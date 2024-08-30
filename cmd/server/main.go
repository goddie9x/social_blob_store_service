package main

import (
	api "blob_store_service/internal/apis"
	"blob_store_service/internal/blobstore"
	"blob_store_service/internal/config"
	eureka_route "blob_store_service/pkg/eureka"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	r.ForwardedByClientIP = true
	r.SetTrustedProxies([]string{"127.0.0.1"})
	handler := api.NewHandler(bs)

	api := r.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			blobs := v1.Group("/blobs")
			{
				blobs.POST("/upload", handler.UploadBlobs)
				blobs.GET("/download", handler.DownloadBlob)
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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	discoveryServerConnect := new(config.DiscoveryServerConnect)
	go func() {
		<-quit
		discoveryServerConnect.DeregisterFromEurekaDiscoveryServer()
		os.Exit(0)
	}()

	discoveryServerConnect.ConnectToEurekaDiscoveryServer(cfg)
	r.Run(fmt.Sprintf(":%d", cfg.Port))
}
