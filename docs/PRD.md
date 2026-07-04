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

- File mana saja yang hanya ada di laptop?
- File mana saja yang hanya ada di Google Drive?
- File mana saja yang hanya ada di Google Photos?
- File mana saja yang ada di beberapa lokasi sekaligus?
- File mana saja yang belum memiliki backup?
- Film apa saja yang sudah ditonton?
- Film apa saja yang ada di watchlist?
- Berapa rating yang diberikan untuk sebuah film?

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

#### Dashboard

- Ringkasan kesehatan backup (fully backed up, local only, drive only, missing backup).
- Jumlah file per kategori.
- Status scan terakhir per provider.
- Ringkasan penggunaan storage.

#### Inventory View

- Daftar seluruh file yang diinventarisasi.
- Informasi lokasi setiap file (ada di Laptop / HDD / Drive / Photos).
- Status backup per file.
- Search dan filter.

#### Compare View

- Menampilkan file yang hanya ada di satu lokasi tertentu.
- Kategori: "Only in Laptop", "Only in Drive", "Only in HDD", "Exists Everywhere".
- Navigasi ke file (Open in Explorer / Open in Google Drive).

#### Settings

- Konfigurasi provider (enable/disable).
- Konfigurasi folder yang di-scan per provider.
- Konfigurasi ignore patterns (node_modules, .git, Thumbs.db, dll.).
- Konfigurasi jadwal scan (manual, setiap startup, harian).

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
 (configure providers & folders)
       │
       ▼
  First Scan
       │
       ▼
   Dashboard
```

### 6.2 Regular Use

```
Open App
   │
   ▼
Dashboard (lihat health overview)
   │
   ├── Scan Now (manual trigger)
   │       │
   │       ▼
   │   Scanning...
   │       │
   │       ▼
   │   Dashboard Updated
   │
   ├── Inventory (browse semua file)
   │       │
   │       ▼
   │   Search / Filter
   │       │
   │       ▼
   │   File Detail (lokasi, status)
   │
   ├── Compare (lihat perbedaan antar lokasi)
   │       │
   │       ▼
   │   "Only in Laptop" / "Only in Drive" / etc.
   │       │
   │       ▼
   │   Open in Explorer / Open in Drive
   │
   └── Settings (ubah konfigurasi)
```

### 6.3 External HDD Plugged In

```
Colok External HDD
       │
       ▼
  Open App
       │
       ▼
 Dashboard: "External HDD — Last scan: 30 days ago"
       │
       ▼
  Scan External HDD
       │
       ▼
 "17 file baru, 3 file berubah"
       │
       ▼
 Dashboard Updated
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

### 7.3 Provider Interface

```go
type Provider interface {
    Name() string
    Scan(ctx context.Context) ([]InventoryItem, error)
}
```

Setiap provider (Local Disk, External HDD, Google Drive) mengimplementasikan interface yang sama. Engine tidak peduli sumber datanya dari mana.

### 7.4 Data Model

```go
type InventoryItem struct {
    Provider     string    // "local", "drive", "photos"
    Device       string    // "laptop", "external-hdd"
    RelativePath string
    Name         string
    Size         int64
    ModifiedAt   time.Time
    RemoteID     string    // kosong untuk local
}
```

### 7.5 Database Schema (Awal)

**providers**

| Column | Type   |
|--------|--------|
| id     | uint   |
| name   | string |
| type   | string |

**devices**

| Column      | Type   |
|-------------|--------|
| id          | uint   |
| name        | string |
| provider_id | uint   |

**files**

| Column        | Type      |
|---------------|-----------|
| id            | uint      |
| filename      | string    |
| relative_path | string    |
| size          | int64     |
| modified_at   | time.Time |
| source        | string    |
| device        | string    |
| remote_id     | string    |
| scan_id       | uint      |
| exists        | bool      |

**scans**

| Column     | Type      |
|------------|-----------|
| id         | uint      |
| provider   | string    |
| device     | string    |
| started_at | time.Time |
| finished_at| time.Time |
| total_files| int       |
| status     | string    |

**ignore_patterns**

| Column  | Type   |
|---------|--------|
| id      | uint   |
| pattern | string |

### 7.6 Google Drive Structure (Target)

```
Google Drive
├── Backup/
│   ├── Laptop/
│   │   ├── Documents/
│   │   ├── Photos/
│   │   └── ...
│   ├── Android/
│   │   ├── DCIM/
│   │   ├── WhatsApp/
│   │   └── ...
│   └── External/
│       ├── Movie/
│       ├── Archive/
│       └── ...
└── .metadata/
    └── config.json
```

---

## 8. UI Design

### 8.1 Layout

Sidebar kiri + konten di kanan. Mirip GitHub Desktop, Obsidian, Linear.

**Sidebar:**
- Dashboard
- Inventory
- Compare
- Reports
- Settings

### 8.2 Dashboard

```
──────────────────────────────────────────
 Personal Digital Inventory
──────────────────────────────────────────
 Health
 ● Fully Backed Up      18,230
 ● Local Only              12
 ● Drive Only              25
 ● Missing Backup           4
──────────────────────────────────────────
 Recent Scans
 Laptop          2 hours ago      ✓
 External HDD    30 days ago      ⚠
 Google Drive     1 day ago       ✓
──────────────────────────────────────────
 Storage Summary
 Laptop           120 GB
 External HDD     500 GB
 Google Drive      80 GB
──────────────────────────────────────────
```

### 8.3 Inventory

```
──────────────────────────────────────────────────────
 [Search...]     [Filter: All ▾]
──────────────────────────────────────────────────────
 Filename         Size    Laptop  HDD  Drive  Status
──────────────────────────────────────────────────────
 IMG001.jpg       4 MB     ✓      ✓     ✓    Complete
 proposal.docx    2 MB     ✓      ✗     ✗    Local Only
 Movie.mp4        1.2 GB   ✗      ✓     ✓    No Local
 Vacation.jpg     5 MB     ✗      ✗     ✓    Drive Only
──────────────────────────────────────────────────────
```

### 8.4 Compare

```
──────────────────────────────────────────
 [Only in Laptop]  [Only in Drive]  [Only in HDD]  [Everywhere]
──────────────────────────────────────────

 Only in Laptop (12 files)
──────────────────────────────────────────
 proposal.docx           2 MB
 screenshot.png        800 KB
 notes.txt               1 KB
──────────────────────────────────────────
 [Open in Explorer]
```

### 8.5 Settings

```
──────────────────────────────────────────
 Providers
──────────────────────────────────────────
 ☑ Laptop
   Folders:
   - D:\Documents
   - D:\Photos
   - D:\Android Backup

 ☑ External HDD
   Folders:
   - E:\Movie
   - E:\Archive

 ☑ Google Drive
   Status: Connected
   [Disconnect]

 ☐ Google Photos (coming soon)
──────────────────────────────────────────
 Ignore Patterns
──────────────────────────────────────────
 node_modules
 .git
 Thumbs.db
 desktop.ini
 [+ Add Pattern]
──────────────────────────────────────────
 Scan Schedule
──────────────────────────────────────────
 ○ Manual Only
 ● Every Startup
 ○ Daily
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

sources:
  - name: laptop
    type: local
    paths:
      - D:\Documents
      - D:\Photos
      - D:\Android Backup

  - name: external-hdd
    type: local
    paths:
      - E:\Movie
      - E:\Archive

  - name: google-drive
    type: drive

ignore:
  - node_modules
  - .git
  - Thumbs.db
  - desktop.ini
  - "*.tmp"
  - "*.cache"

scan:
  schedule: startup  # manual | startup | daily

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
- File entity + repository
- Provider interface
- Local provider (scan, recursive, ignore rules)
- Persist scan results

### Milestone 2 — Comparison & UI

- Comparison engine
- Dashboard UI
- Inventory view UI
- Compare view UI
- Search & filter
- Settings UI

### Milestone 3 — Google Drive

- Google Drive API authentication (OAuth2)
- Google Drive provider implementation
- Google Drive scan & persist

### Milestone 4 — Reports & Polish

- Reports (summary, export)
- Error handling & edge cases
- Performance optimization
- Testing & QA

### Milestone 5 — Future (v2)

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
| Google Photos API limitations (tidak bisa download file yang bukan diupload sendiri) | Gunakan Google Photos hanya untuk inventaris, bukan restore |
| External HDD tidak selalu terhubung | Simpan hasil scan terakhir, tampilkan status "offline" |
| File identity collision (nama + size sama tapi isi berbeda) | Kemungkinan sangat kecil untuk personal use; tambahkan hash on-demand di v2 |
| Scan pertama kali lambat untuk ratusan ribu file | Gunakan worker pool dengan concurrent scanning |
| PostgreSQL harus running sebagai dependency | Dokumentasikan setup PostgreSQL; gunakan Docker Compose untuk kemudahan |

---

## 13. Glossary

| Term | Definition |
|------|-----------|
| Provider | Sumber data yang bisa di-scan (Local Disk, Google Drive, Google Photos) |
| Device | Perangkat fisik dalam sebuah provider (Laptop, External HDD) |
| Inventory | Kumpulan seluruh metadata file dari semua provider |
| Scan | Proses membaca metadata file dari sebuah provider |
| Compare | Proses membandingkan inventaris antar provider |
| File Identity | Kombinasi relative path + filename + size yang mengidentifikasi sebuah file |
