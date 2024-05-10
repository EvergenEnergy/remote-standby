package plan_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/EvergenEnergy/remote-standby/internal/plan"
	"github.com/stretchr/testify/assert"
)

func TestWritesAndReadsAPlan(t *testing.T) {
	optPlan := plan.OptimisationPlan{
		SiteID:       "test-site",
		SetpointType: 1,
		OptimisationIntervals: []plan.OptimisationInterval{
			{
				Interval: plan.OptimisationIntervalTimestamp{
					StartTime: plan.OptimisationTimestamp{Seconds: 1715319000},
					EndTime:   plan.OptimisationTimestamp{Seconds: 1715319900},
				},
				BatteryPower: plan.OptimisationValue{
					Value: 100,
					Unit:  2,
				},
				StateOfCharge: 0.55,
				MeterPower: plan.OptimisationValue{
					Value: 400,
					Unit:  2,
				},
			},
		},
	}
	planPath := fmt.Sprintf("/tmp/plan-%d.json", time.Now().Unix())

	err := plan.WritePlan(optPlan, planPath)
	assert.NoError(t, err)

	plan2, err := plan.ReadPlan(planPath)
	assert.NoError(t, err)
	assert.Equal(t, plan2.SiteID, "test-site")
	assert.Equal(t, plan2.OptimisationIntervals[0].BatteryPower.Value, float32(100))
	assert.Equal(t, plan2.OptimisationIntervals[0].StateOfCharge, float32(0.55))
	os.Remove(planPath)
}
