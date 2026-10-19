# GoFiber Workshop — Exercise 5–7.2

โปรเจกต์ฝึกเขียน REST API ด้วย GoFiber v2 ต่อจาก Exercise 5–6 เพิ่ม GORM และ MySQL สำหรับ Dogs/Company โดยใช้โครงสร้างที่อ่านตามได้ง่าย

## สิ่งที่แต่ละ Exercise ทำ

| Exercise | เนื้อหา |
|---|---|
| 5.0 | Basic Auth ครอบทุก API ที่ขึ้นต้นด้วย `/api` |
| 5.1 | อ่าน Path Parameter แล้วคำนวณ Factorial |
| 5.2–5.3 | อ่าน Query Parameter `tax_id`, แปลงเป็น ASCII และใช้ Controller Import Alias `c` |
| 6 | รับ JSON สมัครสมาชิก ตรวจด้วย validator/v10 โดยไม่บันทึกข้อมูล |
| 7.0 | Dogs CRUD และแยก `DogsRes`, `ResultData` ออกมาไว้ใน `models` |
| 7.0.1 | Company Model, AutoMigrate และ CRUD |
| 7.0.2 | อ่านเฉพาะ Dogs ที่ถูก Soft Delete |
| 7.1 | Query Dogs ที่ `dog_id > 50 AND dog_id < 100` |
| 7.2 | แปลง DogID เป็นสีและนับจำนวนแต่ละสี |

## Tech Stack และโครงสร้างไฟล์

- Go ตาม `go.mod` (1.26.0 ขึ้นไป)
- GoFiber v2, validator/v10
- GORM v2 API ผ่าน `gorm.io/gorm`, MySQL driver `gorm.io/driver/mysql`
- `go-sql-driver/mysql` ใช้จัดรูปแบบ DSN และ godotenv ใช้โหลด `.env` จริง
- MySQL Server และ Postman

```text
main.go                         เริ่ม Server, Basic Auth และ Routes
main_test.go                    ตรวจ API เดิมและ Database unavailable ผ่าน HTTP ในหน่วยความจำ
mysql_integration_test.go       Integration test แบบเปิดใช้เองและ Rollback ข้อมูลทดสอบ
controller/
  controller.go                Handlers เดิมของ Exercise 5–6
  api_helpers.go               Database guard, JSON error และตรวจ Primary Key
  dog_controller.go            Dogs CRUD, Queries และสรุปสี
  dog_controller_test.go       ทดสอบสี, ผลรวม, Request และ ID โดยไม่ใช้ MySQL
  company_controller.go        Company CRUD
models/
  register.go                  Request เดิมของการสมัครสมาชิก
  dog.go                       Dogs database model และ DogRequest
  dog_response.go              DogsRes และ ResultData
  company.go                   Company database model และ CompanyRequest
database/
  database.go                  โหลด config, เชื่อมต่อ MySQL และ AutoMigrate
  database_test.go             ตรวจ config ก่อนเชื่อมต่อ
postman/
  INET_GoFiber_Postman_Collection.json          Collection เดิมของ Exercise 5–6
  GoFiber_Workshop_7.postman_collection.json    Collection ใหม่ของ Exercise 7
.env.example                   ตัวอย่าง config
.gitignore                     กัน .env และ executable ออกจาก Git
```

## ติดตั้ง MySQL บน Windows และสร้าง Database

1. ติดตั้ง MySQL Community Server จาก [เว็บไซต์ MySQL](https://dev.mysql.com/downloads/mysql/) และตั้งรหัสผ่านบัญชีผู้ดูแลระหว่างติดตั้ง ใช้ MySQL Workbench หรือ mysql CLI เพื่อรัน SQL
2. เปิด Services และตรวจว่า service ของ MySQL เช่น `MySQL80` ทำงาน และตรวจพอร์ตจริงของเครื่อง (ตัวอย่างใช้ 3306)
3. เชื่อมต่อด้วยบัญชีผู้ดูแลของเครื่อง เช่นใน PowerShell:

```powershell
& 'C:\Program Files\MySQL\MySQL Server 8.0\bin\mysql.exe' -u root -p
```

ตำแหน่งโปรแกรมอาจต่างกันตามเวอร์ชัน ตัวเลือก `-p` จะถามรหัสผ่านโดยไม่ต้องใส่รหัสผ่านในคำสั่ง

4. รัน SQL ต่อไปนี้ โดยเปลี่ยนรหัสผ่านตัวอย่างเป็นรหัสผ่านที่ใช้ในเครื่องตนเอง:

```sql
CREATE DATABASE IF NOT EXISTS golang_test
  CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

CREATE USER IF NOT EXISTS 'gofiber_workshop'@'localhost'
  IDENTIFIED BY 'replace_with_your_local_password';
GRANT ALL PRIVILEGES ON golang_test.* TO 'gofiber_workshop'@'localhost';
```

`CREATE USER IF NOT EXISTS` ไม่เปลี่ยนรหัสผ่านของบัญชีที่มีอยู่แล้ว ให้ใช้ข้อมูลของบัญชีเดิมหรือจัดการบัญชีผ่านผู้ดูแล หากใช้ MySQL บนเครื่องอื่น ให้ผู้ดูแลกำหนด host ของบัญชีให้ตรงกับเครื่องที่เชื่อมต่อ

ต้องสร้าง Database ก่อน เพราะ `AutoMigrate` จัดการ Tables ภายใน Database และไม่ได้สร้าง Database ให้

## ตั้งค่า Environment Variables หรือ .env

ทำจากโฟลเดอร์โปรเจกต์ใน PowerShell:

```powershell
Copy-Item .env.example .env
notepad .env
```

แก้ไฟล์ให้ตรงกับ MySQL ของตนเอง:

```dotenv
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=gofiber_workshop
DB_PASSWORD="your_local_password"
DB_NAME=golang_test
```

แอปเรียก `godotenv.Load()` ใน `database.Connect()` จึงอ่าน `.env` จริงจาก working directory ค่า Environment Variables ที่ตั้งไว้แล้วจะมีลำดับเหนือ `.env` หากใช้รหัสผ่านที่มีอักขระพิเศษ ให้ใช้รูปแบบการ quote ของ dotenv ให้ถูกต้อง `.env` ถูก ignore โดย Git

อีกทางหนึ่งตั้งค่าใน PowerShell session ที่จะรันแอปได้โดยตรง:

```powershell
$env:DB_HOST = '127.0.0.1'
$env:DB_PORT = '3306'
$env:DB_USER = 'gofiber_workshop'
$env:DB_NAME = 'golang_test'
$credential = Get-Credential -UserName $env:DB_USER -Message 'Enter MySQL credentials'
$env:DB_PASSWORD = $credential.GetNetworkCredential().Password
```

เมื่อไม่กำหนด Host, Port หรือ Database จะใช้ `127.0.0.1`, `3306`, `golang_test` ตามลำดับ แต่ต้องกำหนด `DB_USER`; Password อ่านจาก config เท่านั้น หากบัญชีท้องถิ่นใช้รหัสผ่านว่าง สามารถกำหนดเป็นค่าว่างได้

## ติดตั้ง Dependencies และรัน Server

```powershell
go mod tidy
go build ./...
go test ./...
go run .
```

หาก PowerShell หา `go` ไม่เจอ ให้เพิ่มโฟลเดอร์ Go `bin` ลง PATH หรือใช้ path เต็ม เช่น:

```powershell
& 'C:\Program Files\Go\bin\go.exe' run .
```

Server ใช้พอร์ต **3000** เมื่อเชื่อมต่อและ AutoMigrate สำเร็จจะมีข้อความ `Database connected and migrated` ใน console ถ้าเชื่อมต่อไม่ได้หรือ Migrate ไม่ผ่าน จะแสดงสาเหตุใน console, API ข้อ 5–6 ยังทำงาน และ API Dogs/Company จะตอบ JSON `500`:

```json
{"success": false, "message": "Database unavailable; check server configuration"}
```

หลังแก้การตั้งค่า MySQL ให้หยุดด้วย Ctrl+C และรัน Server ใหม่ ไม่มีการ reconnect หรือ seed ข้อมูลอัตโนมัติ ถ้าฐานข้อมูลขัดข้องหลังเปิดแอปแล้ว Controller จะตรวจ GORM Error และส่ง JSON `500` โดยไม่ส่งรายละเอียด SQL ให้ Client

## Authentication และรายการ API

ทุก endpoint ใช้ Basic Auth ของ Workshop เดิม: Username `gofiber`, Password `21022566` หากไม่ส่งหรือส่งผิดจะได้ JSON `401 Unauthorized`

| Method | Route | ผลสำเร็จ |
|---|---|---|
| GET | `/api/v1/fact/:number` | `200` ผล Factorial ของเลข 0–20 |
| GET | `/api/v3/mean?tax_id=Hello` | `200` รหัส ASCII |
| POST | `/api/v1/register` | `200` ผล Validation เดิม |
| GET | `/api/v1/dog` | `200` Array ของ Dogs ที่ยังไม่ถูกลบ |
| GET | `/api/v1/dog/filter?search=10` | `200` Array ที่มี `dog_id = 10`; ไม่พบเป็น `404` |
| GET | `/api/v1/dog/json` | `200` ข้อมูลพร้อม type, count และจำนวนแต่ละสี |
| POST | `/api/v1/dog/` | `201` Dog ที่สร้าง พร้อม Primary Key `ID` |
| PUT | `/api/v1/dog/:id` | `200` Dog ที่แก้ไข |
| DELETE | `/api/v1/dog/:id` | `200` JSON ยืนยัน Soft Delete |
| GET | `/api/v1/dog/deleted` | `200` `{ "count": ..., "data": [...] }` เฉพาะ Dogs ที่ถูกลบ |
| GET | `/api/v1/dog/range` | `200` `{ "count": ..., "data": [...] }` เฉพาะ DogID 51–99 ที่ยังไม่ถูกลบ |
| POST | `/api/v1/company` | `201` Company ที่สร้าง |
| GET | `/api/v1/company` | `200` `{ "count": ..., "data": [...] }` |
| GET | `/api/v1/company/:id` | `200` Company ที่พบ |
| PUT | `/api/v1/company/:id` | `200` Company ที่แก้ไข |
| DELETE | `/api/v1/company/:id` | `200` JSON ยืนยัน Soft Delete |

Static Routes ของ Dogs ถูกกำหนดก่อน Dynamic Routes โดย Fiber ค่าเริ่มต้นรองรับ trailing slash ทั้งสองรูปแบบ แต่ Collection ใช้รูปแบบตามตาราง

`id` ใน PUT/DELETE Dogs และ Company เป็น Primary Key ที่ Server ส่งกลับใน `ID` ซึ่งต่างจาก `dog_id` ห้ามนำ `dog_id` ไปแทนโดยอัตโนมัติ DogID ไม่ได้กำหนด unique จึงอาจค้นพบหลายรายการใน `/filter`

## Request และ Response ตัวอย่าง

POST/PUT Dogs ใช้ `Content-Type: application/json`:

```json
{"name": "Dog01", "dog_id": 9}
```

POST จะได้ HTTP `201` เช่น:

```json
{
  "ID": 1,
  "CreatedAt": "2026-01-01T00:00:00Z",
  "UpdatedAt": "2026-01-01T00:00:00Z",
  "DeletedAt": null,
  "name": "Dog01",
  "dog_id": 9
}
```

วันที่และ ID เป็นตัวอย่าง Server คำนวณจริง `gorm.Model` ไม่มี JSON Tags จึงใช้ชื่อ `ID`, `CreatedAt`, `UpdatedAt`, `DeletedAt` ตาม Struct ส่วน `/dog/json` ส่งเฉพาะ `name`, `dog_id`, `type` ของ `DogsRes`

POST/PUT Company:

```json
{
  "name": "INET Workshop",
  "address": "Bangkok",
  "phone": "021234567",
  "email": "workshop@example.com"
}
```

Company ที่สร้างหรือแก้ไขจะมีฟิลด์ของ `gorm.Model` แบบเดียวกับ Dogs กฎพื้นฐานเพิ่มเติมที่เลือกใช้สำหรับ Workshop:

- Dogs: name หลังตัดช่องว่างต้องไม่ว่าง ยาวไม่เกิน 100 ตัวอักษร; dog_id ต้องส่งมาเป็นจำนวนเต็มตั้งแต่ 0 ขึ้นไป
- Company: name ต้องไม่ว่างและยาวไม่เกิน 100; address ไม่เกิน 255; phone ไม่เกิน 30; email ไม่บังคับ แต่ถ้าส่งมาต้องถูกต้องและยาวไม่เกิน 254
- PUT ใช้ข้อมูลชุดใหม่: Dogs ต้องส่งทั้ง name และ dog_id; Company ต้องส่ง name และส่ง Optional Fields ตามต้องการ ฟิลด์ Optional ที่ไม่ได้ส่งจะกลายเป็นค่าว่าง
- Client ไม่สามารถกำหนด ID และ timestamps หรือ DeletedAt เพราะรับ Body ผ่าน Request Struct ที่มีเฉพาะฟิลด์ที่อนุญาต

Body ผิดหรือ id ไม่ใช่จำนวนเต็มบวกเป็น `400`; ไม่พบรายการเฉพาะตัวเป็น `404`; Query ล้มเหลวเป็น `500` ตัวอย่าง Error:

```json
{"success": false, "message": "Dog not found"}
```

DELETE สำเร็จตอบ `{"success":true,"message":"Dog deleted"}` หรือ `"Company deleted"` แล้ว GET ปกติจะไม่เห็นข้อมูลนั้น GET แบบรายการที่ไม่พบข้อมูลตอบ `200` และ `[]` หรือ `{ "count": 0, "data": [] }` ตามรูปแบบ endpoint

GET `/api/v1/dog/json` เมื่อไม่มีข้อมูล:

```json
{
  "data": [],
  "name": "golang-test",
  "count": 0,
  "sum_red": 0,
  "sum_green": 0,
  "sum_pink": 0,
  "sum_nocolor": 0
}
```

## GORM และ GetDogsJson ทำงานอย่างไร

`gorm.Model` รวม `ID`, `CreatedAt`, `UpdatedAt`, `DeletedAt` ไว้ให้ ID เป็น Primary Key ส่วน `DeletedAt` ทำให้ `Delete` เป็น Soft Delete โดยตั้งวันเวลาที่ลบแทนการลบแถวออกจริง Query ปกติจะกรองแถวเหล่านี้ออก

`AutoMigrate(&models.Dogs{}, &models.Company{})` ตรวจ Struct แล้วสร้างหรือปรับตารางที่จำเป็น ชื่อตารางตาม GORM Default คือ `dogs` และ `companies` โค้ดนี้ไม่ได้กำหนด `TableName()`, ไม่ Drop Table และไม่ Reset ข้อมูล

`Find` อ่านรายการ, `First` อ่านแถวเฉพาะ, `Create` เพิ่มข้อมูล, `Updates` แก้ข้อมูล, `Delete` ทำ Soft Delete โดยตรวจ `.Error` ทุก operation ถ้าใช้ `Updates` กับ Struct จะข้ามค่าศูนย์บางชนิด ใน Workshop จึงใช้ Map เฉพาะฟิลด์ที่อนุญาต เพื่ออัปเดต DogID เป็น 0 หรือล้าง email ได้

`Where("dog_id > ? AND dog_id < ?", 50, 100)` ใช้ parameter แยกจากคำสั่ง ไม่ใช่ต่อค่าที่รับจากผู้ใช้เข้า SQL โดยตรง และยังใช้เงื่อนไข Soft Delete ของ GORM ตามปกติ

`Unscoped().Find()` จะเห็นทั้งข้อมูลปกติและถูกลบ จึงต้องใช้ `Unscoped().Where("deleted_at IS NOT NULL").Find(&dogs)` ใน endpoint ที่ต้องการเฉพาะข้อมูลถูกลบ API นี้ไม่ลบหรือ restore ข้อมูล

`GetDogsJson()` Query Dogs ที่ยังไม่ถูกลบ แล้วส่งให้ `buildDogSummary()` วน `for range` หนึ่งรอบ สร้าง `models.DogsRes` พร้อมสีและเพิ่มตัวนับเพียงกลุ่มเดียว:

| DogID | type |
|---|---|
| 10–50 รวมขอบเขต | red |
| 100–150 รวมขอบเขต | green |
| 200–250 รวมขอบเขต | pink |
| ช่วงอื่น | no color |

`sum_red`, `sum_green`, `sum_pink`, `sum_nocolor` คือ **จำนวนแถว Dogs** ในแต่ละสี ไม่ใช่ผลรวม DogID; DogID ซ้ำกันก็เป็นคนละรายการตาม ID และต้องนับทุกแถวหนึ่งครั้ง ผลรวมทั้งสี่จึงเท่ากับ `count` และ `len(data)` สีคำนวณตอนตอบกลับ ไม่ได้เก็บในฐานข้อมูล

## Import Postman Collection และข้อมูล 12 รายการ

1. เปิด Postman กด Import แล้วเลือก `postman/GoFiber_Workshop_7.postman_collection.json`
2. ใน Collection Variables ตั้ง `baseUrl = http://localhost:3000`, `basicUsername = gofiber`, `basicPassword = 21022566`
3. Authorization ของ Collection เป็น Basic Auth Request ปกติจะ Inherit ส่วนทดสอบ auth ผิด/ไม่ส่ง auth ตั้ง override ไว้เฉพาะ Request
4. เปิด Server พร้อม MySQL แล้วรัน folder Dogs CRUD ตามลำดับ POST → GET/Filter/JSON → PUT → DELETE โดย POST เก็บ Primary Key ใน `dogPrimaryID` อัตโนมัติ
5. รัน Company CRUD ตามลำดับ POST → GET → PUT → DELETE ใช้ `companyPrimaryID` อัตโนมัติ หากกด PUT/DELETE เดี่ยว ๆ ต้องสร้างข้อมูลก่อนหรือกำหนดค่า variable ให้ตรงกับ ID ที่มีอยู่
6. เปิด folder Color Summary และรันตามลำดับเพื่อเพิ่มตัวอย่าง 12 รายการ แล้วอ่าน JSON Summary
7. หลังเตรียมข้อมูลแล้วเรียก Range จะเห็น DogID 51 และ 99 จากชุดตัวอย่าง เรียก Deleted Dogs หลัง DELETE เพื่อตรวจ Soft Delete
8. ดู Test Results ใต้แต่ละ response: มีการตรวจ status, JSON, ช่วง DogID, สี และผลรวม

| Name | DogID | Expected type |
|---|---:|---|
| Dog01 | 9 | no color |
| Dog02 | 10 | red |
| Dog03 | 50 | red |
| Dog04 | 51 | no color |
| Dog05 | 99 | no color |
| Dog06 | 100 | green |
| Dog07 | 150 | green |
| Dog08 | 151 | no color |
| Dog09 | 200 | pink |
| Dog10 | 250 | pink |
| Dog11 | 251 | no color |
| Dog12 | 999 | no color |

Request ตัวอย่างแต่ละตัวมี pre-request script ตรวจชื่อและ dog_id ผ่าน `/filter` ก่อน POST และข้ามหากมีรายการที่ยังไม่ถูกลบคู่นั้นแล้ว ต้องใช้ Postman รุ่นที่รองรับ `pm.execution.skipRequest()` การตรวจนี้ช่วยให้รันซ้ำตามลำดับได้ แต่ไม่ใช่ข้อบังคับ unique ระดับ Database และไม่ได้ป้องกันการส่งพร้อมกันหลาย client ไม่ได้เพิ่ม Batch Endpoint หรือ seed ตอนเปิดแอป

หากมีเฉพาะ 12 รายการนี้ที่ยังไม่ถูกลบ ผลคือ count 12, red 2, green 2, pink 2, no color 6 ถ้ามีข้อมูลอื่นอยู่แล้ว ผลรวมต้องนับข้อมูลเหล่านั้นด้วย Collection ตรวจสีและผลรวมจากข้อมูล response จริง โดยไม่บังคับให้ฐานข้อมูลทั้งหมดเหลือเพียง 12 รายการ และไม่มีการลบข้อมูลเดิมให้

Collection มี saved example responses ของ endpoints ใหม่ ใช้เพื่อดูรูปแบบ ไม่ใช่ข้อมูล Hardcode ที่ Server ส่งจริง ส่วน Collection เดิมสำหรับ Exercise 5–6 ยังใช้ตรวจ Factorial/ASCII/Register ได้

## Unit Tests และ MySQL Integration Tests

```powershell
go test ./... -v
```

ค่าเริ่มต้นไม่ต้องมี MySQL: ทดสอบขอบเขตสี, ชุดตัวอย่าง 12 ตัว, DogID ซ้ำ, รายการว่างต้องเป็น `[]`, validation และ Path ID, config ไม่ครบ, API เดิม, Basic Auth และ Database unavailable โดย `app.Test()` ส่ง HTTP ผ่าน Fiber ในหน่วยความจำและไม่เปิดพอร์ตจริง

Integration test ถูก skip จนกว่าจะเปิดเอง หลังตั้ง `.env` หรือ Environment Variables และสร้าง Database แล้ว:

```powershell
$env:MYSQL_INTEGRATION_TEST = '1'
go test . -run '^TestMySQLIntegration$' -count=1 -v
Remove-Item Env:MYSQL_INTEGRATION_TEST
```

Integration test เชื่อม MySQL จริงและเรียก AutoMigrate แล้วสร้างข้อมูลของตัวเองใน Transaction ตรวจ Dogs/Company CRUD, filter, range, summary, Soft Delete, validation และ Database Error จาก connection ที่ปิด จากนั้น Rollback เฉพาะข้อมูลใน Transaction; ไม่มี Drop/Reset ตาราง ข้อมูลทดสอบไม่คงอยู่หลังเสร็จ แต่เลข Auto Increment อาจเลื่อน และการเปลี่ยน schema จาก AutoMigrate ไม่ได้ถูก Rollback จึงควรใช้ Database สำหรับ Workshop/Test

สถานะการตรวจในสภาพแวดล้อมพัฒนา: `go mod tidy`, `go build ./...` และ unit/HTTP tests แบบไม่ใช้ MySQL รันผ่านแล้ว ตรวจ HTTP บน Server จริงพอร์ต 3000 อีก 9 requests ผ่าน โดยจำลองให้เชื่อมฐานข้อมูลไม่ได้เพื่อยืนยัน API เดิม, Basic Auth และ JSON 500 ของ API ใหม่ ต่อมาได้ตั้งค่าบัญชี MySQL, สร้าง Database `golang_test` และรัน `TestMySQLIntegration` ผ่านกับ MySQL จริง การทดสอบนี้ตรวจ AutoMigrate, CRUD, Filter, Range, Soft Delete, สรุปสี, Validation และ Database Error โดย Rollback ข้อมูลตัวอย่างใน Transaction แล้ว นอกจากนี้เปิด Server กับ MySQL จริงและตรวจ HTTP อีก 7 requests ผ่าน รวม GET Dogs/Company, Input ผิด และ Basic Auth ผิด ยังไม่ได้รัน Collection ในแอป Postman จริง แต่ตรวจว่าไฟล์ JSON และ JavaScript ใน Collection parse ได้

## Demo Walkthrough

```text
Postman
   ↓
GoFiber จับคู่เส้นทางในลำดับที่ลงทะเบียน
   ↓
Basic Auth Middleware ของ /api
   ↓
Database guard ของ Dogs/Company
   ↓
Controller: อ่าน Parameter/Body และ Validate
   ↓
GORM → MySQL
   ↓
Controller: จัดรูปแบบ JSON / คำนวณสี
   ↓
JSON Response → Postman
```

ลำดับนำเสนอที่แนะนำ:

1. เปิด `main.go` อธิบาย Route Group และ Alias `c` แล้วทดสอบ auth ผิดได้ 401
2. แสดง Factorial 5 ได้ 120, ASCII ของ Hello และ Register เพื่อยืนยัน Exercise 5–6
3. เปิด `database/database.go` แสดงการอ่าน config และ AutoMigrate โดยไม่เปิดเผยรหัสผ่านใน `.env`
4. POST Dog แล้วชี้ให้เห็นความต่างระหว่าง Primary Key `ID` กับ `dog_id` จากนั้นใช้ Filter, PUT และ DELETE
5. เปิด Deleted Dogs อธิบาย `Unscoped()` กับ `deleted_at IS NOT NULL`
6. ทำ Company CRUD และอธิบายชื่อตาราง `companies` กับ Soft Delete
7. เพิ่ม Dogs ตัวอย่าง 12 รายการ แสดง Range แล้วเปิด `/dog/json` อธิบาย `for range`, ช่วงสี และผลรวมตัวนับ
8. รัน unit tests แสดงค่าขอบเขต 10, 50, 100, 150, 200, 250 และอธิบายว่าส่วนนี้ทดสอบได้โดยไม่เชื่อม MySQL

คำอธิบายสั้นสำหรับพูดต่อ: “Client ส่ง Request ผ่าน Basic Auth แล้ว Controller อ่านข้อมูลเข้า Struct ตรวจความถูกต้องและเรียก GORM ให้ทำงานกับ MySQL สำหรับ JSON Summary จะอ่าน Dogs ที่ยังไม่ถูกลบมาจัดสีและนับแต่ละกลุ่ม ก่อนส่ง Response กลับให้ Postman”

## เอกสารอ้างอิง

- [GORM: Connecting to a Database](https://gorm.io/docs/connecting_to_the_database.html)
- [GORM: Updates และ Zero Values](https://gorm.io/docs/update.html)
- [GORM: Delete และ Soft Delete](https://gorm.io/docs/delete.html)
- [Postman: pm.execution.skipRequest](https://learning.postman.com/docs/tests-and-scripts/write-scripts/postman-sandbox-reference/pm-execution/)
