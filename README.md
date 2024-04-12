# tt-collector

A PoC implementation for a Tunnel Telemetry collector.

## Build Server

```
go get github.com/ainghazal/tt-collector/cmd/server
```


## Geolocation

For the time being, geolocation in the `tunnel-telemetry` server only works when listening directly on a port exposed to the internet.

(For working behind proxies, the right setting must be configured in the instantiation of the echo server.)


