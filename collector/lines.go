package collector

import (
	"encoding/json"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	linesSubsystem = "lines"

	linesLabels       = []string{"id", "name"}
	linesStatusLabels = append(linesLabels, "status_type")

	linesDesc = map[string]*prometheus.Desc{
		"amps":          colPromDesc(linesSubsystem, "amps", "Floating point branch current in hundredth Amps. Available only if branch current sensing is present and value is known.", linesLabels),
		"amps_capacity": colPromDesc(linesSubsystem, "amps_capacity", "Integer branch current capacity in whole Amps.", linesLabels),
		"state":         colPromDesc(linesSubsystem, "state", "State (1 = On, 0 = Off)).", linesLabels),
		"status":        colPromDesc(linesSubsystem, "status", "Status (1 = Normal, 0 = Not Normal).", linesStatusLabels),
	}

	totalLinesErrors = 0.0
)

func init() {
	registerCollector(linesSubsystem, enabledByDefault, NewLinesCollector)
}

// LinesCollector collects lines metrics, implemented as per the Collector interface.
type LinesCollector struct{}

// NewLinesCollector returns a new LinesCollector.
func NewLinesCollector() Collector {
	return &LinesCollector{}
}

// Get metrics and send to the Prometheus.Metric channel.
func (c *LinesCollector) Get(ch chan<- prometheus.Metric, target, user, pass string) (float64, error) {

	jsonLines, err := getServerTechJSON(target, user, pass, "lines")
	if err != nil {
		totalLinesErrors++
		return totalLinesErrors, fmt.Errorf("cannot get liness: %s", err)
	}

	if err := processLinesStats(ch, jsonLines); err != nil {
		totalLinesErrors++
		return totalLinesErrors, err
	}
	return totalLinesErrors, nil

}

func processLinesStats(ch chan<- prometheus.Metric, jsonLinesSum []byte) error {
	var jsonLiness linesData

	if err := json.Unmarshal(jsonLinesSum, &jsonLiness); err != nil {
		return fmt.Errorf("cannot unmarshal lines json: %s", err)
	}
	for _, data := range jsonLiness {
		labels := []string{data.ID, data.Name}

		newGauge(ch, linesDesc["amps"], data.Current, labels...)
		newGauge(ch, linesDesc["amps_capacity"], data.CurrentCapacity, labels...)

		statusMetric(ch, linesDesc["status"], data.CurrentStatus, "current", labels)
		statusMetric(ch, linesDesc["status"], data.Status, "line", labels)

		stateMetric(ch, linesDesc["state"], data.State, labels)
	}
	return nil
}

type linesData []struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Current         float64 `json:"current"`
	CurrentCapacity float64 `json:"current_capacity"`
	CurrentStatus   string  `json:"current_status"`
	CurrentUtilized float64 `json:"current_utilized"`
	State           string  `json:"state"`
	Status          string  `json:"status"`
}
