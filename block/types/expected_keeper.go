package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type LaneKeeper interface {
	Ratio(ctx sdk.Context, laneName string) (math.LegacyDec, error)
	MaxTxs(ctx sdk.Context, laneName string) (int64, error)
}
