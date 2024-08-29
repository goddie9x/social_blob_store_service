package main

import (
	api "blob_store_service/internal/apis"
	"blob_store_service/internal/blobstore"
	"blob_store_service/internal/config"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	bs, err := blobstore.NewBlobStore(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to create blob store: %v", err)
	}
	r := gin.Default()
	handler := api.NewHandler(bs)

	api := r.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			blobs := v1.Group("/blobs")
			{
				blobs.POST("/upload", handler.UploadBlob)
				blobs.GET("/download", handler.DownloadBlob)
				blobs.DELETE("/delete/:id", handler.DeleteBlob)
			}
		}
	}
	r.Run(cfg.Port)
	log.Printf("Starting server on %s", cfg.Port)
}
