package router

import (
	controller "autocall/api/controller"
	service "autocall/internal/service"

	"net/http"

	_ "autocall/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

func Config(db *gorm.DB, caller *service.Caller) http.Handler {
	router := gin.Default()
	controllers := controller.Create(db, caller)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.Use(cors.New(cors.Config{
		AllowCredentials:       true,
		AllowBrowserExtensions: true,
		AllowHeaders:           []string{"*"},
		AllowOrigins:           []string{"http://localhost:3000"},
		AllowMethods:           []string{"GET", "POST"},
	}))

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
