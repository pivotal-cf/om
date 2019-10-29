package api_test

import (
	"encoding/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/om/api"
	"net/http"
)

var _ = Describe("VMExtensions", func() {
	var (
		client  *ghttp.Server
		service api.Api
	)

	BeforeEach(func() {
		client = ghttp.NewServer()

		service = api.New(api.ApiInput{
			Client: httpClient{
				client.URL(),
			},
		})
	})

	AfterEach(func() {
		client.Close()
	})

	Context("creating a vm extension", func() {
		It("creates a VM Extension", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/api/v0/staged/vm_extensions/some-vm-extension"),
					ghttp.VerifyContentType("application/json"),
					ghttp.VerifyJSON(`{
					"name": "some-vm-extension",
					"cloud_properties": {
						"iam_instance_profile": "some-iam-profile",
						"elbs": ["some-elb"]
					}
				}`),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
			)

			err := service.CreateStagedVMExtension(api.CreateVMExtension{
				Name:            "some-vm-extension",
				CloudProperties: json.RawMessage(`{ "iam_instance_profile": "some-iam-profile", "elbs": ["some-elb"] }`),
			})

			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an error when the http status is non-200 for creating a vm extension", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/api/v0/staged/vm_extensions/some-vm-extension"),
					ghttp.RespondWith(http.StatusTeapot, `{}`),
				),
			)

			err := service.CreateStagedVMExtension(api.CreateVMExtension{
				Name:            "some-vm-extension",
				CloudProperties: json.RawMessage(`{ "iam_instance_profile": "some-iam-profile", "elbs": ["some-elb"] }`),
			})

			Expect(err).To(MatchError(ContainSubstring("request failed")))
		})

		It("returns an error when the api endpoint fails", func() {
			client.Close()

			err := service.CreateStagedVMExtension(api.CreateVMExtension{
				Name:            "some-vm-extension",
				CloudProperties: json.RawMessage(`{ "iam_instance_profile": "some-iam-profile", "elbs": ["some-elb"] }`),
			})
			Expect(err).To(MatchError(ContainSubstring("could not send api request to PUT /api/v0/staged/vm_extensions/some-vm-extension")))
		})
	})

	Context("listing vm extensions", func() {
		It("lists VM Extensions", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/vm_extensions"),
					ghttp.RespondWith(http.StatusOK, `{
					"vm_extensions": [{
						"name": "vm_ext1",
						"cloud_properties": {
							"source_dest_check": false
						}
					}, {
						"name": "vm_ext2",
						"cloud_properties": {
							"key_name": "operations_keypair"
						}
					}]
				}`),
				),
			)

			vmextensions, err := service.ListStagedVMExtensions()

			Expect(err).ToNot(HaveOccurred())
			Expect(len(vmextensions)).Should(Equal(2))
			Expect(vmextensions[0].Name).Should(Equal("vm_ext1"))
			Expect(vmextensions[1].Name).Should(Equal("vm_ext2"))
		})

		It("returns an error when the http status is non-200 for listing vm extensions", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/vm_extensions"),
					ghttp.RespondWith(http.StatusTeapot, `{}`),
				),
			)

			_, err := service.ListStagedVMExtensions()
			Expect(err).To(MatchError(ContainSubstring("request failed")))
		})

		It("returns an error when the api endpoint fails", func() {
			client.Close()

			_, err := service.ListStagedVMExtensions()
			Expect(err).To(MatchError(ContainSubstring("could not send api request to GET /api/v0/staged/vm_extensions")))
		})
	})

	Context("deleting a vm extension", func() {
		It("deletes a VM Extension", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", "/api/v0/staged/vm_extensions/some-vm-extension"),
					ghttp.VerifyContentType("application/json"),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
			)

			err := service.DeleteVMExtension("some-vm-extension")
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an error when the http status is non-200 for deleting vm extensions", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", "/api/v0/staged/vm_extensions/some-vm-extension"),
					ghttp.RespondWith(http.StatusTeapot, `{}`),
				),
			)

			err := service.DeleteVMExtension("some-vm-extension")
			Expect(err).To(MatchError(ContainSubstring("request failed")))
		})
	})
})
