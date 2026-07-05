export interface FileItem {
  relativePath: string;
  name: string;
  extension: string;
  size: number;
}

export interface CompareResult {
  localOnly: FileItem[];
  driveOnly: FileItem[];
  both: FileItem[];
}

export type TabKey = 'localOnly' | 'driveOnly' | 'both';

export function formatSize(bytes: number): string {
  if (bytes === 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const k = 1024;
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  const value = bytes / Math.pow(k, i);
  return `${value.toFixed(i === 0 ? 0 : 1)} ${units[i]}`;
}

export function getItemKey(item: FileItem): string {
  return `${item.relativePath}${item.name}${item.extension}`;
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
  return items.map((item) => builder(base, item)).join('\n');
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
