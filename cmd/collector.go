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
	infoMetricProviders := []pkg.MetricProvider{
		gaugeValueFactory.CreateInfoMetricProvider(
			"nvme_namespace",
			"",
			"NameSpace",
			infoLabels,
		),
		gaugeValueFactory.CreateInfoMetricProvider(
			"nvme_used_bytes",
			"",
			"UsedBytes",
			infoLabels,
		),
		gaugeValueFactory.CreateInfoMetricProvider(
			"nvme_maximum_lba",
			"",
			"MaximumLBA",
			infoLabels,
		),
		gaugeValueFactory.CreateInfoMetricProvider(
			"nvme_physical_size",
			"",
			"PhysicalSize",
			infoLabels,
		),
		gaugeValueFactory.CreateInfoMetricProvider(
			"nvme_sector_size",
			"",
			"SectorSize",
			infoLabels,
		),
	}

	// Smart-log metrics
	logMetricProviders := []pkg.MetricProvider{
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_critical_warning",
			"Critical warnings for the state of the controller",
			"critical_warning",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_temperature",
			"Temperature in degrees fahrenheit",
			"temperature",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_avail_spare",
			"Normalized percentage of remaining spare capacity available",
			"avail_spare",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_spare_thresh",
			"Async event completion may occur when avail spare < threshold",
			"spare_thresh",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_percent_used",
			"Vendor specific estimate of the percentage of life used",
			"percent_used",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_endurance_grp_critical_warning_summary",
			"Critical warnings for the state of endurance groups",
			"endurance_grp_critical_warning_summary",
		),

		counterValueFactory.CreateLogMetricProvider(
			"nvme_data_units_read",
			"Number of 512 byte data units host has read",
			"data_units_read",
		),

		counterValueFactory.CreateLogMetricProvider(
			"nvme_data_units_written",
			"Number of 512 byte data units the host has written",
			"data_units_written",
		),

		counterValueFactory.CreateLogMetricProvider(
			"nvme_host_read_commands",
			"Number of read commands completed",
			"host_read_commands",
		),

		counterValueFactory.CreateLogMetricProvider(
			"nvme_host_write_commands",
			"Number of write commands completed",
			"host_write_commands",
		),

		counterValueFactory.CreateLogMetricProvider(
			"nvme_controller_busy_time",
			"Amount of time in minutes controller busy with IO commands",
			"controller_busy_time",
		),

		counterValueFactory.CreateLogMetricProvider(
			"nvme_power_cycles",
			"Number of power cycles",
			"power_cycles",
		),

		counterValueFactory.CreateLogMetricProvider(
			"nvme_power_on_hours",
			"Number of power on hours",
			"power_on_hours",
		),

		counterValueFactory.CreateLogMetricProvider(
			"nvme_unsafe_shutdowns",
			"Number of unsafe shutdowns",
			"unsafe_shutdowns",
		),

		counterValueFactory.CreateLogMetricProvider(
			"nvme_media_errors",
			"Number of unrecovered data integrity errors",
			"media_errors",
		),

		counterValueFactory.CreateLogMetricProvider(
			"nvme_num_err_log_entries",
			"Lifetime number of error log entries",
			"num_err_log_entries",
		),

		counterValueFactory.CreateLogMetricProvider(
			"nvme_warning_temp_time",
			"Amount of time in minutes temperature > warning threshold",
			"warning_temp_time",
		),

		counterValueFactory.CreateLogMetricProvider(
			"nvme_critical_comp_time",
			"Amount of time in minutes temperature > critical threshold",
			"critical_comp_time",
		),

		counterValueFactory.CreateLogMetricProvider(
			"nvme_thm_temp1_trans_count",
			"Number of times controller transitioned to lower power",
			"thm_temp1_trans_count",
		),

		counterValueFactory.CreateLogMetricProvider(
			"nvme_thm_temp2_trans_count",
			"Number of times controller transitioned to lower power",
			"thm_temp2_trans_count",
		),

		counterValueFactory.CreateLogMetricProvider(
			"nvme_thm_temp1_trans_time",
			"Total number of seconds controller transitioned to lower power",
			"thm_temp1_total_time",
		),

		counterValueFactory.CreateLogMetricProvider(
			"nvme_thm_temp2_trans_time",
			"Total number of seconds controller transitioned to lower power",
			"thm_temp2_total_time",
		),
	}

	// OCP smart-log metrics
	ocpLogMetricProviders := []pkg.MetricProvider{
		counterValueFactory.CreateLogMetricProvider(
			"nvme_physical_media_units_written_hi",
			"Physical meda units written high",
			"Physical media units written.hi",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_physical_media_units_written_lo",
			"Physical meda units written low",
			"Physical media units written.lo",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_physical_media_units_read_hi",
			"Physical meda units read high",
			"Physical media units read.hi",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_physical_media_units_read_lo",
			"Physical meda units read low",
			"Physical media units read.lo",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_bad_user_nand_blocks_raw",
			"",
			"Bad user nand blocks - Raw",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_bad_user_nand_blocks_normalized",
			"",
			"Bad user nand blocks - Normalized",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_bad_system_nand_blocks_raw",
			"",
			"Bad system nand blocks - Raw",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_bad_system_nand_blocks_normalized",
			"",
			"Bad system nand blocks - Normalized",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_xor_recovery_count",
			"",
			"XOR recovery count",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_uncorrectable_uead_error_count",
			"",
			"Uncorrectable read error count",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_soft_ecc_error_count",
			"",
			"Soft ecc error count",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_end_to_end_detected_errors",
			"",
			"End to end detected errors",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_end_to_end_corrected_errors",
			"",
			"End to end corrected errors",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_system_data_percent_used",
			"",
			"System data percent used",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_refresh_counts",
			"",
			"Refresh counts",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_max_user_data_erase_counts",
			"",
			"Max User data erase counts",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_min_user_data_erase_counts",
			"",
			"Min User data erase counts",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_number_of_thermal_throttling_events",
			"",
			"Number of Thermal throttling events",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_current_throttling_status",
			"",
			"Current throttling status",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_pcie_correctable_error_count",
			"",
			"PCIe correctable error count",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_incomplete_shutdowns",
			"",
			"Incomplete shutdowns",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_percent_free_blocks",
			"",
			"Percent free blocks",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_capacitor_health",
			"",
			"Capacitor health",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_unaligned_io",
			"",
			"Unaligned I/O",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_security_version_number",
			"",
			"Security Version Number",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_nuse_namespace_utilization",
			"",
			"NUSE - Namespace utilization",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_plp_start_count",
			"",
			"PLP start count",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_endurance_estimate",
			"",
			"Endurance estimate",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_log_page_version",
			"",
			"Log page version",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_log_page_guid",
			"",
			"Log page GUID",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_errata_version_field",
			"",
			"Errata Version Field",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_point_version_field",
			"",
			"Point Version Field",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_minor_version_field",
			"",
			"Minor Version Field",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_major_version_field",
			"",
			"Major Version Field",
		),
		gaugeValueFactory.CreateLogMetricProvider(
			"nvme_nvme_errata_version",
			"",
			"NVMe Errata Version",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_pcie_link_retraining_count",
			"",
			"PCIe Link Retraining Count",
		),
		counterValueFactory.CreateLogMetricProvider(
			"nvme_power_state_change_count",
			"",
			"Power State Change Count",
		),
	}

	// the info and smart-log collectors are always present
	collectors := []pkg.MetricCollector{
		pkg.NewInfoMetricCollector(infoMetricProviders),
		pkg.NewLogMetricCollector(logMetricProviders, getSmartLogData),
	}

	if ocpEnabled {
		collectors = append(collectors, pkg.NewLogMetricCollector(ocpLogMetricProviders, getOcpSmartLogData))
	}

	return pkg.NewCompositeCollector(collectors)
}
