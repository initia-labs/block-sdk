// Package keeper provides methods to initialize SDK keepers with local storage for test purposes
package keeper

import (
	"testing"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	testkeeper "github.com/skip-mev/chaintestutil/keeper"
	"github.com/skip-mev/chaintestutil/sample"
	"github.com/stretchr/testify/require"

	auctionkeeper "github.com/skip-mev/block-sdk/v2/x/auction/keeper"
	auctiontypes "github.com/skip-mev/block-sdk/v2/x/auction/types"

	lanekeeper "github.com/skip-mev/block-sdk/v2/x/lane/keeper"
	lanetypes "github.com/skip-mev/block-sdk/v2/x/lane/types"

	testtypes "github.com/skip-mev/block-sdk/v2/testutils/types"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// TestKeepers holds all keepers used during keeper tests for all modules
type TestKeepers struct {
	testkeeper.TestKeepers
	AuctionKeeper auctionkeeper.Keeper
	LaneKeeper    lanekeeper.Keeper
}

// TestMsgServers holds all message servers used during keeper tests for all modules
type TestMsgServers struct {
	testkeeper.TestMsgServers
	AuctionMsgServer auctiontypes.MsgServer
	LaneMsgServer    lanetypes.MsgServer
}

var additionalMaccPerms = map[string][]string{
	auctiontypes.ModuleName: nil,
	lanetypes.ModuleName:    nil,
}

// NewTestSetup returns initialized instances of all the keepers and message servers of the modules
func NewTestSetup(t testing.TB, options ...testkeeper.SetupOption) (sdk.Context, testtypes.EncodingConfig, TestKeepers, TestMsgServers) {
	options = append(options, testkeeper.WithAdditionalModuleAccounts(additionalMaccPerms))

	_, tk, tms := testkeeper.NewTestSetup(t, options...)

	// initialize extra keeper
	auctionKeeper := Auction(tk.Initializer, tk.AccountKeeper, tk.BankKeeper, tk.DistrKeeper, tk.StakingKeeper)

	// initialize msg servers
	auctionMsgSrv := auctionkeeper.NewMsgServerImpl(auctionKeeper)

	laneKeeper := Lane(tk.Initializer, tk.AccountKeeper, tk.BankKeeper, tk.DistrKeeper, tk.StakingKeeper)
	require.NoError(t, tk.Initializer.LoadLatest())

	laneMsgSrv := lanekeeper.NewMsgServerImpl(laneKeeper)

	ctx := sdk.NewContext(tk.Initializer.StateStore, tmproto.Header{
		Time:   testkeeper.ExampleTimestamp,
		Height: testkeeper.ExampleHeight,
	}, false, log.NewNopLogger())

	err := auctionKeeper.SetParams(ctx, auctiontypes.DefaultParams())
	require.NoError(t, err)

	err = laneKeeper.SetParams(ctx, lanetypes.DefaultParams())
	require.NoError(t, err)

	testKeepers := TestKeepers{
		TestKeepers:   tk,
		AuctionKeeper: auctionKeeper,
		LaneKeeper:    laneKeeper,
	}

	testMsgServers := TestMsgServers{
		TestMsgServers:   tms,
		AuctionMsgServer: auctionMsgSrv,
		LaneMsgServer:    laneMsgSrv,
	}

	tk.Initializer.Codec = sample.Codec(
		banktypes.RegisterInterfaces,
		cryptocodec.RegisterInterfaces,
		auctiontypes.RegisterInterfaces,
		lanetypes.RegisterInterfaces,
		stakingtypes.RegisterInterfaces,
	)

	encodingConfig := testtypes.EncodingConfig{
		InterfaceRegistry: tk.Initializer.Codec.InterfaceRegistry(),
		Codec:             tk.Initializer.Codec,
		TxConfig:          tx.NewTxConfig(tk.Initializer.Codec, tx.DefaultSignModes),
		Amino:             codec.NewLegacyAmino(),
	}

	return ctx, encodingConfig, testKeepers, testMsgServers
}

// Auction initializes the auction module using the testkeepers intializer.
func Auction(
	initializer *testkeeper.Initializer,
	authKeeper authkeeper.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
	distrKeeper distrkeeper.Keeper,
	stakingKeeper *stakingkeeper.Keeper,
) auctionkeeper.Keeper {
	storeKey := storetypes.NewKVStoreKey(auctiontypes.StoreKey)
	initializer.StateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, initializer.DB)

	return auctionkeeper.NewKeeper(
		initializer.Codec,
		storeKey,
		authKeeper,
		bankKeeper,
		distrKeeper,
		stakingKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
}

// Lane initializes the lane module using the testkeepers intializer.
func Lane(
	initializer *testkeeper.Initializer,
	authKeeper authkeeper.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
	distrKeeper distrkeeper.Keeper,
	stakingKeeper *stakingkeeper.Keeper,
) lanekeeper.Keeper {
	storeKey := storetypes.NewKVStoreKey(lanetypes.StoreKey)
	initializer.StateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, initializer.DB)

	return lanekeeper.NewKeeper(
		initializer.Codec,
		storeKey,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
}
