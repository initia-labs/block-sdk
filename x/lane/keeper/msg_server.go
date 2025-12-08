package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/skip-mev/block-sdk/v2/x/lane/types"
)

var _ types.MsgServer = MsgServer{}

// MsgServer is the wrapper for the lane module's msg service.
type MsgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the lane MsgServer interface.
func NewMsgServerImpl(keeper Keeper) *MsgServer {
	return &MsgServer{Keeper: keeper}
}

func (m MsgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// ensure that the message signer is the authority
	if msg.Authority != m.Keeper.GetAuthority() {
		return nil, fmt.Errorf("this message can only be executed by the authority; expected %s, got %s", m.Keeper.GetAuthority(), msg.Authority)
	}

	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}

	if err := m.Keeper.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
