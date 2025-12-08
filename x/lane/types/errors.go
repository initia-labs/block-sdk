package types

import (
	errorsmod "cosmossdk.io/errors"
)

// Lane Errors
var (
	// ErrLaneNotFound error for the lane not found
	ErrLaneNotFound = errorsmod.Register(ModuleName, 2, "lane not found")
)
