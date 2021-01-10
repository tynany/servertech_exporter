package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"github.com/tynany/servertech_exporter/collector"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	listenAddress = kingpin.Flag("web.listen-address", "Address on which to expose metrics and web interface.").Default(":9778").String()
	telemetryPath = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
	httpOnly      = kingpin.Flag("web.http", "Run in HTTP mode.").Default("False").Bool()
	sslCrt        = kingpin.Flag("web.certificate", "Path to SSL certificate.").String()
	sslKey        = kingpin.Flag("web.key", "Path to SSL certificate key.").String()
)

func handler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	user := r.URL.Query().Get("user")
	pass := r.URL.Query().Get("pass")
	if target == "" {
		http.Error(w, "'target' parameter must be specified", 400)
		return
	}

	registry := prometheus.NewRegistry()

	registry.Register(collector.NewExporter(target, user, pass))

	gatheres := prometheus.Gatherers{
		prometheus.DefaultGatherer,
		registry,
	}
	handlerOpts := promhttp.HandlerOpts{
		ErrorLog:      log.NewErrorLogger(),
		ErrorHandling: promhttp.ContinueOnError,
	}
	promhttp.HandlerFor(gatheres, handlerOpts).ServeHTTP(w, r)
}

func parseCLI() {
	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("servertech_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	if !*httpOnly {
		if *sslCrt == "" || *sslKey == "" {
			log.Fatal("HTTPS mode selected but SSL certificate and key not specified")
		}
	}
}

func main() {
	prometheus.MustRegister(version.NewCollector("servertech_exporter"))

	parseCLI()

	log.Infof("Starting servertech_exporter %s on %s", version.Info(), *listenAddress)

	http.HandleFunc(*telemetryPath, handler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>ServerTech Exporter</title></head>
			<body>
			<h1>ServerTech Exporter</h1>
			<p><a href="` + *telemetryPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	if *httpOnly {
		if err := http.ListenAndServe(*listenAddress, nil); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := http.ListenAndServeTLS(*listenAddress, *sslCrt, *sslKey, nil); err != nil {
			log.Fatal(err)
		}
	}
}
