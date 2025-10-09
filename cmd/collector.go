package main

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/tidwall/gjson"

	"github.com/E4-Computer-Engineering/nvme_exporter/pkg"
	"github.com/E4-Computer-Engineering/nvme_exporter/pkg/utils"
)

func getSmartLogData(devicePath string) gjson.Result {
	smartLog, err := utils.ExecuteJSONCommand("nvme", "smart-log", devicePath, "-o", "json")
	if err != nil {
		log.Printf("Error running smart-log %s -o json: %s\n", devicePath, err)
	}

	return smartLog
}

func getOcpSmartLogData(devicePath string) gjson.Result {
	ocpSmartLog, err := utils.ExecuteJSONCommand("nvme", "ocp", "smart-add-log", devicePath, "-o", "json")
	if err != nil {
		log.Printf("OCP metrics not supported or error running smart-add-log %s -o json: %s "+
			"(continuing with standard metrics)\n", devicePath, err)
		// Return empty result instead of crashing
		return gjson.Result{}
	}

	return ocpSmartLog
}

type ProviderFactory struct {
	valueType     prometheus.ValueType
	defaultLabels []string
}

func (f *ProviderFactory) CreateLogMetricProvider(
	fqName string,
	help string,
	jsonKey string,
) pkg.MetricProvider {
	return pkg.NewMetricProvider(
		prometheus.NewDesc(
			fqName,
			help,
			f.defaultLabels,
			nil,
		),
		f.valueType,
		jsonKey,
	)
}

func (f *ProviderFactory) CreateInfoMetricProvider(
	fqName string,
	help string,
	jsonKey string,
	infoLabels []string,
) pkg.MetricProvider {
	return pkg.NewMetricProvider(
		prometheus.NewDesc(
			fqName,
			help,
			infoLabels,
			nil,
		),
		f.valueType,
		jsonKey,
	)
}

func newNvmeCollector(collectorStates map[string]bool) prometheus.Collector {
	labels := []string{"device"}
	infoLabels := []string{"device", "generic_path", "firmware", "model_number", "serial_number"}

	gaugeValueFactory := ProviderFactory{
		valueType:     prometheus.GaugeValue,
		defaultLabels: labels,
	}

	counterValueFactory := ProviderFactory{
		valueType:     prometheus.CounterValue,
		defaultLabels: labels,
	}

	// Info metrics
	infoMetricProviders := []pkg.MetricProvider{
		gaugeValueFactory.CreateInfoMetricProvider(
			"nvme_namespace",
			"NVMe namespace identifier",
			"NameSpace",
			infoLabels,
		),
		gaugeValueFactory.CreateInfoMetricProvider(
			"nvme_used_bytes",
			"Used storage capacity in bytes",
			"UsedBytes",
			infoLabels,
		),
		gaugeValueFactory.CreateInfoMetricProvider(
			"nvme_maximum_lba",
			"Maximum Logical Block Address",
			"MaximumLBA",
			infoLabels,
		),
		gaugeValueFactory.CreateInfoMetricProvider(
			"nvme_physical_size",
			"Physical size in bytes",
			"PhysicalSize",
			infoLabels,
		),
		gaugeValueFactory.CreateInfoMetricProvider(
			"nvme_sector_size",
			"Sector size in bytes",
			"SectorSize",
			infoLabels,
		),
	}

	// Smart-log metrics
	logMetricProviders := []pkg.MetricProvider{
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_critical_warning",
			"Critical warnings for the controller state. Bits indicate spare capacity, temperature, "+
				"degraded reliability, or read-only mode",
			"critical_warning",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_temperature",
			"Current composite temperature in Kelvin",
			"temperature",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_avail_spare",
			"Available spare capacity as a normalized percentage (0-100)",
			"avail_spare",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_spare_thresh",
			"Available spare capacity threshold below which an asynchronous event is generated",
			"spare_thresh",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_percent_used",
			"Vendor-specific estimate of the percentage of device life used (0-255)",
			"percent_used",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_endurance_grp_critical_warning_summary",
			"Critical warnings for endurance groups. Contains the OR of all critical warnings for all endurance groups",
			"endurance_grp_critical_warning_summary",
		),

		counterValueFactory.CreateLogMetricProvider(
			"nvme_data_units_read",
			"Total number of 512-byte data units read from the NVMe device by the host",
			"data_units_read",
		),

		counterValueFactory.CreateLogMetricProvider(
			"nvme_data_units_written",
			"Total number of 512-byte data units written to the NVMe device by the host",
			"data_units_written",
		),

		counterValueFactory.CreateLogMetricProvider(
			"nvme_host_read_commands",
			"Total number of read commands completed by the controller",
			"host_read_commands",
		),

		counterValueFactory.CreateLogMetricProvider(
			"nvme_host_write_commands",
			"Total number of write commands completed by the controller",
			"host_write_commands",
		),

		counterValueFactory.CreateLogMetricProvider(
			"nvme_controller_busy_time",
			"Total time in minutes the controller was busy processing I/O commands",
			"controller_busy_time",
		),

		counterValueFactory.CreateLogMetricProvider(
			"nvme_power_cycles",
			"Total number of power cycles",
			"power_cycles",
		),

		counterValueFactory.CreateLogMetricProvider(
			"nvme_power_on_hours",
			"Total number of power-on hours. May not include time when the controller was powered but in a low power state",
			"power_on_hours",
		),

		counterValueFactory.CreateLogMetricProvider(
			"nvme_unsafe_shutdowns",
			"Total number of unsafe shutdowns where the controller was not properly notified before power loss",
			"unsafe_shutdowns",
		),

		counterValueFactory.CreateLogMetricProvider(
			"nvme_media_errors",
			"Total number of unrecovered data integrity errors detected by the controller",
			"media_errors",
		),

		counterValueFactory.CreateLogMetricProvider(
			"nvme_num_err_log_entries",
			"Lifetime number of error log entries available in the Error Information Log",
			"num_err_log_entries",
		),

		counterValueFactory.CreateLogMetricProvider(
			"nvme_warning_temp_time",
			"Total time in minutes the controller temperature exceeded the warning threshold",
			"warning_temp_time",
		),

		counterValueFactory.CreateLogMetricProvider(
			"nvme_critical_comp_time",
			"Total time in minutes the controller temperature exceeded the critical composite temperature threshold",
			"critical_comp_time",
		),

		counterValueFactory.CreateLogMetricProvider(
			"nvme_thm_temp1_trans_count",
			"Total number of times the controller transitioned to a lower power state due to thermal management (threshold 1)",
			"thm_temp1_trans_count",
		),

		counterValueFactory.CreateLogMetricProvider(
			"nvme_thm_temp2_trans_count",
			"Total number of times the controller transitioned to a lower power state due to thermal management (threshold 2)",
			"thm_temp2_trans_count",
		),

		counterValueFactory.CreateLogMetricProvider(
			"nvme_thm_temp1_trans_time",
			"Total time in seconds the controller was in a lower power state due to thermal management (threshold 1)",
			"thm_temp1_total_time",
		),

		counterValueFactory.CreateLogMetricProvider(
			"nvme_thm_temp2_trans_time",
			"Total time in seconds the controller was in a lower power state due to thermal management (threshold 2)",
			"thm_temp2_total_time",
		),
	}

	// OCP smart-log metrics
	ocpLogMetricProviders := []pkg.MetricProvider{
		counterValueFactory.CreateLogMetricProvider(
			"nvme_physical_media_units_written_hi",
			"Physical media units written to the device (high 64 bits). Unit size is 1000h sector size",
			"Physical media units written.hi",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_physical_media_units_written_lo",
			"Physical media units written to the device (low 64 bits). Unit size is 1000h sector size",
			"Physical media units written.lo",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_physical_media_units_read_hi",
			"Physical media units read from the device (high 64 bits). Unit size is 1000h sector size",
			"Physical media units read.hi",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_physical_media_units_read_lo",
			"Physical media units read from the device (low 64 bits). Unit size is 1000h sector size",
			"Physical media units read.lo",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_bad_user_nand_blocks_raw",
			"Raw count of user NAND blocks that have been retired due to errors",
			"Bad user nand blocks - Raw",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_bad_user_nand_blocks_normalized",
			"Normalized value (0-100) of bad user NAND blocks relative to the maximum allowed",
			"Bad user nand blocks - Normalized",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_bad_system_nand_blocks_raw",
			"Raw count of system area NAND blocks that have been retired due to errors",
			"Bad system nand blocks - Raw",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_bad_system_nand_blocks_normalized",
			"Normalized value (0-100) of bad system NAND blocks relative to the maximum allowed",
			"Bad system nand blocks - Normalized",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_xor_recovery_count",
			"Total number of times data was recovered using XOR parity",
			"XOR recovery count",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_uncorrectable_read_error_count",
			"Total number of uncorrectable read errors that could not be recovered",
			"Uncorrectable read error count",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_soft_ecc_error_count",
			"Total number of soft ECC errors that were corrected",
			"Soft ecc error count",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_end_to_end_detected_errors",
			"Total number of end-to-end data protection errors detected",
			"End to end detected errors",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_end_to_end_corrected_errors",
			"Total number of end-to-end data protection errors that were corrected",
			"End to end corrected errors",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_system_data_percent_used",
			"Percentage of system data area used (0-100)",
			"System data percent used",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_refresh_counts",
			"Total number of NAND page refresh operations performed",
			"Refresh counts",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_max_user_data_erase_counts",
			"Maximum number of erase cycles performed on any user data block",
			"Max User data erase counts",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_min_user_data_erase_counts",
			"Minimum number of erase cycles performed on any user data block",
			"Min User data erase counts",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_number_of_thermal_throttling_events",
			"Total number of times thermal throttling was activated",
			"Number of Thermal throttling events",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_current_throttling_status",
			"Current thermal throttling status (0=not throttled, 1=throttled)",
			"Current throttling status",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_pcie_correctable_error_count",
			"Total number of PCIe correctable errors detected",
			"PCIe correctable error count",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_incomplete_shutdowns",
			"Total number of incomplete or unsafe shutdown events",
			"Incomplete shutdowns",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_percent_free_blocks",
			"Percentage of free NAND blocks available (0-100)",
			"Percent free blocks",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_capacitor_health",
			"Health indicator of the power loss protection capacitor (vendor-specific scale)",
			"Capacitor health",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_unaligned_io",
			"Total number of unaligned I/O operations performed",
			"Unaligned I/O",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_security_version_number",
			"Security version number of the device firmware",
			"Security Version Number",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_nuse_namespace_utilization",
			"Namespace utilization as reported by the device",
			"NUSE - Namespace utilization",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_plp_start_count",
			"Total number of times the Power Loss Protection (PLP) mechanism was activated",
			"PLP start count",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_endurance_estimate",
			"Estimated remaining endurance of the device as a percentage (0-100)",
			"Endurance estimate",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_log_page_version",
			"Version number of the OCP SMART log page specification",
			"Log page version",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_log_page_guid",
			"GUID (Globally Unique Identifier) of the OCP SMART log page",
			"Log page GUID",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_errata_version_field",
			"Errata version field from the OCP specification version",
			"Errata Version Field",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_point_version_field",
			"Point version field from the OCP specification version",
			"Point Version Field",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_minor_version_field",
			"Minor version field from the OCP specification version",
			"Minor Version Field",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_major_version_field",
			"Major version field from the OCP specification version",
			"Major Version Field",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_nvme_errata_version",
			"NVMe base specification errata version supported by the device",
			"NVMe Errata Version",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_pcie_link_retraining_count",
			"Total number of PCIe link retraining events",
			"PCIe Link Retraining Count",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_power_state_change_count",
			"Total number of power state transitions",
			"Power State Change Count",
		),
	}

	// Build collectors based on enabled states
	collectors := []pkg.MetricCollector{}

	// Add info collector if enabled
	if collectorStates["info"] {
		collectors = append(collectors, pkg.NewInfoMetricCollector(infoMetricProviders))
	}

	// Add smart-log collector if enabled
	if collectorStates["smart"] {
		collectors = append(collectors, pkg.NewLogMetricCollector(logMetricProviders, getSmartLogData))
	}

	// Add OCP collector if enabled (now enabled by default)
	if collectorStates["ocp"] {
		collectors = append(collectors, pkg.NewLogMetricCollector(ocpLogMetricProviders, getOcpSmartLogData))
	}

	return pkg.NewCompositeCollector(collectors)
}
