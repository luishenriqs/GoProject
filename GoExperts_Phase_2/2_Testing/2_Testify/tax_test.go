package tax

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateTax(t *testing.T) {
	tests := []struct {
		name   string
		amount float64
		expect   float64
	}{
		{"amount == 0", 0.0, 0.0},
		{"amount below 1000", 500.0, 5.0},
		{"amount equal 1000", 1000.0, 10.0},
		{"amount above 1000", 1500.0, 10.0},
	}

	for _, item := range tests {
		t.Run(item.name, func(t *testing.T) {
			result := CalculateTax(item.amount)
			assert.Equal(t, item.expect, result, "CalculateTax(%v)", item.amount)
		})
	}
}
