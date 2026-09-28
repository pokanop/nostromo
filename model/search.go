package model

import (
	"sort"

	"github.com/pokanop/nostromo/stringutil"
)

// SearchResult is a single command or substitution match in a manifest
type SearchResult struct {
	Manifest *Manifest
	Command  *Command
	// Sub is set when a substitution matched rather than the command itself
	Sub *Substitution
}

// FindCommands at exactly the given key path across all manifests
func (s *Spaceport) FindCommands(keyPath string) []*SearchResult {
	results := []*SearchResult{}
	for _, m := range s.Manifests() {
		if m == nil {
			continue
		}
		if cmd := m.Find(keyPath); cmd != nil {
			results = append(results, &SearchResult{Manifest: m, Command: cmd})
		}
	}
	return results
}

// Search all manifests for commands and substitutions matching query
//
// Results are ordered by manifest sequence and then key path.
func (s *Spaceport) Search(query string) ([]*SearchResult, []*SearchResult) {
	cmds := []*SearchResult{}
	subs := []*SearchResult{}
	for _, m := range s.Manifests() {
		if m == nil {
			continue
		}
		mc, ms := m.Search(query)
		cmds = append(cmds, mc...)
		subs = append(subs, ms...)
	}
	return cmds, subs
}

// Search this manifest for commands and substitutions matching query
//
// Commands match when query is a case insensitive substring of the name,
// alias or key path. Substitutions match on name or alias and are returned
// with the command they belong to. Results are ordered by key path.
func (m *Manifest) Search(query string) ([]*SearchResult, []*SearchResult) {
	cmds := []*SearchResult{}
	subs := []*SearchResult{}
	for _, cmd := range m.Commands {
		cmd.Walk(func(c *Command, stop *bool) {
			if c.matches(query) {
				cmds = append(cmds, &SearchResult{Manifest: m, Command: c})
			}
			for _, sub := range c.Subs {
				if sub.matches(query) {
					subs = append(subs, &SearchResult{Manifest: m, Command: c, Sub: sub})
				}
			}
		})
	}
	sortResults(cmds)
	sortResults(subs)
	return cmds, subs
}

func (c *Command) matches(query string) bool {
	return stringutil.ContainsCaseInsensitive(c.KeyPath, query) ||
		stringutil.ContainsCaseInsensitive(c.Alias, query) ||
		stringutil.ContainsCaseInsensitive(c.Name, query)
}

func (s *Substitution) matches(query string) bool {
	return stringutil.ContainsCaseInsensitive(s.Alias, query) ||
		stringutil.ContainsCaseInsensitive(s.Name, query)
}

func sortResults(results []*SearchResult) {
	sort.SliceStable(results, func(i, j int) bool {
		a, b := results[i], results[j]
		if a.Command.KeyPath != b.Command.KeyPath {
			return a.Command.KeyPath < b.Command.KeyPath
		}
		if a.Sub == nil || b.Sub == nil {
			return a.Sub == nil && b.Sub != nil
		}
		return a.Sub.Alias < b.Sub.Alias
	})
}
