import { useState, useMemo, useCallback, useRef, useEffect } from 'react';
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

const SIZE_RANGES = [
  { label: 'All', min: 0, max: Infinity },
  { label: '< 1 MB', min: 0, max: 1024 * 1024 },
  { label: '1-10 MB', min: 1024 * 1024, max: 10 * 1024 * 1024 },
  { label: '10-100 MB', min: 10 * 1024 * 1024, max: 100 * 1024 * 1024 },
  { label: '100 MB-1 GB', min: 100 * 1024 * 1024, max: 1024 * 1024 * 1024 },
  { label: '> 1 GB', min: 1024 * 1024 * 1024, max: Infinity },
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

function filterItems(
  items: FileItem[],
  search: string,
  extensions: Set<string>,
  sizeRangeIndex: number,
): FileItem[] {
  return items.filter((item) => {
    if (search && !item.name.toLowerCase().includes(search.toLowerCase())) {
      return false;
    }
    if (extensions.size > 0 && !extensions.has(item.extension)) {
      return false;
    }
    const range = SIZE_RANGES[sizeRangeIndex];
    if (item.size < range.min || item.size >= range.max) {
      return false;
    }
    return true;
  });
}

function SortIndicator({ field, sort }: { field: SortField; sort: SortState }): React.JSX.Element {
  if (sort.field !== field) {
    return <span className="ml-1 text-neutral-600">&#8597;</span>;
  }
  return <span className="ml-1">{sort.direction === 'asc' ? '▲' : '▼'}</span>;
}

const BATCH_SIZE = 50;

function ResultsTable({ result, activeTab, onTabChange }: ResultsTableProps): React.JSX.Element {
  const [sort, setSort] = useState<SortState>({ field: 'name', direction: 'asc' });
  const [search, setSearch] = useState('');
  const [selectedExtensions, setSelectedExtensions] = useState<Set<string>>(new Set());
  const [sizeRangeIndex, setSizeRangeIndex] = useState(0);
  const [visibleCount, setVisibleCount] = useState(BATCH_SIZE);
  const scrollRef = useRef<HTMLDivElement>(null);

  const items = result[activeTab];

  const availableExtensions = useMemo(() => {
    const exts = new Set<string>();
    for (const item of items) {
      if (item.extension) exts.add(item.extension);
    }
    return Array.from(exts).sort();
  }, [items]);

  const filtered = useMemo(
    () => filterItems(items, search, selectedExtensions, sizeRangeIndex),
    [items, search, selectedExtensions, sizeRangeIndex],
  );

  const sorted = useMemo(() => sortItems(filtered, sort), [filtered, sort]);

  const visible = useMemo(() => sorted.slice(0, visibleCount), [sorted, visibleCount]);
  const hasMore = visibleCount < sorted.length;

  useEffect(() => {
    setVisibleCount(BATCH_SIZE);
  }, [activeTab, search, selectedExtensions, sizeRangeIndex, sort]);

  const handleScroll = useCallback(() => {
    const el = scrollRef.current;
    if (!el || !hasMore) return;

    const threshold = 200;
    if (el.scrollHeight - el.scrollTop - el.clientHeight < threshold) {
      setVisibleCount((prev) => Math.min(prev + BATCH_SIZE, sorted.length));
    }
  }, [hasMore, sorted.length]);

  function handleSort(field: SortField): void {
    setSort((prev) => ({
      field,
      direction: prev.field === field && prev.direction === 'asc' ? 'desc' : 'asc',
    }));
  }

  function toggleExtension(ext: string): void {
    setSelectedExtensions((prev) => {
      const next = new Set(prev);
      if (next.has(ext)) {
        next.delete(ext);
      } else {
        next.add(ext);
      }
      return next;
    });
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

      <div className="flex flex-wrap items-center gap-3 border-b border-neutral-800 px-4 py-3">
        <input
          type="text"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search filename..."
          className="w-56 rounded-md border border-neutral-700 bg-neutral-800 px-3 py-1.5 text-sm text-white placeholder-neutral-500 outline-none focus:border-blue-500"
        />

        <div className="flex items-center gap-1.5">
          <span className="text-xs text-neutral-500">Extension:</span>
          <div className="flex flex-wrap gap-1">
            {availableExtensions.map((ext) => (
              <button
                key={ext}
                type="button"
                onClick={() => toggleExtension(ext)}
                className={`rounded px-2 py-0.5 text-xs transition-colors ${
                  selectedExtensions.has(ext)
                    ? 'bg-blue-600 text-white'
                    : 'bg-neutral-800 text-neutral-400 hover:bg-neutral-700'
                }`}
              >
                {ext}
              </button>
            ))}
            {selectedExtensions.size > 0 && (
              <button
                type="button"
                onClick={() => setSelectedExtensions(new Set())}
                className="rounded px-2 py-0.5 text-xs text-neutral-500 hover:text-neutral-300"
              >
                Clear
              </button>
            )}
          </div>
        </div>

        <div className="flex items-center gap-1.5">
          <span className="text-xs text-neutral-500">Size:</span>
          <select
            value={sizeRangeIndex}
            onChange={(e) => setSizeRangeIndex(Number(e.target.value))}
            className="rounded-md border border-neutral-700 bg-neutral-800 px-2 py-1 text-xs text-neutral-300 outline-none"
          >
            {SIZE_RANGES.map((range, i) => (
              <option key={range.label} value={i}>{range.label}</option>
            ))}
          </select>
        </div>

        <span className="ml-auto text-xs text-neutral-500">
          {sorted.length.toLocaleString()} of {items.length.toLocaleString()} files
        </span>
      </div>

      <div className="overflow-auto" ref={scrollRef} onScroll={handleScroll} style={{ maxHeight: '60vh' }}>
        <table className="w-full text-left text-sm">
          <thead>
            <tr className="border-b border-neutral-800 text-neutral-400">
              <th className="sticky top-0 z-10 w-14 bg-neutral-900 px-4 py-3 font-medium">No</th>
              {COLUMNS.map((col) => (
                <th
                  key={col.key}
                  className={`sticky top-0 z-10 bg-neutral-900 ${col.className} cursor-pointer select-none font-medium hover:text-neutral-200`}
                  onClick={() => handleSort(col.key)}
                >
                  {col.label}
                  <SortIndicator field={col.key} sort={sort} />
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {visible.length === 0 ? (
              <tr>
                <td colSpan={5} className="px-4 py-8 text-center text-neutral-500">
                  No files found.
                </td>
              </tr>
            ) : (
              visible.map((item: FileItem, index: number) => (
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
        {hasMore && (
          <div className="flex items-center justify-center gap-2 py-4">
            <div className="h-4 w-4 animate-spin rounded-full border-2 border-neutral-600 border-t-blue-500" />
            <span className="text-xs text-neutral-500">Loading more...</span>
          </div>
        )}
      </div>
    </div>
  );
}

export default ResultsTable;
