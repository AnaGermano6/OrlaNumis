package services

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type fakeCollectionRates struct {
	rates ExchangeRates
	err   error
	calls int
}

func (f *fakeCollectionRates) Rates(context.Context) (ExchangeRates, error) {
	f.calls++
	return f.rates, f.err
}

func setupCollectionCurrencyDB(t *testing.T) (*gorm.DB, *SettingsService) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&models.AppSetting{}, &models.User{}, &models.Coin{}, &models.CoinValueHistory{}, &models.QuickCaptureDraft{},
		&models.ValueSnapshot{}, &models.CoinSet{}, &models.CoinSetValuationSnapshot{}, &models.ValuationRun{}, &models.ValuationResult{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db, NewSettingsService(repository.NewSettingsRepository(db))
}

func TestCollectionCurrencyChangeConvertsStoredAmounts(t *testing.T) {
	db, settings := setupCollectionCurrencyDB(t)
	rates := &fakeCollectionRates{rates: ExchangeRates{Date: "2026-10-09", PerEUR: map[string]float64{"USD": 1.25, "GBP": 0.8}}}
	svc := NewCollectionCurrencyService(repository.NewCollectionCurrencyRepository(db), settings, rates)

	priced := models.Coin{Name: "Priced", UserID: 1, PurchasePrice: floatPtr(12.5), CurrentValue: floatPtr(3.37)}
	unpriced := models.Coin{Name: "Unpriced", UserID: 1}
	db.Create(&priced)
	db.Create(&unpriced)
	db.Create(&models.CoinValueHistory{CoinID: priced.ID, UserID: 1, Value: 25})
	db.Create(&models.ValueSnapshot{UserID: 1, TotalValue: 100, TotalInvested: 50})

	if settings.CollectionCurrency() != CurrencyUSD {
		t.Fatalf("default currency = %q", settings.CollectionCurrency())
	}
	change, err := svc.Change(context.Background(), "eur")
	if err != nil {
		t.Fatalf("change: %v", err)
	}
	if change.Currency != CurrencyEUR || change.Previous != CurrencyUSD || change.CoinsConverted != 1 || change.Factor != 0.8 || change.RateDate != "2026-10-09" {
		t.Fatalf("change = %+v", change)
	}
	if settings.CollectionCurrency() != CurrencyEUR {
		t.Fatalf("currency after change = %q", settings.CollectionCurrency())
	}

	var coin models.Coin
	db.First(&coin, priced.ID)
	if *coin.PurchasePrice != 10 || *coin.CurrentValue != 2.7 || coin.SoldPrice != nil {
		t.Fatalf("coin amounts = %v %v %v", *coin.PurchasePrice, *coin.CurrentValue, coin.SoldPrice)
	}
	var other models.Coin
	db.First(&other, unpriced.ID)
	if other.PurchasePrice != nil || other.CurrentValue != nil {
		t.Fatalf("an unpriced coin must stay unpriced: %v %v", other.PurchasePrice, other.CurrentValue)
	}
	var history models.CoinValueHistory
	db.First(&history)
	var snapshot models.ValueSnapshot
	db.First(&snapshot)
	if history.Value != 20 || snapshot.TotalValue != 80 || snapshot.TotalInvested != 40 {
		t.Fatalf("history %v, snapshot %v/%v", history.Value, snapshot.TotalValue, snapshot.TotalInvested)
	}

	// Switching to the current currency converts nothing; switching back converts back.
	if change, err := svc.Change(context.Background(), "EUR"); err != nil || change.Factor != 0 || rates.calls != 1 {
		t.Fatalf("no-op change = %+v, %v (rate calls %d)", change, err, rates.calls)
	}
	// Any two currencies of the ECB rates convert through the euro: €10 is £8.
	if change, err := svc.Change(context.Background(), "gbp"); err != nil || change.Factor != 0.8 {
		t.Fatalf("to pounds: %+v, %v", change, err)
	}
	var pounds models.Coin
	db.First(&pounds, priced.ID)
	if *pounds.PurchasePrice != 8 || settings.CollectionCurrency() != "GBP" {
		t.Fatalf("in pounds: %v %q", *pounds.PurchasePrice, settings.CollectionCurrency())
	}
	if _, err := svc.Change(context.Background(), "USD"); err != nil {
		t.Fatal(err)
	}
	var back models.Coin
	db.First(&back, priced.ID)
	if *back.PurchasePrice != 12.5 || settings.CollectionCurrency() != CurrencyUSD {
		t.Fatalf("back to dollars: %v %q", *back.PurchasePrice, settings.CollectionCurrency())
	}
}

func TestCollectionCurrencyChangeRejectsUnknownCurrencyAndMissingRate(t *testing.T) {
	db, settings := setupCollectionCurrencyDB(t)
	db.Create(&models.Coin{Name: "Priced", UserID: 1, CurrentValue: floatPtr(10)})

	if _, err := NewCollectionCurrencyService(repository.NewCollectionCurrencyRepository(db), settings, &fakeCollectionRates{}).Change(context.Background(), "XYZ"); !errors.Is(err, ErrUnsupportedCollectionCurrency) {
		t.Fatalf("XYZ: %v", err)
	}
	// A currency the ECB quoted no rate for today converts nothing.
	noPounds := &fakeCollectionRates{rates: ExchangeRates{Date: "2026-10-09", PerEUR: map[string]float64{"USD": 1.25}}}
	if _, err := NewCollectionCurrencyService(repository.NewCollectionCurrencyRepository(db), settings, noPounds).Change(context.Background(), "GBP"); !errors.Is(err, ErrExchangeRateUnavailable) {
		t.Fatalf("GBP without a rate: %v", err)
	}
	_, err := NewCollectionCurrencyService(repository.NewCollectionCurrencyRepository(db), settings, &fakeCollectionRates{err: errors.New("ecb down")}).Change(context.Background(), "EUR")
	if !errors.Is(err, ErrExchangeRateUnavailable) {
		t.Fatalf("missing rate: %v", err)
	}
	var coin models.Coin
	db.First(&coin)
	if *coin.CurrentValue != 10 || settings.CollectionCurrency() != CurrencyUSD {
		t.Fatalf("nothing may change without a rate: %v %q", *coin.CurrentValue, settings.CollectionCurrency())
	}
}

func TestValuationMessageUsesTheCollectionCurrency(t *testing.T) {
	coin := &models.Coin{Name: "Denarius", PurchasePrice: floatPtr(120), CurrentValue: floatPtr(150)}

	dollars := valuationRequestMessage(coin, CurrencyUSD)
	if !strings.Contains(dollars, "Purchase Price: $120.00") || !strings.Contains(dollars, "Previous value: $150.00") || strings.Contains(dollars, "EUR") {
		t.Fatalf("dollar message: %s", dollars)
	}
	euros := valuationRequestMessage(coin, CurrencyEUR)
	if !strings.Contains(euros, "Purchase Price: €120.00") || !strings.Contains(euros, "Previous value: €150.00") || !strings.Contains(euros, "in EUR, not USD") {
		t.Fatalf("euro message: %s", euros)
	}

	pounds := valuationRequestMessage(coin, "GBP")
	if !strings.Contains(pounds, "Purchase Price: £120.00") || !strings.Contains(pounds, "in GBP, not USD") {
		t.Fatalf("pound message: %s", pounds)
	}
	if FormatMoney(120, "CHF") != "120.00 CHF" {
		t.Fatalf("franc: %s", FormatMoney(120, "CHF"))
	}
}
