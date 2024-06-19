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
      - targets: ['localhost:8000']
```

## Metrics

The exporter provides the following metrics:

- RAID Metrics:

  - `raid_status_<vdisk>`: Status of the RAID virtual disk.
  - `raid_redundancy_<vdisk>`: Remaining redundancy of the RAID virtual disk.
  - `raid_size_<vdisk>`: Size of the RAID virtual disk.
  - `raid_layout_<vdisk>`: Layout of the RAID virtual disk.

- NVMe Metrics:

  - `nvme_presence_<drive>`: Presence of the NVMe device.
  - `nvme_health_<drive>`: Health of the NVMe device.
  - `nvme_<drive>_<metric>`: Various NVMe SMART metrics, including temperature, usage, power cycles, etc.

## Development

### Project Structure

```sh
.
├── README.md
├── dell-disk-health-exporter
├── go.mod
├── go.sum
├── main.go
└── pkg
    ├── idrac
    │   ├── idrac.go
    └── smart
        └── smart.go
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

