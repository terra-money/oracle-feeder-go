package main

import (
	"context"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/terra-money/oracle-feeder-go/config"
	"github.com/terra-money/oracle-feeder-go/internal/provider"
)

func main() {
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
		allianceProtocol, err := allianceProvider.GetProtocolsInfo(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, allianceProtocol)
	})
	if os.Getenv("PORT") == "" {
		os.Setenv("PORT", "8532") // use 8532 by default
	}

	r.Run()
	close(stopCh)
}
