// Package router arma la API entera. Lo usan el servidor local (cmd/api), el
// Lambda (cmd/lambda) y los tests.
package router

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	"api-todo/internal/base"
	"api-todo/internal/handlers"
	"api-todo/internal/middleware"
)

// BaseDeEntorno: la tabla TABLA de DynamoDB; sin TABLA, una base en memoria (solo para local).
func BaseDeEntorno(ctx context.Context) (base.Base, error) {
	tabla := os.Getenv("TABLA")
	if tabla == "" {
		log.Print("sin TABLA: los datos quedan en memoria y se pierden al cortar la API")
		return base.NuevaMemoria(), nil
	}
	return base.NuevaDynamo(ctx, tabla)
}

// OrigenesDeEntorno: ALLOWED_ORIGINS, separados por coma (el front en Vercel y en local).
func OrigenesDeEntorno() []string {
	var origenes []string
	for _, o := range strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",") {
		if o = strings.TrimSpace(o); o != "" {
			origenes = append(origenes, o)
		}
	}
	return origenes
}

// GoogleDeEntorno: el verificador del login con el client ID de Google del front (GOOGLE_CLIENT_ID).
func GoogleDeEntorno() (handlers.VerificarGoogle, error) {
	id := os.Getenv("GOOGLE_CLIENT_ID")
	if id == "" {
		return nil, errors.New("falta GOOGLE_CLIENT_ID (el mismo VITE_CLIENT_ID del front)")
	}
	return handlers.VerificadorGoogle(id), nil
}

func New(b base.Base, origenes []string, verificar handlers.VerificarGoogle) http.Handler {
	h := handlers.New(b, verificar)
	con := middleware.Auth(b)
	mux := http.NewServeMux()

	// health: la API está arriba y llega a la base.
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		if _, _, err := b.Leer(r.Context(), "health", "health"); err != nil {
			log.Printf("health: %v", err)
			http.Error(w, "sin conexión con la base", http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("POST /api/authentication/google", h.Google)
	mux.Handle("GET /api/authentication", con(h.Yo))
	mux.Handle("POST /api/authentication/logout", con(h.Logout))

	mux.Handle("GET /api/work_space", con(h.Tableros))
	mux.Handle("POST /api/work_space", con(h.CrearTablero))
	mux.Handle("PATCH /api/work_space", con(h.RenombrarTablero))
	mux.Handle("PUT /api/work_space", con(h.OrdenarColumnas))
	mux.Handle("GET /api/work_space/{id}", con(h.Tablero))
	mux.Handle("DELETE /api/work_space/{id}", con(h.BorrarTablero))

	mux.Handle("POST /api/column", con(h.CrearColumna))
	mux.Handle("PATCH /api/column", con(h.RenombrarColumna))
	mux.Handle("PUT /api/column", con(h.OrdenarNotas))
	mux.Handle("DELETE /api/column/{id}", con(h.BorrarColumna))

	mux.Handle("POST /api/note", con(h.CrearNota))
	mux.Handle("PATCH /api/note", con(h.EditarNota))
	mux.Handle("DELETE /api/note/{id}", con(h.BorrarNota))

	return middleware.Cors(origenes, sinBarraFinal(mux))
}

// sinBarraFinal: el front llama a veces "/api/work_space/" y a veces "/api/work_space".
func sinBarraFinal(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p := r.URL.Path; len(p) > 1 && strings.HasSuffix(p, "/") {
			r.URL.Path, r.URL.RawPath = strings.TrimSuffix(p, "/"), ""
		}
		next.ServeHTTP(w, r)
	})
}
