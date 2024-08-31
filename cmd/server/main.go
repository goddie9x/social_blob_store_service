package main

import (
	api "blob_store_service/internal/apis"
	"blob_store_service/internal/blobstore"
	"blob_store_service/internal/config"
	"blob_store_service/internal/routes"
	"fmt"
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
	r.ForwardedByClientIP = true
	r.SetTrustedProxies([]string{"127.0.0.1"})
	handler := api.NewHandler(bs)
	routes.StartMappingBlobRoute(r, handler)

	discoveryServerConnect := new(config.DiscoveryServerConnect)

	discoveryServerConnect.ConnectToEurekaDiscoveryServer(cfg)
	r.Run(fmt.Sprintf(":%d", cfg.Port))
}
