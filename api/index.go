// api/index.go
package handler

import (
	"context"

	auth "grammarhive-backend/api/routes/auth"
	grammar "grammarhive-backend/api/routes/grammar"
	middleware "grammarhive-backend/api/routes/middleware"
	"grammarhive-backend/core/config"
	"grammarhive-backend/core/database"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)
var (
	dbService     *database.MongoDB
	authenticator *middleware.Authenticator
)
func init() {
	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	dbService, err = database.NewMongoDB(ctx, cfg.MongoURI)
	if err != nil {
		panic(err)
	}

	authenticator, err = middleware.NewAuth0(cfg.Auth0Domain, cfg.Auth0Audience)
	if err != nil {
		panic(err)
	}

	grammar.InitGrammarService(dbService)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	router := mux.NewRouter()

	// All the routes are defined here!!
	router.HandleFunc("/api/login", auth.HandleLogin).Methods("POST")

	router.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Health good"))
    }).Methods("GET")

	// Secured routes
	router.HandleFunc("/api/grammar/generate",
		authenticator.Middleware(grammar.HandleGenerate),
	).Methods("GET")

	router.HandleFunc("/api/grammar/generateList",
		authenticator.Middleware(grammar.HandleGenerateList),
	).Methods("GET")

	// CORS Preflight
	if r.Method == "OPTIONS" {
		middleware.HandleOptions(w, r)
		return
	}

	router.ServeHTTP(w, r)
}
