package fetch

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type (
	Request struct {
		*http.Request
		Client *http.Client
	}

	RequestOptions struct {
		Method  string
		Headers http.Header
		Params  *[][]string
		Body    io.Reader
	}

	Response struct {
		*http.Response
	}
)

func NewRequest(ctx context.Context, requestUrl string, requestOptions ...*RequestOptions) (*Request, error) {
	client := http.Client{}

	options := RequestOptions{}
	if len(requestOptions) > 0 {
		options = *requestOptions[0]
	}

	if options.Method == "" {
		options.Method = "GET"
	}

	if options.Params != nil {
		u, err := url.Parse(requestUrl)
		query := u.Query()
		if err != nil {
			return nil, err
		}
		for _, p := range *options.Params {
			if len(p) != 2 {
				return nil, errors.New(`URL Params must be [["key1", "value1"], ["key2", "value2"]]`)
			}
			if strings.Trim(p[1], " ") == "" {
				continue
			}
			query.Add(p[0], p[1])
		}
		u.RawQuery = query.Encode()
		requestUrl = u.String()
	}

	req, err := http.NewRequestWithContext(ctx, options.Method, requestUrl, options.Body)
	if err != nil {
		return nil, err
	}

	if options.Headers != nil {
		req.Header = options.Headers
	}

	return &Request{Request: req, Client: &client}, nil
}

func (r *Request) WithTransport(transport *http.Transport) *Request {
	r.Client.Transport = transport
	return r
}

func (r *Request) WithoutRedirects() *Request {
	r.Client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return r
}

func (r *Request) WithTimeout(timeout time.Duration) *Request {
	r.Client.Timeout = timeout
	return r
}

func (r *Request) WithHeaders(Headers http.Header) *Request {
	r.Header = Headers
	return r
}

func (r *Request) JSON(payload any) error {
	r.Header.Set("Content-Type", "application/json; charset=UTF-8")

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	r.Body = io.NopCloser(bytes.NewReader(body))

	return nil
}

func (r *Request) Do() (*Response, error) {
	resp, err := r.Client.Do(r.Request)
	if err != nil {
		return nil, err
	}

	return &Response{resp}, nil
}

func (r *Response) EffectiveURL() string {
	return r.Request.URL.String()
}

func (r *Response) ContentType() string {
	return r.Header.Get("Content-Type")
}

func (r *Response) LastModified() string {
	if r.Header.Get("Expires") == "0" {
		return ""
	}
	return r.Header.Get("Last-Modified")
}

func (r *Response) ETag() string {
	if r.Header.Get("Expires") == "0" {
		return ""
	}
	return r.Header.Get("ETag")
}

func (r *Response) IsModified(lastEtagValue, lastModifiedValue string) bool {
	if r.StatusCode == http.StatusNotModified {
		return false
	}

	if r.ETag() != "" && r.ETag() == lastEtagValue {
		return false
	}

	if r.LastModified() != "" && r.LastModified() == lastModifiedValue {
		return false
	}

	return true
}

func (r *Response) Close() {
	if r != nil && r.Body != nil {
		r.Body.Close()
	}
}

func (r *Response) OK() bool {
	return (r.StatusCode/100) >= 2 && (r.StatusCode/100) < 4
}

func (r *Response) ReadBody(maxBodySize int64) ([]byte, error) {
	var buffer []byte
	var err error
	if maxBodySize > 0 {
		limitedReader := http.MaxBytesReader(nil, r.Body, maxBodySize)
		buffer, err = io.ReadAll(limitedReader)
	} else {
		buffer, err = io.ReadAll(r.Body)
	}

	if err != nil && err != io.EOF {
		if err, ok := err.(*http.MaxBytesError); ok {
			return nil, fmt.Errorf("fetcher: response body too large: %d bytes", err.Limit)
		}
		return nil, fmt.Errorf("fetcher: unable to read response body: %w", err)
	}

	if len(buffer) == 0 {
		return nil, fmt.Errorf("fetcher: empty response body")
	}

	return buffer, nil
}

func (r *Response) Text() (string, error) {
	buffer, err := r.ReadBody(0)
	if err != nil {
		return "", err
	}

	return string(buffer), nil
}

func (r *Response) JSON(v any) error {
	buffer, err := r.ReadBody(0)
	if err != nil {
		return err
	}

	err = json.Unmarshal(buffer, v)
	if err != nil {
		return err
	}

	return nil
}

func (r *Response) XML(v any) error {
	buffer, err := r.ReadBody(0)
	if err != nil {
		return err
	}

	err = xml.Unmarshal(buffer, v)
	if err != nil {
		return err
	}

	return nil
}
