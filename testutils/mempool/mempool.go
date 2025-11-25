package mempool

import (
	"cosmossdk.io/log"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	signerextraction "github.com/skip-mev/block-sdk/v2/adapters/signer_extraction_adapter"
	"github.com/skip-mev/block-sdk/v2/block"
	"github.com/skip-mev/block-sdk/v2/block/base"
	defaultlane "github.com/skip-mev/block-sdk/v2/lanes/base"
	"github.com/skip-mev/block-sdk/v2/lanes/free"
	"github.com/skip-mev/block-sdk/v2/lanes/mev"
	"github.com/skip-mev/block-sdk/v2/testutils"
	lanekeeper "github.com/skip-mev/block-sdk/v2/x/lane/keeper"
	lanetypes "github.com/skip-mev/block-sdk/v2/x/lane/types"
)

func CreateMempool(ctx sdk.Context, laneKeeper *lanekeeper.Keeper) *block.LanedMempool {
	encodingConfig := testutils.CreateTestEncodingConfig()
	signerExtractor := signerextraction.NewDefaultAdapter()

	mevConfig := base.LaneConfig{
		SignerExtractor: signerExtractor,
		Logger:          log.NewNopLogger(),
		TxEncoder:       encodingConfig.TxConfig.TxEncoder(),
		TxDecoder:       encodingConfig.TxConfig.TxDecoder(),
		AnteHandler:     nil,
	}
	factory := mev.NewDefaultAuctionFactory(encodingConfig.TxConfig.TxDecoder(), signerExtractor)
	mevLane := mev.NewMEVLane(mevConfig, factory, factory.MatchHandler(), laneKeeper)

	freeConfig := base.LaneConfig{
		SignerExtractor: signerExtractor,
		Logger:          log.NewNopLogger(),
		TxEncoder:       encodingConfig.TxConfig.TxEncoder(),
		TxDecoder:       encodingConfig.TxConfig.TxDecoder(),
		AnteHandler:     nil,
	}
	freeLane := free.NewFreeLane(freeConfig, base.DefaultTxPriority(), free.DefaultMatchHandler(), laneKeeper)

	defaultConfig := base.LaneConfig{
		SignerExtractor: signerExtractor,
		Logger:          log.NewNopLogger(),
		TxEncoder:       encodingConfig.TxConfig.TxEncoder(),
		TxDecoder:       encodingConfig.TxConfig.TxDecoder(),
		AnteHandler:     nil,
	}
	defaultLane := defaultlane.NewDefaultLane(defaultConfig, base.DefaultMatchHandler(), laneKeeper)

	lanes := []block.Lane{mevLane, freeLane, defaultLane}
	mempool, err := block.NewLanedMempool(log.NewNopLogger(), lanes)
	if err != nil {
		panic(err)
	}

	err = laneKeeper.SetParams(ctx, lanetypes.Params{
		Lanes: []lanetypes.LaneParams{
			{
				Name:   "mev",
				Ratio:  math.LegacyMustNewDecFromStr("0.3"),
				MaxTxs: 0, // unlimited
			},
			{
				Name:   "free",
				Ratio:  math.LegacyMustNewDecFromStr("0.3"),
				MaxTxs: 0, // unlimited
			},
			{
				Name:   "default",
				Ratio:  math.LegacyZeroDec(),
				MaxTxs: 0, // unlimited
			},
		},
	})
	if err != nil {
		panic(err)
	}

	return mempool
}
