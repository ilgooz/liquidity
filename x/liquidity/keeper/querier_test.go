package keeper_test

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/liquidity/app"
	"github.com/tendermint/liquidity/x/liquidity/keeper"
	"github.com/tendermint/liquidity/x/liquidity/types"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"strings"
	"testing"
)

const custom = "custom"

func getQueriedLiquidityPool(t *testing.T, ctx sdk.Context, cdc *codec.LegacyAmino, querier sdk.Querier, poolId uint64) (types.LiquidityPool, error) {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryLiquidityPool}, "/"),
		Data: cdc.MustMarshalJSON(types.QueryLiquidityPoolParams{PoolId: poolId}),
	}

	pool := types.LiquidityPool{}
	bz, err := querier(ctx, []string{types.QueryLiquidityPool}, query)
	if err != nil {
		return pool, err
	}
	require.Nil(t, cdc.UnmarshalJSON(bz, &pool))
	return pool, nil
}

func getQueriedLiquidityPools(t *testing.T, ctx sdk.Context, cdc *codec.LegacyAmino, querier sdk.Querier) (types.LiquidityPools, error) {
	queryDelParams := types.NewQueryLiquidityPoolsParams(1, 100)
	bz, errRes := cdc.MarshalJSON(queryDelParams)
	fmt.Println(bz, errRes)
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryLiquidityPools}, "/"),
		Data: bz,
	}

	pools := types.LiquidityPools{}
	bz, err := querier(ctx, []string{types.QueryLiquidityPools}, query)
	if err != nil {
		return pools, err
	}
	require.Nil(t, cdc.UnmarshalJSON(bz, &pools))
	return pools, nil
}

func TestNewQuerier(t *testing.T) {
	cdc := codec.NewLegacyAmino()
	types.RegisterLegacyAminoCodec(cdc)
	simapp := app.Setup(false)
	ctx := simapp.BaseApp.NewContext(false, tmproto.Header{})
	X := sdk.NewInt(1000000000)
	Y := sdk.NewInt(1000000000)

	addrs := app.AddTestAddrsIncremental(simapp, ctx, 20, sdk.NewInt(10000))

	querier := keeper.NewQuerier(simapp.LiquidityKeeper, cdc)

	poolId := app.TestCreatePool(t, simapp, ctx, X, Y, DenomX, DenomY, addrs[0])
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryLiquidityPool}, "/"),
		Data: cdc.MustMarshalJSON(types.QueryLiquidityPoolParams{PoolId: poolId}),
	}
	queryFailCase := abci.RequestQuery{
		Path: strings.Join([]string{"failCustom", "failRoute", "failQuery"}, "/"),
		Data: cdc.MustMarshalJSON(types.LiquidityPool{}),
	}
	pool := types.LiquidityPool{}
	bz, err := querier(ctx, []string{types.QueryLiquidityPool}, query)
	require.NoError(t, err)
	require.Nil(t, cdc.UnmarshalJSON(bz, &pool))

	bz, err = querier(ctx, []string{"fail"}, queryFailCase)
	require.Error(t, err)
	require.Error(t, cdc.UnmarshalJSON(bz, &pool))
}

func TestQueries(t *testing.T) {
	cdc := codec.NewLegacyAmino()
	types.RegisterLegacyAminoCodec(cdc)

	simapp := app.Setup(false)
	ctx := simapp.BaseApp.NewContext(false, tmproto.Header{})
	//_ = simapp.LiquidityKeeper.GetParams(ctx)

	// define test denom X, Y for Liquidity Pool
	denomX, denomY := types.AlphabeticalDenomPair(DenomX, DenomY)
	//denoms := []string{denomX, denomY}

	X := sdk.NewInt(1000000000)
	Y := sdk.NewInt(1000000000)

	addrs := app.AddTestAddrsIncremental(simapp, ctx, 20, sdk.NewInt(10000))

	poolId := app.TestCreatePool(t, simapp, ctx, X, Y, denomX, denomY, addrs[0])
	poolId2 := app.TestCreatePool(t, simapp, ctx, X, Y, denomX, "testDenom", addrs[0])
	require.Equal(t, uint64(1), poolId)
	require.Equal(t, uint64(2), poolId2)

	// begin block, init
	app.TestDepositPool(t, simapp, ctx, X, Y, addrs[1:10], poolId, true)

	querier := keeper.NewQuerier(simapp.LiquidityKeeper, cdc)

	require.Equal(t, uint64(1), poolId)
	poolRes, err := getQueriedLiquidityPool(t, ctx, cdc, querier, poolId)
	require.NoError(t, err)
	require.Equal(t, poolId, poolRes.PoolId)
	require.Equal(t, types.DefaultPoolTypeIndex, poolRes.PoolTypeIndex)
	require.Equal(t, []string{DenomX, DenomY}, poolRes.ReserveCoinDenoms)
	require.NotNil(t, poolRes.PoolCoinDenom)
	require.NotNil(t, poolRes.ReserveAccountAddress)

	poolResEmpty, err := getQueriedLiquidityPool(t, ctx, cdc, querier, uint64(3))
	require.Error(t, err)
	require.Equal(t, uint64(0), poolResEmpty.PoolId)

	poolsRes, err := getQueriedLiquidityPools(t, ctx, cdc, querier)
	require.NoError(t, err)
	require.Equal(t, 2, len(poolsRes))

}
