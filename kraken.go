//Package kraken is the official client's implementation for Kraken.io API
//https://github.com/kraken-io/kraken-go
package kraken

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

var (
	//ErrNoCred indicates that api_key or api_secret were not specified
	ErrNoCred = fmt.Errorf("key and secret should be specified")
)

//ResponseError contains error encountered while processing HTTP response
//with response itself
type ResponseError struct {
	s    string
	Resp *http.Response
}

func (err *ResponseError) Error() string { return err.s }

//Kraken holds api_key, api_secret, and HttpClient
type Kraken struct {
	auth map[string]string
	//If not specified default http.Client{} is used
	HTTPClient *http.Client
}

//New creates new instance of Kraken. key and secret shouldn't be empty,otherwise ErrNoCred will be returned
func New(key, secret string) (*Kraken, error) {
	if key == "" || secret == "" {
		return nil, ErrNoCred
	}
	return &Kraken{
		auth: map[string]string{
			"api_key":    key,
			"api_secret": secret,
		},
		HTTPClient: http.DefaultClient,
	}, nil
}

// URL makes request to the https://api.kraken.io/v1/url endpoint
func (kr *Kraken) URL(params map[string]interface{}) (map[string]interface{}, error) {
	dataReq, err := kr.marshalParams(params)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", "https://api.kraken.io/v1/url", bytes.NewBuffer(dataReq))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return kr.doReq(req)
}

// Upload opens a file and makes multipart request to the https://api.kraken.io/v1/upload endpoint
func (kr *Kraken) Upload(params map[string]interface{}, path string) (map[string]interface{}, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return kr.UploadReader(params, file, filepath.Base(path))
}

// UploadReader makes multipart request to the https://api.kraken.io/v1/upload endpoint
func (kr *Kraken) UploadReader(params map[string]interface{}, f io.Reader, name string) (map[string]interface{}, error) {
	pipeReader, pipeWriter := io.Pipe()
	writer := multipart.NewWriter(pipeWriter)

	req, err := http.NewRequest("POST", "https://api.kraken.io/v1/upload", pipeReader)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())
	cancelCh := make(chan struct{})
	req.Cancel = cancelCh

	errCh := make(chan error, 1)
	go func() {
		defer pipeWriter.Close()
		data, err := kr.marshalParams(params)
		if err != nil {
			errCh <- err
			close(cancelCh)
			return
		}
		dataPart, err := writer.CreateFormField("data")
		if err != nil {
			errCh <- err
			close(cancelCh)
			return
		}
		_, err = dataPart.Write(data)
		if err != nil {
			errCh <- err
			close(cancelCh)
			return
		}
		uploadPart, err := writer.CreateFormFile("upload", name)
		if err != nil {
			errCh <- err
			close(cancelCh)
			return
		}
		_, err = io.Copy(uploadPart, f)
		if err != nil {
			errCh <- err
			close(cancelCh)
			return
		}
		err = writer.Close()
		if err != nil {
			errCh <- err
			close(cancelCh)
		}
	}()

	dataResp, err := kr.doReq(req)
	if err != nil {
		select {
		case err := <-errCh:
			return nil, err
		default:
			return nil, err
		}
	}
	return dataResp, nil
}
func (kr *Kraken) marshalParams(params map[string]interface{}) ([]byte, error) {
	params["auth"] = kr.auth
	return json.Marshal(params)
}
func (kr *Kraken) doReq(req *http.Request) (map[string]interface{}, error) {
	resp, err := kr.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, &ResponseError{err.Error(), resp}
	}
	dataResp := make(map[string]interface{})
	err = json.Unmarshal(body, &dataResp)
	if err != nil {
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		return nil, &ResponseError{err.Error(), resp}
	}
	return dataResp, nil
}
