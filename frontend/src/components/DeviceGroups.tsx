import React, { useEffect, useState } from 'react';
import { superappApi } from '../services/api';

interface DeviceGroup {
  id: string;
  name: string;
  type: 'ZONE' | 'BUILDING' | 'FLOOR' | 'CUSTOM';
  site_id: string;
  device_ids: string[];
  created_at: string;
}

const mockGroups: DeviceGroup[] = [
  { id: 'grp-001', name: 'North Wing Cameras', type: 'ZONE', site_id: 'site-alpha-001', device_ids: ['dev-001', 'dev-002', 'dev-003', 'dev-004'], created_at: '2026-03-01T10:00:00Z' },
  { id: 'grp-002', name: 'HQ Building A', type: 'BUILDING', site_id: 'site-beta-002', device_ids: ['dev-005', 'dev-006', 'dev-007'], created_at: '2026-03-10T09:00:00Z' },
  { id: 'grp-003', name: 'Floor 3 Sensors', type: 'FLOOR', site_id: 'site-alpha-001', device_ids: ['dev-008', 'dev-009'], created_at: '2026-04-01T14:00:00Z' },
  { id: 'grp-004', name: 'Perimeter Alarms', type: 'CUSTOM', site_id: 'site-gamma-003', device_ids: ['dev-010', 'dev-011', 'dev-012'], created_at: '2026-04-15T08:00:00Z' },
];

const DeviceGroups: React.FC = () => {
  const [groups, setGroups] = useState<DeviceGroup[]>([]);
  const [loading, setLoading] = useState(true);
  const [expandedId, setExpandedId] = useState<string | null>(null);

  // Form state
  const [showForm, setShowForm] = useState(false);
  const [newName, setNewName] = useState('');
  const [newType, setNewType] = useState<'ZONE' | 'BUILDING' | 'FLOOR' | 'CUSTOM'>('ZONE');
  const [newSite, setNewSite] = useState('');

  useEffect(() => {
    loadGroups();
  }, []);

  async function loadGroups() {
    try {
      const data = await superappApi.listGroups();
      setGroups(data.groups || data || []);
    } catch {
      setGroups(mockGroups);
    } finally {
      setLoading(false);
    }
  }

  async function handleCreate() {
    if (!newName.trim() || !newSite.trim()) return;
    const payload = { name: newName, type: newType, site_id: newSite };
    try {
      const result = await superappApi.createGroup(payload);
      setGroups(prev => [...prev, result.group || result]);
    } catch {
      setGroups(prev => [...prev, { id: `grp-${Date.now()}`, ...payload, device_ids: [], created_at: new Date().toISOString() }]);
    }
    setNewName('');
    setNewType('ZONE');
    setNewSite('');
    setShowForm(false);
  }

  async function handleDelete(id: string) {
    if (!window.confirm('Delete this group?')) return;
    try { await superappApi.deleteGroup(id); } catch { /* ok */ }
    setGroups(prev => prev.filter(g => g.id !== id));
    if (expandedId === id) setExpandedId(null);
  }

  function toggleExpand(id: string) {
    setExpandedId(prev => prev === id ? null : id);
  }

  if (loading) return <div className="loading">Loading groups...</div>;

  return (
    <div>
      <div className="page-header">
        <h2>Device Groups</h2>
        <p>Organize devices into zones, buildings, floors, and custom groups</p>
      </div>

      <div style={{ marginBottom: 24 }}>
        <button className="btn btn-primary" onClick={() => setShowForm(!showForm)}>
          {showForm ? 'Cancel' : 'Create Group'}
        </button>
      </div>

      {showForm && (
        <div className="stat-card" style={{ marginBottom: 24 }}>
          <h3 style={{ marginBottom: 16 }}>New Group</h3>
          <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap', alignItems: 'flex-end' }}>
            <div style={{ flex: 2, minWidth: 180 }}>
              <label style={{ display: 'block', color: 'var(--text-muted)', marginBottom: 4, fontSize: 13 }}>Name</label>
              <input className="filter-bar" style={{ width: '100%', padding: '8px 12px' }} value={newName} onChange={e => setNewName(e.target.value)} placeholder="Group name" />
            </div>
            <div style={{ flex: 1, minWidth: 140 }}>
              <label style={{ display: 'block', color: 'var(--text-muted)', marginBottom: 4, fontSize: 13 }}>Type</label>
              <select className="filter-bar" style={{ width: '100%', padding: '8px 12px' }} value={newType} onChange={e => setNewType(e.target.value as any)}>
                <option value="ZONE">Zone</option>
                <option value="BUILDING">Building</option>
                <option value="FLOOR">Floor</option>
                <option value="CUSTOM">Custom</option>
              </select>
            </div>
            <div style={{ flex: 2, minWidth: 180 }}>
              <label style={{ display: 'block', color: 'var(--text-muted)', marginBottom: 4, fontSize: 13 }}>Site ID</label>
              <input className="filter-bar" style={{ width: '100%', padding: '8px 12px' }} value={newSite} onChange={e => setNewSite(e.target.value)} placeholder="site-xxx" />
            </div>
            <button className="btn btn-primary" onClick={handleCreate}>Create</button>
          </div>
        </div>
      )}

      {groups.length === 0 ? (
        <div className="empty-state"><p>No device groups created yet</p></div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          {groups.map(g => (
            <div key={g.id} className="stat-card">
              <div
                style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', cursor: 'pointer' }}
                onClick={() => toggleExpand(g.id)}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                  <span style={{ color: 'var(--accent-cyan)', fontSize: 18 }}>{expandedId === g.id ? '▼' : '▶'}</span>
                  <div>
                    <strong style={{ color: 'var(--text-primary)', fontSize: 16 }}>{g.name}</strong>
                    <div style={{ display: 'flex', gap: 8, marginTop: 4 }}>
                      <span className={`badge badge-${g.type.toLowerCase()}`}>{g.type}</span>
                      <span className="badge badge-online">{g.device_ids.length} device{g.device_ids.length !== 1 ? 's' : ''}</span>
                      <span style={{ color: 'var(--text-muted)', fontSize: 12 }}>Site: {g.site_id}</span>
                    </div>
                  </div>
                </div>
                <button className="btn btn-danger" onClick={e => { e.stopPropagation(); handleDelete(g.id); }}>Delete</button>
              </div>
              {expandedId === g.id && (
                <div style={{ marginTop: 16, paddingTop: 16, borderTop: '1px solid var(--border)' }}>
                  <h4 style={{ color: 'var(--text-secondary)', marginBottom: 8 }}>Devices in this group</h4>
                  {g.device_ids.length === 0 ? (
                    <p style={{ color: 'var(--text-muted)', fontSize: 14 }}>No devices assigned to this group.</p>
                  ) : (
                    <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
                      {g.device_ids.map(did => (
                        <span key={did} className="badge" style={{ fontSize: 13, padding: '4px 10px', background: 'var(--bg-tertiary)', color: 'var(--accent-cyan)' }}>{did}</span>
                      ))}
                    </div>
                  )}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default DeviceGroups;
