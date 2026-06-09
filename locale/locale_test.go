package locale

import (
	"reflect"
	"strings"
	"testing"
)

func TestLoad_english(t *testing.T) {
	Load("en")
	if Active != &English {
		t.Error("Load(\"en\") should set Active to &English")
	}
}

func TestLoad_empty(t *testing.T) {
	Load("")
	if Active != &English {
		t.Error("Load(\"\") should default to &English")
	}
}

func TestLoad_unknown(t *testing.T) {
	Load("zz")
	if Active != &English {
		t.Error("Load with unknown language should fall back to &English")
	}
}

// TestEnglish_noEmptyStrings verifies every string field in the English locale is populated.
func TestEnglish_noEmptyStrings(t *testing.T) {
	v := reflect.ValueOf(English)
	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := typ.Field(i)
		val := v.Field(i).String()
		if strings.TrimSpace(val) == "" {
			t.Errorf("English.%s is empty", field.Name)
		}
	}
}
