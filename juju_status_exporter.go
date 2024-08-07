package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/push"
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
			[]string{"charm_name", "charm_rev", "charm_channel", "model_name"}, nil,
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
	for _, app := range collector.jujuStatus.Applications {
		ch <- prometheus.MustNewConstMetric(
			collector.appStatusDesc,
			prometheus.GaugeValue,
			1,
			app.CharmName,
			strconv.Itoa(app.CharmRev),
			app.CharmChannel,
			collector.modelName,
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

func main() {
	var pushGatewayURL string
	var modelName string
	var interval int

	flag.StringVar(&pushGatewayURL, "push_gateway", "", "The Pushgateway URL to push metrics to. If not set, metrics will be exposed on /metrics.")
	flag.StringVar(&modelName, "model_name", "instance_1", "The model name to use as the instance label in the Pushgateway.")
	flag.IntVar(&interval, "interval", 30, "The interval (in seconds) to push metrics to the Pushgateway. Ignored if push_gateway is not set.")
	flag.Parse()

	jujuStatus, err := getJujuStatus()
	if err != nil {
		log.Fatalf("Error getting juju status: %v", err)
	}

	collector := NewJujuCollector(jujuStatus, modelName)

	if pushGatewayURL != "" {
		// Push metrics to the Pushgateway
		registry := prometheus.NewRegistry()
		registry.MustRegister(collector)

		for {
			if err := push.New(pushGatewayURL, "juju_exporter").
				Collector(collector).
				Grouping("job", "juju").
				Grouping("instance", modelName).
				Push(); err != nil {
				log.Printf("Could not push to Pushgateway: %v", err)
			}
			time.Sleep(time.Duration(interval) * time.Second)
		}
	} else {
		// Expose metrics via HTTP
		prometheus.MustRegister(collector)
		http.Handle("/metrics", promhttp.Handler())
		log.Println("Beginning to serve on port :8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}
}
