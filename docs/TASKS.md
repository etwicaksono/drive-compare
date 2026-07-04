# Task Breakdown

## Personal Digital Inventory

**Last Updated:** 2026-07-04

---

## Execution Rules

- Selesaikan **satu task** pada satu waktu.
- Setiap task selesai harus: build pass, test pass, format pass, lalu commit.
- Format commit: `feat(TASK-XXX): short description` atau `fix(TASK-XXX): ...`
- Jangan gabungkan beberapa task dalam satu commit.
- Jangan implementasi fitur yang belum diminta.
- Berhenti setelah commit, tunggu instruksi untuk task berikutnya.

### Branching Rules

- Setiap task dikerjakan di **branch baru** dari `main`.
- Format nama branch: `task-XXX/short-description` (e.g. `task-001/init-wails-project`).
- Setelah task selesai dan di-review/disetujui, **squash merge** ke `main`.
- Hapus branch setelah merge.
- Jangan push langsung ke `main`.

---

## Global Acceptance Criteria

Setiap task **wajib** memenuhi semua kriteria berikut:

- Build berhasil tanpa error.
- Tidak ada compiler warning.
- Code sudah diformat (`gofmt` / `prettier`).
- Linter pass.
- Unit test pass.
- Tidak ada TODO di production code.
- Tidak ada dead code atau unused dependencies.
- Commit dibuat sebelum berhenti.

---

## Milestone 1 — Project Foundation

### TASK-001: Create Go + Wails project

**Description:** Inisialisasi project Go dengan Wails, React, dan TypeScript.

**Acceptance Criteria:**
- Project Go dengan `go.mod` sudah dibuat.
- Wails project ter-inisialisasi dengan React + TypeScript template.
- `wails build` berhasil.
- `wails dev` bisa menampilkan window aplikasi kosong.

---

### TASK-002: Setup project folder structure

**Description:** Buat struktur folder Clean Architecture.

**Acceptance Criteria:**
- Folder structure dibuat:
  ```
  internal/
      domain/
      application/
      infrastructure/
  frontend/
      src/
          pages/
          components/
          lib/
  ```
- Tidak ada business logic, hanya folder dan placeholder.
- Build masih berhasil.

---

### TASK-003: Setup PostgreSQL connection with GORM

**Description:** Konfigurasi koneksi database PostgreSQL menggunakan GORM.

**Acceptance Criteria:**
- GORM terinstall dengan driver PostgreSQL.
- Koneksi ke database berhasil saat aplikasi start.
- Jika koneksi gagal, aplikasi menampilkan error message yang jelas.
- Unit test koneksi database tersedia.

---

### TASK-004: Create configuration loader

**Description:** Buat config loader yang membaca file `config.yaml`.

**Acceptance Criteria:**
- File `config.yaml` di-load saat startup.
- Mendukung konfigurasi: database, ignore patterns, logging level.
- Password database bisa di-override via environment variable `PDI_DB_PASSWORD`.
- Validasi config: jika field wajib kosong, tampilkan error yang jelas.
- Unit test untuk config loading dan validasi.

---

### TASK-005: Implement logging

**Description:** Setup logging menggunakan `log/slog`.

**Acceptance Criteria:**
- `slog` dikonfigurasi dengan level dari `config.yaml`.
- Log output ke console.
- Level yang didukung: debug, info, warn, error.
- Unit test untuk konfigurasi log level.

---

### TASK-006: Setup database migration with goose

**Description:** Setup mekanisme migration menggunakan goose.

**Acceptance Criteria:**
- goose terinstall dan dikonfigurasi.
- Migration dijalankan otomatis saat aplikasi start.
- Migration pertama: create tabel `comparisons`, `comparison_files`, `ignore_patterns`.
- Schema sesuai dengan PRD section 7.4.
- Migration bisa rollback.

---

### TASK-007: Create domain entities and repository interfaces

**Description:** Buat entity dan repository interface di domain layer.

**Acceptance Criteria:**
- Entity `Comparison`, `ComparisonFile`, `IgnorePattern` dibuat di `internal/domain/`.
- Repository interface `ComparisonRepository` dan `IgnorePatternRepository` dibuat.
- Tidak ada implementasi, hanya interface.
- Build berhasil.

---

### TASK-008: Implement GORM repositories

**Description:** Implementasi repository interface menggunakan GORM.

**Acceptance Criteria:**
- `GormComparisonRepository` mengimplementasi `ComparisonRepository`.
- `GormIgnorePatternRepository` mengimplementasi `IgnorePatternRepository`.
- CRUD operations: Create, FindByID, List, Delete.
- Unit test untuk setiap operasi CRUD (menggunakan test database).

---

## Milestone 2 — Compare Feature

### TASK-009: Implement local directory scanner

**Description:** Buat scanner yang membaca file metadata dari local directory secara rekursif.

**Acceptance Criteria:**
- Scan satu folder dan return list `FileItem` (relative path, name, extension, size).
- Recursive scan untuk nested folders.
- Relative path dihitung dari root directory yang dipilih user.
- Extension dipisahkan dari filename.
- Unit test dengan temporary directory.

---

### TASK-010: Implement ignore rules for scanner

**Description:** Tambahkan ignore rules ke local scanner.

**Acceptance Criteria:**
- Scanner membaca ignore patterns dari database.
- Glob pattern matching (e.g. `node_modules`, `.git`, `*.tmp`).
- File/folder yang match pattern di-skip saat scan.
- Hidden files/folders handling (configurable).
- Unit test untuk ignore pattern matching.

---

### TASK-011: Implement Google Drive OAuth2 authentication

**Description:** Implementasi OAuth2 flow untuk koneksi ke Google Drive API.

**Acceptance Criteria:**
- OAuth2 flow bisa di-trigger dari aplikasi.
- Token disimpan secara aman secara lokal.
- Refresh token otomatis saat expired.
- Status koneksi bisa dicek (connected/disconnected).
- Disconnect menghapus token yang tersimpan.

---

### TASK-012: Implement Google Drive scanner

**Description:** Buat scanner yang membaca file metadata dari Google Drive.

**Acceptance Criteria:**
- Scan folder Google Drive dan return list `FileItem`.
- Recursive scan untuk nested folders.
- Relative path dihitung dari root folder Drive yang dipilih user.
- Menangani Google Drive API rate limits (exponential backoff).
- Unit test dengan mock Google Drive API.

---

### TASK-013: Implement comparison engine

**Description:** Buat engine yang membandingkan hasil scan local dan Google Drive.

**Acceptance Criteria:**
- Menerima dua list `FileItem` (local dan drive).
- Membandingkan berdasarkan file identity (relative path + filename + size).
- Menghasilkan 3 kategori: local_only, drive_only, both.
- Menyimpan hasil comparison ke database (tabel `comparisons` dan `comparison_files`).
- Menghitung dan menyimpan total per kategori.
- Unit test untuk berbagai skenario comparison.

---

### TASK-014: Implement comparison use case

**Description:** Buat application layer use case yang mengorkestrasi scan dan compare.

**Acceptance Criteria:**
- Use case menerima local path dan drive path dari user.
- Menjalankan local scanner dan drive scanner.
- Menjalankan comparison engine.
- Menyimpan hasil ke database.
- Return hasil comparison untuk ditampilkan di UI.
- Error handling: jika salah satu scanner gagal, return error yang jelas.

---

### TASK-015: Create frontend layout with sidebar

**Description:** Buat layout utama aplikasi dengan sidebar navigation.

**Acceptance Criteria:**
- Sidebar kiri dengan menu: Compare, Settings.
- Area konten di kanan.
- React Router untuk navigasi antar halaman.
- Styling dengan Tailwind CSS.
- Tampilan mirip GitHub Desktop / Linear (clean, minimal).

---

### TASK-016: Create Compare page — folder picker section

**Description:** Buat bagian atas halaman Compare untuk memilih folder.

**Acceptance Criteria:**
- Input field + Browse button untuk Local Directory.
- Input field + Browse button untuk Google Drive Directory.
- Local directory browse menggunakan native folder picker (Wails dialog).
- Google Drive browse menampilkan folder picker dari Google Drive API.
- Tombol [Compare] untuk memulai proses.
- Disable tombol Compare jika salah satu directory belum dipilih.

---

### TASK-017: Create Compare page — results table

**Description:** Buat tabel hasil comparison dengan tabs.

**Acceptance Criteria:**
- 3 tab: Local Only, Drive Only, Both.
- Setiap tab menampilkan jumlah file (e.g. "Local Only (12)").
- Kolom tabel: No, Filename, Extension, Size (human-readable), Relative Path.
- Data di-load dari backend via Wails bindings.
- Loading state saat proses scan & compare berjalan.

---

### TASK-018: Implement table sorting

**Description:** Tambahkan sorting ke tabel hasil comparison.

**Acceptance Criteria:**
- Semua kolom bisa di-sort ascending/descending.
- Klik header kolom untuk toggle sort direction.
- Indikator sort direction di header (arrow up/down).
- Default sort: Filename ascending.

---

### TASK-019: Implement table filters

**Description:** Tambahkan search dan filter ke tabel.

**Acceptance Criteria:**
- Search by filename: text input, partial match, case-insensitive.
- Filter by extension: dropdown multi-select, daftar extension dari hasil scan.
- Filter by size range: dropdown single-select (< 1 MB, 1-10 MB, 10-100 MB, 100 MB-1 GB, > 1 GB).
- Filter berlaku pada tab yang aktif.
- Jumlah file di tab header ter-update sesuai filter.

---

### TASK-020: Implement infinite scroll

**Description:** Tambahkan infinite scroll ke tabel.

**Acceptance Criteria:**
- Load data per batch (50 items per load).
- Load batch berikutnya saat user scroll mendekati bawah tabel.
- Loading indicator saat memuat batch baru.
- Berhenti load jika semua data sudah ditampilkan.

---

### TASK-021: Create Settings page — Google Drive connection

**Description:** Buat bagian Settings untuk mengelola koneksi Google Drive.

**Acceptance Criteria:**
- Menampilkan status koneksi (Connected / Disconnected).
- Tombol Connect untuk memulai OAuth2 flow.
- Tombol Disconnect untuk menghapus koneksi.
- Status ter-update secara real-time setelah connect/disconnect.

---

### TASK-022: Create Settings page — ignore patterns

**Description:** Buat bagian Settings untuk mengelola ignore patterns.

**Acceptance Criteria:**
- Menampilkan daftar ignore patterns yang tersimpan.
- Tombol [+ Add Pattern] untuk menambah pattern baru.
- Tombol delete untuk menghapus pattern.
- Perubahan tersimpan ke database.
- Validasi: pattern tidak boleh kosong, tidak boleh duplikat.

---

## Milestone 3 — Polish

### TASK-023: Error handling and edge cases

**Description:** Tambahkan penanganan error dan edge cases.

**Acceptance Criteria:**
- Error message yang user-friendly di UI (bukan stack trace).
- Handle: folder tidak ditemukan, permission denied, Google Drive disconnected.
- Handle: scan folder kosong (tampilkan pesan "No files found").
- Handle: koneksi database gagal saat startup.
- Handle: Google Drive API timeout / network error.

---

### TASK-024: Performance optimization

**Description:** Optimasi performa scan dan comparison.

**Acceptance Criteria:**
- Local scan menggunakan concurrent workers (goroutine pool).
- Google Drive scan menggunakan batched API calls.
- Comparison engine efisien untuk dataset besar (200,000+ files).
- Progress indicator di UI saat scan berjalan.
- Scan bisa di-cancel oleh user.

---

### TASK-025: Comprehensive testing

**Description:** Tambahkan test coverage yang komprehensif.

**Acceptance Criteria:**
- Unit test untuk semua domain entities dan use cases.
- Unit test untuk semua repository operations.
- Unit test untuk scanner (local dan Google Drive mock).
- Unit test untuk comparison engine.
- Integration test untuk full compare flow.
- Test coverage minimal 80%.
