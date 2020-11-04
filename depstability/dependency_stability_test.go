package depstability_test

import (
	"io/ioutil"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dependency Topology", func() {
	It("is unchanged, allowing reuse of the 4.3.0 OSL file", func() {
		currentDepnames, err := listDepnamesFromGoSum()
		Expect(err).NotTo(HaveOccurred())
		oldDepnames, err := listDepnamesFromRecords()
		Expect(err).NotTo(HaveOccurred())

		Expect(currentDepnames).To(ConsistOf(oldDepnames), "See readme for guidance on updating fixtures")
	})
})

func listDepnamesFromGoSum() (deplist []string, err error) {
	gosum, err := ioutil.ReadFile("../go.sum")
	if err != nil {
		return
	}
	trimmedGoSum := strings.TrimSpace(string(gosum))
	splitGoSum := strings.Split(trimmedGoSum, "\n")
	var deplistWithDuplication []string
	for _, line := range splitGoSum {
		splitLine := strings.Split(line, " ")
		deplistWithDuplication = append(deplistWithDuplication, splitLine[0])
	}
	for _, line := range deplistWithDuplication {
		if len(deplist) == 0 || deplist[len(deplist)-1] != line {
			deplist = append(deplist, line)
		}
	}
	return
}

func listDepnamesFromRecords() (deplist []string, err error) {
	depRecords, err := ioutil.ReadFile("records/depnames-7.0.0.txt")
	trimmedDepRecords := strings.TrimSpace(string(depRecords))
	deplist = strings.Split(trimmedDepRecords, "\n")
	return
}
