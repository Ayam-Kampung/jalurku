# jalurku

## 🚀 Quick Start

Sebelumnya untuk reproduksi lokal buatlah kontainer database menggunakan docker, podman, atau semacamnya. Dengan opsi variabel berikut
- https://www.docker.com/blog/how-to-use-the-postgres-docker-official-image/

```env
POSTGRES_USER=myuser
POSTGRES_PASSWORD=mypassword
POSTGRES_NAME=mydb
```
Variabel diatas didapatkan dari `.env` di akar proyek.

Terakhir jalankan dengan:

```bash
go run main.go
```

Jika anda menggunakan Visual Studio Code, gunakan fitur Run & Debug agar dapat mempermudah hidup masing masing tim Ayam Kampung 🙏. Dengan mengeksekusi tombol `F5`

## API

### Autentikasi

#### Registrasi
```http
POST /api/auth/register
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "password123"
}
```

#### Log Masuk
```http
POST /api/auth/login
Content-Type: application/json

{
  "identity": "john@example.com",
  "password": "password123"
}
```

Response:
```json
{
  "status": "success",
  "message": "Success login",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "uuid",
      "username": "John Doe",
      "email": "john@example.com",
      "role": "user"
    }
  }
}
```

### Manajemen pengguna

#### Mendapatkan Pengguna Sekarang
```http
GET /api/users/me
Authorization: Bearer <token>
```

#### Mendapatkan pengguna dengan id
```http
GET /api/users/:id
Authorization: Bearer <token>
```

#### Memperbarui pengguna
```http
PUT /api/users/:id
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "John Updated"
}
```

#### Menghapus pengguna
```http
DELETE /api/users/:id
Authorization: Bearer <token>
```