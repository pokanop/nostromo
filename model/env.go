package model

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/pokanop/nostromo/dotenv"
)

// EnvVar is an environment variable exported before a command runs
type EnvVar struct {
	Key   string
	Value string
	// Literal values are exported verbatim, others use shell syntax and are
	// expanded by the shell (e.g. "$HOME/bin:$PATH")
	Literal bool
	// Source is the key path or dotenv file the value came from
	Source string
}

// EnvChanges to apply to a command's environment
type EnvChanges struct {
	// Set adds or replaces env vars
	Set map[string]string
	// Unset removes env vars
	Unset []string
	// Dotenv replaces the dotenv files, nil keeps the existing list
	Dotenv []string
}

var envRefPattern = regexp.MustCompile(`\$\{?([A-Za-z_][A-Za-z0-9_]*)`)

// ParseEnv parses KEY=VALUE pairs as given on the command line
func ParseEnv(pairs []string) (map[string]string, error) {
	env := map[string]string{}
	for _, pair := range pairs {
		key, value, found := strings.Cut(pair, "=")
		if !found {
			return nil, fmt.Errorf("invalid env %q, expected KEY=VALUE", pair)
		}
		if !dotenv.ValidKey(key) {
			return nil, fmt.Errorf("invalid env key %q", key)
		}
		env[key] = value
	}
	return env, nil
}

// ApplyEnv changes to this command
func (c *Command) ApplyEnv(changes EnvChanges) {
	if len(changes.Set) > 0 && c.Env == nil {
		c.Env = map[string]string{}
	}
	for k, v := range changes.Set {
		c.Env[k] = v
	}
	for _, k := range changes.Unset {
		delete(c.Env, k)
	}
	if len(c.Env) == 0 {
		c.Env = nil
	}
	if changes.Dotenv != nil {
		c.Dotenv = changes.Dotenv
		if len(c.Dotenv) == 0 {
			c.Dotenv = nil
		}
	}
}

// Environment returns the effective environment for this command.
//
// Nodes are visited from the root down to this command, loading each node's
// dotenv files in order and then its env, so children override parents and
// env overrides dotenv on the same node. A variable keeps the position of its
// first definition so values referencing it still resolve. Dotenv files that
// cannot be read or parsed are skipped and returned as errors.
func (c *Command) Environment() ([]EnvVar, []error) {
	var nodes []*Command
	c.reverseWalk(func(cmd *Command, stop *bool) {
		nodes = append([]*Command{cmd}, nodes...)
	})

	var vars []EnvVar
	var errs []error
	index := map[string]int{}
	add := func(v EnvVar) {
		if i, ok := index[v.Key]; ok {
			vars[i] = v
			return
		}
		index[v.Key] = len(vars)
		vars = append(vars, v)
	}

	for _, node := range nodes {
		for _, path := range node.Dotenv {
			entries, err := dotenv.Load(path)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s: dotenv %w", node.KeyPath, err))
				continue
			}
			for _, e := range entries {
				add(EnvVar{Key: e.Key, Value: e.Value, Literal: e.Literal, Source: path})
			}
		}
		for _, k := range orderedEnvKeys(node.Env) {
			add(EnvVar{Key: k, Value: node.Env[k], Source: node.KeyPath})
		}
	}

	return vars, errs
}

// orderedEnvKeys sorts keys alphabetically, moving keys after any key of the
// same map their value references so `A: $B` is exported after `B`
func orderedEnvKeys(env map[string]string) []string {
	remaining := make([]string, 0, len(env))
	for k := range env {
		remaining = append(remaining, k)
	}
	sort.Strings(remaining)

	refs := map[string][]string{}
	for k, v := range env {
		for _, m := range envRefPattern.FindAllStringSubmatch(v, -1) {
			if ref := m[1]; ref != k && hasKey(env, ref) {
				refs[k] = append(refs[k], ref)
			}
		}
	}

	ordered := make([]string, 0, len(env))
	done := map[string]bool{}
	for len(remaining) > 0 {
		next := -1
		for i, k := range remaining {
			ready := true
			for _, ref := range refs[k] {
				if !done[ref] {
					ready = false
					break
				}
			}
			if ready {
				next = i
				break
			}
		}
		if next < 0 {
			// Cycle, fall back to alphabetical order
			next = 0
		}
		k := remaining[next]
		ordered = append(ordered, k)
		done[k] = true
		remaining = append(remaining[:next], remaining[next+1:]...)
	}
	return ordered
}

func hasKey(env map[string]string, key string) bool {
	_, ok := env[key]
	return ok
}

// envField joins env as "KEY=VALUE" pairs for logging
func (c *Command) envField() string {
	keys := make([]string, 0, len(c.Env))
	for k := range c.Env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, k+"="+c.Env[k])
	}
	return strings.Join(pairs, ", ")
}
