// El servidor local de la API (make api).
package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"api-todo/internal/router"
)

func main() {
	b, err := router.BaseDeEntorno(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	google, err := router.GoogleDeEntorno()
	if err != nil {
		log.Fatal(err)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("API en http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, router.New(b, router.OrigenesDeEntorno(), google)))
}
