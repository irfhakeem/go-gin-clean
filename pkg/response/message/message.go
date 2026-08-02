package message

import (
	"embed"
	"encoding/json"

	pkgerror "go-gin-clean/pkg/error"
)

type Language string

const (
	EN Language = "en"
	ID Language = "id"
)

//go:embed en.json id.json
var messageFS embed.FS

var texts = make(map[Language]map[pkgerror.ErrCode]string)

func init() {
	texts[EN] = read("en.json")
	texts[ID] = read("id.json")
}

func read(file string) map[pkgerror.ErrCode]string {
	data, err := messageFS.ReadFile(file)
	if err != nil {
		panic(err)
	}

	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		panic(err)
	}

	msgs := make(map[pkgerror.ErrCode]string, len(raw))
	for code, text := range raw {
		msgs[pkgerror.ErrCode(code)] = text
	}
	return msgs
}

func Get(lang Language, code pkgerror.ErrCode) string {
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
