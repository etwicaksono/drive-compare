# Panduan: Memindahkan Google Photos ke Google Drive

Karena Google Photos API sudah membatasi akses ke library user (sejak April 2025), PDI tidak bisa langsung scan Google Photos. Panduan ini menjelaskan cara memindahkan foto/video dari Google Photos ke Google Drive agar bisa digunakan dengan fitur Compare dan Duplicate Detection di PDI.

---

## Metode 1: Google Takeout (Rekomendasi)

Cocok untuk memindahkan **seluruh library** atau album tertentu sekaligus.

### Langkah-langkah:

1. **Buka Google Takeout**
   - Kunjungi [takeout.google.com](https://takeout.google.com)
   - Login dengan akun Google yang sama.

2. **Pilih data yang akan di-export**
   - Klik **Deselect all** untuk membatalkan semua pilihan.
   - Scroll ke bawah dan centang **Google Photos** saja.
   - Klik **All photo albums included** untuk memilih album tertentu, atau biarkan untuk export semua.

3. **Konfigurasi export**
   - Klik **Next step**.
   - **Transfer to:** pilih **Add to Drive** (file akan langsung masuk ke Google Drive).
     - Alternatif: pilih **Download via link** jika ingin download ke komputer dulu.
   - **Frequency:** pilih **Export once**.
   - **File size:** pilih **2 GB** atau **4 GB** (file akan di-split jika melebihi ukuran ini).

4. **Mulai export**
   - Klik **Create export**.
   - Tunggu proses selesai (bisa memakan waktu beberapa menit hingga beberapa hari tergantung ukuran library).
   - Anda akan menerima email notifikasi setelah selesai.

5. **Akses di Google Drive**
   - Buka Google Drive, cari folder **Takeout**.
   - Jika dalam format `.zip` atau `.tgz`, extract terlebih dahulu.
   - File foto/video sekarang tersedia di Drive dan bisa di-scan menggunakan PDI.

### Catatan penting:
- Metadata (tanggal, lokasi) disimpan dalam file `.json` terpisah per foto.
- Struktur folder mengikuti album Google Photos.
- Ukuran storage Google Drive akan bertambah sesuai ukuran foto yang di-export.
- Google Photos dan Google Drive berbagi kuota storage yang sama.

---

## Metode 2: Download Manual + Upload ke Drive

Cocok untuk memindahkan **beberapa foto/video** saja.

### Langkah-langkah:

1. **Download dari Google Photos**
   - Buka [photos.google.com](https://photos.google.com).
   - Pilih foto/video yang ingin dipindahkan (klik satu per satu, atau Shift+Click untuk range select).
   - Klik ikon **Download** (tanda panah ke bawah) atau tekan `Shift+D`.
   - File akan ter-download ke komputer dalam format `.zip` (jika lebih dari 1 file).

2. **Extract file**
   - Extract file `.zip` ke folder di komputer.

3. **Upload ke Google Drive**
   - Buka [drive.google.com](https://drive.google.com).
   - Buat folder tujuan (misal: `Photos Backup`).
   - Drag & drop file dari komputer ke folder Drive, atau klik **New > File upload**.

4. **Scan dengan PDI**
   - Buka PDI, gunakan fitur Compare atau Duplicate Detection.
   - Pilih folder Google Drive yang berisi foto yang sudah di-upload.

---

## Metode 3: Google Photos Desktop Uploader + Drive for Desktop

Cocok untuk **sinkronisasi berkelanjutan** dari perangkat ke Drive.

### Langkah-langkah:

1. **Install Google Drive for Desktop**
   - Download dari [google.com/drive/download](https://www.google.com/drive/download/).
   - Install dan login.

2. **Konfigurasi backup**
   - Buka Google Drive for Desktop settings.
   - Di tab **My Computer**, tambahkan folder foto lokal yang ingin di-backup.
   - Pilih **Back up to Google Drive** (bukan Google Photos).

3. **Scan dengan PDI**
   - Foto sekarang ada di Google Drive.
   - Gunakan PDI untuk compare folder lokal dengan folder Drive.

---

## Setelah dipindahkan ke Google Drive

Setelah foto/video tersedia di Google Drive, Anda bisa menggunakan semua fitur PDI:

- **Compare** — bandingkan folder lokal dengan folder Google Drive untuk menemukan file yang hanya ada di satu tempat.
- **Duplicate Detection** — scan folder Google Drive untuk menemukan file duplikat berdasarkan pola nama dan ukuran.
- **Export** — export daftar file untuk referensi atau pengelolaan lebih lanjut.

---

## FAQ

**Q: Apakah foto di Google Photos akan terhapus setelah di-export ke Drive?**
A: Tidak. Google Takeout hanya menyalin, tidak memindahkan. Foto tetap ada di Google Photos.

**Q: Apakah kuota storage bertambah?**
A: Ya. Jika menggunakan metode "Add to Drive", file akan menempati storage Drive (yang berbagi kuota dengan Google Photos). Jika foto sudah ada di Google Photos, secara efektif storage yang terpakai bisa berlipat.

**Q: Kenapa PDI tidak langsung scan Google Photos?**
A: Sejak April 2025, Google membatasi Photos Library API sehingga hanya bisa mengakses konten yang dibuat oleh aplikasi itu sendiri. Aplikasi pihak ketiga tidak bisa lagi membaca seluruh library foto user.
