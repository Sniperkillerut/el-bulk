package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProduct_ComputePrice(t *testing.T) {
	priceReference := 10.0
	priceCOPOverride := 50000.0

	tests := []struct {
		name           string
		product        Product
		usdToCOP       float64
		eurToCOP       float64
		expectedPrice float64
	}{
		{
			name: "Manual Override",
			product: Product{
				PriceCOPOverride: &priceCOPOverride,
				PriceReference:   &priceReference,
				PriceSource:      PriceSourceTCGPlayer,
			},
			usdToCOP:      4000,
			eurToCOP:      4500,
			expectedPrice: 50000.0,
		},
		{
			name: "TCGPlayer USD",
			product: Product{
				PriceReference: &priceReference,
				PriceSource:    PriceSourceTCGPlayer,
			},
			usdToCOP:      4000,
			eurToCOP:      4500,
			expectedPrice: 40000.0,
		},
		{
			name: "Cardmarket EUR",
			product: Product{
				PriceReference: &priceReference,
				PriceSource:    PriceSourceCardmarket,
			},
			usdToCOP:      4000,
			eurToCOP:      4500,
			expectedPrice: 45000.0,
		},
		{
			name: "Missing Reference and Override",
			product: Product{
				PriceSource: PriceSourceTCGPlayer,
			},
			usdToCOP:      4000,
			eurToCOP:      4500,
			expectedPrice: 0.0,
		},
		{
			name: "Unknown Source",
			product: Product{
				PriceReference: &priceReference,
				PriceSource:    "unknown",
			},
			usdToCOP:      4000,
			eurToCOP:      4500,
			expectedPrice: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.product.ComputePrice(tt.usdToCOP, tt.eurToCOP)
			assert.Equal(t, tt.expectedPrice, got)
		})
	}
}
