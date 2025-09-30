package tax

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateTaxAndSave_Success(t *testing.T) {
	tests := []struct {
		name       string
		input      float64
		wantTax    float64
	}{
		{name: "amount == 0", input: 0.0, wantTax: 0.0},
		{name: "amount below 1000", input: 500.0, wantTax: 5.0},
		{name: "amount equal 1000", input: 1000.0, wantTax: 10.0},
		{name: "amount above 1000", input: 1500.0, wantTax: 10.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(RepositoryMock)
			repo.On("SaveTax", tt.wantTax).Return(nil).Once()

			gotTax, err := CalculateTaxAndSave(tt.input, repo)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantTax, gotTax)
			repo.AssertExpectations(t)
		})
	}
}

func TestCalculateTaxAndSave_SaveError(t *testing.T) {
	repo := new(RepositoryMock)
	repo.On("SaveTax", 10.0).Return(errors.New("persist failed")).Once()

	gotTax, err := CalculateTaxAndSave(1500.0, repo)

	assert.Error(t, err)
	assert.EqualError(t, err, "persist failed")
	assert.Equal(t, 0.0, gotTax)
	repo.AssertExpectations(t)
}
