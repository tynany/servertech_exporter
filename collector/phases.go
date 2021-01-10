package collector

import (
	"encoding/json"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	phasesSubsystem = "phases"

	phasesLabels       = []string{"id", "name"}
	phasesStatusLabels = append(phasesLabels, "status_type")

	phasesDesc = map[string]*prometheus.Desc{
		"watts":           colPromDesc(phasesSubsystem, "watts", "Integer phase power in Watts. Available only if phase power sensing is present and value is known (AC or DC).", phasesLabels),
		"voltamps":        colPromDesc(phasesSubsystem, "voltamps", "Integer phase apparent power in Volt-Amps. Available only if phase apparent power sensing is present and value is known.", phasesLabels),
		"amps":            colPromDesc(phasesSubsystem, "amps", "Floating point phase current in hundredth Amps. Available only if phase current sensing is present and value is known.", phasesLabels),
		"crest_factor":    colPromDesc(phasesSubsystem, "crest_factor", "Floating point phase crest factor in tenths. Available only if phase crest factor sensing is present and value is known.", phasesLabels),
		"kilowatthours":   colPromDesc(phasesSubsystem, "kilowatthours", "Floating point phase energy in tenth kilowatt-hours (kWh). Available only if energy sensing is present and value is known.", phasesLabels),
		"nominal_volts":   colPromDesc(phasesSubsystem, "nominal_volts", "Integer phase nominal voltage in Volts. Available only if phase voltage sensing present.", phasesLabels),
		"power_factor":    colPromDesc(phasesSubsystem, "power_factor", "Floating point phase power factor in hundredths. Available only if AC cord power factor sensing is present and value is known.", phasesLabels),
		"reactance":       colPromDesc(phasesSubsystem, "reactance", "Status of the measured phase reactance. Available only if phasepower factor sensing present and value is known (0 = Unknown, 1 = Capacitive, 2 = Inductive, 3 = Resistive.", phasesLabels),
		"volts":           colPromDesc(phasesSubsystem, "volts", "Floating point phase voltage in tenth Volts. Available only if voltage sensing is present and value is known. ", phasesLabels),
		"volts_deviation": colPromDesc(phasesSubsystem, "volts_deviation", "Floating point phase deviation percentage from nominal voltage in tenths. Available only if phase voltage sensing present.", phasesLabels),
		"state":           colPromDesc(phasesSubsystem, "state", "State (1 = On, 0 = Off)).", phasesLabels),
		"status":          colPromDesc(phasesSubsystem, "status", "Status (1 = Normal, 0 = Not Normal).", phasesStatusLabels),
	}

	totalPhasesErrors = 0.0
)

func init() {
	registerCollector(phasesSubsystem, enabledByDefault, NewPhasesCollector)
}

// PhasesCollector collects phases metrics, implemented as per the Collector interface.
type PhasesCollector struct{}

// NewPhasesCollector returns a new PhasesCollector.
func NewPhasesCollector() Collector {
	return &PhasesCollector{}
}

// Get metrics and send to the Prometheus.Metric channel.
func (c *PhasesCollector) Get(ch chan<- prometheus.Metric, target, user, pass string) (float64, error) {

	jsonPhases, err := getServerTechJSON(target, user, pass, "phases")
	if err != nil {
		totalPhasesErrors++
		return totalPhasesErrors, fmt.Errorf("cannot get phasess: %s", err)
	}

	if err := processPhasesStats(ch, jsonPhases); err != nil {
		totalPhasesErrors++
		return totalPhasesErrors, err
	}
	return totalPhasesErrors, nil

}

func processPhasesStats(ch chan<- prometheus.Metric, jsonPhasesSum []byte) error {
	var jsonPhases phasesData
	if err := json.Unmarshal(jsonPhasesSum, &jsonPhases); err != nil {
		return fmt.Errorf("cannot unmarshal phases json: %s", err)
	}
	for _, data := range jsonPhases {
		labels := []string{data.ID, data.Name}

		newGauge(ch, phasesDesc["watts"], data.ActivePower, labels...)
		newGauge(ch, phasesDesc["voltamps"], data.ApparentPower, labels...)
		newGauge(ch, phasesDesc["amps"], data.Current, labels...)
		newGauge(ch, phasesDesc["crest_factor"], data.CrestFactor, labels...)
		newGauge(ch, phasesDesc["kilowatthours"], data.Energy, labels...)
		newGauge(ch, phasesDesc["nominal_volts"], data.NominalVoltage, labels...)
		newGauge(ch, phasesDesc["power_factor"], data.PowerFactor, labels...)
		newGauge(ch, phasesDesc["volts"], data.Voltage, labels...)
		newGauge(ch, phasesDesc["volts_deviation"], data.VoltageDeviation, labels...)

		reactanceMetric(ch, phasesDesc["reactance"], data.Reactance, labels)

		statusMetric(ch, phasesDesc["status"], data.PowerFactorStatus, "power factor", labels)
		statusMetric(ch, phasesDesc["status"], data.VoltageStatus, "voltage", labels)
		statusMetric(ch, phasesDesc["status"], data.Status, "phase", labels)

		stateMetric(ch, phasesDesc["state"], data.State, labels)
	}
	return nil
}

type phasesData []struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	ActivePower       float64 `json:"active_power"`
	ApparentPower     float64 `json:"apparent_power"`
	CrestFactor       float64 `json:"crest_factor"`
	Current           float64 `json:"current"`
	Energy            float64 `json:"energy"`
	NominalVoltage    float64 `json:"nominal_voltage"`
	PowerFactor       float64 `json:"power_factor"`
	PowerFactorStatus string  `json:"power_factor_status"`
	Reactance         string  `json:"reactance"`
	State             string  `json:"state"`
	Status            string  `json:"status"`
	Voltage           float64 `json:"voltage"`
	VoltageStatus     string  `json:"voltage_status"`
	VoltageDeviation  float64 `json:"voltage_deviation"`
}
