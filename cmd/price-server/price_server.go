package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/terra-money/oracle-feeder-go/config"
	"github.com/terra-money/oracle-feeder-go/internal/provider"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}
	ctx := context.Background()

	stopCh := make(chan struct{})
	manager := provider.NewProviderManager(&config.DefaultPriceServerConfig, stopCh)
	allianceProvider := provider.NewAllianceProvider(&config.DefaultAllianceConfig, manager)

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
	r.GET("/latest", func(c *gin.Context) {
		prices := manager.GetPrices(ctx)
		c.JSON(http.StatusOK, prices)
	})
	r.GET("/alliance/protocol", func(c *gin.Context) {
		allianceProtocolRes, err := allianceProvider.GetProtocolsInfo(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, allianceProtocolRes)
	})
	if os.Getenv("PRICE_SERVER_PORT") == "" {
		os.Setenv("PORT", "8532") // use 8532 by default
	} else {
		os.Setenv("PORT", os.Getenv("PRICE_SERVER_PORT"))
	}

	err = r.Run()
	if err != nil {
		panic(err)
	}

	close(stopCh)
}
