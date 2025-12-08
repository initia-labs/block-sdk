package base_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	signer_extraction "github.com/skip-mev/block-sdk/v2/adapters/signer_extraction_adapter"
	"github.com/skip-mev/block-sdk/v2/block/base"
	testutils "github.com/skip-mev/block-sdk/v2/testutils"
	lanetypes "github.com/skip-mev/block-sdk/v2/x/lane/types"
)

func (s *BaseTestSuite) TestInsert() {
	txEncoder := s.encodingConfig.TxConfig.TxEncoder()

	mempool, err := base.NewMempool("default", base.DefaultTxPriority(), signer_extraction.NewDefaultAdapter(), txEncoder, s.laneKeeper)
	s.Require().NoError(err)

	err = s.laneKeeper.SetParams(s.ctx, lanetypes.Params{
		Lanes: []lanetypes.LaneParams{
			{
				Name:   "default",
				Ratio:  math.LegacyZeroDec(),
				MaxTxs: 3,
			},
		},
	})
	s.Require().NoError(err)

	s.Run("should be able to insert a transaction", func() {
		tx, err := testutils.CreateRandomTx(
			s.encodingConfig.TxConfig,
			s.accounts[0],
			0,
			0,
			0,
			0,
			sdk.NewCoin(s.gasTokenDenom, math.NewInt(100)),
		)
		s.Require().NoError(err)

		err = mempool.Insert(s.ctx, tx)
		s.Require().NoError(err)
		s.Require().True(mempool.Contains(tx))
	})

	s.Run("cannot insert more transactions than the max", func() {
		for i := 0; i < 3; i++ {
			tx, err := testutils.CreateRandomTx(
				s.encodingConfig.TxConfig,
				s.accounts[0],
				uint64(i),
				0,
				0,
				0,
				sdk.NewCoin(s.gasTokenDenom, math.NewInt(int64(100*i))),
			)
			s.Require().NoError(err)

			err = mempool.Insert(s.ctx, tx)
			s.Require().NoError(err)
			s.Require().True(mempool.Contains(tx))
		}

		tx, err := testutils.CreateRandomTx(
			s.encodingConfig.TxConfig,
			s.accounts[0],
			10,
			0,
			0,
			0,
			sdk.NewCoin(s.gasTokenDenom, math.NewInt(100)),
		)
		s.Require().NoError(err)

		err = mempool.Insert(s.ctx, tx)
		s.Require().Error(err)
		s.Require().False(mempool.Contains(tx))
	})
}

func (s *BaseTestSuite) TestRemove() {
	txEncoder := s.encodingConfig.TxConfig.TxEncoder()
	mempool, err := base.NewMempool("default", base.DefaultTxPriority(), signer_extraction.NewDefaultAdapter(), txEncoder, s.laneKeeper)
	s.Require().NoError(err)

	err = s.laneKeeper.SetParams(s.ctx, lanetypes.Params{
		Lanes: []lanetypes.LaneParams{
			{
				Name:   "default",
				Ratio:  math.LegacyZeroDec(),
				MaxTxs: 3,
			},
		},
	})
	s.Require().NoError(err)

	s.Run("should be able to remove a transaction", func() {
		tx, err := testutils.CreateRandomTx(
			s.encodingConfig.TxConfig,
			s.accounts[0],
			0,
			0,
			0,
			0,
			sdk.NewCoin(s.gasTokenDenom, math.NewInt(100)),
		)
		s.Require().NoError(err)

		err = mempool.Insert(s.ctx, tx)
		s.Require().NoError(err)
		s.Require().True(mempool.Contains(tx))

		mempool.Remove(tx)
		s.Require().False(mempool.Contains(tx))
	})

	s.Run("should not error when removing a transaction that does not exist", func() {
		tx, err := testutils.CreateRandomTx(
			s.encodingConfig.TxConfig,
			s.accounts[0],
			0,
			0,
			0,
			0,
			sdk.NewCoin(s.gasTokenDenom, math.NewInt(100)),
		)
		s.Require().NoError(err)

		mempool.Remove(tx)
	})
}

func (s *BaseTestSuite) TestSelect() {
	s.Run("should be able to select transactions in the correct order", func() {
		txEncoder := s.encodingConfig.TxConfig.TxEncoder()
		mempool, err := base.NewMempool("default", base.DefaultTxPriority(), signer_extraction.NewDefaultAdapter(), txEncoder, s.laneKeeper)
		s.Require().NoError(err)

		err = s.laneKeeper.SetParams(s.ctx, lanetypes.Params{
			Lanes: []lanetypes.LaneParams{
				{
					Name:   "default",
					Ratio:  math.LegacyZeroDec(),
					MaxTxs: 3,
				},
			},
		})
		s.Require().NoError(err)

		tx1, err := testutils.CreateRandomTx(
			s.encodingConfig.TxConfig,
			s.accounts[1],
			0,
			0,
			0,
			0,
			sdk.NewCoin(s.gasTokenDenom, math.NewInt(100)),
		)
		s.Require().NoError(err)

		tx2, err := testutils.CreateRandomTx(
			s.encodingConfig.TxConfig,
			s.accounts[1],
			1,
			0,
			0,
			0,
			sdk.NewCoin(s.gasTokenDenom, math.NewInt(200)),
		)
		s.Require().NoError(err)

		// Insert the transactions into the mempool
		s.Require().NoError(mempool.Insert(s.ctx, tx1))
		s.Require().NoError(mempool.Insert(s.ctx, tx2))
		s.Require().Equal(2, mempool.CountTx())

		// Check that the transactions are in the correct order
		iterator := mempool.Select(s.ctx, nil)
		s.Require().NotNil(iterator)
		s.Require().Equal(tx1, iterator.Tx())

		// Check the second transaction
		iterator = iterator.Next()
		s.Require().NotNil(iterator)
		s.Require().Equal(tx2, iterator.Tx())
	})

	s.Run("should be able to select a single transaction", func() {
		txEncoder := s.encodingConfig.TxConfig.TxEncoder()
		mempool, err := base.NewMempool("default", base.DefaultTxPriority(), signer_extraction.NewDefaultAdapter(), txEncoder, s.laneKeeper)
		s.Require().NoError(err)

		err = s.laneKeeper.SetParams(s.ctx, lanetypes.Params{
			Lanes: []lanetypes.LaneParams{
				{
					Name:   "default",
					Ratio:  math.LegacyZeroDec(),
					MaxTxs: 3,
				},
			},
		})
		s.Require().NoError(err)

		tx1, err := testutils.CreateRandomTx(
			s.encodingConfig.TxConfig,
			s.accounts[0],
			0,
			0,
			0,
			0,
			sdk.NewCoin(s.gasTokenDenom, math.NewInt(100)),
		)
		s.Require().NoError(err)

		// Insert the transactions into the mempool
		s.Require().NoError(mempool.Insert(s.ctx, tx1))
		s.Require().Equal(1, mempool.CountTx())

		// Check that the transactions are in the correct order
		iterator := mempool.Select(s.ctx, nil)
		s.Require().NotNil(iterator)
		s.Require().Equal(tx1, iterator.Tx())

		iterator = iterator.Next()
		s.Require().Nil(iterator)
	})
}
