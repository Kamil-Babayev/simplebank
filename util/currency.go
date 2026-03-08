package util

const (
	EUR = "EUR"
	GBP = "GBP"
	JPY = "JPY"
	NZD = "NZD"
)

func IsSupportedCurrency(currency string) bool {
	switch currency {
	case EUR, GBP, JPY, NZD:
		return true
	default:
		return false
	}
}
