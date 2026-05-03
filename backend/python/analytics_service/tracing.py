"""OpenTelemetry distributed tracing for the Sentinel Analytics Service."""

import os
from typing import Optional

from fastapi import FastAPI
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.sdk.resources import Resource, SERVICE_NAME
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor


def setup_tracing(
    app: FastAPI,
    endpoint: Optional[str] = None,
    service_name: str = "sentinel-analytics",
) -> Optional[callable]:
    """
    Configure OpenTelemetry tracing for a FastAPI application.

    Args:
        app: The FastAPI application instance.
        endpoint: OTel Collector gRPC endpoint (e.g. "otel-collector:4317").
                  If None, tracing is disabled.
        service_name: Service name for span attributes.

    Returns:
        A shutdown callable, or None if tracing is disabled.
    """
    if endpoint is None:
        endpoint = os.getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
        if not endpoint or endpoint == "none":
            return None

    resource = Resource.create({SERVICE_NAME: service_name})

    exporter = OTLPSpanExporter(endpoint=endpoint, insecure=True)
    processor = BatchSpanProcessor(exporter)

    provider = TracerProvider(resource=resource)
    provider.add_span_processor(processor)
    trace.set_tracer_provider(provider)

    FastAPIInstrumentor.instrument_app(app)

    def shutdown():
        provider.shutdown()

    return shutdown


def get_trace_id() -> str:
    """Extract the current trace ID from the active span context."""
    span = trace.get_current_span()
    ctx = span.get_span_context()
    if ctx.is_valid:
        return format(ctx.trace_id, "032x")
    return ""
