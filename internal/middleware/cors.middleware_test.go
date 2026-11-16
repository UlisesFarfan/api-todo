package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCors(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusTeapot) })
	handler := Cors([]string{"http://localhost:4200"}, next)

	cases := []struct {
		name, method, origin string
		status               int
		allowed              string
	}{
		{"origen permitido", http.MethodGet, "http://localhost:4200", http.StatusTeapot, "http://localhost:4200"},
		{"preflight permitido", http.MethodOptions, "http://localhost:4200", http.StatusNoContent, "http://localhost:4200"},
		{"otro origen", http.MethodGet, "https://otro.com", http.StatusTeapot, ""},
		{"preflight de otro origen", http.MethodOptions, "https://otro.com", http.StatusTeapot, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := httptest.NewRequest(c.method, "/health", nil)
			r.Header.Set("Origin", c.origin)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			if w.Code != c.status {
				t.Errorf("status %d, quería %d", w.Code, c.status)
			}
			if got := w.Header().Get("Access-Control-Allow-Origin"); got != c.allowed {
				t.Errorf("Allow-Origin %q, quería %q", got, c.allowed)
			}
		})
	}
}
