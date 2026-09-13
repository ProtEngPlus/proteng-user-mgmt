# Backfill Email Lowercase

Script สำหรับแก้ field `email` ใน collection `users` และ `admins` ให้เป็นตัวพิมพ์เล็กทั้งหมด ให้ตรงกับกฎ `NormalizeEmail` ที่ฝั่ง app บังคับใช้อยู่แล้ว

## รันตอนไหน

- หลัง deploy fix เรื่อง normalize email (BLB11) ลง environment นั้น ๆ ครั้งแรก
- ก่อนที่ `EnsureIndexes()` จะสร้าง unique index แบบ case-insensitive บน `email` สำเร็จ - ถ้ายังมี email ซ้ำกันคนละตัวพิมพ์ค้างอยู่ Mongo จะสร้าง index นี้ไม่ผ่าน

## วิธีรัน

รันแบบ `-dry-run=true` (default) ก่อนเสมอ เพื่อดูว่าจะเปลี่ยนอะไรบ้าง โดยยังไม่เขียนอะไรจริง:

```bash
ENV=local go run ./cmd/backfill-lowercase-emails -collection=all
```

Flag ที่มี:

- `-collection` - `users`, `admins`, หรือ `all` (default `all`)
- `-dry-run` - `true` (default, ไม่เขียนข้อมูลจริง) หรือ `false` (apply ของจริง)

ถ้าเจอ email ซ้ำกันคนละตัวพิมพ์ script จะหยุดทำงานทันทีโดยไม่แก้อะไรเลย - ต้องไปจัดการเอง (merge หรือลบ account ที่ซ้ำ) ก่อนรันใหม่

พอ dry-run แล้วผลลัพธ์ดูถูกต้อง ค่อย apply จริง:

```bash
ENV=local go run ./cmd/backfill-lowercase-emails -collection=all -dry-run=false
```

รันซ้ำได้ไม่พัง - document ที่เป็นตัวพิมพ์เล็กอยู่แล้วจะถูกข้ามไป

เปลี่ยน `ENV=local` เป็น environment ปลายทาง (เช่น `ENV=staging`) เวลารันจริงกับ staging/production
