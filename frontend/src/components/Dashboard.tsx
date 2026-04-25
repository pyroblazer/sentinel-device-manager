import React, { useEffect, useState } from 'react';
import { analyticsApi, deviceApi, AnalyticsSummary, Device } from '../services/api';

const Dashboard: React.FC = () => {
  const [summary, setSummary] = useState<AnalyticsSummary | null>(null);
  const [devices, setDevices] = useState<Device[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function load() {
      try {
        const [s, d] = await Promise.all([
          analyticsApi.summary(),
          deviceApi.list(),
        ]);
        setSummary(s);
        setDevices(d.devices);
      } catch (err: any) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    }
    load();
  }, []);

  if (loading) return <div className="loading">Loading dashboard...</div>;
  if (error) return <div className="empty-state"><p>Error: {error}</p></div>;

  const onlineCount = devices.filter(d => d.status === 'ONLINE').length;
  const offlineCount = devices.filter(d => d.status === 'OFFLINE').length;

  return (
    <div>
      <div className="page-header">
        <h2>Dashboard</h2>
        <p>Real-time overview of your physical security infrastructure</p>
      </div>

      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-label">Total Devices</div>
          <div className="stat-value blue">{summary?.total_devices ?? devices.length}</div>
        </div>
        <div className="stat-card">
          <div className="stat-label">Online</div>
          <div className="stat-value green">{summary?.online_devices ?? onlineCount}</div>
        </div>
        <div className="stat-card">
          <div className="stat-label">Offline</div>
          <div className="stat-value red">{summary?.offline_devices ?? offlineCount}</div>
        </div>
        <div className="stat-card">
          <div className="stat-label">Active Alerts</div>
          <div className="stat-value yellow">{summary?.active_alerts ?? 0}</div>
        </div>
        <div className="stat-card">
          <div className="stat-label">Events (24h)</div>
          <div className="stat-value cyan">{summary?.events_last_24h ?? 0}</div>
        </div>
        <div className="stat-card">
          <div className="stat-label">Critical Alerts</div>
          <div className="stat-value red">{summary?.critical_alerts ?? 0}</div>
        </div>
        <div className="stat-card">
          <div className="stat-label">FW Compliance</div>
          <div className="stat-value green">{summary?.firmware_compliance_pct?.toFixed(1) ?? '0.0'}%</div>
        </div>
      </div>

      <div className="table-container">
        <div className="table-toolbar">
          <h3>Recent Devices</h3>
        </div>
        <table>
          <thead>
            <tr>
              <th>Serial</th>
              <th>Type</th>
              <th>Model</th>
              <th>Site</th>
              <th>Status</th>
              <th>Firmware</th>
            </tr>
          </thead>
          <tbody>
            {devices.slice(0, 10).map(d => (
              <tr key={d.device_id}>
                <td className="mono">{d.serial_number}</td>
                <td><span className={`badge badge-${d.device_type.toLowerCase()}`}>{d.device_type}</span></td>
                <td>{d.model}</td>
                <td className="mono">{d.site_id}</td>
                <td><span className={`badge badge-${d.status.toLowerCase()}`}>{d.status}</span></td>
                <td className="mono">{d.firmware_version}</td>
              </tr>
            ))}
            {devices.length === 0 && (
              <tr><td colSpan={6} style={{ textAlign: 'center', color: 'var(--text-muted)', padding: 32 }}>No devices registered</td></tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default Dashboard;
