package collector

import (
	"encoding/json"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	ocpsSubsystem = "ocps"

	ocpsLabels       = []string{"id", "name", "type"}
	ocpsStatusLabels = append(ocpsLabels, "status_type")

	ocpsDesc = map[string]*prometheus.Desc{
		"amps":          colPromDesc(ocpsSubsystem, "amps", "Floating point branch current in hundredth Amps. Available only if branch current sensing is present and value is known.", ocpsLabels),
		"amps_capacity": colPromDesc(ocpsSubsystem, "amps_capacity", "Integer branch current capacity in whole Amps.", ocpsLabels),
		"state":         colPromDesc(ocpsSubsystem, "state", "State (1 = On, 0 = Off)).", ocpsLabels),
		"status":        colPromDesc(ocpsSubsystem, "status", "Status (1 = Normal, 0 = Not Normal).", ocpsStatusLabels),
	}

	totalOcpsErrors = 0.0
)

func init() {
	registerCollector(ocpsSubsystem, enabledByDefault, NewOcpsCollector)
}

// OcpsCollector collects ocps metrics, implemented as per the Collector interface.
type OcpsCollector struct{}

// NewOcpsCollector returns a new OcpsCollector.
func NewOcpsCollector() Collector {
	return &OcpsCollector{}
}

// Get metrics and send to the Prometheus.Metric channel.
func (c *OcpsCollector) Get(ch chan<- prometheus.Metric, target, user, pass string) (float64, error) {

	jsonOcps, err := getServerTechJSON(target, user, pass, "ocps")
	if err != nil {
		totalOcpsErrors++
		return totalOcpsErrors, fmt.Errorf("cannot get ocpss: %s", err)
	}

	if err := processOcpsStats(ch, jsonOcps); err != nil {
		totalOcpsErrors++
		return totalOcpsErrors, err
	}
	return totalOcpsErrors, nil

}

func processOcpsStats(ch chan<- prometheus.Metric, jsonOcpsSum []byte) error {
	var jsonOcpss ocpsData

	if err := json.Unmarshal(jsonOcpsSum, &jsonOcpss); err != nil {
		return fmt.Errorf("cannot unmarshal ocps json: %s", err)
	}
	for _, data := range jsonOcpss {
		labels := []string{data.ID, data.Name, data.Type}

		newGauge(ch, ocpsDesc["amps_capacity"], data.CurrentCapacity, labels...)

		statusMetric(ch, ocpsDesc["status"], data.Status, "ocp", labels)

	}
	return nil
}

type ocpsData []struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	CurrentCapacity float64 `json:"current_capacity"`
	Status          string  `json:"status"`
	Type            string  `json:"type"`
}
