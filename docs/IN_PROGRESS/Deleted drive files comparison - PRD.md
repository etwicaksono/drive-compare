# PRD: Deleted Drive Files Comparison

**Status:** Implemented (TRASH-001…011) — menunggu verifikasi manual TRASH-012
**Last Updated:** 2026-07-30

---

## 1. Overview

Fitur ini menambahkan **filter cakupan file (file scope)** pada setiap Google Drive source, sehingga user bisa membandingkan atau meng-audit file yang **sudah dihapus (masuk Trash) di Google Drive**, bukan hanya file aktif. Filter tersedia di halaman **Resize Audit** dan **Compare** karena keduanya memakai satu choke point scan yang sama (`App.scanSource`).

Tiga mode scope:

| Mode | Arti | Drive query predicate |
|------|------|-----------------------|
| `active` (default) | Hanya file aktif (perilaku saat ini) | `trashed = false` |
| `trashed` | Hanya file yang sudah dihapus / ada di Trash | `trashed = true` |
| `all` | File aktif + file terhapus | tanpa predicate `trashed` |

---

## 2. Problem Statement

Saat ini `DriveScanner.scanFolder` mengunci query ke `trashed = false`:

```go
query := fmt.Sprintf("'%s' in parents and trashed = false", folderID)
```

Akibatnya file yang sudah dihapus di Drive **tidak pernah terlihat** oleh aplikasi. Ini menimbulkan hasil audit yang menyesatkan:

```
Kasus A — Resize Audit
  Source  : /Videos (local)        → video.mp4
  Resized : /Videos-Resized (Drive) → video_resized.mp4 SUDAH DIHAPUS
  Hasil sekarang : "Pending" (padahal resize sudah pernah dibuat, cuma kehapus)
  Yang diinginkan: user bisa cek Trash dan tahu file itu pernah ada

Kasus B — Compare
  Source A: folder lokal
  Source B: folder Drive
  File yang sengaja/tidak sengaja dihapus di Drive tampil sebagai "A only",
  tanpa cara membedakan "belum pernah diupload" vs "sudah dihapus".
```

User butuh cara eksplisit untuk membandingkan terhadap isi Trash Google Drive, dan menandai item mana yang berasal dari Trash.

---

## 3. User Flow

```
Halaman: Resize Audit (dan Compare)
        │
        ▼
Per source row, dropdown tipe source diperluas:
  [ Local ▾ ]
  [ Drive ▾ ]  →  saat "Drive" dipilih, muncul dropdown scope:
                    [ Active files ▾ ]   ← default
                    [ Deleted (Trash) ]
                    [ Active + Deleted ]
        │
        ▼
┌──────────────────────────────────────────────────────────────────────┐
│ Source (Original Files)                                              │
│ [ Local ▾ ] [ D:\Videos                        ] [ Browse ]          │
│                                                                      │
│ Resized (Output Files)                                               │
│ [ Drive ▾ ] [ Active + Deleted ▾ ] [ Videos-Resized ] [ Browse ]     │
└──────────────────────────────────────────────────────────────────────┘
        │
        ▼
User klik "Run Audit"
        │
        ▼
Hasil di ResultsTable, item dari Trash diberi badge "Deleted":
  ┌───────────────────────────────────────────────────────────────────┐
  │  Pending (3)   Resized (128)   Extra Resized (2)                  │
  ├───────────────────────────────────────────────────────────────────┤
  │  No  Filename              Ext   Size    Path        Status       │
  │  1   holiday_resized       .mp4  120 MB  /2024/      🗑 Deleted   │
  │  2   party_resized         .mp4  98 MB   /2024/      Active       │
  └───────────────────────────────────────────────────────────────────┘
        │
        ▼
Selection + Export tetap seperti sekarang (path Drive/local, clipboard/file)
```

---

## 4. Scope Filter & Logika

### 4.1 Mapping mode → Drive query

`DriveScanner` menerima `TrashScope` dan menyusun predicate:

| `TrashScope` | Predicate yang di-append | Catatan |
|--------------|--------------------------|---------|
| `ScopeActive` | `and trashed = false` | perilaku existing, default |
| `ScopeTrashed` | `and trashed = true` | hanya item Trash |
| `ScopeAll` | *(tidak ada predicate)* | Drive API mengembalikan keduanya |

### 4.2 Field tambahan yang harus diminta

`realDriveAPI.ListFiles` saat ini **mengabaikan** parameter `fields` dan hardcode:

```go
Fields("nextPageToken", "files(id, name, size, mimeType)")
```

Harus diperbaiki agar memakai `fields` dari caller, dengan default baru menyertakan `trashed`:

```
files(id, name, size, mimeType, trashed)
```

Tanpa ini, mode `all` tidak bisa membedakan item aktif vs terhapus.

### 4.3 Penandaan hasil

- `domain.FileItem` mendapat field baru `Deleted bool`.
- Local scanner selalu mengisi `Deleted = false`.
- Drive scanner mengisi `Deleted = f.Trashed`.
- Field ini **tidak** ikut jadi bagian key matching (`resizeKey` / compare key) supaya file terhapus tetap bisa dipasangkan dengan pasangan aktifnya.

### 4.4 Contoh perilaku Resize Audit

Suffix `_resized`, source lokal berisi `video.mp4`, folder Drive resized berisi `video_resized.mp4` yang sudah di-Trash:

| Scope resized source | Hasil |
|----------------------|-------|
| `active` | `video.mp4` → **Pending** |
| `trashed` | `video.mp4` → **Resized** (match ke file Trash, badge Deleted) |
| `all` | `video.mp4` → **Resized** (match, badge Deleted) |

---

## 5. Detail Behavior

### 5.1 Traversal folder untuk item Trash

Google Drive menaruh item yang dihapus di Trash tetapi **relasi `parents` umumnya tetap ada** untuk item yang dihapus eksplisit. Namun:

- Jika **folder induk** yang dihapus, anak-anaknya ikut ter-trash secara implisit (`explicitlyTrashed = false`), dan traversal per-parent bisa tidak konsisten.
- Karena itu pada mode `trashed`/`all`, scanner tetap melakukan rekursi per-parent yang sama seperti sekarang, dan folder ber-status trashed **juga** ditelusuri (tidak di-skip).
- Keterbatasan ini didokumentasikan di Edge Cases dan diverifikasi manual saat implementasi (lihat TRASH-002).

### 5.2 Default & backward compatibility

- Default scope = `active`, jadi perilaku existing tidak berubah bila user tidak menyentuh filter.
- Scope kosong (`""`) dari frontend diperlakukan sebagai `active` agar binding lama tetap aman.
- Scope tidak valid → error `invalid drive scope %q`.

### 5.3 Scope hanya berlaku untuk source Drive

- Untuk `sourceType = "local"`, dropdown scope disembunyikan dan nilainya diabaikan backend.
- `DriveFolderPicker` (`ListDriveFolders`) tetap hanya menampilkan folder aktif di V1 — user memilih folder aktif, lalu isinya yang di-scan sesuai scope. Browsing folder yang sudah terhapus adalah enhancement terpisah.

### 5.4 Tampilan status di tabel

- `ResultsTable` mendapat kolom "Status" opsional (`showStatusColumn`), aktif hanya jika ada minimal satu item `deleted = true` di hasil.
- Nilai: `Active` (netral) atau `Deleted` (warna merah/amber + ikon).
- Filter tambahan pada toolbar tabel: dropdown `All / Active only / Deleted only`.

### 5.5 Export

- Export tetap memakai `buildExportText` existing.
- Item terhapus diberi penanda ` [deleted]` di akhir baris agar hasil paste tetap informatif.

---

## 6. Scope (File yang Terlibat)

### 6.1 Backend

| File | Deskripsi |
|------|-----------|
| `internal/domain/entity.go` | Tambah field `Deleted bool` pada `FileItem` |
| `internal/infrastructure/scanner/drive_scanner.go` | `TrashScope` type, parameterisasi query `trashed`, `fields` dipakai beneran, isi `Deleted` dari `f.Trashed`, folder trashed ikut ditelusuri |
| `internal/infrastructure/scanner/drive_scanner_test.go` | Mock `ListFiles` mengenali query per-scope, fixture file trashed |
| `internal/infrastructure/scanner/local_scanner.go` | Pastikan `Deleted = false` (eksplisit) |
| `app.go` | `scanSource(sourceType, sourceID, scope)`, `CompareResultItem.Deleted`, `RunComparison`/`RunResizeAudit`/`RunDuplicateScan` menerima scope |
| `frontend/wailsjs/**` | Regenerate bindings |

### 6.2 Frontend

| File | Deskripsi |
|------|-----------|
| `lib/types.ts` | `FileItem.deleted`, type `DriveScope`, penanda `[deleted]` di `buildExportText` |
| `pages/ResizeAuditPage.tsx` | `SourceState.scope`, dropdown scope di `SourceRow`, kirim scope ke `RunResizeAudit` |
| `pages/ComparePage.tsx` | Perubahan setara untuk source A/B |
| `components/ResultsTable.tsx` | Kolom Status opsional + filter Active/Deleted |

---

## 7. Edge Cases

- **Folder induk yang dihapus** — anak-anaknya `trashed = true` tapi implisit; traversal per-`parents` bisa melewatkannya. Risiko diketahui; verifikasi manual di TRASH-002, dan jika terbukti tidak reliable, tambahkan catatan di UI ("hanya file yang dihapus langsung yang terdeteksi").
- **File dihapus permanen** — tidak ada di API sama sekali, tidak mungkin terdeteksi. Bukan bagian scope fitur.
- **Shared drive / file milik orang lain di Trash** — `trashed` bisa bernilai berbeda per-user; V1 hanya mengandalkan `trashed` seperti yang dikembalikan API untuk akun yang terautentikasi.
- **Nama duplikat aktif + terhapus** — Drive membolehkan `video.mp4` aktif dan `video.mp4` di Trash sekaligus. Pada mode `all`, keduanya masuk hasil sebagai dua entry dengan key yang sama; UI harus tetap render keduanya (key untuk selection perlu menyertakan status agar tidak bentrok).
- **Mode `trashed` tapi Trash kosong** — hasil kosong, tampilkan empty state existing.
- **Ignore patterns** — tetap diterapkan sama untuk item Trash.
- **Google Docs di Trash** — tetap di-skip seperti sekarang (`isGoogleDoc`).
- **Scope dipilih untuk source Local** — diabaikan, tidak error.
- **Selection lintas scope** — mengganti scope lalu re-run audit mereset selection (perilaku existing `handleAudit`).

---

## 8. Task Breakdown

| No | Task | Deskripsi |
|----|------|-----------|
| 1 | Domain: `Deleted` marker | Tambah field `Deleted` di `FileItem`, sesuaikan test yang ada |
| 2 | Scanner: `TrashScope` + fields fix | Parameterisasi predicate `trashed`, pakai `fields` dari caller, isi `Deleted` |
| 3 | Backend: scope di `scanSource` & binding | Validasi scope, teruskan ke Drive scanner, DTO `deleted`, regenerate bindings |
| 4 | Frontend: types + scope state | `DriveScope`, `FileItem.deleted`, penanda export |
| 5 | Frontend: dropdown scope | `SourceRow` di Resize Audit + Compare |
| 6 | Frontend: kolom & filter status | `ResultsTable` kolom Status + filter Active/Deleted |
| 7 | Testing | Unit test scanner per-scope, verifikasi manual traversal Trash, coverage ≥ 90% |

---

## 9. Status Implementasi & Verifikasi

### 9.1 Yang sudah selesai (otomatis terverifikasi)

| Task | Commit | Catatan |
|------|--------|---------|
| TRASH-001 | `9a52e33` | `FileItem.Deleted` |
| TRASH-002 | `6c3a5da` | `TrashScope` + predicate per scope |
| TRASH-003 | `dccfda4` | `fields` dipakai, `trashed` diminta |
| TRASH-004 | `1271f24` | `Deleted = false` eksplisit di local scanner |
| TRASH-005 | `ebad5ff` | scope diteruskan via `scanSource` |
| TRASH-006 | `e554fe1` | DTO `deleted` + binding + regenerate wailsjs |
| TRASH-007 | `41b333c` | `DriveScope`, key selection, penanda `[deleted]` |
| TRASH-008 | `2fba9ab` | dropdown scope di Resize Audit |
| TRASH-009 | `8774efe` | dropdown scope di Compare |
| TRASH-010 | `c0d7c53` | kolom + filter Status |
| TRASH-011 | `213cd85` | coverage package scanner **85.1% → 96.3%** |

Gate yang dijalankan: `make be-test` pass, `go build ./...` pass, `go vet ./...` pass,
`make fe-build` pass, `npx tsc --noEmit` pass.

### 9.2 Yang BELUM terverifikasi (butuh akun Drive nyata)

Skenario manual TRASH-012 **belum dijalankan** karena butuh kredensial Google Drive
dan interaksi GUI. Yang masih harus dicek user:

1. Upload `video.mp4` + `video_resized.mp4` ke Drive, hapus `video_resized.mp4`,
   lalu Resize Audit dengan resized scope `active` → harus **Pending**.
2. Ulangi dengan scope `trashed` dan `all` → harus **Resized** + status `Deleted`.
3. **Hapus satu folder berisi file**, cek apakah isinya terdeteksi pada scope `trashed`.
   Ini risiko utama yang belum terbukti (lihat §5.1 dan Edge Cases): anak dari folder
   yang dihapus punya `explicitlyTrashed = false`, dan belum diketahui apakah relasi
   `parents`-nya masih bisa ditelusuri lewat query `'<id>' in parents and trashed = true`.
   Jika ternyata tidak terdeteksi, tambahkan catatan di UI:
   "hanya file yang dihapus langsung yang terdeteksi".
4. Export hasil ke clipboard → pastikan penanda `[deleted]` muncul.

### 9.3 Catatan teknis lain

- `make fe-lint` **gagal** dengan `Cannot find package '@eslint/js'` — eslint tidak
  terdaftar di `frontend/package.json` walau `eslint.config.mjs` ada. Kondisi
  pre-existing, tidak disebabkan fitur ini. Typecheck (`tsc --noEmit`) dipakai sebagai
  gantinya.
- `make be-fmt-check` menandai `app.go`, `internal/domain/resize_audit.go`,
  `internal/domain/resize_audit_test.go`. Sudah diverifikasi bahwa ketiganya identik
  dengan output `gofmt` kecuali line ending (CRLF); bukan akibat perubahan fitur ini
  dan sengaja tidak direformat.
- `npx prettier --check` menandai 16 file frontend termasuk file yang tidak disentuh
  (`main.tsx`, `Sidebar.tsx`, `style.css`). Drift pre-existing, sengaja tidak
  direformat agar diff fitur tetap fokus.
- `NewRealDriveAPI` tercatat 75% coverage: branch error konstruksi `drive.NewService`
  tidak bisa dipicu dengan `http.Client` yang valid.
