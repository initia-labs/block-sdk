package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/skip-mev/block-sdk/v2/x/lane/types"
)

func (k Keeper) Ratio(ctx sdk.Context, laneName string) (math.LegacyDec, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return math.LegacyZeroDec(), err
	}

	for _, lane := range params.Lanes {
		if lane.Name == laneName {
			return lane.Ratio, nil
		}
	}
	// If the lane is not found, return a negative ratio to prevent halting the chain
	return math.LegacyOneDec().Neg(), nil
}

func (k Keeper) MaxTxs(ctx sdk.Context, laneName string) (int64, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return 0, err
	}

	for _, lane := range params.Lanes {
		if lane.Name == laneName {
			return lane.MaxTxs, nil
		}
	}
	return 0, types.ErrLaneNotFound
}
