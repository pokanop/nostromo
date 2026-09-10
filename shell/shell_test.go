package shell

import (
	"reflect"
	"strings"
	"testing"

	"github.com/pokanop/nostromo/config"
	"github.com/pokanop/nostromo/model"
	"github.com/spf13/cobra"
)

func TestValidLanguages(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{"valid languages", []string{"sh", "ruby", "python", "perl", "js"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SupportedLanguages(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidLanguages() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsSupportedLanguage(t *testing.T) {
	type args struct {
		language string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"supported", args{"python"}, true},
		{"not supported", args{"jython"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSupportedLanguage(tt.args.language); got != tt.want {
				t.Errorf("IsSupportedLanguage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvalString(t *testing.T) {
	type args struct {
		command  string
		language string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"empty command", args{"", ""}, "", true},
		{"command with newline", args{"echo foo\n", ""}, "echo foo", false},
		{"no lang", args{"echo foo", ""}, "echo foo", false},
		{"node", args{"console.log(\"hello world\")", "js"}, "node -e 'console.log(\"hello world\")'", false},
		{"ruby", args{"puts \"hello world\"", "ruby"}, "ruby -e 'puts \"hello world\"'", false},
		{"python", args{"print()", "python"}, "python -c 'print()'", false},
		{"perl", args{"print \"hello world\";", "perl"}, "perl -e 'print \"hello world\";'", false},
		{"sh", args{"echo foo", "sh"}, "echo foo", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EvalString(tt.args.command, tt.args.language, true)
			if (err != nil) != tt.wantErr {
				t.Errorf("EvalString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("EvalString() got = %v, want %v", got, tt.want)
			}
		})
	}
}

var allShells = []string{Bash, Zsh, Fish, Powershell}

func TestShellWrapperFunc(t *testing.T) {
	tests := []struct {
		sh   string
		want string
	}{
		{Bash, "__nostromo_cmd() { command nostromo \"$@\"; }\nnostromo() { __nostromo_cmd \"$@\" && eval \"$(__nostromo_cmd completion bash)\"; }"},
		{Zsh, "__nostromo_cmd() { command nostromo \"$@\"; }\nnostromo() { __nostromo_cmd \"$@\" && eval \"$(__nostromo_cmd completion zsh)\"; }"},
		{Fish, "function __nostromo_cmd; command nostromo $argv; end\nfunction nostromo; __nostromo_cmd $argv; and __nostromo_cmd completion fish | source; end"},
		{Powershell, "function __nostromo_cmd { & (Get-Command nostromo -CommandType Application | Select-Object -First 1).Source @args }\nfunction nostromo { __nostromo_cmd @args; if ($?) { __nostromo_cmd completion powershell | Out-String | Invoke-Expression } }"},
	}
	for _, tt := range tests {
		t.Run(tt.sh, func(t *testing.T) {
			if got := shellWrapperFunc(tt.sh); got != tt.want {
				t.Errorf("shellWrapperFunc() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestShellAliasFuncs(t *testing.T) {
	tests := []struct {
		sh   string
		want string
	}{
		{Bash, "\nalias three='command'\none() { eval $(__nostromo_cmd eval one \"$@\"); }\ntwo() { eval $(__nostromo_cmd eval two \"$@\"); }\n"},
		{Zsh, "\nalias three='command'\none() { eval $(__nostromo_cmd eval one \"$@\"); }\ntwo() { eval $(__nostromo_cmd eval two \"$@\"); }\n"},
		{Fish, "\nalias three='command'\nfunction one; eval (__nostromo_cmd eval one $argv | string collect); end\nfunction two; eval (__nostromo_cmd eval two $argv | string collect); end\n"},
		{Powershell, "\nfunction one { Invoke-Expression (__nostromo_cmd eval one @args | Out-String) }\nfunction three { command @args }\nfunction two { Invoke-Expression (__nostromo_cmd eval two @args | Out-String) }\n"},
	}
	for _, tt := range tests {
		t.Run(tt.sh, func(t *testing.T) {
			if got := shellAliasFuncs(tt.sh, fakeManifest()); got != tt.want {
				t.Errorf("shellAliasFuncs() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestShellAliasFuncsSkipUnavailable(t *testing.T) {
	t.Setenv(model.PlatformEnv, "windows/amd64")
	tests := []struct {
		sh   string
		want string
	}{
		{Bash, "\nalias three='command'\none() { eval $(__nostromo_cmd eval one \"$@\"); }\n"},
		{Zsh, "\nalias three='command'\none() { eval $(__nostromo_cmd eval one \"$@\"); }\n"},
		{Fish, "\nalias three='command'\nfunction one; eval (__nostromo_cmd eval one $argv | string collect); end\n"},
		{Powershell, "\nfunction one { Invoke-Expression (__nostromo_cmd eval one @args | Out-String) }\nfunction three { command @args }\n"},
	}
	for _, tt := range tests {
		t.Run(tt.sh, func(t *testing.T) {
			m := fakeManifest()
			m.Find("two").Platforms = []string{"linux", "darwin"}
			m.Find("one.two.three").Platforms = []string{"windows"}
			if got := shellAliasFuncs(tt.sh, m); got != tt.want {
				t.Errorf("shellAliasFuncs() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestManifestCompletionSkipsUnavailable(t *testing.T) {
	t.Setenv(model.PlatformEnv, "windows/amd64")
	for _, sh := range allShells {
		t.Run(sh, func(t *testing.T) {
			m := fakeManifest()
			m.Find("two").Platforms = []string{"linux", "darwin"}
			m.Find("one.two").Platforms = []string{"linux"}
			completions, err := ManifestCompletion(sh, m)
			if err != nil {
				t.Fatalf("ManifestCompletion(%s) error: %v", sh, err)
			}
			joined := strings.Join(completions, "\n")
			if !strings.Contains(joined, "__nostromo_cmd __complete run one ") {
				t.Errorf("%s completion missing available command one", sh)
			}
			if strings.Contains(joined, "__nostromo_cmd __complete run two ") {
				t.Errorf("%s completion includes unavailable command two", sh)
			}
			cmd := m.Find("one").CobraCommand()
			if len(cmd.Commands()) != 0 {
				t.Errorf("CobraCommand() for one includes unavailable child two")
			}
		})
	}
}

func TestShellAliasFuncAliasOnly(t *testing.T) {
	c := &model.Command{Alias: "gs", Name: "git status", AliasOnly: true}
	tests := []struct {
		sh   string
		want string
	}{
		{Bash, "alias gs='git status'"},
		{Zsh, "alias gs='git status'"},
		{Fish, "alias gs='git status'"},
		{Powershell, "function gs { git status @args }"},
	}
	for _, tt := range tests {
		t.Run(tt.sh, func(t *testing.T) {
			if got := shellAliasFunc(tt.sh, c); got != tt.want {
				t.Errorf("shellAliasFunc() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCompletionRedirectsToNostromo(t *testing.T) {
	m := fakeManifest()
	for _, sh := range allShells {
		t.Run(sh, func(t *testing.T) {
			for _, cmd := range m.Commands {
				s, err := CommandCompletion(sh, cmd)
				if err != nil {
					t.Fatalf("CommandCompletion(%s, %s) error: %v", sh, cmd.Alias, err)
				}
				if strings.Contains(s, cmd.Alias+" __complete") {
					t.Errorf("%s completion for %s invokes the user command for completions", sh, cmd.Alias)
				}
				for _, bad := range []string{"${words[0]} __complete", "${words[1]} __complete", "$args[1] __complete", "$Program __complete"} {
					if strings.Contains(s, bad) {
						t.Errorf("%s completion for %s still contains %q", sh, cmd.Alias, bad)
					}
				}
				want := "__nostromo_cmd __complete run " + cmd.Alias + " "
				if !strings.Contains(s, want) {
					t.Errorf("%s completion for %s missing %q", sh, cmd.Alias, want)
				}
			}
		})
	}
}

func TestCompletionRootNotRedirected(t *testing.T) {
	root := &cobra.Command{Use: rootCommandName}
	root.AddCommand(&cobra.Command{Use: "run", Run: func(cmd *cobra.Command, args []string) {}})
	for _, sh := range allShells {
		t.Run(sh, func(t *testing.T) {
			s, err := Completion(sh, root)
			if err != nil {
				t.Fatalf("Completion(%s) error: %v", sh, err)
			}
			if s == "" {
				t.Fatalf("Completion(%s) returned empty script", sh)
			}
			if strings.Contains(s, "__nostromo_cmd __complete run") {
				t.Errorf("%s completion for nostromo itself should not be redirected to run", sh)
			}
			if strings.Contains(s, completionRequests[sh].find) {
				t.Errorf("%s completion for nostromo itself still invokes the wrapper function", sh)
			}
			if !strings.Contains(s, completionRequests[sh].root) {
				t.Errorf("%s completion for nostromo itself should invoke the binary directly", sh)
			}
		})
	}
}

func TestRedirectCompletionRequestErrors(t *testing.T) {
	if _, err := redirectCompletionRequest("tcsh", "foo", "anything"); err == nil {
		t.Errorf("expected error for unsupported shell")
	}
	if _, err := redirectCompletionRequest(Bash, "foo", "no request here"); err == nil {
		t.Errorf("expected error when request pattern is missing")
	}
}

func TestSpaceportCompletionPerShell(t *testing.T) {
	for _, sh := range allShells {
		t.Run(sh, func(t *testing.T) {
			completions, err := ManifestCompletion(sh, fakeManifest())
			if err != nil {
				t.Fatalf("ManifestCompletion(%s) error: %v", sh, err)
			}
			joined := strings.Join(completions, "\n")
			switch sh {
			case Fish:
				if strings.Contains(joined, "() {") {
					t.Errorf("fish output contains bash function syntax")
				}
			case Powershell:
				if strings.Contains(joined, "() {") || strings.Contains(joined, "alias ") {
					t.Errorf("powershell output contains bash syntax")
				}
			}
		})
	}
}

func fakeManifest() *model.Manifest {
	m, _ := config.NewCoreManifest()
	m.AddCommand("one.two.three", "command", "", &model.Code{}, false, "concatenate", nil)
	m.AddSubstitution("one.two", "name", "alias")
	m.AddCommand("two", "command", "", &model.Code{}, false, "concatenate", nil)
	m.AddCommand("three", "command", "", &model.Code{}, true, "concatenate", nil)
	return m
}
