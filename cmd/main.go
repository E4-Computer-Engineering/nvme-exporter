package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/E4-Computer-Engineering/nvme_exporter/pkg"
	"github.com/E4-Computer-Engineering/nvme_exporter/pkg/utils"
)

const _minimumSupportedVersion = "2.8"

var (
	validationState = struct {
		sync.RWMutex

		isValid      bool
		errorMessage string
	}{
		isValid: true,
	}

	scrapeFailuresTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "nvme_exporter_scrape_failures_total",
			Help: "Total number of scrape failures due to validation errors " +
				"(not root, nvme-cli not found, or unsupported version)",
		},
	)
)

// Collector represents a metric collector with enable/disable capability.
type Collector struct {
	name         string
	defaultState bool
	enabled      *bool
	description  string
}

var (
	collectors = map[string]*Collector{
		"smart": {
			name:         "smart",
			defaultState: true,
			description:  "NVMe SMART log metrics",
		},
		"info": {
			name:         "info",
			defaultState: true,
			description:  "NVMe device info metrics",
		},
		"ocp": {
			name:         "ocp",
			defaultState: true,
			description:  "NVMe OCP (Open Compute Project) SMART log metrics",
		},
	}

	disableDefaultCollectors = flag.Bool(
		"collector.disable-defaults",
		false,
		"Disable all default collectors",
	)
)

func isSupportedVersion(version string) bool {
	versionParts := strings.Split(version, ".")
	minVersionParts := strings.Split(_minimumSupportedVersion, ".")

	if len(versionParts) < 2 || len(minVersionParts) < 2 {
		return false
	}

	vMajor, err1 := strconv.Atoi(versionParts[0])
	vMinor, err2 := strconv.Atoi(versionParts[1])
	minMajor, err3 := strconv.Atoi(minVersionParts[0])
	minMinor, err4 := strconv.Atoi(minVersionParts[1])

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return false
	}

	if vMajor > minMajor {
		return true
	}

	if vMajor < minMajor {
		return false
	}

	return vMinor >= minMinor
}

func setValidationError(msg string) {
	validationState.Lock()
	defer validationState.Unlock()

	validationState.isValid = false
	validationState.errorMessage = msg
}

func isValidationValid() bool {
	validationState.RLock()
	defer validationState.RUnlock()

	return validationState.isValid
}

func initCollectorFlags() {
	// Register flags for each collector
	for name, collector := range collectors {
		flagName := "collector." + name
		noFlagName := "no-collector." + name

		defaultValue := collector.defaultState

		// --collector.X flag to enable
		enableFlag := flag.Bool(
			flagName,
			defaultValue,
			fmt.Sprintf("Enable the %s collector (default: %t)", collector.description, defaultValue),
		)

		// --no-collector.X flag to disable
		disableFlag := flag.Bool(
			noFlagName,
			false,
			fmt.Sprintf("Disable the %s collector", collector.description),
		)

		collector.enabled = enableFlag

		// Store both flags so we can resolve them later
		collectors[name] = collector

		// The disable flag is handled in resolveCollectorStates
		_ = disableFlag
	}
}

func resolveCollectorStates() map[string]bool {
	states := make(map[string]bool)

	for name, collector := range collectors {
		// Start with default state
		enabled := collector.defaultState

		// If disable-defaults is set, start with false
		if *disableDefaultCollectors {
			enabled = false
		}

		// Check if explicit enable flag was set
		enableFlagName := "collector." + name
		disableFlagName := "no-collector." + name

		// Check if the flag was explicitly set
		explicitlyEnabled := false
		explicitlyDisabled := false

		flag.Visit(func(flagItem *flag.Flag) {
			if flagItem.Name == enableFlagName {
				explicitlyEnabled = true
				enabled = *collector.enabled
			}

			if flagItem.Name == disableFlagName {
				explicitlyDisabled = true
			}
		})

		// Disable flag takes precedence
		if explicitlyDisabled {
			enabled = false
		} else if explicitlyEnabled {
			enabled = true
		}

		states[name] = enabled
	}

	return states
}

func printUsage() {
	fmt.Println("nvme_exporter - Prometheus exporter for NVMe device metrics")
	fmt.Println("\nExports NVMe SMART log and OCP SMART log metrics in Prometheus format.")
	fmt.Println("\nDocumentation:")
	fmt.Println("  NVMe SMART log specification (page 209):")
	fmt.Println("    https://nvmexpress.org/wp-content/uploads/" +
		"NVM-Express-Base-Specification-Revision-2.1-2024.08.05-Ratified.pdf")
	fmt.Println("  OCP SMART log specification (page 24):")
	fmt.Println("    https://www.opencompute.org/documents/datacenter-nvme-ssd-specification-v2-5-pdf")
	fmt.Printf("\nMinimum supported nvme-cli version: %s\n", _minimumSupportedVersion)
	fmt.Println("\nUsage: nvme_exporter [options]")
	fmt.Println("\nWeb server options:")
	fmt.Println("  --web.listen-address string")
	fmt.Println("        Address on which to expose metrics and web interface (default \":9998\")")
	fmt.Println("  --web.telemetry-path string")
	fmt.Println("        Path under which to expose metrics (default \"/metrics\")")
	fmt.Println("\nCollector options:")
	fmt.Println("  --collector.<name>")
	fmt.Println("        Enable the specified collector (enabled by default)")
	fmt.Println("  --no-collector.<name>")
	fmt.Println("        Disable the specified collector")
	fmt.Println("  --collector.disable-defaults")
	fmt.Println("        Disable all default collectors")
	fmt.Println("\nAvailable collectors:")

	for name, collector := range collectors {
		defaultStr := ""
		if collector.defaultState {
			defaultStr = " (enabled by default)"
		}

		fmt.Printf("  %-10s %s%s\n", name, collector.description, defaultStr)
	}

	fmt.Println("\nExamples:")
	fmt.Println("  # Start with all default collectors on default port")
	fmt.Println("  nvme_exporter")
	fmt.Println("\n  # Listen on a specific address and port")
	fmt.Println("  nvme_exporter --web.listen-address=\":9100\"")
	fmt.Println("\n  # Disable OCP metrics collection")
	fmt.Println("  nvme_exporter --no-collector.ocp")
	fmt.Println("\n  # Only collect SMART metrics (disable info and OCP)")
	fmt.Println("  nvme_exporter --collector.disable-defaults --collector.smart")
}

func validatePrerequisites() {
	// Validate current user
	err := utils.CheckCurrentUser("root")
	if err != nil {
		log.Printf("WARNING: current user is not root: %s", err.Error())
		log.Printf("WARNING: exporter will continue running but scrapes will fail")
		setValidationError("not running as root: " + err.Error())

		return
	}

	// Check for nvme-cli version
	validateNVMeCLI()
}

func validateNVMeCLI() {
	out, err := utils.ExecuteCommand("nvme", "--version")
	if err != nil {
		log.Printf("WARNING: nvme binary not found or error executing: %s", err.Error())
		log.Printf("WARNING: exporter will continue running but scrapes will fail")
		setValidationError("nvme binary not available: " + err.Error())

		return
	}

	re := regexp.MustCompile(`nvme version (\d+\.\d+)`)
	match := re.FindStringSubmatch(out)

	if match == nil {
		log.Printf("WARNING: Unable to find NVMe CLI version in output: %s", out)
		log.Printf("WARNING: exporter will continue running but scrapes may fail")
		setValidationError("unable to parse nvme-cli version from output: " + out)

		return
	}

	version := match[1]
	if !isSupportedVersion(version) {
		log.Printf("WARNING: NVMe cli version %s not supported, minimum required version is %s",
			version, _minimumSupportedVersion)
		log.Printf("WARNING: exporter will continue running but scrapes may fail or produce incorrect data")
		setValidationError(fmt.Sprintf("unsupported nvme-cli version %s (minimum: %s)",
			version, _minimumSupportedVersion))

		return
	}

	log.Printf("NVMe cli version %s detected and supported", version)
}

func main() {
	// Initialize collector flags before parsing
	initCollectorFlags()

	flag.Usage = printUsage

	// Define flags following Prometheus node_exporter conventions
	listenAddress := flag.String("web.listen-address", ":9998", "Address on which to expose metrics and web interface")
	metricsPath := flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics")
	flag.Parse()

	// Ensure metrics path starts with /
	if !strings.HasPrefix(*metricsPath, "/") {
		*metricsPath = "/" + *metricsPath
	}

	// Register the scrape failures metric
	prometheus.MustRegister(scrapeFailuresTotal)

	// Validate prerequisites - log errors but don't exit
	validatePrerequisites()

	// Resolve collector states based on flags
	collectorStates := resolveCollectorStates()

	// Log enabled collectors
	log.Printf("Enabled collectors:")

	for name, enabled := range collectorStates {
		if enabled {
			log.Printf("  - %s", name)
		}
	}

	// Set up validation checker and scrape failure incrementer
	pkg.SetValidationChecker(isValidationValid)
	pkg.SetScrapeFailureIncrementer(func() {
		scrapeFailuresTotal.Inc()
	})

	prometheus.MustRegister(newNvmeCollector(collectorStates))
	http.Handle(*metricsPath, promhttp.Handler())

	// Add a landing page like node_exporter
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/" {
			http.NotFound(writer, request)

			return
		}

		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(writer, `<html>
<head><title>NVMe Exporter</title></head>
<body>
<h1>NVMe Exporter</h1>
<p><a href="%s">Metrics</a></p>
</body>
</html>`, *metricsPath)
	})

	log.Printf("Starting nvme_exporter on %s", *listenAddress)
	log.Printf("Metrics path: %s", *metricsPath)

	server := &http.Server{
		Addr:              *listenAddress,
		ReadHeaderTimeout: 3 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}
