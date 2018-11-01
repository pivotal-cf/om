package network_test

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf/om/network"
	"github.com/pivotal-cf/om/network/fakes"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
)

var _ = Describe("DecryptClient", func() {
	var (
		fakeClient *fakes.HttpClient
	)

	BeforeEach(func() {
		fakeClient = &fakes.HttpClient{}
	})

	const correctDP = `correct-decryption-passphrase`

	Describe("Do", func() {
		successfulRequestOnIndex := func(index int) {
			fakeClient.DoReturnsOnCall(index, &http.Response{ // /api/v0/unlock
				StatusCode:    http.StatusOK,
				ContentLength: int64(len([]byte("{}"))),
				Body:          ioutil.NopCloser(strings.NewReader("{}")),
			}, nil)
			fakeClient.DoReturnsOnCall(index+1, &http.Response{ // /api/v0/ensure_availability
				StatusCode:    http.StatusOK,
				ContentLength: int64(len([]byte("Waiting for authentication system to start..."))),
				Body:          ioutil.NopCloser(strings.NewReader("Waiting for authentication system to start...")),
			}, nil)
			fakeClient.DoReturnsOnCall(index+2, &http.Response{ // /api/v0/ensure_availability
				StatusCode: http.StatusFound,
				Header: map[string][]string{
					"Location": []string{
						"https://example.com/auth/cloudfoundry",
					},
				},
				ContentLength: int64(len([]byte("Waiting for authentication system to start..."))),
				Body:          ioutil.NopCloser(strings.NewReader("Waiting for authentication system to start...")),
			}, nil)
			fakeClient.DoReturnsOnCall(index+3, &http.Response{StatusCode: http.StatusOK}, nil) // actual request
			fakeClient.DoReturnsOnCall(index+4, &http.Response{StatusCode: http.StatusOK}, nil) // actual request
		}

		Context("when the response is successful", func() {
			It("returns the response", func() {
				successfulRequestOnIndex(0)
				out := gbytes.NewBuffer()
				decryptClient := network.NewDecryptClient(fakeClient, fakeClient, correctDP, out)

				req := http.Request{Method: "some-method"}
				resp, err := decryptClient.Do(&req)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeClient.DoCallCount()).To(Equal(4))
				Expect(fakeClient.DoArgsForCall(3).Method).To(Equal("some-method"))
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(string(out.Contents())).To(ContainSubstring("Waiting for Ops Manager's auth systems to start. This may take a few minutes..."))
			})

			It("should only try to unlock once", func() {
				successfulRequestOnIndex(0)
				out := gbytes.NewBuffer()
				decryptClient := network.NewDecryptClient(fakeClient, fakeClient, correctDP, out)

				req := http.Request{Method: "some-method"}
				_, err := decryptClient.Do(&req)
				Expect(err).NotTo(HaveOccurred())

				resp, err := decryptClient.Do(&req)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeClient.DoCallCount()).To(Equal(5))
				Expect(fakeClient.DoArgsForCall(3).Method).To(Equal("some-method"))
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(string(out.Contents())).To(ContainSubstring("Waiting for Ops Manager's auth systems to start. This may take a few minutes..."))
			})
		})

		Context("when the response is error", func() {
			BeforeEach(func() {
				fakeClient.DoReturns(nil, errors.New("some-error"))
			})

			It("returns error", func() {
				out := gbytes.NewBuffer()
				decryptClient := network.NewDecryptClient(fakeClient, fakeClient, correctDP, out)

				req := http.Request{Method: "some-method"}
				_, err := decryptClient.Do(&req)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the network timeout", func() {
			It("should retry a couple of times then return the error", func() {
				fakeClient.DoReturnsOnCall(0, nil, &net.DNSError{IsTimeout: true})
				fakeClient.DoReturnsOnCall(1, nil, &net.DNSError{IsTimeout: true})
				fakeClient.DoReturnsOnCall(2, nil, &net.DNSError{IsTimeout: true})

				out := gbytes.NewBuffer()
				decryptClient := network.NewDecryptClient(fakeClient, fakeClient, correctDP, out)

				req := http.Request{Method: "some-method"}
				_, err := decryptClient.Do(&req)
				Expect(err).To(HaveOccurred())

				Expect(fakeClient.DoCallCount()).To(Equal(3))
				Expect(fakeClient.DoArgsForCall(2).Method).To(Equal("PUT"))
			})
		})

		Context("when the decryption passphrase provided is wrong", func() {
			BeforeEach(func() {
				fakeClient.DoReturnsOnCall(0, &http.Response{ // /api/v0/unlock
					StatusCode:    http.StatusForbidden,
					ContentLength: int64(len([]byte("{}"))),
					Body:          ioutil.NopCloser(strings.NewReader("{}")),
				}, nil)
			})

			It("returns the error", func() {
				out := gbytes.NewBuffer()
				decryptClient := network.NewDecryptClient(fakeClient, fakeClient, correctDP, out)

				req := http.Request{Method: "some-method"}
				_, err := decryptClient.Do(&req)
				Expect(err).To(MatchError("could not unlock ops manager, check if the decryption passphrase is correct"))
			})
		})
	})
})
