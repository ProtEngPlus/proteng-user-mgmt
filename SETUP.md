# Setup

Setup ทั้งระบบครั้งแรกดูที่ [Guidebook](https://github.com/ProtEngPlus/manual-guides-2023/blob/main/README.md) ไฟล์นี้มีแค่รายละเอียดเฉพาะของ `proteng-user-mgmt`

## รันบนเครื่อง

ขั้นตอนหลัก (`cp .env.example .env.local` → `go mod tidy` → `./run.sh` และต้องมี RabbitMQ กับ
MongoDB local) อยู่ใน Guidebook §4.3–4.4 `.env.example` มี default local ครบแล้ว (`RABBITMQ_URL`,
`MONGO_URI`, `MONGO_DB`)

ที่ต้องรู้เพิ่มเฉพาะ user-mgmt:

- `ACCESS_TOKEN_PRIVATE_KEY` ต้องเป็นค่าจริง เป็น base64 ของ PEM RSA key (PKCS1 หรือ PKCS8)
  gen เองได้:

  ```sh
  openssl genrsa 2048 | tr -d '\r' | openssl base64 -A
  ```

  public key คู่นี้เอาไปใส่ `ACCESS_TOKEN_PUBLIC_KEY` ของ `proteng-bff` ด้วย
- `SMTP_*` block ต้องใส่เฉพาะตอนจะส่งอีเมลจริง
- ไม่อยากรัน Mongo local จะชี้ `MONGO_URI` ไป shared cluster ก็ได้ ตั้ง `MONGO_DB` เป็นชื่อตัวเอง
  อย่าใช้ `proteng-dev` / `proteng-production`
- **ต้องรันจาก root ของ repo** เพราะ email template โหลดด้วย relative glob path
- เสร็จเมื่อ terminal พิมพ์ `proteng-user-mgmt is running on :8082` (หรือ `HTTP_PORT` ที่ตั้ง) แล้วไม่ crash

## Format & lint

`gofmt` autofix ตอน save/commit, `go vet` รายงานอย่างเดียวต้องแก้เอง รันมือทั้ง repo:

```sh
gofmt -l -w .
go vet ./...
```

ทั้งคู่รันเป็น pre-commit hook ให้อัตโนมัติ (ดู [CONTRIBUTING.md](./CONTRIBUTING.md)) และรันใน CI
ทุก push ด้วย

## API docs

user-mgmt ถูกเรียกจาก proteng-bff เท่านั้น (frontend ไม่เรียกตรง) API docs อยู่ที่ Swagger ของ
**bff** ไม่ใช่ที่นี่: `http://localhost:8080/swagger/index.html` (ดู `proteng-bff/SETUP.md`)

## Build (ถ้าจะทดสอบ deploy)

env var ไม่ถูก bake เข้า image ส่งตอน run:

```sh
docker build -t proteng-user-mgmt .
docker run -d --name proteng-user-mgmt --env-file .env.local -p 8082:8082 proteng-user-mgmt
```
