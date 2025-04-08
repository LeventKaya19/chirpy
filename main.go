package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"

	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileServerHits atomic.Int32
}

func main() {
	httpServerMux := http.NewServeMux()
	httpServer := http.Server{
		Handler: httpServerMux,
		Addr:    ":8080",
	}
	apiCfg := &apiConfig{}
	httpServerMux.Handle("/app/", apiCfg.middleWareMetricsInc(fileServerHandler()))
	httpServerMux.HandleFunc("GET /api/healthz", healthHandler)
	httpServerMux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	httpServerMux.HandleFunc("POST /admin/reset", apiCfg.resetMetricsHandler)
	httpServerMux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)
	httpServer.ListenAndServe()
}

func fileServerHandler() http.Handler {
	return http.StripPrefix("/app/", http.FileServer(http.Dir(".")))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	header := w.Header()
	header.Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) middleWareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Increment the counter for each request
		cfg.fileServerHits.Add(1)

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	hits := cfg.fileServerHits.Load()
	body := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", hits)
	w.Write([]byte(body))
}

func (cfg *apiConfig) resetMetricsHandler(w http.ResponseWriter, r *http.Request) {
	cfg.fileServerHits.Store(0)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Counter reset to 0"))
}

func validateChirpHandler(w http.ResponseWriter, r *http.Request) {

	type post struct {
		Body string `json:body`
	}

	decoder := json.NewDecoder(r.Body)
	chirp := post{}
	err := decoder.Decode(&chirp)
	if err != nil {

		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}
	type returnVals struct {
		Error       string `json:"error,omitempty"`
		CleanedBody string `json:"cleaned_body,omitempty"`
	}
	w.Header().Set("Content-Type", "application/json")
	if len(chirp.Body) < 141 {

		w.WriteHeader(http.StatusOK)
		returnVal := returnVals{
			CleanedBody: profanityCheck(chirp.Body),
		}
		dat, _ := json.Marshal(returnVal)
		w.Write(dat)
		return
	}
	// returnVal := returnVals{
	// 	Error: "Chirp is too long",
	// }
	respondWithError(w, 400, "Chirp is too long")
	// dat, _ := json.Marshal(returnVal)

	// w.WriteHeader(400)
	// w.Write(dat)

}

func profanityCheck(text string) string {
	profane_words := []string{
		"kerfuffle",
		"sharbert",
		"fornax",
	}
	newSentence := text
	for _, profane_word := range profane_words {
		newSentence = profanityCheckHelper(newSentence, profane_word)
	}

	return newSentence
}

func profanityCheckHelper(text string, profanity string) string {
	words := strings.Split(text, " ")
	for i, word := range words {
		lowercaseWord := strings.ToLower(word)
		if lowercaseWord == profanity {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) error {
	response, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	w.Write(response)
	return nil
}

func respondWithError(w http.ResponseWriter, code int, msg string) error {
	return respondWithJSON(w, code, map[string]string{"error": msg})
}
