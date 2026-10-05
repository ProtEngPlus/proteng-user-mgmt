# Contributing

กติกาเรื่อง branch, commit message, PR และ docs ของทุก repo อยู่ที่ [CONTRIBUTING.md ของ hub](https://github.com/ProtEngPlus/manual-guides-2023/blob/main/CONTRIBUTING.md) หน้านี้มีเฉพาะเรื่องของ proteng-user-mgmt

- แตก branch จาก `dev` และเปิด PR เข้า `dev` push เข้า `dev` จะ deploy ขึ้น dev ทันที และ push เข้า `main` จะ deploy ขึ้น production ทันที
- รัน `make check` ก่อน push ทุกครั้ง คำสั่งนี้ตรวจ gofmt, `go vet` และ `go test`
- ฟังก์ชันใหม่ต้องมี test ใน `_test.go` ของ package เดียวกัน และอยู่ใน commit เดียวกัน
- repo นี้เป็น public ห้ามใส่ credential และรายละเอียดของเครื่องที่ใช้ deploy
- ห้าม log เนื้อหาอีเมล, token หรือค่าของ SMTP

## แก้คู่กับ repo อื่น

ของต่อไปนี้ต้องแก้พร้อมกับอีก repo ในงานชุดเดียวกัน และ PR ของทั้งสองฝั่งต้องใส่ `Related: ProtEngPlus/<repo>#<เลข PR>` ถึงกัน รายการเต็มและลำดับการ merge อยู่ใน [CONTRIBUTING ของ hub](https://github.com/ProtEngPlus/manual-guides-2023/blob/main/CONTRIBUTING.md#ของที่ต้องแก้คู่กันข้าม-repo)

- `ACCESS_TOKEN_PRIVATE_KEY` ต้องเป็นคู่กับ `ACCESS_TOKEN_PUBLIC_KEY` ของ bff ใน environment เดียวกัน
- ลิงก์ในอีเมลที่สร้างจาก `ORIGIN` (เช่น `/reset-password` และ `/success-verified`) ต้องตรงกับ route ใน `src/routes/` ของ frontend
- path ของ endpoint ที่เปลี่ยนต้องแก้ใน forwarder ของ bff ด้วย และ env ใหม่ต้องเพิ่มใน devops-k8s

## hook

`make setup` ติดตั้ง hook ให้ด้วย pre-commit

| ตอน | hook |
| --- | --- |
| commit | `gofmt -l -w` และ `go vet` กับไฟล์ Go ที่ staged |
| push | `go build` และ `go test` เพิ่มจากตอน commit |
| เขียน commit message | ปฏิเสธ message ที่ไม่ตรงกับ Conventional Commits |

ถ้า gofmt แก้ไฟล์ให้ระหว่าง commit ให้ `git add` ไฟล์นั้นซ้ำแล้ว commit อีกครั้ง ถ้าบน Windows `gofmt -l .` แสดงไฟล์ออกมาเต็มไปหมดทั้งที่ไม่ได้แก้ แปลว่า working tree เป็น CRLF ดูวิธีแก้ในหัวข้อ Line endings ของ CONTRIBUTING ของ hub
