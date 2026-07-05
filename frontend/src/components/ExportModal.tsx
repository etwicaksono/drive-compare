import { useState } from 'react';

export type ExportContentType = 'items' | 'headers';

interface ExportModalProps {
  open: boolean;
  onClose: () => void;
  onExport: (pathSource: 'local' | 'drive', target: 'clipboard' | 'file', contentType: ExportContentType) => void;
  selectedCount: number;
  activeTab: string;
  showContentOption?: boolean;
  contentLabels?: { headers: string; items: string };
}

function ExportModal({ open, onClose, onExport, selectedCount, activeTab, showContentOption, contentLabels }: ExportModalProps): React.JSX.Element | null {
  const defaultPathSource = activeTab === 'sourceBOnly' || activeTab === 'driveOnly' ? 'drive' : 'local';
  const [pathSource, setPathSource] = useState<'local' | 'drive'>(defaultPathSource);
  const [target, setTarget] = useState<'clipboard' | 'file'>('clipboard');
  const [contentType, setContentType] = useState<ExportContentType>('items');

  if (!open) return null;

  const showPathSource = activeTab === 'both';
  const effectivePathSource = showPathSource ? pathSource : defaultPathSource;
  const headersLabel = contentLabels?.headers ?? 'Group names (headers)';
  const itemsLabel = contentLabels?.items ?? 'All files (items)';

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60">
      <div className="w-[420px] rounded-xl border border-neutral-700 bg-neutral-900 shadow-2xl">
        <div className="flex items-center justify-between border-b border-neutral-800 px-5 py-3">
          <h3 className="text-sm font-semibold text-white">
            Export {selectedCount} selected {showContentOption ? 'group' : 'file'}{selectedCount !== 1 ? 's' : ''}
          </h3>
          <button type="button" onClick={onClose} className="text-neutral-500 transition-colors hover:text-white">
            &times;
          </button>
        </div>

        <div className="flex flex-col gap-5 px-5 py-4">
          {showContentOption && (
            <div className="flex flex-col gap-2">
              <span className="text-xs font-medium text-neutral-400 uppercase tracking-wide">Export content</span>
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="radio"
                  name="contentType"
                  checked={contentType === 'headers'}
                  onChange={() => setContentType('headers')}
                  className="accent-blue-500"
                />
                <span className="text-sm text-neutral-300">{headersLabel}</span>
              </label>
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="radio"
                  name="contentType"
                  checked={contentType === 'items'}
                  onChange={() => setContentType('items')}
                  className="accent-blue-500"
                />
                <span className="text-sm text-neutral-300">{itemsLabel}</span>
              </label>
            </div>
          )}

          {showPathSource && (
            <div className="flex flex-col gap-2">
              <span className="text-xs font-medium text-neutral-400 uppercase tracking-wide">Path source</span>
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="radio"
                  name="pathSource"
                  checked={pathSource === 'local'}
                  onChange={() => setPathSource('local')}
                  className="accent-blue-500"
                />
                <span className="text-sm text-neutral-300">Local Storage</span>
              </label>
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="radio"
                  name="pathSource"
                  checked={pathSource === 'drive'}
                  onChange={() => setPathSource('drive')}
                  className="accent-blue-500"
                />
                <span className="text-sm text-neutral-300">Google Drive</span>
              </label>
            </div>
          )}

          <div className="flex flex-col gap-2">
            <span className="text-xs font-medium text-neutral-400 uppercase tracking-wide">Export to</span>
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="radio"
                name="target"
                checked={target === 'clipboard'}
                onChange={() => setTarget('clipboard')}
                className="accent-blue-500"
              />
              <span className="text-sm text-neutral-300">Clipboard</span>
            </label>
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="radio"
                name="target"
                checked={target === 'file'}
                onChange={() => setTarget('file')}
                className="accent-blue-500"
              />
              <span className="text-sm text-neutral-300">Save as .txt file</span>
            </label>
          </div>
        </div>

        <div className="flex justify-end gap-2 border-t border-neutral-800 px-5 py-3">
          <button
            type="button"
            onClick={onClose}
            className="rounded-md border border-neutral-700 px-4 py-1.5 text-sm text-neutral-300 transition-colors hover:bg-neutral-800"
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={() => onExport(effectivePathSource, target, showContentOption ? contentType : 'items')}
            className="rounded-md bg-blue-600 px-4 py-1.5 text-sm font-medium text-white transition-colors hover:bg-blue-500"
          >
            Export
          </button>
        </div>
      </div>
    </div>
  );
}

export default ExportModal;
