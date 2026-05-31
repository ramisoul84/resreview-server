package server

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/ramisoul84/resreview-server/internal/config"
	"github.com/ramisoul84/resreview-server/internal/server/http/handler"
	"github.com/ramisoul84/resreview-server/internal/server/http/middleware"
	"github.com/ramisoul84/resreview-server/pkg/logger"
)

type Server struct {
	app *fiber.App
	cfg *config.Config
	log logger.Logger
}

func New(cfg *config.Config) *Server {
	app := fiber.New(fiber.Config{
		ReadTimeout:           cfg.Server.ReadTimeout,
		WriteTimeout:          cfg.Server.WriteTimeout,
		IdleTimeout:           cfg.Server.IdleTimeout,
		BodyLimit:             cfg.Server.BodyLimitMB * 1024 * 1024,
		Concurrency:           256 * 1024,
		DisableStartupMessage: true,
	})

	s := &Server{app: app, cfg: cfg, log: logger.Get()}
	s.registerMiddleware()
	return s
}

func (s *Server) App() *fiber.App { return s.app }

func (s *Server) RegisterRoutes(
	auth handler.AuthHandler,
	users handler.UserHandler,
	sessions handler.SessionHandler,
	products handler.ProductHandler,
	versions handler.VersionHandler,
	annotations handler.AnnotationHandler,
) {
	api := s.app.Group("/api/v1")
	api.Get("/health", s.healthHandler())

	authLimiter := limiter.New(limiter.Config{
		Max:          10,
		Expiration:   time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string { return c.IP() },
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "too many attempts, try again later"})
		},
	})

	authGroup := api.Group("/auth", authLimiter)
	authGroup.Post("/login", auth.Login)
	authGroup.Post("/register", auth.Register)
	authGroup.Post("/refresh", auth.Refresh)
	authGroup.Post("/logout", auth.Logout)

	// Protected routes (require JWT)
	protected := api.Group("", middleware.JWTAuth(s.cfg.JWT.Secret))
	protected.Get("/users", users.ListUsers)
	protected.Patch("/users/profile", users.UpdateProfile)
	protected.Patch("/users/:userId/role", users.UpdateRole)

	protected.Post("/sessions", sessions.CreateSession)
	protected.Get("/sessions", sessions.ListSessions)
	protected.Get("/sessions/:sessionId", sessions.GetSession)
	protected.Put("/sessions/:sessionId", sessions.UpdateSession)
	protected.Delete("/sessions/:sessionId", sessions.DeleteSession)

	protected.Post("/products", products.CreateProduct)
	protected.Get("/products", products.ListProducts)
	protected.Get("/products-with-versions", products.ListProductsWithVersions)
	protected.Get("/products/:productId", products.GetProduct)
	protected.Put("/products/:productId", products.UpdateProduct)
	protected.Delete("/products/:productId", products.DeleteProduct)

	protected.Post("/products/:productId/versions", versions.CreateVersion)
	protected.Get("/products/:productId/versions", versions.ListVersions)
	protected.Get("/products/:productId/versions/:versionId", versions.GetVersion)
	protected.Put("/products/:productId/versions/:versionId", versions.UpdateVersion)
	protected.Delete("/products/:productId/versions/:versionId", versions.DeleteVersion)
	protected.Post("/products/:productId/versions/:versionId/upload", versions.UploadImage)

	protected.Post("/products/:productId/versions/:versionId/annotations", annotations.CreateAnnotation)
	protected.Get("/products/:productId/versions/:versionId/annotations", annotations.ListAnnotations)
	protected.Put("/products/:productId/versions/:versionId/annotations/:annotationId", annotations.UpdateAnnotation)
	protected.Delete("/products/:productId/versions/:versionId/annotations/:annotationId", annotations.DeleteAnnotation)
}

func (s *Server) registerMiddleware() {
	s.app.Use(middleware.InjectRequestID())
	s.app.Use(cors.New(cors.Config{
		AllowOrigins:     s.cfg.Server.AllowedOrigins,
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Authorization,X-Request-ID",
		AllowCredentials: true,
		MaxAge:           86400,
	}))
	s.app.Use(compress.New(compress.Config{Level: compress.LevelBestSpeed}))
}

func (s *Server) healthHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": s.cfg.App.Name, "version": s.cfg.App.Version, "env": s.cfg.App.Env})
	}
}
