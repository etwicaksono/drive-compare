# Task Breakdown — Small Improvements

**Last Updated:** 2026-07-06
**PRD:** `docs/TODO/Small improvements - PRD.md`

---

## Execution Rules

- Selesaikan **satu task** pada satu waktu.
- Setiap task selesai harus: build pass, test pass, format pass, lalu commit.
- Format commit: `feat(IMP-XXX): short description` atau `fix(IMP-XXX): ...`
- Jangan gabungkan beberapa task dalam satu commit.

### Branching Rules

- Setiap task dikerjakan di **branch baru** dari `main`.
- Format nama branch: `imp-XXX/short-description`.
- Setelah task selesai, **squash merge** ke `main`.
- Setelah merge, rename branch lama ke `merged/{nama-branch}`.

---

## Milestone 1 — Quick Wins

### IMP-001: Sticky header di ResultsTable

**Description:** Buat header tabel tetap terlihat di atas saat user scroll ke bawah.

**Files:**
- `frontend/src/components/ResultsTable.tsx`

**Detail:**
- Tambahkan Tailwind classes pada setiap `<th>`:
  - `sticky top-0 z-10 bg-neutral-900`
- Scroll container sudah ada (`<div ref={scrollRef} onScroll={handleScroll}>`), `<thead>` sticky relatif terhadap container ini.
- Pastikan background solid agar text body tidak tembus di belakang header.

**Acceptance Criteria:**
- Scroll ke bawah → header (No, Filename, Extension, Size, Relative Path) tetap visible.
- Header tidak transparan, text body tidak tembus.
- Tidak ada layout shift atau kolom misalign.
- Build pass, TSC pass.

---

### IMP-002: Ganti app icon

**Description:** Replace icon default Wails dengan icon yang representatif.

**Files:**
- `build/appicon.png`

**Detail:**
- Generate icon PNG 1024x1024 yang merepresentasikan Personal Digital Inventory.
  - Konsep: dua folder dengan magnifying glass, atau ikon inventory/comparison.
  - Warna: sesuai tema app (dark blue / neutral).
- Replace `build/appicon.png`.
- Wails otomatis resize dan embed saat build.
- Juga update `build/windows/icon.ico` jika ada (Wails generate dari appicon.png).

**Acceptance Criteria:**
- `build/appicon.png` sudah diganti.
- `wails build` berhasil, icon baru terlihat di taskbar dan title bar window.
- Icon terlihat jelas di ukuran kecil (16x16, 32x32).
- Build pass.

---

## Milestone 2 — Respect .gitignore

### IMP-003: Add RespectGitignore config field

**Description:** Tambahkan field konfigurasi untuk toggle respect .gitignore.

**Files:**
- `internal/infrastructure/config/config.go`
- `config.yaml`

**Detail:**
- Ubah field `Ignore` di Config dari `[]string` menjadi struct:
  ```go
  type IgnoreConfig struct {
      Patterns         []string `yaml:"patterns"`
      RespectGitignore bool     `yaml:"respect_gitignore"`
  }
  ```
- Update `Config.Ignore` dari `[]string` menjadi `IgnoreConfig`.
- Default `RespectGitignore`: `true`.
- Update `config.yaml`:
  ```yaml
  ignore:
    patterns:
      - node_modules
      - .git
      - Thumbs.db
      - desktop.ini
      - "*.tmp"
      - "*.cache"
    respect_gitignore: true
  ```
- Update semua referensi ke `cfg.Ignore` di `app.go` → `cfg.Ignore.Patterns`.
- Update unit test config.

**Acceptance Criteria:**
- Config load berhasil dengan format baru.
- Backward compatibility: field `respect_gitignore` opsional, default `true`.
- Semua referensi `cfg.Ignore` di app.go sudah diupdate.
- Unit test config pass.
- Build pass.

---

### IMP-004: Parse .gitignore in LocalScanner

**Description:** Saat scan, baca file `.gitignore` dan apply pattern-nya.

**Files:**
- `internal/infrastructure/scanner/local_scanner.go`
- `internal/infrastructure/scanner/local_scanner_test.go`

**Detail:**
- Tambahkan field `respectGitignore bool` di `LocalScanner`.
- Update `NewLocalScanner(ignorePatterns []string, respectGitignore bool)`.
- Saat `WalkDir`:
  1. Jika ketemu file `.gitignore` dan `respectGitignore == true`:
     - Baca isinya.
     - Parse: skip baris kosong dan baris `#` (comment).
     - Trim trailing `/` (directory indicator) — treat sebagai name pattern.
     - Simpan patterns di map `gitignorePatterns[dirPath][]string`.
  2. Saat cek `shouldIgnore` untuk file/folder:
     - Cek patterns dari `.gitignore` di folder saat ini dan semua parent folders hingga root.
     - Cek menggunakan `filepath.Match` (sama seperti ignore patterns biasa).
- `.gitignore` file itu sendiri **tidak** di-skip (tetap masuk scan result, kecuali ada ignore pattern yang match).

**Acceptance Criteria:**
- Folder dengan `.gitignore` berisi `dist` → folder `dist/` di-skip.
- `.gitignore` berisi `*.log` → file `error.log` di-skip.
- `.gitignore` berisi comment `# build output` → comment diabaikan.
- `.gitignore` berisi baris kosong → diabaikan.
- Nested `.gitignore` di subfolder → pattern berlaku untuk subfolder tersebut.
- `respectGitignore = false` → `.gitignore` diabaikan sepenuhnya.
- Unit test minimal 6 test cases.
- Build pass, semua test pass.

---

### IMP-005: Update app.go and SettingsPage for gitignore toggle

**Description:** Tambahkan Wails binding dan UI toggle untuk respect .gitignore.

**Files:**
- `app.go`
- `frontend/src/pages/SettingsPage.tsx`

**Detail:**
- **Backend:**
  - `GetRespectGitignore() bool` — return `cfg.Ignore.RespectGitignore`.
  - `SetRespectGitignore(value bool)` — set `cfg.Ignore.RespectGitignore = value`.
  - Update `RunComparison` dan `RunDuplicateScan` (jika ada) untuk pass `respectGitignore` ke `NewLocalScanner`.
- **Frontend (SettingsPage):**
  - Section baru atau extend Ignore Patterns section.
  - Toggle switch / checkbox: "Respect .gitignore files".
  - Label hint: "When enabled, patterns in .gitignore files are automatically applied during local scans."
  - Fetch status dari `GetRespectGitignore()` saat mount.
  - Toggle → call `SetRespectGitignore(value)`.

**Acceptance Criteria:**
- Toggle muncul di SettingsPage.
- Toggle on → local scan menghormati .gitignore.
- Toggle off → .gitignore diabaikan.
- State persist selama app berjalan (in-memory via cfg).
- Build pass, Wails bindings ter-generate.
