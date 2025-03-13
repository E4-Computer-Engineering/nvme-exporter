# nvme_exporter

Prometheus exporter for nvme smart-log and OCP smart-log metrics inspired by [fritchie nvme exporter](https://github.com/fritchie/nvme_exporter)

Specification versions of reference:

* nvme smart-log field descriptions can be found on page 209 of:
https://nvmexpress.org/wp-content/uploads/NVM-Express-Base-Specification-Revision-2.1-2024.08.05-Ratified.pdf

* nvme ocp-smart-log field descriptions can be found on page 24 of:
https://www.opencompute.org/documents/datacenter-nvme-ssd-specification-v2-5-pdf

Supported [NVMe CLI](https://github.com/linux-nvme/nvme-cli) versions:

| Version | Supported |
|----|----|
|2.9 | OK |
|2.10 | TBD |
|2.11 | TBD |

## Metrics

This collector exports the output of the following `nvme` cli commands:

``` bash
nvme list
nvme smart-log <device_name>
nvme ocp-smart-add-log <device_name>
```

## Content

* Docker: A sample Dockerfile and docker-compose.yaml are provided.
* Kubernetes: In [resources](resources/k8s/).
* Grafana: In [resources](resources/grafana/) for dashboards.
  * [smart-log and OCP dashboard](https://github.com/E4-Computer-Engineering/nvme-exporter/blob/main/resources/grafana/dashboard_SMART_OCP.json)
  * [smart-log dashboard](https://github.com/E4-Computer-Engineering/nvme-exporter/blob/main/resources/grafana/dashboard_SMART.json)
* Prometheus: In [resources](resources/prom/) for recording and alert rules.
* Systemd: In [resources](resources/systemd/) for executing the exporter as unit.
* Scripts: In [resources](resources/scripts/) for package installation hooks.

## Running

Running the exporter requires the nvme-cli package to be installed on the host and be `root` account.

``` bash
nvme_exporter -h
```

### Flags

| Name | Description | Default |
|----|----|----|
|port | Listen port number. Type: String. | `9998` |
|ocp | Enable OCP smart log metrics. Type: Bool. | `false` |
|endpoint | The endpoint to query for metrics. Type: String. | `/metrics` |
