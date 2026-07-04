import type { FileItem, TabKey, CompareResult } from '../lib/types';
import { formatSize } from '../lib/types';

interface ResultsTableProps {
  result: CompareResult;
  activeTab: TabKey;
  onTabChange: (tab: TabKey) => void;
}

const TAB_LABELS: Record<TabKey, string> = {
  localOnly: 'Local Only',
  driveOnly: 'Drive Only',
  both: 'Both',
};

function ResultsTable({ result, activeTab, onTabChange }: ResultsTableProps): React.JSX.Element {
  const items = result[activeTab];

  return (
    <div className="flex flex-col gap-0 rounded-lg border border-neutral-800 bg-neutral-900">
      <div className="flex border-b border-neutral-800">
        {(Object.keys(TAB_LABELS) as TabKey[]).map((key) => (
          <button
            key={key}
            type="button"
            onClick={() => onTabChange(key)}
            className={`px-5 py-3 text-sm font-medium transition-colors ${
              activeTab === key
                ? 'border-b-2 border-blue-500 text-white'
                : 'text-neutral-400 hover:text-neutral-200'
            }`}
          >
            {TAB_LABELS[key]} ({result[key].length.toLocaleString()})
          </button>
        ))}
      </div>

      <div className="overflow-auto">
        <table className="w-full text-left text-sm">
          <thead>
            <tr className="border-b border-neutral-800 text-neutral-400">
              <th className="w-14 px-4 py-3 font-medium">No</th>
              <th className="px-4 py-3 font-medium">Filename</th>
              <th className="w-24 px-4 py-3 font-medium">Extension</th>
              <th className="w-28 px-4 py-3 font-medium">Size</th>
              <th className="px-4 py-3 font-medium">Relative Path</th>
            </tr>
          </thead>
          <tbody>
            {items.length === 0 ? (
              <tr>
                <td colSpan={5} className="px-4 py-8 text-center text-neutral-500">
                  No files found.
                </td>
              </tr>
            ) : (
              items.map((item: FileItem, index: number) => (
                <tr
                  key={`${item.relativePath}${item.name}${item.extension}`}
                  className="border-b border-neutral-800/50 hover:bg-neutral-800/30"
                >
                  <td className="px-4 py-2.5 text-neutral-500">{index + 1}</td>
                  <td className="px-4 py-2.5 text-white">{item.name}</td>
                  <td className="px-4 py-2.5 text-neutral-400">{item.extension}</td>
                  <td className="px-4 py-2.5 text-neutral-400">{formatSize(item.size)}</td>
                  <td className="px-4 py-2.5 text-neutral-500">{item.relativePath || '/'}</td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}

export default ResultsTable;
