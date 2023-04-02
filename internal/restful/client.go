package restful

import (
	"fmt"
	"strings"

	"github.com/terra-money/oracle-feeder-go/internal/restful/internal/coingecko"
	"github.com/terra-money/oracle-feeder-go/internal/restful/internal/exchangerate"
	"github.com/terra-money/oracle-feeder-go/internal/restful/internal/fer"
	"github.com/terra-money/oracle-feeder-go/internal/restful/internal/frankfurter"
	internal_types "github.com/terra-money/oracle-feeder-go/internal/types"
)

type restfulClient interface {
	FetchAndParse(symbols []string, timeout int) (map[string]internal_types.PriceBySymbol, error)
}

func NewRESTfulClient(exchange string, symbols []string) (restfulClient, error) {
	var client restfulClient
	switch strings.ToLower(exchange) {
	case "coingecko":
		client = coingecko.NewCoingeckoClient()
	case "exchangerate":
		client = exchangerate.NewExchangeRateClient()
	case "fer":
		client = fer.NewFerClient()
	case "frankfurter":
		client = frankfurter.NewFrankFurterClient()
	default:
		return nil, fmt.Errorf("unknown RESTful exchange: %s", exchange)
	}
	return client, nil
}
