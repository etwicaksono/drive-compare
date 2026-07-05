# Task Breakdown — Duplicate Detection

**Last Updated:** 2026-07-06
**PRD:** `docs/TODO/Duplicate detection - PRD.md`

---

## Execution Rules

- Selesaikan **satu task** pada satu waktu.
- Setiap task selesai harus: build pass, test pass, format pass, lalu commit.
- Format commit: `feat(DUP-XXX): short description` atau `fix(DUP-XXX): ...`
- Jangan gabungkan beberapa task dalam satu commit.

### Branching Rules

- Setiap task dikerjakan di **branch baru** dari `main`.
- Format nama branch: `dup-XXX/short-description`.
- Setelah task selesai, **squash merge** ke `main`.
- Setelah merge, rename branch lama ke `merged/{nama-branch}`.

---

## Milestone 1 — Domain Engine

### DUP-001: Create duplication pattern registry

**Description:** Buat struct dan registry untuk built-in duplication patterns.

**Files:**
- `internal/domain/duplicate.go`

**Detail:**
- `DuplicatePattern` struct: `ID string`, `Label string`, `Regex *regexp.Regexp`.
- `BuiltinPatterns()` function yang return 4 patterns:
  - `windows_copy`: `^(.+)\s\((\d+)\)$`
  - `macos_copy`: `^(.+)\scopy(?:\s(\d+))?$`
  - `underscore_num`: `^(.+)_(\d+)$`
  - `dash_num`: `^(.+)-(\d+)$`
- `ExtractBaseName(name string, patterns []DuplicatePattern) string` — apply patterns terhadap name (tanpa extension), return base name. Jika tidak match pattern mana pun, return name as-is.

**Acceptance Criteria:**
- `BuiltinPatterns()` return 4 patterns dengan regex valid.
- `ExtractBaseName("photo (1)", patterns)` → `"photo"`.
- `ExtractBaseName("report copy 2", patterns)` → `"report"`.
- `ExtractBaseName("data_3", patterns)` → `"data"`.
- `ExtractBaseName("file-7", patterns)` → `"file"`.
- `ExtractBaseName("normal", patterns)` → `"normal"` (no match).
- Unit test untuk semua contoh di atas + edge cases.

---

### DUP-002: Create duplicate detection engine

**Description:** Buat fungsi `DetectDuplicates` yang menerima list `FileItem` dan list pattern, lalu return duplicate groups.

**Files:**
- `internal/domain/duplicate.go` (extend)
- `internal/domain/duplicate_test.go`

**Detail:**
- `DuplicateGroup` struct: `BaseName string`, `Extension string`, `RelativePath string`, `Files []FileItem`.
- `DetectDuplicates(files []FileItem, patterns []DuplicatePattern) []DuplicateGroup`.
- Logika:
  1. Untuk setiap file, extract base name menggunakan `ExtractBaseName(file.Name, patterns)`.
  2. Group key = `baseName + extension + relativePath`.
  3. Kumpulkan file-file per group key.
  4. Filter: hanya return group yang memiliki > 1 file.
  5. Sort groups by base name ascending.

**Acceptance Criteria:**
- `["photo.jpg", "photo (1).jpg", "photo (2).jpg"]` di path sama → 1 group, 3 files.
- `["report.pdf", "report copy.pdf"]` → 1 group, 2 files.
- `["unique.txt"]` → tidak masuk hasil (hanya 1 file).
- `["photo.jpg" di /a/, "photo (1).jpg" di /b/]` → tidak digroup (beda relativePath).
- File tanpa extension: `["Makefile", "Makefile (1)"]` → 1 group.
- Empty input → empty result.
- Unit test minimal 8 test cases.

---

### DUP-003: Add pattern filtering by ID

**Description:** Tambahkan kemampuan memilih subset patterns berdasarkan ID.

**Files:**
- `internal/domain/duplicate.go` (extend)

**Detail:**
- `GetPatternsByIDs(ids []string) []DuplicatePattern` — filter `BuiltinPatterns()` berdasarkan list ID.
- Validasi: ID yang tidak dikenal diabaikan (tidak error).

**Acceptance Criteria:**
- `GetPatternsByIDs(["windows_copy"])` return 1 pattern.
- `GetPatternsByIDs(["windows_copy", "macos_copy"])` return 2 patterns.
- `GetPatternsByIDs(["nonexistent"])` return empty slice.
- `GetPatternsByIDs([])` return empty slice.
- Unit test.

---

## Milestone 2 — Backend Binding

### DUP-004: Create RunDuplicateScan Wails binding

**Description:** Buat Wails binding yang menjalankan scan + duplicate detection dan return hasil ke frontend.

**Files:**
- `app.go`

**Detail:**
- `DuplicateGroupResponse` struct:
  ```go
  type DuplicateGroupResponse struct {
      BaseName     string              `json:"baseName"`
      Extension    string              `json:"extension"`
      RelativePath string              `json:"relativePath"`
      Files        []CompareResultItem `json:"files"`
  }
  ```
- `DuplicateScanResponse` struct:
  ```go
  type DuplicateScanResponse struct {
      Groups     []DuplicateGroupResponse `json:"groups"`
      TotalGroups int                     `json:"totalGroups"`
      TotalFiles  int                     `json:"totalFiles"`
  }
  ```
- `RunDuplicateScan(sourceType string, sourceID string, patternIDs []string) (*DuplicateScanResponse, error)`:
  1. Validasi: sourceType harus `"local"` atau `"drive"`.
  2. Scan menggunakan `LocalScanner` atau `DriveScanner` sesuai sourceType.
  3. `GetPatternsByIDs(patternIDs)` untuk filter patterns.
  4. `DetectDuplicates(files, patterns)` untuk deteksi.
  5. Return `DuplicateScanResponse`.

**Acceptance Criteria:**
- `sourceType = "local"` → scan pakai LocalScanner.
- `sourceType = "drive"` → scan pakai DriveScanner (harus connected).
- `sourceType = "invalid"` → return error.
- `patternIDs` kosong → return error "at least one pattern is required".
- Build pass, Wails bindings ter-generate.

---

### DUP-005: Create GetDuplicatePatterns Wails binding

**Description:** Buat binding untuk frontend mendapatkan daftar available patterns.

**Files:**
- `app.go`

**Detail:**
- `DuplicatePatternInfo` struct:
  ```go
  type DuplicatePatternInfo struct {
      ID      string `json:"id"`
      Label   string `json:"label"`
      Example string `json:"example"`
  }
  ```
- `GetDuplicatePatterns() []DuplicatePatternInfo` — return daftar built-in patterns dengan label dan contoh.

**Acceptance Criteria:**
- Return 4 patterns dengan ID, Label, dan Example.
- Build pass, Wails bindings ter-generate.

---

## Milestone 3 — Frontend UI

### DUP-006: Add Duplicates route and sidebar link

**Description:** Tambahkan halaman Duplicates ke routing dan sidebar navigation.

**Files:**
- `frontend/src/App.tsx`
- `frontend/src/components/Sidebar.tsx`
- `frontend/src/pages/DuplicatesPage.tsx` (placeholder)

**Acceptance Criteria:**
- Link "Duplicates" muncul di sidebar (antara Compare dan Settings).
- Klik navigasi ke `/duplicates`.
- Halaman placeholder dengan heading "Duplicates" ditampilkan.
- Build pass.

---

### DUP-007: Create DuplicatesPage scan form

**Description:** Buat form untuk memilih scan source, patterns, dan tombol scan.

**Files:**
- `frontend/src/pages/DuplicatesPage.tsx`
- `frontend/src/lib/types.ts` (tambah types)

**Detail:**
- **Source picker:**
  - Radio button: Local Directory / Google Drive.
  - Local: input readonly + Browse button (reuse `SelectLocalDirectory`).
  - Drive: input readonly + Browse button (reuse `DriveFolderPicker`).
- **Pattern selector:**
  - Fetch patterns dari `GetDuplicatePatterns()` saat mount.
  - Checkbox per pattern dengan label dan contoh.
  - Default: semua tercentang.
- **Scan button:**
  - Disabled jika: source belum dipilih, atau tidak ada pattern tercentang.
  - Label: "Scan for Duplicates" / "Scanning...".
- **Loading state:** spinner saat scan berjalan.
- **Error state:** ErrorBanner untuk error.

**Acceptance Criteria:**
- Source picker (local + Drive) berfungsi.
- Pattern checkboxes muncul dari backend data.
- Scan button disabled/enabled sesuai kondisi.
- Loading dan error state tampil.
- Build pass.

---

### DUP-008: Create DuplicateResultsTable component

**Description:** Buat tabel grouped untuk menampilkan hasil duplicate detection.

**Files:**
- `frontend/src/components/DuplicateResultsTable.tsx`

**Detail:**
- **Summary bar:** "Found X duplicate groups (Y files total)".
- **Group header row:**
  - Collapse/expand toggle (▼/▶).
  - Base filename + extension.
  - File count badge: "(N files)".
  - Default: expanded.
- **File rows (di dalam group):**
  - Kolom: No, Filename, Extension, Size, Relative Path.
  - Sama seperti ResultsTable.
- **Empty state:** "No duplicates found with the selected patterns."

**Acceptance Criteria:**
- Groups ditampilkan dengan header collapsible.
- Collapse/expand berfungsi per group.
- File rows muncul di dalam group.
- Summary bar menunjukkan jumlah group dan total files.
- Empty state muncul jika tidak ada duplikat.
- Build pass.

---

### DUP-009: Add search and filtering to DuplicateResultsTable

**Description:** Tambahkan search dan filter ke tabel duplikat.

**Files:**
- `frontend/src/components/DuplicateResultsTable.tsx`

**Detail:**
- **Search:** filter across groups berdasarkan filename (case-insensitive).
- **Extension filter:** multi-select toggle buttons (sama seperti ResultsTable).
- **Size filter:** dropdown range (sama seperti ResultsTable).
- Filter diterapkan pada file-file di dalam group.
- Group yang tidak memiliki file setelah filter → di-hide.
- Summary bar update sesuai hasil filter: "Showing X of Y groups (Z files)".

**Acceptance Criteria:**
- Search berfungsi across groups.
- Extension dan size filter berfungsi.
- Group kosong setelah filter tidak ditampilkan.
- Summary bar update.
- Build pass.

---

### DUP-010: Add sorting to DuplicateResultsTable

**Description:** Tambahkan sorting di dalam setiap group.

**Files:**
- `frontend/src/components/DuplicateResultsTable.tsx`

**Detail:**
- Sorting berlaku di dalam setiap group (bukan antar group).
- Kolom sortable: Filename, Extension, Size, Relative Path.
- Sort indicator (▲/▼) di header.
- Default: Filename ascending.
- Group order: tetap sorted by base name ascending.

**Acceptance Criteria:**
- Klik header kolom → sort di dalam semua groups.
- Sort indicator muncul.
- Default sort Filename asc.
- Build pass.

---

### DUP-011: Wire up DuplicatesPage with backend

**Description:** Hubungkan form scan dengan backend binding dan tampilkan hasil.

**Files:**
- `frontend/src/pages/DuplicatesPage.tsx`

**Detail:**
- Klik "Scan for Duplicates" → call `RunDuplicateScan(sourceType, sourceID, patternIDs)`.
- Response di-pass ke `DuplicateResultsTable`.
- Error ditampilkan di ErrorBanner.
- Loading state aktif selama scan.

**Acceptance Criteria:**
- Scan local directory → results muncul di tabel.
- Scan Google Drive → results muncul di tabel.
- Error handling: Drive not connected, folder not found, dll.
- Build pass.

---

## Milestone 4 — Testing

### DUP-012: Comprehensive testing

**Description:** Tambahkan test coverage untuk seluruh fitur duplicate detection.

**Files:**
- `internal/domain/duplicate_test.go` (extend)

**Acceptance Criteria:**
- Unit test untuk `ExtractBaseName` — semua 4 patterns + edge cases.
- Unit test untuk `DetectDuplicates` — grouping, filtering, empty input, single file.
- Unit test untuk `GetPatternsByIDs` — valid IDs, invalid IDs, empty.
- Edge cases:
  - File tanpa extension.
  - Pattern overlap (file match multiple patterns → same group).
  - Nested copy `photo (1) (1)` — single-pass behavior.
  - Case sensitivity.
  - Very long filenames.
- Coverage minimal 90% untuk `duplicate.go`.
- Semua test pass.
