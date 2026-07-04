import { useState, useEffect } from 'react';
import {
  GetDriveConnectionStatus,
  ConnectGoogleDrive,
  DisconnectGoogleDrive,
  GetIgnorePatterns,
  AddIgnorePattern,
  RemoveIgnorePattern,
} from '../../wailsjs/go/main/App';
import ErrorBanner from '../components/ErrorBanner';

function formatError(err: unknown): string {
  if (err instanceof Error) return err.message;
  return String(err);
}

function SettingsPage(): React.JSX.Element {
  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-xl font-semibold">Settings</h1>
      <GoogleDriveSection />
      <IgnorePatternsSection />
    </div>
  );
}

function GoogleDriveSection(): React.JSX.Element {
  const [connected, setConnected] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    GetDriveConnectionStatus()
      .then(setConnected)
      .catch(() => setConnected(false))
      .finally(() => setLoading(false));
  }, []);

  async function handleConnect(): Promise<void> {
    try {
      setError('');
      const url = await ConnectGoogleDrive();
      if (url) {
        window.open(url, '_blank');
      }
    } catch (err) {
      setError(`Failed to start Google Drive connection: ${formatError(err)}`);
    }
  }

  async function handleDisconnect(): Promise<void> {
    try {
      setError('');
      await DisconnectGoogleDrive();
      setConnected(false);
    } catch (err) {
      setError(`Failed to disconnect Google Drive: ${formatError(err)}`);
    }
  }

  return (
    <div className="rounded-lg border border-neutral-800 bg-neutral-900 p-5">
      <h2 className="mb-4 text-sm font-semibold text-neutral-300 uppercase tracking-wide">
        Google Drive
      </h2>

      {error && <div className="mb-3"><ErrorBanner message={error} onDismiss={() => setError('')} /></div>}

      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div
            className={`h-2.5 w-2.5 rounded-full ${
              connected ? 'bg-green-500' : 'bg-neutral-600'
            }`}
          />
          <span className="text-sm text-neutral-300">
            {loading ? 'Checking...' : connected ? 'Connected' : 'Disconnected'}
          </span>
        </div>

        {!loading && (
          connected ? (
            <button
              type="button"
              onClick={handleDisconnect}
              className="rounded-md border border-red-800 bg-red-900/30 px-4 py-1.5 text-sm text-red-400 transition-colors hover:bg-red-900/50"
            >
              Disconnect
            </button>
          ) : (
            <button
              type="button"
              onClick={handleConnect}
              className="rounded-md bg-blue-600 px-4 py-1.5 text-sm font-medium text-white transition-colors hover:bg-blue-500"
            >
              Connect
            </button>
          )
        )}
      </div>
    </div>
  );
}

function IgnorePatternsSection(): React.JSX.Element {
  const [patterns, setPatterns] = useState<string[]>([]);
  const [newPattern, setNewPattern] = useState('');
  const [error, setError] = useState('');

  useEffect(() => {
    GetIgnorePatterns()
      .then((p) => setPatterns(p ?? []))
      .catch(() => setPatterns([]));
  }, []);

  async function handleAdd(): Promise<void> {
    const trimmed = newPattern.trim();
    if (trimmed === '') {
      setError('Pattern cannot be empty.');
      return;
    }
    try {
      await AddIgnorePattern(trimmed);
      setPatterns((prev) => [...prev, trimmed]);
      setNewPattern('');
      setError('');
    } catch (err) {
      setError(`Failed to add pattern: ${formatError(err)}`);
    }
  }

  async function handleRemove(pattern: string): Promise<void> {
    try {
      setError('');
      await RemoveIgnorePattern(pattern);
      setPatterns((prev) => prev.filter((p) => p !== pattern));
    } catch (err) {
      setError(`Failed to remove pattern: ${formatError(err)}`);
    }
  }

  function handleKeyDown(e: React.KeyboardEvent): void {
    if (e.key === 'Enter') {
      handleAdd();
    }
  }

  return (
    <div className="rounded-lg border border-neutral-800 bg-neutral-900 p-5">
      <h2 className="mb-4 text-sm font-semibold text-neutral-300 uppercase tracking-wide">
        Ignore Patterns
      </h2>

      {error && <div className="mb-3"><ErrorBanner message={error} onDismiss={() => setError('')} /></div>}

      <div className="flex flex-col gap-3">
        <div className="flex flex-col gap-2">
          {patterns.map((pattern) => (
            <div
              key={pattern}
              className="flex items-center justify-between rounded-md bg-neutral-800 px-3 py-2"
            >
              <code className="text-sm text-neutral-300">{pattern}</code>
              <button
                type="button"
                onClick={() => handleRemove(pattern)}
                className="text-xs text-neutral-500 transition-colors hover:text-red-400"
              >
                Remove
              </button>
            </div>
          ))}
          {patterns.length === 0 && (
            <p className="text-sm text-neutral-500">No ignore patterns configured.</p>
          )}
        </div>

        <div className="flex gap-2">
          <input
            type="text"
            value={newPattern}
            onChange={(e) => { setNewPattern(e.target.value); setError(''); }}
            onKeyDown={handleKeyDown}
            placeholder="e.g. node_modules, *.tmp, .git"
            className="flex-1 rounded-md border border-neutral-700 bg-neutral-800 px-3 py-1.5 text-sm text-white placeholder-neutral-500 outline-none focus:border-blue-500"
          />
          <button
            type="button"
            onClick={handleAdd}
            className="rounded-md bg-neutral-700 px-4 py-1.5 text-sm text-neutral-300 transition-colors hover:bg-neutral-600"
          >
            + Add Pattern
          </button>
        </div>
      </div>
    </div>
  );
}

export default SettingsPage;
