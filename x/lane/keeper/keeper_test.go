package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	testutils "github.com/skip-mev/block-sdk/v2/testutils"
	"github.com/skip-mev/block-sdk/v2/x/lane/keeper"
	"github.com/skip-mev/block-sdk/v2/x/lane/types"
)

func setupKeeper(t *testing.T) (keeper.Keeper, sdk.Context, sdk.AccAddress) {
	t.Helper()

	encCfg := testutils.CreateTestEncodingConfig()
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, storeKey, storetypes.NewTransientStoreKey("transient_test"))

	authority := sdk.AccAddress([]byte("authority"))

	return keeper.NewKeeper(encCfg.Codec, storeKey, authority.String()), testCtx.Ctx, authority
}

func TestKeeperAuthority(t *testing.T) {
	k, _, authority := setupKeeper(t)
	require.Equal(t, authority.String(), k.GetAuthority())
}

func TestKeeperGetParamsEmpty(t *testing.T) {
	k, ctx, _ := setupKeeper(t)

	params, err := k.GetParams(ctx)
	require.NoError(t, err)
	require.Empty(t, params.Lanes)
}

func TestKeeperSetGetParams(t *testing.T) {
	k, ctx, _ := setupKeeper(t)

	original := types.Params{
		Lanes: []types.LaneParams{
			{Name: "a", Ratio: types.DefaultParams().Lanes[0].Ratio, MaxTxs: 1},
			{Name: "b", Ratio: types.DefaultParams().Lanes[0].Ratio, MaxTxs: 2},
		},
	}

	require.NoError(t, k.SetParams(ctx, original))

	got, err := k.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, original, got)
}

func TestKeeperDeleteParams(t *testing.T) {
	k, ctx, _ := setupKeeper(t)

	require.NoError(t, k.SetParams(ctx, types.DefaultParams()))

	require.NoError(t, k.DeleteParams(ctx))

	params, err := k.GetParams(ctx)
	require.NoError(t, err)
	require.Empty(t, params.Lanes)
}

func TestKeeperAddLaneConfig(t *testing.T) {
	k, ctx, _ := setupKeeper(t)

	laneA := types.LaneParams{Name: "a", Ratio: math.LegacyMustNewDecFromStr("0.1")}
	laneB := types.LaneParams{Name: "b", Ratio: math.LegacyMustNewDecFromStr("0.2")}

	require.NoError(t, k.AddLaneConfig(ctx, laneA))
	require.NoError(t, k.AddLaneConfig(ctx, laneB))

	params, err := k.GetParams(ctx)
	require.NoError(t, err)

	require.Len(t, params.Lanes, 2)
	require.Equal(t, laneA.Name, params.Lanes[0].Name)
	require.Equal(t, laneB.Name, params.Lanes[1].Name)
}
