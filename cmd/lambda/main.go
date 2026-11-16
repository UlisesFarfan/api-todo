// Command lambda corre la API en AWS Lambda con el runtime "provided.al2023".
// Habla directo con la Runtime API de Lambda, sin el SDK (como gia-pedidos). Se
// compila con make build (build/bootstrap). La configuración llega como
// variables de entorno (ver template.yaml).
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"api-todo/internal/lambdahttp"
	"api-todo/internal/router"
)

func main() {
	ctx := context.Background()
	b, err := router.BaseDeEntorno(ctx)
	if err != nil {
		log.Fatal(err)
	}
	google, err := router.GoogleDeEntorno()
	if err != nil {
		log.Fatal(err)
	}
	srv := router.New(b, router.OrigenesDeEntorno(), google)

	base := "http://" + os.Getenv("AWS_LAMBDA_RUNTIME_API") + "/2018-06-01/runtime"
	client := &http.Client{}
	for {
		resp, err := client.Get(base + "/invocation/next")
		if err != nil {
			log.Fatalf("runtime next: %v", err)
		}
		requestID := resp.Header.Get("Lambda-Runtime-Aws-Request-Id")
		raw, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			postError(client, base, requestID, err)
			continue
		}
		var ev lambdahttp.Event
		if err := json.Unmarshal(raw, &ev); err != nil {
			postError(client, base, requestID, err)
			continue
		}
		out, err := lambdahttp.Handle(srv, ev)
		if err != nil {
			postError(client, base, requestID, err)
			continue
		}
		payload, _ := json.Marshal(out)
		if _, err := client.Post(fmt.Sprintf("%s/invocation/%s/response", base, requestID), "application/json", bytes.NewReader(payload)); err != nil {
			log.Printf("post response: %v", err)
		}
	}
}

func postError(client *http.Client, base, requestID string, cause error) {
	log.Printf("invocation %s failed: %v", requestID, cause)
	body, _ := json.Marshal(map[string]string{"errorMessage": cause.Error(), "errorType": "AdapterError"})
	_, _ = client.Post(fmt.Sprintf("%s/invocation/%s/error", base, requestID), "application/json", bytes.NewReader(body))
}
