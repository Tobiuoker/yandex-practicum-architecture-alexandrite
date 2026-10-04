package main

import (
	"context"
	"log"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	ctx := context.Background()

	tp, err := initTracer(ctx, "service-b")
	if err != nil {
		log.Fatal(err)
	}
	defer tp.Shutdown(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"service":"service-b"}`))
	})

	handler := otelhttp.NewHandler(mux, "service-b")

	log.Println("service-b listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
