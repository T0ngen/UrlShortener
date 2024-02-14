package main

import (
	"log/slog"
	"net/http"
	"os"
	"url-shortener/internal/config"
	"url-shortener/internal/http-server/middleware/logger/handlers/redirect"
	"url-shortener/internal/http-server/middleware/logger/handlers/url/save"

	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage/sqlite"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)



const (
	envLocal = "local"
	envDev = "dev"
	envProd = "prod"
)


func main(){
	cfg := config.MustLoad()
	log := setupLogger((cfg.Env))
	log.Info("starting url-shortener", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil{
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}
	router := chi.NewRouter()
	//middlewareч
	_ = storage
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shortener", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
			// cfg.HTTPServer.User: cfg.HTTPServer.Password,
			// cfg.HTTPServer.User: cfg.HTTPServer.Password,
			// cfg.HTTPServer.User: cfg.HTTPServer.Password,
			//Добавление друзей
		}))

		r.Post("/", save.New(log, storage))
		// TODO: add DELETE /url/{id}
	})
	router.Get("/{alias}", redirect.New(log, storage))
	//TODO: DELETE METHOD
	// router.Delete("/url/{alias}", delete.New(log, storage))
	// router.Use(middleware.RealIP)

	log.Info("starting server", slog.String("address", cfg.Address))

	srv := &http.Server{
		Addr: cfg.Address,
		Handler: router,
		ReadTimeout: cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout: cfg.HTTPServer.Timeout,
	}
	if err:= srv.ListenAndServe(); err != nil{
		log.Error("failed to start server", sl.Err(err))
		os.Exit(1)
	}
}



func setupLogger(env string) *slog.Logger{
	var log *slog.Logger
	switch env{
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug} ),)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
		
	}
	return log
}