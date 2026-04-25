import React from 'react';
import { Routes, Route, NavLink } from 'react-router-dom';
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
