import { useState } from 'react';
import { SelectLocalDirectory } from '../../wailsjs/go/main/App';

interface FolderPickerProps {
  label: string;
  value: string;
  onSelect: (path: string) => void;
  onBrowse: () => void;
  placeholder: string;
}

function FolderPicker({ label, value, onSelect, onBrowse, placeholder }: FolderPickerProps): React.JSX.Element {
  return (
    <div className="flex flex-col gap-1.5">
      <label className="text-sm font-medium text-neutral-300">{label}</label>
      <div className="flex gap-2">
        <input
          type="text"
          value={value}
          onChange={(e) => onSelect(e.target.value)}
          placeholder={placeholder}
          className="flex-1 rounded-md border border-neutral-700 bg-neutral-800 px-3 py-2 text-sm text-white placeholder-neutral-500 outline-none focus:border-blue-500"
          readOnly
        />
        <button
          type="button"
          onClick={onBrowse}
          className="rounded-md border border-neutral-700 bg-neutral-800 px-4 py-2 text-sm text-neutral-300 transition-colors hover:bg-neutral-700"
        >
          Browse
        </button>
      </div>
    </div>
  );
}

function ComparePage(): React.JSX.Element {
  const [localDir, setLocalDir] = useState('');
  const [driveDir, setDriveDir] = useState('');

  const canCompare = localDir !== '' && driveDir !== '';

  async function handleBrowseLocal(): Promise<void> {
    try {
      const dir = await SelectLocalDirectory();
      if (dir) {
        setLocalDir(dir);
      }
    } catch (err) {
      console.error('Failed to select directory:', err);
    }
  }

  function handleBrowseDrive(): void {
    // TODO: Implement Google Drive folder picker in future task
  }

  function handleCompare(): void {
    // TODO: Implement compare in future task
  }

  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-xl font-semibold">Compare</h1>

      <div className="flex flex-col gap-4 rounded-lg border border-neutral-800 bg-neutral-900 p-5">
        <FolderPicker
          label="Local Directory"
          value={localDir}
          onSelect={setLocalDir}
          onBrowse={handleBrowseLocal}
          placeholder="Select a local folder..."
        />

        <FolderPicker
          label="Google Drive Directory"
          value={driveDir}
          onSelect={setDriveDir}
          onBrowse={handleBrowseDrive}
          placeholder="Select a Drive folder..."
        />

        <button
          type="button"
          onClick={handleCompare}
          disabled={!canCompare}
          className="mt-1 w-full rounded-md bg-blue-600 px-4 py-2.5 text-sm font-medium text-white transition-colors hover:bg-blue-500 disabled:cursor-not-allowed disabled:opacity-40"
        >
          Compare
        </button>
      </div>
    </div>
  );
}

export default ComparePage;
