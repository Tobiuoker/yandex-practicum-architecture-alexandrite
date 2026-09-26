package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	ctx := context.Background()

	tp, err := initTracer(ctx, "service-a")
	if err != nil {
		log.Fatal(err)
	}
	defer tp.Shutdown(ctx)

	serviceBURL := os.Getenv("SERVICE_B_URL")
	if serviceBURL == "" {
		serviceBURL = "http://service-b:8080/"
	}

	client := &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, serviceBURL, nil)
		if err != nil {
			http.Error(w, "cannot create request", http.StatusInternalServerError)
			return
		}

		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "service-b is unavailable", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "cannot read service-b response", http.StatusBadGateway)
			return
		}

		if resp.StatusCode >= 400 {
			http.Error(w, "service-b returned an error", http.StatusBadGateway)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"service":"service-a","serviceB":%s}`, strings.TrimSpace(string(body)))
	})

	handler := otelhttp.NewHandler(mux, "service-a")

	log.Println("service-a listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
