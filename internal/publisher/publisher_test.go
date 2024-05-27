package publisher_test

import (
	"testing"

	"github.com/EvergenEnergy/remote-standby/internal/plan"
	"github.com/EvergenEnergy/remote-standby/internal/publisher"
	"github.com/stretchr/testify/assert"
)

func TestBuildCommandPayloads(t *testing.T) {
	type test struct {
		meterPower float32
		meterUnit  int
		expected   float64
	}

	tests := []test{
		{meterPower: 1234, meterUnit: 1, expected: 1.234},
		{meterPower: 1234, meterUnit: 2, expected: 1234},
		{meterPower: 1.234, meterUnit: 3, expected: 1234},
	}

	for _, tc := range tests {
		payload := publisher.BuildCommandPayload("actionvalue", plan.OptimisationInterval{
			MeterPower: plan.OptimisationValue{Value: tc.meterPower, Unit: tc.meterUnit},
		})
		assert.InDelta(t, tc.expected, payload.Value, 0.0001)
	}
}
