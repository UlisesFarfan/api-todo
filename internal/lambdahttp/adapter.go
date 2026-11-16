// Package lambdahttp adapts API Gateway HTTP API (payload format 2.0) events
// to net/http handlers, and back. In that format cookies travel apart from the
// headers, in "cookies", both ways.
package lambdahttp

import (
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
)

// Event is the subset of the API Gateway HTTP API v2 event that we use.
type Event struct {
	RawPath         string            `json:"rawPath"`
	RawQueryString  string            `json:"rawQueryString"`
	Headers         map[string]string `json:"headers"`
	Cookies         []string          `json:"cookies"`
	Body            string            `json:"body"`
	IsBase64Encoded bool              `json:"isBase64Encoded"`
	RequestContext  struct {
		HTTP struct {
			Method string `json:"method"`
			Path   string `json:"path"`
		} `json:"http"`
	} `json:"requestContext"`
}

// Response is the payload format 2.0 response returned to API Gateway.
type Response struct {
	StatusCode      int               `json:"statusCode"`
	Headers         map[string]string `json:"headers"`
	Cookies         []string          `json:"cookies,omitempty"`
	Body            string            `json:"body"`
	IsBase64Encoded bool              `json:"isBase64Encoded"`
}

// ToRequest builds the net/http request.
func ToRequest(ev Event) (*http.Request, error) {
	target := ev.RawPath
	if target == "" {
		target = ev.RequestContext.HTTP.Path
	}
	if ev.RawQueryString != "" {
		target += "?" + ev.RawQueryString
	}

	var body io.Reader = strings.NewReader("")
	if ev.Body != "" {
		raw := []byte(ev.Body)
		if ev.IsBase64Encoded {
			decoded, err := base64.StdEncoding.DecodeString(ev.Body)
			if err != nil {
				return nil, err
			}
			raw = decoded
		}
		body = strings.NewReader(string(raw))
	}

	req, err := http.NewRequest(ev.RequestContext.HTTP.Method, "https://api.local"+target, body)
	if err != nil {
		return nil, err
	}
	for k, v := range ev.Headers {
		req.Header.Set(k, v)
	}

	if len(ev.Cookies) > 0 {
		req.Header.Set("Cookie", strings.Join(ev.Cookies, "; "))
	}
	return req, nil
}

// Handle runs one event through h and converts the recorded response.
func Handle(h http.Handler, ev Event) (Response, error) {
	req, err := ToRequest(ev)
	if err != nil {
		return Response{}, err
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	headers := map[string]string{}
	for k, vs := range rec.Header() {
		if len(vs) > 0 && k != "Set-Cookie" {
			headers[k] = vs[0]
		}
	}
	return Response{
		StatusCode: rec.Code,
		Headers:    headers,
		Cookies:    rec.Header().Values("Set-Cookie"),
		Body:       rec.Body.String(),
	}, nil
}
