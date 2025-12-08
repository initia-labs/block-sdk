package testutils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	signer_extraction "github.com/skip-mev/block-sdk/v2/adapters/signer_extraction_adapter"
	"github.com/skip-mev/block-sdk/v2/block/base"
	"github.com/skip-mev/block-sdk/v2/lanes/mev"
	testutils "github.com/skip-mev/block-sdk/v2/testutils"
	testkeeper "github.com/skip-mev/block-sdk/v2/testutils/keeper"
	testtypes "github.com/skip-mev/block-sdk/v2/testutils/types"

	lanekeeper "github.com/skip-mev/block-sdk/v2/x/lane/keeper"
	lanetypes "github.com/skip-mev/block-sdk/v2/x/lane/types"

	tmprototypes "github.com/cometbft/cometbft/proto/tendermint/types"
)

type MEVLaneTestSuiteBase struct {
	suite.Suite

	EncCfg        testtypes.EncodingConfig
	Config        mev.Factory
	Ctx           sdk.Context
	Accounts      []testutils.Account
	GasTokenDenom string

	laneKeeper lanekeeper.Keeper
}

func (s *MEVLaneTestSuiteBase) SetupTest() {
	ctx, encodingConfig, testKeepers, _ := testkeeper.NewTestSetup(s.T())

	// Init encoding config
	// s.EncCfg = testutils.CreateTestEncodingConfig()
	s.EncCfg = encodingConfig
	s.Config = mev.NewDefaultAuctionFactory(s.EncCfg.TxConfig.TxDecoder(), signer_extraction.NewDefaultAdapter())
	// testCtx := testutil.DefaultContextWithDB(s.T(), storetypes.NewKVStoreKey("test"), storetypes.NewTransientStoreKey("transient_test"))
	s.Ctx = ctx.WithExecMode(sdk.ExecModePrepareProposal).WithConsensusParams(tmprototypes.ConsensusParams{
		Block: &tmprototypes.BlockParams{
			MaxBytes: 10000,
			MaxGas:   10000,
		},
	})
	s.Ctx = s.Ctx.WithBlockHeight(1)

	// Init accounts
	random := rand.New(rand.NewSource(time.Now().Unix()))
	s.Accounts = testutils.RandomAccounts(random, 10)
	s.GasTokenDenom = "stake"
	s.laneKeeper = testKeepers.LaneKeeper
}

func (s *MEVLaneTestSuiteBase) InitLane(
	maxBlockSpace math.LegacyDec,
	expectedExecution map[sdk.Tx]bool,
	matchAll bool,
) *mev.MEVLane {
	config := base.NewLaneConfig(
		log.NewNopLogger(),
		s.EncCfg.TxConfig.TxEncoder(),
		s.EncCfg.TxConfig.TxDecoder(),
		s.SetUpAnteHandler(expectedExecution),
		signer_extraction.NewDefaultAdapter(),
	)

	factory := mev.NewDefaultAuctionFactory(s.EncCfg.TxConfig.TxDecoder(), signer_extraction.NewDefaultAdapter())
	matchHandler := factory.MatchHandler()
	if matchAll {
		matchHandler = func(_ sdk.Context, _ sdk.Tx) bool {
			return true
		}
	}

	err := s.laneKeeper.SetParams(s.Ctx, lanetypes.Params{
		Lanes: []lanetypes.LaneParams{
			{
				Name:   "mev",
				Ratio:  maxBlockSpace,
				MaxTxs: 0, // unlimited
			},
		},
	})
	s.Require().NoError(err)

	return mev.NewMEVLane(config, factory, matchHandler, s.laneKeeper)
}

func (s *MEVLaneTestSuiteBase) SetUpAnteHandler(expectedExecution map[sdk.Tx]bool) sdk.AnteHandler {
	txCache := make(map[string]bool)
	for tx, pass := range expectedExecution {
		bz, err := s.EncCfg.TxConfig.TxEncoder()(tx)
		s.Require().NoError(err)

		hash := sha256.Sum256(bz)
		hashStr := hex.EncodeToString(hash[:])
		txCache[hashStr] = pass
	}

	anteHandler := func(ctx sdk.Context, tx sdk.Tx, _ bool) (newCtx sdk.Context, err error) {
		bz, err := s.EncCfg.TxConfig.TxEncoder()(tx)
		s.Require().NoError(err)

		hash := sha256.Sum256(bz)
		hashStr := hex.EncodeToString(hash[:])

		pass, found := txCache[hashStr]
		if !found {
			return ctx, fmt.Errorf("tx not found")
		}

		if pass {
			return ctx, nil
		}

		return ctx, fmt.Errorf("tx failed")
	}

	return anteHandler
}
