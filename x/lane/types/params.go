package types

import (
	"errors"
	fmt "fmt"

	"cosmossdk.io/math"
)

// NewParams returns a new Params instance with the provided values.
func NewParams(
	lanes map[string]LaneParams,
) Params {
	lp := make([]LaneParams, 0, len(lanes))
	for name, lane := range lanes {
		lane.Name = name
		lp = append(lp, lane)
	}
	return Params{
		Lanes: lp,
	}
}

// DefaultParams returns the default parameters for the lane module.
func DefaultParams() Params {
	return Params{
		Lanes: []LaneParams{
			{
				Name:   "default",
				Ratio:  math.LegacyZeroDec(),
				MaxTxs: 0,
			},
		},
	}
}

// Validate performs basic validation on the parameters.
func (p Params) Validate() error {
	seenZeroMaxBlockSpace := false
	sumRatio := math.LegacyZeroDec()
	laneNames := make(map[string]struct{})
	for _, lane := range p.Lanes {
		if _, ok := laneNames[lane.Name]; ok {
			return fmt.Errorf("lane name %s is already defined", lane.Name)
		}
		laneNames[lane.Name] = struct{}{}
		if lane.Name == "" {
			return errors.New("lane name cannot be empty")
		}
		if lane.Ratio.IsNegative() || lane.Ratio.GT(math.LegacyOneDec()) {
			return fmt.Errorf("lane ratio cannot be negative or greater than 1; %s", lane.Ratio)
		}
		sumRatio = sumRatio.Add(lane.Ratio)
		if lane.Ratio.IsZero() {
			if seenZeroMaxBlockSpace {
				return fmt.Errorf("multiple lanes with zero max block space")
			}
			seenZeroMaxBlockSpace = true
		}
	}

	if sumRatio.GT(math.LegacyOneDec()) {
		return fmt.Errorf("sum of lane ratios cannot be greater than 1; %s", sumRatio)
	} else if !seenZeroMaxBlockSpace && sumRatio.LT(math.LegacyOneDec()) {
		return fmt.Errorf("sum of lane ratios cannot be less than 1; %s", sumRatio)
	}
	return nil
}
