package network

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("url", func() {
	Describe("parseURL", func() {
		It("Should parse target URL with http protocol", func() {
			u, err := parseURL("http://opsman.example.com")
			Expect(err).ToNot(HaveOccurred())
			Expect(u.String()).To(Equal("http://opsman.example.com"))
		})

		It("Should parse target URL with https protocol", func() {
			u, err := parseURL("https://opsman.example.com")
			Expect(err).ToNot(HaveOccurred())
			Expect(u.String()).To(Equal("https://opsman.example.com"))
		})

		It("Should parse target URL without protocol", func() {
			u, err := parseURL("opsman.example.com")
			Expect(err).ToNot(HaveOccurred())
			Expect(u.String()).To(Equal("https://opsman.example.com"))
		})

		It("Should return target flag is required when target URL is empty", func() {
			_, err := parseURL("")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("target flag is required, run `om help` for more info"))
		})

		It("Should return target flag is required when target URL is missing host", func() {
			_, err := parseURL("https://")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("target flag is required, run `om help` for more info"))
		})

		It("Should not parse target URL with incorrect protocol", func() {
			_, err := parseURL("smb://opsman.example.com")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("error parsing URL, expected http(s) protocol but got smb"))
		})

		It("Should not parse target URL with bad URL", func() {
			_, err := parseURL("a bad\\url")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("parse \"https://a bad\\\\url\": invalid character \" \" in host name"))
		})
	})

	Describe("parseOpsmanAndUAAURLs", func() {
		It("Should default the UAA target to /uaa under OpsMan target", func() {
			opsmanURL, uaaURL, err := parseOpsmanAndUAAURLs("https://opsman.example.com", "")
			Expect(err).ToNot(HaveOccurred())
			Expect(opsmanURL.String()).To(Equal("https://opsman.example.com"))
			Expect(uaaURL.String()).To(Equal("https://opsman.example.com/uaa"))
		})

		It("Should use the UAA target when specified and valid", func() {
			opsmanURL, uaaURL, err := parseOpsmanAndUAAURLs("https://opsman.example.com", "https://uaa.example.com")
			Expect(err).ToNot(HaveOccurred())
			Expect(opsmanURL.String()).To(Equal("https://opsman.example.com"))
			Expect(uaaURL.String()).To(Equal("https://uaa.example.com"))
		})

		It("Should return an error when OpsMan URL is empty", func() {
			_, _, err := parseOpsmanAndUAAURLs("", "https://uaa.example.com")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("could not parse Opsman target URL: target flag is required, run `om help` for more info"))
		})

		It("Should return an error when OpsMan URL is invalid", func() {
			_, _, err := parseOpsmanAndUAAURLs("smb://opsman.example.com", "")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("could not parse Opsman target URL: error parsing URL, expected http(s) protocol but got smb"))
		})

		It("Should return an error when UAA URL is invalid", func() {
			_, _, err := parseOpsmanAndUAAURLs("https://opsman.example.com", "smb://uaa.example.com")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("could not parse UAA target URL: error parsing URL, expected http(s) protocol but got smb"))
		})
	})
})
