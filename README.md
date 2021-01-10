# ServerTech Exporter
A Prometheus exporter that collects metrics from ServerTech PDUs using the ServerTech JAWS API.

## Getting Started
Start servertech_exporter with valid flags. To then collect the metrics of a PDU, pass the 'target', 'user' and 'pass' parameter to the exporter's web interface. For example, http://exporter:9778/metrics?target=192.168.77.9&user=admn&pass=admn. By default, servertech_exporter runs in HTTPS mode and a valid certificate and key need to be passed using the `--web.certificate` and `--web.key` flags.

To run servertech_exporter:
```
./servertech_exporter [flags]
```

To view available flags:
```
./servertech_exporter -h
usage: servertech_exporter [<flags>]

Flags:
  -h, --help                Show context-sensitive help (also try --help-long and --help-man).
      --servertech.http.timeout="20s"
                            The HTTP timeout when scraping the ServerTech API.
      --collector.branches  Enable the branches collector (default: enabled).
      --collector.cords     Enable the cords collector (default: enabled).
      --collector.lines     Enable the lines collector (default: enabled).
      --collector.ocps      Enable the ocps collector (default: enabled).
      --collector.outlets   Enable the outlets collector (default: enabled).
      --collector.phases    Enable the phases collector (default: enabled).
      --collector.system    Enable the system collector (default: enabled).
      --collector.units     Enable the units collector (default: enabled).
      --web.listen-address=":9778"
                            Address on which to expose metrics and web interface.
      --web.telemetry-path="/metrics"
                            Path under which to expose metrics.
      --web.http            Run in HTTP mode.
      --web.certificate=WEB.CERTIFICATE
                            Path to SSL certificate.
      --web.key=WEB.KEY     Path to SSL certificate key.
      --log.level="info"    Only log messages with the given severity or above. Valid levels: [debug, info, warn, error, fatal]
      --log.format="logger:stderr"
                            Set the log target and format. Example: "logger:syslog?appname=bob&local=7" or "logger:stdout?json=true"
      --version             Show application version.
```

Promethues configuraiton:
```
scrape_configs:
  - job_name: servertech
    static_configs:
      - targets:
        - device1
    params:
      user: admn
      pass: admn
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - target_label: __address__
        replacement: localhost:9778  # In this example, localhost is running servertech_exporter
```

Docker:
```
docker run --restart unless-stopped -d -p 9778:9778 -v /path/to/server.crt:/server/crt -v /path/to/server.key:/server.key tynany/servertech_exporter
```
The Docker containers expects the SSL certificate be located at /server.crt and the key be located at /server.key.

## ServerTech API 

### Metric Descriptions
Metric descriptions have been taken from [ServerTech's JAWS API Documentation](https://cdn10.servertech.com/assets/documents/documents/808/original/JSON_API_Web_Service_%28JAWS%29_V1.01.pdf?1562965069).