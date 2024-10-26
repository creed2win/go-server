package main

import (
	"go-server/internal/database"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQ            *database.Queries
}
