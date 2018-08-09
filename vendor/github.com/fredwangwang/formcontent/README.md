This helper library allows you upload large file without loading the file content into the memory.

[non-threaded](https://github.com/fredwangwang/formcontent/tree/non-threaded)

[example](https://gist.github.com/fredwangwang/8c86a34da27c2bc9ded3a38968576e4a):
```go
package main

import (
	"github.com/fredwangwang/formcontent"
	"net/http"
	"log"
	"io/ioutil"
	"fmt"
)

func main() {
	var err error

	client := &http.Client{}

	uri := "https://full.qualified.domain.name/post/route"

	// initialize multipart form
	multipart := formcontent.NewForm()

	// add fields to the form
	if err = multipart.AddFile("file", "/path/to/the/file.ext"); err != nil {
		log.Fatal(err)
	}
	if err = multipart.AddField("attribute1", "value1"); err != nil {
		log.Fatal(err)
	}
	if err = multipart.AddField("attribute2", "value2"); err != nil {
		log.Fatal(err)
	}

	// finish editing the form
	form := multipart.Finalize()

	// create a request
	req, err := http.NewRequest("POST", uri, form.Content)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", form.ContentType)
	req.ContentLength = form.ContentLength

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	respContent, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.StatusCode)
	fmt.Println(resp.Header)
	fmt.Println(string(respContent))
}

```