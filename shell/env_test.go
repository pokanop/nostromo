package shell

import (
	"testing"

	"github.com/pokanop/nostromo/model"
)

func TestIsSupported(t *testing.T) {
	for _, sh := range allShells {
		if !IsSupported(sh) {
			t.Errorf("IsSupported(%s) = false", sh)
		}
	}
	for _, sh := range []string{"", "sh", "tcsh", "Bash"} {
		if IsSupported(sh) {
			t.Errorf("IsSupported(%s) = true", sh)
		}
	}
}

func TestExport(t *testing.T) {
	tests := []struct {
		name string
		v    model.EnvVar
		want map[string]string
	}{
		{
			"plain",
			model.EnvVar{Key: "FOO", Value: "bar"},
			map[string]string{
				Bash:       `export FOO="bar"`,
				Zsh:        `export FOO="bar"`,
				Fish:       `set -gx FOO "bar"`,
				Powershell: `$env:FOO = "bar"`,
			},
		},
		{
			"empty",
			model.EnvVar{Key: "FOO"},
			map[string]string{
				Bash:       `export FOO=""`,
				Fish:       `set -gx FOO ""`,
				Powershell: `$env:FOO = ""`,
			},
		},
		{
			"spaces and specials",
			model.EnvVar{Key: "FOO", Value: `a b; c | d & e > f 'g' (h) *`},
			map[string]string{
				Bash:       `export FOO="a b; c | d & e > f 'g' (h) *"`,
				Fish:       `set -gx FOO "a b; c | d & e > f 'g' (h) *"`,
				Powershell: `$env:FOO = "a b; c | d & e > f 'g' (h) *"`,
			},
		},
		{
			"double quotes and backticks",
			model.EnvVar{Key: "FOO", Value: "say \"hi\" `now`"},
			map[string]string{
				Bash:       `export FOO="say \"hi\" \` + "`" + `now\` + "`" + `"`,
				Fish:       `set -gx FOO "say \"hi\" ` + "`now`" + `"`,
				Powershell: "$env:FOO = \"say `\"hi`\" ``now``\"",
			},
		},
		{
			"var reference",
			model.EnvVar{Key: "PATH", Value: "$HOME/bin:$PATH"},
			map[string]string{
				Bash:       `export PATH="$HOME/bin:$PATH"`,
				Zsh:        `export PATH="$HOME/bin:$PATH"`,
				Fish:       `set -gx PATH "$HOME/bin:$PATH"`,
				Powershell: `$env:PATH = "${env:HOME}/bin:${env:PATH}"`,
			},
		},
		{
			"braced var reference",
			model.EnvVar{Key: "FOO", Value: "pre${BAR}post"},
			map[string]string{
				Bash:       `export FOO="pre${BAR}post"`,
				Fish:       `set -gx FOO "pre"$BAR"post"`,
				Powershell: `$env:FOO = "pre${env:BAR}post"`,
			},
		},
		{
			"powershell scoped reference kept",
			model.EnvVar{Key: "FOO", Value: "$env:BAR/${env:BAZ}/$script:X"},
			map[string]string{
				Bash:       `export FOO="$env:BAR/${env:BAZ}/$script:X"`,
				Powershell: `$env:FOO = "$env:BAR/${env:BAZ}/$script:X"`,
			},
		},
		{
			"escaped dollar and bare dollar",
			model.EnvVar{Key: "FOO", Value: `\$5 costs $ and $(date) $1`},
			map[string]string{
				Bash:       `export FOO="\$5 costs $ and $(date) $1"`,
				Fish:       `set -gx FOO "\$5 costs $ and $(date) $1"`,
				Powershell: `$env:FOO = "\$5 costs $ and $(date) $1"`,
			},
		},
		{
			"literal",
			model.EnvVar{Key: "FOO", Value: `$HOME "quoted" it's \ back`, Literal: true},
			map[string]string{
				Bash:       `export FOO='$HOME "quoted" it'\''s \ back'`,
				Zsh:        `export FOO='$HOME "quoted" it'\''s \ back'`,
				Fish:       `set -gx FOO '$HOME "quoted" it\'s \\ back'`,
				Powershell: `$env:FOO = '$HOME "quoted" it''s \ back'`,
			},
		},
		{
			"multiline",
			model.EnvVar{Key: "FOO", Value: "line1\nline2"},
			map[string]string{
				Bash:       "export FOO=\"line1\nline2\"",
				Fish:       "set -gx FOO \"line1\nline2\"",
				Powershell: "$env:FOO = \"line1\nline2\"",
			},
		},
	}
	for _, tt := range tests {
		for sh, want := range tt.want {
			t.Run(tt.name+"/"+sh, func(t *testing.T) {
				if got := Export(sh, tt.v); got != want {
					t.Errorf("Export() = %s, want %s", got, want)
				}
			})
		}
	}
}

func TestExports(t *testing.T) {
	vars := []model.EnvVar{{Key: "A", Value: "1"}, {Key: "B", Value: "$A"}}
	tests := []struct {
		sh   string
		want string
	}{
		{Bash, `export A="1"; export B="$A"`},
		{Zsh, `export A="1"; export B="$A"`},
		{Fish, `set -gx A "1"; set -gx B "$A"`},
		{Powershell, `$env:A = "1"; $env:B = "${env:A}"`},
	}
	for _, tt := range tests {
		t.Run(tt.sh, func(t *testing.T) {
			if got := Exports(tt.sh, vars); got != tt.want {
				t.Errorf("Exports() = %s, want %s", got, tt.want)
			}
		})
	}
	if got := Exports(Bash, nil); got != "" {
		t.Errorf("Exports(nil) = %q, want empty", got)
	}
}

func TestEvalStringWithEnv(t *testing.T) {
	vars := []model.EnvVar{{Key: "A", Value: "1"}}
	tests := []struct {
		name     string
		sh       string
		command  string
		language string
		vars     []model.EnvVar
		want     string
		wantErr  bool
	}{
		{"no env", Bash, "echo foo", "", nil, "echo foo", false},
		{"bash", Bash, "echo $A", "", vars, `export A="1"; echo $A`, false},
		{"fish", Fish, "echo $A", "", vars, `set -gx A "1"; echo $A`, false},
		{"powershell", Powershell, "echo $env:A", "", vars, `$env:A = "1"; echo $env:A`, false},
		{"language", Bash, "print()", "python", vars, `export A="1"; python -c 'print()'`, false},
		{"empty command", Bash, "", "", vars, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EvalString(tt.sh, tt.command, tt.language, tt.vars, false)
			if (err != nil) != tt.wantErr {
				t.Fatalf("EvalString() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("EvalString() = %s, want %s", got, tt.want)
			}
		})
	}
}
