package routes

import (
	api "blob_store_service/internal/apis"
	eureka_route "blob_store_service/pkg/eureka"

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
				blobs.POST("/upload", handler.UploadBlobs)
				blobs.GET("/download/:id", handler.DownloadBlob)
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
