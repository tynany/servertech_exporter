package collector

import (
	"encoding/json"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	branchesSubsystem = "branches"

	branchesLabels       = []string{"id", "name", "phase_id", "ocp_id"}
	branchesStatusLabels = append(branchesLabels, "status_type")

	branchesDesc = map[string]*prometheus.Desc{
		"amps":          colPromDesc(branchesSubsystem, "amps", "Floating point branch current in hundredth Amps. Available only if branch current sensing is present and value is known.", branchesLabels),
		"amps_capacity": colPromDesc(branchesSubsystem, "amps_capacity", "Integer branch current capacity in whole Amps.", branchesLabels),
		"state":         colPromDesc(branchesSubsystem, "state", "State (1 = On, 0 = Off)).", branchesLabels),
		"status":        colPromDesc(branchesSubsystem, "status", "Status (1 = Normal, 0 = Not Normal).", branchesStatusLabels),
	}

	totalBranchesErrors = 0.0
)

func init() {
	registerCollector(branchesSubsystem, enabledByDefault, NewBranchesCollector)
}

// BranchesCollector collects branches metrics, implemented as per the Collector interface.
type BranchesCollector struct{}

// NewBranchesCollector returns a new BranchesCollector.
func NewBranchesCollector() Collector {
	return &BranchesCollector{}
}

// Get metrics and send to the Prometheus.Metric channel.
func (c *BranchesCollector) Get(ch chan<- prometheus.Metric, target, user, pass string) (float64, error) {

	jsonBranches, err := getServerTechJSON(target, user, pass, "branches")
	if err != nil {
		totalBranchesErrors++
		return totalBranchesErrors, fmt.Errorf("cannot get branchess: %s", err)
	}

	if err := processBranchesStats(ch, jsonBranches); err != nil {
		totalBranchesErrors++
		return totalBranchesErrors, err
	}
	return totalBranchesErrors, nil

}

func processBranchesStats(ch chan<- prometheus.Metric, jsonBranchesSum []byte) error {
	var jsonBranchess branchesData

	if err := json.Unmarshal(jsonBranchesSum, &jsonBranchess); err != nil {
		return fmt.Errorf("cannot unmarshal branches json: %s", err)
	}
	for _, data := range jsonBranchess {
		labels := []string{data.ID, data.Name, data.OcpID, data.PhaseID}

		newGauge(ch, branchesDesc["amps"], data.Current, labels...)
		newGauge(ch, branchesDesc["amps_capacity"], data.CurrentCapacity, labels...)

		statusMetric(ch, branchesDesc["status"], data.CurrentStatus, "current", labels)
		statusMetric(ch, branchesDesc["status"], data.Status, "branche", labels)

		stateMetric(ch, branchesDesc["state"], data.State, labels)
	}
	return nil
}

type branchesData []struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Current         float64 `json:"current"`
	CurrentCapacity float64 `json:"current_capacity"`
	CurrentStatus   string  `json:"current_status"`
	CurrentUtilized float64 `json:"current_utilized"`
	OcpID           string  `json:"ocp_id"`
	PhaseID         string  `json:"phase_id"`
	State           string  `json:"state"`
	Status          string  `json:"status"`
}
