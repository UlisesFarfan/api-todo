package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"api-todo/internal/base"
	"api-todo/internal/httputil"
	"api-todo/internal/middleware"
	"api-todo/internal/security"
)

const duracionSesion = 7 * 24 * time.Hour

// usuarioJSON: lo que sale de un usuario hacia el front.
type usuarioJSON struct {
	ID          string    `json:"_id"`
	Nombre      string    `json:"name"`
	Email       string    `json:"email"`
	Img         string    `json:"img"`
	Rol         string    `json:"rol"`
	Creado      time.Time `json:"created_at"`
	Actualizado time.Time `json:"update_at"`
}

func aJSON(u middleware.Usuario) usuarioJSON {
	return usuarioJSON{u.ID, u.Nombre, u.Email, u.Img, u.Rol, u.Creado, u.Actualizado}
}

// POST /api/authentication/google {credential}: entra con la cuenta de Google
// (la primera vez la crea) y devuelve el token de la sesión y el usuario.
func (h *Handlers) Google(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Credential string `json:"credential"`
	}
	if !httputil.LeerJSON(w, r, &body) {
		return
	}
	ctx := r.Context()
	perfil, err := h.verificar(ctx, body.Credential)
	if err != nil {
		log.Printf("login con Google: %v", err)
		httputil.Error(w, http.StatusUnauthorized, "No pudimos verificar tu cuenta de Google. Probá de nuevo.")
		return
	}
	u, err := h.usuarioDeGoogle(r, perfil)
	if err != nil {
		responderError(w, err)
		return
	}
	token, _ := security.NuevoToken()
	datos, _ := json.Marshal(middleware.Sesion{Email: u.Email})
	sesion := base.Fila{PK: middleware.PKSesion(token), SK: "SESION", Datos: datos, Vence: time.Now().Add(duracionSesion).Unix()}
	if err := h.base.Escribir(ctx, base.Cambio{Fila: sesion}); err != nil {
		httputil.Interno(w, err)
		return
	}
	httputil.JSON(w, http.StatusOK, "Entraste.", map[string]any{"token_type": "Bearer", "token": token, "user": aJSON(u)})
}

// usuarioDeGoogle devuelve el usuario de ese email, o lo crea si es la primera vez.
func (h *Handlers) usuarioDeGoogle(r *http.Request, p PerfilGoogle) (middleware.Usuario, error) {
	ctx := r.Context()
	for range 2 {
		var u middleware.Usuario
		fila, existe, err := h.base.Leer(ctx, middleware.PKUsuario(p.Email), "PERFIL")
		if err == nil && existe {
			err = json.Unmarshal(fila.Datos, &u)
		}
		if err != nil {
			return u, err
		}
		if existe {
			// Google puede reasignar un email (cuentas de empresa): no es la misma persona.
			if u.ID != p.Sub {
				return u, errHTTP{http.StatusConflict, "Ese email ya está asociado a otra cuenta de Google."}
			}
			return u, nil
		}
		ahora := time.Now()
		nombre := strings.TrimSpace(p.Nombre)
		if nombre == "" {
			nombre, _, _ = strings.Cut(p.Email, "@")
		}
		u = middleware.Usuario{ID: p.Sub, Nombre: nombre, Email: p.Email, Img: p.Foto, Rol: "user", Creado: ahora, Actualizado: ahora}
		datos, _ := json.Marshal(u)
		err = h.base.Escribir(ctx, base.Cambio{Fila: base.Fila{PK: middleware.PKUsuario(p.Email), SK: "PERFIL", Datos: datos}})
		// ErrConflicto: entró dos veces a la vez y el otro pedido lo creó; se vuelve a leer
		if !errors.Is(err, base.ErrConflicto) {
			return u, err
		}
	}
	return middleware.Usuario{}, errors.New("no se pudo crear el usuario")
}

// GET /api/authentication: el usuario de la sesión. El front lo pide al abrir.
func (h *Handlers) Yo(w http.ResponseWriter, r *http.Request) {
	httputil.JSON(w, http.StatusOK, "", aJSON(middleware.UsuarioDe(r.Context())))
}

// POST /api/authentication/logout: cierra esta sesión (el token deja de servir).
func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	token, _ := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	err := h.base.Escribir(r.Context(), base.Cambio{Fila: base.Fila{PK: middleware.PKSesion(token), SK: "SESION"}, Borrar: true})
	if err != nil {
		httputil.Interno(w, err)
		return
	}
	httputil.JSON(w, http.StatusOK, "Saliste.", nil)
}
