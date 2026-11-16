package lambdahttp_test

import (
	"net/http"
	"testing"

	"api-todo/internal/lambdahttp"
)

func TestQueryStringIsPreserved(t *testing.T) {
	var ev lambdahttp.Event
	ev.RawPath = "/api/tasks"
	ev.RawQueryString = "status=done"
	ev.RequestContext.HTTP.Method = "GET"
	req, err := lambdahttp.ToRequest(ev)
	if err != nil {
		t.Fatal(err)
	}
	if req.URL.RawQuery != "status=done" || req.URL.Path != "/api/tasks" {
		t.Fatalf("url = %s", req.URL.String())
	}
}

// Las cookies llegan en "cookies" y vuelven en "cookies", no en los headers.
func TestCookiesBothWays(t *testing.T) {
	var ev lambdahttp.Event
	ev.RawPath = "/api/me"
	ev.RequestContext.HTTP.Method = "GET"
	ev.Cookies = []string{"session=abc", "theme=dark"}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("session")
		if err != nil || c.Value != "abc" {
			t.Errorf("session cookie = %v, %v", c, err)
		}
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "nuevo"})
		http.SetCookie(w, &http.Cookie{Name: "otra", Value: "x"})
	})
	resp, err := lambdahttp.Handle(h, ev)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Cookies) != 2 || resp.Cookies[0] != "session=nuevo" || resp.Headers["Set-Cookie"] != "" {
		t.Fatalf("cookies = %v, headers = %v", resp.Cookies, resp.Headers)
	}
}
