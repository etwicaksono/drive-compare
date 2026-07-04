import { useState, useEffect } from 'react';
import {
  GetDriveConnectionStatus,
  ConnectGoogleDrive,
  DisconnectGoogleDrive,
} from '../../wailsjs/go/main/App';

function SettingsPage(): React.JSX.Element {
  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-xl font-semibold">Settings</h1>
      <GoogleDriveSection />
    </div>
  );
}

function GoogleDriveSection(): React.JSX.Element {
  const [connected, setConnected] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    GetDriveConnectionStatus()
      .then(setConnected)
      .catch(() => setConnected(false))
      .finally(() => setLoading(false));
  }, []);

  async function handleConnect(): Promise<void> {
    try {
      const url = await ConnectGoogleDrive();
      window.open(url, '_blank');
    } catch (err) {
      console.error('Failed to start OAuth flow:', err);
    }
  }

  async function handleDisconnect(): Promise<void> {
    try {
      await DisconnectGoogleDrive();
      setConnected(false);
    } catch (err) {
      console.error('Failed to disconnect:', err);
    }
  }

  return (
    <div className="rounded-lg border border-neutral-800 bg-neutral-900 p-5">
      <h2 className="mb-4 text-sm font-semibold text-neutral-300 uppercase tracking-wide">
        Google Drive
      </h2>

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

export default SettingsPage;
