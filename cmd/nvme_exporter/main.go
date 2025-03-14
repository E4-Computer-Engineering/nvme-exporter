package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"os/user"
	"regexp"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tidwall/gjson"
)

var _supportedVersions = map[string]bool{
	"2.9": true,
	"2.10": true,
	"2.11": true,
}

func isSupportedVersion(version string) bool {
	_, ok := _supportedVersions[version]

	return ok
}

type nvmeCollector struct {
	ocp                                    bool
	nvmeCriticalWarning                    *prometheus.Desc
	nvmeTemperature                        *prometheus.Desc
	nvmeAvailSpare                         *prometheus.Desc
	nvmeSpareThresh                        *prometheus.Desc
	nvmePercentUsed                        *prometheus.Desc
	nvmeEnduranceGrpCriticalWarningSummary *prometheus.Desc
	nvmeDataUnitsRead                      *prometheus.Desc
	nvmeDataUnitsWritten                   *prometheus.Desc
	nvmeHostReadCommands                   *prometheus.Desc
	nvmeHostWriteCommands                  *prometheus.Desc
	nvmeControllerBusyTime                 *prometheus.Desc
	nvmePowerCycles                        *prometheus.Desc
	nvmePowerOnHours                       *prometheus.Desc
	nvmeUnsafeShutdowns                    *prometheus.Desc
	nvmeMediaErrors                        *prometheus.Desc
	nvmeNumErrLogEntries                   *prometheus.Desc
	nvmeWarningTempTime                    *prometheus.Desc
	nvmeCriticalCompTime                   *prometheus.Desc
	nvmeThmTemp1TransCount                 *prometheus.Desc
	nvmeThmTemp2TransCount                 *prometheus.Desc
	nvmeThmTemp1TotalTime                  *prometheus.Desc
	nvmeThmTemp2TotalTime                  *prometheus.Desc
	nvmePhysicalMediaUnitsWrittenHi        *prometheus.Desc
	nvmePhysicalMediaUnitsWrittenLo        *prometheus.Desc
	nvmePhysicalMediaUnitsReadHi           *prometheus.Desc
	nvmePhysicalMediaUnitsReadLo           *prometheus.Desc
	nvmeBadUserNandBlocksRaw               *prometheus.Desc
	nvmeBadUserNandBlocksNormalized        *prometheus.Desc
	nvmeBadSystemNandBlocksRaw             *prometheus.Desc
	nvmeBadSystemNandBlocksNormalized      *prometheus.Desc
	nvmeXorRecoveryCount                   *prometheus.Desc
	nvmeUncorrectableReadErrorCount        *prometheus.Desc
	nvmeSoftEccErrorCount                  *prometheus.Desc
	nvmeEndToEndDetectedErrors             *prometheus.Desc
	nvmeEndToEndCorrectedErrors            *prometheus.Desc
	nvmeSystemDataPercentUsed              *prometheus.Desc
	nvmeRefreshCounts                      *prometheus.Desc
	nvmeMaxUserDataEraseCounts             *prometheus.Desc
	nvmeMinUserDataEraseCounts             *prometheus.Desc
	nvmeNumberOfThermalThrottlingEvents    *prometheus.Desc
	nvmeCurrentThrottlingStatus            *prometheus.Desc
	nvmePcieCorrectableErrorCount          *prometheus.Desc
	nvmeIncompleteShutdowns                *prometheus.Desc
	nvmePercentFreeBlocks                  *prometheus.Desc
	nvmeCapacitorHealth                    *prometheus.Desc
	nvmeUnalignedIo                        *prometheus.Desc
	nvmeSecurityVersionNumber              *prometheus.Desc
	nvmeNuseNamespaceUtilization           *prometheus.Desc
	nvmePlpStartCount                      *prometheus.Desc
	nvmeEnduranceEstimate                  *prometheus.Desc
	nvmeLogPageVersion                     *prometheus.Desc
	nvmeLogPageGUID                        *prometheus.Desc
	nvmeErrataVersionField                 *prometheus.Desc
	nvmePointVersionField                  *prometheus.Desc
	nvmeMinorVersionField                  *prometheus.Desc
	nvmeMajorVersionField                  *prometheus.Desc
	nvmeNvmeErrataVersion                  *prometheus.Desc
	nvmePcieLinkRetrainingCount            *prometheus.Desc
	nvmePowerStateChangeCount              *prometheus.Desc
	nvmeNameSpace                          *prometheus.Desc
	nvmeUsedBytes                          *prometheus.Desc
	nvmeMaximumLba                         *prometheus.Desc
	nvmePhysicalSize                       *prometheus.Desc
	nvmeSectorSize                         *prometheus.Desc
}

func newNvmeCollector(ocp bool) prometheus.Collector {
	labels := []string{"device"}
	infoLabels := []string{"device", "generic_path", "firmware", "model_number", "serial_number"}

	return &nvmeCollector{
		ocp: ocp,
		nvmeCriticalWarning: prometheus.NewDesc(
			"nvme_critical_warning",
			"Critical warnings for the state of the controller",
			labels,
			nil,
		),
		nvmeTemperature: prometheus.NewDesc(
			"nvme_temperature",
			"Temperature in degrees fahrenheit",
			labels,
			nil,
		),
		nvmeAvailSpare: prometheus.NewDesc(
			"nvme_avail_spare",
			"Normalized percentage of remaining spare capacity available",
			labels,
			nil,
		),
		nvmeSpareThresh: prometheus.NewDesc(
			"nvme_spare_thresh",
			"Async event completion may occur when avail spare < threshold",
			labels,
			nil,
		),
		nvmePercentUsed: prometheus.NewDesc(
			"nvme_percent_used",
			"Vendor specific estimate of the percentage of life used",
			labels,
			nil,
		),
		nvmeEnduranceGrpCriticalWarningSummary: prometheus.NewDesc(
			"nvme_endurance_grp_critical_warning_summary",
			"Critical warnings for the state of endurance groups",
			labels,
			nil,
		),
		nvmeDataUnitsRead: prometheus.NewDesc(
			"nvme_data_units_read",
			"Number of 512 byte data units host has read",
			labels,
			nil,
		),
		nvmeDataUnitsWritten: prometheus.NewDesc(
			"nvme_data_units_written",
			"Number of 512 byte data units the host has written",
			labels,
			nil,
		),
		nvmeHostReadCommands: prometheus.NewDesc(
			"nvme_host_read_commands",
			"Number of read commands completed",
			labels,
			nil,
		),
		nvmeHostWriteCommands: prometheus.NewDesc(
			"nvme_host_write_commands",
			"Number of write commands completed",
			labels,
			nil,
		),
		nvmeControllerBusyTime: prometheus.NewDesc(
			"nvme_controller_busy_time",
			"Amount of time in minutes controller busy with IO commands",
			labels,
			nil,
		),
		nvmePowerCycles: prometheus.NewDesc(
			"nvme_power_cycles",
			"Number of power cycles",
			labels,
			nil,
		),
		nvmePowerOnHours: prometheus.NewDesc(
			"nvme_power_on_hours",
			"Number of power on hours",
			labels,
			nil,
		),
		nvmeUnsafeShutdowns: prometheus.NewDesc(
			"nvme_unsafe_shutdowns",
			"Number of unsafe shutdowns",
			labels,
			nil,
		),
		nvmeMediaErrors: prometheus.NewDesc(
			"nvme_media_errors",
			"Number of unrecovered data integrity errors",
			labels,
			nil,
		),
		nvmeNumErrLogEntries: prometheus.NewDesc(
			"nvme_num_err_log_entries",
			"Lifetime number of error log entries",
			labels,
			nil,
		),
		nvmeWarningTempTime: prometheus.NewDesc(
			"nvme_warning_temp_time",
			"Amount of time in minutes temperature > warning threshold",
			labels,
			nil,
		),
		nvmeCriticalCompTime: prometheus.NewDesc(
			"nvme_critical_comp_time",
			"Amount of time in minutes temperature > critical threshold",
			labels,
			nil,
		),
		nvmeThmTemp1TransCount: prometheus.NewDesc(
			"nvme_thm_temp1_trans_count",
			"Number of times controller transitioned to lower power",
			labels,
			nil,
		),
		nvmeThmTemp2TransCount: prometheus.NewDesc(
			"nvme_thm_temp2_trans_count",
			"Number of times controller transitioned to lower power",
			labels,
			nil,
		),
		nvmeThmTemp1TotalTime: prometheus.NewDesc(
			"nvme_thm_temp1_trans_time",
			"Total number of seconds controller transitioned to lower power",
			labels,
			nil,
		),
		nvmeThmTemp2TotalTime: prometheus.NewDesc(
			"nvme_thm_temp2_trans_time",
			"Total number of seconds controller transitioned to lower power",
			labels,
			nil,
		),
		nvmePhysicalMediaUnitsWrittenHi: prometheus.NewDesc(
			"nvme_physical_media_units_written_hi",
			"Physical meda units written high",
			labels,
			nil,
		),
		nvmePhysicalMediaUnitsWrittenLo: prometheus.NewDesc(
			"nvme_physical_media_units_written_lo",
			"Physical meda units written low",
			labels,
			nil,
		),
		nvmePhysicalMediaUnitsReadHi: prometheus.NewDesc(
			"nvme_physical_media_units_read_hi",
			"Physical meda units read high",
			labels,
			nil,
		),
		nvmePhysicalMediaUnitsReadLo: prometheus.NewDesc(
			"nvme_physical_media_units_read_lo",
			"Physical meda units read low",
			labels,
			nil,
		),
		nvmeBadUserNandBlocksRaw: prometheus.NewDesc(
			"nvme_bad_user_nand_blocks_raw",
			"",
			labels,
			nil,
		),
		nvmeBadUserNandBlocksNormalized: prometheus.NewDesc(
			"nvme_bad_user_nand_blocks_normalized",
			"",
			labels,
			nil,
		),
		nvmeBadSystemNandBlocksRaw: prometheus.NewDesc(
			"nvme_bad_system_nand_blocks_raw",
			"",
			labels,
			nil,
		),
		nvmeBadSystemNandBlocksNormalized: prometheus.NewDesc(
			"nvme_bad_system_nand_blocks_normalized",
			"",
			labels,
			nil,
		),
		nvmeXorRecoveryCount: prometheus.NewDesc(
			"nvme_xor_recovery_count",
			"",
			labels,
			nil,
		),
		nvmeUncorrectableReadErrorCount: prometheus.NewDesc(
			"nvme_uncorrectable_uead_error_count",
			"",
			labels,
			nil,
		),
		nvmeSoftEccErrorCount: prometheus.NewDesc(
			"nvme_soft_ecc_error_count",
			"",
			labels,
			nil,
		),
		nvmeEndToEndDetectedErrors: prometheus.NewDesc(
			"nvme_end_to_end_detected_errors",
			"",
			labels,
			nil,
		),
		nvmeEndToEndCorrectedErrors: prometheus.NewDesc(
			"nvme_end_to_end_corrected_errors",
			"",
			labels,
			nil,
		),
		nvmeSystemDataPercentUsed: prometheus.NewDesc(
			"nvme_system_data_percent_used",
			"",
			labels,
			nil,
		),
		nvmeRefreshCounts: prometheus.NewDesc(
			"nvme_refresh_counts",
			"",
			labels,
			nil,
		),
		nvmeMaxUserDataEraseCounts: prometheus.NewDesc(
			"nvme_max_user_data_erase_counts",
			"",
			labels,
			nil,
		),
		nvmeMinUserDataEraseCounts: prometheus.NewDesc(
			"nvme_min_user_data_erase_counts",
			"",
			labels,
			nil,
		),
		nvmeNumberOfThermalThrottlingEvents: prometheus.NewDesc(
			"nvme_number_of_thermal_throttling_events",
			"",
			labels,
			nil,
		),
		nvmeCurrentThrottlingStatus: prometheus.NewDesc(
			"nvme_current_throttling_status",
			"",
			labels,
			nil,
		),
		nvmePcieCorrectableErrorCount: prometheus.NewDesc(
			"nvme_pcie_correctable_error_count",
			"",
			labels,
			nil,
		),
		nvmeIncompleteShutdowns: prometheus.NewDesc(
			"nvme_incomplete_shutdowns",
			"",
			labels,
			nil,
		),
		nvmePercentFreeBlocks: prometheus.NewDesc(
			"nvme_percent_free_blocks",
			"",
			labels,
			nil,
		),
		nvmeCapacitorHealth: prometheus.NewDesc(
			"nvme_capacitor_health",
			"",
			labels,
			nil,
		),
		nvmeUnalignedIo: prometheus.NewDesc(
			"nvme_unaligned_io",
			"",
			labels,
			nil,
		),
		nvmeSecurityVersionNumber: prometheus.NewDesc(
			"nvme_security_version_number",
			"",
			labels,
			nil,
		),
		nvmeNuseNamespaceUtilization: prometheus.NewDesc(
			"nvme_nuse_namespace_utilization",
			"",
			labels,
			nil,
		),
		nvmePlpStartCount: prometheus.NewDesc(
			"nvme_plp_start_count",
			"",
			labels,
			nil,
		),
		nvmeEnduranceEstimate: prometheus.NewDesc(
			"nvme_endurance_estimate",
			"",
			labels,
			nil,
		),
		nvmeLogPageVersion: prometheus.NewDesc(
			"nvme_log_page_version",
			"",
			labels,
			nil,
		),
		nvmeLogPageGUID: prometheus.NewDesc(
			"nvme_log_page_guid",
			"",
			labels,
			nil,
		),
		nvmeErrataVersionField: prometheus.NewDesc(
			"nvme_errata_version_field",
			"",
			labels,
			nil,
		),
		nvmePointVersionField: prometheus.NewDesc(
			"nvme_point_version_field",
			"",
			labels,
			nil,
		),
		nvmeMinorVersionField: prometheus.NewDesc(
			"nvme_minor_version_field",
			"",
			labels,
			nil,
		),
		nvmeMajorVersionField: prometheus.NewDesc(
			"nvme_major_version_field",
			"",
			labels,
			nil,
		),
		nvmeNvmeErrataVersion: prometheus.NewDesc(
			"nvme_nvme_errata_version",
			"",
			labels,
			nil,
		),
		nvmePcieLinkRetrainingCount: prometheus.NewDesc(
			"nvme_pcie_link_retraining_count",
			"",
			labels,
			nil,
		),
		nvmePowerStateChangeCount: prometheus.NewDesc(
			"nvme_power_state_change_count",
			"",
			labels,
			nil,
		),
		nvmeNameSpace: prometheus.NewDesc(
			"nvme_namespace",
			"",
			infoLabels,
			nil,
		),
		nvmeUsedBytes: prometheus.NewDesc(
			"nvme_used_bytes",
			"",
			infoLabels,
			nil,
		),
		nvmeMaximumLba: prometheus.NewDesc(
			"nvme_maximum_lba",
			"",
			infoLabels,
			nil,
		),
		nvmePhysicalSize: prometheus.NewDesc(
			"nvme_physical_size",
			"",
			infoLabels,
			nil,
		),
		nvmeSectorSize: prometheus.NewDesc(
			"nvme_sector_size",
			"",
			infoLabels,
			nil,
		),
	}
}

func (c *nvmeCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.nvmeCriticalWarning
	ch <- c.nvmeTemperature
	ch <- c.nvmeAvailSpare
	ch <- c.nvmeSpareThresh
	ch <- c.nvmePercentUsed
	ch <- c.nvmeEnduranceGrpCriticalWarningSummary
	ch <- c.nvmeDataUnitsRead
	ch <- c.nvmeDataUnitsWritten
	ch <- c.nvmeHostReadCommands
	ch <- c.nvmeHostWriteCommands
	ch <- c.nvmeControllerBusyTime
	ch <- c.nvmePowerCycles
	ch <- c.nvmePowerOnHours
	ch <- c.nvmeUnsafeShutdowns
	ch <- c.nvmeMediaErrors
	ch <- c.nvmeNumErrLogEntries
	ch <- c.nvmeWarningTempTime
	ch <- c.nvmeCriticalCompTime
	ch <- c.nvmeThmTemp1TransCount
	ch <- c.nvmeThmTemp2TransCount
	ch <- c.nvmeThmTemp1TotalTime
	ch <- c.nvmeThmTemp2TotalTime
	ch <- c.nvmePhysicalMediaUnitsWrittenHi
	ch <- c.nvmePhysicalMediaUnitsWrittenLo
	ch <- c.nvmePhysicalMediaUnitsReadHi
	ch <- c.nvmePhysicalMediaUnitsReadLo
	ch <- c.nvmeBadUserNandBlocksRaw
	ch <- c.nvmeBadUserNandBlocksNormalized
	ch <- c.nvmeBadSystemNandBlocksRaw
	ch <- c.nvmeBadSystemNandBlocksNormalized
	ch <- c.nvmeXorRecoveryCount
	ch <- c.nvmeUncorrectableReadErrorCount
	ch <- c.nvmeSoftEccErrorCount
	ch <- c.nvmeEndToEndDetectedErrors
	ch <- c.nvmeEndToEndCorrectedErrors
	ch <- c.nvmeSystemDataPercentUsed
	ch <- c.nvmeRefreshCounts
	ch <- c.nvmeMaxUserDataEraseCounts
	ch <- c.nvmeMinUserDataEraseCounts
	ch <- c.nvmeNumberOfThermalThrottlingEvents
	ch <- c.nvmeCurrentThrottlingStatus
	ch <- c.nvmePcieCorrectableErrorCount
	ch <- c.nvmeIncompleteShutdowns
	ch <- c.nvmePercentFreeBlocks
	ch <- c.nvmeCapacitorHealth
	ch <- c.nvmeUnalignedIo
	ch <- c.nvmeSecurityVersionNumber
	ch <- c.nvmeNuseNamespaceUtilization
	ch <- c.nvmePlpStartCount
	ch <- c.nvmeEnduranceEstimate
	ch <- c.nvmeLogPageVersion
	ch <- c.nvmeLogPageGUID
	ch <- c.nvmeErrataVersionField
	ch <- c.nvmePointVersionField
	ch <- c.nvmeMinorVersionField
	ch <- c.nvmeMajorVersionField
	ch <- c.nvmeNvmeErrataVersion
	ch <- c.nvmePcieLinkRetrainingCount
	ch <- c.nvmePowerStateChangeCount
	ch <- c.nvmeNameSpace
	ch <- c.nvmeUsedBytes
	ch <- c.nvmeMaximumLba
	ch <- c.nvmePhysicalSize
	ch <- c.nvmeSectorSize
}

func executeCommand(cmd string, args ...string) ([]byte, error) {
	command := exec.Command(cmd, args...)

	output, err := command.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error running %s command: %w, output: %s", cmd, err, string(output))
	}

	if !gjson.Valid(string(output)) {
		return nil, fmt.Errorf("invalid JSON output from %s command: %s", cmd, string(output))
	}

	return output, nil
}

func (c *nvmeCollector) Collect(ch chan<- prometheus.Metric) {
	nvmeDeviceList := c.getDeviceList()
	for _, nvmeDevice := range nvmeDeviceList {
		c.sendInfoMetrics(ch, nvmeDevice)
		nvmeDevicePath := nvmeDevice.Get("DevicePath")
		c.collectSmartLogMetrics(ch, nvmeDevicePath)

		if c.ocp {
			c.collectOcpSmartLogMetrics(ch, nvmeDevicePath)
		}
	}
}

func (c *nvmeCollector) getDeviceList() []gjson.Result {
	nvmeDeviceCmd, err := executeCommand("nvme", "list", "-o", "json")
	if err != nil {
		log.Printf("Error running nvme list -o json: %s\n", err)
	}

	// return gjson.Get(string(nvmeDeviceCmd), "Devices.#.DevicePath").Array()
	return gjson.Get(string(nvmeDeviceCmd), "Devices").Array()
}

func (c *nvmeCollector) collectSmartLogMetrics(ch chan<- prometheus.Metric, device gjson.Result) {
	nvmeSmartLog, err := executeCommand("nvme", "smart-log", device.String(), "-o", "json")
	if err != nil {
		log.Printf("Error running smart-log %s -o json: %s\n", device.String(), err)
	}

	nvmeSmartLogMetrics := gjson.GetMany(string(nvmeSmartLog),
		"critical_warning",
		"temperature",
		"avail_spare",
		"spare_thresh",
		"percent_used",
		"endurance_grp_critical_warning_summary",
		"data_units_read",
		"data_units_written",
		"host_read_commands",
		"host_write_commands",
		"controller_busy_time",
		"power_cycles",
		"power_on_hours",
		"unsafe_shutdowns",
		"media_errors",
		"num_err_log_entries",
		"warning_temp_time",
		"critical_comp_time",
		"thm_temp1_trans_count",
		"thm_temp2_trans_count",
		"thm_temp1_total_time",
		"thm_temp2_total_time")
	c.sendSmartLogMetrics(ch, nvmeSmartLogMetrics, device.String())
}

func (c *nvmeCollector) collectOcpSmartLogMetrics(ch chan<- prometheus.Metric, device gjson.Result) {
	nvmeOcpSmartLog, err := executeCommand("nvme", "ocp", "smart-add-log", device.String(), "-o", "json")
	if err != nil {
		log.Printf("Error running smart-add-log %s -o json: %s\n", device.String(), err)
	}

	nvmeOcpSmartLogMetrics := gjson.GetMany(string(nvmeOcpSmartLog),
		"Physical media units written.hi",
		"Physical media units written.lo",
		"Physical media units read.hi",
		"Physical media units read.lo",
		"Bad user nand blocks - Raw",
		"Bad user nand blocks - Normalized",
		"Bad system nand blocks - Raw",
		"Bad system nand blocks - Normalized",
		"XOR recovery count",
		"Uncorrectable read error count",
		"Soft ecc error count",
		"End to end detected errors",
		"End to end corrected errors",
		"System data percent used",
		"Refresh counts",
		"Max User data erase counts",
		"Min User data erase counts",
		"Number of Thermal throttling events",
		"Current throttling status",
		"PCIe correctable error count",
		"Incomplete shutdowns",
		"Percent free blocks",
		"Capacitor health",
		"Unaligned I/O",
		"Security Version Number",
		"NUSE - Namespace utilization",
		"PLP start count",
		"Endurance estimate",
		"Log page version",
		"Log page GUID",
		"Errata Version Field",
		"Point Version Field",
		"Minor Version Field",
		"Major Version Field",
		"NVMe Errata Version",
		"PCIe Link Retraining Count",
		"Power State Change Count")
	c.sendOcpSmartLogMetrics(ch, nvmeOcpSmartLogMetrics, device.String())
}

func (c *nvmeCollector) sendInfoMetrics(ch chan<- prometheus.Metric, device gjson.Result) {
	nameSpace := device.Get("NameSpace").Float()
	devicePath := device.Get("DevicePath").String()
	genericPath := device.Get("GenericPath").String()
	firmware := device.Get("Firmware").String()
	modelNumber := device.Get("ModelNumber").String()
	serialNumber := device.Get("SerialNumber").String()
	usedBytes := device.Get("UsedBytes").Float()
	maximumLba := device.Get("MaximumLBA").Float()
	physicalSize := device.Get("PhysicalSize").Float()
	sectorSize := device.Get("SectorSize").Float()
	ch <- prometheus.MustNewConstMetric(
		c.nvmeNameSpace, prometheus.GaugeValue, nameSpace, devicePath, genericPath, firmware, modelNumber, serialNumber)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeUsedBytes, prometheus.GaugeValue, usedBytes, devicePath, genericPath, firmware, modelNumber, serialNumber)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeMaximumLba, prometheus.GaugeValue, maximumLba, devicePath, genericPath, firmware, modelNumber, serialNumber)
	ch <- prometheus.MustNewConstMetric(
		c.nvmePhysicalSize, prometheus.GaugeValue, physicalSize, devicePath, genericPath, firmware, modelNumber, serialNumber)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeSectorSize, prometheus.GaugeValue, sectorSize, devicePath, genericPath, firmware, modelNumber, serialNumber)
}

func (c *nvmeCollector) sendSmartLogMetrics(ch chan<- prometheus.Metric, metrics []gjson.Result, device string) {
	ch <- prometheus.MustNewConstMetric(
		c.nvmeCriticalWarning, prometheus.GaugeValue, metrics[0].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeTemperature, prometheus.GaugeValue, metrics[1].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeAvailSpare, prometheus.GaugeValue, metrics[2].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeSpareThresh, prometheus.GaugeValue, metrics[3].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmePercentUsed, prometheus.GaugeValue, metrics[4].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeEnduranceGrpCriticalWarningSummary, prometheus.GaugeValue, metrics[5].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeDataUnitsRead, prometheus.CounterValue, metrics[6].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeDataUnitsWritten, prometheus.CounterValue, metrics[7].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeHostReadCommands, prometheus.CounterValue, metrics[8].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeHostWriteCommands, prometheus.CounterValue, metrics[9].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeControllerBusyTime, prometheus.CounterValue, metrics[10].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmePowerCycles, prometheus.CounterValue, metrics[11].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmePowerOnHours, prometheus.CounterValue, metrics[12].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeUnsafeShutdowns, prometheus.CounterValue, metrics[13].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeMediaErrors, prometheus.CounterValue, metrics[14].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeNumErrLogEntries, prometheus.CounterValue, metrics[15].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeWarningTempTime, prometheus.CounterValue, metrics[16].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeCriticalCompTime, prometheus.CounterValue, metrics[17].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeThmTemp1TransCount, prometheus.CounterValue, metrics[18].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeThmTemp2TransCount, prometheus.CounterValue, metrics[19].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeThmTemp1TotalTime, prometheus.CounterValue, metrics[20].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeThmTemp2TotalTime, prometheus.CounterValue, metrics[21].Float(), device)
}

func (c *nvmeCollector) sendOcpSmartLogMetrics(ch chan<- prometheus.Metric, metrics []gjson.Result, device string) {
	ch <- prometheus.MustNewConstMetric(
		c.nvmePhysicalMediaUnitsWrittenHi, prometheus.CounterValue, metrics[0].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmePhysicalMediaUnitsWrittenLo, prometheus.CounterValue, metrics[1].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmePhysicalMediaUnitsReadHi, prometheus.CounterValue, metrics[2].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmePhysicalMediaUnitsReadLo, prometheus.CounterValue, metrics[3].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeBadUserNandBlocksRaw, prometheus.CounterValue, metrics[4].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeBadUserNandBlocksNormalized, prometheus.CounterValue, metrics[5].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeBadSystemNandBlocksRaw, prometheus.CounterValue, metrics[6].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeBadSystemNandBlocksNormalized, prometheus.CounterValue, metrics[7].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeXorRecoveryCount, prometheus.CounterValue, metrics[8].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeUncorrectableReadErrorCount, prometheus.CounterValue, metrics[9].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeSoftEccErrorCount, prometheus.CounterValue, metrics[10].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeEndToEndDetectedErrors, prometheus.CounterValue, metrics[11].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeEndToEndCorrectedErrors, prometheus.CounterValue, metrics[12].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeSystemDataPercentUsed, prometheus.GaugeValue, metrics[13].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeRefreshCounts, prometheus.CounterValue, metrics[14].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeMaxUserDataEraseCounts, prometheus.CounterValue, metrics[15].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeMinUserDataEraseCounts, prometheus.CounterValue, metrics[16].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeNumberOfThermalThrottlingEvents, prometheus.CounterValue, metrics[17].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeCurrentThrottlingStatus, prometheus.GaugeValue, metrics[18].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmePcieCorrectableErrorCount, prometheus.CounterValue, metrics[19].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeIncompleteShutdowns, prometheus.CounterValue, metrics[20].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmePercentFreeBlocks, prometheus.GaugeValue, metrics[21].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeCapacitorHealth, prometheus.GaugeValue, metrics[22].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeUnalignedIo, prometheus.CounterValue, metrics[23].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeSecurityVersionNumber, prometheus.GaugeValue, metrics[24].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeNuseNamespaceUtilization, prometheus.GaugeValue, metrics[25].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmePlpStartCount, prometheus.CounterValue, metrics[26].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeEnduranceEstimate, prometheus.GaugeValue, metrics[27].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeLogPageVersion, prometheus.GaugeValue, metrics[28].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeLogPageGUID, prometheus.GaugeValue, metrics[29].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeErrataVersionField, prometheus.GaugeValue, metrics[30].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmePointVersionField, prometheus.GaugeValue, metrics[31].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeMinorVersionField, prometheus.GaugeValue, metrics[32].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeMajorVersionField, prometheus.GaugeValue, metrics[33].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmeNvmeErrataVersion, prometheus.GaugeValue, metrics[34].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmePcieLinkRetrainingCount, prometheus.CounterValue, metrics[35].Float(), device)
	ch <- prometheus.MustNewConstMetric(
		c.nvmePowerStateChangeCount, prometheus.CounterValue, metrics[36].Float(), device)
}

func main() {
	flag.Usage = func() {
		fmt.Println("nvme_exporter - Exports NVMe smart-log and smart-ocp-log metrics in Prometheus format")
		fmt.Println("Validated with nvme smart-log field descriptions can be found on page 209 of:")
		fmt.Println(
			"https://nvmexpress.org/wp-content/uploads/NVM-Express-Base-Specification-Revision-2.1-2024.08.05-Ratified.pdf")
		fmt.Println("Validated with nvme ocp-smart-log field descriptions can be found on page 24 of:")
		fmt.Println("https://www.opencompute.org/documents/datacenter-nvme-ssd-specification-v2-5-pdf */")
		fmt.Printf("It has been tested with nvme-cli versions:%v\n", _supportedVersions)
		fmt.Println("Usage: nvme_exporter [options]")
		flag.PrintDefaults()
	}
	port := flag.String("port", "9998", "port to listen on")
	ocp := flag.Bool("ocp", false, "Enable OCP smart log metrics")
	endpoint := flag.String("endpoint", "/metrics", "Specify the endpoint to expose metrics")
	flag.Parse()

	if !strings.HasPrefix(*endpoint, "/") {
		*endpoint = "/" + *endpoint
	}

	// check user
	currentUser, err := user.Current()
	if err != nil {
		log.Fatalf("Error getting current user  %s\n", err)
	}

	if currentUser.Username != "root" {
		log.Fatalln("Error: you must be root to use nvme-cli")
	}

	// check for nvme-cli executable
	_, err = exec.LookPath("nvme")
	if err != nil {
		log.Fatalf("Cannot find NVMe cli command in path: %s\n", err)
	}
	// check for nvme-cli version
	command := exec.Command("nvme", "--version")

	out, err := command.CombinedOutput()
	if err != nil {
		log.Fatalf("error running nvme --version command: %s", err)
	}

	re := regexp.MustCompile(`nvme version (\d+\.\d+)\.\d+`)
	match := re.FindStringSubmatch(string(out))

	if match != nil {
		version := match[1]
		if !isSupportedVersion(version) {
			log.Printf("NVMe cli version %s not supported, supported versions are: %v", version, _supportedVersions)
		}
	} else {
		log.Fatalf("Unable to find NVMe CLI version in output: %s", string(out))
	}

	prometheus.MustRegister(newNvmeCollector(*ocp))
	http.Handle(*endpoint, promhttp.Handler())
	log.Printf("Starting newNvmeCollector on port: %s, metrics endpoint: %s\n", *port, *endpoint)
	log.Printf("newNvmeCollector is collecting OCP smart-log metrics: %t\n", *ocp)

	server := &http.Server{
		Addr:              ":" + *port,
		ReadHeaderTimeout: 3 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}
