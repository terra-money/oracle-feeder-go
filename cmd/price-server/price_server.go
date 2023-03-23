package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/terra-money/oracle-feeder-go/configs"
	"github.com/terra-money/oracle-feeder-go/internal/provider"
)

func main() {
	stopCh := make(chan struct{})
	manager := provider.NewProviderManager(&configs.DefaultConfig, stopCh)

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
	r.GET("/latest", func(c *gin.Context) {
		prices := manager.GetPrices()
		c.JSON(http.StatusOK, prices)
	})
	if os.Getenv("PORT") == "" {
		os.Setenv("PORT", "8532") // use 8532 by default
	}

	r.Run()
	close(stopCh)
}
