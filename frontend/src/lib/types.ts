export interface FileItem {
  relativePath: string;
  name: string;
  extension: string;
  size: number;
  deleted: boolean;
}

export type DriveScope = 'active' | 'trashed' | 'all';

export const DRIVE_SCOPE_LABELS: Record<DriveScope, string> = {
  active: 'Active files',
  trashed: 'Deleted (Trash)',
  all: 'Active + Deleted',
};

export interface CompareResult {
  sourceAOnly: FileItem[];
  sourceBOnly: FileItem[];
  both: FileItem[];
}

export type TabKey = 'sourceAOnly' | 'sourceBOnly' | 'both';

export function formatSize(bytes: number): string {
  if (bytes === 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const k = 1024;
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  const value = bytes / Math.pow(k, i);
  return `${value.toFixed(i === 0 ? 0 : 1)} ${units[i]}`;
}

// The deleted flag is part of the key so an active file and its trashed
// namesake stay independently selectable when scope is "all".
export function getItemKey(item: FileItem): string {
  return `${item.relativePath}${item.name}${item.extension}${item.deleted ? '#deleted' : ''}`;
}

export function buildLocalPath(localDir: string, item: FileItem): string {
  const rel = item.relativePath.replace(/\//g, '\\');
  return `${localDir}\\${rel}${item.name}${item.extension}`;
}

export function buildDrivePath(driveLabel: string, item: FileItem): string {
  return `${driveLabel}/${item.relativePath}${item.name}${item.extension}`;
}

export function buildExportText(
  items: FileItem[],
  pathSource: 'local' | 'drive',
  localDir: string,
  driveLabel: string,
): string {
  const builder = pathSource === 'local' ? buildLocalPath : buildDrivePath;
  const base = pathSource === 'local' ? localDir : driveLabel;
  return items.map((item) => `${builder(base, item)}${item.deleted ? ' [deleted]' : ''}`).join('\n');
}

export interface DuplicatePatternInfo {
  id: string;
  label: string;
  example: string;
}

export interface DuplicateGroup {
  baseName: string;
  extension: string;
  relativePath: string;
  files: FileItem[];
}

export interface DuplicateScanResult {
  groups: DuplicateGroup[];
  totalGroups: number;
  totalFiles: number;
}

export type ResizeAuditTabKey = 'pending' | 'resized' | 'extraRes';

export interface ResizeAuditResult {
  pending: FileItem[];
  resized: FileItem[];
  extraRes: FileItem[];
}
