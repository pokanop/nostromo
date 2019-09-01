package log

import (
	"fmt"
)

type fieldStyle int

const (
	keyFieldStyle fieldStyle = iota
	valueFieldStyle
	headerFieldStyle
	cellFieldStyle
)

var omitKeysCompact = []string{"description", "code"}

// FieldMapper type returns fields as a map for logging
type FieldMapper interface {
	Keys() []string
	Fields() map[string]interface{}
}

// Fields logs key value pairs formatted in sections
func Fields(mapper FieldMapper) {
	fields(mapper, false)
}

// FieldsCompact logs key value pairs formatted on a single line
func FieldsCompact(mapper FieldMapper) {
	fields(mapper, true)
}

func fields(mapper FieldMapper, compact bool) {
	if mapper == nil {
		return
	}

	keys := mapper.Keys()
	fields := mapper.Fields()

	if keys == nil || fields == nil {
		return
	}

	for _, key := range keys {
		value := fields[key]
		if value == nil || omitKey(key, compact) {
			continue
		}
		if compact {
			fmt.Print(currentTheme.formatStyle(keyFieldStyle, key))
			fmt.Print(currentTheme.formatStyle(valueFieldStyle, ": "+fmt.Sprint(value)))
			fmt.Print(" ")
		} else {
			fmt.Print(currentTheme.formatStyle(keyFieldStyle, key))
			fmt.Println(currentTheme.formatStyle(valueFieldStyle, ": "+fmt.Sprint(value)))
		}
	}
	if compact {
		fmt.Println()
	}
}

// Table logs key value pairs as a table with keys for the header
func Table(mapper FieldMapper) {

}

func omitKey(key string, compact bool) bool {
	if compact {
		for _, k := range omitKeysCompact {
			if k == key {
				return true
			}
		}
	}
	return false
}
