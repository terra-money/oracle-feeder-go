package internal

import (
	"context"
	"crypto/tls"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	codecTypes "github.com/cosmos/cosmos-sdk/codec/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
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
	authCredentials := grpc.WithTransportCredentials(insecure.NewCredentials())
	callOptions := grpc.WithDefaultCallOptions()

	if strings.Contains(nodeUrl, ":443") {
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

func (p *BaseGrpc) GetCallOption() grpc.CallOption {
	interfaceRegistry := codecTypes.NewInterfaceRegistry()

	govtypes.RegisterInterfaces(interfaceRegistry)
	govtypesv1.RegisterInterfaces(interfaceRegistry)

	return grpc.ForceCodec(codec.NewProtoCodec(interfaceRegistry).GRPCCodec())
}
