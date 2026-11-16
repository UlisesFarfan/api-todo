package middleware

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"api-todo/internal/base"
	"api-todo/internal/httputil"
	"api-todo/internal/security"
)

// Usuario es la fila U#<email>/PERFIL. ID es el "sub" de su cuenta de Google.
type Usuario struct {
	ID          string    `json:"id"`
	Nombre      string    `json:"nombre"`
	Email       string    `json:"email"`
	Img         string    `json:"img"`
	Rol         string    `json:"rol"`
	Creado      time.Time `json:"creado"`
	Actualizado time.Time `json:"actualizado"`
}

// Sesion es la fila S#<hash del token>/SESION.
type Sesion struct {
	Email string `json:"email"`
}

func PKUsuario(email string) string { return "U#" + email }

func PKSesion(token string) string { return "S#" + hex.EncodeToString(security.HashToken(token)) }

type claveUsuario struct{}

// Auth deja pasar solo con el token de una sesión vigente ("Authorization:
// Bearer <token>", el front lo guarda al entrar) y deja el usuario en el
// contexto (ver UsuarioDe).
func Auth(b base.Base) func(http.HandlerFunc) http.Handler {
	return func(next http.HandlerFunc) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !ok || token == "" {
				httputil.SesionInvalida(w)
				return
			}
			fila, existe, err := b.Leer(r.Context(), PKSesion(token), "SESION")
			if err != nil {
				httputil.Interno(w, err)
				return
			}
			// DynamoDB tarda en borrar las vencidas: se miran igual.
			if !existe || fila.Vence < time.Now().Unix() {
				httputil.SesionInvalida(w)
				return
			}
			var s Sesion
			var u Usuario
			if err := json.Unmarshal(fila.Datos, &s); err != nil {
				httputil.Interno(w, err)
				return
			}
			fila, existe, err = b.Leer(r.Context(), PKUsuario(s.Email), "PERFIL")
			if err == nil && existe {
				err = json.Unmarshal(fila.Datos, &u)
			}
			if err != nil {
				httputil.Interno(w, err)
				return
			}
			if !existe {
				httputil.SesionInvalida(w)
				return
			}
			next(w, r.WithContext(context.WithValue(r.Context(), claveUsuario{}, u)))
		})
	}
}

// UsuarioDe devuelve el usuario que dejó Auth.
func UsuarioDe(ctx context.Context) Usuario {
	u, _ := ctx.Value(claveUsuario{}).(Usuario)
	return u
}
