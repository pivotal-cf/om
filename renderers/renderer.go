package renderers

// Renderer defines a rendering interface
type Renderer interface {
	RenderEnvironmentVariable(variable string, value string) string
	RenderUnsetVariable(variable string) string
	Type() string
}
