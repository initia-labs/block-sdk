package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	return math.LegacyZeroDec(), fmt.Errorf("lane %s not found", laneName)
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
	return 0, fmt.Errorf("lane %s not found", laneName)
}
