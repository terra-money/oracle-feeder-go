package internal

import (
	"context"
	"crypto/tls"
	"strings"

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

	if strings.Contains(nodeUrl, "carbon") ||
		strings.Contains(nodeUrl, "pisco") ||
		strings.Contains(nodeUrl, "phoenix") {
		authCredentials = grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{}))
	}

	return grpc.DialContext(
		ctx,
		nodeUrl,
		authCredentials,
	)
}
