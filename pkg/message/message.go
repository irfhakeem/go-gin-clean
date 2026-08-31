package message

import (
	"embed"
	"encoding/json"
)

type Language string

const (
	EN Language = "en"
	ID Language = "id"
)

//go:embed en.json id.json
var messageFS embed.FS

var texts = make(map[Language]map[MessageKey]string)

func init() {
	texts[EN] = read("en.json")
	texts[ID] = read("id.json")
}

func read(file string) map[MessageKey]string {
	data, err := messageFS.ReadFile(file)
	if err != nil {
		panic(err)
	}

	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		panic(err)
	}

	msgs := make(map[MessageKey]string, len(raw))
	for code, text := range raw {
		msgs[MessageKey(code)] = text
	}
	return msgs
}

func Get(lang Language, code MessageKey) string {
	if translations, ok := texts[lang]; ok {
		if text, ok := translations[code]; ok && text != "" {
			return text
		}
	}
	if fallback, ok := texts[EN]; ok {
		if text, ok := fallback[code]; ok {
			return text
		}
	}
	return string(code)
}
