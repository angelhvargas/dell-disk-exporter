# Dell Disk Health Exporter

## Overview

The Dell Disk Health Exporter is a Prometheus exporter for monitoring the health of RAID controllers and NVMe drives in Dell servers. This exporter collects metrics from iDRAC RAID controllers and NVMe SMART logs, making them available for Prometheus to scrape and visualize.

## Features

- Monitors the health status of iDRAC RAID controllers.
- Collects NVMe SMART metrics, including temperature, usage, power cycles, and more.
- Supports multiple architectures (amd64, arm64).
- Exposes metrics at `/metrics` endpoint.

## Requirements

- Go 1.21 or higher
- Dell iDRAC tools (`racadm`)
- NVMe CLI tools (`nvme`)

## Installation

1. **Clone the repository:**

```sh
git clone https://github.com/angelhvargas/dell-disk-health-exporter.git
cd dell-disk-health-exporter
```

2. **Build the exporter:**

```sh
go build -o dell-disk-exporter main.go
```

1. **Run the exporter:**
   
```sh
./dell-disk-exporter
```

## Usage

The exporter listens on port `9077` and exposes metrics at the `/metrics` endpoint. Configure your Prometheus server to scrape metrics from this endpoint.

Example Prometheus configuration:

```yaml
scrape_configs:
  - job_name: 'dell_disk_health'
    static_configs:
      - targets: ['<TARGET_IP>:9077']
```

## Metrics

The exporter provides the following metrics:

### RAID Metrics

- raid_status{vdisk}: Status of the RAID virtual disk.
- raid_redundancy{vdisk}: Remaining redundancy of the RAID virtual disk.
- raid_size{vdisk}: Size of the RAID virtual disk.
- raid_layout{vdisk}: Layout of the RAID virtual disk.

### NVMe Metrics

- nvme_presence{device}: Presence of the NVMe device.
- nvme_smart_log{device,metric}: Various NVMe SMART metrics, including:
  - avail_spare
  - controller_busy_time
  - critical_comp_time
  - critical_warning
  - data_units_read
  - data_units_written
  - endurance_grp_critical_warning_summary
  - host_read_commands
  - host_write_commands
  - media_errors
  - num_err_log_entries
  - percent_used
  - power_cycles
  - power_on_hours
  - spare_thresh
  - temperature
  - thm_temp1_total_time
  - thm_temp1_trans_count
  - thm_temp2_total_time
  - thm_temp2_trans_count
  - unsafe_shutdowns
  - warning_temp_time

## Development

### Project Structure

```sh
.
├── README.md
├── dell-disk-health-exporter
├── go.mod
├── go.sum
├── main.go
├── main_test.go
└── pkg
    ├── idrac
    │   ├── idrac.go
    │   └── idrac_test.go
    └── smart
        ├── smart.go
        └── smart_test.go
```

- `main.go`: Entry point of the application.
- `pkg/smart`: Package for NVMe SMART metrics.
- `pkg/idrac`: Package for iDRAC RAID controller metrics.

## Building and Running

To build and run the project locally, use the following commands:

```sh
go build -o dell-disk-exporter main.go
./dell-disk-exporter

```

## GitHub Actions

The project includes a GitHub Actions workflow to automate the build and release process. The workflow builds the project for both amd64 and arm64 architectures and creates a release on GitHub.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.

## License

dell-disk-exporter is licensed under the [![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE). See the LICENSE file for more information.

## Setting Alerts in Prometheus

To set alerts based on the metrics provided by this exporter, you can add rules to your Prometheus configuration. For example, to alert when an NVMe device is absent:

```yaml
groups:
- name: NVMe Alerts
  rules:
  - alert: NVMeDeviceAbsent
    expr: nvme_presence == 0
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: "NVMe Device Absent (instance {{ $labels.instance }})"
      description: "NVMe device {{ $labels.device }} is absent for more than 5 minutes."

```

For RAID metrics, you can add rules such as:

```yaml
groups:
- name: RAID Alerts
  rules:
  - alert: RAIDStatusNotOk
    expr: raid_status != 1
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "RAID Status Not OK (instance {{ $labels.instance }})"
      description: "RAID virtual disk {{ $labels.vdisk }} has a status other than OK."

```
