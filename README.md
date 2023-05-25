# vitastor-exporter

Prometheus exporter for Vitastor - Open Source Software-Defined Storage [vitastor.io](https://vitastor.io)

## Installation and Usage

To run the exporter, you need to run the binary somewhere where etcd is available (for example, on the monitor node).

If you run the exporter on a node where the `vitastor.conf` config is available in the standard path (`/etc/vitastor/vitastor.conf`), just run the binary without any parameters, it will take the necessary parameters from there:

```bash
user@host bin % vitastor-exporter
```

If you want to run it without vitastor.conf, you can pass parameters to excutable:

```bash
user@host bin % vitastor-exporter --help
Usage of ./vitastor-exporter:
  -etcd-url string
        Comma-separated list of etcd urls. WARNING: setting that param will override --vitastor-conf. Default: empty
  -metrics-path string
        Path to expose metrics. Default: /metrics (default "/metrics")
  -port int
        Port to expose metrics. Default: 8080 (default 8080)
  -vitastor-conf string
        Path to vitastor.conf (to obtain etcd connection params). Default: /etc/vitastor/vitastor.conf (default "/etc/vitastor/vitastor.conf")
  -vitastor-prefix string
        Etcd tree prefix for Vitastor cluster info. Default: /vitastor (default "/vitastor")
```