package shell

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/pokanop/nostromo/config"
	"github.com/pokanop/nostromo/model"
)

func TestStartupFile(t *testing.T) {
	man := func() *model.Manifest {
		m, err := config.NewCoreManifest()
		if err != nil {
			panic(err)
		}
		return m
	}

	tests := []struct {
		name        string
		path        string
		content     string
		manifest    *model.Manifest
		preferred   bool
		pristine    bool
		expParseErr bool
		expApplyErr bool
		expContent  string
	}{
		{"nil manifest", ".profile", "", nil, false, true, false, true, ""},
		{"malformed block 1", ".zshrc", "export PATH=/usr/local/bin\nexport FOO=bar\n\n# nostromo [section begin]\neval \"$(nostromo completion)\"\nalias foo='nostromo eval foo \"$*\"'\nalias bar='nostromo eval bar \"$*\"'", makeManifest("foo", "baz"), true, false, true, true, ""},
		{"malformed block 2", ".zshrc", "export PATH=/usr/local/bin\nexport FOO=bar\n\n# nostromo [section begin]\neval \"$(nostromo completion)\"\nalias foo='nostromo eval foo \"$*\"'\nalias bar='nostromo eval bar \"$*\"'# nostromo [section begin]", makeManifest("foo", "baz"), true, false, true, true, ""},
		{"empty profile", ".profile", "", man(), false, true, false, false, ""},
		{"empty bash_profile", ".bash_profile", "", man(), false, true, false, false, ""},
		{"empty bashrc", ".bashrc", "", man(), true, true, false, false, "\n# nostromo [section begin]\nsource <(nostromo completion bash)\n# nostromo [section end]\n"},
		{"empty zshrc", ".zshrc", "", man(), true, true, false, false, "\n# nostromo [section begin]\nautoload -U compinit; compinit\nsource <(nostromo completion zsh)\n# nostromo [section end]\n"},
		{"existing non-preferred no commands", ".profile", "export PATH=/usr/local/bin\nexport FOO=bar", man(), false, true, false, false, "export PATH=/usr/local/bin\nexport FOO=bar"},
		{"existing preferred no commands", ".zshrc", "export PATH=/usr/local/bin\nexport FOO=bar", man(), true, true, false, false, "export PATH=/usr/local/bin\nexport FOO=bar\n# nostromo [section begin]\nautoload -U compinit; compinit\nsource <(nostromo completion zsh)\n# nostromo [section end]\n"},
		{"empty fish config", "/home/user/.config/fish/config.fish", "", man(), true, true, false, false, "\n# nostromo [section begin]\nnostromo completion fish | source\n# nostromo [section end]\n"},
		{"empty powershell profile", "/home/user/.config/powershell/Microsoft.PowerShell_profile.ps1", "", man(), true, true, false, false, "\n# nostromo [section begin]\nnostromo completion powershell | Out-String | Invoke-Expression\n# nostromo [section end]\n"},
		{"existing fish config with block", "config.fish", "set -x FOO bar\n# nostromo [section begin]\nold\n# nostromo [section end]\n", man(), true, false, false, false, "set -x FOO bar\n# nostromo [section begin]\nnostromo completion fish | source\n# nostromo [section end]\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f := newStartupFile(test.path, test.content, os.ModeAppend)

			err := f.parse()
			if err == nil && test.expParseErr {
				t.Errorf("expected parse error but got none")
			} else if err != nil && !test.expParseErr {
				t.Errorf("expected no parse error but got: %s", err)
			}

			if f.pristine != test.pristine {
				t.Errorf("pristine mismatch, expected: %t, actual: %t", test.pristine, f.pristine)
			}
			if f.preferred != test.preferred {
				t.Errorf("preferred mismatch, expected: %t, actual: %t", test.preferred, f.preferred)
			}

			err = f.apply(test.manifest)
			if err == nil && test.expApplyErr {
				t.Errorf("expected apply error but got none")
			} else if err != nil && !test.expApplyErr {
				t.Errorf("expected no apply error but got: %s", err)
			}

			if f.updatedContent != test.expContent {
				t.Errorf("expected content '%s' but got '%s'", test.expContent, f.updatedContent)
			}
		})
	}
}

func TestIsPreferredFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{"empty string", "", false},
		{"profile", ".profile", false},
		{"bash_profile", ".bash_profile", false},
		{"bashrc", ".bashrc", true},
		{"zshrc", ".zshrc", true},
		{"substring 1", "/path/to/.zshrc", true},
		{"substring 2", "~/.zshrc", true},
		{"fish config", "/home/user/.config/fish/config.fish", true},
		{"powershell profile", "/home/user/.config/powershell/Microsoft.PowerShell_profile.ps1", true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isPreferredFilename(test.filename); got != test.want {
				t.Errorf("isPreferredFilename() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestStartupFileShell(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/home/user/.bashrc", Bash},
		{"/home/user/.profile", Bash},
		{"/home/user/.zshrc", Zsh},
		{"/home/user/.config/fish/config.fish", Fish},
		{"/home/user/.config/powershell/Microsoft.PowerShell_profile.ps1", Powershell},
		{"C:\\Users\\user\\Documents\\PowerShell\\Microsoft.PowerShell_profile.ps1", Powershell},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			s := newStartupFile(tt.path, "", os.ModeAppend)
			if got := s.shell(); got != tt.want {
				t.Errorf("shell() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStartupFilePath(t *testing.T) {
	home := filepath.Join("home", "user")

	os.Unsetenv("XDG_CONFIG_HOME")
	os.Unsetenv("ZDOTDIR")

	pwshDir := filepath.Join(home, ".config", "powershell")
	if runtime.GOOS == "windows" {
		pwshDir = filepath.Join(home, "Documents", "PowerShell")
	}

	tests := []struct {
		name string
		env  map[string]string
		file string
		want string
	}{
		{"bashrc", nil, ".bashrc", filepath.Join(home, ".bashrc")},
		{"zshrc", nil, ".zshrc", filepath.Join(home, ".zshrc")},
		{"zshrc with ZDOTDIR", map[string]string{"ZDOTDIR": filepath.Join("zdot")}, ".zshrc", filepath.Join("zdot", ".zshrc")},
		{"bashrc ignores ZDOTDIR", map[string]string{"ZDOTDIR": filepath.Join("zdot")}, ".bashrc", filepath.Join(home, ".bashrc")},
		{"fish default", nil, fishConfigFilename, filepath.Join(home, ".config", "fish", "config.fish")},
		{"fish XDG_CONFIG_HOME", map[string]string{"XDG_CONFIG_HOME": filepath.Join("xdg")}, fishConfigFilename, filepath.Join("xdg", "fish", "config.fish")},
		{"powershell", nil, pwshProfileFilename, filepath.Join(pwshDir, pwshProfileFilename)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}
			if got := startupFilePath(home, tt.file); got != tt.want {
				t.Errorf("startupFilePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPreferredStartupFiles(t *testing.T) {
	tests := []struct {
		name  string
		files []*startupFile
		want  []*startupFile
	}{
		{"nil list", nil, nil},
		{"empty list", []*startupFile{}, nil},
		{"no preferred", []*startupFile{makeStartupFile(false)}, nil},
		{"only preferred", []*startupFile{makeStartupFile(true), makeStartupFile(true)}, []*startupFile{makeStartupFile(true), makeStartupFile(true)}},
		{"mix preferred", []*startupFile{makeStartupFile(true), makeStartupFile(false), makeStartupFile(true)}, []*startupFile{makeStartupFile(true), makeStartupFile(true)}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := preferredStartupFiles(test.files); !reflect.DeepEqual(got, test.want) {
				t.Errorf("preferredStartupFiles() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestStartupFileCanCommit(t *testing.T) {
	type fields struct {
		updatedContent string
		preferred      bool
		pristine       bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"pristine preferred updated", fields{"foo", true, true}, true},
		{"pristine not preferred updated", fields{"foo", false, true}, false},
		{"pristine not preferred not updated", fields{"", false, true}, false},
		{"pristine preferred not updated", fields{"", true, true}, false},
		{"not pristine preferred updated", fields{"foo", true, false}, true},
		{"not pristine not preferred updated", fields{"foo", false, false}, true},
		{"not pristine not preferred not updated", fields{"", false, false}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &startupFile{
				updatedContent: tt.fields.updatedContent,
				preferred:      tt.fields.preferred,
				pristine:       tt.fields.pristine,
			}
			if got := s.canCommit(); got != tt.want {
				t.Errorf("canCommit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func makeManifest(cmds ...string) *model.Manifest {
	return makeManifestLong(true, false, cmds...)
}

func makeManifestLong(match bool, aliasOnly bool, cmds ...string) *model.Manifest {
	m, err := config.NewCoreManifest()
	if err != nil {
		panic(err)
	}
	for _, cmd := range cmds {
		alias := cmd
		if !match {
			cmd = ""
		}
		m.AddCommand(alias, cmd, "", nil, aliasOnly, "concatenate")
	}
	return m
}

func makeStartupFile(preferred bool) *startupFile {
	return makeStartupFileCommon("path", "", preferred)
}

func makeStartupFileCommon(name, content string, preferred bool) *startupFile {
	s := newStartupFile(name, content, os.ModeAppend)
	s.preferred = preferred
	return s
}
