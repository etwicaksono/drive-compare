# Product Requirements Document (PRD)

## Personal Digital Inventory

**Version:** 1.0
**Status:** Draft
**Last Updated:** 2026-07-04

---

## 1. Overview

Personal Digital Inventory adalah aplikasi desktop yang menginventarisasi file di berbagai lokasi penyimpanan dan membantu pengguna memahami di mana setiap file berada.

Aplikasi ini **bukan** aplikasi sinkronisasi. Aplikasi ini **bukan** aplikasi backup otomatis. Aplikasi ini **tidak pernah** menghapus file secara otomatis. Aplikasi ini **tidak pernah** mengupload file secara otomatis. Aplikasi ini hanya melakukan scan, membandingkan, dan melaporkan. Pengguna selalu melakukan upload atau penghapusan secara manual.

---

## 2. Problem Statement

Pengguna memiliki file yang tersebar di beberapa lokasi penyimpanan:

- Laptop (local storage)
- External HDD
- Google Drive
- Google Photos

Saat ini tidak ada cara mudah untuk mengetahui:

- File mana yang hanya ada di satu lokasi dan belum memiliki backup.
- File mana yang sudah dihapus dari local storage tetapi masih ada di cloud.
- File mana yang hanya ada di cloud tetapi tidak ada salinan lokal.
- Status keseluruhan kesehatan backup.

Aplikasi yang ada (Google Drive Desktop, rclone, FreeFileSync, dll.) fokus pada sinkronisasi folder, bukan pada audit inventaris file lintas lokasi.

---

## 3. Target User

Pengguna tunggal (personal use). Aplikasi ini ditujukan untuk penggunaan pribadi di satu komputer desktop/laptop Windows.

---

## 4. Goals

Aplikasi harus mampu menjawab pertanyaan-pertanyaan berikut:

- File mana saja yang hanya ada di local storage (laptop / external HDD)?
- File mana saja yang hanya ada di Google Drive?
- File mana saja yang ada di kedua lokasi?

---

## 5. Scope

### 5.1 In Scope (Version 1)

#### Core — Inventory Engine

- Scan file dari local folders (laptop).
- Scan file dari external HDD.
- Scan file dari Google Drive (via Google Drive API).
- Menyimpan metadata file ke database (PostgreSQL).
- Membandingkan inventaris antar lokasi.
- Menampilkan laporan perbedaan (file yang hanya ada di satu lokasi, file yang missing backup, dll.).

#### File Identity

File diidentifikasi berdasarkan kombinasi:

- **Relative path** (relatif terhadap root folder yang di-scan)
- **Filename**
- **File size** (dalam bytes)

Hash (SHA256/BLAKE3) **tidak** diimplementasikan di versi 1. Akan ditambahkan di versi mendatang jika diperlukan.

Asumsi: pengguna tidak akan melakukan rename file jika ingin melakukan tracking history.

#### Compare View

- User memilih root directory local (laptop atau external HDD).
- User memilih root directory Google Drive.
- Sistem melakukan scanning kedua lokasi dan komparasi otomatis.
- Sistem menampilkan laporan dalam bentuk tabel dengan 3 kategori:
  - **Local Only**: file yang hanya ada di local storage.
  - **Drive Only**: file yang hanya ada di Google Drive.
  - **Both**: file yang ada di kedua lokasi.

#### Settings

- Konfigurasi koneksi Google Drive (OAuth2 connect/disconnect).
- Konfigurasi ignore patterns (node_modules, .git, Thumbs.db, dll.).

### 5.2 In Scope (Version 2 — Future)

- **Google Photos** sebagai provider tambahan.
- **Movie Library**: watched/watchlist, rating, review, integrasi TMDB untuk metadata film otomatis (poster, genre, cast, director).
- **Hash-based identity** (BLAKE3) sebagai validasi opsional atau on-demand.
- **Snapshot/versioning**: melihat kondisi inventaris pada tanggal tertentu.
- **Integrity check**: verifikasi hash secara berkala.

### 5.3 Out of Scope

- Sinkronisasi otomatis file antar lokasi.
- Upload otomatis ke Google Drive / Google Photos.
- Penghapusan otomatis file di lokasi manapun.
- Enkripsi file.
- Dukungan multi-user atau server mode.
- Mobile app (Android/iOS).
- Dukungan cloud storage selain Google (OneDrive, Dropbox, S3, dll.) di versi 1.

---

## 6. User Flows

### 6.1 First Launch

```
Install & Open App
       │
       ▼
 Settings Page
 (connect Google Drive via OAuth2)
       │
       ▼
  Compare Page
```

### 6.2 Compare Flow

```
Open App
   │
   ▼
Compare Page
   │
   ▼
Pilih Root Directory Local
(laptop atau external HDD via folder picker)
   │
   ▼
Pilih Root Directory Google Drive
(browse folder Google Drive)
   │
   ▼
Klik [Compare]
   │
   ▼
Scanning & Comparing...
(scan kedua lokasi, bandingkan berdasarkan relative path + name + size)
   │
   ▼
Laporan Tabel
┌──────────────────────────────────────────────────────────┐
│ [Local Only]          [Drive Only]          [Both]       │
├──────────────────────────────────────────────────────────┤
│ Filename       Size       Relative Path                  │
│ proposal.docx  2 MB       Documents/                     │
│ notes.txt      1 KB       Documents/Work/                │
│ ...                                                      │
└──────────────────────────────────────────────────────────┘
```

---

## 7. Architecture

### 7.1 Tech Stack

| Component       | Technology                          |
|-----------------|-------------------------------------|
| Backend/Core    | Go                                  |
| Desktop Shell   | Wails                               |
| Frontend        | React + TypeScript                  |
| UI Library      | Tailwind CSS + shadcn/ui            |
| Database        | PostgreSQL                          |
| ORM             | GORM                                |
| Migration       | goose                               |
| Configuration   | YAML                                |
| Logging         | log/slog (Go standard library)      |
| Google Drive    | Google Drive API                    |
| File Watching   | fsnotify (future)                   |
| Hashing         | BLAKE3 (future, on-demand)          |

### 7.2 Clean Architecture Layers

```
┌─────────────────────────────────────┐
│            UI (React/Wails)         │
├─────────────────────────────────────┤
│         Application (Use Cases)     │
├─────────────────────────────────────┤
│         Domain (Entities, Interfaces)│
├─────────────────────────────────────┤
│     Infrastructure (DB, API, FS)    │
└─────────────────────────────────────┘
```

- **UI**: React components, pages, routing. Tidak boleh ada business logic.
- **Application**: Use cases / services. Orchestrasi domain logic.
- **Domain**: Entity, value object, repository interface, provider interface. Tidak bergantung pada framework apapun.
- **Infrastructure**: Implementasi repository (GORM/PostgreSQL), implementasi provider (local disk, Google Drive API), konfigurasi.

### 7.3 Data Model

```go
type FileItem struct {
    RelativePath string
    Name         string
    Extension    string
    Size         int64
}
```

### 7.4 Database Schema (Awal)

**comparisons**

| Column            | Type      | Description                                      |
|-------------------|-----------|--------------------------------------------------|
| id                | uint      | Primary key                                      |
| local_path        | string    | Root directory local yang dipilih user            |
| drive_path        | string    | Root directory Google Drive yang dipilih user     |
| started_at        | time.Time | Waktu mulai compare                              |
| finished_at       | time.Time | Waktu selesai compare                            |
| total_local_only  | int       | Jumlah file yang hanya ada di local              |
| total_drive_only  | int       | Jumlah file yang hanya ada di drive              |
| total_both        | int       | Jumlah file yang ada di keduanya                 |

**comparison_files**

| Column         | Type      | Description                                         |
|----------------|-----------|-----------------------------------------------------|
| id             | uint      | Primary key                                         |
| comparison_id  | uint      | Foreign key ke comparisons                          |
| filename       | string    | Nama file tanpa extension                           |
| extension      | string    | Extension file (.jpg, .png, .mp4, dll.)             |
| relative_path  | string    | Path relatif terhadap root directory                |
| size           | int64     | Ukuran file dalam bytes                             |
| status         | string    | "local_only", "drive_only", "both"                  |

**ignore_patterns**

| Column  | Type   | Description                                          |
|---------|--------|------------------------------------------------------|
| id      | uint   | Primary key                                          |
| pattern | string | Glob pattern yang di-ignore saat scan (e.g. .git)    |


---

## 8. UI Design

### 8.1 Layout

Sidebar kiri + konten di kanan. Mirip GitHub Desktop, Obsidian, Linear.

**Sidebar:**
- Compare
- Settings

### 8.2 Compare

```
──────────────────────────────────────────────────────────────────────────
 Compare
──────────────────────────────────────────────────────────────────────────

 Local Directory
 [D:\Photos                              ] [Browse]

 Google Drive Directory
 [Backup/Photos                          ] [Browse]

 [Compare]
──────────────────────────────────────────────────────────────────────────

 Results:  Local Only (12)  |  Drive Only (8)  |  Both (1,204)

 [Search filename...]

 Filters:
 Extension  [All ▾]  (.jpg, .png, .mp4, .pdf, .docx, ...)
 Size Range [All ▾]  (< 1 MB, 1-10 MB, 10-100 MB, 100 MB-1 GB, > 1 GB)

──────────────────────────────────────────────────────────────────────────
 Tab: [Local Only]
──────────────────────────────────────────────────────────────────────────
 No ▴  Filename ▴▾       Extension ▴▾  Size ▴▾     Relative Path ▴▾
──────────────────────────────────────────────────────────────────────────
 1     screenshot        .png          800 KB      Screenshots/
 2     proposal          .docx         2 MB        Documents/Work/
 3     notes             .txt          1 KB        Documents/
 ...   (infinite scroll)
──────────────────────────────────────────────────────────────────────────

 Tab: [Drive Only]
──────────────────────────────────────────────────────────────────────────
 No ▴  Filename ▴▾       Extension ▴▾  Size ▴▾     Relative Path ▴▾
──────────────────────────────────────────────────────────────────────────
 1     Vacation          .jpg          5 MB        Holiday/2024/
 2     old_backup        .zip          120 MB      Archive/
 ...   (infinite scroll)
──────────────────────────────────────────────────────────────────────────

 Tab: [Both]
──────────────────────────────────────────────────────────────────────────
 No ▴  Filename ▴▾       Extension ▴▾  Size ▴▾     Relative Path ▴▾
──────────────────────────────────────────────────────────────────────────
 1     IMG001            .jpg          4 MB        DCIM/Camera/
 2     IMG002            .jpg          3.8 MB      DCIM/Camera/
 3     Movie             .mp4          1.2 GB      Videos/
 ...   (infinite scroll)
──────────────────────────────────────────────────────────────────────────

 Kolom Tabel:
 - No          : Nomor urut baris
 - Filename    : Nama file tanpa extension
 - Extension   : Extension file (.jpg, .png, .mp4, dll.)
 - Size        : Ukuran file (human-readable: KB, MB, GB)
 - Relative Path: Path relatif terhadap root directory yang dipilih

 Sorting:
 - Semua kolom bisa di-sort ascending/descending
 - Default sort: Filename ascending

 Filter:
 - Search      : Text search pada filename (partial match, case-insensitive)
 - Extension   : Dropdown multi-select, daftar extension diambil dari hasil scan
 - Size Range  : Dropdown single-select (< 1 MB, 1-10 MB, 10-100 MB,
                 100 MB-1 GB, > 1 GB)

 Scrolling:
 - Infinite scroll (load batch berikutnya saat user scroll ke bawah)
```

### 8.3 Settings

```
──────────────────────────────────────────
 Google Drive
──────────────────────────────────────────
 Status: Connected
 [Disconnect]
──────────────────────────────────────────
 Ignore Patterns
──────────────────────────────────────────
 node_modules
 .git
 Thumbs.db
 desktop.ini
 [+ Add Pattern]
──────────────────────────────────────────
```

---

## 9. Configuration

```yaml
database:
  host: localhost
  port: 5432
  user: inventory
  password: ""  # use environment variable PDI_DB_PASSWORD
  name: personal_digital_inventory
  sslmode: disable

ignore:
  - node_modules
  - .git
  - Thumbs.db
  - desktop.ini
  - "*.tmp"
  - "*.cache"

logging:
  level: info  # debug | info | warn | error
```

---

## 10. Non-Functional Requirements

| Requirement      | Target                                              |
|------------------|-----------------------------------------------------|
| Platform         | Windows 10/11                                       |
| Performance      | Scan 200,000 file metadata dalam < 60 detik (local) |
| Database Size    | Tidak terbatas (PostgreSQL)                         |
| Startup Time     | < 3 detik                                           |
| Binary Size      | Single executable via Wails (requires PostgreSQL running) |
| Offline          | Berfungsi penuh tanpa internet (kecuali Google API) |
| Database         | PostgreSQL 15+ running locally                      |
| Data Safety      | Tidak pernah menghapus/mengubah file pengguna        |

---

## 11. Milestones

### Milestone 0 — Design

- PRD
- User Flow
- Wireframe
- Mockup

### Milestone 1 — Project Foundation

- Project setup (Go + Wails)
- Folder structure (Clean Architecture)
- Database setup (GORM + PostgreSQL)
- Configuration loader (YAML)
- Logging (slog)
- Migration mechanism (goose)

### Milestone 2 — Compare Feature

- Local directory scanner (recursive, ignore rules)
- Google Drive API authentication (OAuth2)
- Google Drive scanner
- Comparison engine
- Compare UI (folder picker, tabel hasil, filter, sorting, infinite scroll)
- Settings UI (Google Drive connection, ignore patterns)

### Milestone 3 — Polish

- Error handling & edge cases
- Performance optimization
- Testing & QA

### Milestone 4 — Future (v2)

- Google Photos provider
- Movie Library (watched, watchlist, rating, review, TMDB integration)
- Hash-based identity (BLAKE3)
- Snapshot / versioning
- Integrity check

---

## 12. Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Google Drive API rate limits | Implement exponential backoff dan caching |
| File identity collision (nama + size sama tapi isi berbeda) | Kemungkinan sangat kecil untuk personal use; tambahkan hash on-demand di v2 |
| Scan pertama kali lambat untuk ratusan ribu file | Gunakan worker pool dengan concurrent scanning |
| PostgreSQL harus running sebagai dependency | Dokumentasikan setup PostgreSQL; gunakan Docker Compose untuk kemudahan |

---

## 13. Glossary

| Term | Definition |
|------|-----------|
| Local Storage | Folder di laptop atau external HDD yang dipilih user untuk di-compare |
| Google Drive | Cloud storage Google yang diakses via Google Drive API |
| Compare | Proses membandingkan file antara local storage dan Google Drive |
| File Identity | Kombinasi relative path + filename + size yang mengidentifikasi sebuah file |
| Ignore Pattern | Glob pattern untuk mengecualikan file/folder tertentu dari scan |
