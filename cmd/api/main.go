package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/teodosiopiera/api-source-proxy/internal/config"
	"github.com/teodosiopiera/api-source-proxy/internal/handler"
	"github.com/teodosiopiera/api-source-proxy/internal/repository/postgres"
	mongorepo "github.com/teodosiopiera/api-source-proxy/internal/repository/mongo"
	"github.com/teodosiopiera/api-source-proxy/internal/service"
	"github.com/teodosiopiera/api-source-proxy/pkg/database"
)

func main() {
	cfg := config.Load()

	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	pg, err := database.NewPostgres(
		cfg.Database.Postgres.Host,
		cfg.Database.Postgres.Port,
		cfg.Database.Postgres.User,
		cfg.Database.Postgres.Password,
		cfg.Database.Postgres.DBName,
		cfg.Database.Postgres.SSLMode,
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to PostgreSQL")
	}
	defer pg.Close()

	var mgo *mongo.Database
	mgo, err = database.NewMongo(cfg.Database.Mongo.URI, cfg.Database.Mongo.DBName)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}

	userRepo := postgres.NewUserRepo(pg)
	apiKeyRepo := postgres.NewApiKeyRepo(pg)
	apiSourceRepo := postgres.NewApiSourceRepo(pg)
	logRepo := mongorepo.NewLogRepo(mgo)

	authService := service.NewAuthService(
		userRepo, apiKeyRepo,
		cfg.JWT.Secret, cfg.JWT.ExpiryHour,
		cfg.Admin.Username, cfg.Admin.Password, cfg.Admin.Email,
	)

	if err := authService.InitAdmin(context.Background()); err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize admin user")
	}

	userService := service.NewUserService(userRepo, apiKeyRepo)

	proxyService := service.NewProxyService(apiSourceRepo, logRepo, cfg.Server.Proxy)

	if err := proxyService.InitDefaultSources(context.Background()); err != nil {
		logger.Warn().Err(err).Msg("Failed to initialize default sources")
	}

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	apiKeyHandler := handler.NewApiKeyHandler(userService)
	proxyHandler := handler.NewProxyHandler(proxyService, userService)
	logHandler := handler.NewLogHandler(proxyService)
	dashboardHandler := handler.NewDashboardHandler()

	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(60 * time.Second))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-API-Key", "X-Client"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/login", authHandler.Login)

		r.Group(func(r chi.Router) {
			r.Use(handler.AdminAuth(cfg.JWT.Secret))
			r.Post("/users", userHandler.Create)
			r.Get("/users", userHandler.List)
			r.Put("/users/{id}", userHandler.Update)

			r.Post("/api-keys", apiKeyHandler.Create)
			r.Get("/api-keys", apiKeyHandler.List)
			r.Delete("/api-keys/{id}", apiKeyHandler.Revoke)

			r.Get("/api-sources", proxyHandler.ListSources)
			r.Post("/api-sources", proxyHandler.CreateSource)
			r.Put("/api-sources/{id}", proxyHandler.UpdateSource)
			r.Delete("/api-sources/{id}", proxyHandler.DeleteSource)

			r.Get("/activity-logs", logHandler.List)
		})
	})

	r.Route("/api/v1/user", func(r chi.Router) {
		r.Use(handler.UserAuth(cfg.JWT.Secret))
		r.Get("/activity-logs", logHandler.UserList)
		r.Get("/api-sources", proxyHandler.ListSources)
		r.Post("/proxy-test/{source}", proxyHandler.ProxyTest)
	})

	r.Route("/api/v1/proxy", func(r chi.Router) {
		r.Use(handler.ApiKeyAuth(authService))
		r.Post("/{source}/getLocation", proxyHandler.Proxy)
	})

	r.Get("/dashboard", dashboardHandler.Serve)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info().Int("port", cfg.Server.Port).Msg("Server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Server failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	logger.Info().Msg("Server stopped gracefully")
}
