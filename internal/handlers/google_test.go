package handlers

import (
	"context"
	"testing"
)

func TestVerificadorGoogleRechaza(t *testing.T) {
	ctx := context.Background()
	// sin client ID, idtoken no miraría la audiencia: se rechaza antes
	if _, err := VerificadorGoogle("")(ctx, "x.y.z"); err == nil {
		t.Fatal("sin client ID aceptó el token")
	}
	if _, err := VerificadorGoogle("cliente")(ctx, "no-es-un-jwt"); err == nil {
		t.Fatal("aceptó un token roto")
	}
}
