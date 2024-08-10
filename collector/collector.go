package collector

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const (
	// The namespace used by all metrics.
	namespace = "servertech"

	enabledByDefault  = true
	disabledByDefault = false
)

var (
	servertechTotalScrapeCount = 0.0

	servertechLabels = []string{"collector"}
	servertechDesc   = map[string]*prometheus.Desc{
		"scrapesTotal":   promDesc("scrapes_total", "Total number of times servertech_exporter has been scraped.", nil),
		"scrapeErrTotal": promDesc("scrape_errors_total", "Total number of errors from a collector.", servertechLabels),
		"scrapeDuration": promDesc("scrape_duration_seconds", "Time it took for a collector's scrape to complete.", servertechLabels),
		"collectorUp":    promDesc("collector_up", "Whether the collector's last scrape was successful (1 = successful, 0 = unsuccessful).", servertechLabels),
	}

	allCollectors  = make(map[string]func() Collector)
	collectorState = make(map[string]*bool)
	httpTimeout    = kingpin.Flag("servertech.http.timeout", "The HTTP timeout when scraping the ServerTech API.").Default("20s").String()
)

func registerCollector(name string, enabledByDefault bool, collector func() Collector) {
	defaultState := "disabled"
	if enabledByDefault {
		defaultState = "enabled"
	}

	allCollectors[name] = collector
	collectorState[name] = kingpin.Flag(fmt.Sprintf("collector.%s", name), fmt.Sprintf("Enable the %s collector (default: %s).", name, defaultState)).Default(strconv.FormatBool(enabledByDefault)).Bool()
}

// Collector is the interface a collector has to implement.
type Collector interface {
	// Gets metrics and sends to the Prometheus.Metric channel.
	Get(ch chan<- prometheus.Metric, target, user, pass string) (float64, error)
}

// Exporter collects all collector metrics, implemented as per the prometheus.Collector interface.
type Exporter struct {
	Collectors map[string]Collector
	Target     string
	User       string
	Pass       string
}

// NewExporter returns a new Exporter.
func NewExporter(target, user, pass string) *Exporter {
	enabledCollectors := make(map[string]Collector)
	for name, collector := range allCollectors {
		if *collectorState[name] {
			enabledCollectors[name] = collector()
		}
	}
	return &Exporter{
		Collectors: enabledCollectors,
		Target:     target,
		User:       user,
		Pass:       pass,
	}
}

// Collect implemented as per the prometheus.Collector interface.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	servertechTotalScrapeCount++
	ch <- prometheus.MustNewConstMetric(servertechDesc["scrapesTotal"], prometheus.CounterValue, servertechTotalScrapeCount)

	wg := &sync.WaitGroup{}
	for name, collector := range e.Collectors {
		wg.Add(1)
		go e.runCollector(ch, name, collector, wg)
	}
	wg.Wait()
}

func (e *Exporter) runCollector(ch chan<- prometheus.Metric, name string, collector Collector, wg *sync.WaitGroup) {
	defer wg.Done()

	startTime := time.Now()
	totalErrors, err := collector.Get(ch, e.Target, e.User, e.Pass)

	ch <- prometheus.MustNewConstMetric(servertechDesc["scrapeDuration"], prometheus.GaugeValue, float64(time.Since(startTime).Seconds()), name)
	ch <- prometheus.MustNewConstMetric(servertechDesc["scrapeErrTotal"], prometheus.GaugeValue, totalErrors, name)

	if err != nil {
		ch <- prometheus.MustNewConstMetric(servertechDesc["collectorUp"], prometheus.GaugeValue, 0, name)
		log.Errorf("collector %q scrape failed: %s", name, err)
	} else {
		ch <- prometheus.MustNewConstMetric(servertechDesc["collectorUp"], prometheus.GaugeValue, 1, name)
	}

}

// Describe implemented as per the prometheus.Collector interface.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range servertechDesc {
		ch <- desc
	}
}

func promDesc(metricName string, metricDescription string, labels []string) *prometheus.Desc {
	return prometheus.NewDesc(namespace+"_"+metricName, metricDescription, labels, nil)
}

func colPromDesc(subsystem string, metricName string, metricDescription string, labels []string) *prometheus.Desc {
	return prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, metricName), metricDescription, labels, nil)
}

func newGauge(ch chan<- prometheus.Metric, descName *prometheus.Desc, metric float64, labels ...string) {
	ch <- prometheus.MustNewConstMetric(descName, prometheus.GaugeValue, metric, labels...)
}

func newCounter(ch chan<- prometheus.Metric, descName *prometheus.Desc, metric float64, labels ...string) {
	ch <- prometheus.MustNewConstMetric(descName, prometheus.CounterValue, metric, labels...)
}

func getServerTechJSON(target, user, pass, path string) ([]byte, error) {
	// todo: work out how to properly handle TLS verification, and avoid possible MITM attacks
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transport}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s/jaws/monitor/%s", target, path), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %v", err)
	}
	basicAuth := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
	req.Header.Add("Authorization", "Basic "+basicAuth)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform http request: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("incorrect status code received from device: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body of request from device: %v", err)
	}

	return body, nil
}

func statusMetric(ch chan<- prometheus.Metric, desc *prometheus.Desc, metric, statusType string, labels []string) {
	status := float64(0)
	if strings.ToLower(metric) == "normal" {
		status = 1
	}
	statusLabels := append(labels, statusType)
	newGauge(ch, desc, status, statusLabels...)
}

func stateMetric(ch chan<- prometheus.Metric, desc *prometheus.Desc, stateStr string, labels []string) {
	state := float64(0)
	if strings.ToLower(stateStr) == "on" {
		state = 1
	}
	newGauge(ch, desc, state, labels...)
}

func reactanceMetric(ch chan<- prometheus.Metric, desc *prometheus.Desc, reactanceStr string, labels []string) {
	reactance := float64(0)
	if strings.ToLower(reactanceStr) == "capacitive" {
		reactance = 1
	} else if strings.ToLower(reactanceStr) == "inductive" {
		reactance = 2
	} else if strings.ToLower(reactanceStr) == "resistive" {
		reactance = 3
	}
	newGauge(ch, desc, reactance, labels...)
}
