import React from 'react';
import { Routes, Route, NavLink } from 'react-router-dom';
import { useAuth } from './contexts/AuthContext';
import LoginPage from './components/LoginPage';
import Dashboard from './components/Dashboard';
import DeviceList from './components/DeviceList';
import DeviceDetail from './components/DeviceDetail';
import AlertViewer from './components/AlertViewer';
import FirmwareManager from './components/FirmwareManager';
import NotificationCenter from './components/NotificationCenter';
import DeviceGroups from './components/DeviceGroups';
import ConfigTemplates from './components/ConfigTemplates';
import BulkOperations from './components/BulkOperations';
import AdminPanel from './components/AdminPanel';

const App: React.FC = () => {
  const { isAuthenticated, user, logout } = useAuth();

  if (!isAuthenticated) return <LoginPage />;

  return (
    <div className="app">
      <nav className="sidebar">
        <div className="sidebar-logo">
          <h1>Sentinel</h1>
          <span className="sidebar-subtitle">Device Manager</span>
        </div>
        <ul className="nav-links">
          <li>
            <NavLink to="/" end className={({ isActive }) => isActive ? 'nav-link active' : 'nav-link'}>
              Dashboard
            </NavLink>
          </li>
          <li>
            <NavLink to="/devices" className={({ isActive }) => isActive ? 'nav-link active' : 'nav-link'}>
              Devices
            </NavLink>
          </li>
          <li>
            <NavLink to="/alerts" className={({ isActive }) => isActive ? 'nav-link active' : 'nav-link'}>
              Alerts
            </NavLink>
          </li>
          <li>
            <NavLink to="/firmware" className={({ isActive }) => isActive ? 'nav-link active' : 'nav-link'}>
              Firmware
            </NavLink>
          </li>
          <li>
            <NavLink to="/notifications" className={({ isActive }) => isActive ? 'nav-link active' : 'nav-link'}>
              Notifications
            </NavLink>
          </li>
          <li>
            <NavLink to="/groups" className={({ isActive }) => isActive ? 'nav-link active' : 'nav-link'}>
              Groups
            </NavLink>
          </li>
          <li>
            <NavLink to="/templates" className={({ isActive }) => isActive ? 'nav-link active' : 'nav-link'}>
              Templates
            </NavLink>
          </li>
          <li>
            <NavLink to="/bulk" className={({ isActive }) => isActive ? 'nav-link active' : 'nav-link'}>
              Bulk Ops
            </NavLink>
          </li>
          <li>
            <NavLink to="/admin" className={({ isActive }) => isActive ? 'nav-link active' : 'nav-link'}>
              Admin
            </NavLink>
          </li>
        </ul>
        <div style={{
          marginTop: 'auto',
          paddingTop: 16,
          borderTop: '1px solid var(--border)',
        }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
            <div style={{
              width: 32,
              height: 32,
              borderRadius: '50%',
              background: 'var(--accent-blue)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              color: '#fff',
              fontWeight: 700,
              fontSize: 14,
            }}>
              {user?.name?.charAt(0) || 'U'}
            </div>
            <div>
              <div style={{ color: 'var(--text-primary)', fontSize: 13, fontWeight: 600 }}>{user?.name}</div>
              <span className={`badge badge-${user?.role === 'admin' ? 'online' : user?.role === 'operator' ? 'warning' : 'info'}`} style={{ fontSize: 10, padding: '2px 6px' }}>
                {user?.role}
              </span>
            </div>
          </div>
          <button
            className="btn btn-secondary"
            style={{ width: '100%', fontSize: 12, padding: '6px 12px' }}
            onClick={logout}
          >
            Sign Out
          </button>
        </div>
      </nav>
      <main className="main-content">
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/devices" element={<DeviceList />} />
          <Route path="/devices/:deviceId" element={<DeviceDetail />} />
          <Route path="/alerts" element={<AlertViewer />} />
          <Route path="/firmware" element={<FirmwareManager />} />
          <Route path="/notifications" element={<NotificationCenter />} />
          <Route path="/groups" element={<DeviceGroups />} />
          <Route path="/templates" element={<ConfigTemplates />} />
          <Route path="/bulk" element={<BulkOperations />} />
          <Route path="/admin" element={<AdminPanel />} />
        </Routes>
      </main>
    </div>
  );
};

export default App;
