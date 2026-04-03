package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCustomCategory_JSON(t *testing.T) {
	now := time.Now()
	bgColor := "#ff0000"
	cat := CustomCategory{
		ID:         "cat_1",
		Name:       "Test Category",
		Slug:       "test-category",
		IsActive:   true,
		ShowBadge:  true,
		Searchable: true,
		BgColor:    &bgColor,
		CreatedAt:  now,
		ItemCount:  10,
	}

	data, err := json.Marshal(cat)
	assert.NoError(t, err)

	var cat2 CustomCategory
	err = json.Unmarshal(data, &cat2)
	assert.NoError(t, err)
	assert.Equal(t, cat.ID, cat2.ID)
	assert.Equal(t, cat.Name, cat2.Name)
	assert.Equal(t, *cat.BgColor, *cat2.BgColor)
	// Time comparison can be tricky with JSON marshaling due to precision
	assert.WithinDuration(t, cat.CreatedAt, cat2.CreatedAt, time.Second)
}

func TestBounty_JSON(t *testing.T) {
	targetPrice := 100.0
	bounty := Bounty{
		ID:            "b-1",
		Name:          "Black Lotus",
		TCG:           "mtg",
		Language:      "en",
		TargetPrice:   &targetPrice,
		QuantityNeeded: 1,
		IsActive:      true,
	}

	data, err := json.Marshal(bounty)
	assert.NoError(t, err)

	var bounty2 Bounty
	err = json.Unmarshal(data, &bounty2)
	assert.NoError(t, err)
	assert.Equal(t, bounty.ID, bounty2.ID)
	assert.Equal(t, *bounty.TargetPrice, *bounty2.TargetPrice)
}

func TestProduct_JSON(t *testing.T) {
	price := 50.0
	prod := Product{
		ID:    "p-1",
		Name:  "Test Product",
		TCG:   "mtg",
		Price: price,
	}

	data, err := json.Marshal(prod)
	assert.NoError(t, err)

	var prod2 Product
	err = json.Unmarshal(data, &prod2)
	assert.NoError(t, err)
	assert.Equal(t, prod.ID, prod2.ID)
	assert.Equal(t, prod.Price, prod2.Price)
}

func TestCustomer_JSON(t *testing.T) {
	email := "test@example.com"
	cust := Customer{
		ID:        "c-1",
		FirstName: "John",
		LastName:  "Doe",
		Email:     &email,
	}

	data, err := json.Marshal(cust)
	assert.NoError(t, err)

	var cust2 Customer
	err = json.Unmarshal(data, &cust2)
	assert.NoError(t, err)
	assert.Equal(t, cust.ID, cust2.ID)
	assert.Equal(t, *cust.Email, *cust2.Email)
}
