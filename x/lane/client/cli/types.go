package cli

type LaneConfig struct {
	Name   string `json:"name"`
	Ratio  string `json:"ratio"`
	MaxTxs int    `json:"max_txs"`
}
