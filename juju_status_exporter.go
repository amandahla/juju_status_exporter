package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// JujuStatus is a struct to hold the juju status JSON data
type JujuStatus struct {
	Applications map[string]struct {
		CharmName    string `json:"charm-name"`
		CharmRev     int    `json:"charm-rev"`
		CharmChannel string `json:"charm-channel"`
	} `json:"applications"`
}

// JujuCollector is a struct to implement the Prometheus Collector interface
type JujuCollector struct {
	appStatusDesc *prometheus.Desc
	jujuStatus    *JujuStatus
	modelName     string
}

// NewJujuCollector creates a new JujuCollector
func NewJujuCollector(jujuStatus *JujuStatus, modelName string) *JujuCollector {
	return &JujuCollector{
		appStatusDesc: prometheus.NewDesc(
			"juju_status_applications",
			"Status of Juju applications",
			[]string{"charm_name", "charm_rev", "charm_channel", "model_name", "application_name"}, nil,
		),
		jujuStatus: jujuStatus,
		modelName:  modelName,
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (collector *JujuCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.appStatusDesc
}

// Collect is called by the Prometheus registry when collecting metrics.
func (collector *JujuCollector) Collect(ch chan<- prometheus.Metric) {
	for name, app := range collector.jujuStatus.Applications {
		ch <- prometheus.MustNewConstMetric(
			collector.appStatusDesc,
			prometheus.GaugeValue,
			1,
			app.CharmName,
			strconv.Itoa(app.CharmRev),
			app.CharmChannel,
			collector.modelName,
			name,
		)
	}
}

// getJujuStatus runs the "juju status --format=json" command and parses the output.
func getJujuStatus() (*JujuStatus, error) {
	cmd := exec.Command("juju", "status", "--format=json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var jujuStatus JujuStatus
	err = json.Unmarshal(output, &jujuStatus)
	if err != nil {
		return nil, err
	}

	return &jujuStatus, nil
}

func lookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func lookupEnvOrInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.Atoi(val)
		if err != nil {
			log.Fatalf("LookupEnvOrInt[%s]: %v", key, err)
		}
		return v
	}
	return defaultVal
}

func main() {
	var modelName string
	var port string

	flag.StringVar(&modelName, "model_name", lookupEnvOrString("JUJU_MODEL", "model_name_1"), "The model name to use as the instance label in the Pushgateway.")
	flag.StringVar(&port, "port", lookupEnvOrString("JUJU_STATUS_EXPORTER_PORT", "8080"), "The port on which the exporter will start listening")
	flag.Parse()

	jujuStatus, err := getJujuStatus()
	if err != nil {
		log.Fatalf("Error getting juju status: %v", err)
	}

	// case JUJU_MODEL is admin/model
	modelNameParts := strings.Split(modelName, "/")
	modelNameFmt := modelNameParts[len(modelNameParts)-1]

	collector := NewJujuCollector(jujuStatus, modelNameFmt)

	prometheus.MustRegister(collector)
	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Listening on port: %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
