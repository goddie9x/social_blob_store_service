package routes

import (
	api "blob_store_service/internal/apis"
	eureka_route "github.com/goddie9x/social_utils_go/eureka"
	"github.com/goddie9x/social_utils_go/middlewares"

	"time"

	"github.com/gin-gonic/gin"
)

func StartMappingBlobRoute(r *gin.Engine, handler *api.Handler) {
	api := r.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			blobs := v1.Group("/blobs")
			{
				// Auth middleware must be registered BEFORE the routes -
				// Gin only applies Use() to routes registered after it.
				blobs.Use(middlewares.PutAuthToContext)
				blobs.GET("/download/:id", handler.DownloadBlob)
				blobs.POST("/upload", handler.UploadBlobs)
				blobs.DELETE("/delete/:id", handler.DeleteBlob)
			}
		}
	}
	startTime := time.Now()
	global := r.Group("")
	{
		global.GET("/health", eureka_route.Health)
		global.GET("/status", eureka_route.Status(startTime))
	}
}
