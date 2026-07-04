import { useState, useMemo } from 'react';
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

type SortField = 'name' | 'extension' | 'size' | 'relativePath';
type SortDirection = 'asc' | 'desc';

interface SortState {
  field: SortField;
  direction: SortDirection;
}

const COLUMNS: { key: SortField; label: string; className: string }[] = [
  { key: 'name', label: 'Filename', className: 'px-4 py-3' },
  { key: 'extension', label: 'Extension', className: 'w-24 px-4 py-3' },
  { key: 'size', label: 'Size', className: 'w-28 px-4 py-3' },
  { key: 'relativePath', label: 'Relative Path', className: 'px-4 py-3' },
];

function sortItems(items: FileItem[], sort: SortState): FileItem[] {
  return [...items].sort((a, b) => {
    let cmp = 0;
    if (sort.field === 'size') {
      cmp = a.size - b.size;
    } else {
      cmp = a[sort.field].localeCompare(b[sort.field]);
    }
    return sort.direction === 'asc' ? cmp : -cmp;
  });
}

function SortIndicator({ field, sort }: { field: SortField; sort: SortState }): React.JSX.Element | null {
  if (sort.field !== field) {
    return <span className="ml-1 text-neutral-600">&#8597;</span>;
  }
  return <span className="ml-1">{sort.direction === 'asc' ? '▲' : '▼'}</span>;
}

function ResultsTable({ result, activeTab, onTabChange }: ResultsTableProps): React.JSX.Element {
  const [sort, setSort] = useState<SortState>({ field: 'name', direction: 'asc' });

  const items = result[activeTab];
  const sorted = useMemo(() => sortItems(items, sort), [items, sort]);

  function handleSort(field: SortField): void {
    setSort((prev) => ({
      field,
      direction: prev.field === field && prev.direction === 'asc' ? 'desc' : 'asc',
    }));
  }

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
              {COLUMNS.map((col) => (
                <th
                  key={col.key}
                  className={`${col.className} cursor-pointer select-none font-medium hover:text-neutral-200`}
                  onClick={() => handleSort(col.key)}
                >
                  {col.label}
                  <SortIndicator field={col.key} sort={sort} />
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {sorted.length === 0 ? (
              <tr>
                <td colSpan={5} className="px-4 py-8 text-center text-neutral-500">
                  No files found.
                </td>
              </tr>
            ) : (
              sorted.map((item: FileItem, index: number) => (
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
