// api/index.go
package handler

import (
	"context"
	"fmt"

	auth "grammarhive-backend/api/routes/auth"
	handler "grammarhive-backend/api/routes/handler"
	middleware "grammarhive-backend/api/routes/middleware"
	"grammarhive-backend/core/config"
	"grammarhive-backend/core/database"

	"net/http"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/gorilla/mux"
)

type App struct {
	dbService     *database.MongoDB
	authenticator *middleware.Authenticator
	grammar       *handler.GrammarHandler
	profile      *handler.ProfileHandler
}

var app = NewApp()

func NewApp() *App {
	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbService, err := database.NewMongoDB(ctx, cfg.MongoURI)
	if err != nil {
		panic(err)
	}

	authenticator, err := middleware.NewAuth0(cfg.Auth0Domain, cfg.Auth0Audience)
	if err != nil {
		panic(err)
	}

	grammar := handler.NewGrammarHandler(dbService)
	profile := handler.NewProfileHandler(dbService)

	return &App{
		dbService:     dbService,
		authenticator: authenticator,
		grammar:       grammar,
		profile:       profile,
	}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	log.Info(fmt.Sprintf(
		"%s: method=%s, uri=%s", r.Proto, r.Method, r.RequestURI),
	)

	router := mux.NewRouter()

	// All the routes are defined here!!
	router.HandleFunc("/api/login", auth.HandleLogin).Methods("POST")

	router.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Health good"))
    }).Methods("GET")

	// Secured routes
	router.HandleFunc("/api/grammar/generate",
		app.authenticator.Middleware(app.grammar.HandleGenerate),
	).Methods("GET")

	router.HandleFunc("/api/grammar/generateList",
		app.authenticator.Middleware(app.grammar.HandleGenerateList),
	).Methods("GET")

	router.HandleFunc("/api/user/profile/grammar/upload",
		app.authenticator.Middleware(app.profile.HandleUpload),
	).Methods("POST")

	router.HandleFunc("/api/user/profile/grammar",
		app.authenticator.Middleware(app.profile.HandleGetGrammarByUsername),
	).Methods("GET")

	// CORS Preflight
	if r.Method == "OPTIONS" {
		middleware.HandleOptions(w, r)
		return
	}

	router.ServeHTTP(w, r)
}
