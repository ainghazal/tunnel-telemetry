# Open Tunnel Telemetry

A PoC implementation for a Open Tunnel Telemetry collector and client. See the [spec].

## Build Server

This project requires `go1.21` or higher.

```
go install github.com/ainghazal/tunnel-telemetry/cmd/tt-server@latest
```

## Concepts

* **Client**: is an application client that generates connectivity reports; and sends them to its configured collector.
* **Endpoint**: the circumvention proxy that the client attempts to connect to.
* **Collector**: a HTTPS endpoint that can receive client reports. Trust is important, because the collector sees the client IPs, and it knows the endpoint IPs.
* **Scrubbing**: the collector can be configured to scrub potentially sensitive information (client and endpoint IPs).
* **Relay**: after scrubbing, the collector can relay reports to a secondary
  collector. The secondary collector receives individual reports (or
  aggregates) by the primary collector, and is able to run data processing
  pipelines with a broader context than individual collectors. In this case,
  OONI acts as a secondary collector.

## Running

```
tt-server
```

This will run server in port `:8080`.

## Sending a report

A minimal report is a json containing three mandatory fields:

* `report-type`: MUST be `tunnel-telemetry`.
* `t`: MUST be the timestamp for the observation contained in the report (TODO: standardize if start or end time). The collector will not process reports sent from too far in the future or the past.
* `endpoint`: MUST be the endpoint that the client attempted to connect to, in the format `protocol://ip_address:port`.

A few optional fields are also understood:

* `config`: an arbitrary `map[str]str` containing relevant configurations used in the connection. Sensitive information should not be sent here.
* `failure`: in the form `{"op": "operation.detail", "error": "error message"}`, or `null`. A missing `failure` field is understood as a successful connection.
* `uuid`: the client can add an `uuid`. If empty, one will be generated.

```bash
curl -X POST \
    -H 'Content-Type: application/json' \
    -H 'X-Forwarded-For: 2.3.4.5' \
    -d '{"report-type": "tunnel-telemetry", "t": "2024-04-12T00:00:00Z", "endpoint": "ss://1.1.1.1:443", "config": {"prefix": "asdf"}}' \
    http://localhost:8080/report
```

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
clients report to the collector using their public IP.  Obviously, this point
will need to be reconsidered, because from a privacy perspective it does not
make sense to abandon the tunnel to submit a report.

For the time being, geolocation in the `tunnel-telemetry` server only works when listening directly on a port exposed to the internet.

For working behind proxies, the right setting must be configured in the instantiation of the echo server (TBD).
