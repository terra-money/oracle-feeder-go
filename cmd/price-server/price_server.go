package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/terra-money/oracle-feeder-go/config"
	"github.com/terra-money/oracle-feeder-go/internal/provider"
	alliance_provider "github.com/terra-money/oracle-feeder-go/internal/provider/alliance"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Print("Error loading .env file:", err)
	}
	ctx := context.Background()

	stopCh := make(chan struct{})
	manager := provider.NewProviderManager(&config.DefaultPriceServerConfig, stopCh)
	allianceProvider := alliance_provider.NewAllianceProvider(&config.AllianceDefaultConfig, manager)

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"https://*.terra.money"},
		AllowMethods:  []string{"GET", "OPTIONS"},
		AllowHeaders:  []string{"Origin"},
		ExposeHeaders: []string{"Content-Length"},
	}))

	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
	r.GET("/latest", func(c *gin.Context) {
		prices := manager.GetPrices(ctx)
		c.JSON(http.StatusOK, prices)
	})
	r.GET("/alliance/protocol", func(c *gin.Context) {
		allianceProtocolRes, err := allianceProvider.GetProtocolsInfo(ctx)
		// allianceProtocolRes.UpdateChainsInfo.ChainsInfo.ProtocolsInfo[0].ChainId = "narwhal-1"
		// allianceProtocolRes.UpdateChainsInfo.ChainsInfo.ProtocolsInfo[1].ChainId = "harpoon-4"
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, allianceProtocolRes)
	})
	r.GET("/alliance/rebalance", func(c *gin.Context) {
		allianceRebalanceVals, err := allianceProvider.GetAllianceRedelegateReq(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, allianceRebalanceVals)
	})
	r.GET("/alliance/delegations", func(c *gin.Context) {
		allianceDelegatios, err := allianceProvider.GetAllianceInitialDelegations(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, allianceDelegatios)
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
