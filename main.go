package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
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

func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	result := struct {
		Error string `json:"error,omitempty"`
	}{
		Error: msg,
	}

	dat, err := json.Marshal(result)
	if err != nil {
		log.Println("error Marshalling error message")
		return
	}

	w.Write(dat)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	dat, err := json.Marshal(payload)
	if err != nil {
		log.Println("error Marshalling error message")
		return
	}

	w.Write(dat)
}

func badWordReplacer(s string) string {
	result := []string{}
	wordList := strings.Split(s, " ")
	badWords := []string{"kerfuffle", "sharbert", "fornax"}

	for _, word := range wordList {
		if slices.Contains(badWords, strings.ToLower(word)) {
			result = append(result, "****")
		} else {
			result = append(result, word)
		}
	}

	return strings.Join(result, " ")
}

func handleValidateChirp(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type valid struct {
		//	the key will be the name of struct field unless you give it an explicit JSON tag
		CleanedBody string `json:"cleaned_body,omitempty"`
	}

	w.Header().Set("Content-Type", "application/json")

	result := valid{}

	decoder := json.NewDecoder(req.Body)
	// decoder.DisallowUnknownFields()

	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, "error decoding request json")

		// w.WriteHeader(400)
		// result := result{
		// 	Error: 		}
		// dat, err := json.Marshal(result)
		// if err != nil {
		// 	log.Println("error Marshalling error message")
		// 	return
		// }
		//
		// w.Write(dat)
		return
	}

	if len(params.Body) <= 140 {
		result.CleanedBody = badWordReplacer(params.Body)
		respondWithJSON(w, 200, result)
	} else {
		respondWithError(w, 400, "Chirp is too long")
		return
	}
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
