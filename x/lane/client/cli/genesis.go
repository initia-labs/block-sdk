package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"cosmossdk.io/math"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	blocklanetypes "github.com/skip-mev/block-sdk/v2/x/lane/types"
)

func ConfigureLanes(mbm module.BasicManager, txEncCfg client.TxEncodingConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configure-lanes [lane-config-path]",
		Short: "Configure lanes",
		Args:  cobra.ExactArgs(1),
		Long: fmt.Sprintf(`Configure lanes with the lane config file.

Example:
$ %s configure-lanes config.json --home=/path/to/home/dir 
`, version.AppName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverCtx := server.GetServerContextFromCmd(cmd)
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			cdc := clientCtx.Codec

			laneConfigPath := args[0]
			laneConfigBz, err := os.ReadFile(laneConfigPath)
			if err != nil {
				return fmt.Errorf("failed to read lane config file: %w", err)
			}
			var laneConfig []LaneConfig
			if err := json.Unmarshal(laneConfigBz, &laneConfig); err != nil {
				return fmt.Errorf("failed to unmarshal lane config file: %w", err)
			}

			config := serverCtx.Config
			config.SetRoot(clientCtx.HomeDir)

			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			laneState := blocklanetypes.GetGenesisStateFromAppState(cdc, appState)

			laneParams := make(map[string]blocklanetypes.LaneParams)
			for _, lane := range laneConfig {
				laneParams[lane.Name] = blocklanetypes.LaneParams{
					Name:   lane.Name,
					Ratio:  math.LegacyMustNewDecFromStr(lane.Ratio),
					MaxTxs: int64(lane.MaxTxs),
				}
			}
			laneState.Params = blocklanetypes.NewParams(laneParams)

			err = laneState.Params.Validate()
			if err != nil {
				return fmt.Errorf("failed to validate lane params: %w", err)
			}

			blockLaneGenStateBz, err := cdc.MarshalJSON(laneState)
			if err != nil {
				return fmt.Errorf("failed to marshal lane genesis state: %w", err)
			}
			appState[blocklanetypes.ModuleName] = blockLaneGenStateBz

			if err = mbm.ValidateGenesis(cdc, txEncCfg, appState); err != nil {
				return errors.Wrap(err, "failed to validate genesis state")
			}
			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}

			genDoc.AppState = appStateJSON
			if err = genutil.ExportGenesisFile(genDoc, config.GenesisFile()); err != nil {
				return errors.New("Failed to export genesis file")
			}

			return nil
		},
	}

	return cmd
}
