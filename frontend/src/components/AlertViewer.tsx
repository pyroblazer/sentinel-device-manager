import React, { useEffect, useState } from 'react';
import { analyticsApi, Alert } from '../services/api';

const AlertViewer: React.FC = () => {
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [loading, setLoading] = useState(true);
  const [statusFilter, setStatusFilter] = useState('');

  useEffect(() => {
    loadAlerts();
  }, [statusFilter]);

  async function loadAlerts() {
    try {
      const params: Record<string, string> = {};
      if (statusFilter) params.status = statusFilter;
      const data = await analyticsApi.alerts(params);
      setAlerts(data);
    } catch {
      // handle error
    } finally {
      setLoading(false);
    }
  }

  async function handleAcknowledge(alertId: string) {
    try {
      await analyticsApi.acknowledgeAlert(alertId, 'admin@sentinel.io');
      loadAlerts();
    } catch {
      // handle error
    }
  }

  if (loading) return <div className="loading">Loading alerts...</div>;

  return (
    <div>
      <div className="page-header">
        <h2>Alerts</h2>
        <p>Security alerts and event notifications</p>
      </div>

      <div className="filter-bar" style={{ marginBottom: 24 }}>
        <select value={statusFilter} onChange={e => setStatusFilter(e.target.value)}>
          <option value="">All Statuses</option>
          <option value="ACTIVE">Active</option>
          <option value="ACKNOWLEDGED">Acknowledged</option>
          <option value="RESOLVED">Resolved</option>
        </select>
      </div>

      <div className="table-container">
        <div className="table-toolbar">
          <h3>{alerts.length} Alerts</h3>
        </div>
        <table>
          <thead>
            <tr>
              <th>Alert ID</th>
              <th>Device</th>
              <th>Type</th>
              <th>Severity</th>
              <th>Status</th>
              <th>Message</th>
              <th>Created</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {alerts.map(a => (
              <tr key={a.alert_id}>
                <td className="mono">{a.alert_id.slice(0, 8)}...</td>
                <td className="mono">{a.device_id.slice(0, 8)}...</td>
                <td>{a.alert_type}</td>
                <td><span className={`badge badge-${a.severity.toLowerCase()}`}>{a.severity}</span></td>
                <td><span className={`badge badge-${a.status.toLowerCase()}`}>{a.status}</span></td>
                <td style={{ maxWidth: 300, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{a.message}</td>
                <td>{new Date(a.created_at).toLocaleString()}</td>
                <td>
                  {a.status === 'ACTIVE' && (
                    <button className="btn btn-primary" onClick={() => handleAcknowledge(a.alert_id)}>Acknowledge</button>
                  )}
                </td>
              </tr>
            ))}
            {alerts.length === 0 && (
              <tr><td colSpan={8}><div className="empty-state"><p>No alerts found</p></div></td></tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default AlertViewer;
