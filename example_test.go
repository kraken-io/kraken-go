package kraken_test

import (
	"log"
	"os"
	"strings"

	"github.com/kraken-io/kraken-go"
)

// Examples without an "Output:" comment are compiled but not run, so these keep
// the README honest: if the documented usage stops compiling, CI fails.

func ExampleKraken_URL() {
	kr, err := kraken.New("your-api-key", "your-api-secret")
	if err != nil {
		log.Fatal(err)
	}

	data, err := kr.URL(map[string]interface{}{
		"wait":  true,
		"lossy": true,
		"url":   "https://example.com/file.jpg",
	})
	if err != nil {
		log.Fatal(err)
	}

	if data["success"] != true {
		log.Println("failed:", data["message"])
		return
	}
	log.Println("optimized:", data["kraked_url"])
}

func ExampleKraken_Upload() {
	kr, err := kraken.New("your-api-key", "your-api-secret")
	if err != nil {
		log.Fatal(err)
	}

	data, err := kr.Upload(map[string]interface{}{
		"wait":  true,
		"lossy": true,
	}, "./photo.jpg")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(data["kraked_url"])
}

func ExampleKraken_UploadReader() {
	kr, err := kraken.New("your-api-key", "your-api-secret")
	if err != nil {
		log.Fatal(err)
	}

	// Any io.Reader works — an HTTP body, an S3 object, an in-memory buffer.
	r := strings.NewReader("...image bytes...")

	data, err := kr.UploadReader(map[string]interface{}{"wait": true}, r, "photo.jpg")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(data["kraked_url"])
}

// Resizing, converting and storage are all just API parameters.
func ExampleKraken_URL_resize() {
	kr, _ := kraken.New(os.Getenv("KRAKEN_API_KEY"), os.Getenv("KRAKEN_API_SECRET"))

	data, err := kr.URL(map[string]interface{}{
		"wait":  true,
		"lossy": true,
		"url":   "https://example.com/file.jpg",
		"resize": map[string]interface{}{
			"width":    800,
			"height":   600,
			"strategy": "fit",
		},
		"convert": map[string]interface{}{
			"format": "webp",
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Println(data["kraked_url"])
}

// A malformed response comes back as *ResponseError, which carries the
// *http.Response so the status and body can be inspected.
func ExampleResponseError() {
	kr, _ := kraken.New("your-api-key", "your-api-secret")

	_, err := kr.URL(map[string]interface{}{"url": "https://example.com/file.jpg"})
	if respErr, ok := err.(*kraken.ResponseError); ok {
		log.Println("status:", respErr.Resp.StatusCode)
		log.Println("error:", respErr.Error())
	}
}
