package idrac

import (
	"log"
	"os/exec"
	"strings"
	"time"
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
	executor CommandExecutor
}

func NewClient(executor CommandExecutor) *Client {
	return &Client{executor: executor}
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
			// Update Prometheus metrics here
		}
		time.Sleep(30 * time.Second) // Adjust the interval as needed
	}
}
