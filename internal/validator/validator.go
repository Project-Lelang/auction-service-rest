package validator

import (
	"reflect"
	"strings"
	"sync"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/id"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	idTranslations "github.com/go-playground/validator/v10/translations/id"
)

var (
	once        sync.Once
	v           *validator.Validate
	uni         *ut.UniversalTranslator
	Translators map[string]ut.Translator
)

type StructValidationErrors = validator.ValidationErrors

func init() {
	once.Do(func() {
		enLocale := en.New()
		idLocale := id.New()
		uni = ut.New(enLocale, enLocale, idLocale)

		v = validator.New()

		// use JSON tag name as field name
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})

		enTrans, _ := uni.GetTranslator("en")
		idTrans, _ := uni.GetTranslator("id")

		_ = enTranslations.RegisterDefaultTranslations(v, enTrans)
		_ = idTranslations.RegisterDefaultTranslations(v, idTrans)

		Translators = map[string]ut.Translator{
			"en": enTrans,
			"id": idTrans,
		}

		binding.Validator = &defaultValidator{validate: v}
	})
}

type defaultValidator struct {
	validate *validator.Validate
}

func (dv *defaultValidator) ValidateStruct(obj interface{}) error {
	if kindOfData(obj) == reflect.Struct {
		if err := dv.validate.Struct(obj); err != nil {
			return err
		}
	}
	return nil
}

func (dv *defaultValidator) Engine() interface{} {
	return dv.validate
}

func kindOfData(data interface{}) reflect.Kind {
	value := reflect.ValueOf(data)
	valueType := value.Kind()
	if valueType == reflect.Ptr {
		valueType = value.Elem().Kind()
	}
	return valueType
}

func GetTranslator(locale string) ut.Translator {
	if t, ok := Translators[locale]; ok {
		return t
	}
	t, _ := uni.GetTranslator("en")
	return t
}
