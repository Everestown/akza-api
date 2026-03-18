package handler

import (
	"github.com/akza/akza-api/internal/pkg/apperror"
	"github.com/go-playground/validator/v10"
)

func validationErr(err error) error {
	if ve, ok := err.(validator.ValidationErrors); ok {
		return apperror.Validation(ve.Error())
	}
	return apperror.Validation(err.Error())
}
