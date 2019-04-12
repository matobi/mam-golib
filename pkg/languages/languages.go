package languages

import (
	"fmt"
	"strings"
)

type Language struct {
	KeyShort          string
	KeyLong           string
	NameLocal         string
	NameInternational string
}

var langs = []Language{
	Language{"da", "dan", "Dansk", "Danish"},
	Language{"en", "eng", "English", "English"},
	Language{"fi", "fin", "Suomi", "Finnish"},
	Language{"no", "nor", "Norsk", "Norwegian"},
	Language{"sv", "swe", "Svenska", "Swedish"},
}

func Search(keys []string) (*Language, error) {
	for _, key := range keys {
		key := strings.TrimSpace(strings.ToLower(key))
		for i := range langs {
			if key == langs[i].KeyShort || key == langs[i].KeyLong {
				lang := langs[i]
				return &lang, nil
			}
		}
	}
	return nil, fmt.Errorf("language not found: %s", strings.Join(keys, ","))
}
