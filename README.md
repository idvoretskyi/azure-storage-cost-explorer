# Azure Storage Cost Explorer

[![CI](https://github.com/idvoretskyi/azure-storage-cost-explorer/actions/workflows/ci.yml/badge.svg)](https://github.com/idvoretskyi/azure-storage-cost-explorer/actions/workflows/ci.yml)
[![Release](https://github.com/idvoretskyi/azure-storage-cost-explorer/actions/workflows/release.yml/badge.svg)](https://github.com/idvoretskyi/azure-storage-cost-explorer/actions/workflows/release.yml)
[![Latest Release](https://img.shields.io/github/v/release/idvoretskyi/azure-storage-cost-explorer)](https://github.com/idvoretskyi/azure-storage-cost-explorer/releases/latest)
[![Go Version](https://img.shields.io/github/go-mod/go-version/idvoretskyi/azure-storage-cost-explorer)](go.mod)
[![Go Report Card](https://goreportcard.com/badge/github.com/idvoretskyi/azure-storage-cost-explorer)](https://goreportcard.com/report/github.com/idvoretskyi/azure-storage-cost-explorer)

A CLI tool to retrieve Azure Storage costs and usage across Blob, File, Queue, and Table services in your Azure subscription.

Inspired by [aws-s3-cost-explorer](https://github.com/idvoretskyi/aws-s3-cost-explorer).

Written in Go. Produces a single self-contained binary — no runtime or virtualenv required.

## Installation

### Using Homebrew

```bash
brew install idvoretskyi/tap/azure-storage-cost-explorer
```

### From source

```bash
git clone https://github.com/idvoretskyi/azure-storage-cost-explorer.git
cd azure-storage-cost-explorer
go build -o azure-storage-cost-explorer .
```

### Using `go install`

```bash
go install github.com/idvoretskyi/azure-storage-cost-explorer@latest
```

Ensure `$(go env GOPATH)/bin` is on your `$PATH`.

## Prerequisites

Azure credentials must be available via one of the standard methods supported by `DefaultAzureCredential`:

```bash
az login                                    # Azure CLI
# or environment variables (Service Principal):
export AZURE_TENANT_ID=...
export AZURE_CLIENT_ID=...
export AZURE_CLIENT_SECRET=...
export AZURE_SUBSCRIPTION_ID=...           # optional; resolves automatically otherwise
# or workload / managed identity (when running in Azure)
```

The subscription is resolved in this order:

1. `--subscription <id>` flag
2. `AZURE_SUBSCRIPTION_ID` environment variable
3. The first subscription returned by `az account list`

## Usage

### Get storage costs for the last 30 days

```bash
./azure-storage-cost-explorer costs
```

### Get costs for a specific period

```bash
./azure-storage-cost-explorer costs --days 7
```

### Export costs to CSV

```bash
./azure-storage-cost-explorer costs --csv costs.csv
./azure-storage-cost-explorer costs --days 7 --csv costs_7days.csv
```

### List all storage accounts with capacity summary

```bash
./azure-storage-cost-explorer accounts
```

### List all storage accounts with per-service breakdown

```bash
./azure-storage-cost-explorer accounts --detailed
```

### Export account information to CSV

```bash
./azure-storage-cost-explorer accounts --csv accounts.csv
./azure-storage-cost-explorer accounts --detailed --csv accounts_detailed.csv
```

### Get detailed capacity for a specific storage account

```bash
./azure-storage-cost-explorer account-details mystorageaccount
./azure-storage-cost-explorer account-details mystorageaccount --csv account.csv
```

### List all blob containers (with per-tier sizes)

```bash
./azure-storage-cost-explorer containers
./azure-storage-cost-explorer containers --detailed
./azure-storage-cost-explorer containers --csv containers.csv
```

> Note: Container-level sizes are computed by enumerating blobs (Azure Monitor only exposes account-level capacity). This may be slow on very large containers.

### Get per-tier breakdown for a specific container

```bash
./azure-storage-cost-explorer container-details mystorageaccount/mycontainer
./azure-storage-cost-explorer container-details mystorageaccount/mycontainer --csv container.csv
```

### List file shares, queues, tables

```bash
./azure-storage-cost-explorer shares [--csv shares.csv]
./azure-storage-cost-explorer queues [--csv queues.csv]
./azure-storage-cost-explorer tables [--csv tables.csv]
```

## Features

- Total Azure Storage costs with breakdown by `MeterSubCategory` (Blob, File, Queue, Table)
- Per-account capacity across all four storage services via Azure Monitor metrics
- Per-tier blob breakdown (Hot / Cool / Cold / Archive) at account and container level
- File-share quotas + used bytes
- Queue and Table enumeration
- Falls back to data-plane blob enumeration where Azure Monitor only exposes account-level metrics
- Human-readable size formatting (B / KB / MB / GB / TB / PB)
- Clean grid-style tabular output
- CSV export for all commands
- Subscription resolution via flag, env var, or auto-detect

## Required Azure RBAC Roles

| Service | Roles |
|---|---|
| Subscription | `Reader` |
| Storage Accounts (data plane) | `Storage Blob Data Reader` (for `containers`, `container-details`) |
| Cost Management | `Cost Management Reader` (for `costs`) |
| Azure Monitor | `Monitoring Reader` (typically included in `Reader`) |

## Caveats

- **Container-level metrics**: Azure Monitor's `BlobCapacity` is exposed only at the storage-account level (with a `Tier` dimension). Per-container size is computed by listing blobs and summing `Content-Length` per `Access Tier` — same trade-off as the AWS S3 `ListObjectsV2` fallback.
- **Cost data freshness**: Azure Cost Management can lag actual usage by up to 8–24 hours.
- **Account kinds**: Some account kinds (`BlobStorage`, `FileStorage`) do not support all four services; missing services simply report `0 B`.
- **Cold tier**: Included alongside Hot / Cool / Archive. Newer tiers added by Azure later will appear under their literal names.

## License

MIT — see [LICENSE](LICENSE).
