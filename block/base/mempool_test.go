package base_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/math"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsign "github.com/cosmos/cosmos-sdk/x/auth/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	signer_extraction "github.com/skip-mev/block-sdk/v2/adapters/signer_extraction_adapter"
	testtypes "github.com/skip-mev/block-sdk/v2/testutils/types"

	"github.com/skip-mev/block-sdk/v2/block/base"

	testkeeper "github.com/skip-mev/block-sdk/v2/testutils/keeper"
	lanetypes "github.com/skip-mev/block-sdk/v2/x/lane/types"
)

func Test_MempoolInsert(t *testing.T) {
	signerExtractor := signer_extraction.NewDefaultAdapter()
	// encodingConfig := testutils.CreateTestEncodingConfig()
	// txEncoder := encodingConfig.TxConfig.TxEncoder()

	ctx, encodingConfig, testKeepers, _ := testkeeper.NewTestSetup(t)

	ctx = ctx.WithConsensusParams(cmtproto.ConsensusParams{
		Block: &cmtproto.BlockParams{
			MaxBytes: 1000000,
			MaxGas:   1000000,
		},
	})

	// valid creation
	mempool, err := base.NewMempool(
		"default",
		base.NewDefaultTxPriority(),
		signerExtractor,
		encodingConfig.TxConfig.TxEncoder(),
		testKeepers.LaneKeeper,
	)
	require.NoError(t, err)

	err = testKeepers.LaneKeeper.SetParams(ctx, lanetypes.Params{
		Lanes: []lanetypes.LaneParams{
			{
				Name:   "default",
				Ratio:  math.LegacyMustNewDecFromStr("0.01"),
				MaxTxs: 1,
			},
		},
	})
	require.NoError(t, err)

	priv, _, addr := testdata.KeyTestPubAddr()
	defaultSignMode, err := authsign.APISignModeToInternal(encodingConfig.TxConfig.SignModeHandler().DefaultMode())
	require.NoError(t, err)

	// valid gas limit
	txBuilder := encodingConfig.TxConfig.NewTxBuilder()
	txBuilder.SetGasLimit(10000)
	txBuilder.SetMemo("")
	txBuilder.SetMsgs(&banktypes.MsgSend{FromAddress: addr.String(), ToAddress: addr.String(), Amount: sdk.Coins{}})
	sigV2 := signing.SignatureV2{
		PubKey: priv.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  defaultSignMode,
			Signature: nil,
		},
		Sequence: 1,
	}
	err = txBuilder.SetSignatures(sigV2)
	require.NoError(t, err)
	err = mempool.Insert(ctx, txBuilder.GetTx())
	require.NoError(t, err)

	// high gas limit than max gas
	txBuilder.SetGasLimit(10001)
	err = mempool.Insert(ctx, txBuilder.GetTx())
	require.ErrorContains(t, err, "exceeds max lane gas limit")

	// rollback gas limit and set memo to exceed max block space
	txBuilder.SetGasLimit(10000)
	txBuilder.SetMemo(string(make([]byte, 10000)))
	err = mempool.Insert(ctx, txBuilder.GetTx())
	require.ErrorContains(t, err, "exceeds max lane size")
}

func createTx(t *testing.T, ctx sdk.Context, encodingConfig testtypes.EncodingConfig, addr sdk.AccAddress, priv cryptotypes.PrivKey, defaultSignMode signing.SignMode, sequence uint64) sdk.Tx {
	txBuilder := encodingConfig.TxConfig.NewTxBuilder()
	txBuilder.SetGasLimit(10000)
	txBuilder.SetMemo("")
	txBuilder.SetMsgs(&banktypes.MsgSend{FromAddress: addr.String(), ToAddress: addr.String(), Amount: sdk.Coins{}})
	err := txBuilder.SetSignatures(signing.SignatureV2{
		PubKey: priv.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  defaultSignMode,
			Signature: nil,
		},
		Sequence: sequence,
	})
	require.NoError(t, err)
	return txBuilder.GetTx()
}

func Test_Mempool_SkipListBuffer(t *testing.T) {
	signerExtractor := signer_extraction.NewDefaultAdapter()
	ctx, encodingConfig, testKeepers, _ := testkeeper.NewTestSetup(t)

	ctx = ctx.WithConsensusParams(cmtproto.ConsensusParams{
		Block: &cmtproto.BlockParams{
			MaxBytes: 22 * 1024 * 1024,
			MaxGas:   200_000_000,
		},
	})

	// valid creation
	mempool, err := base.NewMempool(
		"default",
		base.NewDefaultTxPriority(),
		signerExtractor,
		encodingConfig.TxConfig.TxEncoder(),
		testKeepers.LaneKeeper,
	)
	require.NoError(t, err)

	err = testKeepers.LaneKeeper.SetParams(ctx, lanetypes.Params{
		Lanes: []lanetypes.LaneParams{
			{
				Name:   "default",
				Ratio:  math.LegacyMustNewDecFromStr("1.0"),
				MaxTxs: 10000,
			},
		},
	})
	require.NoError(t, err)

	// total create 4000 txs, but we have lrucache to limit the number of sender indices to 1000
	// so we should evict some txs from the mempool
	for range 2000 {
		priv, _, addr := testdata.KeyTestPubAddr()
		defaultSignMode, err := authsign.APISignModeToInternal(encodingConfig.TxConfig.SignModeHandler().DefaultMode())
		require.NoError(t, err)

		// create 2 transactions from the same sender
		tx1 := createTx(t, ctx, encodingConfig, addr, priv, defaultSignMode, 1)
		tx2 := createTx(t, ctx, encodingConfig, addr, priv, defaultSignMode, 2)
		err = mempool.Insert(ctx, tx1)
		require.NoError(t, err)
		err = mempool.Insert(ctx, tx2)
		require.NoError(t, err)
	}

	// it must report 4000 txs
	require.Equal(t, 4000, mempool.CountTx())

	numTxs := 0
	txes := make([]sdk.Tx, 4000)
	for iter := mempool.Select(ctx, nil); iter != nil; iter = iter.Next() {
		tx := iter.Tx()
		txes[numTxs] = tx

		signers, err := signerExtractor.GetSigners(tx)
		require.NoError(t, err)
		require.Equal(t, 1, len(signers))

		numTxs++
	}

	// check if the skip lists are empty
	require.Equal(t, 0, mempool.SkipListBufferLenForTesting())

	// check if the txes are unique
	for _, tx := range txes {
		err := mempool.Remove(tx)
		require.NoError(t, err)
	}

	// check if the txes are removed
	require.Equal(t, 0, mempool.CountTx())

	// check if the skip lists are fully filled
	require.Equal(t, 1000, mempool.SkipListBufferLenForTesting())
}
