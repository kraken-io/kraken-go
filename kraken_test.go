package kraken_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kraken-io/kraken-go"
)

// redirect sends every request the client makes to srv instead of the real API,
// so the tests need no network and no credentials. Kraken exposes HTTPClient,
// so this needs no change to the library itself.
type redirect struct{ to *url.URL }

func (r redirect) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = r.to.Scheme
	req.URL.Host = r.to.Host
	return http.DefaultTransport.RoundTrip(req)
}

// newClient returns a Kraken pointed at srv.
func newClient(t *testing.T, srv *httptest.Server) *kraken.Kraken {
	t.Helper()
	kr, err := kraken.New("api_key", "api_secret")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse %q: %v", srv.URL, err)
	}
	kr.HTTPClient = &http.Client{Transport: redirect{to: u}}
	return kr
}

func TestNewRequiresBothCredentials(t *testing.T) {
	for _, tc := range []struct{ key, secret string }{
		{"", ""},
		{"api_key", ""},
		{"", "api_secret"},
	} {
		if _, err := kraken.New(tc.key, tc.secret); err != kraken.ErrNoCred {
			t.Errorf("New(%q, %q) = %v, want ErrNoCred", tc.key, tc.secret, err)
		}
	}
	if _, err := kraken.New("api_key", "api_secret"); err != nil {
		t.Errorf("New with both credentials: %v", err)
	}
}

func TestURLSendsParamsAndAuth(t *testing.T) {
	var got map[string]interface{}
	var path, method, contentType string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path, method = r.URL.Path, r.Method
		contentType = r.Header.Get("Content-Type")
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"success":true,"kraked_url":"https://dl.kraken.io/x.jpg","saved_bytes":1024}`)
	}))
	defer srv.Close()

	data, err := newClient(t, srv).URL(map[string]interface{}{
		"wait": true,
		"url":  "https://example.com/a.jpg",
	})
	if err != nil {
		t.Fatalf("URL: %v", err)
	}

	if method != http.MethodPost {
		t.Errorf("method = %q, want POST", method)
	}
	if path != "/v1/url" {
		t.Errorf("path = %q, want /v1/url", path)
	}
	if contentType != "application/json" {
		t.Errorf("content-type = %q, want application/json", contentType)
	}

	auth, ok := got["auth"].(map[string]interface{})
	if !ok {
		t.Fatalf("auth missing from request body: %#v", got)
	}
	if auth["api_key"] != "api_key" || auth["api_secret"] != "api_secret" {
		t.Errorf("auth = %#v, want the credentials from New", auth)
	}
	if got["url"] != "https://example.com/a.jpg" {
		t.Errorf("url = %v, want it passed through", got["url"])
	}
	if data["success"] != true {
		t.Errorf("success = %v, want true", data["success"])
	}
	if data["kraked_url"] != "https://dl.kraken.io/x.jpg" {
		t.Errorf("kraked_url = %v", data["kraked_url"])
	}
}

func TestURLReturnsAPIFailureAsData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"success":false,"message":"Unknown API Key"}`)
	}))
	defer srv.Close()

	// A well-formed failure is data, not a transport error.
	data, err := newClient(t, srv).URL(map[string]interface{}{"url": "https://example.com/a.jpg"})
	if err != nil {
		t.Fatalf("URL: %v", err)
	}
	if data["success"] != false {
		t.Errorf("success = %v, want false", data["success"])
	}
	if data["message"] != "Unknown API Key" {
		t.Errorf("message = %v", data["message"])
	}
}

func TestNonJSONResponseIsResponseError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		io.WriteString(w, "<html>502</html>")
	}))
	defer srv.Close()

	_, err := newClient(t, srv).URL(map[string]interface{}{"url": "https://example.com/a.jpg"})
	if err == nil {
		t.Fatal("want an error for a non-JSON body")
	}
	respErr, ok := err.(*kraken.ResponseError)
	if !ok {
		t.Fatalf("err is %T, want *kraken.ResponseError", err)
	}
	if respErr.Resp == nil {
		t.Fatal("ResponseError.Resp is nil, want the response for inspection")
	}
	if respErr.Resp.StatusCode != http.StatusBadGateway {
		t.Errorf("status = %d, want 502", respErr.Resp.StatusCode)
	}
	// The body is restored so a caller can still read it.
	body, err := io.ReadAll(respErr.Resp.Body)
	if err != nil {
		t.Fatalf("read restored body: %v", err)
	}
	if !strings.Contains(string(body), "502") {
		t.Errorf("restored body = %q, want the original payload", body)
	}
}

func TestUploadSendsMultipart(t *testing.T) {
	var fileName, fileBody string
	var data map[string]interface{}
	var path string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Errorf("ParseMultipartForm: %v", err)
			return
		}
		if err := json.Unmarshal([]byte(r.FormValue("data")), &data); err != nil {
			t.Errorf("decode data field: %v", err)
		}
		f, hdr, err := r.FormFile("upload")
		if err != nil {
			t.Errorf("FormFile(upload): %v", err)
			return
		}
		defer f.Close()
		fileName = hdr.Filename
		b, _ := io.ReadAll(f)
		fileBody = string(b)

		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"success":true,"file_name":"photo.jpg"}`)
	}))
	defer srv.Close()

	dir := t.TempDir()
	img := filepath.Join(dir, "photo.jpg")
	if err := os.WriteFile(img, []byte("not a real jpeg"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	got, err := newClient(t, srv).Upload(map[string]interface{}{"wait": true}, img)
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}

	if path != "/v1/upload" {
		t.Errorf("path = %q, want /v1/upload", path)
	}
	if fileName != "photo.jpg" {
		t.Errorf("filename = %q, want the basename of the path", fileName)
	}
	if fileBody != "not a real jpeg" {
		t.Errorf("uploaded body = %q", fileBody)
	}
	if auth, ok := data["auth"].(map[string]interface{}); !ok || auth["api_key"] != "api_key" {
		t.Errorf("auth missing from the data field: %#v", data)
	}
	if got["success"] != true {
		t.Errorf("success = %v, want true", got["success"])
	}
}

func TestUploadMissingFile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("no request should be made when the file cannot be opened")
	}))
	defer srv.Close()

	_, err := newClient(t, srv).Upload(map[string]interface{}{}, filepath.Join(t.TempDir(), "nope.jpg"))
	if err == nil {
		t.Fatal("want an error for a missing file")
	}
	if !os.IsNotExist(err) {
		t.Errorf("err = %v, want a not-exist error", err)
	}
}

func TestUploadReaderUsesGivenName(t *testing.T) {
	var fileName string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Errorf("ParseMultipartForm: %v", err)
			return
		}
		_, hdr, err := r.FormFile("upload")
		if err != nil {
			t.Errorf("FormFile(upload): %v", err)
			return
		}
		fileName = hdr.Filename
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"success":true}`)
	}))
	defer srv.Close()

	_, err := newClient(t, srv).UploadReader(
		map[string]interface{}{}, strings.NewReader("bytes"), "given-name.png")
	if err != nil {
		t.Fatalf("UploadReader: %v", err)
	}
	if fileName != "given-name.png" {
		t.Errorf("filename = %q, want given-name.png", fileName)
	}
}
