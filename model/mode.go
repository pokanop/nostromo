package model

type Mode int

const (
	// ConcatenateMode will join this command to sub-commands and is the default. Use
	// this when combining the command nodes together to generate a single command string to
	// evaluate.
	ConcatenateMode Mode = iota

	// IndependentMode will run this command independently of others by suffixing the
	// execution string with a ';'. Note the parent commands may still be run depending on
	// the mode set on them. This is convenient for separating commands that should always run.
	IndependentMode

	// ExclusiveMode will run this command exclusively without walking up the command tree
	// to join or run other commands. This is similar to `IndependentMode` and will add a
	// ';' at the end of the execution string so sub-commands cannot be concatenated. However,
	// it means that `nostromo` will only run this command. Substitutions are still fully scoped.
	ExclusiveMode
)

var supportedModes = map[string]Mode{
	ConcatenateMode.String(): ConcatenateMode,
	IndependentMode.String(): IndependentMode,
	ExclusiveMode.String():   ExclusiveMode,
}

func (m Mode) String() string {
	switch m {
	case ConcatenateMode:
		return "concatenate"
	case IndependentMode:
		return "independent"
	case ExclusiveMode:
		return "exclusive"
	}
	return "unknown"
}

// IsModeSupported returns true if mode is supported and false otherwise.
func IsModeSupported(mode string) bool {
	_, ok := supportedModes[mode]
	return ok
}

// ModeFromString converts a string to a Mode and defaults to ConcatenateMode if not mappable.
func ModeFromString(mode string) Mode {
	m, ok := supportedModes[mode]
	if !ok {
		m = ConcatenateMode // default
	}
	return m
}

func SupportedModes() []string {
	var modes []string
	for _, mode := range supportedModes {
		modes = append(modes, mode.String())
	}
	return modes
}
