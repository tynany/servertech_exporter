package collector

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	outletsSubsystem = "outlets"

	outletsLabels       = []string{"id", "name", "branch_id", "ocp_id", "phase_id", "socket_adapter", "socket_type"}
	outletsStatusLabels = append(outletsLabels, "status_type")

	outletsDesc = map[string]*prometheus.Desc{
		"watts":          colPromDesc(outletsSubsystem, "watts", "Integer outlet power in Watts. Available only if outlet power sensing is present and value is known (AC or DC).", outletsLabels),
		"watts_capacity": colPromDesc(outletsSubsystem, "watts_capacity", "Integer power capacity in VA for AC products and Watts for DC products.", outletsLabels),
		"voltamps":       colPromDesc(outletsSubsystem, "voltamps", "Integer outlet apparent power in Volt-Amps. Available only if outlet apparent power sensing is present and value is known.", outletsLabels),
		"amps":           colPromDesc(outletsSubsystem, "amps", "Floating point outlet current in hundredth Amps. Available only if outlet current sensing is present and value is known.", outletsLabels),
		"amps_capacity":  colPromDesc(outletsSubsystem, "amps_capacity", "Integer outlet current capacity in whole Amps.", outletsLabels),
		"crest_factor":   colPromDesc(outletsSubsystem, "crest_factor", "Floating point outlet crest factor in tenths. Available only if outlet crest factor sensing is present and value is known.", outletsLabels),
		"kilowatthours":  colPromDesc(outletsSubsystem, "kilowatthours", "Floating point outlet energy in tenth kilowatt-hours (kWh). Available only if energy sensing is present and value is known.", outletsLabels),
		"power_factor":   colPromDesc(outletsSubsystem, "power_factor", "Floating point outlet power factor in hundredths. Available only if AC cord power factor sensing is present and value is known.", outletsLabels),
		"reactance":      colPromDesc(outletsSubsystem, "reactance", "Status of the measured outlet reactance. Available only if outletpower factor sensing present and value is known (0 = Unknown, 1 = Capacitive, 2 = Inductive, 3 = Resistive.", outletsLabels),
		"volts":          colPromDesc(outletsSubsystem, "volts", "Floating point outlet voltage in tenth Volts. Available only if voltage sensing is present and value is known.", outletsLabels),
		"state":          colPromDesc(outletsSubsystem, "state", "State (1 = On, 0 = Off)).", outletsLabels),
		"status":         colPromDesc(outletsSubsystem, "status", "Status (1 = Normal, 0 = Not Normal).", outletsStatusLabels),
	}

	totalOutletsErrors = 0.0
)

func init() {
	registerCollector(outletsSubsystem, enabledByDefault, NewOutletsCollector)
}

// OutletsCollector collects outlets metrics, implemented as per the Collector interface.
type OutletsCollector struct{}

// NewOutletsCollector returns a new OutletsCollector.
func NewOutletsCollector() Collector {
	return &OutletsCollector{}
}

// Get metrics and send to the Prometheus.Metric channel.
func (c *OutletsCollector) Get(ch chan<- prometheus.Metric, target, user, pass string) (float64, error) {

	jsonOutlets, err := getServerTechJSON(target, user, pass, "outlets")
	if err != nil {
		totalOutletsErrors++
		return totalOutletsErrors, fmt.Errorf("cannot get outletss: %s", err)
	}

	if err := processOutletsStats(ch, jsonOutlets); err != nil {
		totalOutletsErrors++
		return totalOutletsErrors, err
	}
	return totalOutletsErrors, nil

}

func processOutletsStats(ch chan<- prometheus.Metric, jsonOutletsSum []byte) error {
	var jsonOutlets outletsData
	if err := json.Unmarshal(jsonOutletsSum, &jsonOutlets); err != nil {
		return fmt.Errorf("cannot unmarshal outlets json: %s", err)
	}
	for _, data := range jsonOutlets {
		labels := []string{data.ID, data.Name, data.BranchID, data.OcpID, data.PhaseID, data.SocketAdapter, data.SocketType}

		newGauge(ch, outletsDesc["watts"], data.ActivePower, labels...)
		newGauge(ch, outletsDesc["watts_capacity"], data.PowerCapacity, labels...)
		newGauge(ch, outletsDesc["voltamps"], data.ApparentPower, labels...)
		newGauge(ch, outletsDesc["amps"], data.Current, labels...)
		newGauge(ch, outletsDesc["amps_capacity"], data.CurrentCapacity, labels...)
		newGauge(ch, outletsDesc["crest_factor"], data.CrestFactor, labels...)
		newGauge(ch, outletsDesc["kilowatthours"], data.Energy, labels...)
		newGauge(ch, outletsDesc["power_factor"], data.PowerFactor, labels...)
		newGauge(ch, outletsDesc["volts"], data.Voltage, labels...)

		reactance := float64(0)
		if strings.ToLower(data.Reactance) == "capacitive" {
			reactance = 1
		} else if strings.ToLower(data.Reactance) == "inductive" {
			reactance = 2
		} else if strings.ToLower(data.Reactance) == "resistive" {
			reactance = 3
		}
		newGauge(ch, outletsDesc["reactance"], reactance, labels...)

		statusMetric(ch, outletsDesc["status"], data.ActivePowerStatus, "active power", labels)
		statusMetric(ch, outletsDesc["status"], data.CurrentStatus, "current", labels)
		statusMetric(ch, outletsDesc["status"], data.PowerFactorStatus, "power factor", labels)
		statusMetric(ch, outletsDesc["status"], data.Status, "outlet", labels)

		stateMetric(ch, outletsDesc["state"], data.State, labels)
	}
	return nil
}

type outletsData []struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	ActivePower       float64 `json:"active_power"`
	ActivePowerStatus string  `json:"active_power_status"`
	ApparentPower     float64 `json:"apparent_power"`
	BranchID          string  `json:"branch_id"`
	ControlState      string  `json:"control_state"`
	Current           float64 `json:"current"`
	CurrentCapacity   float64 `json:"current_capacity"`
	CurrentStatus     string  `json:"current_status"`
	CurrentUtilized   float64 `json:"current_utilized"`
	Energy            float64 `json:"energy"`
	OcpID             string  `json:"ocp_id"`
	PhaseID           string  `json:"phase_id"`
	PowerCapacity     float64 `json:"power_capacity"`
	PowerFactorStatus string  `json:"power_factor_status"`
	SocketAdapter     string  `json:"socket_adapter"`
	SocketType        string  `json:"socket_type"`
	State             string  `json:"state"`
	Status            string  `json:"status"`
	Voltage           float64 `json:"voltage"`
	CrestFactor       float64 `json:"crest_factor,omitempty"`
	PowerFactor       float64 `json:"power_factor,omitempty"`
	Reactance         string  `json:"reactance,omitempty"`
}
