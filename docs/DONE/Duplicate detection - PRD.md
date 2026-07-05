# PRD: Duplicate Detection

**Status:** Draft
**Last Updated:** 2026-07-06

---

## 1. Overview

Fitur ini mendeteksi file duplikat di dalam satu lokasi (local storage atau Google Drive) berdasarkan pattern penamaan yang umum terjadi saat copy/download. User memilih pattern duplikasi yang ingin dideteksi, lalu sistem mengelompokkan file-file yang kemungkinan merupakan duplikat dari file yang sama.

---

## 2. Problem Statement

Saat menyalin file, OS atau browser sering menambahkan suffix otomatis:

```
photo.jpg           ← original
photo (1).jpg       ← Windows copy
photo (2).jpg       ← Windows copy ke-2
photo copy.jpg      ← macOS copy
photo copy 2.jpg    ← macOS copy ke-2
photo_1.jpg         ← download ulang
photo-1.jpg         ← download ulang
photo (1) (1).jpg   ← copy dari copy
```

User ingin menemukan file-file seperti ini agar bisa dibersihkan secara manual.

---

## 3. User Flow

```
Halaman baru: "Duplicates" (menu sidebar)
        │
        ▼
User memilih scan source:
  ○ Local Directory  [Browse]
  ○ Google Drive     [Browse]
        │
        ▼
User memilih duplication patterns (multi-select checkbox):
  ☑ name (1).ext, name (2).ext, ...       — Windows copy
  ☑ name copy.ext, name copy 2.ext, ...   — macOS copy
  ☑ name_1.ext, name_2.ext, ...           — Underscore numbering
  ☑ name-1.ext, name-2.ext, ...           — Dash numbering
  ☐ Custom pattern...                      — User-defined regex
        │
        ▼
User klik "Scan for Duplicates"
        │
        ▼
Hasil ditampilkan sebagai tabel grouped:
  ┌─────────────────────────────────────────────────────────┐
  │  Found 24 duplicate groups (68 files)                   │
  ├─────────────────────────────────────────────────────────┤
  │  ▼ photo.jpg (3 files)                                  │
  │    1  photo.jpg         2.1 MB  /vacation/              │
  │    2  photo (1).jpg     2.1 MB  /vacation/              │
  │    3  photo (2).jpg     2.1 MB  /vacation/              │
  ├─────────────────────────────────────────────────────────┤
  │  ▼ report.pdf (2 files)                                 │
  │    1  report.pdf        540 KB  /documents/             │
  │    2  report copy.pdf   540 KB  /documents/             │
  ├─────────────────────────────────────────────────────────┤
  │  ▶ budget.xlsx (2 files)  [collapsed]                   │
  └─────────────────────────────────────────────────────────┘
        │
        ▼
User bisa select items → Export (sama seperti fitur Export di Compare)
```

---

## 4. Duplication Patterns

### 4.1 Built-in Patterns

Setiap pattern memiliki regex untuk mencocokkan filename dan mengekstrak "base name" (nama file asli).

| ID | Label | Regex Pattern | Contoh Match | Base Name |
|----|-------|---------------|-------------|-----------|
| `windows_copy` | Windows copy `(N)` | `^(.+)\s\((\d+)\)$` | `photo (1)` → `photo` | `photo` |
| `macos_copy` | macOS copy | `^(.+)\scopy(?:\s(\d+))?$` | `photo copy` → `photo`, `photo copy 2` → `photo` | `photo` |
| `underscore_num` | Underscore `_N` | `^(.+)_(\d+)$` | `photo_1` → `photo` | `photo` |
| `dash_num` | Dash `-N` | `^(.+)-(\d+)$` | `photo-1` → `photo` | `photo` |

**Catatan:** Regex di-apply pada **name tanpa extension**. Extension harus sama persis agar dianggap duplikat.

### 4.2 Logika Deteksi

```
Untuk setiap file:
  1. Cek apakah name cocok dengan salah satu pattern yang dipilih
  2. Jika cocok → ekstrak base_name
  3. Jika tidak cocok → base_name = name (file ini mungkin "original")
  4. Group key = base_name + extension + relativePath
  5. Kumpulkan semua file dengan group key yang sama
  6. Hanya tampilkan group yang memiliki > 1 file
```

### 4.3 Contoh Grouping

Files:
```
/photos/photo.jpg
/photos/photo (1).jpg
/photos/photo (2).jpg
/photos/photo_1.jpg        ← jika pattern underscore_num dipilih
/documents/report.pdf
/documents/report copy.pdf
```

Dengan pattern `windows_copy` + `macos_copy` + `underscore_num`:

| Group Key | Files |
|-----------|-------|
| `photo.jpg\|/photos/` | `photo.jpg`, `photo (1).jpg`, `photo (2).jpg`, `photo_1.jpg` |
| `report.pdf\|/documents/` | `report.pdf`, `report copy.pdf` |

---

## 5. Detail Behavior

### 5.1 Scan Source

- Sama seperti Compare: local directory browse (native dialog) atau Google Drive folder picker.
- Hanya scan **satu** source per operasi (bukan cross-location).
- Menggunakan `LocalScanner` atau `DriveScanner` yang sudah ada.
- Ignore patterns tetap berlaku.

### 5.2 Results Table

- **Grouped view**: file dikelompokkan per duplicate group.
- Group header: menampilkan base filename + jumlah file dalam group.
- Group bisa di-collapse/expand (default: expanded).
- Dalam group: tabel dengan kolom No, Filename, Extension, Size, Relative Path.
- Sorting dan filtering berlaku di dalam group.
- Search bisa mencari across groups.
- **Summary bar** di atas: "Found X duplicate groups (Y files total)".

### 5.3 Selection & Export

- Reuse mekanisme selection + export dari fitur Export Result (PRD terpisah).
- User bisa select individual files dari group mana saja.
- Path source otomatis mengikuti scan source (local → local path, drive → drive path).
- Export ke clipboard atau .txt.

---

## 6. Scope

### 6.1 Backend

| File | Deskripsi |
|------|-----------|
| `internal/domain/duplicate.go` | `DuplicatePattern` struct, `DetectDuplicates()` function, grouping logic |
| `internal/domain/duplicate_test.go` | Unit test: pattern matching, grouping, edge cases |
| `app.go` | `RunDuplicateScan(source, sourceID, patterns[])` Wails binding |

### 6.2 Frontend

| File | Deskripsi |
|------|-----------|
| `pages/DuplicatesPage.tsx` | New page — source picker, pattern selector, scan button |
| `components/DuplicateResultsTable.tsx` | New — grouped table, collapse/expand, summary |
| `components/Sidebar.tsx` | Tambah link "Duplicates" |
| `App.tsx` | Tambah route `/duplicates` |
| `lib/types.ts` | Tambah `DuplicateGroup`, `DuplicateScanResult` types |

---

## 7. Edge Cases

- **File tanpa extension** — base_name matching tetap berlaku, extension = empty string.
- **Nested copy**: `photo (1) (1).jpg` — pattern `windows_copy` match outer `(1)`, base = `photo (1)`, bisa chain ke `photo` jika multi-pass. **V1: single-pass saja**, tidak recursive resolve.
- **Case sensitivity** — Windows: case-insensitive match. Google Drive: case-sensitive. Ikuti source.
- **Cross-folder duplicates** — V1 hanya deteksi duplikat **dalam folder yang sama** (relativePath sama). Cross-folder detection bisa jadi enhancement.
- **Tidak ada duplikat ditemukan** — tampilkan message "No duplicates found with the selected patterns."
- **Pattern overlap** — file bisa match multiple patterns. Tetap masuk satu group saja (group by base_name).
- **File original tidak ada** — misal hanya ada `photo (1).jpg` tanpa `photo.jpg`. Tetap tampilkan sebagai group jika ada file lain yang match ke base_name yang sama. Jika hanya 1 file, tidak ditampilkan (bukan group).

---

## 8. Task Breakdown

| No | Task | Deskripsi |
|----|------|-----------|
| 1 | Domain: duplicate detection engine | Pattern registry, base name extraction, grouping logic, unit tests |
| 2 | Backend: Wails binding | `RunDuplicateScan` — scan + detect + return grouped results |
| 3 | Frontend: DuplicatesPage | Source picker, pattern selector (checkboxes), scan button |
| 4 | Frontend: DuplicateResultsTable | Grouped table, collapse/expand, summary bar, sorting/filtering |
| 5 | Frontend: Sidebar + routing | Tambah menu Duplicates, route /duplicates |
| 6 | Selection & Export integration | Reuse export mechanism dari Export Result PRD |
| 7 | Testing | Domain unit tests, pattern matching edge cases |
