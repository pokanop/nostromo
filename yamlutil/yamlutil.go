package yamlutil

import (
	"fmt"
	"reflect"
	"strings"

	"gopkg.in/yaml.v2"
)

// Unmarshal decodes YAML into out like yaml.Unmarshal but also accepts
// documents whose mapping keys differ only by case from the yaml field names
// of the target structs, e.g. "aliasonly" for a field tagged "aliasOnly".
//
// Files written before yaml tags existed used lowercased field names, so keys
// are matched case-insensitively against the struct fields and rewritten to
// the tagged names before decoding. Nested structs, maps and slices are
// walked the same way.
//
// Returns true if any key was rewritten, which means the source was in the
// legacy format and should be saved again to migrate it.
func Unmarshal(b []byte, out interface{}) (bool, error) {
	if out == nil {
		return false, fmt.Errorf("cannot unmarshal into nil")
	}

	var doc interface{}
	if err := yaml.Unmarshal(b, &doc); err != nil {
		return false, err
	}

	if !normalize(doc, reflect.TypeOf(out)) {
		return false, yaml.Unmarshal(b, out)
	}

	b, err := yaml.Marshal(doc)
	if err != nil {
		return false, err
	}
	return true, yaml.Unmarshal(b, out)
}

// field describes a yaml mapping entry of a struct
type field struct {
	name string
	typ  reflect.Type
}

// normalize walks a generically decoded document alongside the type it will
// be decoded into and rewrites legacy keys in place. Returns true if any key
// was changed.
func normalize(v interface{}, t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.Struct:
		m, ok := v.(map[interface{}]interface{})
		if !ok {
			return false
		}
		return normalizeStruct(m, t)
	case reflect.Map:
		m, ok := v.(map[interface{}]interface{})
		if !ok {
			return false
		}
		changed := false
		for _, val := range m {
			if normalize(val, t.Elem()) {
				changed = true
			}
		}
		return changed
	case reflect.Slice, reflect.Array:
		s, ok := v.([]interface{})
		if !ok {
			return false
		}
		changed := false
		for _, val := range s {
			if normalize(val, t.Elem()) {
				changed = true
			}
		}
		return changed
	}

	return false
}

// normalizeStruct rewrites the keys of a mapping to the yaml field names of
// the struct type and recurses into the values
func normalizeStruct(m map[interface{}]interface{}, t reflect.Type) bool {
	fields := fieldsOf(t)
	changed := false

	// Collect renames first so the map is not mutated while iterating
	renames := map[string]string{}
	for k, val := range m {
		key, ok := k.(string)
		if !ok {
			continue
		}
		f, ok := fields[strings.ToLower(key)]
		if !ok {
			continue
		}
		if key != f.name {
			renames[key] = f.name
		}
		if normalize(val, f.typ) {
			changed = true
		}
	}

	for legacy, name := range renames {
		val := m[legacy]
		delete(m, legacy)
		// Prefer an existing canonical key over the legacy duplicate
		if _, exists := m[name]; !exists {
			m[name] = val
		}
		changed = true
	}

	return changed
}

// fieldsOf returns the yaml mapping entries of a struct type indexed by
// their lowercased key, following the naming rules of yaml.v2
func fieldsOf(t reflect.Type) map[string]field {
	fields := map[string]field{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			// Unexported
			continue
		}

		tag := f.Tag.Get("yaml")
		if tag == "-" {
			continue
		}

		name := tag
		inline := false
		if idx := strings.Index(tag, ","); idx != -1 {
			name = tag[:idx]
			for _, opt := range strings.Split(tag[idx+1:], ",") {
				if opt == "inline" {
					inline = true
				}
			}
		}

		if inline {
			ft := f.Type
			for ft.Kind() == reflect.Ptr {
				ft = ft.Elem()
			}
			if ft.Kind() == reflect.Struct {
				for k, v := range fieldsOf(ft) {
					fields[k] = v
				}
			}
			continue
		}

		if len(name) == 0 {
			name = strings.ToLower(f.Name)
		}
		fields[strings.ToLower(name)] = field{name, f.Type}
	}
	return fields
}
