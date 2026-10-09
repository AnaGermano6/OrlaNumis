package services

import (
	"context"
	"errors"
	"testing"
	"time"
)

// Shaped like the ECB eurofxref-daily.xml; rates are made up.
const testECBRates = `<?xml version="1.0" encoding="UTF-8"?>
<gesmes:Envelope xmlns:gesmes="http://www.gesmes.org/xml/2002-08-01" xmlns="http://www.ecb.int/vocabulary/2002-08-01/eurofxref">
	<gesmes:subject>Reference rates</gesmes:subject>
	<Cube>
		<Cube time='2026-10-09'>
			<Cube currency='JPY' rate='177.34'/>
			<Cube currency='USD' rate='1.2500'/>
		</Cube>
	</Cube>
</gesmes:Envelope>`

type fakeECBClient struct {
	body  string
	err   error
	calls int
}

func (f *fakeECBClient) DailyRates(context.Context) ([]byte, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return []byte(f.body), nil
}

func TestParseECBDollarRate(t *testing.T) {
	rate, err := parseECBDollarRate([]byte(testECBRates))
	if err != nil || rate.USDPerEUR != 1.25 || rate.Date != "2026-10-09" {
		t.Fatalf("rate = %+v (%v)", rate, err)
	}
	if _, err := parseECBDollarRate([]byte(`<Envelope><Cube><Cube time="2026-10-09"><Cube currency="JPY" rate="1"/></Cube></Cube></Envelope>`)); err == nil {
		t.Fatal("missing USD must be an error")
	}
	if _, err := parseECBDollarRate([]byte("not xml")); err == nil {
		t.Fatal("invalid XML must be an error")
	}
}

func TestExchangeRateService_CachesAndKeepsTheLastRate(t *testing.T) {
	_, _, store, clock := newMarketValueFixture()
	client := &fakeECBClient{body: testECBRates}
	svc := NewExchangeRateService(client, NewNumistaResponseCache(store, clock, nil))
	ctx := context.Background()

	for range 3 {
		if rate, err := svc.USDPerEUR(ctx); err != nil || rate.USDPerEUR != 1.25 {
			t.Fatalf("rate = %+v (%v)", rate, err)
		}
	}
	if client.calls != 1 {
		t.Fatalf("expected one ECB call within the cache window, got %d", client.calls)
	}

	clock.now = clock.now.Add(7 * time.Hour)
	client.err = errors.New("ECB down")
	if rate, err := svc.USDPerEUR(ctx); err != nil || rate.Date != "2026-10-09" {
		t.Fatalf("the last stored rate should be used when the ECB is down: %+v (%v)", rate, err)
	}

	_, _, emptyStore, emptyClock := newMarketValueFixture()
	failing := NewExchangeRateService(&fakeECBClient{err: errors.New("offline")}, NewNumistaResponseCache(emptyStore, emptyClock, nil))
	if _, err := failing.USDPerEUR(ctx); err == nil {
		t.Fatal("no rate and no stored copy must be an error")
	}
}
