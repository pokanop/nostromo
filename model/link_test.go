package model

import (
	"reflect"
	"strings"
	"testing"

	"github.com/pokanop/nostromo/version"
	"gopkg.in/yaml.v2"
)

func linkedManifest(name string, links ...string) *Manifest {
	m := NewManifest(name, "https://example.com/"+name+".yaml", "", &version.Info{UUID: name + "-uuid"})
	for _, l := range links {
		m.Links = append(m.Links, &LinkedManifest{UUID: l + "-uuid", Name: l, Source: "https://example.com/" + l + ".yaml"})
	}
	return m
}

func linkedSpaceport(docked []string, manifests ...*Manifest) *Spaceport {
	s := &Spaceport{Docked: docked}
	s.Init()
	s.Import(manifests)
	return s
}

func names(manifests []*Manifest) []string {
	n := []string{}
	for _, m := range manifests {
		n = append(n, m.Name)
	}
	return n
}

func TestManifestLinks(t *testing.T) {
	m := linkedManifest("manifest")
	l := NewLinkedManifest(linkedManifest("tools"))
	if l.UUID != "tools-uuid" || l.Name != "tools" || l.Source != "https://example.com/tools.yaml" {
		t.Errorf("unexpected link record: %+v", l)
	}

	if !m.AddLink(l) {
		t.Errorf("want link added")
	}
	if m.AddLink(&LinkedManifest{Name: "other", Source: l.Source}) {
		t.Errorf("want duplicate source rejected")
	}
	if m.AddLink(nil) {
		t.Errorf("want nil link rejected")
	}
	if len(m.Links) != 1 {
		t.Errorf("want 1 link, got %d", len(m.Links))
	}

	if m.FindLink("tools") != l || m.FindLink(l.Source) != l || m.FindLink("") != nil || m.FindLink("missing") != nil {
		t.Errorf("unexpected FindLink results")
	}
	if !m.LinksTo(linkedManifest("tools")) || m.LinksTo(linkedManifest("other")) || m.LinksTo(nil) {
		t.Errorf("unexpected LinksTo results")
	}

	if m.RemoveLink("missing") != nil {
		t.Errorf("want nil when removing missing link")
	}
	if m.RemoveLink("tools") != l || len(m.Links) != 0 {
		t.Errorf("want link removed")
	}
}

func TestManifestLinksYAML(t *testing.T) {
	m := linkedManifest("manifest", "tools")
	out := m.AsYAML()
	if !strings.Contains(out, "links:\n- uuid: tools-uuid\n  name: tools\n  source: https://example.com/tools.yaml") {
		t.Errorf("unexpected yaml:\n%s", out)
	}

	parsed := &Manifest{}
	if err := yaml.Unmarshal([]byte(out), parsed); err != nil {
		t.Fatalf("unable to parse yaml: %s", err)
	}
	if !reflect.DeepEqual(parsed.Links, m.Links) {
		t.Errorf("links = %+v, want %+v", parsed.Links, m.Links)
	}

	// Manifests without links must not write an empty block
	if out := linkedManifest("plain").AsYAML(); strings.Contains(out, "links") {
		t.Errorf("want links omitted, got:\n%s", out)
	}
}

func TestSpaceportDocked(t *testing.T) {
	s := linkedSpaceport(nil, linkedManifest(CoreManifestName), linkedManifest("tools"))
	// Unreachable manifests are treated as docked when imported
	if !s.IsDocked("tools") {
		t.Errorf("want tools docked after import")
	}
	s.Dock(CoreManifestName)
	s.Dock("tools")
	if !reflect.DeepEqual(s.Docked, []string{"tools"}) {
		t.Errorf("docked = %v, want [tools]", s.Docked)
	}
	s.Undock("tools")
	if s.IsDocked("tools") {
		t.Errorf("want tools undocked")
	}
	s.Dock("tools")
	s.RemoveManifest("tools")
	if s.IsDocked("tools") || s.FindManifest("tools") != nil {
		t.Errorf("want tools removed")
	}
	if !reflect.DeepEqual(s.Sequence, []string{CoreManifestName}) {
		t.Errorf("sequence = %v", s.Sequence)
	}
}

func TestSpaceportReconcile(t *testing.T) {
	tests := []struct {
		name      string
		docked    []string
		manifests []*Manifest
		want      []string
	}{
		{"legacy spaceport marks everything docked", nil, []*Manifest{linkedManifest(CoreManifestName), linkedManifest("a"), linkedManifest("b")}, []string{"a", "b"}},
		{"linked manifests stay linked", nil, []*Manifest{linkedManifest(CoreManifestName, "a"), linkedManifest("a", "b"), linkedManifest("b")}, []string{}},
		{"missing docked dropped", []string{"gone", CoreManifestName}, []*Manifest{linkedManifest(CoreManifestName)}, []string{}},
		{"unreachable cycle marked docked", nil, []*Manifest{linkedManifest(CoreManifestName), linkedManifest("a", "b"), linkedManifest("b", "a")}, []string{"a", "b"}},
		{"docked kept", []string{"a"}, []*Manifest{linkedManifest(CoreManifestName, "a"), linkedManifest("a")}, []string{"a"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := linkedSpaceport(tt.docked, tt.manifests...)
			if !reflect.DeepEqual(s.Docked, tt.want) {
				t.Errorf("docked = %v, want %v", s.Docked, tt.want)
			}
		})
	}
}

func TestSpaceportOrphans(t *testing.T) {
	core := linkedManifest(CoreManifestName, "a")
	s := linkedSpaceport(nil, core, linkedManifest("a", "b"), linkedManifest("b", "a"), linkedManifest("c"))
	if got := names(s.Orphans()); len(got) != 0 {
		t.Fatalf("want no orphans after import, got %v", got)
	}

	// Dropping the link from core orphans the a <-> b cycle but not docked c
	core.Links = nil
	if got := names(s.Orphans()); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Errorf("orphans = %v, want [a b]", got)
	}
	if got := s.Reachable(); !got[CoreManifestName] || !got["c"] || got["a"] || got["b"] {
		t.Errorf("unexpected reachable set %v", got)
	}
}

func TestSpaceportLinkers(t *testing.T) {
	bySource := linkedManifest("renamed")
	bySource.Links = []*LinkedManifest{{Name: "old-name", Source: "https://example.com/b.yaml"}}
	s := linkedSpaceport(nil, linkedManifest(CoreManifestName, "a"), linkedManifest("a", "b"), linkedManifest("b"), bySource)

	tests := []struct {
		name string
		want []string
	}{
		{"a", []string{CoreManifestName}},
		{"b", []string{"a", "renamed"}},
		{CoreManifestName, []string{}},
		{"missing", []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := names(s.Linkers(tt.name)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("linkers = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSpaceportResolve(t *testing.T) {
	s := linkedSpaceport(nil, linkedManifest(CoreManifestName), linkedManifest("a"))
	if s.Resolve(nil) != nil {
		t.Errorf("want nil for nil link")
	}
	if m := s.Resolve(&LinkedManifest{Name: "a"}); m == nil || m.Name != "a" {
		t.Errorf("want resolve by name")
	}
	if m := s.Resolve(&LinkedManifest{Name: "other", Source: "https://example.com/a.yaml"}); m == nil || m.Name != "a" {
		t.Errorf("want resolve by source")
	}
	if s.Resolve(&LinkedManifest{Name: "other"}) != nil {
		t.Errorf("want nil for unknown link")
	}
}

func TestSpaceportLinkTree(t *testing.T) {
	s := linkedSpaceport([]string{"c"},
		linkedManifest(CoreManifestName, "a", "missing"),
		linkedManifest("a", "b"),
		linkedManifest("b", "a", "c"),
		linkedManifest("c"),
	)

	tree := s.LinkTree(s.CoreManifest())
	want := &LinkNode{Label: CoreManifestName, Nodes: []*LinkNode{
		{Label: "a <- https://example.com/a.yaml", Nodes: []*LinkNode{
			{Label: "b <- https://example.com/b.yaml", Nodes: []*LinkNode{
				{Label: "a <- https://example.com/a.yaml (circular)"},
				{Label: "c <- https://example.com/c.yaml (docked)"},
			}},
		}},
		{Label: "missing <- https://example.com/missing.yaml (missing)"},
	}}
	if !reflect.DeepEqual(tree, want) {
		t.Errorf("tree = %s, want %s", tree.AsString(), want.AsString())
	}

	graph := s.LinkGraph()
	if graph.Label != DefaultSpaceportName || len(graph.Nodes) != 2 || graph.Nodes[1].Label != "c (docked)" {
		t.Errorf("unexpected graph: %s", graph.AsString())
	}
	if tree.Data() != CoreManifestName || len(tree.Children()) != 2 {
		t.Errorf("unexpected node interface results")
	}
}
