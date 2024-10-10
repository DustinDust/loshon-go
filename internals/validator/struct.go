package validator

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type Validator struct {
	Validator *validator.Validate
}

type StructValidationErrors struct {
	FieldErrors validator.ValidationErrors
}

func NewValidator() *Validator {
	v := validator.New(validator.WithRequiredStructEnabled())

	// Get json name from field using reflect & tags
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Validator{
		Validator: v,
	}
}

func (v *Validator) ValidateStruct(args interface{}) error {
	if err := v.Validator.Struct(args); err != nil {
		verr, ok := err.(validator.ValidationErrors)
		if !ok {
			return err
		}
		return &StructValidationErrors{
			FieldErrors: verr,
		}
	}
	return nil
}

// Debug & help subscribe StructValidationError to the Error interface
func (ve *StructValidationErrors) Error() string {
	return ve.FieldErrors.Error()
}

// Return a echo.HttpError
func (ve *StructValidationErrors) TranslateToHttpError() error {
	errMessage := "invalid data format"
	errData := []interface{}{}

	for _, e := range ve.FieldErrors {
		errData = append(errData, echo.Map{
			"field": e.Field(),
			"expected": fmt.Sprintf("%s%s", e.ActualTag(), func() string {
				if e.Param() != "" {
					return "=" + e.Param()
				} else {
					return ""
				}
			}()),
			"got":   e.Value(),
			"error": e.Error(),
		})
	}
	return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
		"message": errMessage,
		"errors":  errData,
	})
}
