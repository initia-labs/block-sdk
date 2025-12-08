package keeper

import (
	"errors"

	"cosmossdk.io/collections"
	"cosmossdk.io/log"

	corestoretypes "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/skip-mev/block-sdk/v2/x/lane/types"
)

type Keeper struct {
	cdc          codec.BinaryCodec
	storeService corestoretypes.KVStoreService

	Schema collections.Schema
	Params collections.Item[types.Params]

	// The address that is capable of executing a MsgUpdateParams message.
	// Typically this will be the governance module's address.
	authority string
}

// NewKeeper is a new keeper for the lane module.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService corestoretypes.KVStoreService,

	authority string,
) Keeper {
	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		cdc:          cdc,
		storeService: storeService,
		Params:       collections.NewItem(sb, types.KeyParams, "params", codec.CollValue[types.Params](cdc)),
		authority:    authority,
	}
	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
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
	params, err := k.Params.Get(ctx)
	if errors.Is(err, collections.ErrNotFound) {
		return types.Params{}, nil
	} else if err != nil {
		return types.Params{}, err
	}
	return params, nil
}

// SetParams sets the lane module's parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	return k.Params.Set(ctx, params)
}

func (k Keeper) DeleteParams(ctx sdk.Context) error {
	return k.Params.Remove(ctx)
}

func (k Keeper) AddLaneConfig(ctx sdk.Context, laneConfig types.LaneParams) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	params.Lanes = append(params.Lanes, laneConfig)

	if err := params.ValidatePartialParams(); err != nil {
		return err
	}

	return k.SetParams(ctx, params)
}
