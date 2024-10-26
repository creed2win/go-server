package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"go-server/internal/database"
	"io"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Println("error opening sql database", err)
	}

	dbQueries := database.New(db)

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		dbQ:            dbQueries,
	}

	serverMux := http.ServeMux{}
	serverMux.HandleFunc("GET /api/healthz", respWriter)
	serverMux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	serverMux.HandleFunc("POST /admin/reset", apiCfg.handleReset)
	serverMux.HandleFunc("POST /api/validate_chirp", apiCfg.handleValidation)
	serverMux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	server := http.Server{}

	server.Handler = &serverMux
	server.Addr = ":8080"
	server.ListenAndServe()
}

func respWriter(writer http.ResponseWriter, req *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(200)
	writer.Write([]byte("OK"))
}

func (cfg *apiConfig) handlerMetrics(writer http.ResponseWriter, req *http.Request) {
	writer.Header().Add("Content-Type", "text/html; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>
`, cfg.fileserverHits.Load())))
	fmt.Println("hits logged: ", cfg.fileserverHits.Load())
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		cfg.fileserverHits.Add(1)
		fmt.Println("hits added: ", cfg.fileserverHits.Load())
		next.ServeHTTP(writer, req)
	})
}
func (cfg *apiConfig) handleReset(writer http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits.Store(0)
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("Hits reset to 0."))
}

func (cfg *apiConfig) handleValidation(writer http.ResponseWriter, req *http.Request) {
	data, err := io.ReadAll(req.Body)
	if err != nil {
		fmt.Println("error reading data", err)
	}
	type respBody struct {
		Valid       bool   `json:"valid"`
		Error       string `json:"error"`
		CleanedBody string `json:"cleaned_body"`
	}
	if len(string(data)) > 140 {
		body := respBody{
			Valid: false,
			Error: "",
		}
		dat, err := json.Marshal(body)
		if err != nil {
			fmt.Println("error marshaling data while invalid", err)
		}

		writer.WriteHeader(400)
		writer.Write(dat)
		return
	}
	type responseBody struct {
		Body string `json:"Body"`
	}

	bodyUnmarsh := responseBody{}

	err = json.Unmarshal(data, &bodyUnmarsh)
	if err != nil {
		fmt.Println("error unmarshaling data", err)
	}

	cleanBody := replaceBadWords(bodyUnmarsh.Body)

	body := respBody{
		CleanedBody: cleanBody,
		Valid:       true,
	}

	dat, err := json.Marshal(body)
	if err != nil {
		fmt.Println("error marshaling data while valid", err)
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(200)
	writer.Write(dat)

}
