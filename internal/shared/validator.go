package shared

import (
	"regexp"
	"strings"

	validator "github.com/go-playground/validator/v10"
)

func NewValidator() (*validator.Validate, error) {
	v := validator.New()

	if err := v.RegisterValidation("password_chars_contain", PasswordCharsContain); err != nil {
		return nil, err
	}

	if err := v.RegisterValidation("username_chars_allowed", UsernameCharsAllowed); err != nil {
		return nil, err
	}

	if err := v.RegisterValidation("code_chars_allowed", CodeCharsAllowed); err != nil {
		return nil, err
	}

	if err := v.RegisterValidation("alpha_space", AlphaSpaceOnly); err != nil {
		return nil, err
	}

	if err := v.RegisterValidation("no_spaces", NoSpaces); err != nil {
		return nil, err
	}

	v.RegisterAlias("password", "min=8,password_chars_contain")

	return v, nil
}

func PasswordCharsContain(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSymbol := regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(password)

	return hasUpper && hasLower && hasNumber && hasSymbol
}

func UsernameCharsAllowed(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	isAllowed := regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(username)

	return isAllowed
}

func CodeCharsAllowed(fl validator.FieldLevel) bool {
	code := fl.Field().String()
	isAllowed := regexp.MustCompile(`^[A-Z0-9._]+$`).MatchString(code)

	return isAllowed
}

func AlphaSpaceOnly(fl validator.FieldLevel) bool {
	text := fl.Field().String()
	isValid := regexp.MustCompile(`^[a-zA-Z\s]+$`).MatchString(text)

	return isValid
}

func NoSpaces(fl validator.FieldLevel) bool {
	text := fl.Field().String()

	return !strings.Contains(text, " ")
}
