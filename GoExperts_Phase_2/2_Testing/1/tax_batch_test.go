package tax

import (
	"math"
	"testing"
)

// Rodar o teste:  go test .
// Rodar o teste verboso:  go test . -v
// Rodar o teste com cobertura:  go test . -coverprofile=coverage.out

// Descobrir o que esta fora da cobertura:
// go tool cover -html=coverage.out

func TestCalculateTaxBatch(t *testing.T) {
	tests := []struct {
		name   string
		amount float64
		expect float64
	}{
		{
			name:   "amount below 1000",
			amount: 500.0,
			expect: 5.0,
		},
		{
			name:   "amount equal to 1000",
			amount: 1000.0,
			expect: 10.0,
		},
		{
			name:   "amount above 1000",
			amount: 1500.0,
			expect: 10.0,
		},
	}

	for _, item := range tests {
		t.Run(item.name, func(t *testing.T) {
			result := CalculateTax(item.amount)
			if result != item.expect {
				t.Errorf("CalculateTax(%v) = %v; expect %v", item.amount, result, item.expect)
			}
		})
	}
}

// BENCHMARK

// go test -bench=.
func BenchmarkCalculateTax(b *testing.B) {
	for b.Loop() {
		_ = CalculateTax(1500.0)
	}
}

// Comparando algorítimos
func BenchmarkCalculateTax2(b *testing.B) {
	for b.Loop() {
		_ = CalculateTax2(1500.0)
	}
}

// **Sub-benchmarks**
//   - Você pode rodar vários cenários no mesmo benchmark:
func BenchmarkCalculateTax3(b *testing.B) {
	b.Run("below-1000", func(b *testing.B) {
		for b.Loop() {
			_ = CalculateTax(500)
		}
	})
	b.Run("equal-1000", func(b *testing.B) {
		for b.Loop() {
			_ = CalculateTax(1000)
		}
	})
	b.Run("above-1000", func(b *testing.B) {
		for b.Loop() {
			_ = CalculateTax(1500)
		}
	})
}

// FUZZ

// go test -fuzz=Fuzz
// go test -fuzz=Fuzz -fuzztime=10s -run=^$
// (-run=^$ evita rodar testes não-fuzz.)
func FuzzCalculateTax(f *testing.F) {
	// Seeds cobrindo pontos relevantes e casos de borda simples
	f.Add(float64(0))
	f.Add(float64(-50))
	f.Add(999.99)
	f.Add(1000.0)
	f.Add(1000.01)

	f.Fuzz(func(t *testing.T, amount float64) {
		// Evita valores não finitos, que atrapalham a oráculo simples
		if math.IsNaN(amount) || math.IsInf(amount, 0) {
			t.Skip()
		}

		got := CalculateTax(amount)

		if got != 0.0 && got != 5.0 && got != 10.0 {
			t.Fatalf("valor inesperado: CalculateTax(%v) = %v (esperado 0.0, 5.0 ou 10.0)", amount, got)
		}

		if amount == 0 && got != 0.0 {
			t.Fatalf("esperado 0.0 para amount=0; amount=%v, got=%v", amount, got)
		}
		if amount > 0 && amount < 1000 && got != 5.0 {
			t.Fatalf("esperado 5.0 para 0<amount<1000; amount=%v, got=%v", amount, got)
		}
		if amount >= 1000 && got != 10.0 {
			t.Fatalf("esperado 10.0 para amount>=1000; amount=%v, got=%v", amount, got)
		}

	})
}
