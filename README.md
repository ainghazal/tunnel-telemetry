# Open Tunnel Telemetry

A PoC implementation for a Open Tunnel Telemetry collector and client. See the [spec].

## Build Server

```
go get github.com/ainghazal/tunnel-telemetry/cmd/tt-server
```

## Geolocation

For the time being, geolocation in the `tunnel-telemetry` server only works when listening directly on a port exposed to the internet.

(For working behind proxies, the right setting must be configured in the instantiation of the echo server.)


