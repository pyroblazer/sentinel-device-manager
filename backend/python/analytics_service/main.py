"""Sentinel Analytics Service - Event processing, alerting, and analytics."""

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from .handlers import create_router
from .metrics import metrics_endpoint, metrics_middleware
from .store import EventStore
from .tracing import setup_tracing

app = FastAPI(
    title="Sentinel Analytics Service",
    description="Event processing, alerting, and analytics for the Sentinel device management platform.",
    version="1.0.0",
)

app.middleware("http")(metrics_middleware)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

tracer_shutdown = setup_tracing(app, service_name="sentinel-analytics")

store = EventStore()
router = create_router(store)
app.include_router(router)


@app.get("/health")
def health_check():
    return {"status": "ok"}


@app.get("/metrics")
def get_metrics():
    return metrics_endpoint()


def shutdown():
    if tracer_shutdown:
        tracer_shutdown()


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="0.0.0.0", port=8081)
