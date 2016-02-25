package kraken_test

import (
	"testing"

	"github.com/kraken-io/kraken-go"
)

func TestErrors(t *testing.T) {
	kr, err := kraken.New("api_key", "")
	if err != kraken.ErrNoCred {
		t.Fatal("expected ErrNoAuth error")
	}

	kr, err = kraken.New("api_key", "api_secret")
	if err != nil {
		t.Fatal(err)
	}

	imgPath := "./notexist.jpeg"
	params := map[string]interface{}{}
	_, err = kr.Upload(params, imgPath)
	if err == nil {
		t.Fatal("Should throw error")
	}
}

func TestURL(t *testing.T) {
	kr, err := kraken.New("api_key", "api_secret")
	if err != nil {
		t.Fatal(err)
	}
	params := map[string]interface{}{
		"wait": true,
		"url":  "https://www.planwallpaper.com/static/images/canberra_hero_image_JiMVvYU.jpg",
		"resize": map[string]interface{}{
			"width":    100,
			"height":   75,
			"strategy": "crop",
		},
	}
	data, err := kr.URL(params)
	if err != nil {
		t.Fatal(err)
	}
	if data["success"] == true {
		t.Fatal("success's value should be false")
	}
	if data["message"] != "Unnknown API Key. Please check your API key and try again." {
		t.Fatal("Unexpected message ", data["message"])
	}
}
func TestUpload(t *testing.T) {
	kr, err := kraken.New("api_key", "api_secret")
	if err != nil {
		t.Fatal(err)
	}
	params := map[string]interface{}{
		"wait": true,
		"resize": map[string]interface{}{
			"width":    100,
			"height":   75,
			"strategy": "crop",
		},
	}
	imgPath := "./img.jpeg"
	data, err := kr.Upload(params, imgPath)
	if err != nil {
		t.Fatal(err)
	}
	if data["success"] == true {
		t.Fatal("success's value should be false")
	}
	if data["message"] != "Unnknown API Key. Please check your API key and try again." {
		t.Fatal("Unexpected message ", data["message"])
	}
}
