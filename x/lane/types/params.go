package types

import (
	"errors"
	fmt "fmt"

	"cosmossdk.io/math"
)

var (
	DefaultLaneParams = []LaneParams{
		{
			Name:   "default",
			Ratio:  math.LegacyZeroDec(),
			MaxTxs: 0,
		},
	}
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
		Lanes: DefaultLaneParams,
	}
}

// validateCommon performs validation common to both full and partial params.
func (p Params) validateCommon() (seenZeroMaxBlockSpace bool, sumRatio math.LegacyDec, err error) {
	sumRatio = math.LegacyZeroDec()
	laneNames := make(map[string]struct{})
	for _, lane := range p.Lanes {
		if _, ok := laneNames[lane.Name]; ok {
			return false, sumRatio, fmt.Errorf("lane name %s is already defined", lane.Name)
		}
		laneNames[lane.Name] = struct{}{}
		if lane.Name == "" {
			return false, sumRatio, errors.New("lane name cannot be empty")
		}
		if lane.Ratio.IsNegative() || lane.Ratio.GT(math.LegacyOneDec()) {
			return false, sumRatio, fmt.Errorf("lane ratio cannot be negative or greater than 1; %s", lane.Ratio)
		}
		sumRatio = sumRatio.Add(lane.Ratio)
		if lane.Ratio.IsZero() {
			if seenZeroMaxBlockSpace {
				return false, sumRatio, fmt.Errorf("multiple lanes with zero max block space")
			}
			seenZeroMaxBlockSpace = true
		}
	}
	if sumRatio.GT(math.LegacyOneDec()) {
		return seenZeroMaxBlockSpace, sumRatio, fmt.Errorf("sum of lane ratios cannot be greater than 1; %s", sumRatio)
	}
	return seenZeroMaxBlockSpace, sumRatio, nil
}

func (p Params) Validate() error {
	seenZeroMaxBlockSpace, sumRatio, err := p.validateCommon()
	if err != nil {
		return err
	}
	if !seenZeroMaxBlockSpace && sumRatio.LT(math.LegacyOneDec()) {
		return fmt.Errorf("sum of lane ratios cannot be less than 1; %s", sumRatio)
	}
	return nil
}

func (p Params) ValidatePartialParams() error {
	_, _, err := p.validateCommon()
	return err
}
