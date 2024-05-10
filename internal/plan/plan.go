package plan

import (
	"encoding/json"
	"fmt"
	"os"
)

type PlanHandler struct {
	OptimisationPlan OptimisationPlan
}

type OptimisationPlan struct {
	SiteID                string                 `json:"site_id"`
	OptimisationTimestamp OptimisationTimestamp  `json:"optimisation_timestamp"`
	OptimisationIntervals []OptimisationInterval `json:"optimisation_intervals"`
	SetpointType          int                    `json:"setpoint_type"`
}

type OptimisationTimestamp struct {
	Seconds int64 `json:"seconds"`
	Nanos   int64 `json:"nanos"`
}

type OptimisationInterval struct {
	Interval      OptimisationIntervalTimestamp `json:"optimisation_interval"`
	BatteryPower  OptimisationValue             `json:"battery_power"`
	StateOfCharge float32                       `json:"state_of_charge"`
	MeterPower    OptimisationValue             `json:"meter_power"`
}

type OptimisationIntervalTimestamp struct {
	StartTime OptimisationTimestamp `json:"start_time"`
	EndTime   OptimisationTimestamp `json:"end_time"`
}

type OptimisationValue struct {
	Value float32 `json:"value"`
	Unit  int     `json:"unit"`
}

func ReadPlan(path string) (OptimisationPlan, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return OptimisationPlan{}, fmt.Errorf("reading plan from file: %w", err)
	}
	optPlan := OptimisationPlan{}
	err = json.Unmarshal(content, &optPlan)
	if err != nil {
		fmt.Println(err)
		return OptimisationPlan{}, fmt.Errorf("unmarshalling plan: %w", err)
	}
	return optPlan, nil
}

func WritePlan(optPlan OptimisationPlan, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating plan backup file at %s: %w", path, err)
	}

	encodedPlan, err := json.Marshal(optPlan)
	if err != nil {
		return fmt.Errorf("marshalling optimisation plan: %w", err)
	}

	_, err = f.Write(encodedPlan)
	if err != nil {
		return fmt.Errorf("writing plan to file at %s: %w", path, err)
	}

	return nil
}
