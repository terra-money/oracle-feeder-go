package provider

import (
	"context"
	"fmt"
	"os"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/tendermint/tmlibs/bech32"
	"github.com/terra-money/oracle-feeder-go/internal/provider/internal"
	"github.com/terra-money/oracle-feeder-go/internal/types"

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
	internal.BaseGrpc
	privKey                    cryptoTypes.PrivKey
	nodeGrpcUrl                string
	oracleAddress              string
	allianceHubContractAddress string
	address                    sdk.AccAddress
	prefix                     string
	denom                      string
	ChainId                    string
	feederType                 types.FeederType
}

func NewTransactionsProvider(
	feederType types.FeederType,
) TransactionsProvider {
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

	var allianceHubContractAddress string
	if allianceHubContractAddress = os.Getenv("ALLIANCE_HUB_CONTRACT_ADDRESS"); len(allianceHubContractAddress) == 0 {
		panic("ALLIANCE_HUB_CONTRACT_ADDRESS env variable is not set!")
	}

	var ChainId string
	if ChainId = os.Getenv("CHAIN_ID"); len(ChainId) == 0 {
		panic("ORACLE_ADDRESS env variable is not set!")
	}

	privKeyBytes, err := hd.Secp256k1.Derive()(mnemonic, "", "m/44'/330'/0'/0/0")
	if err != nil {
		panic(err)
	}

	privKey := hd.Secp256k1.Generate()(privKeyBytes)
	address := sdk.AccAddress(privKey.PubKey().Address())
	return TransactionsProvider{
		BaseGrpc:                   *internal.NewBaseGrpc(),
		privKey:                    privKey,
		nodeGrpcUrl:                nodeGrpcUrl,
		address:                    address,
		oracleAddress:              oracleAddress,
		allianceHubContractAddress: allianceHubContractAddress,
		feederType:                 feederType,
		ChainId:                    ChainId,
		prefix:                     "terra",
		denom:                      "uluna",
	}
}

func (p *TransactionsProvider) SubmitAlliancesTransaction(
	ctx context.Context,
	msg []byte,
) (string, error) {
	// Get the bech address and...
	bech32Addr, err := bech32.ConvertAndEncode(p.prefix, p.address)
	if err != nil {
		return "", err
	}
	// ... build the messages to be signed
	msgs := []sdk.Msg{
		&wasmtypes.MsgExecuteContract{
			Sender:   bech32Addr,
			Contract: p.getContractAddress(),
			Msg:      msg,
			Funds:    nil,
		},
	}
	var account authTypes.AccountI
	txBuilder, txConfig, interfaceRegistry := p.getTxClients()

	// create gRPC connection
	grpcConn, err := p.BaseGrpc.Connection(ctx, p.nodeGrpcUrl)
	if err != nil {
		return "", err
	}
	defer grpcConn.Close()

	fromAddress, err := bech32.ConvertAndEncode(p.prefix, p.address)
	if err != nil {
		return "", err
	}

	authClient := authTypes.NewQueryClient(grpcConn)
	accRes, err := authClient.Account(ctx, &authTypes.QueryAccountRequest{
		Address: fromAddress,
	})
	if err != nil {
		return "", err
	}

	err = interfaceRegistry.UnpackAny(accRes.Account, &account)
	if err != nil {
		return "", err
	}
	accSeq := account.GetSequence()

	pubKey := p.privKey.PubKey()
	signMode := txConfig.SignModeHandler().DefaultMode()

	signerData := signing.SignerData{
		ChainID:       p.ChainId,
		AccountNumber: account.GetAccountNumber(),
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
		return "", err
	}

	err = txBuilder.SetSignatures(sigv2)
	if err != nil {
		return "", err
	}

	// simulate transaction to get gas cost and see if it will fail
	simulateBytes, err := txConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return "", err
	}
	queryClient := txTypes.NewServiceClient(grpcConn)
	simRes, err := queryClient.Simulate(ctx, &txTypes.SimulateRequest{
		TxBytes: simulateBytes,
	})
	if err != nil {
		return "", err
	}

	// set the gas needed with some allowance
	gasUsed := simRes.GetGasInfo().GetGasUsed()

	// calculate fee
	fee := sdk.NewIntFromUint64(uint64(float64(gasUsed) * 0.0155 * 1.5))

	// set fee amount from gasused
	txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(p.denom, fee)))

	txBuilder.SetGasLimit(uint64(float64(gasUsed) * 1.5))
	if err != nil {
		return "", err
	}

	// sign the final message with the private key
	bytesToSign, err := txConfig.SignModeHandler().GetSignBytes(signMode, signerData, txBuilder.GetTx())
	if err != nil {
		return "", err
	}
	sig, err := p.privKey.Sign(bytesToSign)
	if err != nil {
		return "", err
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
		return "", err
	}

	// encode and broadcast transaction
	txBytes, err := txConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return "", err
	}
	// Increase cached account sequence before broadcasting tx
	bRes, err := queryClient.BroadcastTx(ctx,
		&txTypes.BroadcastTxRequest{
			Mode:    txTypes.BroadcastMode_BROADCAST_MODE_SYNC,
			TxBytes: txBytes,
		},
	)
	if err != nil {
		return "", err
	}
	if bRes.TxResponse.Code != 0 {
		return "", fmt.Errorf("tx failed with code %d, %s", bRes.TxResponse.Code, bRes.TxResponse.RawLog)
	}
	return bRes.TxResponse.TxHash, err
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

func (p *TransactionsProvider) getContractAddress() string {
	if p.feederType == types.AllianceUpdateRewards ||
		p.feederType == types.AllianceRebalanceEmissions ||
		p.feederType == types.AllianceRebalanceFeeder ||
		p.feederType == types.AllianceInitialDelegation {
		return p.allianceHubContractAddress
	} else if p.feederType == types.AllianceOracleFeeder {
		return p.oracleAddress
	}

	panic("Unknown feeder type " + p.feederType)
}
