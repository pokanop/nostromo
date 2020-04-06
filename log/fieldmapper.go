package log

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
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

		svalue := fmt.Sprint(value)
		if len(svalue) == 0 {
			continue
		}

		fmt.Print(opt.theme.formatStyle(keyFieldStyle, key))
		fmt.Print(opt.theme.formatStyle(valueFieldStyle, ": "+svalue))
		fmt.Print(" ")
	}

	fmt.Println()
}

// Table logs key value pairs as a table with keys for the header
func Table(mapper FieldMapper) {
	if mapper == nil {
		return
	}

	keys := mapper.Keys()
	fields := mapper.Fields()

	if keys == nil || fields == nil {
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetColMinWidth(0, 12)
	table.SetColWidth(68)

	for _, key := range keys {
		value := fields[key]
		if value == nil || omitKey(key) {
			continue
		}

		svalue := fmt.Sprint(value)
		if len(svalue) == 0 {
			continue
		}

		key = opt.theme.formatStyle(keyFieldStyle, key).String()
		table.Append([]string{key, svalue})
	}

	table.Render()
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
