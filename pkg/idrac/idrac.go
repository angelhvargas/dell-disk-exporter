package idrac

import (
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// CommandExecutor defines an interface for executing commands
type CommandExecutor interface {
	ExecuteCommand(name string, args ...string) ([]byte, error)
}

// DefaultCommandExecutor implements CommandExecutor
type DefaultCommandExecutor struct{}

func (e *DefaultCommandExecutor) ExecuteCommand(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.CombinedOutput()
}

type Client struct {
	executor       CommandExecutor
	raidStatus     *prometheus.GaugeVec
	raidRedundancy *prometheus.GaugeVec
	raidSize       *prometheus.GaugeVec
	raidLayout     *prometheus.GaugeVec
}

func NewClient(executor CommandExecutor, registry *prometheus.Registry) *Client {
	raidStatus := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "raid_status",
			Help: "Status of the RAID controller",
		},
		[]string{"vdisk"},
	)
	raidRedundancy := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "raid_redundancy",
			Help: "Remaining redundancy of the RAID controller",
		},
		[]string{"vdisk"},
	)
	raidSize := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "raid_size",
			Help: "Size of the RAID controller",
		},
		[]string{"vdisk"},
	)
	raidLayout := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "raid_layout",
			Help: "Layout of the RAID controller",
		},
		[]string{"vdisk"},
	)

	registry.MustRegister(raidStatus)
	registry.MustRegister(raidRedundancy)
	registry.MustRegister(raidSize)
	registry.MustRegister(raidLayout)

	return &Client{
		executor:       executor,
		raidStatus:     raidStatus,
		raidRedundancy: raidRedundancy,
		raidSize:       raidSize,
		raidLayout:     raidLayout,
	}
}

func (c *Client) GetRAIDStatus() (map[string]map[string]string, error) {
	output, err := c.executor.ExecuteCommand("racadm", "raid", "get", "vdisks", "-o", "-p", "layout,status,RemainingRedundancy,Size")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	raidStatuses := make(map[string]map[string]string)
	var currentVdisk string

	for _, line := range lines {
		if strings.HasPrefix(line, "Disk.Virtual") {
			currentVdisk = strings.TrimSpace(strings.Split(line, ":")[1])
			raidStatuses[currentVdisk] = make(map[string]string)
		} else if currentVdisk != "" && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			raidStatuses[currentVdisk][key] = value
		}
	}

	return raidStatuses, nil
}

func (c *Client) UpdateMetrics() {
	for {
		statuses, err := c.GetRAIDStatus()
		if err != nil {
			log.Printf("Error fetching RAID status: %v", err)
			return
		}
		for vdisk, metrics := range statuses {
			log.Printf("RAID Status for %s: %v", vdisk, metrics)
			c.raidStatus.WithLabelValues(vdisk).Set(1) // Assuming Status is Ok
			c.raidRedundancy.WithLabelValues(vdisk).Set(parseToFloat(metrics["RemainingRedundancy"]))
			c.raidSize.WithLabelValues(vdisk).Set(parseToFloat(metrics["Size"]))
			c.raidLayout.WithLabelValues(vdisk).Set(1) // Assuming Layout is set
		}
		time.Sleep(30 * time.Second) // Adjust the interval as needed
	}
}

func parseToFloat(value string) float64 {
	parsed, _ := strconv.ParseFloat(strings.Fields(value)[0], 64)
	return parsed
}
