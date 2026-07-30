# Task Breakdown — Deleted Drive Files Comparison

**Last Updated:** 2026-07-30
**PRD:** `docs/IN_PROGRESS/Deleted drive files comparison - PRD.md`

---

## Execution Rules

- Selesaikan **satu task** pada satu waktu.
- Setiap task selesai harus: build pass, test pass, format pass, lalu commit.
- Format commit: `feat(TRASH-XXX): short description` atau `fix(TRASH-XXX): ...`
- Jangan gabungkan beberapa task dalam satu commit.

### Branching Rules

- Setiap task dikerjakan di **branch baru** dari `main`.
- Format nama branch: `trash-XXX/short-description`.
- Setelah task selesai, **squash merge** ke `main`.
- Setelah merge, rename branch lama ke `merged/{nama-branch}`.

---

## Milestone 1 — Domain & Scanner

### TRASH-001: Add `Deleted` marker to FileItem

**Description:** Tambahkan penanda status terhapus pada entity file.

**Files:**
- `internal/domain/entity.go`
- `internal/domain/compare_test.go`, `internal/domain/resize_audit_test.go`, `internal/domain/duplicate_test.go` (sesuaikan jika perlu)

**Detail:**
- `FileItem` menjadi:
  ```go
  type FileItem struct {
      RelativePath string
      Name         string
      Extension    string
      Size         int64
      Deleted      bool
  }
  ```
- **Jangan** ubah `resizeKey` maupun compare key — matching tetap berbasis `relativePath + name + extension` supaya file Trash bisa dipasangkan dengan pasangan aktifnya.

**Acceptance Criteria:**
- Zero value `Deleted = false` untuk semua pembuatan `FileItem` existing.
- `CompareFiles`, `ResizeAudit`, `DetectDuplicates` tetap lulus semua test tanpa perubahan logika.
- `make be-test` pass, `make be-fmt-check` pass.

---

### TRASH-002: Parameterize Drive query with TrashScope

**Description:** Buat `DriveScanner` bisa scan file aktif, file Trash, atau keduanya.

**Files:**
- `internal/infrastructure/scanner/drive_scanner.go`

**Detail:**
- Tambah type dan konstanta:
  ```go
  type TrashScope string

  const (
      ScopeActive  TrashScope = "active"
      ScopeTrashed TrashScope = "trashed"
      ScopeAll     TrashScope = "all"
  )
  ```
- `NewDriveScanner(api DriveAPI, ignorePatterns []string, scope TrashScope)` — scope kosong (`""`) di-normalisasi ke `ScopeActive`.
- `scanFolder` menyusun query:
  - `ScopeActive` → `'<id>' in parents and trashed = false`
  - `ScopeTrashed` → `'<id>' in parents and trashed = true`
  - `ScopeAll` → `'<id>' in parents`
- Scope tidak dikenal → return error `invalid drive scope %q`.
- Pada scope `trashed`/`all`, folder dengan `trashed = true` **tetap** ditelusuri (jangan di-skip).
- Isi `Deleted: f.Trashed` pada `domain.FileItem` yang dihasilkan.
- Verifikasi manual terhadap Drive nyata: hapus 1 file langsung + hapus 1 folder berisi file, lalu cek mana yang terdeteksi. Catat hasilnya sebagai komentar/limitation di kode bila traversal folder ter-trash tidak reliable.

**Acceptance Criteria:**
- Query yang dikirim ke API sesuai tabel di atas per scope.
- `Deleted` terisi sesuai `f.Trashed`.
- Scope invalid → error, tidak panic.
- Build pass, test existing pass.

---

### TRASH-003: Honor `fields` parameter in realDriveAPI

**Description:** `realDriveAPI.ListFiles` saat ini mengabaikan argumen `fields`; perbaiki dan minta field `trashed`.

**Files:**
- `internal/infrastructure/scanner/drive_scanner.go`

**Detail:**
- Jika `fields` kosong → default `"nextPageToken, files(id, name, size, mimeType, trashed)"`.
- Jika `fields` diisi caller → pakai nilai itu.
- `DriveScanner.scanFolder` mengirim fields eksplisit yang menyertakan `trashed`.
- `GetFile` juga menyertakan `trashed` di field list.

**Acceptance Criteria:**
- Tidak ada lagi hardcode `Fields(...)` yang mengabaikan argumen.
- Scan mode `all` bisa membedakan item aktif vs terhapus.
- Build pass.

---

### TRASH-004: Set `Deleted = false` explicitly in LocalScanner

**Description:** Samakan kontrak entity untuk source lokal.

**Files:**
- `internal/infrastructure/scanner/local_scanner.go`

**Detail:**
- Isi `Deleted: false` secara eksplisit saat membangun `domain.FileItem` supaya intent-nya jelas dan tahan terhadap perubahan default di masa depan.

**Acceptance Criteria:**
- Semua item hasil local scan punya `Deleted = false`.
- Test local scanner pass.

---

## Milestone 2 — Backend Binding

### TRASH-005: Thread scope through scanSource

**Description:** Teruskan scope dari frontend ke Drive scanner via choke point `scanSource`.

**Files:**
- `app.go`

**Detail:**
- Ubah signature: `scanSource(sourceType string, sourceID string, scope string) ([]domain.FileItem, error)`.
- `sourceType = "local"` → scope diabaikan (tidak error walau diisi).
- `sourceType = "drive"` → validasi scope (`""`/`active`/`trashed`/`all`), lalu `scanner.NewDriveScanner(api, patterns, scanner.TrashScope(scope))`.
- Scope invalid → error `invalid drive scope %q, must be "active", "trashed", or "all"`.
- Log scope di `slog.Info` saat scan Drive.

**Acceptance Criteria:**
- Semua caller `scanSource` ter-update dan build pass.
- Scope invalid → error message jelas.

---

### TRASH-006: Expose `deleted` in DTO and update bindings

**Description:** Bawa status terhapus ke frontend dan tambahkan parameter scope pada binding.

**Files:**
- `app.go`
- `frontend/wailsjs/**` (regenerate)

**Detail:**
- `CompareResultItem` tambah `Deleted bool \`json:"deleted"\``; `toResultItems` mengisinya dari `domain.FileItem.Deleted`.
- Signature binding baru:
  - `RunComparison(sourceTypeA, sourceIDA, scopeA, sourceTypeB, sourceIDB, scopeB string)`
  - `RunResizeAudit(sourceType, sourceID, sourceScope, resizedType, resizedID, resizedScope, suffix string)`
  - `RunDuplicateScan(sourceType, sourceID, scope string, patternIDs []string)`
- Regenerate bindings (`wails build` / `make build`) agar `App.d.ts`, `App.js`, `models.ts` ikut update.

**Acceptance Criteria:**
- Build pass, bindings ter-generate dengan signature baru.
- Response JSON mengandung `deleted` per item.
- Tidak ada caller frontend yang masih memakai signature lama (dicek saat TRASH-007/008).

---

## Milestone 3 — Frontend

### TRASH-007: Add DriveScope types and export marker

**Description:** Siapkan tipe dan helper di layer types.

**Files:**
- `frontend/src/lib/types.ts`

**Detail:**
- `export type DriveScope = 'active' | 'trashed' | 'all';`
- `FileItem` tambah `deleted: boolean;`.
- `getItemKey` tambahkan penanda status agar item aktif dan item Trash dengan nama sama tidak bentrok saat selection: `${relativePath}${name}${extension}${deleted ? '#deleted' : ''}`.
- `buildLocalPath`/`buildDrivePath` tidak berubah; `buildExportText` menambahkan ` [deleted]` di akhir baris untuk item `deleted = true`.

**Acceptance Criteria:**
- Typecheck pass (`make fe-build`).
- Selection tidak bentrok untuk pasangan aktif/terhapus bernama sama.
- Export text menampilkan penanda `[deleted]`.

---

### TRASH-008: Add scope dropdown to Resize Audit source rows

**Description:** Tambahkan filter scope di halaman Resize Audit.

**Files:**
- `frontend/src/pages/ResizeAuditPage.tsx`

**Detail:**
- `SourceState` tambah `scope: DriveScope` (default `'active'`).
- `SourceRow`: jika `source.type === 'drive'`, render `<select>` kedua dengan opsi:
  - `active` → "Active files"
  - `trashed` → "Deleted (Trash)"
  - `all` → "Active + Deleted"
  Disembunyikan saat type `local`.
- `handleAudit` memanggil `RunResizeAudit(sourceDir.type, sourceID, sourceDir.scope, resizedDir.type, resizedID, resizedDir.scope, suffix)`.
- Map `deleted` dari response ke state hasil.
- Ganti type source ke `local` → scope direset ke `active`.

**Acceptance Criteria:**
- Dropdown scope hanya muncul untuk source Drive.
- Audit dengan scope `trashed` menghasilkan file dari Trash.
- Loading/error/success state tetap berfungsi.
- `make fe-lint` + `make fe-build` pass.

---

### TRASH-009: Add scope dropdown to Compare page

**Description:** Perubahan setara untuk halaman Compare (source A dan B).

**Files:**
- `frontend/src/pages/ComparePage.tsx`

**Detail:**
- Pola state dan UI mengikuti TRASH-008.
- Panggil `RunComparison` dengan scope A dan B.

**Acceptance Criteria:**
- Compare local↔drive-trashed, drive-active↔drive-trashed berjalan.
- Build + lint pass.

---

### TRASH-010: Status column and filter in ResultsTable

**Description:** Tampilkan dan filter status Active/Deleted di tabel hasil.

**Files:**
- `frontend/src/components/ResultsTable.tsx`

**Detail:**
- Prop baru `showStatusColumn?: boolean` — halaman mengaktifkannya jika ada minimal satu item `deleted`.
- Kolom "Status": `Active` (netral) atau `Deleted` (warna amber/merah).
- Toolbar tambah dropdown status: `All` / `Active only` / `Deleted only`, ikut memengaruhi count yang ditampilkan pada baris hasil (bukan angka pada tab).
- Kolom Status **tidak** sortable di V1.

**Acceptance Criteria:**
- Kolom Status muncul hanya saat ada item terhapus.
- Filter status mempersempit baris dengan benar dan kompatibel dengan search/extension/size filter existing.
- Incremental rendering (`BATCH_SIZE`) tetap berfungsi.
- Build + lint pass.

---

## Milestone 4 — Testing

### TRASH-011: Scanner unit tests per scope

**Description:** Perluas test Drive scanner untuk semua scope.

**Files:**
- `internal/infrastructure/scanner/drive_scanner_test.go`

**Detail:**
- `mockDriveAPI.ListFiles` saat ini mencocokkan query persis `'%s' in parents and trashed = false`. Refactor agar mengenali tiga bentuk query (active / trashed / all) dan mengembalikan subset fixture yang sesuai berdasarkan flag `Trashed` pada `*drive.File`.
- Tambahkan `addFiles` fixture dengan `Trashed: true`.
- Test cases baru:
  - scope `active` → item trashed tidak muncul (regresi perilaku lama).
  - scope `trashed` → hanya item trashed, `Deleted = true`.
  - scope `all` → keduanya muncul dengan `Deleted` sesuai.
  - scope kosong → diperlakukan sebagai `active`.
  - scope invalid → error.
  - folder trashed di scope `trashed` → isinya ikut ditelusuri.
  - nama sama aktif + trashed pada scope `all` → dua entry.
  - ignore patterns tetap berlaku pada item trashed.
  - Google Docs di Trash tetap di-skip.

**Acceptance Criteria:**
- Semua test baru pass.
- Coverage `drive_scanner.go` ≥ 90% (`make be-test-cover`).

---

### TRASH-012: End-to-end verification and doc update

**Description:** Verifikasi manual terhadap Google Drive nyata dan tutup dokumentasi.

**Files:**
- `docs/IN_PROGRESS/Deleted drive files comparison - PRD.md` → pindah ke `docs/DONE/` setelah selesai

**Detail:**
- Skenario manual:
  1. Upload `video.mp4` + `video_resized.mp4` ke Drive, hapus `video_resized.mp4`.
  2. Resize Audit dengan resized scope `active` → Pending.
  3. Ulangi dengan scope `trashed` dan `all` → Resized + badge Deleted.
  4. Hapus satu **folder** berisi file, cek apakah isinya terdeteksi pada scope `trashed`.
  5. Export hasil ke clipboard, pastikan penanda `[deleted]` muncul.
- Catat hasil langkah 4 di PRD (Edge Cases) sebagai perilaku terverifikasi atau limitation.

**Acceptance Criteria:**
- Semua skenario terverifikasi dan hasilnya tercatat di PRD.
- `make be-test`, `make be-fmt-check`, `make fe-lint`, `make fe-build` pass.
- Dokumen PRD + TASKS dipindahkan ke `docs/DONE/`.

**Status: BLOCKED** — butuh akun Google Drive nyata + interaksi GUI, tidak bisa
dijalankan otomatis. Hasil gate otomatis dan sisa checklist manual dicatat di
PRD §9. Dokumen tetap di `docs/IN_PROGRESS/` sampai skenario manual selesai.
