package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"
)

// Embed locales directly into the binary
//
//go:embed locales/*.json
var embeddedLocales embed.FS

type Messages struct {
	App      map[string]string
	Menu     map[string]string
	Download map[string]string
	Settings map[string]string
	Prompt   map[string]string
	Error    map[string]string
	Success  map[string]string
	History  map[string]string
	Archive  map[string]string
}

type Translator struct {
	mu       sync.RWMutex
	lang     string
	messages map[string]*Messages
}

var globalTranslator *Translator

// LoadFromEmbed loads locales from embedded files
func LoadFromEmbed() error {
	globalTranslator = &Translator{
		lang:     "en",
		messages: make(map[string]*Messages),
	}

	langs := []string{"en", "ru"}
	for _, lang := range langs {
		data, err := embeddedLocales.ReadFile(filepath.Join("locales", lang+".json"))
		if err != nil {
			return fmt.Errorf("failed to load embedded %s: %w", lang, err)
		}

		msgs := &Messages{}
		if err := json.Unmarshal(data, msgs); err != nil {
			return fmt.Errorf("failed to parse %s: %w", lang, err)
		}
		globalTranslator.messages[lang] = msgs
	}

	return nil
}

func Load(localesPath string) error {
	globalTranslator = &Translator{
		lang:     "en",
		messages: make(map[string]*Messages),
	}

	langs := []string{"en", "ru"}
	for _, lang := range langs {
		jsonPath := filepath.Join(localesPath, lang+".json")
		data, err := ioutil.ReadFile(jsonPath)
		if err != nil {
			return fmt.Errorf("failed to load %s: %w", lang, err)
		}

		msgs := &Messages{}
		if err := json.Unmarshal(data, msgs); err != nil {
			return fmt.Errorf("failed to parse %s: %w", lang, err)
		}
		globalTranslator.messages[lang] = msgs
	}

	return nil
}

func SetLanguage(lang string) {
	if globalTranslator == nil {
		return
	}
	globalTranslator.mu.Lock()
	defer globalTranslator.mu.Unlock()

	if _, ok := globalTranslator.messages[lang]; ok {
		globalTranslator.lang = lang
	}
}

func SetGlobal(lang string) {
	SetLanguage(lang)
}

func T(section, key string) string {
	if globalTranslator == nil {
		return key
	}

	globalTranslator.mu.RLock()
	defer globalTranslator.mu.RUnlock()

	msgs, ok := globalTranslator.messages[globalTranslator.lang]
	if !ok {
		msgs = globalTranslator.messages["en"]
	}

	// Safety check: if msgs is still nil, return key
	if msgs == nil {
		return key
	}

	switch section {
	case "menu":
		if v, ok := msgs.Menu[key]; ok {
			return v
		}
	case "download":
		if v, ok := msgs.Download[key]; ok {
			return v
		}
	case "settings":
		if v, ok := msgs.Settings[key]; ok {
			return v
		}
	case "prompt":
		if v, ok := msgs.Prompt[key]; ok {
			return v
		}
	case "error":
		if v, ok := msgs.Error[key]; ok {
			return v
		}
	case "success":
		if v, ok := msgs.Success[key]; ok {
			return v
		}
	case "app":
		if v, ok := msgs.App[key]; ok {
			return v
		}
	case "history":
		if v, ok := msgs.History[key]; ok {
			return v
		}
	case "archive":
		if v, ok := msgs.Archive[key]; ok {
			return v
		}
	}

	return key
}
