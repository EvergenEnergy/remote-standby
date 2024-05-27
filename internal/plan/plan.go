package plan

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"
)

type PlanHandler struct {
	logger *slog.Logger
	mu     *sync.Mutex
	Path   string `required:"True"`
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

func (o OptimisationPlan) IsEmpty(logger *slog.Logger) bool {
	return o.SiteID == "" && len(o.OptimisationIntervals) == 0 && o.OptimisationTimestamp.Seconds == 0
}

func NewHandler(logger *slog.Logger, path string) PlanHandler {
	return PlanHandler{logger: logger, mu: new(sync.Mutex), Path: path}
}

func (p PlanHandler) ReadPlan() (OptimisationPlan, error) {
	content, err := os.ReadFile(p.Path)
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

func (p PlanHandler) WritePlan(optPlan OptimisationPlan) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	f, err := os.Create(p.Path)
	if err != nil {
		return fmt.Errorf("creating plan backup file at %s: %w", p.Path, err)
	}
	defer f.Close()

	encodedPlan, err := json.Marshal(optPlan)
	if err != nil {
		return fmt.Errorf("marshalling optimisation plan: %w", err)
	}

	_, err = f.Write(encodedPlan)
	if err != nil {
		return fmt.Errorf("writing plan to file at %s: %w", p.Path, err)
	}

	return nil
}

func (p PlanHandler) TrimPlan(targetTime time.Time) error {
	plan, err := p.ReadPlan()
	if err != nil {
		return fmt.Errorf("reading current plan: %w", err)
	}

	newIntervals := []OptimisationInterval{}
	for i, intv := range plan.OptimisationIntervals {
		if time.Unix(intv.Interval.StartTime.Seconds, 0).Compare(targetTime) >= 0 {
			newIntervals = append(newIntervals, plan.OptimisationIntervals[i:]...)
			break
		}
	}
	plan.OptimisationIntervals = newIntervals
	err = p.WritePlan(plan)
	if err != nil {
		return fmt.Errorf("writing plan: %w", err)
	}

	return nil
}

func (p PlanHandler) GetCurrentInterval(targetTime time.Time) (OptimisationInterval, error) {
	plan, err := p.ReadPlan()
	if err != nil {
		return OptimisationInterval{}, fmt.Errorf("reading current plan: %w", err)
	}
	for _, intv := range plan.OptimisationIntervals {
		if time.Unix(intv.Interval.StartTime.Seconds, 0).Compare(targetTime) >= 0 {
			return intv, nil
		}
	}
	return OptimisationInterval{}, fmt.Errorf("no current interval found in plan")
}
