package base_test

import (
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	testutils "github.com/skip-mev/block-sdk/v2/testutils"
	testkeeper "github.com/skip-mev/block-sdk/v2/testutils/keeper"
	testtypes "github.com/skip-mev/block-sdk/v2/testutils/types"

	tmprototypes "github.com/cometbft/cometbft/proto/tendermint/types"
	lanekeeper "github.com/skip-mev/block-sdk/v2/x/lane/keeper"
)

type BaseTestSuite struct {
	suite.Suite

	ctx            sdk.Context
	encodingConfig testtypes.EncodingConfig
	random         *rand.Rand
	accounts       []testutils.Account
	gasTokenDenom  string

	laneKeeper lanekeeper.Keeper
}

func TestBaseTestSuite(t *testing.T) {
	suite.Run(t, new(BaseTestSuite))
}

func (s *BaseTestSuite) SetupTest() {
	// Set up basic TX encoding config.
	// s.encodingConfig = testutils.CreateTestEncodingConfig()

	ctx, encodingConfig, testKeepers, _ := testkeeper.NewTestSetup(s.T())
	s.ctx = ctx.WithConsensusParams(tmprototypes.ConsensusParams{
		Block: &tmprototypes.BlockParams{
			MaxBytes: 1000000000000000,
			MaxGas:   1000000000000000,
		},
	})
	s.encodingConfig = encodingConfig
	s.laneKeeper = testKeepers.LaneKeeper

	// Create a few random accounts
	s.random = rand.New(rand.NewSource(1))
	s.accounts = testutils.RandomAccounts(s.random, 5)
	s.gasTokenDenom = "stake"

	// key := storetypes.NewKVStoreKey(types.StoreKey)
	// s.ctx = testutil.DefaultContext(key, storetypes.NewTransientStoreKey("transient_key"))
}
