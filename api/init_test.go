package api_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"

	"testing"
)

func TestApi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "api")
}

type ensureHandler struct {
	handlers []http.HandlerFunc
}

func (e *ensureHandler) Ensure(funs ...http.HandlerFunc) []http.HandlerFunc {
	for _, fun := range funs {
		e.handlers = append(e.handlers, func(writer http.ResponseWriter, request *http.Request) {
			fun(writer, request)
			e.handlers = e.handlers[1:]
		})
	}

	return e.handlers
}

func (e *ensureHandler) Handlers() []http.HandlerFunc {
	return e.handlers
}
