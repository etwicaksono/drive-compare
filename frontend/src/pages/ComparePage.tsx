import { useState } from 'react';
import { SelectLocalDirectory, RunComparison, ExportToClipboard, ExportToFile } from '../../wailsjs/go/main/App';
import ResultsTable from '../components/ResultsTable';
import ErrorBanner from '../components/ErrorBanner';
import DriveFolderPicker from '../components/DriveFolderPicker';
import ExportModal from '../components/ExportModal';
import type { CompareResult, TabKey, FileItem } from '../lib/types';
import { getItemKey, buildExportText } from '../lib/types';

function formatError(err: unknown): string {
  if (err instanceof Error) return err.message;
  return String(err);
}

const EMPTY_SELECTION: Record<TabKey, Set<string>> = {
  localOnly: new Set(),
  driveOnly: new Set(),
  both: new Set(),
};

function ComparePage(): React.JSX.Element {
  const [localDir, setLocalDir] = useState('');
  const [driveDir, setDriveDir] = useState('');
  const [driveDirLabel, setDriveDirLabel] = useState('');
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<CompareResult | null>(null);
  const [activeTab, setActiveTab] = useState<TabKey>('localOnly');
  const [error, setError] = useState('');
  const [successMsg, setSuccessMsg] = useState('');
  const [showDrivePicker, setShowDrivePicker] = useState(false);
  const [showExportModal, setShowExportModal] = useState(false);
  const [selections, setSelections] = useState<Record<TabKey, Set<string>>>({ ...EMPTY_SELECTION });

  const canCompare = localDir !== '' && driveDir !== '' && !loading;
  const currentSelection = selections[activeTab];

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

  function handleDriveSelect(folderId: string, folderName: string): void {
    setDriveDir(folderId);
    setDriveDirLabel(folderName);
    setShowDrivePicker(false);
  }

  async function handleCompare(): Promise<void> {
    setLoading(true);
    setResult(null);
    setError('');
    setSuccessMsg('');
    setSelections({ localOnly: new Set(), driveOnly: new Set(), both: new Set() });

    try {
      const resp = await RunComparison(localDir, driveDir);
      setResult({
        localOnly: resp.localOnly ?? [],
        driveOnly: resp.driveOnly ?? [],
        both: resp.both ?? [],
      });
    } catch (err) {
      setError(`Comparison failed: ${formatError(err)}`);
    } finally {
      setLoading(false);
    }
  }

  function handleSelectionChange(keys: Set<string>): void {
    setSelections((prev) => ({ ...prev, [activeTab]: keys }));
  }

  async function handleExport(pathSource: 'local' | 'drive', target: 'clipboard' | 'file', _contentType: 'items' | 'headers'): Promise<void> {
    if (!result) return;

    const items = result[activeTab];
    const selectedItems: FileItem[] = items.filter((item) => currentSelection.has(getItemKey(item)));

    if (selectedItems.length === 0) return;

    const text = buildExportText(selectedItems, pathSource, localDir, driveDirLabel);

    try {
      if (target === 'clipboard') {
        await ExportToClipboard(text);
        setSuccessMsg(`Exported ${selectedItems.length} file path${selectedItems.length !== 1 ? 's' : ''} to clipboard`);
      } else {
        const path = await ExportToFile(text);
        if (path) {
          setSuccessMsg(`Saved ${selectedItems.length} file path${selectedItems.length !== 1 ? 's' : ''} to ${path}`);
        }
      }
    } catch (err) {
      setError(`Export failed: ${formatError(err)}`);
    }

    setShowExportModal(false);
  }

  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-xl font-semibold">Compare</h1>

      {error && <ErrorBanner message={error} onDismiss={() => setError('')} />}

      {successMsg && (
        <div className="flex items-center justify-between rounded-lg border border-green-800/50 bg-green-900/20 px-4 py-3">
          <p className="text-sm text-green-400">{successMsg}</p>
          <button type="button" onClick={() => setSuccessMsg('')} className="ml-4 text-sm text-green-500 transition-colors hover:text-green-300">
            Dismiss
          </button>
        </div>
      )}

      <div className="flex flex-col gap-4 rounded-lg border border-neutral-800 bg-neutral-900 p-5">
        <div className="flex flex-col gap-1.5">
          <label className="text-sm font-medium text-neutral-300">Local Directory</label>
          <div className="flex gap-2">
            <input
              type="text"
              value={localDir}
              placeholder="Select a local folder..."
              className="flex-1 rounded-md border border-neutral-700 bg-neutral-800 px-3 py-2 text-sm text-white placeholder-neutral-500 outline-none focus:border-blue-500 disabled:opacity-50"
              readOnly
              disabled={loading}
            />
            <button
              type="button"
              onClick={handleBrowseLocal}
              disabled={loading}
              className="rounded-md border border-neutral-700 bg-neutral-800 px-4 py-2 text-sm text-neutral-300 transition-colors hover:bg-neutral-700 disabled:cursor-not-allowed disabled:opacity-50"
            >
              Browse
            </button>
          </div>
        </div>

        <div className="flex flex-col gap-1.5">
          <label className="text-sm font-medium text-neutral-300">Google Drive Folder</label>
          <div className="flex gap-2">
            <input
              type="text"
              value={driveDirLabel}
              placeholder="Select a Drive folder..."
              className="flex-1 rounded-md border border-neutral-700 bg-neutral-800 px-3 py-2 text-sm text-white placeholder-neutral-500 outline-none focus:border-blue-500 disabled:opacity-50"
              readOnly
              disabled={loading}
            />
            <button
              type="button"
              onClick={() => setShowDrivePicker(true)}
              disabled={loading}
              className="rounded-md border border-neutral-700 bg-neutral-800 px-4 py-2 text-sm text-neutral-300 transition-colors hover:bg-neutral-700 disabled:cursor-not-allowed disabled:opacity-50"
            >
              Browse
            </button>
          </div>
        </div>

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
          selectedKeys={currentSelection}
          onSelectionChange={handleSelectionChange}
          onExport={() => setShowExportModal(true)}
        />
      )}

      <DriveFolderPicker
        open={showDrivePicker}
        onSelect={handleDriveSelect}
        onClose={() => setShowDrivePicker(false)}
      />

      <ExportModal
        open={showExportModal}
        onClose={() => setShowExportModal(false)}
        onExport={handleExport}
        selectedCount={currentSelection.size}
        activeTab={activeTab}
      />
    </div>
  );
}

export default ComparePage;
