# jalurku
Website untuk JHIC

## Reproduksi

### Jagoan Cloud
`jalurku` sudah dideploy melalui Jagoan Cloud, untuk mensinkronkan, ikuti langkah berikut.

`jalurku > Deployments > ROOT > Update from GIT` 

![Update from GIT](jagoan.png)

### Linux/MacOS/Mirip-UNIX
Sebelumnya terlebih dahulu klon repositori ini, dan pasang `go` sesuai dengan cara pemasagan paket pada setiap sistem operasi.

Praktik baik yaitu mengunduh dan merapikan keperluan dependensi. Maka dari itu lakukan berikut.
```sh
$ go mod tidy
```

Terakhir, jalankan websitenya. Biasanya berada di http://127.0.0.1:3000
```sh
$ go run main.go
```

### Windows/NT
Sebelumnya terlebih dahulu klon repositori ini, dan pasang `go`. Versi 1.17 keatas.
- https://go.dev/dl/ 

Praktik baik yaitu mengunduh dan merapikan keperluan dependensi. Maka dari itu lakukan berikut.
```sh
$ go mod tidy
```

Terakhir, jalankan websitenya. Biasanya berada di http://127.0.0.1:3000
```sh
$ go run main.go
```