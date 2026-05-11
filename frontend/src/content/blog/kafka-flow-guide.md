---
title: "Luồng hoạt động Kafka - Minh họa & Ví dụ"
description: "Giải thích chi tiết kiến trúc và luồng hoạt động của Apache Kafka qua các ví dụ thực tế, so sánh trực quan và cấu hình mẫu cho server nhỏ."
date: 2026-05-11
category: "cong-nghe"
tags: ["kafka", "message-queue", "backend", "distributed-systems", "tutorial"]
draft: false
coverImage: "https://images.viblo.asia/35bd305f-4ccd-45d4-b230-a4317e27ac15.png"
---

# Luồng hoạt động Kafka - Minh họa & Ví dụ

## 1. Tổng quan kiến trúc Kafka

```
┌──────────┐       ┌─────────────────────────────────┐       ┌──────────────┐
│ Producer │──────▶│         KAFKA CLUSTER            │◀──────│   Consumer   │
│ (API)    │       │                                   │  pull │   (Worker)   │
└──────────┘       │  ┌───────────┐  ┌───────────┐   │       └──────┬───────┘
                   │  │Partition 0│  │Partition 1│   │              │
                   │  └───────────┘  └───────────┘   │              ▼
                   │  ┌───────────┐                   │       ┌──────────────┐
                   │  │Partition 2│                   │       │   Database   │
                   │  └───────────┘                   │       └──────────────┘
                   └─────────────────────────────────┘
```

## 2. Các thành phần chính

| Thành phần | Vai trò | Ví dụ đời thực |
|---|---|---|
| Producer | Gửi message vào Kafka | Người gửi bưu kiện đến bưu điện |
| Topic | Chủ đề/kênh chứa messages | Tên loại bưu kiện (thư, hàng hoá, EMS) |
| Partition | Ngăn chia nhỏ trong topic | Các ô phân loại trong kho bưu điện |
| Consumer | Kéo message ra xử lý | Nhân viên giao hàng lấy bưu kiện đi giao |
| Consumer Group | Nhóm consumers cùng xử lý 1 topic | Đội giao hàng (mỗi người phụ trách 1 khu vực) |
| Offset | Vị trí đọc hiện tại của consumer | Số thứ tự bưu kiện đã giao đến đâu |
| Broker | Server Kafka lưu trữ data | Kho bưu điện |

## 3. Luồng hoạt động chi tiết

### Bước 1: Producer gửi message

```
[API nhận request]
      │
      ▼
[Producer serialize data thành message]
      │
      ▼
[Gửi vào Topic "orders"]
      │
      ├── Có key (userId=123) → hash(key) % số_partition → vào partition cố định
      │
      └── Không có key → Round-robin → phân bổ đều các partition
```

### Bước 2: Kafka lưu trữ

```
Topic "orders" (3 partitions):

Partition 0:  [offset 0] [offset 1] [offset 2] [offset 3] ...
               msg_A      msg_D      msg_G      msg_J

Partition 1:  [offset 0] [offset 1] [offset 2] [offset 3] ...
               msg_B      msg_E      msg_H      msg_K

Partition 2:  [offset 0] [offset 1] [offset 2] [offset 3] ...
               msg_C      msg_F      msg_I      msg_L
```

- Messages được ghi **tuần tự** vào cuối partition (append-only)
- Mỗi message có 1 **offset** (số thứ tự) duy nhất trong partition
- Data lưu trên **disk**, không mất khi restart

### Bước 3: Consumer pull messages

```
Consumer Group "order-service" (3 consumers):

Consumer A ← được gán Partition 0 → pull [msg_A, msg_D, msg_G...]
Consumer B ← được gán Partition 1 → pull [msg_B, msg_E, msg_H...]
Consumer C ← được gán Partition 2 → pull [msg_C, msg_F, msg_I...]

Mỗi consumer pull batch (ví dụ 500 messages/lần) → INSERT vào DB
```

### Bước 4: Commit offset

```
Consumer A xử lý xong msg_A, msg_D, msg_G
      │
      ▼
Commit offset = 3 (đã đọc đến vị trí 3)
      │
      ▼
Nếu Consumer A crash → restart → đọc lại từ offset 3 (không mất data)
```

## 4. Ví dụ thực tế: Insert 1 triệu đơn hàng

### Kịch bản

- Flash sale 12h trưa
- 1,000,000 đơn hàng đổ về trong 10 giây
- DB PostgreSQL chỉ chịu được 5,000 writes/giây

### Luồng xử lý

```
                    10 giây
[1M đơn hàng] ─────────────────▶ [Kafka Topic: orders]
                                   (3 partitions)
                                        │
                    ┌───────────────────┼───────────────────┐
                    ▼                   ▼                   ▼
             [Partition 0]       [Partition 1]       [Partition 2]
              ~333K msgs          ~333K msgs          ~333K msgs
                    │                   │                   │
                    ▼                   ▼                   ▼
             [Consumer A]        [Consumer B]        [Consumer C]
              pull 500/lần        pull 500/lần        pull 500/lần
                    │                   │                   │
                    ▼                   ▼                   ▼
             [Batch INSERT]      [Batch INSERT]      [Batch INSERT]
              500 rows/lần        500 rows/lần        500 rows/lần
                    │                   │                   │
                    └───────────────────┼───────────────────┘
                                        ▼
                                  [PostgreSQL]
                                 ~1500 writes/s
                                 (500 x 3 consumers)
                                        │
                                        ▼
                              Hoàn thành sau ~11 phút
                              (1M ÷ 1500/s ≈ 667 giây)
```

### Timeline

```
T+0s     : 1M requests đến API
T+2s     : Producer đẩy xong 1M messages vào Kafka (Kafka nhanh!)
T+2s     : API trả response "Đã tiếp nhận" cho tất cả users
T+2s~11m : 3 consumers xử lý song song, INSERT batch vào DB
T+11m    : Hoàn thành 1M records trong DB
```

## 5. Ví dụ đời thực: Nhà máy đóng gói

```
┌─────────────────────────────────────────────────────────────────┐
│                        NHÀ MÁY ĐÓNG GÓI                        │
│                                                                   │
│  [Xe tải chở hàng đến]     = Producer (API nhận request)        │
│         │                                                         │
│         ▼                                                         │
│  [Kho chứa hàng lớn]       = Kafka Topic                        │
│         │                                                         │
│         ├── Kệ A (hàng điện tử)  = Partition 0                  │
│         ├── Kệ B (hàng thực phẩm) = Partition 1                 │
│         └── Kệ C (hàng gia dụng)  = Partition 2                 │
│                │                                                  │
│                ▼                                                  │
│  [3 dây chuyền đóng gói]   = 3 Consumers                        │
│         │                                                         │
│         ▼                                                         │
│  [Xe tải giao hàng]        = INSERT vào DB                      │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘

Quy tắc:
- Xe tải đến bất kỳ lúc nào → kho luôn nhận (Kafka throughput cao)
- Mỗi kệ chỉ 1 dây chuyền xử lý (1 partition = 1 consumer)
- Dây chuyền lấy hàng theo tốc độ của mình (pull-based)
- Nếu 1 dây chuyền hỏng → hàng vẫn nằm trên kệ, không mất
```

## 6. So sánh: Có Kafka vs Không Kafka

```
KHÔNG CÓ KAFKA:
═══════════════
[1M requests] ═══════════════════════▶ [DB] 💥 CRASH!
               đồng thời, không kiểm soát


CÓ KAFKA:
═════════
[1M requests] ──▶ [Kafka] ──▶ [Consumer] ──batch──▶ [DB] ✅ OK
               2s    giữ lại    pull từ từ   500/lần    ổn định
```

## 7. Các tình huống đặc biệt

### Consumer crash giữa chừng

```
Consumer A đang xử lý Partition 0:
  - Đã xử lý: offset 0 → 999 (commit offset = 1000)
  - Đang xử lý: offset 1000 → 1499
  - CRASH! 💥

Kafka phát hiện Consumer A chết:
  - Gán Partition 0 cho Consumer B (rebalance)
  - Consumer B đọc từ offset 1000 (chỗ commit cuối)
  - Không mất data!
```

### Thêm consumer khi tải cao

```
Ban đầu: 3 partitions, 3 consumers → mỗi consumer 1 partition

Thêm Consumer D:
  - 3 partitions, 4 consumers
  - Consumer D KHÔNG được gán partition nào (ngồi chờ dự phòng)
  - Vì: số consumer > số partition → consumer thừa

Muốn 4 consumers hoạt động → phải tăng lên 4 partitions
```

### Message bị lỗi khi INSERT

```
Consumer pull batch 500 messages:
  - 498 messages INSERT thành công
  - 2 messages bị lỗi (duplicate key, invalid data)

Xử lý:
  - Gửi 2 messages lỗi vào "dead-letter-topic" (topic chứa lỗi)
  - Commit offset (tiếp tục xử lý batch tiếp)
  - Sau đó review dead-letter-topic để fix thủ công
```

## 8. Cấu hình mẫu cho server 2vCPU / 4GB RAM

```yaml
# Kafka topic config
topic:
  name: orders
  partitions: 3          # 3 đường song song
  replication-factor: 1  # Server đơn, không replicate

# Consumer config
consumer:
  group-id: order-service
  max-poll-records: 500       # Pull tối đa 500 messages/lần
  auto-offset-reset: earliest # Đọc từ đầu nếu chưa có offset
  enable-auto-commit: false   # Commit thủ công sau khi INSERT thành công

# Producer config  
producer:
  batch-size: 16384      # Gom messages trước khi gửi (16KB)
  linger-ms: 5           # Chờ 5ms để gom thêm messages
  acks: 1                # Chờ leader xác nhận (cân bằng tốc độ/an toàn)
```

## 9. Tóm tắt 1 câu

```
Producer đẩy nhanh → Kafka giữ an toàn trên disk theo partitions
→ Consumers tự pull theo tốc độ của mình → INSERT batch vào DB
→ DB không bao giờ bị quá tải, data không bao giờ mất.
```
