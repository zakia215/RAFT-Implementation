# Tubes IF3230 Sistem Paralel dan Terdistribusi

## Implementasi Protokol Konsensus: Raft

Oleh kelompok sistol (Kelompok 11 K02)

-   13521121 Saddam Annais Shaquille
-   13521142 Enrique Alifio Ditya
-   13521146 Muhammad Zaki Amanullah
-   13521166 Mohammad Rifqi Farhansyah
-   13521170 Haziq Abiyyu Mahdy

## Cara Menjalankan Program

### Cara Menjalankan Node

1. Leader node

    ```
    go run ./server [IP] [Port]
    ```

    Contoh:

    ```
    go run ./server localhost 8000
    ```

2. Follower node

    ```
    go run ./server [IP] [Port] [Leader IP] [Leader Port]
    ```

    Contoh:

    ```
    go run ./server localhost 8001 localhost 8000
    ```

### Cara Menjalankan Client (CLI)

1. Jalankan command berikut

    ```
    go run ./client [Node IP] [Node Port]
    ```

    IP dan Port yang ingin dihubungkan dengan client tidak harus leader node, karena program client otomatis akan melakukan redirect ke leader apabila client meminta eksekusi command ke node yang bukan leader

    Contoh:

    ```
    go run ./client localhost 8000
    ```

2. Jalankan command yang diinginkan (`ping`, `set`, `get`, `del`, `append`, `strln`, dan `log`)
3. Untuk menghentikan program, jalankan command `exit`

### Cara Menjalankan Web Client/Dashboard

1. Install dependencies pada directory `webclient`

    ```
    cd webclient
    npm install
    ```

2. Jalankan program pada development server

    ```
    npm run dev
    ```

3. Pindah ke root directory dan jalankan proxy server. Proxy server bertujuan sebagai perantara antara webclient dan node. Proxy server ini menerima HTTP request dan meneruskannya ke node melalui RPC. Proxy juga bertanggung jawab untuk melakukan redirect jika command dilakukan terhadap node yang bukan leader.

    ```
    cd ..
    go run ./proxy
    ```

4. Buka browser dan masukkan URL web client (defaultnya adalah http://localhost:5173)
   ![](./img/screenshot.jpg)

5. Masukkan IP dan Port dari Node yang ingin dihubungkan (tidak harus leader node, tetapi harus aktif) dan klik `Connect`. Log dari node tersebut akan muncul secara otomatis

6. Pilih command yang diinginkan pada dropdown. Masukkan parameternya (misalnya key atau value) dan klik `Execute`. Hasil operasi akan muncul pada kotak `Server Reply`
