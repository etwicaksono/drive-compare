import { useState } from 'react';
import { SelectLocalDirectory } from '../../wailsjs/go/main/App';
import ResultsTable from '../components/ResultsTable';
import ErrorBanner from '../components/ErrorBanner';
import type { CompareResult, TabKey } from '../lib/types';

interface FolderPickerProps {
  label: string;
  value: string;
  onSelect: (path: string) => void;
  onBrowse: () => void;
  placeholder: string;
  disabled: boolean;
}

function FolderPicker({ label, value, onSelect, onBrowse, placeholder, disabled }: FolderPickerProps): React.JSX.Element {
  return (
    <div className="flex flex-col gap-1.5">
      <label className="text-sm font-medium text-neutral-300">{label}</label>
      <div className="flex gap-2">
        <input
          type="text"
          value={value}
          onChange={(e) => onSelect(e.target.value)}
          placeholder={placeholder}
          className="flex-1 rounded-md border border-neutral-700 bg-neutral-800 px-3 py-2 text-sm text-white placeholder-neutral-500 outline-none focus:border-blue-500 disabled:opacity-50"
          readOnly
          disabled={disabled}
        />
        <button
          type="button"
          onClick={onBrowse}
          disabled={disabled}
          className="rounded-md border border-neutral-700 bg-neutral-800 px-4 py-2 text-sm text-neutral-300 transition-colors hover:bg-neutral-700 disabled:cursor-not-allowed disabled:opacity-50"
        >
          Browse
        </button>
      </div>
    </div>
  );
}

function formatError(err: unknown): string {
  if (err instanceof Error) return err.message;
  return String(err);
}

function ComparePage(): React.JSX.Element {
  const [localDir, setLocalDir] = useState('');
  const [driveDir, setDriveDir] = useState('');
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<CompareResult | null>(null);
  const [activeTab, setActiveTab] = useState<TabKey>('localOnly');
  const [error, setError] = useState('');

  const canCompare = localDir !== '' && driveDir !== '' && !loading;

  async function handleBrowseLocal(): Promise<void> {
    try {
      setError('');
      const dir = await SelectLocalDirectory();
      if (dir) {
        setLocalDir(dir);
      }
    } catch (err) {
      setError(`Failed to select directory: ${formatError(err)}`);
    }
  }

  function handleBrowseDrive(): void {
    setError('Google Drive folder picker is not yet available. Please enter the folder ID manually.');
  }

  function handleCompare(): void {
    setLoading(true);
    setResult(null);
    setError('');

    setTimeout(() => {
      setResult({ localOnly: [], driveOnly: [], both: [] });
      setLoading(false);
    }, 100);
  }

  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-xl font-semibold">Compare</h1>

      {error && <ErrorBanner message={error} onDismiss={() => setError('')} />}

      <div className="flex flex-col gap-4 rounded-lg border border-neutral-800 bg-neutral-900 p-5">
        <FolderPicker
          label="Local Directory"
          value={localDir}
          onSelect={setLocalDir}
          onBrowse={handleBrowseLocal}
          placeholder="Select a local folder..."
          disabled={loading}
        />

        <FolderPicker
          label="Google Drive Directory"
          value={driveDir}
          onSelect={setDriveDir}
          onBrowse={handleBrowseDrive}
          placeholder="Enter Drive folder ID..."
          disabled={loading}
        />

        <button
          type="button"
          onClick={handleCompare}
          disabled={!canCompare}
          className="mt-1 w-full rounded-md bg-blue-600 px-4 py-2.5 text-sm font-medium text-white transition-colors hover:bg-blue-500 disabled:cursor-not-allowed disabled:opacity-40"
        >
          {loading ? 'Scanning & Comparing...' : 'Compare'}
        </button>
      </div>

      {loading && (
        <div className="flex items-center justify-center gap-3 rounded-lg border border-neutral-800 bg-neutral-900 p-8">
          <div className="h-5 w-5 animate-spin rounded-full border-2 border-neutral-600 border-t-blue-500" />
          <span className="text-sm text-neutral-400">Scanning directories and comparing files...</span>
        </div>
      )}

      {result && !loading && (
        <ResultsTable
          result={result}
          activeTab={activeTab}
          onTabChange={setActiveTab}
        />
      )}
    </div>
  );
}

export default ComparePage;
