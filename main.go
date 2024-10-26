package main

import (
	"database/sql"
	"fmt"
	"go-server/internal/database"
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
	serverMux.HandleFunc("POST /api/users", apiCfg.handleUsers)
	serverMux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	server := http.Server{}

	server.Handler = &serverMux
	server.Addr = ":8080"
	server.ListenAndServe()
}
