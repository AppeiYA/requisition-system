package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// init teaches gin's shared validator to report fields by their JSON key
// instead of their Go struct name. It must run before any handler binds a
// request, so it lives here where validation errors are interpreted.
func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(jsonFieldName)
	}
}

// jsonFieldName returns the name part of a field's `json` tag, falling back to
// the Go field name (handled by the validator) when the tag is absent.
func jsonFieldName(fld reflect.StructField) string {
	name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
	if name == "-" {
		return ""
	}
	return name
}

// ValidationErrors interprets the error returned by gin's binding methods
// (c.ShouldBindJSON, c.ShouldBind, ...) into field-level details suitable for
// the Details field of ErrorWithDetails.
//
// It understands:
//   - validator.ValidationErrors: one entry per failed field rule.
//   - *json.UnmarshalTypeError: a field carrying the wrong JSON type.
//   - *json.SyntaxError: a malformed JSON body.
//   - io.EOF / io.ErrUnexpectedEOF: an empty or truncated body.
//
// It never returns nil for a non-nil error, so callers can pass the result
// straight through and still get a useful response.
func ValidationErrors(err error) []FieldError {
	if err == nil {
		return nil
	}

	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		return fromValidationErrors(validationErrs)
	}

	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		field := typeErr.Field
		if field == "" {
			field = bodyField
		}
		return []FieldError{{
			Field:   field,
			Message: fmt.Sprintf("must be a valid %s", typeErr.Type),
		}}
	}

	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return []FieldError{{
			Field:   bodyField,
			Message: fmt.Sprintf("malformed JSON body at position %d", syntaxErr.Offset),
		}}
	}

	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return []FieldError{{
			Field:   bodyField,
			Message: "request body must not be empty",
		}}
	}

	return []FieldError{{
		Field:   bodyField,
		Message: "invalid request body",
	}}
}

// bodyField is used for errors that cannot be attributed to a single field.
const bodyField = "body"

func fromValidationErrors(errs validator.ValidationErrors) []FieldError {
	details := make([]FieldError, 0, len(errs))
	for _, fieldErr := range errs {
		details = append(details, FieldError{
			Field:   fieldErr.Field(),
			Message: validationMessage(fieldErr),
		})
	}
	return details
}

// validationMessage renders a human-readable message for a single failed rule.
func validationMessage(fieldErr validator.FieldError) string {
	switch fieldErr.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email address"
	case "url":
		return "must be a valid URL"
	case "uuid":
		return "must be a valid UUID"
	case "alphanum":
		return "must contain only alphanumeric characters"
	case "numeric":
		return "must be numeric"
	case "min":
		return fmt.Sprintf("must be at least %s characters long", fieldErr.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters long", fieldErr.Param())
	case "len":
		return fmt.Sprintf("must be exactly %s characters long", fieldErr.Param())
	case "gte":
		return fmt.Sprintf("must be greater than or equal to %s", fieldErr.Param())
	case "lte":
		return fmt.Sprintf("must be less than or equal to %s", fieldErr.Param())
	case "gt":
		return fmt.Sprintf("must be greater than %s", fieldErr.Param())
	case "lt":
		return fmt.Sprintf("must be less than %s", fieldErr.Param())
	case "oneof":
		return fmt.Sprintf("must be one of [%s]", fieldErr.Param())
	default:
		return fmt.Sprintf("failed the '%s' validation", fieldErr.Tag())
	}
}
