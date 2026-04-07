package i18n

import (
	"fmt"
	"path"

	"auction-service/global"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
)

var bundle *i18n.Bundle

func init() {
	languageDir := path.Join(global.GetBaseDir(), "storage", "i18n", "language")

	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("yml", yaml.Unmarshal)
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
	bundle.LoadMessageFile(path.Join(languageDir, "en.yml"))
	bundle.LoadMessageFile(path.Join(languageDir, "id.yml"))
}

type Localizer struct {
	l *i18n.Localizer
}

func NewLocalizer(langs ...string) *Localizer {
	return &Localizer{
		l: i18n.NewLocalizer(bundle, langs...),
	}
}

func (l *Localizer) Translate(message string, contents ...string) (string, error) {
	localization, err := l.l.Localize(&i18n.LocalizeConfig{
		MessageID: message,
	})

	if err != nil {
		if _, ok := err.(*i18n.MessageNotFoundErr); ok {
			return message, err
		}
		return "", err
	}

	if len(contents) > 0 {
		datas := []any{}
		for _, content := range contents {
			datas = append(datas, content)
		}
		localization = fmt.Sprintf(localization, datas...)
	}

	return localization, err
}

func (l *Localizer) MustTranslate(message string, contents ...string) string {
	localization, err := l.l.Localize(&i18n.LocalizeConfig{
		MessageID: message,
	})

	if err != nil {
		if _, ok := err.(*i18n.MessageNotFoundErr); ok {
			return message
		}
	}

	if len(contents) > 0 {
		datas := []any{}
		for _, content := range contents {
			datas = append(datas, content)
		}
		localization = fmt.Sprintf(localization, datas...)
	}

	return localization
}
