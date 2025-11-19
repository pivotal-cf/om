package pivnet_test

import (
	"fmt"
	"github.com/pivotal-cf/go-pivnet/v7/go-pivnetfakes"
	"net/http"

	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/go-pivnet/v7"
	"github.com/pivotal-cf/go-pivnet/v7/logger"
	"github.com/pivotal-cf/go-pivnet/v7/logger/loggerfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PivnetClient - user groups", func() {
	var (
		server     *ghttp.Server
		client     pivnet.Client
		apiAddress string
		userAgent  string

		newClientConfig        pivnet.ClientConfig
		fakeLogger             logger.Logger
		fakeAccessTokenService *gopivnetfakes.FakeAccessTokenService

		response           interface{}
		responseStatusCode int
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		apiAddress = server.URL()
		userAgent = "pivnet-resource/0.1.0 (some-url)"

		fakeLogger = &loggerfakes.FakeLogger{}
		fakeAccessTokenService = &gopivnetfakes.FakeAccessTokenService{}
		newClientConfig = pivnet.ClientConfig{
			Host:      apiAddress,
			UserAgent: userAgent,
		}
		client = pivnet.NewClient(fakeAccessTokenService, newClientConfig, fakeLogger)

		responseStatusCode = http.StatusOK
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("List", func() {
		It("returns all user groups", func() {
			response := `{"user_groups": [{"id":2,"name":"group 1"},{"id": 3, "name": "group 2"}]}`

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/user_groups", apiPrefix)),
					ghttp.RespondWith(http.StatusOK, response),
				),
			)

			userGroups, err := client.UserGroups.List()
			Expect(err).NotTo(HaveOccurred())

			Expect(userGroups).To(HaveLen(2))
			Expect(userGroups[0].ID).To(Equal(2))
			Expect(userGroups[1].ID).To(Equal(3))
		})

		Context("when the server responds with a non-2XX status code", func() {
			var (
				body []byte
			)

			BeforeEach(func() {
				body = []byte(`{"message":"foo message"}`)
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("%s/user_groups", apiPrefix)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				_, err := client.UserGroups.List()
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("%s/user_groups", apiPrefix)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.UserGroups.List()
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("List for release", func() {
		var (
			releaseID int
		)

		BeforeEach(func() {
			releaseID = 1234
		})

		It("returns the user groups for the product slug", func() {
			response := `{"user_groups": [{"id":2,"name":"group 1"},{"id": 3, "name": "group 2"}]}`

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products/banana/releases/%d/user_groups", apiPrefix, releaseID)),
					ghttp.RespondWith(http.StatusOK, response),
				),
			)

			userGroups, err := client.UserGroups.ListForRelease("banana", releaseID)
			Expect(err).NotTo(HaveOccurred())

			Expect(userGroups).To(HaveLen(2))
			Expect(userGroups[0].ID).To(Equal(2))
			Expect(userGroups[1].ID).To(Equal(3))
		})

		Context("when the server responds with a non-2XX status code", func() {
			var (
				body []byte
			)

			BeforeEach(func() {
				body = []byte(`{"message":"foo message"}`)
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products/banana/releases/%d/user_groups", apiPrefix, releaseID)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				_, err := client.UserGroups.ListForRelease("banana", releaseID)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products/banana/releases/%d/user_groups", apiPrefix, releaseID)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.UserGroups.ListForRelease("banana", releaseID)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Add To Release", func() {
		var (
			productSlug = "banana-slug"
			releaseID   = 2345
			userGroupID = 3456

			expectedRequestBody = `{"user_group":{"id":3456}}`
		)

		Context("when the server responds with a 204 status code", func() {
			It("returns without error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d/add_user_group",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.VerifyJSON(expectedRequestBody),
						ghttp.RespondWith(http.StatusNoContent, nil),
					),
				)

				err := client.UserGroups.AddToRelease(productSlug, releaseID, userGroupID)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the server responds with a non-204 status code", func() {
			var (
				body []byte
			)

			BeforeEach(func() {
				body = []byte(`{"message":"foo message"}`)
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d/add_user_group",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				err := client.UserGroups.AddToRelease(productSlug, releaseID, userGroupID)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d/add_user_group",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				err := client.UserGroups.AddToRelease(productSlug, releaseID, userGroupID)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Remove From Release", func() {
		var (
			productSlug = "banana-slug"
			releaseID   = 2345
			userGroupID = 3456

			expectedRequestBody = `{"user_group":{"id":3456}}`
		)

		Context("when the server responds with a 204 status code", func() {
			It("returns without error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d/remove_user_group",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.VerifyJSON(expectedRequestBody),
						ghttp.RespondWith(http.StatusNoContent, nil),
					),
				)

				err := client.UserGroups.RemoveFromRelease(productSlug, releaseID, userGroupID)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the server responds with a non-204 status code", func() {
			var (
				body []byte
			)

			BeforeEach(func() {
				body = []byte(`{"message":"foo message"}`)
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d/remove_user_group",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				err := client.UserGroups.RemoveFromRelease(productSlug, releaseID, userGroupID)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d/remove_user_group",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				err := client.UserGroups.RemoveFromRelease(productSlug, releaseID, userGroupID)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Get User Group", func() {
		var (
			userGroupID int
		)

		BeforeEach(func() {
			userGroupID = 1234

			response = pivnet.UserGroup{
				ID:   userGroupID,
				Name: "something",
			}
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf(
							"%s/user_groups/%d",
							apiPrefix,
							userGroupID,
						),
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("returns the user group without error", func() {
			userGroup, err := client.UserGroups.Get(
				userGroupID,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(userGroup.ID).To(Equal(userGroupID))
			Expect(userGroup.Name).To(Equal("something"))
		})

		Context("when the server responds with a non-2XX status code", func() {
			var (
				body []byte
			)

			BeforeEach(func() {
				response = pivnetErr{Message: "foo message"}
				responseStatusCode = http.StatusTeapot
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(
							"GET",
							fmt.Sprintf(
								"%s/user_groups/%d",
								apiPrefix,
								userGroupID,
							),
						),
						ghttp.RespondWith(responseStatusCode, body),
					),
				)

				_, err := client.UserGroups.Get(
					userGroupID,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			BeforeEach(func() {
				response = "%%%"
			})

			It("forwards the error", func() {
				_, err := client.UserGroups.Get(userGroupID)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("json"))
			})
		})
	})

	Describe("Create", func() {
		var (
			name        string
			description string
			members     []string

			expectedRequestBody string

			returnedUserGroup pivnet.UserGroup
		)

		BeforeEach(func() {
			name = "some name"
			description = "some description"
			members = []string{"some member"}

			expectedRequestBody = fmt.Sprintf(
				`{"user_group":{"name":"%s","description":"%s","members":["some member"]}}`,
				name,
				description,
			)
		})

		JustBeforeEach(func() {
			returnedUserGroup = pivnet.UserGroup{
				ID:          1234,
				Name:        name,
				Description: description,
				Members:     members,
			}
		})

		It("creates new user group without error", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", fmt.Sprintf(
						"%s/user_groups",
						apiPrefix,
					)),
					ghttp.VerifyJSON(expectedRequestBody),
					ghttp.RespondWithJSONEncoded(http.StatusCreated, returnedUserGroup),
				),
			)

			userGroup, err := client.UserGroups.Create(name, description, members)
			Expect(err).NotTo(HaveOccurred())

			Expect(userGroup.ID).To(Equal(returnedUserGroup.ID))
			Expect(userGroup.Name).To(Equal(name))
			Expect(userGroup.Description).To(Equal(description))
		})

		Context("when members is nil", func() {
			BeforeEach(func() {
				members = nil

				expectedRequestBody = fmt.Sprintf(
					`{"user_group":{"name":"%s","description":"%s","members":[]}}`,
					name,
					description,
				)
			})

			It("successfully sends empty array in json body", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", fmt.Sprintf(
							"%s/user_groups",
							apiPrefix,
						)),
						ghttp.VerifyJSON(expectedRequestBody),
						ghttp.RespondWithJSONEncoded(http.StatusCreated, returnedUserGroup),
					),
				)

				_, err := client.UserGroups.Create(name, description, members)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the server responds with a non-201 status code", func() {
			var (
				body []byte
			)

			BeforeEach(func() {
				body = []byte(`{"message":"foo message"}`)
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", fmt.Sprintf(
							"%s/user_groups",
							apiPrefix,
						)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				_, err := client.UserGroups.Create(name, description, members)

				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", fmt.Sprintf(
							"%s/user_groups",
							apiPrefix,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.UserGroups.Create(name, description, members)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Update", func() {
		var (
			userGroup pivnet.UserGroup

			expectedRequestBody string

			response pivnet.UpdateUserGroupResponse
		)

		BeforeEach(func() {
			userGroup = pivnet.UserGroup{
				ID:          1234,
				Name:        "some name",
				Description: "some description",
			}

			expectedRequestBody = fmt.Sprintf(
				`{"user_group":{"name":"%s","description":"%s"}}`,
				userGroup.Name,
				userGroup.Description,
			)

			response = pivnet.UpdateUserGroupResponse{userGroup}
		})

		It("returns without error", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PATCH", fmt.Sprintf(
						"%s/user_groups/%d",
						apiPrefix,
						userGroup.ID,
					)),
					ghttp.VerifyJSON(expectedRequestBody),
					ghttp.RespondWithJSONEncoded(http.StatusOK, response),
				),
			)

			returned, err := client.UserGroups.Update(userGroup)
			Expect(err).NotTo(HaveOccurred())

			Expect(returned.ID).To(Equal(userGroup.ID))
			Expect(returned.Name).To(Equal(userGroup.Name))
			Expect(returned.Description).To(Equal(userGroup.Description))
		})

		Context("when the server responds with a non-200 status code", func() {
			var (
				body []byte
			)

			BeforeEach(func() {
				body = []byte(`{"message":"foo message"}`)
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/user_groups/%d",
							apiPrefix,
							userGroup.ID,
						)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				_, err := client.UserGroups.Update(userGroup)

				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/user_groups/%d",
							apiPrefix,
							userGroup.ID,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.UserGroups.Update(userGroup)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Delete", func() {
		var (
			userGroup pivnet.UserGroup
		)

		BeforeEach(func() {
			userGroup = pivnet.UserGroup{
				ID: 1234,
			}
		})

		It("deletes the release", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", fmt.Sprintf("%s/user_groups/%d", apiPrefix, userGroup.ID)),
					ghttp.RespondWith(http.StatusNoContent, nil),
				),
			)

			err := client.UserGroups.Delete(userGroup.ID)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the server responds with a non-204 status code", func() {
			var (
				body []byte
			)

			BeforeEach(func() {
				body = []byte(`{"message":"foo message"}`)
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", fmt.Sprintf("%s/user_groups/%d", apiPrefix, userGroup.ID)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				err := client.UserGroups.Delete(userGroup.ID)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", fmt.Sprintf("%s/user_groups/%d", apiPrefix, userGroup.ID)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				err := client.UserGroups.Delete(userGroup.ID)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("AddMemberToGroup", func() {
		var (
			memberEmailAddress string
			admin              bool
			userGroup          pivnet.UserGroup

			expectedRequestBody string

			response pivnet.UpdateUserGroupResponse
		)

		BeforeEach(func() {
			memberEmailAddress = "some email address"
			admin = true

			userGroup = pivnet.UserGroup{
				ID:          1234,
				Name:        "some name",
				Description: "some description",
			}

			expectedRequestBody = fmt.Sprintf(
				`{"member":{"email":"%s","admin":true}}`,
				memberEmailAddress,
			)

			response = pivnet.UpdateUserGroupResponse{userGroup}
		})

		It("returns without error", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PATCH", fmt.Sprintf(
						"%s/user_groups/%d/add_member",
						apiPrefix,
						userGroup.ID,
					)),
					ghttp.VerifyJSON(expectedRequestBody),
					ghttp.RespondWithJSONEncoded(http.StatusOK, response),
				),
			)

			returned, err := client.UserGroups.AddMemberToGroup(
				userGroup.ID,
				memberEmailAddress,
				admin,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(returned.ID).To(Equal(userGroup.ID))
			Expect(returned.Name).To(Equal(userGroup.Name))
			Expect(returned.Description).To(Equal(userGroup.Description))
		})

		Context("when the server responds with a non-200 status code", func() {
			var (
				body []byte
			)

			BeforeEach(func() {
				body = []byte(`{"message":"foo message"}`)
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/user_groups/%d/add_member",
							apiPrefix,
							userGroup.ID,
						)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				_, err := client.UserGroups.AddMemberToGroup(
					userGroup.ID,
					memberEmailAddress,
					admin,
				)

				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/user_groups/%d/add_member",
							apiPrefix,
							userGroup.ID,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.UserGroups.AddMemberToGroup(
					userGroup.ID,
					memberEmailAddress,
					admin,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("RemoveMemberFromGroup", func() {
		var (
			memberEmailAddress string
			userGroup          pivnet.UserGroup

			expectedRequestBody string

			response pivnet.UpdateUserGroupResponse
		)

		BeforeEach(func() {
			memberEmailAddress = "some email address"

			userGroup = pivnet.UserGroup{
				ID:          1234,
				Name:        "some name",
				Description: "some description",
			}

			expectedRequestBody = fmt.Sprintf(
				`{"member":{"email":"%s"}}`,
				memberEmailAddress,
			)

			response = pivnet.UpdateUserGroupResponse{userGroup}
		})

		It("returns without error", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PATCH", fmt.Sprintf(
						"%s/user_groups/%d/remove_member",
						apiPrefix,
						userGroup.ID,
					)),
					ghttp.VerifyJSON(expectedRequestBody),
					ghttp.RespondWithJSONEncoded(http.StatusOK, response),
				),
			)

			returned, err := client.UserGroups.RemoveMemberFromGroup(userGroup.ID, memberEmailAddress)
			Expect(err).NotTo(HaveOccurred())

			Expect(returned.ID).To(Equal(userGroup.ID))
			Expect(returned.Name).To(Equal(userGroup.Name))
			Expect(returned.Description).To(Equal(userGroup.Description))
		})

		Context("when the server responds with a non-200 status code", func() {
			var (
				body []byte
			)

			BeforeEach(func() {
				body = []byte(`{"message":"foo message"}`)
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/user_groups/%d/remove_member",
							apiPrefix,
							userGroup.ID,
						)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				_, err := client.UserGroups.RemoveMemberFromGroup(userGroup.ID, memberEmailAddress)

				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/user_groups/%d/remove_member",
							apiPrefix,
							userGroup.ID,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.UserGroups.RemoveMemberFromGroup(userGroup.ID, memberEmailAddress)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})
})
