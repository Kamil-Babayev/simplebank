package api

import (
	"simplebank/util"

	"github.com/go-playground/validator/v10"
)

func validCurrency(fieldLevel validator.FieldLevel) bool {
	if currency, ok := fieldLevel.Field().Interface().(string); ok {
		return util.IsSupportedCurrency(currency)
	}

	return false
}
