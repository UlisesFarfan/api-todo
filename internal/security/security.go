// Package security: tokens de sesión.
package security

import (
	"crypto/rand"
	"crypto/sha256"
)

// NuevoToken devuelve el token para el front y el hash que se guarda en la base.
func NuevoToken() (token string, hash []byte) {
	token = rand.Text()
	return token, HashToken(token)
}

func HashToken(token string) []byte {
	h := sha256.Sum256([]byte(token))
	return h[:]
}
