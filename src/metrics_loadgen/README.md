# Metrics Load Generator

## Overview

The `metrics_loadgen` application is a Go-based tool that continuously monitors a specified input directory for JSON metric files. It processes each file by updating the `k8s.cluster.name` attribute for multiple simulated clusters and outputs the modified JSON data. The program runs in an infinite loop until interrupted with `CTRL+C`.

## Features

- Reads configuration from `config.yaml` (default) or a user-specified file via `--config`.
- Monitors `input_dir` for JSON metric files.
- Updates `k8s.cluster.name` based on the `no_clusters` configuration.
- Outputs the modified JSON data instead of saving it to disk.
- Supports graceful shutdown with `CTRL+C`.

## Installation & Setup

### Prerequisites

- Go 1.16 or later installed.
- A valid `config.yaml` file in the root of the repository.

### Project Structure

```
/your-project-root/
│── config.yaml
│── go.mod
│── src/
│   ├── metrics_loadgen/
│   │   ├── main.go
```

### Initializing the Go Module

```sh
cd /your-project-root
go mod init your_project_name
```

### Building the Application

```sh
cd src/metrics_loadgen
go build -o ../../metrics_loadgen
```

### Running the Application

#### Default Configuration

```sh
./metrics_loadgen
```

#### Using a Custom Configuration File

```sh
./metrics_loadgen --config=my_custom_config.yaml
```

#### Displaying Help

```sh
./metrics_loadgen -h
```

## Configuration (`config.yaml`)

```yaml
base_cluster_name: "demo"
no_clusters: 2
access_token: "your-access-token"
rum_token: "your-rum-token"
api_token: "your-api-token"
input_dir: "./metrics-org"
output_dir: "./metrics-new"
input_file: "./metric.json"
```

## How It Works

1. The application loads settings from `config.yaml`.
2. It continuously monitors `input_dir` for JSON metric files.
3. It reads each file and updates the `k8s.cluster.name` field based on the `no_clusters` setting.
4. Instead of saving the modified JSON, it prints it to the console.
5. The loop continues until the process is stopped with `CTRL+C`.

## License
This project is open-source and available under the MIT License.