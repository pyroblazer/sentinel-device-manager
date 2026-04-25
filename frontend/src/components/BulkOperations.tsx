import React, { useState } from 'react';
import { superappApi } from '../services/api';

type Action = 'delete' | 'update' | 'tag' | 'export-json' | 'export-csv' | null;

const BulkOperations: React.FC = () => {
  const [deviceIdsText, setDeviceIdsText] = useState('');
  const [action, setAction] = useState<Action>(null);
  const [busy, setBusy] = useState(false);
  const [result, setResult] = useState<string | null>(null);

  // Bulk update form
  const [updatePairs, setUpdatePairs] = useState<{ key: string; value: string }[]>([{ key: '', value: '' }]);

  // Bulk tag form
  const [tagsText, setTagsText] = useState('');

  function parseDeviceIds(): string[] {
    return deviceIdsText
      .split('\n')
      .map(line => line.trim())
      .filter(line => line.length > 0);
  }

  function addUpdatePair() {
    setUpdatePairs(prev => [...prev, { key: '', value: '' }]);
  }

  function removeUpdatePair(index: number) {
    setUpdatePairs(prev => prev.filter((_, i) => i !== index));
  }

  function updateUpdatePair(index: number, field: 'key' | 'value', val: string) {
    setUpdatePairs(prev => prev.map((p, i) => i === index ? { ...p, [field]: val } : p));
  }

  async function executeAction() {
    const ids = parseDeviceIds();
    if (ids.length === 0) {
      setResult('No device IDs provided. Enter one ID per line.');
      return;
    }

    setBusy(true);
    setResult(null);

    try {
      switch (action) {
        case 'delete': {
          if (!window.confirm(`Delete ${ids.length} device(s)? This cannot be undone.`)) { setBusy(false); return; }
          const res = await superappApi.bulkDelete(ids);
          setResult(`Deleted ${ids.length} device(s).`);
          break;
        }
        case 'update': {
          const updates: Record<string, string> = {};
          updatePairs.forEach(p => { if (p.key.trim()) updates[p.key.trim()] = p.value; });
          if (Object.keys(updates).length === 0) { setResult('No update fields specified.'); setBusy(false); return; }
          const res = await superappApi.bulkUpdate(ids, updates);
          setResult(`Updated ${ids.length} device(s) with ${Object.keys(updates).length} field(s).`);
          break;
        }
        case 'tag': {
          const tags = tagsText.split(',').map(t => t.trim()).filter(t => t.length > 0);
          if (tags.length === 0) { setResult('No tags specified.'); setBusy(false); return; }
          const res = await superappApi.bulkTag(ids, tags);
          setResult(`Applied ${tags.length} tag(s) to ${ids.length} device(s).`);
          break;
        }
        case 'export-json': {
          const url = superappApi.exportDevices('json');
          window.open(url, '_blank');
          setResult(`Exported ${ids.length} device(s) as JSON. Download started.`);
          break;
        }
        case 'export-csv': {
          const url = superappApi.exportDevices('csv');
          window.open(url, '_blank');
          setResult(`Exported ${ids.length} device(s) as CSV. Download started.`);
          break;
        }
      }
    } catch (err: any) {
      setResult(`Error: ${err.message}`);
    } finally {
      setBusy(false);
    }
  }

  function selectAction(a: Action) {
    setAction(prev => prev === a ? null : a);
    setResult(null);
  }

  const ids = parseDeviceIds();

  return (
    <div>
      <div className="page-header">
        <h2>Bulk Operations</h2>
        <p>Perform batch actions across multiple devices</p>
      </div>

      <div className="stat-card" style={{ marginBottom: 24 }}>
        <h3 style={{ marginBottom: 12 }}>Device IDs</h3>
        <textarea
          style={{
            width: '100%',
            minHeight: 140,
            padding: 12,
            background: 'var(--bg-primary)',
            border: '1px solid var(--border)',
            borderRadius: 6,
            color: 'var(--text-primary)',
            fontFamily: 'monospace',
            fontSize: 13,
            resize: 'vertical',
          }}
          placeholder="Enter device IDs, one per line&#10;dev-001&#10;dev-002&#10;dev-003"
          value={deviceIdsText}
          onChange={e => setDeviceIdsText(e.target.value)}
        />
        <p style={{ color: 'var(--text-muted)', marginTop: 8, fontSize: 13 }}>
          {ids.length} device(s) entered
        </p>
      </div>

      <div style={{ marginBottom: 24 }}>
        <h3 style={{ marginBottom: 12, color: 'var(--text-primary)' }}>Actions</h3>
        <div style={{ display: 'flex', gap: 10, flexWrap: 'wrap' }}>
          <button
            className={`btn ${action === 'delete' ? 'btn-danger' : 'btn-secondary'}`}
            onClick={() => selectAction('delete')}
          >
            Bulk Delete
          </button>
          <button
            className={`btn ${action === 'update' ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => selectAction('update')}
          >
            Bulk Update
          </button>
          <button
            className={`btn ${action === 'tag' ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => selectAction('tag')}
          >
            Bulk Tag
          </button>
          <button
            className={`btn ${action === 'export-json' ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => selectAction('export-json')}
          >
            Export JSON
          </button>
          <button
            className={`btn ${action === 'export-csv' ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => selectAction('export-csv')}
          >
            Export CSV
          </button>
        </div>
      </div>

      {/* Action-specific forms */}
      {action === 'update' && (
        <div className="stat-card" style={{ marginBottom: 24 }}>
          <h3 style={{ marginBottom: 12 }}>Update Fields</h3>
          {updatePairs.map((pair, i) => (
            <div key={i} style={{ display: 'flex', gap: 8, marginBottom: 8, alignItems: 'center' }}>
              <input
                className="filter-bar"
                style={{ flex: 1, padding: '6px 10px' }}
                placeholder="Field name (e.g. firmware_version)"
                value={pair.key}
                onChange={e => updateUpdatePair(i, 'key', e.target.value)}
              />
              <span style={{ color: 'var(--text-muted)' }}>=</span>
              <input
                className="filter-bar"
                style={{ flex: 1, padding: '6px 10px' }}
                placeholder="Value"
                value={pair.value}
                onChange={e => updateUpdatePair(i, 'value', e.target.value)}
              />
              {updatePairs.length > 1 && (
                <button className="btn btn-danger" style={{ padding: '4px 10px', fontSize: 12 }} onClick={() => removeUpdatePair(i)}>X</button>
              )}
            </div>
          ))}
          <button className="btn btn-secondary" style={{ fontSize: 13 }} onClick={addUpdatePair}>+ Add Field</button>
        </div>
      )}

      {action === 'tag' && (
        <div className="stat-card" style={{ marginBottom: 24 }}>
          <h3 style={{ marginBottom: 12 }}>Tags</h3>
          <input
            className="filter-bar"
            style={{ width: '100%', padding: '8px 12px' }}
            placeholder="Enter tags separated by commas (e.g. maintenance, priority, site-alpha)"
            value={tagsText}
            onChange={e => setTagsText(e.target.value)}
          />
        </div>
      )}

      {/* Execute button */}
      {action && (
        <div style={{ marginBottom: 24 }}>
          <button
            className="btn btn-primary"
            disabled={busy}
            onClick={executeAction}
          >
            {busy ? 'Processing...' : `Execute: ${action.replace('-', ' ').toUpperCase()}`}
          </button>
        </div>
      )}

      {/* Result */}
      {result && (
        <div className="stat-card" style={{
          borderLeft: result.startsWith('Error')
            ? '4px solid var(--accent-red)'
            : '4px solid var(--accent-cyan)',
        }}>
          <p style={{ color: 'var(--text-primary)', fontSize: 14 }}>{result}</p>
        </div>
      )}
    </div>
  );
};

export default BulkOperations;
