package internal

import (
	"crypto/tls"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/codec/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type BaseGrpc struct {
}

func NewBaseGrpc() *BaseGrpc {
	return &BaseGrpc{}
}

func (p *BaseGrpc) Connection(nodeUrl string, interfaceRegistry sdk.InterfaceRegistry) (*grpc.ClientConn, error) {
	var authCredentials = grpc.WithTransportCredentials(insecure.NewCredentials())

	if strings.Contains(nodeUrl, "carbon") {
		authCredentials = grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{}))
	}

	return grpc.Dial(
		nodeUrl,
		authCredentials,
		grpc.WithDefaultCallOptions(
			grpc.ForceCodec(codec.NewProtoCodec(interfaceRegistry).GRPCCodec()),
			grpc.MaxCallRecvMsgSize(1024*1024*16), // 16MB
		))
}
