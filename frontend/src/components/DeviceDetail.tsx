import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { deviceApi, Device, DeviceHealth } from '../services/api';

const DeviceDetail: React.FC = () => {
  const { deviceId } = useParams<{ deviceId: string }>();
  const navigate = useNavigate();
  const [device, setDevice] = useState<Device | null>(null);
  const [health, setHealth] = useState<DeviceHealth | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!deviceId) return;
    async function load() {
      try {
        const d = await deviceApi.get(deviceId!);
        setDevice(d);
        try {
          const h = await deviceApi.health(deviceId!);
          setHealth(h);
        } catch {
          // health endpoint may not be available
        }
      } catch {
        // device not found
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [deviceId]);

  async function handleDecommission() {
    if (!device || !window.confirm('Decommission this device?')) return;
    const updated = await deviceApi.update(device.device_id, { status: 'DECOMMISSIONED' } as any);
    setDevice(updated);
  }

  if (loading) return <div className="loading">Loading device...</div>;
  if (!device) return <div className="empty-state"><p>Device not found</p></div>;

  return (
    <div className="device-detail">
      <div className="device-detail-header">
        <div>
          <div className="page-header" style={{ marginBottom: 0 }}>
            <h2>{device.serial_number}</h2>
            <p>{device.device_type} - {device.model}</p>
          </div>
        </div>
        <div style={{ display: 'flex', gap: 8 }}>
          <button className="btn btn-secondary" onClick={() => navigate('/devices')}>Back</button>
          <button className="btn btn-danger" onClick={handleDecommission}>Decommission</button>
        </div>
      </div>

      <div className="detail-grid">
        <div className="detail-field">
          <label>Device ID</label>
          <span>{device.device_id}</span>
        </div>
        <div className="detail-field">
          <label>Status</label>
          <span className={`badge badge-${device.status.toLowerCase()}`}>{device.status}</span>
        </div>
        <div className="detail-field">
          <label>Firmware</label>
          <span>{device.firmware_version}</span>
        </div>
        <div className="detail-field">
          <label>Site</label>
          <span>{device.site_id}</span>
        </div>
        <div className="detail-field">
          <label>IP Address</label>
          <span>{device.ip_address || 'N/A'}</span>
        </div>
        <div className="detail-field">
          <label>MAC Address</label>
          <span>{device.mac_address || 'N/A'}</span>
        </div>
        <div className="detail-field">
          <label>Created</label>
          <span>{new Date(device.created_at).toLocaleString()}</span>
        </div>
        <div className="detail-field">
          <label>Last Heartbeat</label>
          <span>{new Date(device.last_heartbeat).toLocaleString()}</span>
        </div>
      </div>

      {health && (
        <>
          <h3 style={{ marginBottom: 16, color: 'var(--text-secondary)' }}>Health Metrics</h3>
          <div className="detail-grid">
            <div className="detail-field">
              <label>CPU Usage</label>
              <span>{health.cpu_usage.toFixed(1)}%</span>
            </div>
            <div className="detail-field">
              <label>Memory Usage</label>
              <span>{health.memory_usage.toFixed(1)}%</span>
            </div>
            <div className="detail-field">
              <label>Temperature</label>
              <span>{health.temperature_c.toFixed(1)} C</span>
            </div>
            <div className="detail-field">
              <label>Network Latency</label>
              <span>{health.network_latency_ms.toFixed(1)} ms</span>
            </div>
          </div>
        </>
      )}
    </div>
  );
};

export default DeviceDetail;
