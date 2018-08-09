package builder

type logger interface {
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

type Metadata map[string]interface{}
