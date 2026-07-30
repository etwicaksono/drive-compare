import { useState } from 'react';
import { SelectLocalDirectory, RunResizeAudit, ExportToClipboard, ExportToFile } from '../../wailsjs/go/main/App';
import ResultsTable from '../components/ResultsTable';
import ErrorBanner from '../components/ErrorBanner';
import DriveFolderPicker from '../components/DriveFolderPicker';
import ExportModal from '../components/ExportModal';
import type { ResizeAuditResult, ResizeAuditTabKey, FileItem, DriveScope } from '../lib/types';
import { getItemKey, buildExportText, DRIVE_SCOPE_LABELS } from '../lib/types';

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
  scope: DriveScope;
}

const EMPTY_SELECTION: Record<ResizeAuditTabKey, Set<string>> = {
  pending: new Set(),
  resized: new Set(),
  extraRes: new Set(),
};

function ResizeAuditPage(): React.JSX.Element {
  const [sourceDir, setSourceDir] = useState<SourceState>({ type: 'local', localDir: '', driveId: '', driveLabel: '', scope: 'active' });
  const [resizedDir, setResizedDir] = useState<SourceState>({ type: 'local', localDir: '', driveId: '', driveLabel: '', scope: 'active' });
  const [suffix, setSuffix] = useState('_resized');
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<ResizeAuditResult | null>(null);
  const [activeTab, setActiveTab] = useState<ResizeAuditTabKey>('pending');
  const [error, setError] = useState('');
  const [successMsg, setSuccessMsg] = useState('');
  const [drivePicker, setDrivePicker] = useState<'source' | 'resized' | null>(null);
  const [showExportModal, setShowExportModal] = useState(false);
  const [selections, setSelections] = useState<Record<ResizeAuditTabKey, Set<string>>>({ ...EMPTY_SELECTION });

  const sourceID = sourceDir.type === 'local' ? sourceDir.localDir : sourceDir.driveId;
  const resizedID = resizedDir.type === 'local' ? resizedDir.localDir : resizedDir.driveId;
  const canAudit = sourceID !== '' && resizedID !== '' && suffix !== '' && !loading;
  const currentSelection = selections[activeTab];

  const tabLabels: Record<ResizeAuditTabKey, string> = {
    pending: 'Pending',
    resized: 'Resized',
    extraRes: 'Extra Resized',
  };

  const resultForTable = result ? {
    sourceAOnly: result.pending,
    sourceBOnly: result.extraRes,
    both: result.resized,
  } : null;

  const hasDeletedItems = resultForTable
    ? [...resultForTable.sourceAOnly, ...resultForTable.sourceBOnly, ...resultForTable.both].some((item) => item.deleted)
    : false;

  const tabKeyMap: Record<ResizeAuditTabKey, 'sourceAOnly' | 'sourceBOnly' | 'both'> = {
    pending: 'sourceAOnly',
    resized: 'both',
    extraRes: 'sourceBOnly',
  };

  const reverseTabKeyMap: Record<string, ResizeAuditTabKey> = {
    sourceAOnly: 'pending',
    both: 'resized',
    sourceBOnly: 'extraRes',
  };

  const tableTabLabels: Record<'sourceAOnly' | 'sourceBOnly' | 'both', string> = {
    sourceAOnly: tabLabels.pending,
    sourceBOnly: tabLabels.extraRes,
    both: tabLabels.resized,
  };

  const tableTabTooltips: Record<'sourceAOnly' | 'sourceBOnly' | 'both', string> = {
    sourceAOnly: 'Source files that have not been resized yet',
    both: 'Source files that already have a matching resized file',
    sourceBOnly: 'Resized files without a matching source file',
  };

  async function handleBrowseLocal(target: 'source' | 'resized'): Promise<void> {
    try {
      setError('');
      const dir = await SelectLocalDirectory();
      if (dir) {
        if (target === 'source') {
          setSourceDir((prev) => ({ ...prev, localDir: dir }));
        } else {
          setResizedDir((prev) => ({ ...prev, localDir: dir }));
        }
      }
    } catch (err) {
      setError(`Failed to select directory: ${formatError(err)}`);
    }
  }

  function handleDriveSelect(folderId: string, folderName: string): void {
    if (drivePicker === 'source') {
      setSourceDir((prev) => ({ ...prev, driveId: folderId, driveLabel: folderName }));
    } else {
      setResizedDir((prev) => ({ ...prev, driveId: folderId, driveLabel: folderName }));
    }
    setDrivePicker(null);
  }

  async function handleAudit(): Promise<void> {
    setLoading(true);
    setResult(null);
    setError('');
    setSuccessMsg('');
    setSelections({ pending: new Set(), resized: new Set(), extraRes: new Set() });

    try {
      const resp = await RunResizeAudit(sourceDir.type, sourceID, sourceDir.scope, resizedDir.type, resizedID, resizedDir.scope, suffix);
      setResult({
        pending: resp.pending ?? [],
        resized: resp.resized ?? [],
        extraRes: resp.extraRes ?? [],
      });
    } catch (err) {
      setError(`Audit failed: ${formatError(err)}`);
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

    const localDir = sourceDir.type === 'local' ? sourceDir.localDir : resizedDir.localDir;
    const driveLabel = sourceDir.type === 'drive' ? sourceDir.driveLabel : resizedDir.driveLabel;
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
      <h1 className="text-xl font-semibold">Resize Audit</h1>

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
          label="Source (Original Files)"
          source={sourceDir}
          onChange={setSourceDir}
          onBrowseLocal={() => handleBrowseLocal('source')}
          onBrowseDrive={() => setDrivePicker('source')}
          disabled={loading}
        />

        <SourceRow
          label="Resized (Output Files)"
          source={resizedDir}
          onChange={setResizedDir}
          onBrowseLocal={() => handleBrowseLocal('resized')}
          onBrowseDrive={() => setDrivePicker('resized')}
          disabled={loading}
        />

        <div className="flex flex-col gap-1.5">
          <label className="text-sm font-medium text-neutral-300">Suffix</label>
          <input
            type="text"
            value={suffix}
            onChange={(e) => setSuffix(e.target.value)}
            placeholder="_resized"
            className="w-56 rounded-md border border-neutral-700 bg-neutral-800 px-3 py-2 text-sm text-white placeholder-neutral-500 outline-none focus:border-blue-500 disabled:opacity-50"
            disabled={loading}
          />
          <span className="text-xs text-neutral-500">
            e.g. video.mp4 → video{suffix}.mp4
          </span>
        </div>

        <button
          type="button"
          onClick={handleAudit}
          disabled={!canAudit}
          className="mt-1 w-full rounded-md bg-blue-600 px-4 py-2.5 text-sm font-medium text-white transition-colors hover:bg-blue-500 disabled:cursor-not-allowed disabled:opacity-40"
        >
          {loading ? 'Scanning...' : 'Run Audit'}
        </button>
      </div>

      {loading && (
        <div className="flex items-center justify-center gap-3 rounded-lg border border-neutral-800 bg-neutral-900 p-8">
          <div className="h-5 w-5 animate-spin rounded-full border-2 border-neutral-600 border-t-blue-500" />
          <span className="text-sm text-neutral-400">Scanning and auditing files...</span>
        </div>
      )}

      {resultForTable && !loading && (
        <ResultsTable
          result={resultForTable}
          activeTab={tabKeyMap[activeTab]}
          onTabChange={(key) => setActiveTab(reverseTabKeyMap[key])}
          selectedKeys={currentSelection}
          onSelectionChange={handleSelectionChange}
          onExport={() => setShowExportModal(true)}
          tabLabels={tableTabLabels}
          tabTooltips={tableTabTooltips}
          showStatusColumn={hasDeletedItems}
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
          onChange={(e) => {
            const type = e.target.value as SourceType;
            onChange({ ...source, type, scope: type === 'local' ? 'active' : source.scope });
          }}
          disabled={disabled}
          className="w-28 rounded-md border border-neutral-700 bg-neutral-800 px-2 py-2 text-sm text-neutral-300 outline-none disabled:opacity-50"
        >
          <option value="local">Local</option>
          <option value="drive">Drive</option>
        </select>
        {source.type === 'drive' && (
          <select
            value={source.scope}
            onChange={(e) => onChange({ ...source, scope: e.target.value as DriveScope })}
            disabled={disabled}
            title="Which Drive files to include in the scan"
            className="w-40 rounded-md border border-neutral-700 bg-neutral-800 px-2 py-2 text-sm text-neutral-300 outline-none disabled:opacity-50"
          >
            {(Object.keys(DRIVE_SCOPE_LABELS) as DriveScope[]).map((scope) => (
              <option key={scope} value={scope}>
                {DRIVE_SCOPE_LABELS[scope]}
              </option>
            ))}
          </select>
        )}
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

export default ResizeAuditPage;
