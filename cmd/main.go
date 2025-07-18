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

	"github.com/E4-Computer-Engineering/nvme_exporter/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var _supportedVersions = map[string]bool{
	"2.9":  true,
	"2.10": true,
	"2.11": true,
}

func isSupportedVersion(version string) bool {
	_, ok := _supportedVersions[version]

	return ok
}

func checkNvmeVersion() error {
	validator := func(out string) bool {
		re := regexp.MustCompile(`nvme version (\d+\.\d+)\.\d+`)
		match := re.FindStringSubmatch(string(out))

		if match != nil {
			version := match[1]
			return isSupportedVersion(version)
		}
		return false
	}
	onError := func(out string) error {
		return fmt.Errorf("NVMe cli version not supported, supported versions are: %v", _supportedVersions)
	}

	shell := utils.NewShell(
		utils.WithValidators(validator),
		utils.WithOnValidationError(onError),
	)

	_, err := shell.Run("nvme", "--version")
	return err
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

	if err = checkNvmeVersion(); err != nil {
		log.Fatal(err)
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
