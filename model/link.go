package model

import (
	"fmt"
	"strings"

	"github.com/shivamMg/ppds/tree"
)

// LinkedManifest is a dependency record pointing at another manifest
//
// The UUID is the version identifier of the linked manifest at the time it
// was linked and is kept for future versioning support.
type LinkedManifest struct {
	UUID   string `json:"uuid"`
	Name   string `json:"name"`
	Source string `json:"source"`
}

// NewLinkedManifest returns a link record for the given manifest
func NewLinkedManifest(m *Manifest) *LinkedManifest {
	l := &LinkedManifest{Name: m.Name, Source: m.Source}
	if m.Version != nil {
		l.UUID = m.Version.UUID
	}
	return l
}

// Matches checks if the link refers to the given name or source
func (l *LinkedManifest) Matches(nameOrSource string) bool {
	return len(nameOrSource) > 0 && (l.Name == nameOrSource || l.Source == nameOrSource)
}

// Refers checks if the link points at the given manifest by name or source
func (l *LinkedManifest) Refers(m *Manifest) bool {
	if m == nil {
		return false
	}
	return l.Name == m.Name || (len(l.Source) > 0 && l.Source == m.Source)
}

// AddLink to the manifest, returns false if the source is already linked
func (m *Manifest) AddLink(l *LinkedManifest) bool {
	if l == nil {
		return false
	}
	for _, existing := range m.Links {
		if existing.Source == l.Source {
			return false
		}
	}
	m.Links = append(m.Links, l)
	return true
}

// RemoveLink matching name or source, returns the removed link or nil
func (m *Manifest) RemoveLink(nameOrSource string) *LinkedManifest {
	for i, l := range m.Links {
		if l.Matches(nameOrSource) {
			m.Links = append(m.Links[:i], m.Links[i+1:]...)
			return l
		}
	}
	return nil
}

// FindLink matching name or source or nil if missing
func (m *Manifest) FindLink(nameOrSource string) *LinkedManifest {
	for _, l := range m.Links {
		if l.Matches(nameOrSource) {
			return l
		}
	}
	return nil
}

// LinksTo checks if this manifest directly links the given manifest
func (m *Manifest) LinksTo(other *Manifest) bool {
	for _, l := range m.Links {
		if l.Refers(other) {
			return true
		}
	}
	return false
}

// Linkers returns the manifests that directly link the named manifest
func (s *Spaceport) Linkers(name string) []*Manifest {
	target := s.manifests[name]
	linkers := []*Manifest{}
	for _, m := range s.Manifests() {
		if m == nil || m.Name == name {
			continue
		}
		for _, l := range m.Links {
			if l.Name == name || (target != nil && l.Refers(target)) {
				linkers = append(linkers, m)
				break
			}
		}
	}
	return linkers
}

// Dock marks a manifest as explicitly docked by the user
func (s *Spaceport) Dock(name string) {
	if name == CoreManifestName || s.IsDocked(name) {
		return
	}
	s.Docked = append(s.Docked, name)
}

// Undock clears the explicitly docked mark for a manifest
func (s *Spaceport) Undock(name string) {
	for i, n := range s.Docked {
		if n == name {
			s.Docked = append(s.Docked[:i], s.Docked[i+1:]...)
			return
		}
	}
}

// IsDocked checks if a manifest was explicitly docked by the user
func (s *Spaceport) IsDocked(name string) bool {
	for _, n := range s.Docked {
		if n == name {
			return true
		}
	}
	return false
}

// Roots are the manifests that anchor the link graph: the core manifest and
// every explicitly docked manifest
func (s *Spaceport) Roots() []*Manifest {
	roots := []*Manifest{}
	if m := s.CoreManifest(); m != nil {
		roots = append(roots, m)
	}
	for _, name := range s.Docked {
		if m := s.manifests[name]; m != nil {
			roots = append(roots, m)
		}
	}
	return roots
}

// Resolve a link to a docked manifest by name or source, nil if missing
func (s *Spaceport) Resolve(l *LinkedManifest) *Manifest {
	if l == nil {
		return nil
	}
	if m := s.manifests[l.Name]; m != nil {
		return m
	}
	if len(l.Source) == 0 {
		return nil
	}
	for _, m := range s.Manifests() {
		if m != nil && m.Source == l.Source {
			return m
		}
	}
	return nil
}

// Reachable returns the names of manifests reachable from the roots via links
func (s *Spaceport) Reachable() map[string]bool {
	reachable := map[string]bool{}
	var walk func(m *Manifest)
	walk = func(m *Manifest) {
		if m == nil || reachable[m.Name] {
			return
		}
		reachable[m.Name] = true
		for _, l := range m.Links {
			walk(s.Resolve(l))
		}
	}
	for _, m := range s.Roots() {
		walk(m)
	}
	return reachable
}

// Orphans returns manifests that are neither docked nor linked from a root
func (s *Spaceport) Orphans() []*Manifest {
	reachable := s.Reachable()
	orphans := []*Manifest{}
	for _, m := range s.Manifests() {
		if m != nil && !reachable[m.Name] {
			orphans = append(orphans, m)
		}
	}
	return orphans
}

// reconcile the docked list with the loaded manifests
//
// Manifests that no longer exist are dropped and any manifest that cannot
// be reached through links is treated as explicitly docked, which also
// migrates spaceports from before links existed.
func (s *Spaceport) reconcile() {
	docked := []string{}
	for _, name := range s.Docked {
		if s.manifests[name] != nil && name != CoreManifestName {
			docked = append(docked, name)
		}
	}
	s.Docked = docked
	for _, m := range s.Orphans() {
		s.Dock(m.Name)
	}
}

// LinkNode is a printable node in the manifest dependency graph
type LinkNode struct {
	Label string
	Nodes []*LinkNode
}

// Data method for Node interface to print tree
func (n *LinkNode) Data() interface{} {
	return n.Label
}

// Children method for Node interface to print tree
func (n *LinkNode) Children() []tree.Node {
	nodes := make([]tree.Node, 0, len(n.Nodes))
	for _, c := range n.Nodes {
		nodes = append(nodes, c)
	}
	return nodes
}

// AsString renders the graph as a horizontal tree
func (n *LinkNode) AsString() string {
	return tree.SprintHr(n)
}

// LinkTree builds the dependency graph rooted at the given manifest
//
// Circular references are cut and labeled, as are links whose manifest is
// not docked.
func (s *Spaceport) LinkTree(m *Manifest) *LinkNode {
	return s.linkTree(m, []string{})
}

func (s *Spaceport) linkTree(m *Manifest, path []string) *LinkNode {
	label := m.Name
	if s.IsDocked(m.Name) {
		label += " (docked)"
	}
	node := &LinkNode{Label: label}
	path = append(path, m.Name)
	for _, l := range m.Links {
		target := s.Resolve(l)
		name := l.Name
		if target != nil {
			name = target.Name
		}
		label := fmt.Sprintf("%s <- %s", name, l.Source)
		if target == nil {
			node.Nodes = append(node.Nodes, &LinkNode{Label: label + " (missing)"})
			continue
		}
		if contains(path, target.Name) {
			node.Nodes = append(node.Nodes, &LinkNode{Label: label + " (circular)"})
			continue
		}
		child := s.linkTree(target, path)
		child.Label = label
		if s.IsDocked(target.Name) {
			child.Label += " (docked)"
		}
		node.Nodes = append(node.Nodes, child)
	}
	return node
}

// LinkGraph builds the dependency graph for all roots in the spaceport
func (s *Spaceport) LinkGraph() *LinkNode {
	root := &LinkNode{Label: DefaultSpaceportName}
	for _, m := range s.Roots() {
		root.Nodes = append(root.Nodes, s.LinkTree(m))
	}
	return root
}

func contains(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}

func joinedLinks(links []*LinkedManifest) string {
	names := []string{}
	for _, l := range links {
		names = append(names, l.Name)
	}
	return strings.Join(names, ", ")
}
