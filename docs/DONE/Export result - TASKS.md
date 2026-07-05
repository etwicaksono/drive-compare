# Task Breakdown — Export Comparison Result

**Last Updated:** 2026-07-06
**PRD:** `docs/TODO/Export result - PRD.md`

---

## Execution Rules

- Selesaikan **satu task** pada satu waktu.
- Setiap task selesai harus: build pass, test pass, format pass, lalu commit.
- Format commit: `feat(EXP-XXX): short description` atau `fix(EXP-XXX): ...`
- Jangan gabungkan beberapa task dalam satu commit.

### Branching Rules

- Setiap task dikerjakan di **branch baru** dari `main`.
- Format nama branch: `exp-XXX/short-description`.
- Setelah task selesai, **squash merge** ke `main`.
- Setelah merge, rename branch lama ke `merged/{nama-branch}`.

---

## Milestone 1 — Backend Bindings

### EXP-001: Create ExportToClipboard Wails binding

**Description:** Buat Wails binding untuk menulis text ke system clipboard.

**Files:**
- `app.go`

**Detail:**
- `ExportToClipboard(text string) error` — panggil `runtime.ClipboardSetText(a.ctx, text)`.
- Validasi: text kosong → return error.

**Acceptance Criteria:**
- Binding ter-generate di `frontend/wailsjs/go/main/App.d.ts`.
- Text kosong → return error `"nothing to export"`.
- Build pass.

---

### EXP-002: Create ExportToFile Wails binding

**Description:** Buat Wails binding untuk menyimpan text ke file .txt via Save dialog.

**Files:**
- `app.go`

**Detail:**
- `ExportToFile(content string) (string, error)`:
  1. Validasi: content kosong → return error.
  2. Buka `runtime.SaveFileDialog()` dengan filter `*.txt`.
  3. Jika user cancel dialog → return empty string, no error.
  4. Tulis content ke file yang dipilih via `os.WriteFile()`.
  5. Return path file yang disimpan.

**Acceptance Criteria:**
- Content kosong → return error `"nothing to export"`.
- User cancel dialog → return `""`, nil.
- File berhasil ditulis → return path file.
- Build pass.

---

## Milestone 2 — Selection UI

### EXP-003: Add selection state to ResultsTable

**Description:** Tambahkan selection state management ke ResultsTable.

**Files:**
- `frontend/src/components/ResultsTable.tsx`
- `frontend/src/lib/types.ts`

**Detail:**
- Tambahkan props baru di `ResultsTableProps`:
  - `selectedKeys: Set<string>` — set of selected item keys.
  - `onSelectionChange: (keys: Set<string>) => void` — callback saat selection berubah.
- Item key: `${item.relativePath}${item.name}${item.extension}` (sudah dipakai sebagai React key).
- Selection state di-maintain di `ComparePage` per tab (3 Set terpisah).
- Tambah helper `getItemKey(item: FileItem): string` di `types.ts`.

**Acceptance Criteria:**
- Props `selectedKeys` dan `onSelectionChange` diterima oleh ResultsTable.
- ComparePage maintain 3 Set selection (per tab).
- Tab switch → selection yang aktif mengikuti tab.
- Build pass.

---

### EXP-004: Add hover checkbox to No column

**Description:** Ubah kolom No agar menampilkan checkbox saat hover.

**Files:**
- `frontend/src/components/ResultsTable.tsx`

**Detail:**
- **Default:** tampilkan nomor urut.
- **Hover row:** nomor berubah menjadi checkbox.
  - Gunakan CSS `group` / `group-hover` dari Tailwind.
  - Checkbox unchecked: `□` (neutral border).
  - Checkbox checked: `☑` (blue fill).
- **Row yang sudah diselect:** checkbox selalu tampil (tidak perlu hover), menggantikan nomor.
- **Klik checkbox:** toggle item di `selectedKeys` via `onSelectionChange`.
- Lebar kolom No tetap sama (`w-14`), tidak bergeser saat toggle.

**Acceptance Criteria:**
- Hover row → checkbox muncul menggantikan nomor.
- Klik checkbox → item ter-select/deselect.
- Row yang sudah select → checkbox selalu tampil.
- Tidak ada layout shift saat toggle.
- Build pass.

---

### EXP-005: Add Select All checkbox to header

**Description:** Tambahkan Select All / Deselect All di header kolom No.

**Files:**
- `frontend/src/components/ResultsTable.tsx`

**Detail:**
- Header "No" saat hover → tampilkan checkbox.
- State checkbox header:
  - Unchecked: tidak ada item terselect di filter saat ini.
  - Checked: semua item yang lolos filter saat ini terselect.
  - Indeterminate: sebagian item terselect.
- Klik → toggle:
  - Jika belum semua terselect → select all filtered items.
  - Jika sudah semua → deselect all filtered items.
- **Penting:** Select All hanya berlaku untuk item yang lolos filter, bukan semua data di tab.

**Acceptance Criteria:**
- Header hover → checkbox muncul.
- Select All → semua filtered items ter-select.
- Deselect All → semua filtered items ter-deselect.
- Indeterminate state saat sebagian terselect.
- Build pass.

---

### EXP-006: Add selection counter and Export button to toolbar

**Description:** Tampilkan counter selected items dan tombol Export di toolbar filter.

**Files:**
- `frontend/src/components/ResultsTable.tsx`

**Detail:**
- Di toolbar filter (baris yang berisi search, extension filter, size filter):
  - Tambahkan di sebelah kanan (setelah "X of Y files"):
    - `"N selected"` — hanya muncul jika ada item terselect.
    - Tombol `"Export Selected"` — hanya muncul jika ada item terselect.
- Tombol Export memanggil callback `onExport()` yang di-pass dari parent.
- Tambah props `onExport: () => void` di `ResultsTableProps`.

**Acceptance Criteria:**
- Counter "N selected" muncul saat ada selection.
- Tombol "Export Selected" muncul saat ada selection.
- Tidak ada selection → counter dan tombol hidden.
- Klik Export → callback `onExport()` dipanggil.
- Build pass.

---

## Milestone 3 — Export Modal & Wiring

### EXP-007: Create ExportModal component

**Description:** Buat modal Export dengan pilihan path source dan export target.

**Files:**
- `frontend/src/components/ExportModal.tsx` (new)

**Detail:**
- Props:
  - `open: boolean`
  - `onClose: () => void`
  - `onExport: (pathSource: 'local' | 'drive', target: 'clipboard' | 'file') => void`
  - `selectedCount: number`
  - `activeTab: TabKey`
- UI:
  - Heading: `"Export {N} selected files"`.
  - **Path source** (radio buttons):
    - Tab `localOnly` → tidak tampil, otomatis `local`.
    - Tab `driveOnly` → tidak tampil, otomatis `drive`.
    - Tab `both` → radio: "Local Storage" / "Google Drive", default "Local Storage".
  - **Export to** (radio buttons): "Clipboard" / "Save as .txt file", default "Clipboard".
  - Tombol Cancel dan Export.
- Styling konsisten dengan DriveFolderPicker modal (border, bg, padding).

**Acceptance Criteria:**
- Modal muncul/hilang sesuai `open` prop.
- Tab localOnly → path source hidden, fixed local.
- Tab driveOnly → path source hidden, fixed drive.
- Tab both → radio path source muncul.
- Export to radio berfungsi.
- Klik Export → callback `onExport` dipanggil dengan nilai yang benar.
- Klik Cancel → `onClose` dipanggil.
- Build pass.

---

### EXP-008: Implement path builder utility

**Description:** Buat utility function untuk menyusun full path dari item data.

**Files:**
- `frontend/src/lib/types.ts` (extend)

**Detail:**
- `buildLocalPath(localDir: string, item: FileItem): string`
  - Format: `{localDir}\{relativePath}{name}{ext}`
  - `relativePath` menggunakan `/` → convert ke `\` untuk Windows.
  - Jika `relativePath` kosong → `{localDir}\{name}{ext}`.
- `buildDrivePath(driveLabel: string, item: FileItem): string`
  - Format: `{driveLabel}/{relativePath}{name}{ext}`
  - Jika `relativePath` kosong → `{driveLabel}/{name}{ext}`.
- `buildExportText(items: FileItem[], pathSource: 'local' | 'drive', localDir: string, driveLabel: string): string`
  - Susun path per item, gabung dengan newline `\n`.

**Acceptance Criteria:**
- `buildLocalPath("D:\\Photos", {relativePath: "2024/", name: "photo", extension: ".jpg"})` → `"D:\\Photos\\2024\\photo.jpg"`.
- `buildLocalPath("D:\\Photos", {relativePath: "", name: "photo", extension: ".jpg"})` → `"D:\\Photos\\photo.jpg"`.
- `buildDrivePath("My Drive/Photos", {relativePath: "2024/", name: "photo", extension: ".jpg"})` → `"My Drive/Photos/2024/photo.jpg"`.
- `buildExportText(...)` → multi-line string.
- Build pass.

---

### EXP-009: Wire up export flow in ComparePage

**Description:** Hubungkan semua komponen: selection → modal → path builder → backend binding.

**Files:**
- `frontend/src/pages/ComparePage.tsx`

**Detail:**
- State baru:
  - `showExportModal: boolean`
  - `selections: Record<TabKey, Set<string>>` (per-tab selection).
- Pass `localDir` dan `driveDirLabel` ke ResultsTable (untuk path building).
- `onExport` callback dari ResultsTable → buka ExportModal.
- ExportModal `onExport(pathSource, target)`:
  1. Ambil selected items dari current tab data.
  2. `buildExportText(selectedItems, pathSource, localDir, driveDirLabel)`.
  3. Jika target = `clipboard` → panggil `ExportToClipboard(text)`.
  4. Jika target = `file` → panggil `ExportToFile(text)`.
  5. Tutup modal, tampilkan success/error feedback.
- Success feedback: toast-like message "Exported N files to clipboard" / "Saved to {path}".

**Acceptance Criteria:**
- Full flow: select items → Export Selected → modal → pilih options → Export → success.
- Clipboard export berfungsi (bisa di-paste).
- File export berfungsi (file .txt tersimpan dengan isi yang benar).
- Error handling: export gagal → ErrorBanner.
- Tab localOnly → path source otomatis local.
- Tab driveOnly → path source otomatis drive.
- Tab both → user pilih path source di modal.
- Build pass.

---

## Milestone 4 — Polish & Testing

### EXP-010: Reset selection on tab switch and filter change

**Description:** Pastikan selection behavior konsisten saat tab switch dan filter berubah.

**Files:**
- `frontend/src/pages/ComparePage.tsx`
- `frontend/src/components/ResultsTable.tsx`

**Detail:**
- Tab switch → selection di-reset untuk tab baru (masing-masing tab punya Set sendiri, yang sudah di-maintain di ComparePage).
- Filter/search berubah → selection TIDAK di-reset (item yang sudah terselect tapi ter-filter-out tetap dipertahankan di Set, akan muncul kembali jika filter di-clear).
- Re-compare (klik Compare ulang) → semua selection di-reset.

**Acceptance Criteria:**
- Tab switch → selection aktif mengikuti tab tersebut.
- Filter berubah → selection tidak hilang.
- Re-compare → selection di-reset semua.
- Build pass.

---

### EXP-011: Testing and edge cases

**Description:** Verifikasi semua edge cases dan pastikan build clean.

**Files:**
- Semua file yang diubah.

**Acceptance Criteria:**
- Select All → hanya item yang lolos filter.
- Export 0 items → tidak mungkin (tombol hidden).
- Path separator benar: `\` untuk local (Windows), `/` untuk Drive.
- Export file sangat besar (10,000+ items) → tidak hang.
- Modal dismiss (Cancel / click outside) → tidak ada side effect.
- TSC pass, build pass, semua Go test pass.
