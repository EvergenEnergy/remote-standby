package plan_test

import (
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/EvergenEnergy/remote-standby/internal/plan"
	"github.com/stretchr/testify/assert"
)

var testLogger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

func GetOptimisationPlan() plan.OptimisationPlan {
	return plan.OptimisationPlan{
		SiteID:       "test-site",
		SetpointType: 1,
		OptimisationIntervals: []plan.OptimisationInterval{
			{
				Interval: plan.OptimisationIntervalTimestamp{
					StartTime: plan.OptimisationTimestamp{Seconds: 1715319000},
					EndTime:   plan.OptimisationTimestamp{Seconds: 1715319300},
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
			{
				Interval: plan.OptimisationIntervalTimestamp{
					StartTime: plan.OptimisationTimestamp{Seconds: 1715319300},
					EndTime:   plan.OptimisationTimestamp{Seconds: 1715319600},
				},
				BatteryPower: plan.OptimisationValue{
					Value: 110,
					Unit:  2,
				},
				StateOfCharge: 0.55,
				MeterPower: plan.OptimisationValue{
					Value: 390,
					Unit:  2,
				},
			},
			{
				Interval: plan.OptimisationIntervalTimestamp{
					StartTime: plan.OptimisationTimestamp{Seconds: 1715319600},
					EndTime:   plan.OptimisationTimestamp{Seconds: 1715319900},
				},
				BatteryPower: plan.OptimisationValue{
					Value: 120,
					Unit:  2,
				},
				StateOfCharge: 0.55,
				MeterPower: plan.OptimisationValue{
					Value: 380,
					Unit:  2,
				},
			},
		},
	}
}

func TestWritesAndReadsAPlan(t *testing.T) {
	planPath := fmt.Sprintf("/tmp/write-plan-%d.json", time.Now().Unix())

	handler := plan.NewHandler(testLogger, planPath)

	err := handler.WritePlan(GetOptimisationPlan())
	assert.NoError(t, err)

	plan2, err := handler.ReadPlan()
	assert.NoError(t, err)
	assert.Equal(t, plan2.SiteID, "test-site")
	assert.Equal(t, plan2.OptimisationIntervals[0].BatteryPower.Value, float32(100))
	assert.Equal(t, plan2.OptimisationIntervals[0].StateOfCharge, float32(0.55))
	os.Remove(planPath)
}

func TestRemoveExpiredIntervals(t *testing.T) {
	type test struct {
		startTime   int
		expectedNum int
	}

	tests := []test{
		{startTime: 1715318999, expectedNum: 3},
		{startTime: 1715319300, expectedNum: 2},
		{startTime: 1715319901, expectedNum: 0},
	}

	for i, tc := range tests {

		planPath := fmt.Sprintf("/tmp/trim-plan-%d-%d.json", i, time.Now().Unix())

		handler := plan.NewHandler(testLogger, planPath)

		origPlan := GetOptimisationPlan()
		assert.Len(t, origPlan.OptimisationIntervals, 3)

		err := handler.WritePlan(origPlan)
		assert.NoError(t, err)

		startTime := time.Unix(int64(tc.startTime), 0)

		err = handler.TrimPlan(startTime)
		assert.NoError(t, err)

		trimmedPlan, err := handler.ReadPlan()
		assert.NoError(t, err)
		assert.Len(t, trimmedPlan.OptimisationIntervals, tc.expectedNum)

		os.Remove(planPath)
	}
}

func TestGetCurrentInterval(t *testing.T) {
	type test struct {
		startTime          int
		hasInterval        bool
		expectedMeterPower int
	}

	tests := []test{
		{startTime: 1715318999, hasInterval: true, expectedMeterPower: 400},
		{startTime: 1715319300, hasInterval: true, expectedMeterPower: 390},
		{startTime: 1715319901, hasInterval: false, expectedMeterPower: 0},
	}

	for i, tc := range tests {

		planPath := fmt.Sprintf("/tmp/interval-plan-%d-%d.json", i, time.Now().Unix())

		handler := plan.NewHandler(testLogger, planPath)

		origPlan := GetOptimisationPlan()

		err := handler.WritePlan(origPlan)
		assert.NoError(t, err)

		startTime := time.Unix(int64(tc.startTime), 0)

		optInterval, err := handler.GetCurrentInterval(startTime)
		assert.NotEqual(t, tc.hasInterval, optInterval.IsEmpty())
		if tc.hasInterval {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
		assert.EqualValues(t, optInterval.MeterPower.Value, tc.expectedMeterPower)

		os.Remove(planPath)
	}
}
