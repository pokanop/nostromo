package model

// Substitution at a given scope for altering arguments
type Substitution struct {
	Name  string
	Alias string
}

// Keys as ordered list of fields for logging
func (s *Substitution) Keys() []string {
	return []string{"alias", "name"}
}

// Fields interface for logging
func (s *Substitution) Fields() map[string]interface{} {
	return map[string]interface{}{
		"alias": s.Alias,
		"name":  s.Name,
	}
}
