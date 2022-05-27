package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	typestx "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/sei-protocol/sei-chain/app"
	"github.com/sei-protocol/sei-chain/scripts/common"
	dextypes "github.com/sei-protocol/sei-chain/x/dex/types"
	"google.golang.org/grpc"
)

const BATCH_SIZE = 100

func init() {
	cdc := codec.NewLegacyAmino()
	interfaceRegistry := types.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)

	common.TEST_CONFIG = common.EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         marshaler,
		TxConfig:          tx.NewTxConfig(marshaler, tx.DefaultSignModes),
		Amino:             cdc,
	}
	std.RegisterLegacyAminoCodec(common.TEST_CONFIG.Amino)
	std.RegisterInterfaces(common.TEST_CONFIG.InterfaceRegistry)
	app.ModuleBasics.RegisterLegacyAminoCodec(common.TEST_CONFIG.Amino)
	app.ModuleBasics.RegisterInterfaces(common.TEST_CONFIG.InterfaceRegistry)
}

func run(
	contractAddress string,
	numberOfAccounts uint64,
	numberOfOrders uint64,
	longPriceFloor uint64,
	longPriceCeiling uint64,
	shortPriceFloor uint64,
	shortPriceCeiling uint64,
	quantityFloor uint64,
	quantityCeiling uint64,
) {
	grpcConn, _ := grpc.Dial(
		"127.0.0.1:9090",
		grpc.WithInsecure(),
	)
	defer grpcConn.Close()
	common.TX_CLIENT = typestx.NewServiceClient(grpcConn)
	userHomeDir, _ := os.UserHomeDir()
	filename := filepath.Join(userHomeDir, "outputs", "test_tx_hash")
	_ = os.Remove(filename)
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening file %s", err)
		return
	}
	common.TX_HASH_FILE = file

	var wg sync.WaitGroup
	var mu sync.Mutex
	var senders []func()

	for i := uint64(0); i < numberOfOrders/BATCH_SIZE; i++ {
		fmt.Println(fmt.Sprintf("Preparing %d-th order", i))
		accountIdx := i % numberOfAccounts
		key := common.GetKey(fmt.Sprintf("ta%d", accountIdx))
		orderPlacements := []*dextypes.OrderPlacement{}
		longPrice := i%(longPriceCeiling-longPriceFloor) + longPriceFloor
		longQuantity := uint64(rand.Intn(int(quantityCeiling)-int(quantityFloor))) + quantityFloor
		shortPrice := i%(shortPriceCeiling-shortPriceFloor) + shortPriceFloor
		shortQuantity := uint64(rand.Intn(int(quantityCeiling)-int(quantityFloor))) + quantityFloor
		for j := 0; j < BATCH_SIZE; j++ {
			orderPlacements = append(orderPlacements, &dextypes.OrderPlacement{
				Long:       true,
				Price:      longPrice,
				Quantity:   longQuantity,
				PriceDenom: "usei",
				AssetDenom: "luna",
				Open:       true,
				Limit:      true,
				Leverage:   "1",
			}, &dextypes.OrderPlacement{
				Long:       false,
				Price:      shortPrice,
				Quantity:   shortQuantity,
				PriceDenom: "usei",
				AssetDenom: "luna",
				Open:       true,
				Limit:      true,
				Leverage:   "1",
			})
		}
		amount, err := sdk.ParseCoinsNormalized(fmt.Sprintf("%d%s", longPrice*longQuantity+shortPrice*shortQuantity, "usei"))
		if err != nil {
			panic(err)
		}
		msg := dextypes.MsgPlaceOrders{
			Creator:      sdk.AccAddress(key.PubKey().Address()).String(),
			Orders:       orderPlacements,
			ContractAddr: contractAddress,
			Nonce:        i,
			Funds:        amount,
		}
		txBuilder := common.TEST_CONFIG.TxConfig.NewTxBuilder()
		_ = txBuilder.SetMsgs(&msg)
		sender := common.SendTx(key, &txBuilder, &mu, 0)
		wg.Add(1)
		senders = append(senders, func() {
			defer wg.Done()
			sender()
		})
	}

	for _, sender := range senders {
		go sender()
	}

	wg.Wait()
}

func main() {
	args := os.Args[1:]
	contractAddress := args[0]
	numberOfAccounts, _ := strconv.ParseUint(args[1], 10, 64)
	numberOfOrders, _ := strconv.ParseUint(args[2], 10, 64)
	longPriceFloor, _ := strconv.ParseUint(args[3], 10, 64)
	longPriceCeiling, _ := strconv.ParseUint(args[4], 10, 64)
	shortPriceFloor, _ := strconv.ParseUint(args[5], 10, 64)
	shortPriceCeiling, _ := strconv.ParseUint(args[6], 10, 64)
	quantityFloor, _ := strconv.ParseUint(args[7], 10, 64)
	quantityCeiling, _ := strconv.ParseUint(args[8], 10, 64)
	chainId := args[9]
	common.CHAIN_ID = chainId
	run(
		contractAddress,
		numberOfAccounts,
		numberOfOrders,
		longPriceFloor,
		longPriceCeiling,
		shortPriceFloor,
		shortPriceCeiling,
		quantityFloor,
		quantityCeiling,
	)
}
