package router

import (
	controller "core/api/controller"
	service "core/internal/service"

	"net/http"
	_ "core/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Config(db *gorm.DB, caller *service.Caller) http.Handler {
	router := gin.Default()
	controllers := controller.Create(db, caller)

	router.Use(func(c *gin.Context) {
		switch c.Request.URL.Path {
		case "/swagger", "/swagger/":
			c.Redirect(http.StatusTemporaryRedirect, "/swagger/index.html")
			c.Abort()
			return
		}
		c.Next()
	})

	router.Use(cors.New(cors.Config{
		AllowCredentials:       true,
		AllowBrowserExtensions: true,
		AllowHeaders:           []string{"*"},
		AllowOrigins:           []string{"http://localhost:3000"},
		AllowMethods:           []string{"GET", "POST"},
	}))

	router.GET(
		"/swagger/*any",
		ginSwagger.WrapHandler(
			swaggerFiles.Handler,
			ginSwagger.URL("/swagger/doc.json"),
		),
	)

	v1 := router.Group("/api/v1")

	v1.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	autocall := v1.Group("/autocall")
	{
		autocall.POST("/init", controllers.Init)
		autocall.GET("/status", controllers.GetStatus)
		autocall.GET("/leads", controllers.GetLeads)
	}

	return router
}
