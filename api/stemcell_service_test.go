package api_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"net/http"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/pivotal-cf/om/api"
)

var _ = Describe("StemcellService", func() {
	var (
		server  *ghttp.Server
		service api.Api
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

		client := httpClient{
			server.URL(),
		}

		service = api.New(api.ApiInput{
			Client:         client,
			UnauthedClient: client,
		})
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("ListStemcells", func() {
		It("makes a request to list the stemcells", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/stemcell_assignments"),
					ghttp.RespondWith(http.StatusOK, `{
						"products": [{
							"guid": "some-guid",
							"staged_stemcell_version": "1234.5",
							"identifier": "some-product",
							"available_stemcell_versions": [
								"1234.5", "1234.6"
							]
						}]
					}`),
				),
			)

			output, err := service.ListStemcells()
			Expect(err).ToNot(HaveOccurred())
			Expect(output).To(Equal(api.ProductStemcells{
				Products: []api.ProductStemcell{
					{
						GUID:                  "some-guid",
						StagedForDeletion:     false,
						StagedStemcellVersion: "1234.5",
						ProductName:           "some-product",
						AvailableVersions: []string{
							"1234.5",
							"1234.6",
						},
					},
				},
			}))
		})

		When("invalid JSON is returned", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/stemcell_assignments"),
						ghttp.RespondWith(http.StatusOK, `invalid-json`),
					),
				)

				_, err := service.ListStemcells()
				Expect(err).To(MatchError(ContainSubstring("invalid JSON: invalid character 'i' looking for beginning of value")))
			})
		})

		When("the server errors before the request", func() {
			It("returns an error", func() {
				server.Close()

				_, err := service.ListStemcells()
				Expect(err).To(MatchError(ContainSubstring("could not make api request to list stemcells: could not send api request to GET /api/v0/stemcell_assignments")))
			})
		})

		When("the api returns a non-200 status code", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/stemcell_assignments"),
						ghttp.RespondWith(http.StatusTeapot, `{}`),
					),
				)

				_, err := service.ListStemcells()
				Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
			})
		})
	})

	Describe("AssignStemcells", func() {
		It("makes a request to assign the stemcells", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PATCH", "/api/v0/stemcell_assignments"),
					ghttp.VerifyJSON(`{
						"products": [{
							"guid": "some-guid",
							"staged_stemcell_version": "1234.6"
						}]
					}`),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
			)

			err := service.AssignStemcell(api.ProductStemcells{
				Products: []api.ProductStemcell{{
					GUID:                  "some-guid",
					StagedStemcellVersion: "1234.6",
				}},
			})
			Expect(err).ToNot(HaveOccurred())
		})

		When("the server errors before the request", func() {
			It("returns an error", func() {
				server.Close()

				err := service.AssignStemcell(api.ProductStemcells{
					Products: []api.ProductStemcell{{
						GUID:                  "some-guid",
						StagedStemcellVersion: "1234.6",
					}},
				})
				Expect(err).To(MatchError(ContainSubstring("could not send api request to PATCH /api/v0/stemcell_assignments")))
			})
		})

		When("the api returns a non-200 status code", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", "/api/v0/stemcell_assignments"),
						ghttp.VerifyJSON(`{
							"products": [{
								"guid": "some-guid",
								"staged_stemcell_version": "1234.6"
							}]
						}`),
						ghttp.RespondWith(http.StatusTeapot, `{}`),
					),
				)

				err := service.AssignStemcell(api.ProductStemcells{
					Products: []api.ProductStemcell{{
						GUID:                  "some-guid",
						StagedStemcellVersion: "1234.6",
					}},
				})
				Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
			})
		})
	})

	Describe("CheckStemcellAvailability", func() {
		When("the diagnostic report exists with stemcells", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.RespondWith(http.StatusOK, `
				{
					"stemcells": ["light-bosh-stemcell-621.79-google-kvm-ubuntu-xenial-go_agent.tgz"],
					"infrastructure_type": "vsphere-esxi",
					"available_stemcells": [
						{
							"filename": "light-bosh-stemcell-621.80-google-kvm-ubuntu-xenial-go_agent.tgz"
						},
						{
							"filename": "bosh-vsphere-esxi-ubuntu-jammy-go_agent-1.1016.tgz",
							"os": "ubuntu-jammy",
							"version": "1.1016"
						}
					]
				}
				`),
				)
			})

			When("the version of the OpsManager cannot be determined", func() {
				It("returns an error", func() {
					server.AppendHandlers(
						ghttp.RespondWith(http.StatusOK, `{}`),
					)

					_, err := service.CheckStemcellAvailability("light-bosh-stemcell-621.77-google-kvm-ubuntu-xenial-go_agent.tgz")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("could not determine version"))
				})
			})

			When("the version cannot be gotten", func() {
				It("returns an error", func() {
					server.AppendHandlers(
						ghttp.RespondWith(http.StatusNotFound, nil),
					)

					_, err := service.CheckStemcellAvailability("light-bosh-stemcell-621.77-google-kvm-ubuntu-xenial-go_agent.tgz")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("cannot retrieve version"))
				})
			})

			When("the stemcell does not exists", func() {
				It("returns false", func() {
					server.AppendHandlers(
						ghttp.RespondWith(http.StatusOK, `{"info":{"version":"2.4.3"}}`),
					)

					found, err := service.CheckStemcellAvailability("light-bosh-stemcell-621.77-google-kvm-ubuntu-xenial-go_agent.tgz")
					Expect(err).NotTo(HaveOccurred())

					Expect(found).To(BeFalse())
				})

				When("the OpsManager is 2.6+", func() {
					It("returns false", func() {
						server.AppendHandlers(
							ghttp.RespondWith(http.StatusOK, `{"info":{"version":"2.6.3"}}`),
						)

						found, err := service.CheckStemcellAvailability("light-bosh-stemcell-621.77-google-kvm-ubuntu-xenial-go_agent.tgz")
						Expect(err).NotTo(HaveOccurred())

						Expect(found).To(BeFalse())
					})
				})
			})

			When("the stemcell already exists", func() {
				It("returns true", func() {
					server.AppendHandlers(
						ghttp.RespondWith(http.StatusOK, `{"info":{"version":"2.4.3"}}`),
					)

					found, err := service.CheckStemcellAvailability("light-bosh-stemcell-621.79-google-kvm-ubuntu-xenial-go_agent.tgz")
					Expect(err).NotTo(HaveOccurred())

					Expect(found).To(BeTrue())
				})

				When("the OpsMan 2.6+", func() {
					It("returns true", func() {
						server.AppendHandlers(
							ghttp.RespondWith(http.StatusOK, `{"info":{"version":"2.6.3"}}`),
						)

						found, err := service.CheckStemcellAvailability("light-bosh-stemcell-621.80-google-kvm-ubuntu-xenial-go_agent.tgz")
						Expect(err).NotTo(HaveOccurred())

						Expect(found).To(BeTrue())
					})

					It("returns true when same stemcell is provided with different filename (matched by OS, version, infrastructure)", func() {
						// Create a minimal stemcell .tgz with stemcell.MF
						manifestContent := `name: bosh-vsphere-esxi-ubuntu-jammy
version: "1.1016"
operating_system: ubuntu-jammy
cloud_properties:
  infrastructure: vsphere-esxi
`
						var buf bytes.Buffer
						gw := gzip.NewWriter(&buf)
						tw := tar.NewWriter(gw)
						Expect(tw.WriteHeader(&tar.Header{Name: "stemcell.MF", Size: int64(len(manifestContent))})).To(Succeed())
						_, err := tw.Write([]byte(manifestContent))
						Expect(err).NotTo(HaveOccurred())
						Expect(tw.Close()).To(Succeed())
						Expect(gw.Close()).To(Succeed())

						stemcellPath := filepath.Join(os.TempDir(), "bosh-stemcell-1.1016-vsphere-esxi-ubuntu-jammy-go_agent.tgz")
						Expect(os.WriteFile(stemcellPath, buf.Bytes(), 0600)).To(Succeed())
						defer os.Remove(stemcellPath)

						// Ops Manager has the same stemcell under a different filename
						server.AppendHandlers(
							ghttp.RespondWith(http.StatusOK, `{"info":{"version":"2.6.3"}}`),
						)

						found, err := service.CheckStemcellAvailability(stemcellPath)
						Expect(err).NotTo(HaveOccurred())
						Expect(found).To(BeTrue())
					})

					It("falls back to filename matching when manifest extraction fails (e.g. file is not a valid .tgz)", func() {
						// File is not a valid stemcell .tgz (no stemcell.MF), so manifest extraction fails.
						// Should fall back to filename match; report has this filename in available_stemcells.
						invalidPath := filepath.Join(os.TempDir(), "light-bosh-stemcell-621.80-google-kvm-ubuntu-xenial-go_agent.tgz")
						Expect(os.WriteFile(invalidPath, []byte("not a tgz"), 0600)).To(Succeed())
						defer os.Remove(invalidPath)

						server.AppendHandlers(
							ghttp.RespondWith(http.StatusOK, `{"info":{"version":"2.6.3"}}`),
						)

						found, err := service.CheckStemcellAvailability(invalidPath)
						Expect(err).NotTo(HaveOccurred())
						Expect(found).To(BeTrue())
					})
				})
			})
		})

		When("the diagnostic report has different infrastructure than stemcell manifest", func() {
			It("returns false when same OS and version but different infrastructure", func() {
				// Report says google-kvm; stemcell manifest says vsphere-esxi -> no match.
				server.AppendHandlers(
					ghttp.RespondWith(http.StatusOK, `{"stemcells": [], "infrastructure_type": "google-kvm", "available_stemcells": [{"filename": "bosh-google-kvm-ubuntu-jammy-1.1016.tgz", "os": "ubuntu-jammy", "version": "1.1016"}]}`),
					ghttp.RespondWith(http.StatusOK, `{"info":{"version":"2.6.3"}}`),
				)

				manifestContent := `name: bosh-vsphere-esxi-ubuntu-jammy
version: "1.1016"
operating_system: ubuntu-jammy
cloud_properties:
  infrastructure: vsphere-esxi
`
				var buf bytes.Buffer
				gw := gzip.NewWriter(&buf)
				tw := tar.NewWriter(gw)
				Expect(tw.WriteHeader(&tar.Header{Name: "stemcell.MF", Size: int64(len(manifestContent))})).To(Succeed())
				_, err := tw.Write([]byte(manifestContent))
				Expect(err).NotTo(HaveOccurred())
				Expect(tw.Close()).To(Succeed())
				Expect(gw.Close()).To(Succeed())

				stemcellPath := filepath.Join(os.TempDir(), "bosh-stemcell-1.1016-vsphere-esxi-ubuntu-jammy-go_agent.tgz")
				Expect(os.WriteFile(stemcellPath, buf.Bytes(), 0600)).To(Succeed())
				defer os.Remove(stemcellPath)

				found, err := service.CheckStemcellAvailability(stemcellPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})

		When("Docker Ops Manager (warden/docker infrastructure equivalence)", func() {
			It("returns true when stemcell manifest has warden and report has docker", func() {
				// Docker Ops Manager reports infrastructure_type "docker" but stemcells are built for "warden".
				manifestContent := `name: bosh-warden-ubuntu-jammy
version: "1.1016"
operating_system: ubuntu-jammy
cloud_properties:
  infrastructure: warden
`
				var buf bytes.Buffer
				gw := gzip.NewWriter(&buf)
				tw := tar.NewWriter(gw)
				Expect(tw.WriteHeader(&tar.Header{Name: "stemcell.MF", Size: int64(len(manifestContent))})).To(Succeed())
				_, err := tw.Write([]byte(manifestContent))
				Expect(err).NotTo(HaveOccurred())
				Expect(tw.Close()).To(Succeed())
				Expect(gw.Close()).To(Succeed())

				stemcellPath := filepath.Join(os.TempDir(), "bosh-stemcell-1.1016-warden-ubuntu-jammy-go_agent.tgz")
				Expect(os.WriteFile(stemcellPath, buf.Bytes(), 0600)).To(Succeed())
				defer os.Remove(stemcellPath)

				server.AppendHandlers(
					ghttp.RespondWith(http.StatusOK, `{"stemcells": [], "infrastructure_type": "docker", "available_stemcells": [{"filename": "bosh-docker-ubuntu-jammy-1.1016.tgz", "os": "ubuntu-jammy", "version": "1.1016"}]}`),
					ghttp.RespondWith(http.StatusOK, `{"info":{"version":"2.6.3"}}`),
				)

				found, err := service.CheckStemcellAvailability(stemcellPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeTrue())
			})
		})

		When("the diagnostic report fails", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.RespondWith(http.StatusNotFound, nil),
				)
				_, err := service.CheckStemcellAvailability("light-bosh-stemcell-621.79-google-kvm-ubuntu-xenial-go_agent.tgz")
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
