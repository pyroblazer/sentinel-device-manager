import React, { useState } from 'react';

interface FirmwareVersion {
  version: string;
  device_type: string;
  release_notes: string;
  created_at: string;
}

const mockFirmware: FirmwareVersion[] = [
  { version: '2.4.1', device_type: 'CAMERA', release_notes: 'Improved night vision, reduced CPU usage', created_at: '2026-04-20T00:00:00Z' },
  { version: '2.4.0', device_type: 'CAMERA', release_notes: 'Added H.265 encoding support', created_at: '2026-04-01T00:00:00Z' },
  { version: '1.8.3', device_type: 'ACCESS_CONTROL', release_notes: 'Fixed badge reader timeout issue', created_at: '2026-04-15T00:00:00Z' },
  { version: '1.8.2', device_type: 'ACCESS_CONTROL', release_notes: 'Added multi-factor authentication support', created_at: '2026-03-20T00:00:00Z' },
  { version: '3.1.0', device_type: 'ALARM', release_notes: 'New siren patterns, improved zone detection', created_at: '2026-04-10T00:00:00Z' },
  { version: '2.0.5', device_type: 'SENSOR', release_notes: 'Temperature calibration fix', created_at: '2026-04-05T00:00:00Z' },
];

const FirmwareManager: React.FC = () => {
  const [selectedType, setSelectedType] = useState('');
  const [deploying, setDeploying] = useState(false);

  const filtered = selectedType
    ? mockFirmware.filter(f => f.device_type === selectedType)
    : mockFirmware;

  async function handleDeploy(version: string) {
    if (!window.confirm(`Deploy firmware ${version} to all compatible devices?`)) return;
    setDeploying(true);
    // In production, call deviceApi.deployFirmware(version, deviceIds)
    setTimeout(() => {
      setDeploying(false);
      alert(`Firmware ${version} deployment initiated`);
    }, 1500);
  }

  return (
    <div>
      <div className="page-header">
        <h2>Firmware Management</h2>
        <p>Deploy and manage firmware versions across your device fleet</p>
      </div>

      <div className="filter-bar" style={{ marginBottom: 24 }}>
        <select value={selectedType} onChange={e => setSelectedType(e.target.value)}>
          <option value="">All Device Types</option>
          <option value="CAMERA">Camera</option>
          <option value="ACCESS_CONTROL">Access Control</option>
          <option value="ALARM">Alarm</option>
          <option value="SENSOR">Sensor</option>
        </select>
      </div>

      <div className="table-container">
        <div className="table-toolbar">
          <h3>Firmware Versions</h3>
        </div>
        <table>
          <thead>
            <tr>
              <th>Version</th>
              <th>Device Type</th>
              <th>Release Notes</th>
              <th>Released</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {filtered.map(fw => (
              <tr key={`${fw.version}-${fw.device_type}`}>
                <td className="mono" style={{ color: 'var(--accent-cyan)', fontWeight: 700 }}>{fw.version}</td>
                <td><span className={`badge badge-${fw.device_type.toLowerCase()}`}>{fw.device_type}</span></td>
                <td style={{ maxWidth: 400 }}>{fw.release_notes}</td>
                <td>{new Date(fw.created_at).toLocaleDateString()}</td>
                <td>
                  <button
                    className="btn btn-primary"
                    disabled={deploying}
                    onClick={() => handleDeploy(fw.version)}
                  >
                    {deploying ? 'Deploying...' : 'Deploy'}
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default FirmwareManager;
