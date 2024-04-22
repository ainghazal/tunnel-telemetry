# Open Tunnel Telemetry

A PoC implementation for a Open Tunnel Telemetry collector and client. See the [spec].

## Concepts

* **Client**: is an application client that generates connectivity reports; and sends them to its configured collector(s). Equivalent to the **probe** in OONI terminology.
* **Endpoint**: the circumvention proxy that the client attempts to connect to.
* **Collector**: a HTTPS endpoint that can receive client reports. Trust is important, because the collector sees the client IPs, and it knows the endpoint IPs. A client is instructed to send reports to one or more collectors (as backups).
* **Scrubbing**: the collector can be configured to scrub potentially sensitive information (client and endpoint IPs).
* **Relay**: after scrubbing, the collector can relay reports to a **secondary collector**. The secondary collector receives individual reports (or aggregates) by the primary collector, and is able to run data processing pipelines with a broader context than individual collectors. In the case of this implementation of `tunnel-telemetry` collector, OONI acts as a **secondary collector**.

## Server


### Install

```bash
go install github.com/ainghazal/tunnel-telemetry/cmd/tt-server@latest
```

### Run

```bash
tt-server
```

This will run a test `http` server in port `:8080`.

### Autotls

You can enable `autotls` to fetch LetsEncrypt certificates.

```bash
tt-server --autotls --hostname collector.example.org
```

### Help

```bash
tt-server --help
```

### Config

All the configuration flags can be passed as env variables:

```bash
AUTOTLS=true HOSTNAME=collector.example.org tt-server
```

Or as a `yaml` configuration file:


```bash
cat /home/user/ttserver/config.yaml
autotls: true
hostname: collector.asdf.network
```

And then pass that config file to the `tt-server` invocation:

```bash
tt-server --config /home/user/ttserver/config.yaml
```

The default configuration location is `/etc/tunneltelemetry/config.yaml`; any flag or environment variable will take precedence over the options set in the config file.


### Server configuration 

* `autotls`: if true, it will configure LetsEncrypt certificates.
* `autotls-cache-dir`: a dir to cache autotls material (default: "/var/www/.cache").
* `collector-id`: if present, this unique identifier will be added to all reports as an extra annotation. This can be useful to later on query all reports submitted by a given collector.
* `hostname`: the hostname to configure `autotls` certs.
* `listen`: the address to listen on (`:8080` by default; `443` if autotls is used).


## Sending a report

A minimal report is a json containing only three mandatory fields:

* `report-type`: **MUST** be `tunnel-telemetry`.
* `time`: **MUST** be the initial timestamp for the observation contained in the report. The collector will not process reports sent from too far in the future or the past.
* `endpoint`: **MUST** be the endpoint that the client attempted to connect to, in the format `protocol://ip_address:port`.

```bash
$ cat report.json
{
  "report-type": "tunnel-telemetry",
  "time": "2024-04-12T00:00:00Z",
  "endpoint": "ss://1.1.1.1:443",
  "config": {"prefix": "xx"}
}

$ curl -X POST \
    -H 'Content-Type: application/json' \
    -H 'X-Forwarded-For: 2.3.4.5' \
    -d @report.json \
    http://localhost:8080/report
```

### Optional report fields

A few optional fields are also understood:

* `config`: an arbitrary `map[str]str` containing relevant configurations used in the connection. Sensitive information should not be sent here.
* `duration_ms(int)`: a duration, in  milliseconds. This is the delta between the initial time, `time`, and the success or failure indicated by the report.
* `failure`: in the form `{"op": "operation.detail", "msg": "error message", "posix_error": "standard posix error"}`, or `null`. A missing `failure` field is understood as a successful connection.
* `uuid`: the client can add an `uuid`. If empty, one will be generated.


## Viewing a report

Upon a successful processing, and possibly relaying the report, the collector returns a scrubbed report:

```JavaScript
{
  "report-type": "tunnel-telemetry",
  "uuid": "2891c6ff-1b5d-4090-aff9-805d8c4a61c0",
  "ooni-measurement-id": "20240418123955.791083_IT_tunneltelemetry_67c3f38268f4d364",
  "ooni-measurement-link": "https://explorer.ooni.org/m/20240418123955.791083_IT_tunneltelemetry_67c3f38268f4d364",
  "t": "2024-04-12T00:00:00Z",
  "endpoint_port": 443,
  "endpoint_asn": 13335,
  "endpoint_cc": "AU",
  "proto": "ss",
  "config": {
    "prefix": "asdf"
  },
  "client_asn": 3215,
  "client_cc": "FR",
  "sampling_rate": 1
}
```

In this case, the collector was configured to relay reports to the OONI
upstream collector, and it's adding an [OONI Measurement Link](https://explorer.ooni.org/m/20240418123955.791083_IT_tunneltelemetry_67c3f38268f4d364)
where we can share the report in the public OONI Explorer.


## Geolocation

For simplicity, it's assumed that the collector is not blocked, and that
clients report to the collector using their public IP.  It's likely that we'll have 
to reconsider this point, because from a privacy perspective perhaps it does not
make much sense to abandon the tunnel to submit a report. This probably will lead us 
to separate discovery of IP and geolocation itself.

⚠️ For the time being, geolocation in the `tunnel-telemetry` server only works when listening directly on a port exposed to the internet.

For working behind proxies, the right setting must be configured in the instantiation of the echo server (TBD).
