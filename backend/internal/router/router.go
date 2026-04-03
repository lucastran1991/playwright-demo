package router

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/user/app/internal/handler"
	"github.com/user/app/internal/middleware"
)

// Setup creates and configures the Gin engine with routes.
func Setup(authHandler *handler.AuthHandler, blueprintHandler *handler.BlueprintHandler, tracerHandler *handler.TracerHandler, jwtSecret string) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Public auth routes
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
	}

	// Public blueprint read endpoints
	blueprints := r.Group("/api/blueprints")
	{
		blueprints.GET("/types", blueprintHandler.ListTypes)
		blueprints.GET("/nodes", blueprintHandler.ListNodes)
		blueprints.GET("/nodes/:nodeId", blueprintHandler.GetNode)
		blueprints.GET("/edges", blueprintHandler.ListEdges)
		blueprints.GET("/tree/:typeSlug", blueprintHandler.GetTree)
	}

	// Public model + trace endpoints
	models := r.Group("/api/models")
	{
		models.GET("/capacity-nodes", tracerHandler.ListCapacityNodes)
	}
	trace := r.Group("/api/trace")
	{
		trace.GET("/dependencies/:nodeId", tracerHandler.TraceDependencies)
		trace.GET("/impacts/:nodeId", tracerHandler.TraceImpacts)
	}

	// Protected routes
	protected := r.Group("/api")
	protected.Use(middleware.AuthRequired(jwtSecret))
	{
		protected.GET("/auth/me", authHandler.Me)
		protected.POST("/blueprints/ingest", blueprintHandler.Ingest)
		protected.POST("/models/ingest", tracerHandler.IngestModels)
	}

	return r
}
