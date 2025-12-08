package integration_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/skip-mev/chaintestutil/encoding"
	"github.com/stretchr/testify/suite"

	testkeeper "github.com/skip-mev/block-sdk/v2/testutils/keeper"
	testtypes "github.com/skip-mev/block-sdk/v2/testutils/types"
)

type IntegrationTestSuite struct {
	suite.Suite
	testkeeper.TestKeepers
	testkeeper.TestMsgServers

	encCfg encoding.TestEncodingConfig
	ctx    sdk.Context
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupTest() {
	// s.encCfg = encoding.MakeTestEncodingConfig(func(registry types.InterfaceRegistry) {
	// 	auctiontypes.RegisterInterfaces(registry)
	// })

	var encodingConfig testtypes.EncodingConfig
	s.ctx, encodingConfig, s.TestKeepers, s.TestMsgServers = testkeeper.NewTestSetup(s.T())

	s.encCfg = encoding.TestEncodingConfig{
		InterfaceRegistry: encodingConfig.InterfaceRegistry,
		Codec:             encodingConfig.Codec,
		TxConfig:          encodingConfig.TxConfig,
		Amino:             encodingConfig.Amino,
	}
}
