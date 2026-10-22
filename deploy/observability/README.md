# Observability backend

Where the telemetry of the API and the worker goes when you ask it to leave
the process. Four containers, one network, nothing that the application
depends on: with this directory absent, the application still starts and
logs to stdout as it always did.

| | |
| --- | --- |
| Traces | Tempo, queried by Grafana and by the dashboard |
| Metrics | the collector's prometheus exporter, scraped by Prometheus |
| Dashboards | Grafana, provisioned from `grafana/` |
| Logs | none — the application logs to stdout and that is deliberate |

## Start it

```sh
docker compose -f deploy/observability/docker-compose.observability.yml up -d
```

The Compose plugin (`docker compose`) is the usual one. A machine that only
has the standalone binary takes the same arguments and does the same thing:

```sh
docker-compose -f deploy/observability/docker-compose.observability.yml up -d
```

Then point the application at it. The exporter defaults to `stdout`, so
nothing is exported until you say so:

```sh
TELEMETRY_EXPORTER=otlp
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
```

`OTEL_EXPORTER_OTLP_PROTOCOL` accepts `grpc` (the default, port 4317) or
`http` (port 4318). `OTEL_EXPORTER_OTLP_INSECURE=true` is right for a
local stack; any other deployment wants TLS.

| URL | What |
| --- | --- |
| http://localhost:3000 | the dashboard — login `admin` / `admin` |
| http://localhost:9090 | Prometheus, for raw metric queries |
| http://localhost:3200 | Tempo, for raw trace queries |
| http://localhost:13133 | the collector's health endpoint |

Stop it with `down`. Add `-v` to also throw away the stored traces and
metrics.

## What is on the dashboard

`Requisition System`, provisioned from `grafana/dashboards/`. Two variables,
`service` and `env`, filter every panel; `service` separates the API from the
worker.

- **Overview** — request rate, latency percentiles, the 5xx ratio, and which
  routes are being hit.
- **Business operations** — the four counters and durations the application
  reports for invitations and acceptances, split by attempted, succeeded and
  failed.
- **Worker operations** — the same for email delivery and invitation
  expirations, measured in the worker rather than in the request that queued
  them.
- **Dependencies** — Postgres and Redis query durations, and the Redis pool.
- **Traces** — a TraceQL search over the selected services.

Two things are worth knowing about the panels before reading them:

- **The worker panels need the worker.** They are measured in the process that
  runs the task, not in the request that queued it, so they stay empty until
  `go run ./cmd/worker` runs next to the API.
- **The counters start at zero.** Each of them is published with a zero value
  when the process starts, so an outcome that has not happened yet still has
  a line: "nothing failed" reads as zero and not as a missing series.

Two things are missing on purpose, and the dashboard does not pretend
otherwise:

- **No Redis *metrics*.** The Redis instrumentation reports the state of the
  pool and nothing else — it records no query duration at all — so the Redis
  duration is read straight out of the traces with TraceQL metrics. Postgres
  does publish a real histogram, `db.client.operation.duration`, so its
  duration is a metric rather than a trace query: that panel holds even when
  the traces of the moment were not sampled.
- **No service map.** Grafana's service map needs Tempo's metrics generator
  writing span metrics into Prometheus, which this stack does not run. Trace
  relationships within a single trace still work.

## Revalidate the configs

Both of these are worth running after any edit; a collector config that does
not parse is a stack that does not start.

```sh
docker run --rm -v "$PWD/deploy/observability/otel-collector:/etc/otelcol:ro" \
  --entrypoint /otelcol-contrib otel/opentelemetry-collector-contrib:0.163.2 \
  validate --config=/etc/otelcol/config.yaml

docker run --rm -v "$PWD/deploy/observability/prometheus:/etc/prometheus:ro" \
  --entrypoint /bin/promtool prom/prometheus:v3.15.1 \
  check config /etc/prometheus/prometheus.yml
```

Tempo has no equivalent validator: its config is rejected at startup, so
`docker compose ... logs tempo` is the check.

## Where the telemetry is read from

The application speaks OTLP to the collector on 4317 and never opens a
metrics endpoint of its own. The collector fans traces out to Tempo and
leaves metrics on 9464 for Prometheus to pull.

Prometheus scrapes with `honor_labels: true`, so the labels the collector
put on each series survive: `service_name`, `deployment_environment_name`,
`service_version`, `host_name` and, on the HTTP metrics, `http_route`,
`http_request_method` and `http_response_status_code`.

Retention is 24 hours in both backends, so a panel and a trace of the same
minute cover the same window.

## Is it working?

Work outwards from the process. Each step only makes sense if the one before
it passed.

**1. The collector is receiving something.** With the application running and
exporting to `otlp`, its metrics appear on the collector's own endpoint — this
is the check that matters, because it means the application is talking to the
collector and the collector has not dropped anything:

```sh
curl -s http://localhost:9464/metrics | grep -E '^(requisition_|http_server_)' | head
```

Each line carries `service_name`, so `requisition-system` and
`requisition-system-worker` are already told apart. No lines means the
application is not exporting: check `TELEMETRY_EXPORTER=otlp` and
`OTEL_EXPORTER_OTLP_ENDPOINT`.

**2. Prometheus is scraping it.** http://localhost:9090/targets should show
both targets `up`; then:

```sh
curl -s -G http://localhost:9090/api/v1/query \
  --data-urlencode 'query=requisition_admin_invitations_attempted_total'
```

**3. Tempo has the traces.** No trace has to be recorded for this to work, but
one has to exist for anything to come back:

```sh
curl -s -G http://localhost:3200/api/search \
  --data-urlencode 'q={.service.name="requisition-system"}'
```

**4. The dashboard renders.** http://localhost:3000/d/requisition-system — if
the panels are empty but steps 1 to 3 pass, widen the time range; the stack
only keeps 24 hours.

**How long a request takes to show up.** The application exports its metrics
every `OTEL_METRIC_EXPORT_INTERVAL` (10 seconds by default), the collector
batches what it receives for up to 5 seconds, Prometheus scrapes it every 15
seconds and the dashboard refreshes every 30 seconds. A request shows up
within about a minute; the export interval is the one of those four that the
application controls.

## Troubleshooting

**Every panel is empty.** Almost always that the application is still on the
default exporter. `TELEMETRY_EXPORTER=otlp` and
`OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317` — and restart the process, the
exporter is chosen at startup. Confirm with step 1 above.

**Metrics arrive but there is no `service_name` label.** Something turned off
`honor_labels` in `prometheus.yml`. Without it Prometheus discards the labels
the collector put on each series in favour of its own, and every panel
filtering on `service_name` goes blank.

**Metrics arrive but no traces appear.** The traces pipeline is separate: the
collector forwards them to `tempo:4317`, which is a different address from
the one the application uses. Read the collector's logs for a refused
connection and Tempo's for an ingestion error:

```sh
docker compose -f deploy/observability/docker-compose.observability.yml logs otel-collector tempo
```

**Tempo keeps restarting.** Its config is rejected at startup and the error is
in the log. Tempo 3.x does not accept the `ingester` or `compactor` blocks
from earlier versions, and this directory is written for 3.1.1 specifically.

**A container exits with a permission error on start.** A named volume was
created while owned by root, which happens if the mount path is changed to one
the image does not already own. Both volumes are tied to image-owned
directories for exactly this reason; `docker compose ... down -v` and starting
again recreates them.

**A panel that reads from Tempo is empty while the Prometheus panels are
fine.** The Redis duration panel queries spans, so it needs traces to exist —
see step 3. It is also not instantaneous: a query over a window in which
nothing was traced has nothing to measure. The Postgres duration panel reads
a metric and is not affected by any of this.

**A panel of the worker row is empty while the business ones are filled.**
The worker is not running, or it is exporting under another service name:
the worker panels filter on `service_name` like every other panel, and the
worker publishes under `requisition-system-worker` unless `OTEL_SERVICE_NAME`
says otherwise.
