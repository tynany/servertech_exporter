package collector

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	unitsSubsystem = "units"

	unitsLabels       = []string{"id", "name", "type"}
	unitsStatusLabels = append(unitsLabels, "status_type")

	unitsDesc = map[string]*prometheus.Desc{
		"display_orientation": colPromDesc(unitsSubsystem, "display_orientation", "0 = Unknown, 1 = Auto (inverted), 2 = Auto (Normal), 3 = Inverted, 4 = Normal.", unitsLabels),
		"unit_sequence":       colPromDesc(unitsSubsystem, "unit_sequence", "0 = Unknown, 1 = Normal, 2 = Reversed.", unitsLabels),
		"status":              colPromDesc(unitsSubsystem, "status", "Status (1 = Normal, 0 = Not Normal).", unitsStatusLabels),
	}

	totalUnitsErrors = 0.0
)

func init() {
	registerCollector(unitsSubsystem, enabledByDefault, NewUnitsCollector)
}

// UnitsCollector collects units metrics, implemented as per the Collector interface.
type UnitsCollector struct{}

// NewUnitsCollector returns a new UnitsCollector.
func NewUnitsCollector() Collector {
	return &UnitsCollector{}
}

// Get metrics and send to the Prometheus.Metric channel.
func (c *UnitsCollector) Get(ch chan<- prometheus.Metric, target, user, pass string) (float64, error) {

	jsonUnits, err := getServerTechJSON(target, user, pass, "units")
	if err != nil {
		totalUnitsErrors++
		return totalUnitsErrors, fmt.Errorf("cannot get unitss: %s", err)
	}

	if err := processUnitsStats(ch, jsonUnits); err != nil {
		totalUnitsErrors++
		return totalUnitsErrors, err
	}
	return totalUnitsErrors, nil

}

func processUnitsStats(ch chan<- prometheus.Metric, jsonUnitsSum []byte) error {
	var jsonUnits unitsData

	if err := json.Unmarshal(jsonUnitsSum, &jsonUnits); err != nil {
		return fmt.Errorf("cannot unmarshal units json: %s", err)
	}
	for _, data := range jsonUnits {
		labels := []string{data.ID, data.Name, data.Type}

		displayOrientation := float64(0)
		if strings.ToLower(data.DisplayOrientation) == "auto (inverted)" {
			displayOrientation = 1
		} else if strings.ToLower(data.DisplayOrientation) == "auto (normal)" {
			displayOrientation = 2
		} else if strings.ToLower(data.DisplayOrientation) == "inverted" {
			displayOrientation = 3
		} else if strings.ToLower(data.DisplayOrientation) == "normal" {
			displayOrientation = 4
		}
		newGauge(ch, unitsDesc["display_orientation"], displayOrientation, labels...)

		unitSequence := float64(0)
		if strings.ToLower(data.DisplayOrientation) == "normal" {
			unitSequence = 1
		} else if strings.ToLower(data.DisplayOrientation) == "reversed" {
			unitSequence = 2
		}
		newGauge(ch, unitsDesc["unit_sequence"], unitSequence, labels...)

		statusMetric(ch, unitsDesc["status"], data.Status, "unit", labels)

	}
	return nil
}

type unitsData []struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	DisplayOrientation string `json:"display_orientation"`
	OutletSequence     string `json:"unit_sequence"`
	Status             string `json:"status"`
	Type               string `json:"type"`
}
