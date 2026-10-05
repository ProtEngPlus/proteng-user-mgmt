# proteng-user-mgmt

proteng-user-mgmt ดูแลบัญชีของ ProtEngPlus ได้แก่ การสมัครสมาชิก, การยืนยันอีเมล, การ login ที่ออก JWT, การลืมและเปลี่ยนรหัสผ่าน และบัญชี admin นอกจากนี้ยังอ่าน queue `job_status_email_notification` เพื่อส่งอีเมลแจ้งผลของ job service นี้ถูกเรียกจาก [proteng-bff](https://github.com/ProtEngPlus/proteng-bff) เท่านั้น ภาพรวมของระบบอยู่ที่ [manual-guides-2023](https://github.com/ProtEngPlus/manual-guides-2023/blob/main/reference/architecture.md)

## เริ่มใช้

ถ้ายังไม่เคยตั้งเครื่อง ให้ทำตาม [tutorials/01-local-setup.md](https://github.com/ProtEngPlus/manual-guides-2023/blob/main/tutorials/01-local-setup.md) ของ hub ซึ่งตั้งทุก repo พร้อมกัน ถ้าจะตั้งเฉพาะ repo นี้ ให้รันใน Git Bash

```sh
make setup
make -C ../manual-guides-2023 local-keys
make -C ../manual-guides-2023 infra-up
make run
```

- `make setup` ติดตั้ง git hook สร้าง `.env.local` จาก `.env.example` (ถ้ามีอยู่แล้วจะไม่ทับ) และดาวน์โหลด Go module
- `make -C ../manual-guides-2023 local-keys` ใส่ `ACCESS_TOKEN_PRIVATE_KEY` สำหรับเซ็น JWT และใส่ public key ที่เป็นคู่กันให้ bff
- `make run` เปิด Mailhog แล้วรันด้วย `ENV=local` ผ่านเมื่อเห็น `proteng-user-mgmt is running on :8082` อีเมลทุกฉบับที่ส่งในเครื่องจะไปอยู่ที่ <http://localhost:8025>

ต้องใช้ Go 1.22 ขึ้นไป, Docker สำหรับ Mailhog และ `pip install pre-commit`

## คำสั่ง

| คำสั่ง | ทำอะไร |
| --- | --- |
| `make setup` | ติดตั้ง hook, สร้าง `.env.local` และดาวน์โหลด module รันซ้ำได้ |
| `make run` | เปิด Mailhog แล้วรันในเครื่องจาก root ของ repo |
| `make mailhog-up` และ `make mailhog-down` | เปิดหรือปิด Mailhog เอง |
| `make check` | gofmt, `go vet` และ `go test` เหมือนกับ CI |
| `make fmt` | จัด format ด้วย gofmt |
| `make build` และ `make docker-build` | compile ทุก package และ build image ในเครื่อง |

## Config

user-mgmt อ่าน `.env.local` เมื่อรันด้วย `ENV=local` ส่วน dev และ production ได้ค่าจาก ConfigMap และ SealedSecret ใน devops-k8s

| ตัวแปร | ค่าตอนรัน local | ใช้ทำอะไร |
| --- | --- | --- |
| `HTTP_PORT` | `8082` | port ที่ service ฟัง |
| `RABBITMQ_URL` | `amqp://guest:guest@localhost:5672/` | broker |
| `JOB_QUEUE` | `job_status_email_notification` | queue ที่ conductor ส่งเหตุการณ์ของ job มาให้ส่งอีเมล |
| `MONGO_URI` | `mongodb://localhost:27017` | MongoDB |
| `MONGO_DB` | `proteng_local` | ชื่อ database ถ้าไม่ตั้ง ค่า default ในโค้ดคือ `proteng-dev` |
| `ACCESS_TOKEN_PRIVATE_KEY` | ใส่ด้วย `make -C ../manual-guides-2023 local-keys` | base64 ของ RSA private key แบบ PEM (PKCS1 หรือ PKCS8) ใช้เซ็น JWT |
| `ACCESS_TOKEN_EXPIRED_IN`, `ACCESS_TOKEN_MAXAGE` | `1h`, `3600` | อายุของ token |
| `EMAIL_FROM`, `SMTP_HOST`, `SMTP_PORT` | `noreply@localhost`, `localhost`, `1025` | ผู้ส่งและ SMTP server ค่า local ชี้ไปที่ Mailhog `SMTP_PORT` ต้องเป็นตัวเลข ถ้าว่าง service จะหยุดตั้งแต่เริ่ม |
| `SMTP_USER`, `SMTP_PASS` | ว่าง | บัญชี SMTP จริง ใช้เฉพาะ dev และ production |
| `ORIGIN` | `http://localhost:5173` | URL ของหน้าเว็บ ใช้สร้างลิงก์ในอีเมล เช่นลิงก์ยืนยันและลิงก์ reset password |

## โครงสร้างโค้ด

| ที่อยู่ | มีอะไร |
| --- | --- |
| `apis/routes/`, `apis/controllers/` | endpoint ของ `/auth/*`, `/users/*` และ `/admins/*` |
| `models/`, `repositories/` | struct ของ user และ admin และการอ่านเขียน MongoDB `EnsureIndexes` สร้าง unique index ของ email ตอนเริ่ม |
| `internal/rabbitmq/consumer/` | อ่าน `JOB_QUEUE` แล้วส่งอีเมลแจ้งผลของ job ตามการตั้งค่าของผู้ใช้ |
| `utils/email.go` | `SendEmail` สร้างอีเมลจาก template แล้วส่งผ่าน SMTP |
| `utils/token.go`, `utils/password.go` | สร้าง JWT และ hash รหัสผ่าน |
| `utils/email_normalize.go` | `NormalizeEmail` ทำให้ email เป็นตัวพิมพ์เล็กก่อนบันทึกและค้นหา |
| `templates/` | template ของอีเมล โหลดด้วย `template.ParseGlob("templates/*.html")` |
| `cmd/backfill-lowercase-emails/` | tool แก้ email เก่าใน database ให้เป็นตัวพิมพ์เล็ก ดู [README ของ tool](./cmd/backfill-lowercase-emails/README.md) |

## ข้อควรระวัง

- ต้องรันจาก root ของ repo เพราะ template ถูกโหลดด้วย relative path `make run` ทำให้แล้ว ถ้ารันจากที่อื่น service จะ panic ตอนเริ่ม
- private key ต้องเป็นคู่กับ `ACCESS_TOKEN_PUBLIC_KEY` ของ bff ใน environment เดียวกัน ถ้าเปลี่ยน key ต้องเปลี่ยนทั้งสองฝั่งพร้อมกัน วิธีอยู่ที่ [how-to/rotate-secrets.md](https://github.com/ProtEngPlus/manual-guides-2023/blob/main/how-to/rotate-secrets.md)
- ถ้าชี้ `MONGO_URI` ไปที่ MongoDB ที่ใช้ร่วมกัน ให้ตั้ง `MONGO_DB` เป็นชื่อของตัวเอง เช่น `proteng_<ชื่อ>` ห้ามใช้ `proteng-dev` หรือ `proteng-production` และห้ามปล่อยว่าง เพราะค่า default คือ `proteng-dev`
- ห้าม log เนื้อหาอีเมล, token หรือค่าของ SMTP โค้ดตอนนี้ยังทำอยู่ใน `SendEmail` ดูหัวข้อ Security ของ [explanation/known-issues.md](https://github.com/ProtEngPlus/manual-guides-2023/blob/main/explanation/known-issues.md)

## Deploy

push เข้า `dev` จะ build image และ deploy ขึ้น dev ส่วน push เข้า `main` จะ deploy ขึ้น production ทั้งสองแบบเกิดขึ้นทันทีทุกครั้งที่ push แม้จะแก้แค่ docs วิธีตรวจและ rollback อยู่ที่ [how-to/deploy-app.md](https://github.com/ProtEngPlus/manual-guides-2023/blob/main/how-to/deploy-app.md)

## ลิงก์

- กติกาการทำงานและ hook ของ repo นี้: [CONTRIBUTING.md](./CONTRIBUTING.md)
- API ทั้งหมดอยู่ใน Swagger ของ bff ที่ <http://localhost:8080/swagger/index.html> เมื่อรัน bff ในเครื่อง
- เอกสารของทั้งระบบ: [manual-guides-2023](https://github.com/ProtEngPlus/manual-guides-2023/blob/main/README.md)
