package main

import (
	api "blob_store_service/internal/apis"
	"blob_store_service/internal/blobstore"
	"blob_store_service/internal/routes"
	"blob_store_service/pkg/dotenv"
	"log"
	"post_service/pkg/configs"

	"github.com/gin-gonic/gin"
)

func main() {
	bs, err := blobstore.NewBlobStore(dotenv.GetEnvOrDefaultValue("CONNECTION_STRING", "postgres"))
	if err != nil {
		log.Fatalf("Failed to create blob store: %v", err)
	}
	r := gin.Default()
	r.ForwardedByClientIP = true
	r.SetTrustedProxies([]string{"127.0.0.1"})
	handler := api.NewHandler(bs)
	routes.StartMappingBlobRoute(r, handler)

	discoveryServerConnect := new(configs.DiscoveryServerConnect)

	discoveryServerConnect.ConnectToEurekaDiscoveryServer()
	r.Run(":" + dotenv.GetEnvOrDefaultValue("API_PORT", "6543"))
}
