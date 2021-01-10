package collector

import (
	"encoding/json"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	cordsSubsystem = "cords"

	cordsLabels       = []string{"id", "name", "plug_type"}
	cordsStatusLabels = append(cordsLabels, "status_type")

	cordsDesc = map[string]*prometheus.Desc{
		"watts":                 colPromDesc(cordsSubsystem, "watts", "Integer cord power in Watts. Available only if cord power sensing is present and value is known (AC or DC).", cordsLabels),
		"watts_capacity":        colPromDesc(cordsSubsystem, "watts_capacity", "Integer cord power capacity in Watts.", cordsLabels),
		"voltamps":              colPromDesc(cordsSubsystem, "voltamps", "Integer cord apparent power ranging from 0 to maximum rated power in Volt-Amps. Available only if AC cord power sensing is present and value is known.", cordsLabels),
		"kilowatthours":         colPromDesc(cordsSubsystem, "kilowatthours", "Floating point cord energy in tenth kilowatt-hours (kWh). Available only if energy sensing is present and value is known.", cordsLabels),
		"hertz":                 colPromDesc(cordsSubsystem, "hertz", "Floating point cord frequency in tenth Hertz (Hz). Available only if frequency sensing is present and value is known.", cordsLabels),
		"three_phase_imbalance": colPromDesc(cordsSubsystem, "three_phase_imbalance", "Floating point 3 phase out of balance percentage in tenths.. Available only if 3-phase AC cord current sensing is present and value is known.", cordsLabels),
		"power_factor":          colPromDesc(cordsSubsystem, "power_factor", "Floating point cord power factor in hundredths. Available only if AC cord power factor sensing is present and value is known.", cordsLabels),
		"state":                 colPromDesc(cordsSubsystem, "state", "State (1 = On, 0 = Off)).", cordsLabels),
		"status":                colPromDesc(cordsSubsystem, "status", "Status (1 = Normal, 0 = Not Normal).", cordsStatusLabels),
	}

	totalCordsErrors = 0.0
)

func init() {
	registerCollector(cordsSubsystem, enabledByDefault, NewCordsCollector)
}

// CordsCollector collects cords metrics, implemented as per the Collector interface.
type CordsCollector struct{}

// NewCordsCollector returns a new CordsCollector.
func NewCordsCollector() Collector {
	return &CordsCollector{}
}

// Get metrics and send to the Prometheus.Metric channel.
func (c *CordsCollector) Get(ch chan<- prometheus.Metric, target, user, pass string) (float64, error) {

	jsonCords, err := getServerTechJSON(target, user, pass, "cords")
	if err != nil {
		totalCordsErrors++
		return totalCordsErrors, fmt.Errorf("cannot get cordss: %s", err)
	}

	if err := processCordsStats(ch, jsonCords); err != nil {
		totalCordsErrors++
		return totalCordsErrors, err
	}
	return totalCordsErrors, nil

}

func processCordsStats(ch chan<- prometheus.Metric, jsonCordsSum []byte) error {
	var jsonCords cordsData
	if err := json.Unmarshal(jsonCordsSum, &jsonCords); err != nil {
		return fmt.Errorf("cannot unmarshal cords json: %s", err)
	}
	for _, data := range jsonCords {
		labels := []string{data.ID, data.Name, data.PlugType}

		newGauge(ch, cordsDesc["watts"], data.ActivePower, labels...)
		newGauge(ch, cordsDesc["watts_capacity"], data.PowerCapacity, labels...)
		newGauge(ch, cordsDesc["voltamps"], data.ApparentPower, labels...)
		newGauge(ch, cordsDesc["kilowatthours"], data.Energy, labels...)
		newGauge(ch, cordsDesc["hertz"], data.Frequency, labels...)
		newGauge(ch, cordsDesc["three_phase_imbalance"], data.ThreePhaseImbalance, labels...)
		newGauge(ch, cordsDesc["power_factor"], data.PowerFactor, labels...)

		statusMetric(ch, cordsDesc["status"], data.ActivePowerStatus, "active power", labels)
		statusMetric(ch, cordsDesc["status"], data.ApparentPowerStatus, "apparent power", labels)
		statusMetric(ch, cordsDesc["status"], data.PowerFactorStatus, "power factor", labels)
		statusMetric(ch, cordsDesc["status"], data.ThreePhaseImbalanceStatus, "three phase imbalance", labels)
		statusMetric(ch, cordsDesc["status"], data.Status, "cord", labels)

		stateMetric(ch, cordsDesc["state"], data.State, labels)
	}
	return nil
}

type cordsData []struct {
	ID                        string  `json:"id"`
	Name                      string  `json:"name"`
	ActivePower               float64 `json:"active_power"`
	ActivePowerStatus         string  `json:"active_power_status"`
	ApparentPower             float64 `json:"apparent_power"`
	ApparentPowerStatus       string  `json:"apparent_power_status"`
	Energy                    float64 `json:"energy"`
	Frequency                 float64 `json:"frequency"`
	PowerCapacity             float64 `json:"power_capacity"`
	PowerFactor               float64 `json:"power_factor"`
	PowerFactorStatus         string  `json:"power_factor_status"`
	PowerUtilized             float64 `json:"power_utilized"`
	PlugType                  string  `json:"plug_type"`
	State                     string  `json:"state"`
	Status                    string  `json:"status"`
	ThreePhaseImbalance       float64 `json:"three_phase_imbalance"`
	ThreePhaseImbalanceStatus string  `json:"three_phase_imbalance_status"`
}
