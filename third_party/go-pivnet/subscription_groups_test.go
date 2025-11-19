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

var _ = Describe("PivnetClient - subscription groups", func() {
	var (
		server     *ghttp.Server
		client     pivnet.Client
		apiAddress string
		userAgent  string

		newClientConfig        pivnet.ClientConfig
		fakeLogger             logger.Logger
		fakeAccessTokenService *gopivnetfakes.FakeAccessTokenService
		response               interface{}
		responseStatusCode     int
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
		It("returns all subscription groups", func() {
			response := `{"subscription_groups": [{"id":2,"name":"subscription group 1"},{"id": 3, "name": "subscription group 2"}]}`

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/subscription_groups", apiPrefix)),
					ghttp.RespondWith(http.StatusOK, response),
				),
			)

			subscriptionGroups, err := client.SubscriptionGroups.List()
			Expect(err).NotTo(HaveOccurred())

			Expect(subscriptionGroups).To(HaveLen(2))
			Expect(subscriptionGroups[0].ID).To(Equal(2))
			Expect(subscriptionGroups[1].ID).To(Equal(3))
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
						ghttp.VerifyRequest("GET", fmt.Sprintf("%s/subscription_groups", apiPrefix)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				_, err := client.SubscriptionGroups.List()
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("%s/subscription_groups", apiPrefix)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.SubscriptionGroups.List()
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Get Subscription Group", func() {
		var (
			subscriptionGroupID int
		)

		BeforeEach(func() {
			subscriptionGroupID = 1234

			response = pivnet.SubscriptionGroup{
				ID:   subscriptionGroupID,
				Name: "some subscription group",
				Members: []pivnet.SubscriptionGroupMember{
					{
						ID:      4321,
						Name:    "subscription group member 1",
						Email:   "dude@dude.dude",
						IsAdmin: false,
					},
					{
						ID:      9876,
						Name:    "subscription group member 2",
						Email:   "buddy@buddy.buddy",
						IsAdmin: true,
					},
				},
				PendingInvitations: []string{},
				Subscriptions:      []pivnet.SubscriptionGroupSubscription{},
			}
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf(
							"%s/subscription_groups/%d",
							apiPrefix,
							subscriptionGroupID,
						),
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("returns subscription group without errors", func() {
			_, err := client.SubscriptionGroups.Get(subscriptionGroupID)

			Expect(err).NotTo(HaveOccurred())
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
								"%s/subscription_groups/%d",
								apiPrefix,
								subscriptionGroupID,
							),
						),
						ghttp.RespondWith(responseStatusCode, body),
					),
				)

				_, err := client.SubscriptionGroups.Get(
					subscriptionGroupID,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})
	})

	Describe("AddMember", func() {
		var (
			subscriptionGroupID int
			memberEmailAddress  string
			expectedRequestBody string
		)

		BeforeEach(func() {
			subscriptionGroupID = 1234
			memberEmailAddress = "dude@dude.dude"

			response = pivnet.SubscriptionGroup{
				ID:   subscriptionGroupID,
				Name: "some subscription group",
				Members: []pivnet.SubscriptionGroupMember{
					{
						ID:      4321,
						Name:    "subscription group member 1",
						Email:   "dude@dude.dude",
						IsAdmin: false,
					},
					{
						ID:      9876,
						Name:    "subscription group member 2",
						Email:   "buddy@buddy.buddy",
						IsAdmin: true,
					},
				},
				PendingInvitations: []string{},
				Subscriptions:      []pivnet.SubscriptionGroupSubscription{},
			}

			expectedRequestBody = fmt.Sprintf(
				`{"member":{"email":"%s","admin":false}}`,
				memberEmailAddress,
			)
		})

		It("should return the changed subscription group when successful", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"PATCH",
						fmt.Sprintf(
							"%s/subscription_groups/%d/add_member",
							apiPrefix,
							subscriptionGroupID,
						),
					),
					ghttp.VerifyJSON(expectedRequestBody),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)

			_, err := client.SubscriptionGroups.AddMember(subscriptionGroupID, memberEmailAddress, "false")

			Expect(err).NotTo(HaveOccurred())
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
						ghttp.VerifyRequest("PATCH", fmt.Sprintf("%s/subscription_groups/%d/add_member", apiPrefix, 1234)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				_, err := client.SubscriptionGroups.AddMember(1234, "dude@dude.dude", "false")
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/subscription_groups/%d/add_member",
							apiPrefix,
							4321,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.SubscriptionGroups.AddMember(4321, memberEmailAddress, "false")
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("RemoveMember", func() {
		var (
			subscriptionGroupID int
			memberEmailAddress  string
			expectedRequestBody string
		)

		BeforeEach(func() {
			subscriptionGroupID = 1234
			memberEmailAddress = "dude@dude.dude"

			response = pivnet.SubscriptionGroup{
				ID:   subscriptionGroupID,
				Name: "some subscription group",
				Members: []pivnet.SubscriptionGroupMember{
					{
						ID:      4321,
						Name:    "subscription group member 1",
						Email:   "dude@dude.dude",
						IsAdmin: false,
					},
					{
						ID:      9876,
						Name:    "subscription group member 2",
						Email:   "buddy@buddy.buddy",
						IsAdmin: true,
					},
				},
				PendingInvitations: []string{},
				Subscriptions:      []pivnet.SubscriptionGroupSubscription{},
			}

			expectedRequestBody = fmt.Sprintf(
				`{"member":{"email":"%s"}}`,
				memberEmailAddress,
			)
		})

		It("should return the changed subscription group when successful", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"PATCH",
						fmt.Sprintf(
							"%s/subscription_groups/%d/remove_member",
							apiPrefix,
							subscriptionGroupID,
						),
					),
					ghttp.VerifyJSON(expectedRequestBody),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)

			_, err := client.SubscriptionGroups.RemoveMember(subscriptionGroupID, memberEmailAddress)

			Expect(err).NotTo(HaveOccurred())
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
						ghttp.VerifyRequest("PATCH", fmt.Sprintf("%s/subscription_groups/%d/remove_member", apiPrefix, 1234)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				_, err := client.SubscriptionGroups.RemoveMember(1234, "dude@dude.dude")
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/subscription_groups/%d/remove_member",
							apiPrefix,
							4321,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.SubscriptionGroups.RemoveMember(4321, memberEmailAddress)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

})
