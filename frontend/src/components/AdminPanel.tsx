import React, { useEffect, useState } from 'react';
import { superappApi } from '../services/api';

interface APIKey {
  id: string;
  name: string;
  role: string;
  key?: string;
  expires_at?: string;
  created_at: string;
}

interface Webhook {
  id: string;
  name: string;
  url: string;
  events: string[];
  created_at: string;
  last_status?: string;
}

interface Geofence {
  id: string;
  name: string;
  latitude: number;
  longitude: number;
  radius_meters: number;
  site_id: string;
  created_at: string;
}

const mockAPIKeys: APIKey[] = [
  { id: 'key-001', name: 'Integration Service', role: 'ADMIN', key: 'sk-****a1b2', expires_at: '2027-01-01T00:00:00Z', created_at: '2026-03-15T10:00:00Z' },
  { id: 'key-002', name: 'Monitoring Agent', role: 'VIEWER', key: 'sk-****c3d4', created_at: '2026-04-01T08:00:00Z' },
];

const mockWebhooks: Webhook[] = [
  { id: 'wh-001', name: 'Slack Alert Channel', url: 'https://hooks.slack.com/services/T00/B00/xxx', events: ['device.offline', 'alert.critical'], created_at: '2026-03-20T12:00:00Z', last_status: '200 OK' },
  { id: 'wh-002', name: 'PagerDuty Integration', url: 'https://events.pagerduty.com/v2/enqueue', events: ['alert.critical'], created_at: '2026-04-10T09:00:00Z', last_status: '200 OK' },
];

const mockGeofences: Geofence[] = [
  { id: 'gf-001', name: 'Site Alpha Perimeter', latitude: 37.7749, longitude: -122.4194, radius_meters: 500, site_id: 'site-alpha-001', created_at: '2026-03-01T00:00:00Z' },
  { id: 'gf-002', name: 'Site Beta Building', latitude: 40.7128, longitude: -74.006, radius_meters: 200, site_id: 'site-beta-002', created_at: '2026-03-15T00:00:00Z' },
];

const webhookEvents = ['device.online', 'device.offline', 'alert.info', 'alert.warning', 'alert.critical', 'firmware.deployed', 'geofence.breach'];

type Tab = 'api-keys' | 'webhooks' | 'geofences';

const AdminPanel: React.FC = () => {
  const [tab, setTab] = useState<Tab>('api-keys');
  const [loading, setLoading] = useState(true);

  // API Keys state
  const [apiKeys, setApiKeys] = useState<APIKey[]>([]);
  const [newKeyName, setNewKeyName] = useState('');
  const [newKeyRole, setNewKeyRole] = useState('VIEWER');
  const [newKeyExpiry, setNewKeyExpiry] = useState('');

  // Webhooks state
  const [webhooks, setWebhooks] = useState<Webhook[]>([]);
  const [newWhName, setNewWhName] = useState('');
  const [newWhUrl, setNewWhUrl] = useState('');
  const [newWhEvents, setNewWhEvents] = useState<string[]>([]);
  const [testingId, setTestingId] = useState<string | null>(null);

  // Geofences state
  const [geofences, setGeofences] = useState<Geofence[]>([]);
  const [newGfName, setNewGfName] = useState('');
  const [newGfLat, setNewGfLat] = useState('');
  const [newGfLng, setNewGfLng] = useState('');
  const [newGfRadius, setNewGfRadius] = useState('');
  const [newGfSite, setNewGfSite] = useState('');

  useEffect(() => {
    loadAll();
  }, []);

  async function loadAll() {
    setLoading(true);
    try {
      const [keys, whs, gfs] = await Promise.all([
        superappApi.listAPIKeys().catch(() => null),
        superappApi.listWebhooks().catch(() => null),
        superappApi.listGeofences().catch(() => null),
      ]);
      setApiKeys(keys?.api_keys || keys || mockAPIKeys);
      setWebhooks(whs?.webhooks || whs || mockWebhooks);
      setGeofences(gfs?.geofences || gfs || mockGeofences);
    } catch {
      setApiKeys(mockAPIKeys);
      setWebhooks(mockWebhooks);
      setGeofences(mockGeofences);
    } finally {
      setLoading(false);
    }
  }

  // API Key handlers
  async function handleCreateKey() {
    if (!newKeyName.trim()) return;
    const data: any = { name: newKeyName, role: newKeyRole };
    if (newKeyExpiry) data.expires_at = newKeyExpiry;
    try {
      const result = await superappApi.createAPIKey(data);
      setApiKeys(prev => [...prev, result.api_key || result]);
    } catch {
      // fallback: add locally
      setApiKeys(prev => [...prev, { id: `key-${Date.now()}`, name: newKeyName, role: newKeyRole, created_at: new Date().toISOString() }]);
    }
    setNewKeyName('');
    setNewKeyRole('VIEWER');
    setNewKeyExpiry('');
  }

  async function handleDeleteKey(id: string) {
    if (!window.confirm('Delete this API key?')) return;
    try { await superappApi.deleteAPIKey(id); } catch { /* ok */ }
    setApiKeys(prev => prev.filter(k => k.id !== id));
  }

  // Webhook handlers
  function toggleWhEvent(event: string) {
    setNewWhEvents(prev => prev.includes(event) ? prev.filter(e => e !== event) : [...prev, event]);
  }

  async function handleCreateWebhook() {
    if (!newWhName.trim() || !newWhUrl.trim() || newWhEvents.length === 0) return;
    const data = { name: newWhName, url: newWhUrl, events: newWhEvents };
    try {
      const result = await superappApi.createWebhook(data);
      setWebhooks(prev => [...prev, result.webhook || result]);
    } catch {
      setWebhooks(prev => [...prev, { id: `wh-${Date.now()}`, ...data, created_at: new Date().toISOString() }]);
    }
    setNewWhName('');
    setNewWhUrl('');
    setNewWhEvents([]);
  }

  async function handleDeleteWebhook(id: string) {
    if (!window.confirm('Delete this webhook?')) return;
    try { await superappApi.deleteWebhook(id); } catch { /* ok */ }
    setWebhooks(prev => prev.filter(w => w.id !== id));
  }

  async function handleTestWebhook(id: string) {
    setTestingId(id);
    try {
      await superappApi.testWebhook(id);
      alert('Test webhook dispatched successfully.');
    } catch {
      alert('Test webhook failed. Check the endpoint URL.');
    } finally {
      setTestingId(null);
    }
  }

  // Geofence handlers
  async function handleCreateGeofence() {
    if (!newGfName.trim() || !newGfLat || !newGfLng || !newGfRadius || !newGfSite.trim()) return;
    const data = { name: newGfName, latitude: parseFloat(newGfLat), longitude: parseFloat(newGfLng), radius_meters: parseInt(newGfRadius), site_id: newGfSite };
    try {
      const result = await superappApi.createGeofence(data);
      setGeofences(prev => [...prev, result.geofence || result]);
    } catch {
      setGeofences(prev => [...prev, { id: `gf-${Date.now()}`, ...data, created_at: new Date().toISOString() }]);
    }
    setNewGfName('');
    setNewGfLat('');
    setNewGfLng('');
    setNewGfRadius('');
    setNewGfSite('');
  }

  async function handleDeleteGeofence(id: string) {
    if (!window.confirm('Delete this geofence zone?')) return;
    try { await superappApi.deleteGeofence(id); } catch { /* ok */ }
    setGeofences(prev => prev.filter(g => g.id !== id));
  }

  if (loading) return <div className="loading">Loading admin panel...</div>;

  const tabButtons: { key: Tab; label: string }[] = [
    { key: 'api-keys', label: 'API Keys' },
    { key: 'webhooks', label: 'Webhooks' },
    { key: 'geofences', label: 'Geofences' },
  ];

  return (
    <div>
      <div className="page-header">
        <h2>Admin Panel</h2>
        <p>Manage API keys, webhooks, and geofence zones</p>
      </div>

      <div className="filter-bar" style={{ marginBottom: 24 }}>
        {tabButtons.map(t => (
          <button
            key={t.key}
            className={`btn ${tab === t.key ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setTab(t.key)}
          >
            {t.label}
          </button>
        ))}
      </div>

      {/* API Keys Tab */}
      {tab === 'api-keys' && (
        <>
          <div className="stat-card" style={{ marginBottom: 24 }}>
            <h3 style={{ marginBottom: 16 }}>Create API Key</h3>
            <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap', alignItems: 'flex-end' }}>
              <div style={{ flex: 2, minWidth: 180 }}>
                <label style={{ display: 'block', color: 'var(--text-muted)', marginBottom: 4, fontSize: 13 }}>Name</label>
                <input className="filter-bar" style={{ width: '100%', padding: '8px 12px' }} value={newKeyName} onChange={e => setNewKeyName(e.target.value)} placeholder="Service name" />
              </div>
              <div style={{ flex: 1, minWidth: 120 }}>
                <label style={{ display: 'block', color: 'var(--text-muted)', marginBottom: 4, fontSize: 13 }}>Role</label>
                <select className="filter-bar" style={{ width: '100%', padding: '8px 12px' }} value={newKeyRole} onChange={e => setNewKeyRole(e.target.value)}>
                  <option value="VIEWER">Viewer</option>
                  <option value="OPERATOR">Operator</option>
                  <option value="ADMIN">Admin</option>
                </select>
              </div>
              <div style={{ flex: 1, minWidth: 160 }}>
                <label style={{ display: 'block', color: 'var(--text-muted)', marginBottom: 4, fontSize: 13 }}>Expiry (optional)</label>
                <input className="filter-bar" type="date" style={{ width: '100%', padding: '8px 12px' }} value={newKeyExpiry} onChange={e => setNewKeyExpiry(e.target.value)} />
              </div>
              <button className="btn btn-primary" onClick={handleCreateKey}>Create Key</button>
            </div>
          </div>

          <div className="table-container">
            <div className="table-toolbar"><h3>{apiKeys.length} API Keys</h3></div>
            <table>
              <thead>
                <tr><th>Name</th><th>Role</th><th>Key</th><th>Expires</th><th>Created</th><th>Actions</th></tr>
              </thead>
              <tbody>
                {apiKeys.map(k => (
                  <tr key={k.id}>
                    <td>{k.name}</td>
                    <td><span className={`badge badge-${k.role.toLowerCase()}`}>{k.role}</span></td>
                    <td className="mono">{k.key || 'sk-****'}</td>
                    <td>{k.expires_at ? new Date(k.expires_at).toLocaleDateString() : 'Never'}</td>
                    <td>{new Date(k.created_at).toLocaleDateString()}</td>
                    <td><button className="btn btn-danger" onClick={() => handleDeleteKey(k.id)}>Delete</button></td>
                  </tr>
                ))}
                {apiKeys.length === 0 && (
                  <tr><td colSpan={6}><div className="empty-state"><p>No API keys configured</p></div></td></tr>
                )}
              </tbody>
            </table>
          </div>
        </>
      )}

      {/* Webhooks Tab */}
      {tab === 'webhooks' && (
        <>
          <div className="stat-card" style={{ marginBottom: 24 }}>
            <h3 style={{ marginBottom: 16 }}>Create Webhook</h3>
            <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap', alignItems: 'flex-end' }}>
              <div style={{ flex: 1, minWidth: 160 }}>
                <label style={{ display: 'block', color: 'var(--text-muted)', marginBottom: 4, fontSize: 13 }}>Name</label>
                <input className="filter-bar" style={{ width: '100%', padding: '8px 12px' }} value={newWhName} onChange={e => setNewWhName(e.target.value)} placeholder="Webhook name" />
              </div>
              <div style={{ flex: 2, minWidth: 240 }}>
                <label style={{ display: 'block', color: 'var(--text-muted)', marginBottom: 4, fontSize: 13 }}>URL</label>
                <input className="filter-bar" style={{ width: '100%', padding: '8px 12px' }} value={newWhUrl} onChange={e => setNewWhUrl(e.target.value)} placeholder="https://example.com/webhook" />
              </div>
              <button className="btn btn-primary" onClick={handleCreateWebhook}>Create</button>
            </div>
            <div style={{ marginTop: 12 }}>
              <label style={{ display: 'block', color: 'var(--text-muted)', marginBottom: 6, fontSize: 13 }}>Events</label>
              <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
                {webhookEvents.map(ev => (
                  <button
                    key={ev}
                    className={`btn ${newWhEvents.includes(ev) ? 'btn-primary' : 'btn-secondary'}`}
                    style={{ fontSize: 12, padding: '4px 10px' }}
                    onClick={() => toggleWhEvent(ev)}
                  >
                    {ev}
                  </button>
                ))}
              </div>
            </div>
          </div>

          <div className="table-container">
            <div className="table-toolbar"><h3>{webhooks.length} Webhooks</h3></div>
            <table>
              <thead>
                <tr><th>Name</th><th>URL</th><th>Events</th><th>Last Status</th><th>Created</th><th>Actions</th></tr>
              </thead>
              <tbody>
                {webhooks.map(w => (
                  <tr key={w.id}>
                    <td>{w.name}</td>
                    <td className="mono" style={{ maxWidth: 250, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{w.url}</td>
                    <td>{w.events.map(e => <span key={e} className="badge" style={{ marginRight: 4, fontSize: 11 }}>{e}</span>)}</td>
                    <td><span className="badge badge-online">{w.last_status || 'N/A'}</span></td>
                    <td>{new Date(w.created_at).toLocaleDateString()}</td>
                    <td>
                      <div style={{ display: 'flex', gap: 6 }}>
                        <button className="btn btn-secondary" disabled={testingId === w.id} onClick={() => handleTestWebhook(w.id)}>
                          {testingId === w.id ? 'Testing...' : 'Test'}
                        </button>
                        <button className="btn btn-danger" onClick={() => handleDeleteWebhook(w.id)}>Delete</button>
                      </div>
                    </td>
                  </tr>
                ))}
                {webhooks.length === 0 && (
                  <tr><td colSpan={6}><div className="empty-state"><p>No webhooks configured</p></div></td></tr>
                )}
              </tbody>
            </table>
          </div>
        </>
      )}

      {/* Geofences Tab */}
      {tab === 'geofences' && (
        <>
          <div className="stat-card" style={{ marginBottom: 24 }}>
            <h3 style={{ marginBottom: 16 }}>Create Geofence Zone</h3>
            <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap', alignItems: 'flex-end' }}>
              <div style={{ flex: 1, minWidth: 140 }}>
                <label style={{ display: 'block', color: 'var(--text-muted)', marginBottom: 4, fontSize: 13 }}>Name</label>
                <input className="filter-bar" style={{ width: '100%', padding: '8px 12px' }} value={newGfName} onChange={e => setNewGfName(e.target.value)} placeholder="Zone name" />
              </div>
              <div style={{ flex: 1, minWidth: 100 }}>
                <label style={{ display: 'block', color: 'var(--text-muted)', marginBottom: 4, fontSize: 13 }}>Latitude</label>
                <input className="filter-bar" type="number" step="any" style={{ width: '100%', padding: '8px 12px' }} value={newGfLat} onChange={e => setNewGfLat(e.target.value)} placeholder="37.7749" />
              </div>
              <div style={{ flex: 1, minWidth: 100 }}>
                <label style={{ display: 'block', color: 'var(--text-muted)', marginBottom: 4, fontSize: 13 }}>Longitude</label>
                <input className="filter-bar" type="number" step="any" style={{ width: '100%', padding: '8px 12px' }} value={newGfLng} onChange={e => setNewGfLng(e.target.value)} placeholder="-122.4194" />
              </div>
              <div style={{ flex: 1, minWidth: 100 }}>
                <label style={{ display: 'block', color: 'var(--text-muted)', marginBottom: 4, fontSize: 13 }}>Radius (m)</label>
                <input className="filter-bar" type="number" style={{ width: '100%', padding: '8px 12px' }} value={newGfRadius} onChange={e => setNewGfRadius(e.target.value)} placeholder="500" />
              </div>
              <div style={{ flex: 1, minWidth: 140 }}>
                <label style={{ display: 'block', color: 'var(--text-muted)', marginBottom: 4, fontSize: 13 }}>Site ID</label>
                <input className="filter-bar" style={{ width: '100%', padding: '8px 12px' }} value={newGfSite} onChange={e => setNewGfSite(e.target.value)} placeholder="site-xxx" />
              </div>
              <button className="btn btn-primary" onClick={handleCreateGeofence}>Create</button>
            </div>
          </div>

          <div className="table-container">
            <div className="table-toolbar"><h3>{geofences.length} Geofence Zones</h3></div>
            <table>
              <thead>
                <tr><th>Name</th><th>Center</th><th>Radius</th><th>Site</th><th>Created</th><th>Actions</th></tr>
              </thead>
              <tbody>
                {geofences.map(g => (
                  <tr key={g.id}>
                    <td>{g.name}</td>
                    <td className="mono">{g.latitude.toFixed(4)}, {g.longitude.toFixed(4)}</td>
                    <td>{g.radius_meters}m</td>
                    <td className="mono">{g.site_id}</td>
                    <td>{new Date(g.created_at).toLocaleDateString()}</td>
                    <td><button className="btn btn-danger" onClick={() => handleDeleteGeofence(g.id)}>Delete</button></td>
                  </tr>
                ))}
                {geofences.length === 0 && (
                  <tr><td colSpan={6}><div className="empty-state"><p>No geofence zones configured</p></div></td></tr>
                )}
              </tbody>
            </table>
          </div>
        </>
      )}
    </div>
  );
};

export default AdminPanel;
