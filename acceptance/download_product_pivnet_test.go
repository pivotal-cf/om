package acceptance

import (
	"archive/zip"
	"bytes"
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var _ = Describe("download-product command", func() {
	When("downloading from Pivnet", func() {
		var server *ghttp.Server
		var pathToHTTPSPivnet string

		AfterEach(func() {
			server.Close()
		})

		BeforeEach(func() {
			pivotalFile := createPivotalFile("[example-product,1.10.1]example*pivotal", "./fixtures/example-product.yml")
			contents, err := ioutil.ReadFile(pivotalFile)
			Expect(err).ToNot(HaveOccurred())
			modTime := time.Now()

			var fakePivnetMetadataResponse []byte

			fixtureMetadata, err := os.Open("fixtures/example-product.yml")
			defer fixtureMetadata.Close()

			Expect(err).ToNot(HaveOccurred())

			_, err = fixtureMetadata.Read(fakePivnetMetadataResponse)
			Expect(err).ToNot(HaveOccurred())

			server = ghttp.NewTLSServer()
			pathToHTTPSPivnet, err = gexec.Build("github.com/pivotal-cf/om",
				"--ldflags", fmt.Sprintf("-X github.com/pivotal-cf/om/download_clients.pivnetHost=%s", server.URL()))
			Expect(err).ToNot(HaveOccurred())

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v2/products/example-product/releases"),
					ghttp.RespondWith(http.StatusOK, `{
  "releases": [
    {
      "id": 24,
      "version": "1.10.1"
    }
  ]
}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v2/products/example-product/releases/24"),
					ghttp.RespondWith(http.StatusOK, `{"id":24}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v2/products/example-product/releases/24/product_files"),
					ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`{
  "product_files": [
  {
    "id": 1,
    "aws_object_key": "example-product.pivotal",
    "_links": {
      "download": {
        "href": "%s/api/v2/products/example-product/releases/32/product_files/21/download"
      }
    }
  }
]
}`, server.URL())),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v2/products/example-product/releases/24/file_groups"),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v2/products/example-product/releases/24/product_files/1"),
					ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`{
"product_file": {
    "id": 1,
	"_links": {
		"download": {
			"href":"%s/api/v2/products/example-product/releases/24/product_files/1/download"
		}
	}
}
}`, server.URL())),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v2/products/example-product/releases/24/product_files/1/download"),
					ghttp.RespondWith(http.StatusFound, `{}`, http.Header{"Location": {fmt.Sprintf("%s/api/v2/products/example-product/releases/24/product_files/1/download", server.URL())}}),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("HEAD", "/api/v2/products/example-product/releases/24/product_files/1/download"),
					func(w http.ResponseWriter, r *http.Request) {
						http.ServeContent(w, r, "download", modTime, bytes.NewReader(contents))
					},
				),
			)
			server.RouteToHandler("GET", "/api/v2/products/example-product/releases/24/product_files/1/download",
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v2/products/example-product/releases/24/product_files/1/download"),
					func(w http.ResponseWriter, r *http.Request) {
						http.ServeContent(w, r, "download", modTime, bytes.NewReader(contents))
					},
				),
			)
		})

		It("downloads the product", func() {
			tmpDir, err := ioutil.TempDir("", "")
			Expect(err).ToNot(HaveOccurred())
			command := exec.Command(pathToHTTPSPivnet, "download-product",
				"--pivnet-api-token", "token",
				"--file-glob", "example-product.pivotal",
				"--pivnet-product-slug", "example-product",
				"--pivnet-disable-ssl",
				"--product-version", "1.10.1",
				"--output-directory", tmpDir,
			)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(session.Err).To(gbytes.Say(`attempting to download the file.*example-product.pivotal.*from source pivnet`))
			Expect(session.Err).To(gbytes.Say(`Writing a list of downloaded artifact to download-file.json`))

			Expect(filepath.Join(tmpDir, "example-product.pivotal")).To(BeAnExistingFile())
			Expect(filepath.Join(tmpDir, "example-product.pivotal.partial")).ToNot(BeAnExistingFile())
		})
	})
})

func fileContents(paths ...string) []byte {
	path := filepath.Join(paths...)
	contents, err := ioutil.ReadFile(filepath.Join(path))
	Expect(err).ToNot(HaveOccurred())
	return contents
}

func createPivotalFile(productFileName, metadataFilename string) string {
	tempfile, err := ioutil.TempFile("", productFileName)
	Expect(err).ToNot(HaveOccurred())

	zipper := zip.NewWriter(tempfile)
	file, err := zipper.Create("metadata/props.yml")
	Expect(err).ToNot(HaveOccurred())

	contents, err := ioutil.ReadFile(metadataFilename)
	Expect(err).ToNot(HaveOccurred())

	_, err = file.Write(contents)
	Expect(err).ToNot(HaveOccurred())

	Expect(zipper.Close()).ToNot(HaveOccurred())
	return tempfile.Name()
}

func uploadGCSFile(localFile, serviceAccountKey, bucketName, objectName string) {
	f, err := os.Open(localFile)
	Expect(err).ToNot(HaveOccurred())

	defer f.Close()

	cxt := context.Background()
	creds, err := google.CredentialsFromJSON(cxt, []byte(serviceAccountKey), storage.ScopeReadWrite)
	Expect(err).ToNot(HaveOccurred())

	client, err := storage.NewClient(cxt, option.WithCredentials(creds))
	Expect(err).ToNot(HaveOccurred())

	wc := client.Bucket(bucketName).Object(objectName).NewWriter(cxt)
	_, err = io.Copy(wc, f)
	Expect(err).ToNot(HaveOccurred())

	err = wc.Close()
	Expect(err).ToNot(HaveOccurred())
}
