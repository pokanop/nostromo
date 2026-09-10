package model

// Substitution at a given scope for altering arguments
type Substitution struct {
	Name  string `json:"name" yaml:"name"`
	Alias string `json:"alias" yaml:"alias"`
}
