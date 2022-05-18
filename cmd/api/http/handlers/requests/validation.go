package requests

import (
	"github.com/go-playground/validator/v10"
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
)

var (
	validate = validator.New()
	Trans    ut.Translator
	mu       sync.Mutex
)

func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

func ValidateVar(s interface{}, tag string) error {
	return validate.Var(s, tag)
}

func registerTranslations() {
	enTrans := en.New()
	uni := ut.New(enTrans, enTrans)
	Trans, _ = uni.GetTranslator("en")

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	_ = validate.RegisterTranslation(
		"required",
		Trans,
		func(ut ut.Translator) error {
			return ut.Add("required", "field is required", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("required")
			return t
		})
}

func SetupValidator() {
	mu.Lock()
	registerTranslations()
	mu.Unlock()
}
