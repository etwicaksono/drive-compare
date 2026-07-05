import { useState, useEffect } from 'react';
import { SelectLocalDirectory, GetDuplicatePatterns, RunDuplicateScan } from '../../wailsjs/go/main/App';
import DriveFolderPicker from '../components/DriveFolderPicker';
import DuplicateResultsTable from '../components/DuplicateResultsTable';
import ErrorBanner from '../components/ErrorBanner';
import type { DuplicatePatternInfo, DuplicateScanResult } from '../lib/types';

function formatError(err: unknown): string {
  if (err instanceof Error) return err.message;
  return String(err);
}

function DuplicatesPage(): React.JSX.Element {
  const [sourceType, setSourceType] = useState<'local' | 'drive'>('local');
  const [localDir, setLocalDir] = useState('');
  const [driveDir, setDriveDir] = useState('');
  const [driveDirLabel, setDriveDirLabel] = useState('');
  const [showDrivePicker, setShowDrivePicker] = useState(false);

  const [patterns, setPatterns] = useState<DuplicatePatternInfo[]>([]);
  const [selectedPatterns, setSelectedPatterns] = useState<Set<string>>(new Set());

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [result, setResult] = useState<DuplicateScanResult | null>(null);

  useEffect(() => {
    GetDuplicatePatterns()
      .then((p) => {
        const list: DuplicatePatternInfo[] = (p ?? []).map((item: { id: string; label: string; example: string }) => ({
          id: item.id,
          label: item.label,
          example: item.example,
        }));
        setPatterns(list);
        setSelectedPatterns(new Set(list.map((x) => x.id)));
      })
      .catch(() => setPatterns([]));
  }, []);

  const sourceID = sourceType === 'local' ? localDir : driveDir;
  const canScan = sourceID !== '' && selectedPatterns.size > 0 && !loading;

  async function handleBrowseLocal(): Promise<void> {
    try {
      setError('');
      const dir = await SelectLocalDirectory();
      if (dir) setLocalDir(dir);
    } catch (err) {
      setError(`Failed to select directory: ${formatError(err)}`);
    }
  }

  function handleDriveSelect(folderId: string, folderName: string): void {
    setDriveDir(folderId);
    setDriveDirLabel(folderName);
    setShowDrivePicker(false);
  }

  function togglePattern(id: string): void {
    setSelectedPatterns((prev) => {
      const next = new Set(prev);
      if (next.has(id)) { next.delete(id); } else { next.add(id); }
      return next;
    });
  }

  async function handleScan(): Promise<void> {
    setLoading(true);
    setResult(null);
    setError('');

    try {
      const resp = await RunDuplicateScan(sourceType, sourceID, Array.from(selectedPatterns));
      setResult({
        groups: (resp.groups ?? []).map((g: { baseName: string; extension: string; relativePath: string; files: { relativePath: string; name: string; extension: string; size: number }[] }) => ({
          baseName: g.baseName,
          extension: g.extension,
          relativePath: g.relativePath,
          files: g.files ?? [],
        })),
        totalGroups: resp.totalGroups,
        totalFiles: resp.totalFiles,
      });
    } catch (err) {
      setError(`Scan failed: ${formatError(err)}`);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-xl font-semibold">Duplicates</h1>

      {error && <ErrorBanner message={error} onDismiss={() => setError('')} />}

      <div className="flex flex-col gap-4 rounded-lg border border-neutral-800 bg-neutral-900 p-5">
        <div className="flex flex-col gap-2">
          <span className="text-sm font-medium text-neutral-300">Scan Source</span>
          <div className="flex gap-4">
            <label className="flex items-center gap-2 cursor-pointer">
              <input type="radio" name="sourceType" checked={sourceType === 'local'} onChange={() => setSourceType('local')} className="accent-blue-500" />
              <span className="text-sm text-neutral-300">Local Directory</span>
            </label>
            <label className="flex items-center gap-2 cursor-pointer">
              <input type="radio" name="sourceType" checked={sourceType === 'drive'} onChange={() => setSourceType('drive')} className="accent-blue-500" />
              <span className="text-sm text-neutral-300">Google Drive</span>
            </label>
          </div>
        </div>

        {sourceType === 'local' ? (
          <div className="flex gap-2">
            <input
              type="text"
              value={localDir}
              placeholder="Select a local folder..."
              className="flex-1 rounded-md border border-neutral-700 bg-neutral-800 px-3 py-2 text-sm text-white placeholder-neutral-500 outline-none focus:border-blue-500 disabled:opacity-50"
              readOnly
              disabled={loading}
            />
            <button type="button" onClick={handleBrowseLocal} disabled={loading} className="rounded-md border border-neutral-700 bg-neutral-800 px-4 py-2 text-sm text-neutral-300 transition-colors hover:bg-neutral-700 disabled:cursor-not-allowed disabled:opacity-50">
              Browse
            </button>
          </div>
        ) : (
          <div className="flex gap-2">
            <input
              type="text"
              value={driveDirLabel}
              placeholder="Select a Drive folder..."
              className="flex-1 rounded-md border border-neutral-700 bg-neutral-800 px-3 py-2 text-sm text-white placeholder-neutral-500 outline-none focus:border-blue-500 disabled:opacity-50"
              readOnly
              disabled={loading}
            />
            <button type="button" onClick={() => setShowDrivePicker(true)} disabled={loading} className="rounded-md border border-neutral-700 bg-neutral-800 px-4 py-2 text-sm text-neutral-300 transition-colors hover:bg-neutral-700 disabled:cursor-not-allowed disabled:opacity-50">
              Browse
            </button>
          </div>
        )}

        <div className="flex flex-col gap-2">
          <span className="text-sm font-medium text-neutral-300">Duplication Patterns</span>
          <div className="flex flex-col gap-1.5">
            {patterns.map((p) => (
              <label key={p.id} className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={selectedPatterns.has(p.id)}
                  onChange={() => togglePattern(p.id)}
                  className="accent-blue-500"
                />
                <span className="text-sm text-neutral-300">{p.label}</span>
                <span className="text-xs text-neutral-500">({p.example})</span>
              </label>
            ))}
          </div>
        </div>

        <button
          type="button"
          onClick={handleScan}
          disabled={!canScan}
          className="mt-1 w-full rounded-md bg-blue-600 px-4 py-2.5 text-sm font-medium text-white transition-colors hover:bg-blue-500 disabled:cursor-not-allowed disabled:opacity-40"
        >
          {loading ? 'Scanning...' : 'Scan for Duplicates'}
        </button>
      </div>

      {loading && (
        <div className="flex items-center justify-center gap-3 rounded-lg border border-neutral-800 bg-neutral-900 p-8">
          <div className="h-5 w-5 animate-spin rounded-full border-2 border-neutral-600 border-t-blue-500" />
          <span className="text-sm text-neutral-400">Scanning for duplicates...</span>
        </div>
      )}

      {result && !loading && (
        <DuplicateResultsTable
          groups={result.groups}
          totalGroups={result.totalGroups}
          totalFiles={result.totalFiles}
        />
      )}

      <DriveFolderPicker
        open={showDrivePicker}
        onSelect={handleDriveSelect}
        onClose={() => setShowDrivePicker(false)}
      />
    </div>
  );
}

export default DuplicatesPage;
