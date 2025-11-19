package download_test

import (
	"net/http"

	"github.com/pivotal-cf/go-pivnet/v7/download"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BuildRange", func() {
	Context("when an even content length is provided", func() {
		var cr download.Ranger

		BeforeEach(func() {
			cr = download.NewRanger(10)
		})

		It("returns an set of byte ranges", func() {
			contentLength := int64(100)
			r, err := cr.BuildRange(contentLength)
			Expect(err).NotTo(HaveOccurred())

			Expect(r).To(Equal([]download.Range{
				{Lower: 0, Upper: 9, HTTPHeader: http.Header{"Range": []string{"bytes=0-9"}}},
				{Lower: 10, Upper: 19, HTTPHeader: http.Header{"Range": []string{"bytes=10-19"}}},
				{Lower: 20, Upper: 29, HTTPHeader: http.Header{"Range": []string{"bytes=20-29"}}},
				{Lower: 30, Upper: 39, HTTPHeader: http.Header{"Range": []string{"bytes=30-39"}}},
				{Lower: 40, Upper: 49, HTTPHeader: http.Header{"Range": []string{"bytes=40-49"}}},
				{Lower: 50, Upper: 59, HTTPHeader: http.Header{"Range": []string{"bytes=50-59"}}},
				{Lower: 60, Upper: 69, HTTPHeader: http.Header{"Range": []string{"bytes=60-69"}}},
				{Lower: 70, Upper: 79, HTTPHeader: http.Header{"Range": []string{"bytes=70-79"}}},
				{Lower: 80, Upper: 89, HTTPHeader: http.Header{"Range": []string{"bytes=80-89"}}},
				{Lower: 90, Upper: 99, HTTPHeader: http.Header{"Range": []string{"bytes=90-99"}}},
			}))
		})
	})

	Context("when an odd content length is provided", func() {
		var cr download.Ranger

		BeforeEach(func() {
			cr = download.NewRanger(10)
		})

		It("returns the byte ranges", func() {
			contentLength := int64(101)
			r, err := cr.BuildRange(contentLength)
			Expect(err).NotTo(HaveOccurred())

			Expect(r).To(Equal([]download.Range{
				{Lower: 0, Upper: 9, HTTPHeader: http.Header{"Range": []string{"bytes=0-9"}}},
				{Lower: 10, Upper: 19, HTTPHeader: http.Header{"Range": []string{"bytes=10-19"}}},
				{Lower: 20, Upper: 29, HTTPHeader: http.Header{"Range": []string{"bytes=20-29"}}},
				{Lower: 30, Upper: 39, HTTPHeader: http.Header{"Range": []string{"bytes=30-39"}}},
				{Lower: 40, Upper: 49, HTTPHeader: http.Header{"Range": []string{"bytes=40-49"}}},
				{Lower: 50, Upper: 59, HTTPHeader: http.Header{"Range": []string{"bytes=50-59"}}},
				{Lower: 60, Upper: 69, HTTPHeader: http.Header{"Range": []string{"bytes=60-69"}}},
				{Lower: 70, Upper: 79, HTTPHeader: http.Header{"Range": []string{"bytes=70-79"}}},
				{Lower: 80, Upper: 89, HTTPHeader: http.Header{"Range": []string{"bytes=80-89"}}},
				{Lower: 90, Upper: 100, HTTPHeader: http.Header{"Range": []string{"bytes=90-100"}}},
			}))
		})
	})

	Context("when content length is less than the number of hunks", func() {
		var cr download.Ranger

		BeforeEach(func() {
			cr = download.NewRanger(10)
		})

		It("returns as many byte ranges as possible", func() {
			contentLength := int64(3)
			r, err := cr.BuildRange(contentLength)
			Expect(err).NotTo(HaveOccurred())

			Expect(r).To(Equal([]download.Range{
				{Lower: 0, Upper: 2, HTTPHeader: http.Header{"Range": []string{"bytes=0-2"}}},
			}))
		})

		It("returns as many byte ranges as possible", func() {
			contentLength := int64(9)
			r, err := cr.BuildRange(contentLength)
			Expect(err).NotTo(HaveOccurred())

			Expect(r).To(Equal([]download.Range{
				{Lower: 0, Upper: 1, HTTPHeader: http.Header{"Range": []string{"bytes=0-1"}}},
				{Lower: 2, Upper: 3, HTTPHeader: http.Header{"Range": []string{"bytes=2-3"}}},
				{Lower: 4, Upper: 5, HTTPHeader: http.Header{"Range": []string{"bytes=4-5"}}},
				{Lower: 6, Upper: 8, HTTPHeader: http.Header{"Range": []string{"bytes=6-8"}}},
			}))
		})
	})

	Context("when an error occurs", func() {
		Context("when the content length is zero", func() {
			var cr download.Ranger

			BeforeEach(func() {
				cr = download.NewRanger(10)
			})

			It("returns an error", func() {
				contentLength := int64(0)
				_, err := cr.BuildRange(contentLength)
				Expect(err).To(MatchError("content length cannot be zero"))
			})
		})
	})
})
