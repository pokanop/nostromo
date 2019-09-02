package shell

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/pokanop/nostromo/pathutil"
)

const (
	beginBlockComment = "# nostromo [section begin]"
	endBlockComment   = "# nostromo [section end]"
)

var (
	startupFileNames = []string{".profile", ".bash_profile", ".bashrc"}
)

type startupFile struct {
	path      string
	mode      os.FileMode
	content   string
	aliases   []string
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

func preferredStartupFile(files []*startupFile) *startupFile {
	if len(files) == 0 {
		return nil
	}
	for _, s := range files {
		if s.preferred {
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
		aliases:   []string{},
		preferred: strings.Contains(path, ".bashrc"),
	}
}

func (s *startupFile) parse() error {
	re := regexp.MustCompile("alias (.+)='nostromo run .+")
	m := re.FindAllStringSubmatch(s.content, -1)
	if m == nil {
		// No matches
		s.pristine = true
		return nil
	}

	// Add existing aliases
	for _, a := range m {
		if len(a) < 2 {
			return fmt.Errorf("unable to find alias matches")
		}
		s.add(a[1])
	}

	return nil
}

func (s *startupFile) reset() {
	s.aliases = []string{}
}

func (s *startupFile) add(alias string) {
	s.aliases = append(s.aliases, alias)
}

func (s *startupFile) commit() error {
	// No changes were made
	if s.pristine && !s.preferred {
		return nil
	}

	start := strings.Index(s.content, beginBlockComment)
	end := strings.Index(s.content, endBlockComment)
	if (start == -1 && end != -1) || (start != -1 && end == -1) {
		// Malformed block
		return fmt.Errorf("malformed nostromo section found")
	}

	content := s.content
	if start != -1 && end != -1 {
		// Remove existing nostromo block
		start--
		end += len(endBlockComment) + 1
		content = content[:start] + content[end:]
	}

	// Add aliases
	content += s.makeAliasBlock()

	// Save a timestamped backup
	ts := time.Now().UTC().Format("20060102150405")
	backupPath := filepath.Join("/tmp", filepath.Base(s.path)) + "_" + ts
	err := ioutil.WriteFile(backupPath, []byte(s.content), s.mode)
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

func (s *startupFile) makeAliasBlock() string {
	if len(s.aliases) == 0 {
		return ""
	}

	aliases := []string{}
	for _, a := range s.aliases {
		alias := fmt.Sprintf("alias %s='nostromo run %s \"$*\"'", a, a)
		aliases = append(aliases, alias)
	}
	return fmt.Sprintf("\n%s\n%s\n%s\n", beginBlockComment, strings.Join(aliases, "\n"), endBlockComment)
}
