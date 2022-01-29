package shell

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pokanop/nostromo/log"
	"github.com/pokanop/nostromo/model"
	"github.com/pokanop/nostromo/pathutil"
)

const (
	beginBlockComment    = "# nostromo [section begin]"
	endBlockComment      = "# nostromo [section end]"
	bashSourceCompletion = "source <(nostromo completion bash)"
	zshSourceCompletion  = "autoload -U compinit; compinit\nsource <(nostromo completion zsh)"
)

var (
	startupFilenames   = []string{".profile", ".bash_profile", ".bashrc", ".zshrc"}
	preferredFilenames = []string{".bashrc", ".zshrc"}
)

type startupFile struct {
	path           string
	mode           os.FileMode
	content        string
	updatedContent string
	commands       map[string]*model.Command
	preferred      bool
	pristine       bool
}

func isPreferredFilename(filename string) bool {
	for _, preferredFilename := range preferredFilenames {
		if strings.Contains(filename, preferredFilename) {
			return true
		}
	}
	return false
}

func loadStartupFiles() []*startupFile {
	var files []*startupFile
	for _, n := range startupFilenames {
		path, mode, err := findStartupFile(n)
		if err != nil {
			log.Debugf("could not find %s: %s\n", n, err)
			continue
		}

		s, err := parseStartupFile(path, mode)
		if err != nil {
			log.Debugf("could not parse %s: %s\n", n, err)
			continue
		}

		files = append(files, s)
	}
	return files
}

func preferredStartupFiles(files []*startupFile) []*startupFile {
	var p []*startupFile
	for _, s := range files {
		if s.preferred {
			p = append(p, s)
		}
	}
	return p
}

func findStartupFile(name string) (string, os.FileMode, error) {
	home, err := pathutil.HomeDir()
	if err != nil {
		return "", 0, err
	}

	// zsh doesn't always have a ~/.zshrc file, and if it doesn't,
	// it does have a $ZDOTDIR/.zshrc
	// https://wiki.archlinux.org/index.php/Zsh#Startup.2FShutdown_files
	zdotDir := os.Getenv("ZDOTDIR")
	var path string
	if zdotDir != "" {
		path = filepath.Join(zdotDir, name)
	} else {
		path = filepath.Join(home, name)
	}

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
		commands:  map[string]*model.Command{},
		preferred: isPreferredFilename(path),
	}
}

func (s *startupFile) name() string {
	return filepath.Base(s.path)
}

func (s *startupFile) shell() string {
	if strings.Contains(s.path, ".zshrc") {
		return Zsh
	}
	return Bash
}

func (s *startupFile) parse() error {
	// Find the nostromo content block
	content, err := s.contentBlock()
	if err != nil {
		return err
	}

	// No existing content block
	if content == "" {
		s.pristine = true
	}

	return nil
}

func (s *startupFile) apply(manifest *model.Manifest) error {
	if manifest == nil {
		return fmt.Errorf("manifest must not be nil")
	}

	// Since nostromo works by aliasing only the top level commands,
	// iterate the manifest's list and update.
	for _, cmd := range manifest.Commands {
		s.commands[cmd.Alias] = cmd
	}

	// Find the nostromo content block and remove
	content, err := s.contentOmitted()
	if err != nil {
		return err
	}

	// Add aliases to preferred init files only
	if s.preferred {
		content += s.makeNostromoBlock()
	}

	s.updatedContent = content

	return nil
}

func (s *startupFile) canCommit() bool {
	return !s.pristine || (s.preferred && len(s.updatedContent) > 0)
}

func (s *startupFile) commit() error {
	// Only update preferred init files, clean up other files if possible
	if !s.canCommit() {
		return fmt.Errorf("commit now allowed")
	}

	// Save a timestamped backup
	ts := time.Now().UTC().Format("20060102150405")
	backupPath := filepath.Join("/tmp", filepath.Base(s.path)) + "_" + ts
	err := ioutil.WriteFile(backupPath, []byte(s.content), s.mode)
	if err != nil {
		return err
	}

	// Save changes
	err = ioutil.WriteFile(pathutil.Abs(s.path), []byte(s.updatedContent), s.mode)
	if err != nil {
		return err
	}

	return nil
}

func (s *startupFile) contentOmitted() (string, error) {
	start, end := s.contentIndexes()
	if start == -1 && end == -1 {
		// No nostromo block, return content as is
		return s.content, nil
	}

	if start == -1 || end == -1 {
		// Malformed block
		return "", fmt.Errorf("malformed nostromo section found")
	}

	// Remove existing nostromo block
	return s.content[:start] + s.content[end:], nil
}

func (s *startupFile) contentBlock() (string, error) {
	start, end := s.contentIndexes()
	if start == -1 && end == -1 {
		// No content block
		return "", nil
	}

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

	// Adjust if no newline at the end
	if end > len(s.content) {
		end = len(s.content)
	}

	return start, end
}

func (s *startupFile) makeNostromoBlock() string {
	sourceCompletion := bashSourceCompletion
	if s.shell() == Zsh {
		sourceCompletion = zshSourceCompletion
	}
	return fmt.Sprintf("\n%s\n%s\n%s\n", beginBlockComment, sourceCompletion, endBlockComment)
}
