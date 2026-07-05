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

type SourceType = 'local' | 'drive';

interface SourceState {
  type: SourceType;
  localDir: string;
  driveId: string;
  driveLabel: string;
}

const EMPTY_SELECTION: Record<TabKey, Set<string>> = {
  sourceAOnly: new Set(),
  sourceBOnly: new Set(),
  both: new Set(),
};

function sourceLabel(s: SourceState): string {
  if (s.type === 'local') return s.localDir || 'Local';
  return s.driveLabel || 'Drive';
}

function ComparePage(): React.JSX.Element {
  const [sourceA, setSourceA] = useState<SourceState>({ type: 'local', localDir: '', driveId: '', driveLabel: '' });
  const [sourceB, setSourceB] = useState<SourceState>({ type: 'drive', localDir: '', driveId: '', driveLabel: '' });
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<CompareResult | null>(null);
  const [activeTab, setActiveTab] = useState<TabKey>('sourceAOnly');
  const [error, setError] = useState('');
  const [successMsg, setSuccessMsg] = useState('');
  const [drivePicker, setDrivePicker] = useState<'a' | 'b' | null>(null);
  const [showExportModal, setShowExportModal] = useState(false);
  const [selections, setSelections] = useState<Record<TabKey, Set<string>>>({ ...EMPTY_SELECTION });

  const sourceAID = sourceA.type === 'local' ? sourceA.localDir : sourceA.driveId;
  const sourceBID = sourceB.type === 'local' ? sourceB.localDir : sourceB.driveId;
  const canCompare = sourceAID !== '' && sourceBID !== '' && !loading;
  const currentSelection = selections[activeTab];

  const tabLabels: Record<TabKey, string> = {
    sourceAOnly: `${sourceA.type === 'local' ? 'Local' : 'Drive'} 1st Only`,
    sourceBOnly: `${sourceB.type === 'local' ? 'Local' : 'Drive'} 2nd Only`,
    both: 'Both',
  };

  async function handleBrowseLocal(target: 'a' | 'b'): Promise<void> {
    try {
      setError('');
      const dir = await SelectLocalDirectory();
      if (dir) {
        if (target === 'a') {
          setSourceA((prev) => ({ ...prev, localDir: dir }));
        } else {
          setSourceB((prev) => ({ ...prev, localDir: dir }));
        }
      }
    } catch (err) {
      setError(`Failed to select directory: ${formatError(err)}`);
    }
  }

  function handleDriveSelect(folderId: string, folderName: string): void {
    if (drivePicker === 'a') {
      setSourceA((prev) => ({ ...prev, driveId: folderId, driveLabel: folderName }));
    } else {
      setSourceB((prev) => ({ ...prev, driveId: folderId, driveLabel: folderName }));
    }
    setDrivePicker(null);
  }

  async function handleCompare(): Promise<void> {
    setLoading(true);
    setResult(null);
    setError('');
    setSuccessMsg('');
    setSelections({ sourceAOnly: new Set(), sourceBOnly: new Set(), both: new Set() });

    try {
      const resp = await RunComparison(sourceA.type, sourceAID, sourceB.type, sourceBID);
      setResult({
        sourceAOnly: resp.sourceAOnly ?? [],
        sourceBOnly: resp.sourceBOnly ?? [],
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

  function getExportPathSource(): 'local' | 'drive' {
    if (activeTab === 'sourceAOnly') return sourceA.type;
    if (activeTab === 'sourceBOnly') return sourceB.type;
    return sourceA.type;
  }

  async function handleExport(pathSource: 'local' | 'drive', target: 'clipboard' | 'file', _contentType: 'items' | 'headers'): Promise<void> {
    if (!result) return;

    const items = result[activeTab];
    const selectedItems: FileItem[] = items.filter((item) => currentSelection.has(getItemKey(item)));

    if (selectedItems.length === 0) return;

    const localDir = sourceA.type === 'local' ? sourceA.localDir : sourceB.localDir;
    const driveLabel = sourceA.type === 'drive' ? sourceA.driveLabel : sourceB.driveLabel;
    const text = buildExportText(selectedItems, pathSource, localDir, driveLabel);

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
        <SourceRow
          label="First Source"
          source={sourceA}
          onChange={setSourceA}
          onBrowseLocal={() => handleBrowseLocal('a')}
          onBrowseDrive={() => setDrivePicker('a')}
          disabled={loading}
        />

        <SourceRow
          label="Second Source"
          source={sourceB}
          onChange={setSourceB}
          onBrowseLocal={() => handleBrowseLocal('b')}
          onBrowseDrive={() => setDrivePicker('b')}
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
          selectedKeys={currentSelection}
          onSelectionChange={handleSelectionChange}
          onExport={() => setShowExportModal(true)}
          tabLabels={tabLabels}
        />
      )}

      <DriveFolderPicker
        open={drivePicker !== null}
        onSelect={handleDriveSelect}
        onClose={() => setDrivePicker(null)}
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

function SourceRow({ label, source, onChange, onBrowseLocal, onBrowseDrive, disabled }: {
  label: string;
  source: SourceState;
  onChange: (s: SourceState) => void;
  onBrowseLocal: () => void;
  onBrowseDrive: () => void;
  disabled: boolean;
}): React.JSX.Element {
  const displayValue = source.type === 'local' ? source.localDir : source.driveLabel;
  const placeholder = source.type === 'local' ? 'Select a local folder...' : 'Select a Drive folder...';

  return (
    <div className="flex flex-col gap-1.5">
      <label className="text-sm font-medium text-neutral-300">{label}</label>
      <div className="flex gap-2">
        <select
          value={source.type}
          onChange={(e) => onChange({ ...source, type: e.target.value as SourceType })}
          disabled={disabled}
          className="w-28 rounded-md border border-neutral-700 bg-neutral-800 px-2 py-2 text-sm text-neutral-300 outline-none disabled:opacity-50"
        >
          <option value="local">Local</option>
          <option value="drive">Drive</option>
        </select>
        <input
          type="text"
          value={displayValue}
          placeholder={placeholder}
          className="flex-1 rounded-md border border-neutral-700 bg-neutral-800 px-3 py-2 text-sm text-white placeholder-neutral-500 outline-none focus:border-blue-500 disabled:opacity-50"
          readOnly
          disabled={disabled}
        />
        <button
          type="button"
          onClick={source.type === 'local' ? onBrowseLocal : onBrowseDrive}
          disabled={disabled}
          className="rounded-md border border-neutral-700 bg-neutral-800 px-4 py-2 text-sm text-neutral-300 transition-colors hover:bg-neutral-700 disabled:cursor-not-allowed disabled:opacity-50"
        >
          Browse
        </button>
      </div>
    </div>
  );
}

export default ComparePage;
