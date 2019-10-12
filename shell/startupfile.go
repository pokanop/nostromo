package shell

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/pokanop/nostromo/model"
	"github.com/pokanop/nostromo/pathutil"
)

const (
	beginBlockComment = "# nostromo [section begin]"
	endBlockComment   = "# nostromo [section end]"
	sourceCompletion  = "eval \"$(nostromo completion%s)\""
)

var (
	startupFileNames = []string{".profile", ".bash_profile", ".bashrc", ".zshrc"}
)

type startupFile struct {
	path      string
	mode      os.FileMode
	content   string
	cmds      []*model.Command
	preferred bool
	pristine  bool
}

func loadStartupFiles() []*startupFile {
	files := []*startupFile{}
	for _, n := range startupFileNames {
		path, mode, err := findStartupFile(n)
		if err != nil {
			continue
		}

		s, err := parseStartupFile(path, mode)
		if err != nil {
			continue
		}

		files = append(files, s)
	}
	return files
}

func preferredStartupFiles(files []*startupFile) []*startupFile {
	p := []*startupFile{}
	for _, s := range files {
		if s.preferred {
			p = append(p, s)
		}
	}
	return p
}

func currentStartupFile(files []*startupFile) *startupFile {
	sh := Which()
	for _, s := range files {
		if sh == Zsh && strings.Contains(s.path, "zshrc") {
			return s
		} else if sh == Bash && strings.Contains(s.path, "bashrc") {
			return s
		}
	}
	return nil
}

func findStartupFile(name string) (string, os.FileMode, error) {
	home, err := pathutil.HomeDir()
	if err != nil {
		return "", 0, err
	}

	path := filepath.Join(home, name)
	info, err := os.Stat(path)
	if err != nil {
		return "", 0, err
	}

	return path, info.Mode(), nil
}

func parseStartupFile(path string, mode os.FileMode) (*startupFile, error) {
	f, err := os.Open(pathutil.Abs(path))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	s := newStartupFile(path, string(b), mode)
	err = s.parse()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func newStartupFile(path, content string, mode os.FileMode) *startupFile {
	return &startupFile{
		path:      path,
		mode:      mode,
		content:   content,
		cmds:      []*model.Command{},
		preferred: strings.Contains(path, ".bashrc") || strings.Contains(path, ".zshrc"),
	}
}

func (s *startupFile) parse() error {
	// Find the nostromo content block
	content, err := s.contentBlock()
	if err != nil {
		return err
	}

	re := regexp.MustCompile("alias (.+)='(.+)'")
	m := re.FindAllStringSubmatch(content, -1)
	if m == nil {
		// No matches
		s.pristine = true
		return nil
	}

	// Add existing aliases
	for _, a := range m {
		if len(a) < 3 {
			return fmt.Errorf("unable to find alias matches")
		}

		name := a[2]
		alias := a[1]
		aliasOnly := false
		if !strings.Contains(alias, "nostromo run") {
			aliasOnly = true
		}
		s.add(&model.Command{Name: name, Alias: a[1], AliasOnly: aliasOnly})
	}

	return nil
}

func (s *startupFile) reset() {
	s.cmds = []*model.Command{}
}

func (s *startupFile) add(cmd *model.Command) {
	s.cmds = append(s.cmds, cmd)
}

func (s *startupFile) commit() error {
	// No changes were made
	if s.pristine && !s.preferred {
		return nil
	}

	// Find the nostromo content block and remove
	content, err := s.contentOmitted()
	if err != nil {
		return err
	}

	// Add aliases
	content += s.makeAliasBlock()

	// Save a timestamped backup
	ts := time.Now().UTC().Format("20060102150405")
	backupPath := filepath.Join("/tmp", filepath.Base(s.path)) + "_" + ts
	err = ioutil.WriteFile(backupPath, []byte(s.content), s.mode)
	if err != nil {
		return err
	}

	// Save changes
	err = ioutil.WriteFile(pathutil.Abs(s.path), []byte(content), s.mode)
	if err != nil {
		return err
	}

	s.content = content
	return nil
}

func (s *startupFile) contentOmitted() (string, error) {
	start, end := s.contentIndexes()
	if start == -1 || end == -1 {
		// Malformed block
		return "", fmt.Errorf("malformed nostromo section found")
	}

	// Remove existing nostromo block
	return s.content[:start] + s.content[end:], nil
}

func (s *startupFile) contentBlock() (string, error) {
	start, end := s.contentIndexes()
	if start == -1 || end == -1 {
		// Malformed block
		return "", fmt.Errorf("malformed nostromo section found")
	}

	// Return existing nostromo block
	return s.content[start:end], nil
}

func (s *startupFile) contentIndexes() (int, int) {
	start := strings.Index(s.content, beginBlockComment)
	end := strings.Index(s.content, endBlockComment)
	if start == -1 || end == -1 {
		// Malformed block
		return start, end
	}

	// Return adjusted indexes
	start--
	end += len(endBlockComment) + 1
	return start, end
}

func (s *startupFile) makeAliasBlock() string {
	if len(s.cmds) == 0 {
		return ""
	}

	aliases := []string{}
	for _, c := range s.cmds {
		cmd := fmt.Sprintf("nostromo run %s \"$*\"", c.Alias)
		if c.AliasOnly {
			cmd = c.Name
		}
		alias := strings.TrimSpace(fmt.Sprintf("alias %s='%s'", c.Alias, cmd))
		aliases = append(aliases, alias)
	}
	zsh := ""
	if strings.Contains(s.path, "zsh") {
		zsh = " --zsh"
	}
	return fmt.Sprintf("\n%s\n%s\n\n%s\n%s\n", beginBlockComment, fmt.Sprintf(sourceCompletion, zsh), strings.Join(aliases, "\n"), endBlockComment)
}
