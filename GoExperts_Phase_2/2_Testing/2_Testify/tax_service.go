package tax

// CalculateTaxAndSave calcula a taxa a partir de amount e persiste o valor
// usando o Repository. Em caso de erro ao salvar, retorna (0, err).
func CalculateTaxAndSave(amount float64, repository Repository) (float64, error) {
	tax := CalculateTax(amount)
	if err := repository.SaveTax(tax); err != nil {
		return 0, err
	}
	return tax, nil
}
