package collector

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	systemSubsystem = "system"

	systemLabels       = []string{"firmware", "nic_serial_number"}
	systemStatusLabels = append(systemLabels, "status_type")

	systemDesc = map[string]*prometheus.Desc{
		"active_users":   colPromDesc(systemSubsystem, "active_users", "Integer number of active users logged in.", systemLabels),
		"uptime_seconds": colPromDesc(systemSubsystem, "voltamps", "System uptime", systemLabels),
		"status":         colPromDesc(systemSubsystem, "status", "Status (1 = Normal, 0 = Not Normal).", systemStatusLabels),
	}

	totalSystemErrors = 0.0
)

func init() {
	registerCollector(systemSubsystem, enabledByDefault, NewSystemCollector)
}

// SystemCollector collects system metrics, implemented as per the Collector interface.
type SystemCollector struct{}

// NewSystemCollector returns a new SystemCollector.
func NewSystemCollector() Collector {
	return &SystemCollector{}
}

// Get metrics and send to the Prometheus.Metric channel.
func (c *SystemCollector) Get(ch chan<- prometheus.Metric, target, user, pass string) (float64, error) {

	jsonSystem, err := getServerTechJSON(target, user, pass, "system")
	if err != nil {
		totalSystemErrors++
		return totalSystemErrors, fmt.Errorf("cannot get systems: %s", err)
	}

	if err := processSystemStats(ch, jsonSystem); err != nil {
		totalSystemErrors++
		return totalSystemErrors, err
	}
	return totalSystemErrors, nil

}

func processSystemStats(ch chan<- prometheus.Metric, jsonSystemSum []byte) error {
	var data systemData
	if err := json.Unmarshal(jsonSystemSum, &data); err != nil {
		return fmt.Errorf("cannot unmarshal system json: %s", err)
	}
	labels := []string{data.Firmware, data.NicSerialNumber}

	newGauge(ch, systemDesc["active_users"], data.ActiveUsers, labels...)

	statusMetric(ch, systemDesc["status"], data.StatusBranches, "branches", labels)
	statusMetric(ch, systemDesc["status"], data.StatusCords, "cords", labels)
	statusMetric(ch, systemDesc["status"], data.StatusLines, "lines", labels)
	statusMetric(ch, systemDesc["status"], data.StatusOcps, "ocps", labels)
	statusMetric(ch, systemDesc["status"], data.StatusOutlets, "outlets", labels)
	statusMetric(ch, systemDesc["status"], data.StatusPhases, "phases", labels)
	statusMetric(ch, systemDesc["status"], data.StatusUnits, "units", labels)

	r, err := regexp.Compile("(?:(.*) days )?(?:(.*) hours )?(?:(.*) minutes )?(.*) seconds")
	if err != nil {
		return fmt.Errorf("could not compile uptime regex: %v", err)
	}
	reUptime := r.FindStringSubmatch(data.Uptime)
	uptimeDays, err := strconv.Atoi(reUptime[1])
	if err != nil {
		return fmt.Errorf("could not convert uptime day to int: %v", err)
	}
	uptimeHours, err := strconv.Atoi(reUptime[2])
	if err != nil {
		return fmt.Errorf("could not convert uptime hour to int: %v", err)
	}
	uptimeMinutes, err := strconv.Atoi(reUptime[3])
	if err != nil {
		return fmt.Errorf("could not convert uptime minute to int: %v", err)
	}
	uptimeSeconds, err := strconv.Atoi(reUptime[4])
	if err != nil {
		return fmt.Errorf("could not convert uptime second to int: %v", err)
	}
	uptime := (uptimeDays * 86400) + (uptimeHours * 3600) + (uptimeMinutes * 60) + uptimeSeconds

	newCounter(ch, systemDesc["uptime_seconds"], float64(uptime), labels...)

	return nil
}

type systemData struct {
	ActiveUsers     float64 `json:"active_users"`
	Firmware        string  `json:"firmware"`
	NicSerialNumber string  `json:"nic_serial_number"`
	StatusBranches  string  `json:"status_branches"`
	StatusCords     string  `json:"status_cords"`
	StatusLines     string  `json:"status_lines"`
	StatusOcps      string  `json:"status_ocps"`
	StatusOutlets   string  `json:"status_outlets"`
	StatusPhases    string  `json:"status_phases"`
	StatusUnits     string  `json:"status_units"`
	Uptime          string  `json:"uptime"`
}
