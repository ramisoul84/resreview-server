package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ramisoul84/resreview-server/internal/config"
	"github.com/ramisoul84/resreview-server/internal/repository"
	httpServer "github.com/ramisoul84/resreview-server/internal/server/http"
	"github.com/ramisoul84/resreview-server/internal/server/http/handler"
	"github.com/ramisoul84/resreview-server/internal/service"
	"github.com/ramisoul84/resreview-server/internal/ws"
	"github.com/ramisoul84/resreview-server/pkg/database"
	"github.com/ramisoul84/resreview-server/pkg/logger"
	"github.com/ramisoul84/resreview-server/pkg/storage"
)

func main() {
	// ── Config ────────────────────────────────────────────────────────────────
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}
	cfg := config.Load(env)

	// ── Logger ────────────────────────────────────────────────────────────────
	logger.InitGlobal(cfg)
	log := logger.Get()

	log.Info().
		Str("env", cfg.App.Env).
		Str("version", cfg.App.Version).
		Int("pid", os.Getpid()).
		Msg("Starting ResReview Server")

	// ── PostgreSQL ────────────────────────────────────────────────────
	db, err := database.ConnectPostgres(cfg.Postgres)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to PostgreSQL")
	}
	defer db.Close()
	log.Info().Str("host", cfg.Postgres.Host).Msg("PostgreSQL connected")

	// ── Redis ─────────────────────────────────────────────────────────
	redisClient, err := database.ConnectRedis(cfg.Redis)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redisClient.Close()
	log.Info().Str("addr", cfg.Redis.Addr).Msg("Redis connected")

	// ── MinIO ──────────────────────────────────────────────────────────────────
	photoStorage, err := storage.NewMinIOStorage(cfg.MinIO)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init MinIO storage")
	}
	log.Info().Str("endpoint", cfg.MinIO.Endpoint).Str("bucket", cfg.MinIO.Bucket).Msg("MinIO ready")

	// ── WebSocket Hub ─────────────────────────────────────────────────────────
	wsHub := ws.NewHub()
	go wsHub.Run()

	http.HandleFunc("/ws", ws.HandleWebSocket(wsHub, cfg.JWT.Secret))

	wsServer := &http.Server{
		Addr:    ":" + cfg.Server.WSPort,
		Handler: nil, // uses default mux
	}

	go func() {
		log.Info().Str("port", cfg.Server.WSPort).Msg("WebSocket server listening")
		if err := wsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("WebSocket server listen error")
		}
	}()

	// ── Repositories ──────────────────────────────────────────────────────────
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(redisClient)
	sessionRepo := repository.NewSessionRepository(db)
	productRepo := repository.NewProductRepository(db)
	versionRepo := repository.NewVersionRepository(db)
	photoRepo := repository.NewPhotoRepository(db)
	annotationRepo := repository.NewAnnotationRepository(db)

	// ── Services ──────────────────────────────────────────────────────────────
	tokenSvc := service.NewTokenService(tokenRepo, cfg)
	authSvc := service.NewAuthService(userRepo, tokenSvc)
	userSvc := service.NewUserService(userRepo)
	sessionSvc := service.NewSessionService(sessionRepo)
	productSvc := service.NewProductService(productRepo, versionRepo)
	versionSvc := service.NewVersionService(versionRepo, productRepo, photoRepo, photoStorage, cfg.MinIO.MaxSizeBytes())
	annotationSvc := service.NewAnnotationService(annotationRepo, versionRepo, productRepo, wsHub)

	// ── Handlers ──────────────────────────────────────────────────────────────
	authHandler := handler.NewAuthHandler(authSvc, tokenSvc, cfg)
	userHandler := handler.NewUserHandler(userSvc)
	sessionHandler := handler.NewSessionHandler(sessionSvc)
	productHandler := handler.NewProductHandler(productSvc)
	versionHandler := handler.NewVersionHandler(versionSvc)
	annotationHandler := handler.NewAnnotationHandler(annotationSvc)

	// ── HTTP Server ───────────────────────────────────────────────────────────
	srv := httpServer.New(cfg)
	srv.RegisterRoutes(
		authHandler,
		userHandler,
		sessionHandler,
		productHandler,
		versionHandler,
		annotationHandler,
	)

	go func() {
		log.Info().Str("port", cfg.Server.Port).Msg("HTTP server listening")
		if err := srv.App().Listen(":" + cfg.Server.Port); err != nil {
			log.Fatal().Err(err).Msg("Server listen error")
		}
	}()

	// ── Graceful Shutdown ─────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	log.Info().Str("signal", sig.String()).Msg("Shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	wsHub.Shutdown()
	time.Sleep(100 * time.Millisecond)

	if err := wsServer.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("WebSocket server forced shutdown")
	}

	if err := srv.App().ShutdownWithContext(ctx); err != nil {
		log.Error().Err(err).Msg("Forced shutdown after timeout")
	}

	log.Info().Msg("ResReview Server stopped gracefully")
}
