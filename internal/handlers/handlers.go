// Package handlers: los endpoints de la API. Rutas y JSON son los que ya usa
// el front (app-todo); los datos van a la base (DynamoDB) como filas con versión.
package handlers

import (
	"errors"
	"net/http"

	"api-todo/internal/base"
	"api-todo/internal/httputil"
)

type Handlers struct {
	base      base.Base
	verificar VerificarGoogle
}

func New(b base.Base, verificar VerificarGoogle) *Handlers {
	return &Handlers{base: b, verificar: verificar}
}

// errHTTP es un error para mostrarle al usuario, con su status.
type errHTTP struct {
	status  int
	mensaje string
}

func (e errHTTP) Error() string { return e.mensaje }

func invalido(mensaje string) error { return errHTTP{http.StatusBadRequest, mensaje} }

func responderError(w http.ResponseWriter, err error) {
	var e errHTTP
	if errors.As(err, &e) {
		httputil.Error(w, e.status, e.mensaje)
		return
	}
	httputil.Interno(w, err)
}
