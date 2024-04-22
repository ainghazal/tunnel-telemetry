#!/bin/sh
curl -X POST \
    -H 'Content-Type: application/json' \
    -H 'X-Forwarded-For: 2.3.4.5' \
    -d '{"report-type": "tunnel-telemetry", "time": "2024-04-16T00:00:00Z", "duration_ms": 2345, "endpoint": "ss://1.1.1.1:443", "config": {"prefix": "asdf"}}' \
    http://localhost:8080/report

