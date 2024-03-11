package app

import (
	"errors"
	en_locales "github.com/go-playground/locales/en"
	"github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

// InitValidator sets up the validator and validator translator.
func InitValidator() (*validator.Validate, ut.Translator, error) {
	// Initialize translator for validator
	translator := en_locales.New()
	universalTranslator := ut.New(translator, translator)
	trans, found := universalTranslator.GetTranslator("en")
	if !found {
		return nil, nil, errors.New("translator not found")
	}
	// Initialize validator
	valid := validator.New()
	// Register translator with validator
	if err := en_translations.RegisterDefaultTranslations(valid, trans); err != nil {
		return nil, nil, err
	}
	// Register missing required field translation
	if err := valid.RegisterTranslation("required", trans, func(ut ut.Translator) error {
		return ut.Add("required", "{0} is a required field", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("required", fe.Field())
		return t
	}); err != nil {
		return nil, nil, err
	}
	// Register invalid email translation
	if err := valid.RegisterTranslation("email", trans, func(ut ut.Translator) error {
		return ut.Add("email", "{0} must be a valid email", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("email", fe.Field())
		return t
	}); err != nil {
		return nil, nil, err
	}
	// Register invalid password strength translation
	if err := valid.RegisterTranslation("passwd", trans, func(ut ut.Translator) error {
		return ut.Add("passwd", "{0} is not strong enough", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("passwd", fe.Field())
		return t
	}); err != nil {
		return nil, nil, err
	}
	// Register password custom validation
	if err := valid.RegisterValidation("passwd", func(fl validator.FieldLevel) bool {
		return len(fl.Field().String()) > 6
	}); err != nil {
		return nil, nil, err
	}
	return valid, trans, nil
}
