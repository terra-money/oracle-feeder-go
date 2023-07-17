package internal

import (
	"context"
	"crypto/tls"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type BaseGrpc struct {
}

func NewBaseGrpc() *BaseGrpc {
	return &BaseGrpc{}
}

func (p *BaseGrpc) Connection(
	ctx context.Context,
	nodeUrl string,
) (*grpc.ClientConn, error) {
	var authCredentials = grpc.WithTransportCredentials(insecure.NewCredentials())
	var callOptions = grpc.WithDefaultCallOptions()

	if strings.Contains(nodeUrl, "carbon") ||
		strings.Contains(nodeUrl, "pisco") ||
		strings.Contains(nodeUrl, "phoenix") {
		authCredentials = grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{}))
	}

	if strings.Contains(nodeUrl, "migaloo") {
		callOptions = grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.NewProtoCodec(nil).GRPCCodec()))
	}

	return grpc.DialContext(
		ctx,
		nodeUrl,
		authCredentials,
		callOptions,
	)
}
