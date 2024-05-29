package plan

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"
)

type Handler struct {
	logger *slog.Logger
	mu     *sync.RWMutex
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

func (o OptimisationPlan) IsEmpty() bool {
	return o.SiteID == "" && len(o.OptimisationIntervals) == 0 && o.OptimisationTimestamp.Seconds == 0
}

func (i OptimisationInterval) IsEmpty() bool {
	return i.Interval.StartTime.Seconds == 0
}

func (i OptimisationInterval) IsCurrent(targetTime time.Time) bool {
	intStart := time.Unix(i.Interval.StartTime.Seconds, 0)
	intEnd := time.Unix(i.Interval.EndTime.Seconds, 0)

	isAfterStart := targetTime.Equal(intStart) || targetTime.After(intStart)
	isBeforeEnd := targetTime.Before(intEnd)
	return isAfterStart && isBeforeEnd
}

func NewHandler(logger *slog.Logger, path string) Handler {
	return Handler{logger: logger, mu: new(sync.RWMutex), Path: path}
}

func (p Handler) ReadPlan() (OptimisationPlan, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

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

func (p Handler) WritePlan(optPlan OptimisationPlan) error {
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

func (p Handler) GetCurrentInterval(targetTime time.Time) (OptimisationInterval, error) {
	plan, err := p.ReadPlan()
	if err != nil {
		return OptimisationInterval{}, fmt.Errorf("reading current plan: %w", err)
	}
	for _, intv := range plan.OptimisationIntervals {
		if intv.IsCurrent(targetTime) {
			return intv, nil
		}
	}
	return OptimisationInterval{}, fmt.Errorf("no current interval found in plan")
}
