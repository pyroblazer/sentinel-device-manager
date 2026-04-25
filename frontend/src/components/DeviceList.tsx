import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { deviceApi, Device } from '../services/api';

const DeviceList: React.FC = () => {
  const navigate = useNavigate();
  const [devices, setDevices] = useState<Device[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState({ type: '', status: '' });

  useEffect(() => {
    loadDevices();
  }, [filter]);

  async function loadDevices() {
    try {
      const params: Record<string, string> = {};
      if (filter.type) params.type = filter.type;
      if (filter.status) params.status = filter.status;
      const res = await deviceApi.list(params);
      setDevices(res.devices);
    } catch {
      // handle error
    } finally {
      setLoading(false);
    }
  }

  async function handleDelete(id: string) {
    if (!window.confirm('Decommission this device?')) return;
    await deviceApi.delete(id);
    loadDevices();
  }

  if (loading) return <div className="loading">Loading devices...</div>;

  return (
    <div>
      <div className="page-header">
        <h2>Devices</h2>
        <p>Manage cameras, access control, alarms, and sensors</p>
      </div>

      <div className="filter-bar" style={{ marginBottom: 24 }}>
        <select value={filter.type} onChange={e => setFilter(f => ({ ...f, type: e.target.value }))}>
          <option value="">All Types</option>
          <option value="CAMERA">Camera</option>
          <option value="ACCESS_CONTROL">Access Control</option>
          <option value="ALARM">Alarm</option>
          <option value="SENSOR">Sensor</option>
        </select>
        <select value={filter.status} onChange={e => setFilter(f => ({ ...f, status: e.target.value }))}>
          <option value="">All Statuses</option>
          <option value="ONLINE">Online</option>
          <option value="OFFLINE">Offline</option>
          <option value="MAINTENANCE">Maintenance</option>
          <option value="DECOMMISSIONED">Decommissioned</option>
        </select>
      </div>

      <div className="table-container">
        <div className="table-toolbar">
          <h3>{devices.length} Devices</h3>
        </div>
        <table>
          <thead>
            <tr>
              <th>Serial Number</th>
              <th>Type</th>
              <th>Model</th>
              <th>Site</th>
              <th>Status</th>
              <th>Firmware</th>
              <th>Last Heartbeat</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {devices.map(d => (
              <tr key={d.device_id} onClick={() => navigate(`/devices/${d.device_id}`)} style={{ cursor: 'pointer' }}>
                <td className="mono">{d.serial_number}</td>
                <td><span className={`badge badge-${d.device_type.toLowerCase()}`}>{d.device_type}</span></td>
                <td>{d.model}</td>
                <td className="mono">{d.site_id.slice(0, 8)}...</td>
                <td><span className={`badge badge-${d.status.toLowerCase()}`}>{d.status}</span></td>
                <td className="mono">{d.firmware_version}</td>
                <td>{new Date(d.last_heartbeat).toLocaleString()}</td>
                <td>
                  <button className="btn btn-danger" onClick={e => { e.stopPropagation(); handleDelete(d.device_id); }}>Delete</button>
                </td>
              </tr>
            ))}
            {devices.length === 0 && (
              <tr><td colSpan={8}><div className="empty-state"><p>No devices found</p></div></td></tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default DeviceList;
