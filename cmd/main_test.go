package cmd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("main", func() {
	It("Should parse Opsman target URL with https protocol", func() {
		opsmanURL, uaaURL, err := parseTargetURLs("https://opsman.example.com", "")
		Expect(err).ToNot(HaveOccurred())
		Expect(opsmanURL.String()).To(Equal("https://opsman.example.com"))
		Expect(uaaURL.String()).To(Equal("https://opsman.example.com/uaa"))
	})

	It("Should parse Opsman target URL with http protocol", func() {
		opsmanURL, uaaURL, err := parseTargetURLs("http://opsman.example.com", "")
		Expect(err).ToNot(HaveOccurred())
		Expect(opsmanURL.String()).To(Equal("http://opsman.example.com"))
		Expect(uaaURL.String()).To(Equal("http://opsman.example.com/uaa"))
	})

	It("Should parse Opsman target URL without protocol", func() {
		opsmanURL, uaaURL, err := parseTargetURLs("opsman.example.com", "")
		Expect(err).ToNot(HaveOccurred())
		Expect(opsmanURL.String()).To(Equal("https://opsman.example.com"))
		Expect(uaaURL.String()).To(Equal("https://opsman.example.com/uaa"))
	})

	It("Should parse Opsman and UAA target URL with https protocol", func() {
		opsmanURL, uaaURL, err := parseTargetURLs("https://opsman.example.com", "https://uaa.example.com")
		Expect(err).ToNot(HaveOccurred())
		Expect(opsmanURL.String()).To(Equal("https://opsman.example.com"))
		Expect(uaaURL.String()).To(Equal("https://uaa.example.com"))
	})

	It("Should parse Opsman and UAA target URL with http protocol", func() {
		opsmanURL, uaaURL, err := parseTargetURLs("http://opsman.example.com", "http://uaa.example.com")
		Expect(err).ToNot(HaveOccurred())
		Expect(opsmanURL.String()).To(Equal("http://opsman.example.com"))
		Expect(uaaURL.String()).To(Equal("http://uaa.example.com"))
	})

	It("Should parse Opsman and UAA target URL without protocol", func() {
		opsmanURL, uaaURL, err := parseTargetURLs("opsman.example.com", "uaa.example.com")
		Expect(err).ToNot(HaveOccurred())
		Expect(opsmanURL.String()).To(Equal("https://opsman.example.com"))
		Expect(uaaURL.String()).To(Equal("https://uaa.example.com"))
	})

	It("Should return flag required error when Opsman target URL when empty", func() {
		_, _, err := parseTargetURLs("", "")
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError("could not parse Opsman target URL: target flag is required, run `om help` for more info"))
	})

	It("Should not parse Opsman target URL with incorrect protocol", func() {
		_, _, err := parseTargetURLs("smb://opsman.example.com", "")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("error parsing target, expected http(s) protocol but got: smb"))
	})

	It("Should not parse Opsman target URL with bad URL", func() {
		_, _, err := parseTargetURLs("a bad\\url", "")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("could not parse Opsman target URL"))
	})

	It("Should not parse UAA target URL with bad URL", func() {
		_, _, err := parseTargetURLs("opsman.example.com", "a bad\\url")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("could not parse UAA target URL"))
	})
})
