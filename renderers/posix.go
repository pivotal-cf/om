package renderers

import (
	"fmt"
	"strings"
)

type posix struct {
}

// NewPosix defines a new posix renderer
func NewPosix() Renderer {
	return &posix{}
}

func (renderer *posix) RenderEnvironmentVariable(variable string, value string) string {
	if strings.ContainsAny(value, "\n") && !strings.HasSuffix(value, "\n") {
		value += "\n"
	}
	return fmt.Sprintf("export %s=%s", variable, posixQuote(value))
}

// posixQuote wraps value in single quotes, escaping any embedded single
// quotes, so it is safe to eval regardless of shell metacharacters it contains.
func posixQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}

func (renderer *posix) RenderUnsetVariable(variable string) string {
	return fmt.Sprintf("unset %s", variable)
}

func (renderer *posix) Type() string {
	return ShellTypePosix
}
