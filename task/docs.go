package task

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pokanop/nostromo/log"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

const docsNavFile = "SUMMARY.md"

// GenerateDocs writes a markdown page per visible command into dir along with
// a SUMMARY.md navigation file consumed by the docs site.
func GenerateDocs(cmd *cobra.Command, dir string) int {
	root := cmd.Root()
	root.DisableAutoGenTag = true

	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Error(err)
		return -1
	}

	if err := removeGeneratedDocs(root, dir); err != nil {
		log.Error(err)
		return -1
	}

	prepender := func(filename string) string {
		return fmt.Sprintf("---\ntitle: %s\n---\n\n", docTitle(root.Name(), filename))
	}
	linkHandler := func(name string) string {
		return name
	}
	if err := doc.GenMarkdownTreeCustom(root, dir, prepender, linkHandler); err != nil {
		log.Error(err)
		return -1
	}

	if err := promoteHeadings(root, dir); err != nil {
		log.Error(err)
		return -1
	}

	nav := docsNav(root)
	if err := os.WriteFile(filepath.Join(dir, docsNavFile), []byte(nav), 0644); err != nil {
		log.Error(err)
		return -1
	}

	log.Highlightf("generated command reference in %s\n", dir)
	return 0
}

// docTitle derives the page title from a generated file name, e.g.
// "nostromo_add_cmd.md" becomes "add cmd" and "nostromo.md" stays "nostromo".
func docTitle(rootName, filename string) string {
	base := strings.TrimSuffix(filepath.Base(filename), ".md")
	base = strings.TrimPrefix(base, rootName+"_")
	return strings.ReplaceAll(base, "_", " ")
}

// navTitle is the sidebar label for a command: its path without the root
// name, e.g. "add cmd".
func navTitle(cmd *cobra.Command) string {
	if !cmd.HasParent() {
		return cmd.Name()
	}
	return strings.TrimPrefix(cmd.CommandPath(), cmd.Root().Name()+" ")
}

// promoteHeadings rewrites each generated page so the command name is the H1
// and the sections cobra emits (Synopsis, Options, ...) are H2.
func promoteHeadings(root *cobra.Command, dir string) error {
	var cmds []*cobra.Command
	collectDocCommands(root, &cmds)
	for _, c := range cmds {
		path := filepath.Join(dir, docFilename(c))
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		page := strings.Replace(string(b), "## "+c.CommandPath()+"\n", "# "+c.CommandPath()+"\n", 1)
		page = strings.ReplaceAll(page, "\n### ", "\n## ")
		if err := os.WriteFile(path, []byte(page), 0644); err != nil {
			return err
		}
	}
	return nil
}

func collectDocCommands(cmd *cobra.Command, out *[]*cobra.Command) {
	if cmd.HasParent() && (!cmd.IsAvailableCommand() || cmd.IsAdditionalHelpTopicCommand()) {
		return
	}
	*out = append(*out, cmd)
	for _, c := range cmd.Commands() {
		collectDocCommands(c, out)
	}
}

// docFilename mirrors the file naming used by cobra/doc.
func docFilename(cmd *cobra.Command) string {
	return strings.ReplaceAll(cmd.CommandPath(), " ", "_") + ".md"
}

// docsNav renders a nested markdown list of all visible commands.
func docsNav(root *cobra.Command) string {
	var b strings.Builder
	fmt.Fprintf(&b, "* [%s](%s)\n", root.Name(), docFilename(root))
	for _, c := range root.Commands() {
		writeNavEntry(&b, c, 0)
	}
	return b.String()
}

func writeNavEntry(b *strings.Builder, cmd *cobra.Command, depth int) {
	if !cmd.IsAvailableCommand() || cmd.IsAdditionalHelpTopicCommand() {
		return
	}
	indent := strings.Repeat("    ", depth)
	fmt.Fprintf(b, "%s* [%s](%s)\n", indent, navTitle(cmd), docFilename(cmd))
	for _, c := range cmd.Commands() {
		writeNavEntry(b, c, depth+1)
	}
}

// removeGeneratedDocs deletes previously generated pages so commands that no
// longer exist do not linger in the reference.
func removeGeneratedDocs(root *cobra.Command, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	prefix := root.Name() + "_"
	for _, e := range entries {
		name := e.Name()
		generated := name == docsNavFile || name == root.Name()+".md" ||
			(strings.HasPrefix(name, prefix) && strings.HasSuffix(name, ".md"))
		if e.IsDir() || !generated {
			continue
		}
		if err := os.Remove(filepath.Join(dir, name)); err != nil {
			return err
		}
	}
	return nil
}
