package server

import (
	"fmt"
	"net/http"

	"github.com/joshuamURD/wclist/config"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Server struct {
	Config *config.Config
	Server *echo.Echo
}

func NewServer(config *config.Config) *Server {
	return &Server{Config: config}
}

func (s *Server) SetupRoutes() {
	// Add middleware
	s.Server.Use(middleware.Logger())
	s.Server.Use(middleware.Recover())
	s.Server.Use(middleware.CORS())

	// Register routes with handlers
	s.Server.GET("/", s.handleHome)
	s.Server.GET("/health", s.handleHealth)

	// API routes group
	api := s.Server.Group("/api/v1")
	api.GET("/status", s.handleAPIStatus)
}

// Handler for home route
func (s *Server) handleHome(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "WC List Server",
		"version": "1.0.0",
		"status":  "running",
	})
}

// Handler for health check
func (s *Server) handleHealth(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"timestamp": fmt.Sprintf("%d", s.Config.Port),
	})
}

// Handler for API status
func (s *Server) handleAPIStatus(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"api":     "WC List API",
		"version": "v1",
		"status":  "operational",
	})
}

func (s *Server) Start() error {
	s.Server = echo.New()

	// Hide Echo banner for cleaner startup
	s.Server.HideBanner = true

	// Setup routes
	s.SetupRoutes()

	address := fmt.Sprintf("%s:%s", s.Config.Localhost, s.Config.Port)
	fmt.Printf("🚀 Server starting on %s\n", address)

	return s.Server.Start(address)
}
