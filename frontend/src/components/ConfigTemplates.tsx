import React, { useEffect, useState } from 'react';
import { superappApi } from '../services/api';

interface ConfigTemplate {
  id: string;
  name: string;
  device_type: 'CAMERA' | 'ACCESS_CONTROL' | 'ALARM' | 'SENSOR' | 'ALL';
  config: Record<string, string>;
  description: string;
  created_at: string;
}

const mockTemplates: ConfigTemplate[] = [
  {
    id: 'tpl-001',
    name: 'Standard Camera Profile',
    device_type: 'CAMERA',
    config: { resolution: '4K', encoding: 'H.265', fps: '30', night_vision: 'auto', motion_detection: 'enabled' },
    description: 'Default configuration for all standard IP cameras.',
    created_at: '2026-03-01T00:00:00Z',
  },
  {
    id: 'tpl-002',
    name: 'High-Security Access Control',
    device_type: 'ACCESS_CONTROL',
    config: { auth_mode: 'mfa', lock_timeout: '30s', logging: 'verbose', fail_max: '3' },
    description: 'Enhanced security profile for restricted areas.',
    created_at: '2026-03-15T00:00:00Z',
  },
  {
    id: 'tpl-003',
    name: 'Low-Power Sensor Mode',
    device_type: 'SENSOR',
    config: { report_interval: '300s', low_power: 'true', temp_threshold: '35' },
    description: 'Battery-optimized configuration for remote sensors.',
    created_at: '2026-04-01T00:00:00Z',
  },
];

const deviceTypes = ['CAMERA', 'ACCESS_CONTROL', 'ALARM', 'SENSOR', 'ALL'];

const ConfigTemplates: React.FC = () => {
  const [templates, setTemplates] = useState<ConfigTemplate[]>([]);
  const [loading, setLoading] = useState(true);

  // Form state
  const [showForm, setShowForm] = useState(false);
  const [newName, setNewName] = useState('');
  const [newDeviceType, setNewDeviceType] = useState('CAMERA');
  const [newDescription, setNewDescription] = useState('');
  const [configPairs, setConfigPairs] = useState<{ key: string; value: string }[]>([{ key: '', value: '' }]);

  // Detail view
  const [selectedId, setSelectedId] = useState<string | null>(null);

  useEffect(() => {
    loadTemplates();
  }, []);

  async function loadTemplates() {
    try {
      const data = await superappApi.listTemplates();
      setTemplates(data.templates || data || []);
    } catch {
      setTemplates(mockTemplates);
    } finally {
      setLoading(false);
    }
  }

  function addConfigPair() {
    setConfigPairs(prev => [...prev, { key: '', value: '' }]);
  }

  function removeConfigPair(index: number) {
    setConfigPairs(prev => prev.filter((_, i) => i !== index));
  }

  function updateConfigPair(index: number, field: 'key' | 'value', val: string) {
    setConfigPairs(prev => prev.map((p, i) => i === index ? { ...p, [field]: val } : p));
  }

  async function handleCreate() {
    if (!newName.trim()) return;
    const config: Record<string, string> = {};
    configPairs.forEach(p => {
      if (p.key.trim()) config[p.key.trim()] = p.value;
    });
    const payload = { name: newName, device_type: newDeviceType, config, description: newDescription };
    try {
      const result = await superappApi.createTemplate(payload);
      setTemplates(prev => [...prev, result.template || result]);
    } catch {
      setTemplates(prev => [...prev, { id: `tpl-${Date.now()}`, ...payload, device_type: newDeviceType as ConfigTemplate['device_type'], created_at: new Date().toISOString() }]);
    }
    setNewName('');
    setNewDeviceType('CAMERA');
    setNewDescription('');
    setConfigPairs([{ key: '', value: '' }]);
    setShowForm(false);
  }

  async function handleDelete(id: string) {
    if (!window.confirm('Delete this template?')) return;
    try { await superappApi.deleteTemplate(id); } catch { /* ok */ }
    setTemplates(prev => prev.filter(t => t.id !== id));
    if (selectedId === id) setSelectedId(null);
  }

  if (loading) return <div className="loading">Loading templates...</div>;

  const selected = selectedId ? templates.find(t => t.id === selectedId) : null;

  return (
    <div>
      <div className="page-header">
        <h2>Configuration Templates</h2>
        <p>Reusable configuration profiles for device types</p>
      </div>

      <div style={{ marginBottom: 24 }}>
        <button className="btn btn-primary" onClick={() => setShowForm(!showForm)}>
          {showForm ? 'Cancel' : 'Create Template'}
        </button>
      </div>

      {showForm && (
        <div className="stat-card" style={{ marginBottom: 24 }}>
          <h3 style={{ marginBottom: 16 }}>New Template</h3>
          <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap', marginBottom: 12 }}>
            <div style={{ flex: 2, minWidth: 180 }}>
              <label style={{ display: 'block', color: 'var(--text-muted)', marginBottom: 4, fontSize: 13 }}>Name</label>
              <input className="filter-bar" style={{ width: '100%', padding: '8px 12px' }} value={newName} onChange={e => setNewName(e.target.value)} placeholder="Template name" />
            </div>
            <div style={{ flex: 1, minWidth: 160 }}>
              <label style={{ display: 'block', color: 'var(--text-muted)', marginBottom: 4, fontSize: 13 }}>Device Type</label>
              <select className="filter-bar" style={{ width: '100%', padding: '8px 12px' }} value={newDeviceType} onChange={e => setNewDeviceType(e.target.value)}>
                {deviceTypes.map(dt => <option key={dt} value={dt}>{dt.replace('_', ' ')}</option>)}
              </select>
            </div>
          </div>
          <div style={{ marginBottom: 12 }}>
            <label style={{ display: 'block', color: 'var(--text-muted)', marginBottom: 4, fontSize: 13 }}>Description</label>
            <input className="filter-bar" style={{ width: '100%', padding: '8px 12px' }} value={newDescription} onChange={e => setNewDescription(e.target.value)} placeholder="Brief description of this template" />
          </div>
          <div style={{ marginBottom: 16 }}>
            <label style={{ display: 'block', color: 'var(--text-muted)', marginBottom: 8, fontSize: 13 }}>Config (Key-Value Pairs)</label>
            {configPairs.map((pair, i) => (
              <div key={i} style={{ display: 'flex', gap: 8, marginBottom: 8, alignItems: 'center' }}>
                <input className="filter-bar" style={{ flex: 1, padding: '6px 10px' }} placeholder="Key" value={pair.key} onChange={e => updateConfigPair(i, 'key', e.target.value)} />
                <span style={{ color: 'var(--text-muted)' }}>=</span>
                <input className="filter-bar" style={{ flex: 1, padding: '6px 10px' }} placeholder="Value" value={pair.value} onChange={e => updateConfigPair(i, 'value', e.target.value)} />
                {configPairs.length > 1 && (
                  <button className="btn btn-danger" style={{ padding: '4px 10px', fontSize: 12 }} onClick={() => removeConfigPair(i)}>X</button>
                )}
              </div>
            ))}
            <button className="btn btn-secondary" style={{ fontSize: 13 }} onClick={addConfigPair}>+ Add Field</button>
          </div>
          <button className="btn btn-primary" onClick={handleCreate}>Create Template</button>
        </div>
      )}

      {/* Detail view */}
      {selected && (
        <div className="stat-card" style={{ marginBottom: 24 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 16 }}>
            <div>
              <h3 style={{ color: 'var(--text-primary)', marginBottom: 4 }}>{selected.name}</h3>
              <span className={`badge badge-${selected.device_type.toLowerCase()}`}>{selected.device_type}</span>
              <span style={{ color: 'var(--text-muted)', marginLeft: 12, fontSize: 13 }}>Created {new Date(selected.created_at).toLocaleDateString()}</span>
            </div>
            <button className="btn btn-secondary" onClick={() => setSelectedId(null)}>Close</button>
          </div>
          {selected.description && <p style={{ color: 'var(--text-secondary)', marginBottom: 16, fontSize: 14 }}>{selected.description}</p>}
          <h4 style={{ color: 'var(--text-secondary)', marginBottom: 8 }}>Configuration</h4>
          <table>
            <thead>
              <tr><th>Key</th><th>Value</th></tr>
            </thead>
            <tbody>
              {Object.entries(selected.config).map(([k, v]) => (
                <tr key={k}>
                  <td className="mono" style={{ color: 'var(--accent-cyan)' }}>{k}</td>
                  <td className="mono">{v}</td>
                </tr>
              ))}
              {Object.keys(selected.config).length === 0 && (
                <tr><td colSpan={2} style={{ color: 'var(--text-muted)', textAlign: 'center' }}>No config parameters</td></tr>
              )}
            </tbody>
          </table>
        </div>
      )}

      {/* Templates list */}
      <div className="table-container">
        <div className="table-toolbar"><h3>{templates.length} Templates</h3></div>
        <table>
          <thead>
            <tr><th>Name</th><th>Device Type</th><th>Config Keys</th><th>Description</th><th>Created</th><th>Actions</th></tr>
          </thead>
          <tbody>
            {templates.map(t => (
              <tr key={t.id} onClick={() => setSelectedId(t.id)} style={{ cursor: 'pointer' }}>
                <td style={{ fontWeight: 600, color: 'var(--text-primary)' }}>{t.name}</td>
                <td><span className={`badge badge-${t.device_type.toLowerCase()}`}>{t.device_type}</span></td>
                <td style={{ color: 'var(--text-muted)' }}>{Object.keys(t.config).length} keys</td>
                <td style={{ maxWidth: 300, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', color: 'var(--text-secondary)', fontSize: 14 }}>{t.description}</td>
                <td>{new Date(t.created_at).toLocaleDateString()}</td>
                <td>
                  <button className="btn btn-secondary" style={{ marginRight: 6 }} onClick={e => { e.stopPropagation(); setSelectedId(t.id); }}>View</button>
                  <button className="btn btn-danger" onClick={e => { e.stopPropagation(); handleDelete(t.id); }}>Delete</button>
                </td>
              </tr>
            ))}
            {templates.length === 0 && (
              <tr><td colSpan={6}><div className="empty-state"><p>No configuration templates found</p></div></td></tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default ConfigTemplates;
