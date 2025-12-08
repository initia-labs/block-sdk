package types

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestDefaultParamsValidate(t *testing.T) {
	params := DefaultParams()
	require.NoError(t, params.Validate())
}

func TestNewParamsSetsNames(t *testing.T) {
	params := NewParams(map[string]LaneParams{
		"lane1": {
			Name:   "ignored",
			Ratio:  math.LegacyMustNewDecFromStr("0.1"),
			MaxTxs: 1,
		},
		"lane2": {
			Ratio:  math.LegacyMustNewDecFromStr("0.2"),
			MaxTxs: 2,
		},
	})

	require.Len(t, params.Lanes, 2)

	nameToParams := make(map[string]LaneParams, len(params.Lanes))
	for _, lane := range params.Lanes {
		nameToParams[lane.Name] = lane
	}

	require.Equal(t, int64(1), nameToParams["lane1"].MaxTxs)
	require.Equal(t, int64(2), nameToParams["lane2"].MaxTxs)
	require.Equal(t, math.LegacyMustNewDecFromStr("0.1"), nameToParams["lane1"].Ratio)
	require.Equal(t, math.LegacyMustNewDecFromStr("0.2"), nameToParams["lane2"].Ratio)
}

func TestParamsValidate(t *testing.T) {
	tests := []struct {
		name    string
		params  Params
		wantErr bool
	}{
		{
			name: "valid sum equals 1",
			params: Params{Lanes: []LaneParams{
				{Name: "a", Ratio: math.LegacyMustNewDecFromStr("0.4")},
				{Name: "b", Ratio: math.LegacyMustNewDecFromStr("0.6")},
			}},
		},
		{
			name: "valid with zero ratio lane",
			params: Params{Lanes: []LaneParams{
				{Name: "a", Ratio: math.LegacyMustNewDecFromStr("0.3")},
				{Name: "default", Ratio: math.LegacyZeroDec()},
			}},
		},
		{
			name: "duplicate names",
			params: Params{Lanes: []LaneParams{
				{Name: "dup", Ratio: math.LegacyMustNewDecFromStr("0.5")},
				{Name: "dup", Ratio: math.LegacyMustNewDecFromStr("0.5")},
			}},
			wantErr: true,
		},
		{
			name: "empty name",
			params: Params{Lanes: []LaneParams{
				{Name: "", Ratio: math.LegacyMustNewDecFromStr("1")},
			}},
			wantErr: true,
		},
		{
			name: "negative ratio",
			params: Params{Lanes: []LaneParams{
				{Name: "a", Ratio: math.LegacyMustNewDecFromStr("-0.1")},
			}},
			wantErr: true,
		},
		{
			name: "ratio greater than one",
			params: Params{Lanes: []LaneParams{
				{Name: "a", Ratio: math.LegacyMustNewDecFromStr("1.1")},
			}},
			wantErr: true,
		},
		{
			name: "multiple zero ratio lanes",
			params: Params{Lanes: []LaneParams{
				{Name: "a", Ratio: math.LegacyZeroDec()},
				{Name: "b", Ratio: math.LegacyZeroDec()},
			}},
			wantErr: true,
		},
		{
			name: "sum ratios greater than one",
			params: Params{Lanes: []LaneParams{
				{Name: "a", Ratio: math.LegacyMustNewDecFromStr("0.7")},
				{Name: "b", Ratio: math.LegacyMustNewDecFromStr("0.4")},
			}},
			wantErr: true,
		},
		{
			name: "sum ratios less than one without zero lane",
			params: Params{Lanes: []LaneParams{
				{Name: "a", Ratio: math.LegacyMustNewDecFromStr("0.3")},
				{Name: "b", Ratio: math.LegacyMustNewDecFromStr("0.3")},
			}},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := tc.params.Validate()
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
