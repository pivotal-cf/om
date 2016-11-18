package network_test

import (
	"net/http"
	"net/url"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/om/network"
)

var _ = Describe("WebClient", func() {
	var (
		server *ghttp.Server
	)
	BeforeEach(func() {
		server = ghttp.NewServer()
	})

	AfterEach(func() {
		server.Close()
	})
	Context("NewWebClient", func() {
		It("Should initialized a WebClient that has authenticated with OpsManager", func() {
			firstResponseHeaders := http.Header{}
			firstResponseHeaders.Add("Set-Cookie", "X-Uaa-Csrf=csrf_secret")

			secondResponseHeaders := http.Header{}
			secondResponseHeaders.Add("Set-Cookie", "_web_session=web_session; httponly; Path=/")
			secondResponseHeaders.Add("Set-Cookie", "uaa_access_token=uaa_access_token; httponly; Path=/")
			secondResponseHeaders.Add("Set-Cookie", "uaa_refresh_token=uaa_refresh_token; httponly; Path=/")

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/auth/cloudfoundry"),
					ghttp.RespondWith(http.StatusOK, "", firstResponseHeaders),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/uaa/login.do"),
					ghttp.VerifyBody([]byte("X-Uaa-Csrf=csrf_secret&password=keepitsimple&username=admin")),
					ghttp.RespondWith(http.StatusOK, "", secondResponseHeaders),
				),
			)

			webClient, err := network.NewWebClient(server.URL(), "admin", "keepitsimple", true, 1800*time.Second)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(webClient).ShouldNot(BeNil())

			hostURL, _ := url.Parse(server.URL())
			cookies := webClient.HTTPClient.Jar.Cookies(hostURL)
			Expect(len(cookies)).Should(Equal(3))
			var expectedCookies []*http.Cookie
			expectedCookies = append(expectedCookies,
				&http.Cookie{
					Name: "uaa_refresh_token", Value: "uaa_refresh_token",
				},
				&http.Cookie{
					Name: "_web_session", Value: "web_session",
				},
				&http.Cookie{
					Name: "uaa_access_token", Value: "uaa_access_token",
				},
			)
			Expect(cookies).Should(ConsistOf(expectedCookies))
		})
	})
})
