import { useState, useEffect } from 'react';
import { ListDriveFolders } from '../../wailsjs/go/main/App';
import type { main } from '../../wailsjs/go/models';

interface DriveFolderPickerProps {
  open: boolean;
  onSelect: (folderId: string, folderName: string) => void;
  onClose: () => void;
}

interface BreadcrumbItem {
  id: string;
  name: string;
}

function DriveFolderPicker({ open, onSelect, onClose }: DriveFolderPickerProps): React.JSX.Element | null {
  const [folders, setFolders] = useState<main.DriveFolder[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [breadcrumbs, setBreadcrumbs] = useState<BreadcrumbItem[]>([
    { id: 'root', name: 'My Drive' },
  ]);

  const currentFolder = breadcrumbs[breadcrumbs.length - 1];

  useEffect(() => {
    if (!open) return;
    loadFolders(currentFolder.id);
  }, [open, currentFolder.id]);

  async function loadFolders(parentID: string): Promise<void> {
    setLoading(true);
    setError('');
    try {
      const result = await ListDriveFolders(parentID);
      setFolders(result ?? []);
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
      setFolders([]);
    } finally {
      setLoading(false);
    }
  }

  function handleFolderClick(folder: main.DriveFolder): void {
    setBreadcrumbs((prev) => [...prev, { id: folder.id, name: folder.name }]);
  }

  function handleBreadcrumbClick(index: number): void {
    setBreadcrumbs((prev) => prev.slice(0, index + 1));
  }

  function handleSelectCurrent(): void {
    onSelect(currentFolder.id, breadcrumbs.map((b) => b.name).join(' / '));
  }

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60">
      <div className="flex w-[520px] max-h-[480px] flex-col rounded-xl border border-neutral-700 bg-neutral-900 shadow-2xl">
        <div className="flex items-center justify-between border-b border-neutral-800 px-5 py-3">
          <h3 className="text-sm font-semibold text-white">Select Google Drive Folder</h3>
          <button
            type="button"
            onClick={onClose}
            className="text-neutral-500 transition-colors hover:text-white"
          >
            &times;
          </button>
        </div>

        <div className="flex flex-wrap items-center gap-1 border-b border-neutral-800 px-5 py-2">
          {breadcrumbs.map((crumb, i) => (
            <span key={crumb.id} className="flex items-center gap-1">
              {i > 0 && <span className="text-xs text-neutral-600">/</span>}
              <button
                type="button"
                onClick={() => handleBreadcrumbClick(i)}
                className={`text-xs transition-colors ${
                  i === breadcrumbs.length - 1
                    ? 'font-medium text-white'
                    : 'text-neutral-400 hover:text-blue-400'
                }`}
              >
                {crumb.name}
              </button>
            </span>
          ))}
        </div>

        <div className="flex-1 overflow-y-auto px-2 py-2" style={{ minHeight: '200px', maxHeight: '320px' }}>
          {loading && (
            <div className="flex items-center justify-center py-12">
              <div className="h-5 w-5 animate-spin rounded-full border-2 border-neutral-600 border-t-blue-500" />
              <span className="ml-3 text-sm text-neutral-400">Loading folders...</span>
            </div>
          )}

          {error && (
            <div className="rounded-md bg-red-900/20 px-4 py-3">
              <p className="text-sm text-red-400">{error}</p>
            </div>
          )}

          {!loading && !error && folders.length === 0 && (
            <p className="py-8 text-center text-sm text-neutral-500">No subfolders in this folder.</p>
          )}

          {!loading && !error && folders.map((folder) => (
            <button
              key={folder.id}
              type="button"
              onClick={() => handleFolderClick(folder)}
              className="flex w-full items-center gap-3 rounded-md px-3 py-2 text-left transition-colors hover:bg-neutral-800"
            >
              <span className="text-lg">&#128193;</span>
              <span className="text-sm text-neutral-200">{folder.name}</span>
            </button>
          ))}
        </div>

        <div className="flex items-center justify-between border-t border-neutral-800 px-5 py-3">
          <span className="text-xs text-neutral-500 truncate max-w-[280px]">
            {breadcrumbs.map((b) => b.name).join(' / ')}
          </span>
          <div className="flex gap-2">
            <button
              type="button"
              onClick={onClose}
              className="rounded-md border border-neutral-700 px-4 py-1.5 text-sm text-neutral-300 transition-colors hover:bg-neutral-800"
            >
              Cancel
            </button>
            <button
              type="button"
              onClick={handleSelectCurrent}
              className="rounded-md bg-blue-600 px-4 py-1.5 text-sm font-medium text-white transition-colors hover:bg-blue-500"
            >
              Select This Folder
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

export default DriveFolderPicker;
