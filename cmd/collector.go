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
		log.Printf("Error running smart-add-log %s -o json: %s\n", devicePath, err)
	}

	return ocpSmartLog
}

type ProviderFactory struct {
	valueType     prometheus.ValueType
	defaultLabels []string
}

func (f *ProviderFactory) NewLogMetricProvider(
	fqName string,
	help string,
	jsonKey string,
) pkg.LogMetricProvider {
	return pkg.NewLogMetricProvider(
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

func (f *ProviderFactory) NewInfoMetricProvider(
	fqName string,
	help string,
	jsonKey string,
	infoLabels []string,
) pkg.InfoMetricProvider {
	return pkg.NewInfoMetricProvider(
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

func newNvmeCollector(ocpEnabled bool) prometheus.Collector {
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
	infoMetricProviders := []pkg.InfoMetricProvider{
		gaugeValueFactory.NewInfoMetricProvider(
			"nvme_namespace",
			"",
			"NameSpace",
			infoLabels,
		),
		gaugeValueFactory.NewInfoMetricProvider(
			"nvme_used_bytes",
			"",
			"UsedBytes",
			infoLabels,
		),
		gaugeValueFactory.NewInfoMetricProvider(
			"nvme_maximum_lba",
			"",
			"MaximumLBA",
			infoLabels,
		),
		gaugeValueFactory.NewInfoMetricProvider(
			"nvme_physical_size",
			"",
			"PhysicalSize",
			infoLabels,
		),
		gaugeValueFactory.NewInfoMetricProvider(
			"nvme_sector_size",
			"",
			"SectorSize",
			infoLabels,
		),
	}

	// Smart-log metrics
	logMetricProviders := []pkg.LogMetricProvider{
		gaugeValueFactory.NewLogMetricProvider(
			"nvme_critical_warning",
			"Critical warnings for the state of the controller",
			"critical_warning",
		),
		gaugeValueFactory.NewLogMetricProvider(
			"nvme_temperature",
			"Temperature in degrees fahrenheit",
			"temperature",
		),
		gaugeValueFactory.NewLogMetricProvider(
			"nvme_avail_spare",
			"Normalized percentage of remaining spare capacity available",
			"avail_spare",
		),
		gaugeValueFactory.NewLogMetricProvider(
			"nvme_spare_thresh",
			"Async event completion may occur when avail spare < threshold",
			"spare_thresh",
		),
		gaugeValueFactory.NewLogMetricProvider(
			"nvme_percent_used",
			"Vendor specific estimate of the percentage of life used",
			"percent_used",
		),
		gaugeValueFactory.NewLogMetricProvider(
			"nvme_endurance_grp_critical_warning_summary",
			"Critical warnings for the state of endurance groups",
			"endurance_grp_critical_warning_summary",
		),

		counterValueFactory.NewLogMetricProvider(
			"nvme_data_units_read",
			"Number of 512 byte data units host has read",
			"data_units_read",
		),

		counterValueFactory.NewLogMetricProvider(
			"nvme_data_units_written",
			"Number of 512 byte data units the host has written",
			"data_units_written",
		),

		counterValueFactory.NewLogMetricProvider(
			"nvme_host_read_commands",
			"Number of read commands completed",
			"host_read_commands",
		),

		counterValueFactory.NewLogMetricProvider(
			"nvme_host_write_commands",
			"Number of write commands completed",
			"host_write_commands",
		),

		counterValueFactory.NewLogMetricProvider(
			"nvme_controller_busy_time",
			"Amount of time in minutes controller busy with IO commands",
			"controller_busy_time",
		),

		counterValueFactory.NewLogMetricProvider(
			"nvme_power_cycles",
			"Number of power cycles",
			"power_cycles",
		),

		counterValueFactory.NewLogMetricProvider(
			"nvme_power_on_hours",
			"Number of power on hours",
			"power_on_hours",
		),

		counterValueFactory.NewLogMetricProvider(
			"nvme_unsafe_shutdowns",
			"Number of unsafe shutdowns",
			"unsafe_shutdowns",
		),

		counterValueFactory.NewLogMetricProvider(
			"nvme_media_errors",
			"Number of unrecovered data integrity errors",
			"media_errors",
		),

		counterValueFactory.NewLogMetricProvider(
			"nvme_num_err_log_entries",
			"Lifetime number of error log entries",
			"num_err_log_entries",
		),

		counterValueFactory.NewLogMetricProvider(
			"nvme_warning_temp_time",
			"Amount of time in minutes temperature > warning threshold",
			"warning_temp_time",
		),

		counterValueFactory.NewLogMetricProvider(
			"nvme_critical_comp_time",
			"Amount of time in minutes temperature > critical threshold",
			"critical_comp_time",
		),

		counterValueFactory.NewLogMetricProvider(
			"nvme_thm_temp1_trans_count",
			"Number of times controller transitioned to lower power",
			"thm_temp1_trans_count",
		),

		counterValueFactory.NewLogMetricProvider(
			"nvme_thm_temp2_trans_count",
			"Number of times controller transitioned to lower power",
			"thm_temp2_trans_count",
		),

		counterValueFactory.NewLogMetricProvider(
			"nvme_thm_temp1_trans_time",
			"Total number of seconds controller transitioned to lower power",
			"thm_temp1_total_time",
		),

		counterValueFactory.NewLogMetricProvider(
			"nvme_thm_temp2_trans_time",
			"Total number of seconds controller transitioned to lower power",
			"thm_temp2_total_time",
		),
	}

	// OCP smart-log metrics
	ocpLogMetricProviders := []pkg.LogMetricProvider{
		counterValueFactory.NewLogMetricProvider(
			"nvme_physical_media_units_written_hi",
			"Physical meda units written high",
			"Physical media units written.hi",
		),
		counterValueFactory.NewLogMetricProvider(
			"nvme_physical_media_units_written_lo",
			"Physical meda units written low",
			"Physical media units written.lo",
		),
		counterValueFactory.NewLogMetricProvider(
			"nvme_physical_media_units_read_hi",
			"Physical meda units read high",
			"Physical media units read.hi",
		),
		counterValueFactory.NewLogMetricProvider(
			"nvme_physical_media_units_read_lo",
			"Physical meda units read low",
			"Physical media units read.lo",
		),
		counterValueFactory.NewLogMetricProvider(
			"nvme_bad_user_nand_blocks_raw",
			"",
			"Bad user nand blocks - Raw",
		),
		counterValueFactory.NewLogMetricProvider(
			"nvme_bad_user_nand_blocks_normalized",
			"",
			"Bad user nand blocks - Normalized",
		),
		counterValueFactory.NewLogMetricProvider(
			"nvme_bad_system_nand_blocks_raw",
			"",
			"Bad system nand blocks - Raw",
		),
		counterValueFactory.NewLogMetricProvider(
			"nvme_bad_system_nand_blocks_normalized",
			"",
			"Bad system nand blocks - Normalized",
		),
		counterValueFactory.NewLogMetricProvider(
			"nvme_xor_recovery_count",
			"",
			"XOR recovery count",
		),
		counterValueFactory.NewLogMetricProvider(
			"nvme_uncorrectable_uead_error_count",
			"",
			"Uncorrectable read error count",
		),
		counterValueFactory.NewLogMetricProvider(
			"nvme_soft_ecc_error_count",
			"",
			"Soft ecc error count",
		),
		counterValueFactory.NewLogMetricProvider(
			"nvme_end_to_end_detected_errors",
			"",
			"End to end detected errors",
		),
		counterValueFactory.NewLogMetricProvider(
			"nvme_end_to_end_corrected_errors",
			"",
			"End to end corrected errors",
		),
		gaugeValueFactory.NewLogMetricProvider(
			"nvme_system_data_percent_used",
			"",
			"System data percent used",
		),
		counterValueFactory.NewLogMetricProvider(
			"nvme_refresh_counts",
			"",
			"Refresh counts",
		),
		counterValueFactory.NewLogMetricProvider(
			"nvme_max_user_data_erase_counts",
			"",
			"Max User data erase counts",
		),
		counterValueFactory.NewLogMetricProvider(
			"nvme_min_user_data_erase_counts",
			"",
			"Min User data erase counts",
		),
		counterValueFactory.NewLogMetricProvider(
			"nvme_number_of_thermal_throttling_events",
			"",
			"Number of Thermal throttling events",
		),
		gaugeValueFactory.NewLogMetricProvider(
			"nvme_current_throttling_status",
			"",
			"Current throttling status",
		),
		counterValueFactory.NewLogMetricProvider(
			"nvme_pcie_correctable_error_count",
			"",
			"PCIe correctable error count",
		),
		counterValueFactory.NewLogMetricProvider(
			"nvme_incomplete_shutdowns",
			"",
			"Incomplete shutdowns",
		),
		gaugeValueFactory.NewLogMetricProvider(
			"nvme_percent_free_blocks",
			"",
			"Percent free blocks",
		),
		gaugeValueFactory.NewLogMetricProvider(
			"nvme_capacitor_health",
			"",
			"Capacitor health",
		),
		counterValueFactory.NewLogMetricProvider(
			"nvme_unaligned_io",
			"",
			"Unaligned I/O",
		),
		gaugeValueFactory.NewLogMetricProvider(
			"nvme_security_version_number",
			"",
			"Security Version Number",
		),
		gaugeValueFactory.NewLogMetricProvider(
			"nvme_nuse_namespace_utilization",
			"",
			"NUSE - Namespace utilization",
		),
		counterValueFactory.NewLogMetricProvider(
			"nvme_plp_start_count",
			"",
			"PLP start count",
		),
		gaugeValueFactory.NewLogMetricProvider(
			"nvme_endurance_estimate",
			"",
			"Endurance estimate",
		),
		gaugeValueFactory.NewLogMetricProvider(
			"nvme_log_page_version",
			"",
			"Log page version",
		),
		gaugeValueFactory.NewLogMetricProvider(
			"nvme_log_page_guid",
			"",
			"Log page GUID",
		),
		gaugeValueFactory.NewLogMetricProvider(
			"nvme_errata_version_field",
			"",
			"Errata Version Field",
		),
		gaugeValueFactory.NewLogMetricProvider(
			"nvme_point_version_field",
			"",
			"Point Version Field",
		),
		gaugeValueFactory.NewLogMetricProvider(
			"nvme_minor_version_field",
			"",
			"Minor Version Field",
		),
		gaugeValueFactory.NewLogMetricProvider(
			"nvme_major_version_field",
			"",
			"Major Version Field",
		),
		gaugeValueFactory.NewLogMetricProvider(
			"nvme_nvme_errata_version",
			"",
			"NVMe Errata Version",
		),
		counterValueFactory.NewLogMetricProvider(
			"nvme_pcie_link_retraining_count",
			"",
			"PCIe Link Retraining Count",
		),
		counterValueFactory.NewLogMetricProvider(
			"nvme_power_state_change_count",
			"",
			"Power State Change Count",
		),
	}

	var collectors []pkg.MetricCollector
	collectors = append(collectors, pkg.NewInfoMetricCollector(infoMetricProviders))
	collectors = append(collectors, pkg.NewLogMetricCollector(logMetricProviders, getSmartLogData))

	if ocpEnabled {
		collectors = append(collectors, pkg.NewLogMetricCollector(ocpLogMetricProviders, getOcpSmartLogData))
	}

	return pkg.NewCompositeCollector(collectors)
}
