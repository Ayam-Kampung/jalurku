# Api Jalurku
Sebuah angket sederhana menggunakan go fiber sebagai REST Api.
## Manajemen Pengguna
### Register/Mendaftar
```http
POST /api/auth/register
Content-Type: application/json
{
	"username": "john"
  	"email": "john@example.com",
  	"password": "password123"
}
```
Respon
```http
{
  "data": {
    "id": "215dc4e6-0434-425f-a340-46d01dca17c7",
    "username": "john",
    "email": "john@example.com",
    "role": "user"
  },
  "message": "user_created",
  "status": "success"
}
```
### Log Masuk
```http
POST /api/auth/login
Content-Type: application/json
{
  	"identity": "john@example.com",
  	"password": "password123"
}
```
Respon simpan di cookie frontend.
```http
{
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "215dc4e6-0434-425f-a340-46d01dca17c7",
      "username": "john",
      "email": "john@example.com",
      "role": "user"
    }
  },
  "message": "login",
  "status": "success"
}
```
### Mendapatkan Info akun sendiri (Middleware Protected)
```http
GET /api/user/me
Authorization: Bearer <token>
```
Respon juga memperlihatkan hasil angket, yang paling baru akan di tulis paling pertama di urutan JSONnya.
```http
{
  "data": {
    "ID": "215dc4e6-0434-425f-a340-46d01dca17c7",
    "Name": "john",
    "Email": "john@example.com",
    "Role": "user",
    "CreatedAt": "2025-10-11T16:20:18.124834+07:00",
    "UpdatedAt": "2025-10-11T16:20:18.124834+07:00",
    "DeletedAt": null,
    "HasilAngket": [
      {
        "id": "bd9de930-67be-47ae-98c0-67fdb11612d0",
        "user_id": "215dc4e6-0434-425f-a340-46d01dca17c7",
        "jurusan_id": 3,
        "CreatedAt": "2025-10-12T12:26:18.232076+07:00",
        "UpdatedAt": "2025-10-12T12:26:18.232076+07:00",
        "DeletedAt": null
      },
      {
        "id": "6b7ca480-b244-4ae4-baf7-cc75e9ea15f3",
        "user_id": "215dc4e6-0434-425f-a340-46d01dca17c7",
        "jurusan_id": 1,
        "CreatedAt": "2025-10-11T22:20:51.876922+07:00",
        "UpdatedAt": "2025-10-11T22:20:51.876922+07:00",
        "DeletedAt": null
      },
		...
    ]
  },
  "message": "user_found",
  "status": "success"
}
```
## Manajemen Angket
### 1. Pembuatan sesi angket
Pertama buatlah sesi angket. Sesinya akan dihandle di Backend redis, memiliki TTL selama 1 jam. Dan tidak di Middleware protected, tetapi jika terdapat user log masuk maka jurusan akan dijadikan rekor histori seperti pada pengambilan info akun sendiri `HasilAngket: [..]`. Guest akan hanya disimpan di redis, dan dihapus pada saat angket selesai, selama sesi masih valid.
```http
POST /api/angket/mulai
```
Respon berupa id sesi, id sesi berguna agar jawaban pertanyaan dapat disimpan untuk sementara. Gunakan id sesi ini untuk melakukan `POST /api/angket/submit`.
```http
{
  "message": "angket_started",
  "session_id": "67353c73-e188-4f1b-a14a-4e9924725f93"
}
```
### 2. Pengambilan Pertanyaan angket/Handler Jawaban angket
Pengambilan pertanyaannya diambil dari database. Pastikan frontend hanya fetching semua pertanyaan dari awal, dan disimpan di penyimpanan sementara frontend. Hingga selesai angket.
```http
GET /api/pertanyaan/rand
```
Respon dari server berupa `"data":[...]` kumpulan soal urutan random dari server, setiap pertanyaan akan ditampilkan di halaman berbeda. Misal pertanyaan `fcaba5da-1492-4f89-ab7a-47c0a013f049` di halaman pertama, `01444e0e-281c-4a46-9548-c8e0c117a59c` di halaman kedua, dst. Pada saat ganti halaman maka lakukan lah `POST /api/angket/submit`.
```http
{
  "data": [
    {
      "id": "fcaba5da-1492-4f89-ab7a-47c0a013f049",
      "text": "Kamu suka berfikir secara logis",
      "meta": "minat",
      "image": "",
      "jurusan_id": 3,
      "CreatedAt": "2025-10-12T12:22:33.398941+07:00",
      "UpdatedAt": "2025-10-12T12:22:33.398941+07:00",
      "Jurusan": {
        "id": 3,
        "name": "TKJ",
        "CreatedAt": "2025-10-11T20:01:58.554739+07:00",
        "UpdatedAt": "2025-10-11T20:01:58.554739+07:00",
        "Pertanyaan": null,
        "HasilAngket": null
      }
    },
    {
      "id": "01444e0e-281c-4a46-9548-c8e0c117a59c",
      "text": "Suka belajar diluar?",
      "meta": "gaya_belajar",
      "image": "outside.png",
      "jurusan_id": 1,
      "CreatedAt": "2025-10-11T20:02:23.397592+07:00",
      "UpdatedAt": "2025-10-11T20:02:23.397592+07:00",
      "Jurusan": {
        "id": 1,
        "name": "PG",
        "CreatedAt": "2025-10-11T20:01:58.554739+07:00",
        "UpdatedAt": "2025-10-11T20:01:58.554739+07:00",
        "Pertanyaan": null,
        "HasilAngket": null
      }
    },
    {
      "id": "02a32d0d-876b-4ad4-a867-c7a596be2dff",
      "text": "Kamu lebih suka belajar secara visual",
      "meta": "gaya_belajar",
      "image": "visual.png",
      "jurusan_id": 1,
      "CreatedAt": "2025-10-11T20:02:33.94736+07:00",
      "UpdatedAt": "2025-10-11T20:49:02.794785+07:00",
      "Jurusan": {
        "id": 1,
        "name": "PG",
        "CreatedAt": "2025-10-11T20:01:58.554739+07:00",
        "UpdatedAt": "2025-10-11T20:01:58.554739+07:00",
        "Pertanyaan": null,
        "HasilAngket": null
      }
    },
	...
  ],
  "message": "pertanyaan_fetched_allrand",
  "status": "success"
```
Lalu membaca dari radio button yang menunjukan level skor likert. 1-5. Yang lalu akan disubmit.
```http
POST /api/angket/submit
Content-Type: application/json
{
  "session_id": "67353c73-e188-4f1b-a14a-4e9924725f93", <- id sesi redis
  "question_id": "fcaba5da-1492-4f89-ab7a-47c0a013f049",
  "selected_option": 3 <- Skala likert
}
```
Respon:
```http
{
  "data": {
    "session_id": "67353c73-e188-4f1b-a14a-4e9924725f93",
    "question_id": "fcaba5da-1492-4f89-ab7a-47c0a013f049",
    "selected_option": 3
  },
  "message": "jawaban_saved_temp"
}
```
Proses ini dilanjutkan oleh pertanyaan kedua di `data: [..]`, dan setersnya... sampai semua pertanyaan di `data: [..]` sudah dijawab semua.
### 3. Mengakhiri Sesi angket
```http
POST /api/angket/selesai
Content-Type: application/json
{
  "session_id": "67353c73-e188-4f1b-a14a-4e9924725f93"
}
```
Respon dari server, detail skor menunjukan angka, dan angka tersebut merepresentasikan jurusannya (1:PG, 2:RPL, 3:TKJ, 4:TJA). Tampilkan jurusan terbaik
```http
{
  "hasil": {
    "detail_skor": {
      "1": 24 -> 1 di database itu PG
		"2": 18
		"3": 19
    },
    "jurusan_terbaik": "PG",
	 "jurusan_terbaik_id": "1",
    "session_id": "fae34a56-c558-46a5-83da-ee62f5f048b3",
    "total_skor": 9
  },
  "message": "angket_finished"
}
```