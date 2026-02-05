package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/shipit-ai-demo-org/shipping-rates-service/internal/carriers"
	"github.com/shipit-ai-demo-org/shipping-rates-service/internal/quote"
)

func main() {
	eng := quote.NewEngine(
		carriers.NewUPS(),
		carriers.NewFedEx(),
	)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","service":"shipping-rates-service"}`))
	})

	mux.HandleFunc("POST /v1/quotes", func(w http.ResponseWriter, r *http.Request) {
		var shp quote.Shipment
		if err := json.NewDecoder(r.Body).Decode(&shp); err != nil {
			http.Error(w, `{"error":"invalid_body"}`, http.StatusBadRequest)
			return
		}
		best, alts, err := eng.BestRate(r.Context(), shp)
		if err != nil {
			if errors.Is(err, quote.ErrNoQuotes) {
				http.Error(w, `{"error":"no_quotes"}`, http.StatusBadGateway)
				return
			}
			http.Error(w, `{"error":"internal"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"best": best, "alternatives": alts})
	})

	addr := ":" + envOr("PORT", "8081")
	log.Printf("shipping-rates-service listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
