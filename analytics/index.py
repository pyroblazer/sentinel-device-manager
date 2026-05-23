"""Vercel Python serverless handler for Sentinel Analytics Service."""

import sys
import os

# Make the analytics_service package importable from backend/python/
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'backend', 'python'))

from analytics_service.main import app  # noqa: E402
