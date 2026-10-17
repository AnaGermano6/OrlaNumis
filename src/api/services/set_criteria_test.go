package services

import (
	"strings"
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupSetCriteriaTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(
		&models.CoinSet{}, &models.CoinSetMembership{}, &models.CoinSetTarget{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func newTestSetService(t *testing.T) (*SetService, *gorm.DB) {
	t.Helper()
	db := setupSetCriteriaTestDB(t)
	setRepo := repository.NewSetRepository(db)
	svc := NewSetService(setRepo)
	return svc, db
}

// ---- ValidateSmartCriteria ----

func TestValidateSmartCriteria_ValidSingleRule(t *testing.T) {
	criteria := map[string]interface{}{
		"operator": "and",
		"rules": []interface{}{
			map[string]interface{}{"field": "material", "op": "eq", "value": "Silver"},
		},
	}
	if err := ValidateSmartCriteria(criteria); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidateSmartCriteria_InvalidField(t *testing.T) {
	criteria := map[string]interface{}{
		"operator": "and",
		"rules": []interface{}{
			map[string]interface{}{"field": "badfield", "op": "eq", "value": "x"},
		},
	}
	err := ValidateSmartCriteria(criteria)
	if err == nil {
		t.Fatal("expected error for invalid field")
	}
	if !strings.Contains(err.Error(), "badfield") {
		t.Errorf("error should mention bad field, got: %v", err)
	}
}

func TestValidateSmartCriteria_InvalidOperator(t *testing.T) {
	criteria := map[string]interface{}{
		"operator": "and",
		"rules": []interface{}{
			map[string]interface{}{"field": "material", "op": "LIKE", "value": "Silver"},
		},
	}
	err := ValidateSmartCriteria(criteria)
	if err == nil {
		t.Fatal("expected error for invalid operator")
	}
}

func TestValidateSmartCriteria_InvalidGroupOperator(t *testing.T) {
	criteria := map[string]interface{}{
		"operator": "xor",
		"rules": []interface{}{
			map[string]interface{}{"field": "material", "op": "eq", "value": "Gold"},
		},
	}
	err := ValidateSmartCriteria(criteria)
	if err == nil {
		t.Fatal("expected error for invalid group operator")
	}
}

func TestValidateSmartCriteria_EmptyRules(t *testing.T) {
	criteria := map[string]interface{}{"operator": "and", "rules": []interface{}{}}
	if err := ValidateSmartCriteria(criteria); err == nil {
		t.Fatal("expected error for empty rules")
	}
}

func TestValidateSmartCriteria_NestedGroups(t *testing.T) {
	criteria := map[string]interface{}{
		"operator": "or",
		"rules": []interface{}{
			map[string]interface{}{"field": "material", "op": "eq", "value": "Silver"},
			map[string]interface{}{
				"operator": "and",
				"rules": []interface{}{
					map[string]interface{}{"field": "category", "op": "eq", "value": "Roman"},
					map[string]interface{}{"field": "isWishlist", "op": "eq", "value": true},
				},
			},
		},
	}
	if err := ValidateSmartCriteria(criteria); err != nil {
		t.Errorf("expected no error for nested groups, got %v", err)
	}
}

func TestValidateSmartCriteria_SimpleFiltersAndKindOfCoin(t *testing.T) {
	rules := func(rules ...interface{}) map[string]interface{} {
		return map[string]interface{}{"operator": "and", "rules": rules}
	}
	valid := rules(
		map[string]interface{}{"field": "country", "op": "eq", "value": "Portugal"},
		map[string]interface{}{"field": "year", "op": "between", "value": []interface{}{2002, 2010}},
		map[string]interface{}{"field": "presentation", "op": "eq", "value": "coincard"},
		map[string]interface{}{"field": "kind", "op": "eq", "value": "coin"},
		map[string]interface{}{"field": "classification", "op": "eq", "value": "commemorative"},
	)
	if err := ValidateSmartCriteria(valid); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	for _, bad := range []map[string]interface{}{
		rules(map[string]interface{}{"field": "classification", "op": "eq", "value": "ancient"}),
		rules(map[string]interface{}{"field": "classification", "op": "neq", "value": "collector"}),
		rules(map[string]interface{}{"field": "id", "op": "in", "value": []interface{}{1, 2}}),
	} {
		if err := ValidateSmartCriteria(bad); err == nil {
			t.Errorf("expected an error for %v", bad)
		}
	}
}
