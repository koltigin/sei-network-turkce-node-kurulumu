package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sei-protocol/sei-chain/x/dex/types"
)

func (k Keeper) SetTwap(ctx sdk.Context, twap types.Twap, contractAddr string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.TwapPrefix(contractAddr))
	b := k.cdc.MustMarshal(&twap)
	store.Set(GetKeyForTwap(twap.PriceDenom, twap.AssetDenom), b)
}

func GetKeyForTwap(priceDenom string, assetDenom string) []byte {
	return append([]byte(priceDenom), []byte(assetDenom)...)
}
