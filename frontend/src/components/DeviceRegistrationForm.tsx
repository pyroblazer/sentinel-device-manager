import React, { useState } from 'react';
import { deviceApi } from '../services/api';

interface Props {
  onCreated: () => void;
}

const DeviceRegistrationForm: React.FC<Props> = ({ onCreated }) => {
  const [form, setForm] = useState({
    serial_number: '',
    device_type: 'CAMERA' as 'CAMERA' | 'ACCESS_CONTROL' | 'ALARM' | 'SENSOR',
    model: '',
    site_id: '',
    organization_id: 'org-demo-001',
    ip_address: '',
    mac_address: '',
  });
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');

  function update(field: string, value: string) {
    setForm(prev => ({ ...prev, [field]: value }));
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!form.serial_number.trim() || !form.site_id.trim() || !form.organization_id.trim()) {
      setError('Serial number, site ID, and organization ID are required.');
      return;
    }
    setSubmitting(true);
    setError('');
    try {
      await deviceApi.create(form);
      onCreated();
    } catch (err: any) {
      setError(err.message || 'Registration failed');
    } finally {
      setSubmitting(false);
    }
  }

  const inputStyle: React.CSSProperties = {
    width: '100%', padding: '8px 12px',
    background: 'var(--bg-primary)', border: '1px solid var(--border)',
    borderRadius: 6, color: 'var(--text-primary)', fontSize: 14, boxSizing: 'border-box',
  };
  const labelStyle: React.CSSProperties = {
    display: 'block', color: 'var(--text-muted)', marginBottom: 4, fontSize: 13,
  };

  return (
    <div className="stat-card" style={{ marginBottom: 24 }}>
      <h3 style={{ marginBottom: 16 }}>Register New Device</h3>
      <form onSubmit={handleSubmit}>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: 16 }}>
          <div>
            <label style={labelStyle}>Serial Number *</label>
            <input style={inputStyle} value={form.serial_number}
              onChange={e => update('serial_number', e.target.value)}
              placeholder="e.g. VKD-CAM-005" />
          </div>
          <div>
            <label style={labelStyle}>Device Type *</label>
            <select style={inputStyle} value={form.device_type}
              onChange={e => update('device_type', e.target.value)}>
              <option value="CAMERA">Camera</option>
              <option value="ACCESS_CONTROL">Access Control</option>
              <option value="ALARM">Alarm</option>
              <option value="SENSOR">Sensor</option>
            </select>
          </div>
          <div>
            <label style={labelStyle}>Model</label>
            <input style={inputStyle} value={form.model}
              onChange={e => update('model', e.target.value)}
              placeholder="e.g. D30" />
          </div>
          <div>
            <label style={labelStyle}>Site ID *</label>
            <input style={inputStyle} value={form.site_id}
              onChange={e => update('site_id', e.target.value)}
              placeholder="e.g. site-alpha-001" />
          </div>
          <div>
            <label style={labelStyle}>Organization ID *</label>
            <input style={inputStyle} value={form.organization_id}
              onChange={e => update('organization_id', e.target.value)}
              placeholder="e.g. org-demo-001" />
          </div>
          <div>
            <label style={labelStyle}>IP Address</label>
            <input style={inputStyle} value={form.ip_address}
              onChange={e => update('ip_address', e.target.value)}
              placeholder="e.g. 10.0.1.150" />
          </div>
          <div>
            <label style={labelStyle}>MAC Address</label>
            <input style={inputStyle} value={form.mac_address}
              onChange={e => update('mac_address', e.target.value)}
              placeholder="e.g. AA:BB:CC:DD:EE:FF" />
          </div>
        </div>
        {error && <p style={{ color: 'var(--accent-red)', fontSize: 13, marginTop: 12 }}>{error}</p>}
        <div style={{ marginTop: 16, display: 'flex', gap: 8 }}>
          <button className="btn btn-primary" type="submit" disabled={submitting}>
            {submitting ? 'Registering...' : 'Register Device'}
          </button>
        </div>
      </form>
    </div>
  );
};

export default DeviceRegistrationForm;
