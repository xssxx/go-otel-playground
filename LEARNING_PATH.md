# แผนการเรียน OpenTelemetry — ทีละก้าว

---

### ✅ Step 0 — ที่ทำไปแล้ว
**Trace endpoint เดียว**
- HTTP handler สร้าง span
- ส่ง trace ไป Jaeger ผ่าน OTLP/gRPC
- ดู trace ใน Jaeger UI ได้

---

### Step 1 — ทำให้ Span มีความหมายขึ้น
**เป้าหมาย:** รู้ว่าใส่อะไรใน span ได้บ้าง

สิ่งที่ทำ:
- เพิ่ม **attributes** บน span เช่น `user.id`, `dice.sides`, `http.route`
- เพิ่ม **events** บน span (จุดเวลาสำคัญในชีวิต request)
- ทำให้ span แสดง **error** เมื่อเกิดข้อผิดพลาด (`span.RecordError`, `span.SetStatus`)

ทำไมต้องทำก่อน: ถ้าข้ามไป multi-span โดยที่ span แต่ละอันว่างเปล่า trace ก็ไม่มีประโยชน์

---

### Step 2 — หลาย Span ใน Service เดียว (Child Spans)
**เป้าหมาย:** เห็น trace tree แบบ parent → child ใน Jaeger

สิ่งที่ทำ:
- แยก logic ออกเป็นฟังก์ชัน แล้วให้แต่ละฟังก์ชันสร้าง span ของตัวเอง
- เช่น `rolldice` → เรียก `validateInput()` → เรียก `computeRoll()` → เรียก `saveResult()`
- ส่ง `ctx` ต่อกันทุกฟังก์ชัน (นี่คือกุญแจสำคัญ)

```
[HTTP span]
  └── [validateInput span]
  └── [computeRoll span]
  └── [saveResult span]
```

ทำไมต้องทำก่อน: เข้าใจว่า context propagation ทำงานยังไง ก่อนที่จะข้าม process boundary

---

### Step 3 — เพิ่ม Metrics และ Logs ให้ครบ
**เป้าหมาย:** เห็น 3 signals พร้อมกัน (Traces + Metrics + Logs)

สิ่งที่ทำ:
- แก้ `otel.go` ให้ register `meterProvider` จริงๆ (ตอนนี้มันสร้างแต่ไม่ได้ใช้)
- เพิ่ม metric ประเภทอื่น: `Histogram` วัด latency, `Gauge` วัด active connections
- ดู log + trace ใน terminal และดู metric ผ่าน Prometheus/Grafana

ทำไมต้องทำก่อน: distributed system จริงต้องใช้ทั้ง 3 signals ถ้าไม่คุ้นก็จะสับสนทีหลัง

---

### Step 4 — สอง Service คุยกัน (Context Propagation ข้าม Process)
**เป้าหมาย:** trace เดียวกันวิ่งข้าม 2 service

สิ่งที่ทำ:
- สร้าง Service B (อีก Go binary หรืออีก port)
- Service A (`rolldice`) เรียก Service B ผ่าน HTTP โดยใช้ `otelhttp.DefaultClient` หรือ inject header เอง
- ใน Jaeger จะเห็น trace เดียวที่มี span จากทั้ง 2 service

```
[Service A: HTTP span]
  └── [Service A: call-service-b span]
        └── [Service B: HTTP span]   ← คนละ process แต่ trace เดียวกัน
```

สิ่งที่ต้องเข้าใจ: W3C `traceparent` header คืออะไร, `Propagator` inject/extract ทำงานยังไง

---

### Step 5 — ระบบจริง: หลาย Service + Message Queue
**เป้าหมาย:** trace ผ่าน async boundary

สิ่งที่ทำ:
- เพิ่ม Kafka หรือ RabbitMQ ระหว่าง service
- inject trace context ลงใน message header
- extract context ฝั่ง consumer แล้วสร้าง span ต่อ
- เห็น trace ที่มี "gap" ของเวลาระหว่าง produce → consume

ทำไมยากขึ้น: async ไม่มี request-response ชัดเจน ต้องจัดการ context ด้วยตัวเอง

---

### Step 6 — Observability Stack จริง
**เป้าหมาย:** เลิกใช้ Jaeger all-in-one, ใช้ stack ที่ production-ready

สิ่งที่ทำ:
- เพิ่ม **OpenTelemetry Collector** เป็น middle layer (รับ signal แล้ว route ไปหลายที่)
- Traces → Tempo หรือ Jaeger
- Metrics → Prometheus → Grafana
- Logs → Loki → Grafana
- ทำ **exemplars**: link จาก metric graph ไปถึง trace ที่ทำให้เกิด spike

---

### สรุป Roadmap

```
Step 0  ✅  Trace endpoint เดียว
Step 1      ทำ span ให้มีข้อมูล (attributes, events, errors)
Step 2      หลาย span ใน service เดียว (child spans)
Step 3      Metrics + Logs ให้ครบ 3 signals
Step 4      2 service คุยกันผ่าน HTTP (context propagation)
Step 5      Async / Message Queue
Step 6      Production stack (OTel Collector + Grafana)
```

**ตอนนี้ควรทำ Step 1 ก่อน** — ใส่ข้อมูลให้ span แล้วลองดูใน Jaeger ว่ามันเปลี่ยนยังไง จะทำให้ Step 2-4 มีความหมายขึ้นมาก
