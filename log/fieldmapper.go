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
		if value == nil || omitKey(key) {
			continue
		}
		if !opt.verbose {
			fmt.Print(opt.theme.formatStyle(keyFieldStyle, key))
			fmt.Print(opt.theme.formatStyle(valueFieldStyle, ": "+fmt.Sprint(value)))
			fmt.Print(" ")
		} else {
			fmt.Print(opt.theme.formatStyle(keyFieldStyle, key))
			fmt.Println(opt.theme.formatStyle(valueFieldStyle, ": "+fmt.Sprint(value)))
		}
	}
	if !opt.verbose {
		fmt.Println()
	}
}

// Table logs key value pairs as a table with keys for the header
func Table(mapper FieldMapper) {

}

func omitKey(key string) bool {
	if !opt.verbose {
		for _, k := range omitKeysCompact {
			if k == key {
				return true
			}
		}
	}
	return false
}
