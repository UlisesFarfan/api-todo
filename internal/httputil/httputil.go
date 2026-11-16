// Package httputil: respuestas JSON y lectura del cuerpo, igual en toda la API.
// Todas las respuestas tienen la forma que ya usa el front (app-todo, en
// Vercel): {code, status, message, data}; en un error, message es para mostrar.
package httputil

import (
	"encoding/json"
	"log"
	"net/http"
)

type respuesta struct {
	Code    int    `json:"code"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func JSON(w http.ResponseWriter, status int, mensaje string, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(respuesta{status, http.StatusText(status), mensaje, data}); err != nil {
		log.Printf("respuesta JSON: %v", err)
	}
}

func Error(w http.ResponseWriter, status int, mensaje string) { JSON(w, status, mensaje, nil) }

func SesionInvalida(w http.ResponseWriter) {
	Error(w, http.StatusUnauthorized, "Tu sesión terminó. Volvé a entrar.")
}

// Interno registra el error y responde 500 sin mostrar el detalle.
func Interno(w http.ResponseWriter, err error) {
	log.Printf("error interno: %v", err)
	Error(w, http.StatusInternalServerError, "Algo salió mal. Probá de nuevo.")
}

// LeerJSON lee un cuerpo de hasta 64 KB. Acepta campos de más porque el front
// los manda (signal, columnId). Si falla, ya respondió 400 y devuelve false.
func LeerJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(v); err != nil {
		Error(w, http.StatusBadRequest, "Los datos enviados no son válidos.")
		return false
	}
	return true
}
