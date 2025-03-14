# NVMe Exporter

[![build](https://github.com/E4-Computer-Engineering/nvme-exporter/actions/workflows/build.yml/badge.svg)](https://github.com/E4-Computer-Engineering/nvme-exporter/actions/workflows/build.yml)
![Latest GitHub release](https://img.shields.io/github/release/E4-Computer-Engineering/nvme-exporter.svg)
[![GitHub license](https://img.shields.io/github/license/E4-Computer-Engineering/nvme-exporter)](https://github.com/E4-Computer-Engineering/nvme-exporter/blob/master/LICENSE)
![GitHub all releases](https://img.shields.io/github/downloads/E4-Computer-Engineering/nvme-exporter/total)

Prometheus exporter for nvme smart-log and OCP smart-log metrics inspired by [fritchie nvme exporter](https://github.com/fritchie/nvme_exporter).

Specification versions of reference:

* nvme smart-log field descriptions can be found on page 209 of [NVMe specifications](https://nvmexpress.org/wp-content/uploads/NVM-Express-Base-Specification-Revision-2.1-2024.08.05-Ratified.pdf)

* nvme ocp-smart-log field descriptions can be found on page 24 of [Opencompute NVMe SSD specifications](https://www.opencompute.org/documents/datacenter-nvme-ssd-specification-v2-5-pdf)

Supported [NVMe CLI](https://github.com/linux-nvme/nvme-cli) versions:

| Version | Supported |
|----|----|
|2.9 | OK |
|2.10 | OK |
|2.11 | TBD |

## Repo Content

* Docker: A sample `Dockerfile` is provided.
* Kubernetes: In [resources](resources/k8s/).
* Grafana: In [resources](resources/grafana/) for dashboards.
  * [smart-log and OCP dashboard](https://github.com/E4-Computer-Engineering/nvme-exporter/blob/main/resources/grafana/dashboard_SMART_OCP.json)
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

### Systemd

By installing the packaged version: RPM or DEB, the systemd unit will be automatically deployed and started as `nvme_exporter.service`.
If you are installing from `tar.gz` the [systemd unit file](resources/systemd/nvme_exporter.service) is provided in this repo.

> NOTE: if you want to execute with custom flags you will need to modify the unit file

### Container

To run the exporter as a container with, for example, OCP metrics enabled:

``` bash
podman run --rm -d --network=host --privileged nvme_exporter -ocp
```

## Visualization

This is how the dashboard visualizes:

![OCP metrics](https://raw.githubusercontent.com/E4-Computer-Engineering/nvme-exporter/refs/heads/main/resources/grafana/nvme_ocp.png)

![Endurance metrics](https://raw.githubusercontent.com/E4-Computer-Engineering/nvme-exporter/refs/heads/main/resources/grafana/nvme_endurance.png)

![Stats metrics](https://raw.githubusercontent.com/E4-Computer-Engineering/nvme-exporter/refs/heads/main/resources/grafana/nvme_stats.png)

![Errors metrics](https://raw.githubusercontent.com/E4-Computer-Engineering/nvme-exporter/refs/heads/main/resources/grafana/nvme_errors.png)

## Metrics

This collector exports the output of the following `nvme` cli commands:

``` bash
nvme list
nvme smart-log <device_name>
nvme ocp-smart-add-log <device_name>
```

|metric_name|description|
|---|---|
|nvme_avail_spare|---|
|nvme_bad_system_nand_blocks_normalized|---|
|nvme_bad_system_nand_blocks_raw|---|
|nvme_bad_user_nand_blocks_normalized|---|
|nvme_bad_user_nand_blocks_raw|---|
|nvme_capacitor_health|---|
|nvme_controller_busy_time|---|
|nvme_critical_comp_time|---|
|nvme_critical_warning|---|
|nvme_current_throttling_status|---|
|nvme_data_units_read|---|
|nvme_data_units_written|---|
|nvme_end_to_end_corrected_errors|---|
|nvme_end_to_end_detected_errors|---|
|nvme_endurance_estimate|---|
|nvme_endurance_grp_critical_warning_summary|---|
|nvme_errata_version_field|---|
|nvme_host_read_commands|---|
|nvme_host_write_commands|---|
|nvme_incomplete_shutdowns|---|
|nvme_log_page_guid|---|
|nvme_log_page_version|---|
|nvme_major_version_field|---|
|nvme_maximum_lba|---|
|nvme_max_user_data_erase_counts|---|
|nvme_media_errors|---|
|nvme_minor_version_field|---|
|nvme_min_user_data_erase_counts|---|
|nvme_namespace|---|
|nvme_number_of_thermal_throttling_events|---|
|nvme_num_err_log_entries|---|
|nvme_nuse_namespace_utilization|---|
|nvme_nvme_errata_version|---|
|nvme_pcie_correctable_error_count|---|
|nvme_pcie_link_retraining_count|---|
|nvme_percent_free_blocks|---|
|nvme_percent_used|---|
|nvme_physical_media_units_read_hi|---|
|nvme_physical_media_units_read_lo|---|
|nvme_physical_media_units_written_hi|---|
|nvme_physical_media_units_written_lo|---|
|nvme_physical_size|---|
|nvme_plp_start_count|---|
|nvme_point_version_field|---|
|nvme_power_cycles|---|
|nvme_power_on_hours|---|
|nvme_power_state_change_count|---|
|nvme_refresh_counts|---|
|nvme_sector_size|---|
|nvme_security_version_number|---|
|nvme_soft_ecc_error_count|---|
|nvme_spare_thresh|---|
|nvme_system_data_percent_used|---|
|nvme_temperature|---|
|nvme_thm_temp1_trans_count|---|
|nvme_thm_temp1_trans_time|---|
|nvme_thm_temp2_trans_count|---|
|nvme_thm_temp2_trans_time|---|
|nvme_unaligned_io|---|
|nvme_uncorrectable_uead_error_count|---|
|nvme_unsafe_shutdowns|---|
|nvme_used_bytes|---|
|nvme_warning_temp_time|---|
|nvme_xor_recovery_count|---|
