package acceptance

import (
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"

	"github.com/onsi/gomega/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MatchSHASumOfMatcher", func() {
	Describe("Match", func() {
		Context("when the files have the same contents", func() {
			It("returns true", func() {
				tempDir, err := ioutil.TempDir("", "")
				Expect(err).NotTo(HaveOccurred())

				file1 := filepath.Join(tempDir, "file-1")
				file2 := filepath.Join(tempDir, "file-2")

				err = ioutil.WriteFile(file1, []byte("file contents"), 0644)
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(file2, []byte("file contents"), 0644)
				Expect(err).NotTo(HaveOccurred())

				matcher := MatchSHASumOf(file1)
				success, err := matcher.Match(file2)
				Expect(err).NotTo(HaveOccurred())
				Expect(success).To(BeTrue())
			})
		})

		Context("when the files do not have the same contents", func() {
			It("returns true", func() {
				tempDir, err := ioutil.TempDir("", "")
				Expect(err).NotTo(HaveOccurred())

				file1 := filepath.Join(tempDir, "file-1")
				file2 := filepath.Join(tempDir, "file-2")

				err = ioutil.WriteFile(file1, []byte("some contents"), 0644)
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(file2, []byte("other contents"), 0644)
				Expect(err).NotTo(HaveOccurred())

				matcher := MatchSHASumOf(file1)
				success, err := matcher.Match(file2)
				Expect(err).NotTo(HaveOccurred())
				Expect(success).To(BeFalse())
			})
		})
	})
})

func MatchSHASumOf(expected interface{}) types.GomegaMatcher {
	return &matchSHASumOfMatcher{
		expected: expected,
	}
}

type matchSHASumOfMatcher struct {
	expected interface{}
}

func (matcher *matchSHASumOfMatcher) Match(actual interface{}) (success bool, err error) {
	expectedFile, err := os.Open(matcher.expected.(string))
	if err != nil {
		return false, err
	}

	actualFile, err := os.Open(actual.(string))
	if err != nil {
		return false, err
	}

	expectedHash := sha1.New()
	actualHash := sha1.New()

	_, err = io.Copy(expectedHash, expectedFile)
	if err != nil {
		return false, err
	}

	_, err = io.Copy(actualHash, actualFile)
	if err != nil {
		return false, err
	}

	return reflect.DeepEqual(actualHash.Sum(nil), expectedHash.Sum(nil)), nil
}

func (matcher *matchSHASumOfMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nto contain the same contents as\n\t%#v", actual, matcher.expected)
}

func (matcher *matchSHASumOfMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nnot to contain the same contents as\n\t%#v", actual, matcher.expected)
}
