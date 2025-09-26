package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	cfg.fileserverHits.Add(1)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handleMetrics(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `
		<html>
			<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %d times!</p>
			</body>
		</html>`,
		cfg.fileserverHits.Load())
}

func (cfg *apiConfig) handleReset(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
}

func handleValidateChirp(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type result struct {
		//	the key will be the name of struct field unless you give it an explicit JSON tag
		Valid bool   `json:"valid,omitempty"`
		Error string `json:"error,omitempty"`
	}

	w.Header().Set("Content-Type", "application/json")

	res := result{}

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()

	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		w.WriteHeader(400)
		res := result{
			Error: "error decoding request json",
		}
		dat, err := json.Marshal(res)
		if err != nil {
			log.Println("error Marshalling error message")
			return
		}

		w.Write(dat)
		return
	}

	var statusCode int

	if len(params.Body) <= 140 {
		statusCode = 200
		res.Valid = true
	} else {
		statusCode = 400
		res.Error = "Chirp is too long"
	}

	dat, err := json.Marshal(res)
	if err != nil {
		log.Println("Error marshalling result")
		return
	}

	w.WriteHeader(statusCode)
	w.Write(dat)
}

func main() {
	mux := http.NewServeMux()
	apiCfg := &apiConfig{
		fileserverHits: atomic.Int32{},
	}

	handler := (http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(handler))
	mux.Handle("/app", apiCfg.middlewareMetricsInc(handler))

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("GET /admin/metrics", apiCfg.handleMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handleReset)
	mux.HandleFunc("POST /api/validate_chirp", handleValidateChirp)

	server := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	server.ListenAndServe()
}
