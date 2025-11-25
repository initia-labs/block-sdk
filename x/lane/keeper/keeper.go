package keeper

import (
	"cosmossdk.io/log"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/skip-mev/block-sdk/v2/x/lane/types"
)

type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey

	// The address that is capable of executing a MsgUpdateParams message.
	// Typically this will be the governance module's address.
	authority string
}

// NewKeeper is a wrapper around NewKeeperWithRewardsAddressProvider for backwards compatibility.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	authority string,
) Keeper {
	return Keeper{
		cdc:       cdc,
		storeKey:  storeKey,
		authority: authority,
	}
}

// Logger returns a lane module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// GetAuthority returns the address that is capable of executing a MsgUpdateParams message.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// GetParams returns the lane module's parameters.
func (k Keeper) GetParams(ctx sdk.Context) (types.Params, error) {
	store := ctx.KVStore(k.storeKey)

	key := types.KeyParams
	bz := store.Get(key)

	if len(bz) == 0 {
		return types.Params{}, nil
	}

	params := types.Params{}
	if err := params.Unmarshal(bz); err != nil {
		return types.Params{}, err
	}

	return params, nil
}

// SetParams sets the lane module's parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	store := ctx.KVStore(k.storeKey)

	bz, err := params.Marshal()
	if err != nil {
		return err
	}

	store.Set(types.KeyParams, bz)

	return nil
}

func (k Keeper) DeleteParams(ctx sdk.Context) error {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.KeyParams)
	return nil
}

func (k Keeper) AddLaneConfig(ctx sdk.Context, laneConfig types.LaneParams) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	params.Lanes = append(params.Lanes, laneConfig)

	return k.SetParams(ctx, params)
}
