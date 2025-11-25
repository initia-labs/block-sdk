package base

import (
	"fmt"

	"cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"

	signer_extraction "github.com/skip-mev/block-sdk/v2/adapters/signer_extraction_adapter"
)

// LaneConfig defines the basic configurations needed for a lane.
type LaneConfig struct {
	Logger      log.Logger
	TxEncoder   sdk.TxEncoder
	TxDecoder   sdk.TxDecoder
	AnteHandler sdk.AnteHandler

	// SignerExtractor defines the interface used for extracting the expected signers of a transaction
	// from the transaction.
	SignerExtractor signer_extraction.Adapter
}

// NewLaneConfig returns a new LaneConfig. This will be embedded in a lane.
func NewLaneConfig(
	logger log.Logger,
	txEncoder sdk.TxEncoder,
	txDecoder sdk.TxDecoder,
	anteHandler sdk.AnteHandler,
	signerExtractor signer_extraction.Adapter,
) LaneConfig {
	return LaneConfig{
		Logger:          logger,
		TxEncoder:       txEncoder,
		TxDecoder:       txDecoder,
		AnteHandler:     anteHandler,
		SignerExtractor: signerExtractor,
	}
}

// ValidateBasic validates the lane configuration.
func (c *LaneConfig) ValidateBasic() error {
	if c.Logger == nil {
		return fmt.Errorf("logger cannot be nil")
	}

	if c.TxEncoder == nil {
		return fmt.Errorf("tx encoder cannot be nil")
	}

	if c.TxDecoder == nil {
		return fmt.Errorf("tx decoder cannot be nil")
	}

	if c.SignerExtractor == nil {
		return fmt.Errorf("signer extractor cannot be nil")
	}
	return nil
}
