// internal/entity/product_test.go
package entity

import (
	"testing"

	"github.com/google/uuid"
)

func TestProductValidate_Success(t *testing.T) {
	p := &Product{
		ID:    uuid.New(), // compatível com pkgentity.ID (alias de uuid.UUID)
		Name:  "Produto X",
		Price: 10.50,
		// CreatedAt não é validado
	}
	if err := p.validate(); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestProductValidate_NilReceiver(t *testing.T) {
	var p *Product
	err := p.validate()
	if err == nil {
		t.Fatalf("expected error for nil receiver, got nil")
	}
	if err != ErrRequired {
		t.Fatalf("expected ErrRequired, got %v", err)
	}
}

func TestProductValidate_InvalidID(t *testing.T) {
	p := &Product{
		ID:    uuid.Nil,
		Name:  "Produto Y",
		Price: 1.0,
	}
	err := p.validate()
	if err == nil {
		t.Fatalf("expected error for invalid id, got nil")
	}
	if err != ErrInvalidID {
		t.Fatalf("expected ErrInvalidID, got %v", err)
	}
}

func TestProductValidate_EmptyName(t *testing.T) {
	tests := []struct {
		name string
		in   *Product
	}{
		{
			name: "empty",
			in: &Product{
				ID:    uuid.New(),
				Name:  "",
				Price: 5.0,
			},
		},
		{
			name: "spaces",
			in: &Product{
				ID:    uuid.New(),
				Name:  "   ",
				Price: 5.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.in.validate()
			if err == nil {
				t.Fatalf("expected error for empty/blank name, got nil")
			}
			if err != ErrRequired {
				t.Fatalf("expected ErrRequired, got %v", err)
			}
		})
	}
}

func TestProductValidate_ZeroPrice(t *testing.T) {
	p := &Product{
		ID:    uuid.New(),
		Name:  "Produto Z",
		Price: 0,
	}
	err := p.validate()
	if err == nil {
		t.Fatalf("expected error for zero price, got nil")
	}
	if err != ErrRequired {
		t.Fatalf("expected ErrRequired for zero price, got %v", err)
	}
}

func TestProductValidate_NegativePrice(t *testing.T) {
	p := &Product{
		ID:    uuid.New(),
		Name:  "Produto W",
		Price: -1.0,
	}
	err := p.validate()
	if err == nil {
		t.Fatalf("expected error for negative price, got nil")
	}
	if err != ErrInvalidPrice {
		t.Fatalf("expected ErrInvalidPrice, got %v", err)
	}
}
