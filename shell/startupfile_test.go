package shell

import (
	"github.com/pokanop/nostromo/model"
	"os"
	"reflect"
	"testing"
)

func TestStartupFile(t *testing.T) {
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
		{"malformed block 1", ".zshrc", "export PATH=/usr/local/bin\nexport FOO=bar\n\n# nostromo [section begin]\neval \"$(nostromo completion --zsh)\"\nalias foo='nostromo eval foo \"$*\"'\nalias bar='nostromo eval bar \"$*\"'", makeManifest("foo", "baz"), true, false, true, true, ""},
		{"malformed block 2", ".zshrc", "export PATH=/usr/local/bin\nexport FOO=bar\n\n# nostromo [section begin]\neval \"$(nostromo completion --zsh)\"\nalias foo='nostromo eval foo \"$*\"'\nalias bar='nostromo eval bar \"$*\"'# nostromo [section begin]", makeManifest("foo", "baz"), true, false, true, true, ""},
		{"empty profile", ".profile", "", model.NewManifest(), false, true, false, false, ""},
		{"empty bash_profile", ".bash_profile", "", model.NewManifest(), false, true, false, false, ""},
		{"empty bashrc", ".bashrc", "", model.NewManifest(), true, true, false, false, ""},
		{"empty zshrc", ".zshrc", "", model.NewManifest(), true, true, false, false, ""},
		{"existing non-preferred no commands", ".profile", "export PATH=/usr/local/bin\nexport FOO=bar", model.NewManifest(), false, true, false, false, "export PATH=/usr/local/bin\nexport FOO=bar"},
		{"existing preferred no commands", ".zshrc", "export PATH=/usr/local/bin\nexport FOO=bar", model.NewManifest(), true, true, false, false, "export PATH=/usr/local/bin\nexport FOO=bar"},
		{"existing non-preferred same commands", ".profile", "export PATH=/usr/local/bin\nexport FOO=bar\n\n# nostromo [section begin]\neval \"$(nostromo completion)\"\nfoo() { eval $(nostromo eval foo \"$*\") }\nbar() { eval $(nostromo eval bar \"$*\") }\n# nostromo [section end]", makeManifest("foo", "bar"), false, false, false, false, "export PATH=/usr/local/bin\nexport FOO=bar\n"},
		{"existing preferred same commands", ".zshrc", "export PATH=/usr/local/bin\nexport FOO=bar\n\n# nostromo [section begin]\neval \"$(nostromo completion --zsh)\"\nfoo() { eval $(nostromo eval foo \"$*\") }\nbar() { eval $(nostromo eval bar \"$*\") }\n# nostromo [section end]", makeManifest("foo", "bar"), true, false, false, false, "export PATH=/usr/local/bin\nexport FOO=bar\n\n# nostromo [section begin]\neval \"$(nostromo completion --zsh)\"\n\nbar() { eval $(nostromo eval bar \"$*\") }\nfoo() { eval $(nostromo eval foo \"$*\") }\n# nostromo [section end]\n"},
		{"existing non-preferred diff commands", ".profile", "export PATH=/usr/local/bin\nexport FOO=bar\n\n# nostromo [section begin]\neval \"$(nostromo completion)\"\nfoo() { eval $(nostromo eval foo \"$*\") }\nbar() { eval $(nostromo eval bar \"$*\") }\n# nostromo [section end]", makeManifest("baz"), false, false, false, false, "export PATH=/usr/local/bin\nexport FOO=bar\n"},
		{"existing preferred diff commands", ".zshrc", "export PATH=/usr/local/bin\nexport FOO=bar\n\n# nostromo [section begin]\neval \"$(nostromo completion --zsh)\"\nfoo() { eval $(nostromo eval foo \"$*\") }\nbar() { eval $(nostromo eval bar \"$*\") }\n# nostromo [section end]", makeManifest("foo", "baz"), true, false, false, false, "export PATH=/usr/local/bin\nexport FOO=bar\n\n# nostromo [section begin]\neval \"$(nostromo completion --zsh)\"\n\nbaz() { eval $(nostromo eval baz \"$*\") }\nfoo() { eval $(nostromo eval foo \"$*\") }\n# nostromo [section end]\n"},
		{"empty commands", ".zshrc", "export PATH=/usr/local/bin\nexport FOO=bar\n\n# nostromo [section begin]\neval \"$(nostromo completion --zsh)\"\nalias foo='nostromo eval foo \"$*\"'\nalias bar='nostromo eval bar \"$*\"'\n# nostromo [section end]", makeManifestLong(false, false, "foo", "bar", "baz", "qux"), true, false, false, false, "export PATH=/usr/local/bin\nexport FOO=bar\n\n# nostromo [section begin]\neval \"$(nostromo completion --zsh)\"\n\nbar() { eval $(nostromo eval bar \"$*\") }\nbaz() { eval $(nostromo eval baz \"$*\") }\nfoo() { eval $(nostromo eval foo \"$*\") }\nqux() { eval $(nostromo eval qux \"$*\") }\n# nostromo [section end]\n"},
		{"aliases only", ".zshrc", "export PATH=/usr/local/bin\nexport FOO=bar\n\n# nostromo [section begin]\neval \"$(nostromo completion --zsh)\"\nalias foo='nostromo eval foo \"$*\"'\nalias bar='nostromo eval bar \"$*\"'\n# nostromo [section end]", makeManifestLong(true, true, "baz", "qux"), true, false, false, false, "export PATH=/usr/local/bin\nexport FOO=bar\n\n# nostromo [section begin]\neval \"$(nostromo completion --zsh)\"\n\nalias baz='baz'\nalias qux='qux'\n# nostromo [section end]\n"},
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
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isPreferredFilename(test.filename); got != test.want {
				t.Errorf("isPreferredFilename() = %v, want %v", got, test.want)
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

func TestCurrentStartupFile(t *testing.T) {
	type args struct {
		env   string
		files []*startupFile
	}
	tests := []struct {
		name string
		args args
		want *startupFile
	}{
		{"nil files", args{"", nil}, nil},
		{"not startup file", args{"zsh", []*startupFile{makeStartupFileCommon(".foo", "", true)}}, nil},
		{"zsh", args{"zsh", []*startupFile{makeStartupFileCommon(".zshrc", "", true)}}, makeStartupFileCommon(".zshrc", "", true)},
		{"bash", args{"bash", []*startupFile{makeStartupFileCommon(".bashrc", "", true)}}, makeStartupFileCommon(".bashrc", "", true)},
	}
	for _, tt := range tests {
		os.Setenv("SHELL", tt.args.env)
		t.Run(tt.name, func(t *testing.T) {
			if got := currentStartupFile(tt.args.files); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("currentStartupFile() = %v, want %v", got, tt.want)
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
	m := model.NewManifest()
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
