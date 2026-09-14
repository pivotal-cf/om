package renderers

import (
	"fmt"
	"strings"
)

type powershell struct {
}

// NewPowershell creates a new Powershell Renderer
func NewPowershell() Renderer {
	return &powershell{}
}

func (renderer *powershell) RenderEnvironmentVariable(variable string, value string) string {
	if strings.ContainsAny(value, "\n") {
		if !strings.HasSuffix(value, "\r\n") {
			value += "\r\n"
		}
		value = "\r\n" + value
	}
	return fmt.Sprintf("$env:%s=%s", variable, powershellQuote(value))
}

// powershellQuote wraps value in single quotes, doubling any embedded
// single quotes, so it is safe to iex regardless of `"`, backtick, or
// $() metacharacters it contains.
func powershellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func (renderer *powershell) RenderUnsetVariable(variable string) string {
	return fmt.Sprintf("$env:%s=$null", variable)
}

func (renderer *powershell) Type() string {
	return ShellTypePowershell
}
