"""Prometheus metrics instrumentation for the Sentinel Analytics Service."""

import time
from typing import Callable

from fastapi import Request, Response
from prometheus_client import (
    Counter,
    Histogram,
    Gauge,
    generate_latest,
    CONTENT_TYPE_LATEST,
    REGISTRY,
)

# --- HTTP Metrics ---
HTTP_REQUESTS_TOTAL = Counter(
    "sentinel_python_http_requests_total",
    "Total HTTP requests by method, path and status code.",
    ["method", "path", "status_code"],
)

HTTP_REQUEST_DURATION_SECONDS = Histogram(
    "sentinel_python_http_request_duration_seconds",
    "HTTP request duration in seconds.",
    ["method", "path"],
    buckets=(0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0),
)

HTTP_REQUESTS_IN_FLIGHT = Gauge(
    "sentinel_python_http_requests_in_flight",
    "Current in-flight HTTP requests.",
    ["method"],
)

# --- Business Metrics ---
EVENTS_INGESTED_TOTAL = Counter(
    "sentinel_python_events_ingested_total",
    "Total events ingested by severity.",
    ["severity"],
)

ALERTS_CREATED_TOTAL = Counter(
    "sentinel_python_alerts_created_total",
    "Total alerts created by severity.",
    ["severity"],
)

ANALYTICS_QUERIES_TOTAL = Counter(
    "sentinel_python_analytics_queries_total",
    "Total analytics queries by endpoint.",
    ["endpoint"],
)


async def metrics_middleware(request: Request, call_next: Callable) -> Response:
    """FastAPI middleware that records Prometheus metrics for each request."""
    method = request.method
    path = request.url.path

    # Skip the metrics endpoint itself
    if path == "/metrics":
        return await call_next(request)

    HTTP_REQUESTS_IN_FLIGHT.labels(method=method).inc()
    start = time.perf_counter()

    try:
        response = await call_next(request)
        duration = time.perf_counter() - start

        status_code = str(response.status_code)
        HTTP_REQUESTS_TOTAL.labels(method=method, path=path, status_code=status_code).inc()
        HTTP_REQUEST_DURATION_SECONDS.labels(method=method, path=path).observe(duration)

        return response
    except Exception:
        duration = time.perf_counter() - start
        HTTP_REQUESTS_TOTAL.labels(method=method, path=path, status_code="500").inc()
        HTTP_REQUEST_DURATION_SECONDS.labels(method=method, path=path).observe(duration)
        raise
    finally:
        HTTP_REQUESTS_IN_FLIGHT.labels(method=method).dec()


def metrics_endpoint():
    """Return Prometheus-formatted metrics for the /metrics endpoint."""
    return Response(content=generate_latest(REGISTRY), media_type=CONTENT_TYPE_LATEST)


def record_event_ingested(severity: str) -> None:
    EVENTS_INGESTED_TOTAL.labels(severity=severity).inc()


def record_alert_created(severity: str) -> None:
    ALERTS_CREATED_TOTAL.labels(severity=severity).inc()


def record_analytics_query(endpoint: str) -> None:
    ANALYTICS_QUERIES_TOTAL.labels(endpoint=endpoint).inc()
