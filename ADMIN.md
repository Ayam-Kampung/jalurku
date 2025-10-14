# Jalurku Admin API
## Manajemen Pertanyaan
Di frontend bagusnya dijadikan kartu. Terdapat input untuk menggantikan data pertanyaan seperti: `text`, `image`, dan `meta`. Dan setiap kartu pertanyaan terdapat dropdown untuk memilih relasi pertanyaan dengan jurusan, tombol simpan `PUT /api/pertanyaan/:id` dan tombol hapus `DELETE /api/pertanyaan/:id` di setiap kartunya. Tampilkan juga tampilkan id pertanyaan agar keren aja si di setiap kartunya.
Untuk mengetahui semua jurusan di database gunakan `GET /api/jurusan`. Dan tambahkan tombol tambah untuk menambah pertanyaan baru `POST /api/pertanyaan/`.
### Mendapatkan Semua Pertanyaan
Di bagian `"data": [...]` terdapat banyak kumpulan pertanyaan secara urut. Setiap pertanyaan itu sebuah kartu pertanyaan.
```http
GET /api/pertanyaan/
```
Respon:
```http
{
  "data": [
    {
      "id": "c98f5310-9ef0-4620-91b0-f20e0e4c4c8d",
      "text": "maka dari itu ambil saja hal-hal yang bagus itu!",
      "meta": "",
      "image": "",
      "jurusan_id": 1,
      "CreatedAt": "2025-10-11T20:02:05.93154+07:00",
      "UpdatedAt": "2025-10-11T20:02:05.93154+07:00",
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
      "id": "541c5079-f34b-4321-8550-bbec6ec97669",
      "text": "halo teman-teman",
      "meta": "",
      "image": "",
      "jurusan_id": 1,
      "CreatedAt": "2025-10-11T20:02:13.947344+07:00",
      "UpdatedAt": "2025-10-11T20:02:13.947344+07:00",
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
  "message": "pertanyaan_fetched_all",
  "status": "success"
}
```
### Membuat pertanyaan baru
```http
POST /api/pertanyaan/
Content-Type: application/json
    {
      "text": "Pertanyaan baru dan",
      "meta": "",
      "image": "",
      "jurusan_id": 1
    }
```
Respon
```http
{
  "data": {
    "id": "6df26c9e-4818-4918-b0c1-8eb16c728c89",
    "text": "Pertanyaan baru dan",
    "meta": "",
    "image": "",
    "jurusan_id": 1,
    "CreatedAt": "2025-10-14T12:05:33.435371852+07:00",
    "UpdatedAt": "2025-10-14T12:05:33.435371852+07:00",
    "Jurusan": {
      "id": 1,
      "name": "PG",
      "CreatedAt": "2025-10-11T20:01:58.554739+07:00",
      "UpdatedAt": "2025-10-11T20:01:58.554739+07:00"
    }
  },
  "message": "pertanyaan_create",
  "status": "success"
}
```
### Update Pertanyaan
```http
PUT /api/pertanyaan/:id
Content-Type: application/json
    {
      "text": "Pertanyaan baru dan",
      "meta": "",
      "image": "",
      "jurusan_id": 1
    }
```
Respon
```http
{
  "data": {
    "id": "898a7c0c-b026-41b3-94ed-c8334dae57f3",
    "text": "Pertanyaan baasdasdsadasdru dan",
    "meta": "",
    "image": "",
    "jurusan_id": 1,
    "CreatedAt": "2025-10-14T12:09:16.032219+07:00",
    "UpdatedAt": "2025-10-14T12:14:01.954695423+07:00",
    "Jurusan": {
      "id": 1,
      "name": "PG",
      "CreatedAt": "2025-10-11T20:01:58.554739+07:00",
      "UpdatedAt": "2025-10-11T20:01:58.554739+07:00"
    }
  },
  "message": "pertanyaan_update",
  "status": "success"
}
```
### Menghapus Pertanyaan
```http
DELETE /api/pertanyaan/:id
```
Respon:
```
{
  "message": "Pertanyaan berhasil dihapus",
  "status": "success"
}
```

