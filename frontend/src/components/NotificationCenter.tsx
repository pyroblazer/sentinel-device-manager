import React, { useEffect, useState } from 'react';
import { superappApi } from '../services/api';

interface Notification {
  id: string;
  type: 'INFO' | 'WARNING' | 'CRITICAL' | 'DEVICE' | 'SYSTEM';
  title: string;
  message: string;
  read: boolean;
  created_at: string;
}

const mockNotifications: Notification[] = [
  { id: 'n-001', type: 'CRITICAL', title: 'Device Offline', message: 'Camera CAM-0042 at Site-Alpha has gone offline.', read: false, created_at: '2026-04-24T10:15:00Z' },
  { id: 'n-002', type: 'WARNING', title: 'Firmware Outdated', message: '3 devices at Site-Beta are running firmware older than v2.4.0.', read: false, created_at: '2026-04-24T09:30:00Z' },
  { id: 'n-003', type: 'INFO', title: 'Firmware Deployed', message: 'Firmware v2.4.1 successfully deployed to 12 cameras.', read: true, created_at: '2026-04-23T16:00:00Z' },
  { id: 'n-004', type: 'DEVICE', title: 'New Device Registered', message: 'Access control unit AC-0198 auto-registered at Site-Gamma.', read: false, created_at: '2026-04-23T14:22:00Z' },
  { id: 'n-005', type: 'SYSTEM', title: 'Geofence Breach', message: 'Device SEN-0033 moved outside the Site-Alpha perimeter.', read: true, created_at: '2026-04-22T11:45:00Z' },
];

const typeIcons: Record<string, string> = {
  INFO: 'ℹ️',
  WARNING: '⚠️',
  CRITICAL: '🚨',
  DEVICE: '🔧',
  SYSTEM: '⚙️',
};

const NotificationCenter: React.FC = () => {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState('');

  useEffect(() => {
    loadNotifications();
  }, []);

  async function loadNotifications() {
    try {
      const data = await superappApi.listNotifications();
      setNotifications(data.notifications || data || []);
    } catch {
      // Fallback to mock data when auth is disabled / API unavailable
      setNotifications(mockNotifications);
    } finally {
      setLoading(false);
    }
  }

  async function handleMarkRead(id: string) {
    try {
      await superappApi.markRead(id);
    } catch {
      // proceed with optimistic update
    }
    setNotifications(prev => prev.map(n => n.id === id ? { ...n, read: true } : n));
  }

  async function handleDelete(id: string) {
    if (!window.confirm('Delete this notification?')) return;
    try {
      await superappApi.deleteNotification(id);
    } catch {
      // proceed with optimistic update
    }
    setNotifications(prev => prev.filter(n => n.id !== id));
  }

  const unreadCount = notifications.filter(n => !n.read).length;

  const filtered = filter
    ? notifications.filter(n => n.type === filter)
    : notifications;

  if (loading) return <div className="loading">Loading notifications...</div>;

  return (
    <div>
      <div className="page-header">
        <h2>
          Notifications
          {unreadCount > 0 && <span className="badge badge-critical" style={{ marginLeft: 12, fontSize: 14 }}>{unreadCount} unread</span>}
        </h2>
        <p>System alerts and device notifications</p>
      </div>

      <div className="filter-bar" style={{ marginBottom: 24 }}>
        <select value={filter} onChange={e => setFilter(e.target.value)}>
          <option value="">All Types</option>
          <option value="INFO">Info</option>
          <option value="WARNING">Warning</option>
          <option value="CRITICAL">Critical</option>
          <option value="DEVICE">Device</option>
          <option value="SYSTEM">System</option>
        </select>
      </div>

      {filtered.length === 0 ? (
        <div className="empty-state">
          <p>No notifications found</p>
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          {filtered.map(n => (
            <div
              key={n.id}
              className="stat-card"
              style={{
                display: 'flex',
                alignItems: 'flex-start',
                gap: 16,
                opacity: n.read ? 0.6 : 1,
              }}
            >
              <div style={{ fontSize: 24, flexShrink: 0, marginTop: 2 }}>{typeIcons[n.type] || '🔔'}</div>
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
                  <strong style={{ color: 'var(--text-primary)' }}>{n.title}</strong>
                  <span className={`badge badge-${n.type.toLowerCase()}`}>{n.type}</span>
                  {!n.read && <span style={{ width: 8, height: 8, borderRadius: '50%', background: 'var(--accent-cyan)', display: 'inline-block' }}></span>}
                </div>
                <p style={{ color: 'var(--text-secondary)', margin: '4px 0', fontSize: 14 }}>{n.message}</p>
                <span style={{ color: 'var(--text-muted)', fontSize: 12 }}>{new Date(n.created_at).toLocaleString()}</span>
              </div>
              <div style={{ display: 'flex', gap: 8, flexShrink: 0 }}>
                {!n.read && (
                  <button className="btn btn-primary" onClick={() => handleMarkRead(n.id)}>Mark Read</button>
                )}
                <button className="btn btn-danger" onClick={() => handleDelete(n.id)}>Delete</button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default NotificationCenter;
