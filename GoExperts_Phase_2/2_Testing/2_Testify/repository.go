package tax

// Repository define a dependência externa responsável por persistir a taxa.
type Repository interface {
	SaveTax(amount float64) error
}
