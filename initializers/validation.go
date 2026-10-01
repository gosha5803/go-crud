package initializers

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

type ClientError struct {
	Message string `json:"message"`
}

// TODO Вообще обработку ошибок можно сделать в формате цепочки ответственностей
// Ошибка типа валидации, ошибка grpc, ошибка notFound, Forbiden
// Название такое будто она не должна проверятьтип, а тип должен проверяться извне
func FormatValidationErrors(err error) []ClientError {
	var validationErrors validator.ValidationErrors

	if !errors.As(err, &validationErrors) {
		return []ClientError{ClientError{Message: "Неизвестная ошибка. Попробуйте позже"}}
	}

	errors := make([]ClientError, 0, len(validationErrors))

	for _, err := range validationErrors {
		errorMessage := FormatValidationError(err)
		errors = append(errors, ClientError{Message: errorMessage})
	}

	return errors

}

func FormatValidationError(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return err.Field() + " обязательно для заполнения"
	case "min":
		return err.Field() + " должно быть не короче " + err.Param() + " символов"
	case "max":
		return err.Field() + " должно быть не длиннее " + err.Param() + " символов"
	default:
		return err.Field() + " не валидно (" + err.Tag() + ")"
	}
}
