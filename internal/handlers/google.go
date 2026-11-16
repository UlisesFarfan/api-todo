package handlers

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/api/idtoken"
)

// PerfilGoogle: lo que sale del ID token de Google ya verificado.
type PerfilGoogle struct {
	Sub, Email, Nombre, Foto string
}

// VerificarGoogle valida el ID token ("credential") que da el botón de Google.
// En los tests se reemplaza por uno falso.
type VerificarGoogle func(ctx context.Context, credential string) (PerfilGoogle, error)

// VerificadorGoogle revisa firma, audiencia (el client ID del front),
// vencimiento, emisor y que el email esté verificado.
func VerificadorGoogle(clientID string) VerificarGoogle {
	return func(ctx context.Context, credential string) (PerfilGoogle, error) {
		// idtoken no mira la audiencia si viene vacía
		if clientID == "" {
			return PerfilGoogle{}, errors.New("falta GOOGLE_CLIENT_ID")
		}
		p, err := idtoken.Validate(ctx, credential, clientID)
		if err != nil {
			return PerfilGoogle{}, err
		}
		// idtoken no revisa el emisor
		if p.Issuer != "accounts.google.com" && p.Issuer != "https://accounts.google.com" {
			return PerfilGoogle{}, errors.New("emisor del token desconocido: " + p.Issuer)
		}
		if verificado, _ := p.Claims["email_verified"].(bool); !verificado {
			return PerfilGoogle{}, errors.New("email de Google sin verificar")
		}
		email, _ := p.Claims["email"].(string)
		nombre, _ := p.Claims["name"].(string)
		foto, _ := p.Claims["picture"].(string)
		if email == "" || p.Subject == "" {
			return PerfilGoogle{}, errors.New("token sin email o sin sub")
		}
		return PerfilGoogle{Sub: p.Subject, Email: strings.ToLower(email), Nombre: nombre, Foto: foto}, nil
	}
}
