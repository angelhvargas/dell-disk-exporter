package idrac

import (
	"log"
	"os/exec"
	"strings"
)

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) GetRAIDStatus() (map[string]string, error) {
	cmd := exec.Command("racadm", "raid", "get", "vdisks", "-o", "-p", "layout,status,RemainingRedundancy,Size")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				result[key] = value
			}
		}
	}
	return result, nil
}

func (c *Client) UpdateMetrics() {
	// Implement the logic to update metrics
	status, err := c.GetRAIDStatus()
	if err != nil {
		log.Printf("Error fetching RAID status: %v", err)
		return
	}
	log.Printf("RAID Status: %v", status)
}
