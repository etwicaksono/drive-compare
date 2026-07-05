# PRD: Export Comparison Result

**Status:** Draft
**Last Updated:** 2026-07-06

---

## 1. Overview

Fitur ini memungkinkan user memilih (select) item-item dari hasil comparison, lalu mengekspor path dari item yang dipilih ke clipboard atau file `.txt`.

---

## 2. User Flow

```
Hasil Compare muncul di ResultsTable
        │
        ▼
User hover kolom "No" pada row → checkbox muncul menggantikan nomor
User klik checkbox untuk memilih item (bisa multi-select)
Header "No" juga bisa diklik untuk Select All / Deselect All
        │
        ▼
User klik tombol "Export Selected" (muncul di toolbar jika ada item tercentang)
        │
        ▼
Modal Export muncul, berisi:
  ┌──────────────────────────────────────────────────────┐
  │  Export 12 selected files                            │
  │                                                      │
  │  Path source:          (hanya muncul di tab "Both")  │
  │  ○ Local Storage                                     │
  │  ○ Google Drive                                      │
  │                                                      │
  │  Export to:                                          │
  │  ○ Clipboard                                         │
  │  ○ Save as .txt file                                 │
  │                                                      │
  │              [ Cancel ]  [ Export ]                   │
  └──────────────────────────────────────────────────────┘
        │
        ▼
Sistem menyusun daftar path sesuai pilihan, lalu:
  - Clipboard → Wails ClipboardSetText()
  - File .txt → Wails SaveFileDialog() → tulis file
```

---

## 3. Detail Behavior

### 3.1 Selection (di ResultsTable)

- **Tidak ada kolom baru.** Checkbox terintegrasi di kolom "No" yang sudah ada.
- **Default state:** menampilkan nomor urut seperti biasa.
- **Hover row:** nomor berubah menjadi checkbox (unchecked/checked sesuai state).
- **Row yang sudah diselect:** checkbox selalu tampil (tidak perlu hover), menggantikan nomor.
- **Header "No":** hover menampilkan checkbox untuk Select All / Deselect All (terhadap item yang lolos filter).
- Selection state di-maintain per tab (localOnly, driveOnly, both).
- Counter "X selected" muncul di toolbar filter saat ada item tercentang.
- Tombol "Export Selected" muncul di toolbar saat ada item tercentang.

### 3.2 Path Source

Menentukan dari perspektif mana path disusun:

| Path Source    | Format Path                                    | Contoh                          |
|----------------|------------------------------------------------|---------------------------------|
| Local Storage  | `{localDir}\{relativePath}{name}{ext}`         | `D:\Photos\2024\photo.jpg`     |
| Google Drive   | `{driveLabel}/{relativePath}{name}{ext}`        | `My Drive/Photos/2024/photo.jpg`|

**Aturan per tab:**

| Tab          | Path Source                  | Behavior                                      |
|--------------|------------------------------|-----------------------------------------------|
| Local Only   | Local Storage (otomatis)     | Tidak ada pilihan path source di modal         |
| Drive Only   | Google Drive (otomatis)      | Tidak ada pilihan path source di modal         |
| Both         | User memilih                 | Radio button Local Storage / Google Drive      |

### 3.3 Export Format

Output berupa daftar path, satu path per baris:

```
D:\Photos\2024\vacation\photo001.jpg
D:\Photos\2024\vacation\photo002.jpg
D:\Photos\2024\documents\report.pdf
```

### 3.4 Export Target

| Target       | Mekanisme                                                        |
|--------------|------------------------------------------------------------------|
| Clipboard    | Wails `runtime.ClipboardSetText()` — paste langsung ke mana saja |
| File .txt    | Wails `runtime.SaveFileDialog()` → tulis file via Go backend     |

---

## 4. Scope

### 4.1 Data yang Dibutuhkan

ResultsTable saat ini sudah punya `relativePath`, `name`, `extension` per item. Yang perlu ditambahkan:

- **localDir** dan **driveFolderLabel** perlu di-pass ke ResultsTable agar bisa menyusun full path.
- **activeTab** perlu di-pass ke ExportModal agar tahu aturan path source.
- Selection state (Set of selected item keys).

### 4.2 Backend (Wails Bindings)

| Binding                                      | Deskripsi                                    |
|-----------------------------------------------|----------------------------------------------|
| `ExportToClipboard(text string)`              | Tulis text ke system clipboard               |
| `ExportToFile(content string) (string, error)`| Buka Save dialog, tulis content ke file .txt |

### 4.3 Frontend Components

| Component        | Perubahan                                                              |
|------------------|------------------------------------------------------------------------|
| `ResultsTable`   | Kolom No → hover checkbox, selection state, toolbar Export Selected     |
| `ExportModal`    | New — path source radio (hanya tab Both), export target radio          |
| `ComparePage`    | Pass localDir + driveFolderLabel ke ResultsTable                       |

---

## 5. Edge Cases

- **Tidak ada item tercentang** → tombol Export tidak muncul.
- **Select All** → hanya select item yang lolos filter saat ini, bukan semua data.
- **Tab switch** → selection di-reset per tab.
- **Export kosong** → tidak mungkin karena tombol hanya muncul jika ada selection.
- **Path separator** → Local Storage pakai `\` (Windows), Google Drive pakai `/`.

---

## 6. Task Breakdown

| No | Task | Deskripsi |
|----|------|-----------|
| 1 | Selection di ResultsTable | Kolom No → hover checkbox, Select All header, selection counter, per-tab state |
| 2 | Export Modal component | Path source radio (hanya tab Both), export target radio, preview count |
| 3 | Backend bindings | `ExportToClipboard`, `ExportToFile` (Save dialog + write) |
| 4 | Path builder logic | Menyusun full path dari localDir/driveLabel + relativePath + name + ext |
| 5 | Wire up export flow | Connect modal → path builder → backend binding |
| 6 | Testing & edge cases | Tab-specific path source, empty selection guard |
