package shell

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pokanop/nostromo/model"
)

// Shells that can be targeted when rendering
var Shells = []string{Bash, Zsh, Fish, Powershell}

// IsSupported returns true if sh is a supported shell
func IsSupported(sh string) bool {
	for _, s := range Shells {
		if s == sh {
			return true
		}
	}
	return false
}

// Matches $NAME and ${NAME} references, group 1 or 2 holds the name
var varRefPattern = regexp.MustCompile(`\$(?:\{([A-Za-z_][A-Za-z0-9_]*)\}|([A-Za-z_][A-Za-z0-9_]*))`)

// Exports renders env vars as export statements on a single line separated
// by ";" so the result can be prepended to a command and evaluated by sh.
//
// Values are quoted for the target shell. Literal values are single quoted
// so nothing is expanded, other values are double quoted so references to
// other variables like $HOME or ${HOME} are expanded by the shell. Fish does
// not support ${NAME} so it is rewritten as "$NAME", and PowerShell keeps
// environment variables in a separate namespace so $NAME is rewritten as
// ${env:NAME} unless it already carries a scope like $env:NAME.
func Exports(sh string, vars []model.EnvVar) string {
	stmts := make([]string, 0, len(vars))
	for _, v := range vars {
		stmts = append(stmts, Export(sh, v))
	}
	return strings.Join(stmts, "; ")
}

// Export renders a single env var as an export statement for sh
func Export(sh string, v model.EnvVar) string {
	switch sh {
	case Fish:
		return fmt.Sprintf("set -gx %s %s", v.Key, quoteFish(v))
	case Powershell:
		return fmt.Sprintf("$env:%s = %s", v.Key, quotePowershell(v))
	default:
		return fmt.Sprintf("export %s=%s", v.Key, quotePosix(v))
	}
}

func quotePosix(v model.EnvVar) string {
	if v.Literal {
		return "'" + strings.ReplaceAll(v.Value, "'", `'\''`) + "'"
	}
	r := strings.NewReplacer(`"`, `\"`, "`", "\\`")
	return `"` + r.Replace(v.Value) + `"`
}

func quoteFish(v model.EnvVar) string {
	if v.Literal {
		r := strings.NewReplacer(`\`, `\\`, `'`, `\'`)
		return "'" + r.Replace(v.Value) + "'"
	}
	s := strings.ReplaceAll(v.Value, `"`, `\"`)
	s = replaceVarRefs(s, func(name string, braced bool) string {
		if braced {
			return `"$` + name + `"`
		}
		return "$" + name
	})
	return `"` + s + `"`
}

func quotePowershell(v model.EnvVar) string {
	if v.Literal {
		return "'" + strings.ReplaceAll(v.Value, "'", "''") + "'"
	}
	r := strings.NewReplacer("`", "``", `"`, "`\"")
	s := replaceVarRefs(r.Replace(v.Value), func(name string, braced bool) string {
		return "${env:" + name + "}"
	})
	return `"` + s + `"`
}

// replaceVarRefs rewrites $NAME and ${NAME} references using fn, skipping
// scoped references like $env:NAME
func replaceVarRefs(s string, fn func(name string, braced bool) string) string {
	var b strings.Builder
	last := 0
	for _, m := range varRefPattern.FindAllStringSubmatchIndex(s, -1) {
		start, end := m[0], m[1]
		braced := m[2] >= 0
		var name string
		if braced {
			name = s[m[2]:m[3]]
		} else {
			name = s[m[4]:m[5]]
			if end < len(s) && s[end] == ':' {
				continue
			}
		}
		b.WriteString(s[last:start])
		b.WriteString(fn(name, braced))
		last = end
	}
	b.WriteString(s[last:])
	return b.String()
}
