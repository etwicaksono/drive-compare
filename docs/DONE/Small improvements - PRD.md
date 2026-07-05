# PRD: Small Improvements

**Status:** Draft
**Last Updated:** 2026-07-06

---

## 1. Respect .gitignore saat scan

### Deskripsi

Saat LocalScanner menemukan file `.gitignore` di dalam folder yang sedang di-scan, baca isinya dan tambahkan pattern-pattern tersebut ke ignore list secara otomatis. Ini mencegah file-file seperti `node_modules/`, `dist/`, `*.pyc`, dll. masuk ke hasil comparison.

### Behavior

- Jika root folder yang di-scan mengandung `.gitignore`, parse isinya.
- Pattern `.gitignore` berlaku relatif terhadap lokasi file `.gitignore` tersebut.
- Nested `.gitignore` (di subfolder) juga dihormati, berlaku untuk subfolder tersebut.
- Baris kosong dan baris yang diawali `#` (comment) diabaikan.
- Fitur ini **opsional** — user bisa toggle on/off via Settings atau config.
- Default: **enabled**.

### Config

```yaml
ignore:
  respect_gitignore: true  # default true
```

### Scope

| File | Perubahan |
|------|-----------|
| `config.go` | Tambah field `RespectGitignore bool` di Ignore config |
| `local_scanner.go` | Cek `.gitignore` saat walk, parse pattern, apply ke child paths |
| `SettingsPage.tsx` | Toggle on/off untuk respect .gitignore |
| `app.go` | Wails binding untuk get/set respect_gitignore |

---

## 2. Ganti icon default Wails

### Deskripsi

Ganti icon default Wails (logo Wails) dengan icon yang lebih representatif untuk aplikasi Personal Digital Inventory.

### Scope

- Buat atau pilih icon yang merepresentasikan file inventory / comparison (misal: folder dengan magnifying glass, atau dua folder dengan panah).
- Format: PNG 1024x1024 (Wails akan resize otomatis).
- Replace file di `build/appicon.png`.
- Wails otomatis embed icon ini saat build.

| File | Perubahan |
|------|-----------|
| `build/appicon.png` | Replace dengan icon baru |

---

## 3. Header tabel sticky saat scroll

### Deskripsi

Saat user scroll ke bawah di ResultsTable, header kolom (No, Filename, Extension, Size, Relative Path) tetap terlihat di atas tabel (sticky header).

### Behavior

- Header row (`<thead>`) tetap fixed di atas container scroll.
- Body (`<tbody>`) yang scroll, bukan seluruh tabel.
- Header tetap align dengan kolom body.
- Background header solid (bukan transparent) agar tidak overlap dengan content.

### Scope

| File | Perubahan |
|------|-----------|
| `ResultsTable.tsx` | `<thead>` sticky top, background solid, z-index |

### Implementasi

```css
thead th {
  position: sticky;
  top: 0;
  z-index: 10;
  background: /* neutral-900 */;
}
```

---

## 4. Task Breakdown

| No | Task | Estimasi |
|----|------|----------|
| 1 | Sticky header di ResultsTable | Kecil — CSS only |
| 2 | Ganti app icon | Kecil — replace 1 file |
| 3 | Respect .gitignore di LocalScanner | Sedang — parse .gitignore, apply patterns per folder |
