package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/tendermint/tmlibs/bech32"
	"github.com/terra-money/oracle-feeder-go/internal/types"
	pkgtypes "github.com/terra-money/oracle-feeder-go/pkg/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdktypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	cryptoTypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txTypes "github.com/cosmos/cosmos-sdk/types/tx"
	txsigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type TransactionsProvider struct {
	privKey       cryptoTypes.PrivKey
	nodeGrpcUrl   string
	oracleAddress string
	address       sdk.AccAddress
	prefix        string
	denom         string
	chainId       string
}

func NewTransactionsProvider() TransactionsProvider {
	var mnemonic string
	if mnemonic = os.Getenv("MNEMONIC"); len(mnemonic) == 0 {
		panic("MNEMONIC env variable is not set!")
	}

	var nodeGrpcUrl string
	if nodeGrpcUrl = os.Getenv("NODE_GRPC_URL"); len(nodeGrpcUrl) == 0 {
		panic("NODE_GRPC_URL env variable is not set!")
	}

	var oracleAddress string
	if oracleAddress = os.Getenv("ORACLE_ADDRESS"); len(oracleAddress) == 0 {
		panic("ORACLE_ADDRESS env variable is not set!")
	}

	privKeyBytes, err := hd.Secp256k1.Derive()(mnemonic, "", "m/44'/330'/0'/0/0")
	if err != nil {
		panic(err)
	}

	privKey := hd.Secp256k1.Generate()(privKeyBytes)
	address := sdk.AccAddress(privKey.PubKey().Address())

	return TransactionsProvider{
		privKey:       privKey,
		nodeGrpcUrl:   nodeGrpcUrl,
		address:       address,
		oracleAddress: oracleAddress,
		chainId:       "pisco-1",
		prefix:        "terra",
		denom:         "uluna",
	}
}
func (p *TransactionsProvider) ParseAlliancesTransaction(protocolRes *types.AllianceProtocolRes) (msg sdk.Msg, err error) {
	bech32Addr, err := bech32.ConvertAndEncode(p.prefix, p.address)
	if err != nil {
		return msg, err
	}

	executeMsg := pkgtypes.NewMsgUpdateChainsInfo(*protocolRes)
	executeB, err := json.Marshal(executeMsg)
	if err != nil {
		return msg, err
	}

	fmt.Printf("%s", executeB)
	// This needs to be a smart contract send execution
	msg = &wasmtypes.MsgExecuteContract{
		Sender:   bech32Addr,
		Contract: p.oracleAddress,
		Msg:      executeB,
		Funds:    nil,
	}
	return msg, nil
}

func (p *TransactionsProvider) SubmitAlliancesTransaction(
	ctx context.Context,
	msgs []sdk.Msg,
) (*string, error) {
	var faucetAccount authTypes.AccountI
	txBuilder, txConfig, interfaceRegistry := p.getTxClients()

	// create gRPC connection
	grpcConn, err := p.getRPCConnection(p.nodeGrpcUrl, interfaceRegistry)
	if err != nil {
		return nil, err
	}
	defer grpcConn.Close()

	fromAddress, err := bech32.ConvertAndEncode(p.prefix, p.address)
	if err != nil {
		return nil, err
	}

	authClient := authTypes.NewQueryClient(grpcConn)
	accRes, err := authClient.Account(ctx, &authTypes.QueryAccountRequest{
		Address: fromAddress,
	})
	if err != nil {
		return nil, err
	}

	err = interfaceRegistry.UnpackAny(accRes.Account, &faucetAccount)
	if err != nil {
		return nil, err
	}
	accSeq := faucetAccount.GetSequence()

	pubKey := p.privKey.PubKey()
	signMode := txConfig.SignModeHandler().DefaultMode()

	signerData := signing.SignerData{
		ChainID:       p.chainId,
		AccountNumber: faucetAccount.GetAccountNumber(),
		Sequence:      accSeq,
	}
	sigData := txsigning.SingleSignatureData{
		SignMode:  signMode,
		Signature: nil,
	}
	sigv2 := txsigning.SignatureV2{
		PubKey:   pubKey,
		Data:     &sigData,
		Sequence: accSeq,
	}

	// build txn
	err = txBuilder.SetMsgs(msgs...)
	if err != nil {
		return nil, err
	}

	err = txBuilder.SetSignatures(sigv2)
	if err != nil {
		return nil, err
	}

	// simulate transaction to get gas cost and see if it will fail
	simulateBytes, err := txConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, err
	}
	queryClient := txTypes.NewServiceClient(grpcConn)
	simRes, err := queryClient.Simulate(ctx, &txTypes.SimulateRequest{
		TxBytes: simulateBytes,
	})
	if err != nil {
		return nil, err
	}

	// set the gas needed with some allowance
	gasUsed := simRes.GetGasInfo().GetGasUsed()

	// set fee amount from gasused
	txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(p.denom, sdk.NewIntFromUint64(uint64(float64(gasUsed)*1.2)))))

	txBuilder.SetGasLimit(uint64(float64(gasUsed) * 1.5))
	if err != nil {
		return nil, err
	}

	// sign the final message with the private key
	bytesToSign, err := txConfig.SignModeHandler().GetSignBytes(signMode, signerData, txBuilder.GetTx())
	if err != nil {
		return nil, err
	}
	sig, err := p.privKey.Sign(bytesToSign)
	if err != nil {
		return nil, err
	}
	sigData = txsigning.SingleSignatureData{
		SignMode:  signMode,
		Signature: sig,
	}
	sigv2 = txsigning.SignatureV2{
		PubKey:   pubKey,
		Data:     &sigData,
		Sequence: accSeq,
	}
	err = txBuilder.SetSignatures(sigv2)
	if err != nil {
		return nil, err
	}

	// encode and broadcast transaction
	txBytes, err := txConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, err
	}
	// Increase cached account sequence before broadcasting tx
	bRes, err := queryClient.BroadcastTx(ctx,
		&txTypes.BroadcastTxRequest{
			Mode:    txTypes.BroadcastMode_BROADCAST_MODE_BLOCK,
			TxBytes: txBytes,
		},
	)
	if err != nil {
		return nil, err
	}
	return &bRes.TxResponse.TxHash, err
}

func (p *TransactionsProvider) getTxClients() (client.TxBuilder, client.TxConfig, sdktypes.InterfaceRegistry) {
	amino := codec.NewLegacyAmino()
	interfaceRegistry := sdktypes.NewInterfaceRegistry()
	protoCodec := codec.NewProtoCodec(interfaceRegistry)
	txConfig := tx.NewTxConfig(protoCodec, tx.DefaultSignModes)

	std.RegisterLegacyAminoCodec(amino)
	std.RegisterInterfaces(interfaceRegistry)

	authTypes.RegisterLegacyAminoCodec(amino)
	authTypes.RegisterInterfaces(interfaceRegistry)

	txBuilder := txConfig.NewTxBuilder()
	txBuilder.SetMemo("Alliance Oracle designed by Terra Devs")
	txBuilder.SetTimeoutHeight(0)

	return txBuilder, txConfig, interfaceRegistry
}

func (p *TransactionsProvider) getRPCConnection(nodeUrl string, interfaceRegistry sdktypes.InterfaceRegistry) (*grpc.ClientConn, error) {
	return grpc.Dial(
		nodeUrl,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.NewProtoCodec(interfaceRegistry).GRPCCodec())),
	)
}
