package exception

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/cockroachdb/errors"
	validator "github.com/go-playground/validator/v10"
)

type ErrorType string

type FieldErrors map[string][]string

type Exception struct {
	Type     ErrorType      `json:"type"`
	Code     string         `json:"code,omitempty"`
	Message  string         `json:"message"`
	Errors   FieldErrors    `json:"errors,omitempty"`
	Metadata map[string]any `json:"-"`
}

func (e *Exception) Error() string {
	return e.Message
}

func (e *Exception) Is(target error) bool {
	t, ok := target.(*Exception)
	if !ok {
		return false
	}

	return e.Type == t.Type
}

func newException(kind ErrorType, code string, message string) *Exception {
	return &Exception{
		Type:     kind,
		Code:     code,
		Message:  message,
		Errors:   make(FieldErrors),
		Metadata: make(map[string]any),
	}
}

func New(kind ErrorType, code string, message string) error {
	marker := newException(kind, code, message)
	return errors.WithStack(marker)
}

func Newf(kind ErrorType, code string, format string, args ...any) error {
	message := fmt.Sprintf(format, args...)
	marker := newException(kind, code, message)

	return errors.WithStack(marker)
}

func NewWithErrors(kind ErrorType, code string, message string, errs FieldErrors) error {
	marker := newException(kind, code, message)
	marker.Errors = errs

	return errors.WithStack(marker)
}

func Wrap(cause error, kind ErrorType, code string, message string) error {
	if cause == nil {
		return New(kind, code, message)
	}

	marker := newException(kind, code, message)

	return errors.Wrapf(marker, "%w", cause)
}

func Wrapf(cause error, kind ErrorType, code string, format string, args ...any) error {
	if cause == nil {
		return Newf(kind, code, format, args...)
	}

	message := fmt.Sprintf(format, args...)
	marker := newException(kind, code, message)

	return errors.Wrapf(marker, "%w", marker)
}

func GetException(err error) (*Exception, bool) {
	var ex *Exception

	ok := errors.As(err, &ex)

	return ex, ok
}

func WithFieldError(err error, field, message string) error {
	ex, ok := GetException(err)
	if !ok {
		return errors.Wrap(err, "cannot add field error: no Exception found in chain")
	}

	if ex.Errors == nil {
		ex.Errors = make(FieldErrors)
	}

	ex.Errors[field] = append(ex.Errors[field], message)

	return err
}

func WithCode(err error, code string) error {
	ex, ok := GetException(err)
	if !ok {
		return errors.Wrap(err, "cannot add code: no Exception found in chain")
	}

	ex.Code = code

	return err
}

func WithMeta(err error, key string, value any) error {
	ex, ok := GetException(err)
	if !ok {
		return errors.Wrap(err, "cannot add metadata: no Exception found in chain")
	}

	if ex.Metadata == nil {
		ex.Metadata = make(map[string]any)
	}

	ex.Metadata[key] = value

	return err
}

func HasFieldErrors(err error) bool {
	ex, ok := GetException(err)
	return ok && len(ex.Errors) > 0
}

func FromValidationErrors(req any, validationErrors validator.ValidationErrors) error {
	err := New(TypeValidationError, CodeValidationFailed, "validation failed")
	ex, ok := GetException(err)

	if !ok {
		return errors.Wrap(err, "internal inconsistency: failed to get exception from new exception")
	}

	if ex.Errors == nil {
		ex.Errors = make(FieldErrors)
	}

	reqType := reflect.TypeOf(req).Elem()
	for _, fieldErr := range validationErrors {
		jsonPath := getJsonPathFromNamespace(reqType, fieldErr.StructNamespace())

		var message string

		switch fieldErr.Tag() {
		case "required":
			message = "This field is required"
		case "min":
			message = fmt.Sprintf("This field must be at least %s characters long", fieldErr.Param())
		case "max":
			message = fmt.Sprintf("This field must be at most %s characters long", fieldErr.Param())
		case "email":
			message = "Invalid email format"
		case "gt":
			message = "This field must be greater than to " + fieldErr.Param()
		case "lt":
			message = "This field must be less than to " + fieldErr.Param()
		case "gte":
			message = "This field must be greater than or equal to " + fieldErr.Param()
		case "lte":
			message = "This field must be less than or equal to " + fieldErr.Param()
		case "eq":
			message = fmt.Sprintf("This field must be equal to '%s'", fieldErr.Param())
		case "ne":
			message = fmt.Sprintf("This field must be not equal to '%s'", fieldErr.Param())
		default:
			message = fmt.Sprintf("field validation for '%s' failed on the '%s' tag", fieldErr.Field(), fieldErr.Tag())
		}

		ex.Errors[jsonPath] = append(ex.Errors[jsonPath], message)
	}

	return err
}

func getJsonPathFromNamespace(rootType reflect.Type, namespace string) string {
	fields := strings.Split(namespace, ".")

	if len(fields) <= 1 {
		return namespace
	}

	jsonPathParts := make([]string, 0, len(fields)-1)
	currentType := rootType

	for _, fieldName := range fields[1:] {
		splited := strings.Split(fieldName, "[")
		fieldName := splited[0]

		var fieldIndex string
		if len(splited) > 1 {
			fieldIndex = strings.Split(splited[1], "]")[0]
		}

		if fieldIndex != "" {
			fieldIndex = "." + fieldIndex
		}

		field, found := currentType.FieldByName(fieldName)
		if !found {
			jsonPathParts = append(jsonPathParts, fmt.Sprintf("%s%s", strings.ToLower(fieldName), fieldIndex))
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			jsonPathParts = append(jsonPathParts, fmt.Sprintf("%s%s", strings.ToLower(fieldName), fieldIndex))
		} else {
			jsonName := strings.Split(jsonTag, ",")[0]
			jsonPathParts = append(jsonPathParts, fmt.Sprintf("%s%s", jsonName, fieldIndex))
		}

		currentType = field.Type

		if currentType.Kind() == reflect.Ptr || currentType.Kind() == reflect.Slice {
			currentType = currentType.Elem()
		}
	}

	return strings.Join(jsonPathParts, ".")
}
