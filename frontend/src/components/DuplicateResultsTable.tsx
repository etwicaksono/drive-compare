import { useState, useMemo, useRef } from 'react';
import type { DuplicateGroup, FileItem } from '../lib/types';
import { formatSize } from '../lib/types';

interface DuplicateResultsTableProps {
  groups: DuplicateGroup[];
  totalGroups: number;
  totalFiles: number;
  selectedGroupKeys: Set<string>;
  onSelectionChange: (keys: Set<string>) => void;
  onExport: () => void;
}

type SortField = 'name' | 'extension' | 'size' | 'relativePath';
type SortDirection = 'asc' | 'desc';

interface SortState {
  field: SortField;
  direction: SortDirection;
}

const SIZE_RANGES = [
  { label: 'All', min: 0, max: Infinity },
  { label: '< 1 MB', min: 0, max: 1024 * 1024 },
  { label: '1-10 MB', min: 1024 * 1024, max: 10 * 1024 * 1024 },
  { label: '10-100 MB', min: 10 * 1024 * 1024, max: 100 * 1024 * 1024 },
  { label: '100 MB-1 GB', min: 100 * 1024 * 1024, max: 1024 * 1024 * 1024 },
  { label: '> 1 GB', min: 1024 * 1024 * 1024, max: Infinity },
];

function sortFiles(files: FileItem[], sort: SortState): FileItem[] {
  return [...files].sort((a, b) => {
    let cmp = 0;
    if (sort.field === 'size') {
      cmp = a.size - b.size;
    } else {
      cmp = a[sort.field].localeCompare(b[sort.field]);
    }
    return sort.direction === 'asc' ? cmp : -cmp;
  });
}

function SortIndicator({ field, sort }: { field: SortField; sort: SortState }): React.JSX.Element {
  if (sort.field !== field) {
    return <span className="ml-1 text-neutral-600">&#8597;</span>;
  }
  return <span className="ml-1">{sort.direction === 'asc' ? '▲' : '▼'}</span>;
}

const COLUMNS: { key: SortField; label: string; className: string }[] = [
  { key: 'name', label: 'Filename', className: 'px-4 py-2' },
  { key: 'extension', label: 'Extension', className: 'w-24 px-4 py-2' },
  { key: 'size', label: 'Size', className: 'w-28 px-4 py-2' },
  { key: 'relativePath', label: 'Relative Path', className: 'px-4 py-2' },
];

export function makeGroupKey(g: DuplicateGroup): string {
  return `${g.baseName}${g.extension}|${g.relativePath}`;
}

function DuplicateResultsTable({ groups, totalGroups, totalFiles, selectedGroupKeys, onSelectionChange, onExport }: DuplicateResultsTableProps): React.JSX.Element {
  const [collapsed, setCollapsed] = useState<Set<string>>(new Set());
  const [sort, setSort] = useState<SortState>({ field: 'name', direction: 'asc' });
  const [search, setSearch] = useState('');
  const [selectedExtensions, setSelectedExtensions] = useState<Set<string>>(new Set());
  const [sizeRangeIndex, setSizeRangeIndex] = useState(0);
  const lastClickedGroupIndex = useRef<number>(-1);

  const filteredGroups = useMemo(() => {
    return groups
      .map((group) => {
        const filtered = group.files.filter((f) => {
          if (search && !f.name.toLowerCase().includes(search.toLowerCase())) return false;
          if (selectedExtensions.size > 0 && !selectedExtensions.has(f.extension)) return false;
          const range = SIZE_RANGES[sizeRangeIndex];
          if (f.size < range.min || f.size >= range.max) return false;
          return true;
        });
        return { ...group, files: filtered };
      })
      .filter((g) => g.files.length > 0);
  }, [groups, search, selectedExtensions, sizeRangeIndex]);

  const sortedGroups = useMemo(() => {
    return filteredGroups.map((g) => ({ ...g, files: sortFiles(g.files, sort) }));
  }, [filteredGroups, sort]);

  const allGroupKeys = useMemo(() => {
    return new Set(filteredGroups.map(makeGroupKey));
  }, [filteredGroups]);

  const selectedInFilter = useMemo(() => {
    let count = 0;
    for (const key of selectedGroupKeys) {
      if (allGroupKeys.has(key)) count++;
    }
    return count;
  }, [selectedGroupKeys, allGroupKeys]);

  const selectedFileCount = useMemo(() => {
    let count = 0;
    for (const g of filteredGroups) {
      if (selectedGroupKeys.has(makeGroupKey(g))) {
        count += g.files.length;
      }
    }
    return count;
  }, [filteredGroups, selectedGroupKeys]);

  const allSelected = allGroupKeys.size > 0 && selectedInFilter === allGroupKeys.size;
  const someSelected = selectedInFilter > 0 && !allSelected;

  const availableExtensions = useMemo(() => {
    const exts = new Set<string>();
    for (const g of groups) {
      for (const f of g.files) {
        if (f.extension) exts.add(f.extension);
      }
    }
    return Array.from(exts).sort();
  }, [groups]);

  const filteredFileCount = useMemo(() => {
    return filteredGroups.reduce((sum, g) => sum + g.files.length, 0);
  }, [filteredGroups]);

  function toggleCollapse(key: string): void {
    setCollapsed((prev) => {
      const next = new Set(prev);
      if (next.has(key)) { next.delete(key); } else { next.add(key); }
      return next;
    });
  }

  function handleSort(field: SortField): void {
    setSort((prev) => ({
      field,
      direction: prev.field === field && prev.direction === 'asc' ? 'desc' : 'asc',
    }));
  }

  function toggleExtension(ext: string): void {
    setSelectedExtensions((prev) => {
      const next = new Set(prev);
      if (next.has(ext)) { next.delete(ext); } else { next.add(ext); }
      return next;
    });
  }

  function handleGroupClick(key: string, index: number, e: React.MouseEvent): void {
    e.stopPropagation();
    if (e.shiftKey) e.preventDefault();
    if (e.shiftKey && lastClickedGroupIndex.current >= 0) {
      const start = Math.min(lastClickedGroupIndex.current, index);
      const end = Math.max(lastClickedGroupIndex.current, index);
      const next = new Set(selectedGroupKeys);
      for (let i = start; i <= end; i++) {
        next.add(makeGroupKey(sortedGroups[i]));
      }
      onSelectionChange(next);
    } else if (e.ctrlKey || e.metaKey) {
      const next = new Set(selectedGroupKeys);
      if (next.has(key)) { next.delete(key); } else { next.add(key); }
      onSelectionChange(next);
    } else {
      if (selectedGroupKeys.has(key) && selectedGroupKeys.size === 1) {
        onSelectionChange(new Set());
      } else {
        onSelectionChange(new Set([key]));
      }
    }
    lastClickedGroupIndex.current = index;
  }

  function toggleSelectAll(): void {
    if (allSelected) {
      onSelectionChange(new Set());
    } else {
      onSelectionChange(new Set(allGroupKeys));
    }
  }

  if (groups.length === 0) {
    return (
      <div className="rounded-lg border border-neutral-800 bg-neutral-900 px-4 py-8 text-center text-sm text-neutral-500">
        No duplicates found with the selected patterns.
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-0 rounded-lg border border-neutral-800 bg-neutral-900">
      <div className="flex flex-wrap items-center gap-3 border-b border-neutral-800 px-4 py-3">
        <span className="text-sm text-neutral-300">
          Found {totalGroups} duplicate group{totalGroups !== 1 ? 's' : ''} ({totalFiles} files)
        </span>

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
                  selectedExtensions.has(ext) ? 'bg-blue-600 text-white' : 'bg-neutral-800 text-neutral-400 hover:bg-neutral-700'
                }`}
              >
                {ext}
              </button>
            ))}
            {selectedExtensions.size > 0 && (
              <button type="button" onClick={() => setSelectedExtensions(new Set())} className="rounded px-2 py-0.5 text-xs text-neutral-500 hover:text-neutral-300">
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

        <div className="ml-auto flex items-center gap-2">
          <button
            type="button"
            onClick={() => setCollapsed(new Set())}
            className="rounded px-2 py-0.5 text-xs text-neutral-400 hover:text-neutral-200 hover:bg-neutral-800 transition-colors"
          >
            Expand All
          </button>
          <button
            type="button"
            onClick={() => setCollapsed(new Set(filteredGroups.map(makeGroupKey)))}
            className="rounded px-2 py-0.5 text-xs text-neutral-400 hover:text-neutral-200 hover:bg-neutral-800 transition-colors"
          >
            Collapse All
          </button>
          <span className="text-xs text-neutral-500">
            Showing {filteredGroups.length} of {totalGroups} groups ({filteredFileCount} files)
          </span>
          {selectedInFilter > 0 && (
            <>
              <span className="text-xs text-blue-400">
                {selectedInFilter} group{selectedInFilter !== 1 ? 's' : ''} ({selectedFileCount} files)
              </span>
              <button
                type="button"
                onClick={onExport}
                className="rounded-md bg-blue-600 px-3 py-1 text-xs font-medium text-white transition-colors hover:bg-blue-500"
              >
                Export Selected
              </button>
            </>
          )}
        </div>
      </div>

      <div className="overflow-auto" style={{ maxHeight: '60vh' }}>
        <table className="w-full text-left text-sm">
          <thead>
            <tr className="border-b border-neutral-800 text-neutral-400">
              <th
                className="sticky top-0 z-10 w-10 bg-neutral-900 px-3 py-2 font-medium cursor-pointer"
                onClick={toggleSelectAll}
              >
                <input
                  type="checkbox"
                  checked={allSelected}
                  readOnly
                  className="accent-blue-500 pointer-events-none"
                  ref={(el) => { if (el) el.indeterminate = someSelected; }}
                />
              </th>
              <th className="sticky top-0 z-10 w-14 bg-neutral-900 px-4 py-2 font-medium">No</th>
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
            {sortedGroups.map((group, index) => {
              const key = makeGroupKey(group);
              const isCollapsed = collapsed.has(key);
              const isGroupSelected = selectedGroupKeys.has(key);

              return (
                <GroupRows
                  key={key}
                  group={group}
                  files={group.files}
                  isCollapsed={isCollapsed}
                  isSelected={isGroupSelected}
                  onToggleCollapse={() => toggleCollapse(key)}
                  onToggleSelect={(e) => handleGroupClick(key, index, e)}
                />
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function GroupRows({ group, files, isCollapsed, isSelected, onToggleCollapse, onToggleSelect }: {
  group: DuplicateGroup;
  files: FileItem[];
  isCollapsed: boolean;
  isSelected: boolean;
  onToggleCollapse: () => void;
  onToggleSelect: (e: React.MouseEvent) => void;
}): React.JSX.Element {
  return (
    <>
      <tr
        className={`border-b border-neutral-700 bg-neutral-800/50 cursor-pointer hover:bg-neutral-800 ${isSelected ? 'bg-blue-900/20' : ''}`}
        onClick={onToggleCollapse}
      >
        <td className="w-10 px-3 py-2.5" onClick={onToggleSelect}>
          <input
            type="checkbox"
            checked={isSelected}
            readOnly
            className="accent-blue-500 pointer-events-none"
          />
        </td>
        <td colSpan={5} className="px-4 py-2.5">
          <div className="flex items-center gap-2">
            <span className="text-neutral-400">{isCollapsed ? '▶' : '▼'}</span>
            <span className="font-medium text-white">
              {group.baseName}{group.extension}
            </span>
            <span className="text-xs text-neutral-500">
              ({files.length} file{files.length !== 1 ? 's' : ''})
            </span>
            {group.relativePath && (
              <span className="text-xs text-neutral-600 ml-2">{group.relativePath}</span>
            )}
          </div>
        </td>
      </tr>
      {!isCollapsed && files.map((file, i) => (
        <tr
          key={`${file.relativePath}${file.name}${file.extension}`}
          className="border-b border-neutral-800/30 hover:bg-neutral-800/20"
        >
          <td className="w-10 px-3 py-2" />
          <td className="px-4 py-2 text-neutral-500 pl-8">{i + 1}</td>
          <td className="px-4 py-2 text-white">{file.name}</td>
          <td className="px-4 py-2 text-neutral-400">{file.extension}</td>
          <td className="px-4 py-2 text-neutral-400">{formatSize(file.size)}</td>
          <td className="px-4 py-2 text-neutral-500">{file.relativePath || '/'}</td>
        </tr>
      ))}
    </>
  );
}

export default DuplicateResultsTable;
