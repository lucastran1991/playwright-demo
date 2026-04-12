# Giải Thích Chi Tiết: Hệ Thống Tính Toán Load Capacity

## Tổng Quan

Hệ thống Load Capacity thực hiện quy trình **bottom-up aggregation** (tổng hợp từ dưới lên) để tính toán công suất tải và mức sử dụng cho toàn bộ hạ tầng data center. Quy trình gồm 3 giai đoạn chính:

1. **Nhập dữ liệu CSV** → bảng `node_variables` (source: `csv_import`)
2. **Tính toán tổng hợp** → bảng `node_variables` (source: `computed`)
3. **Truy vấn API** → cung cấp dữ liệu cho frontend DAG

---

## Cấu Trúc Database

### Bảng `node_variables` (Bảng chính)

Lưu trữ các chỉ số capacity dạng key-value cho từng node.

| Cột | Kiểu | Mô tả |
|-----|------|-------|
| `node_id` | string | ID của node trong blueprint (VD: `RACK-01-A`) |
| `variable_name` | string | Tên biến (VD: `rated_capacity`, `allocated_load`) |
| `value` | float64 | Giá trị số (đơn vị kW) |
| `unit` | string | Đơn vị đo (mặc định: `kW`) |
| `source` | string | Nguồn gốc: `csv_import` hoặc `computed` |

**Ràng buộc duy nhất:** `(node_id, variable_name)` — mỗi node chỉ có 1 giá trị cho mỗi biến.

### Bảng `capacity_node_types` (Phân loại node)

| Cột | Mô tả |
|-----|-------|
| `node_type` | Loại node (VD: `Rack`, `RPP`, `UPS`) |
| `topology` | Thuộc domain nào (Electrical, Cooling, Spatial, Whitespace) |
| `is_capacity_node` | Có phải node capacity không (ảnh hưởng hiển thị Load trace) |
| `active_constraint` | Có phải ràng buộc hoạt động không |

### Bảng `blueprint_edges` (Quan hệ cha-con)

Lưu quan hệ phân cấp giữa các node theo từng topology. Được sử dụng bởi truy vấn đệ quy để tìm tất cả Rack con của một node cha.

---

## Giai Đoạn 1: Nhập Dữ Liệu CSV

### File nguồn
`blueprint/ISET capacity - rack load flow.csv` — chứa 35+ cột dữ liệu cho ~657 node.

### Quy trình chuyển đổi cột CSV → biến chuẩn

Mỗi loại node có bản đồ chuyển đổi riêng từ tên cột CSV sang tên biến chuẩn hóa:

**Rack:**
| Cột CSV | Biến chuẩn | Đơn vị |
|---------|-----------|--------|
| `Rack_Circuit_Capacity_(design)` | `design_capacity` | kW |
| `Rack_Circuit_Capacity_(rated)` | `rated_capacity` | kW |
| `Allocated_ITLoad` | `allocated_load` | kW |
| `Allocated_LiquidCool_Load` | `allocated_liquid_load` | kW |
| `Allocated_AirCool_Load` | `allocated_air_load` | kW |

**RPP (Remote Power Panel):**
| Cột CSV | Biến chuẩn |
|---------|-----------|
| `RPP_Panel_Capacity_(rated)` | `rated_capacity` |
| `Allocated_ITLoad` | `allocated_load` |

**Air Zone / Air Cooling Unit:**
| Cột CSV | Biến chuẩn |
|---------|-----------|
| `AirZone_Cooling_Capacity` | `design_capacity` |
| `Rated_Cooling_Capacity` | `rated_capacity` |

*(Tương tự cho UPS, Room PDU, CDU, RDHx, DTC, Liquid Loop, v.v.)*

### Bước xử lý

1. Đọc file CSV, xây dựng bản đồ header → index cột
2. Với mỗi dòng dữ liệu: trích xuất `node_id`, `node_type`, `name`
3. Tra cứu bản đồ chuyển đổi theo `node_type`
4. Parse giá trị số, bỏ qua nếu rỗng
5. **Upsert vào bảng `node_variables`** với `source = "csv_import"`
   - Dùng `ON CONFLICT (node_id, variable_name) DO UPDATE` để cập nhật nếu đã tồn tại

---

## Giai Đoạn 2: Tính Toán Bottom-Up Aggregation

Đây là phần cốt lõi. Sau khi nhập CSV, hệ thống tự động kích hoạt `ComputeAll()`.

### Bước 1: Xóa dữ liệu computed cũ

```
DELETE FROM node_variables WHERE source = 'computed'
```

### Bước 2: Load dữ liệu CSV vào bộ nhớ

Tạo bản đồ: `map[nodeID] → map[variableName] → float64` từ tất cả record có `source = "csv_import"`.

### Bước 3: Tính chỉ số cấp Rack (đơn vị gốc)

Với mỗi Rack có dữ liệu từ CSV:

```
available_capacity = rated_capacity - allocated_load
utilization_pct    = (allocated_load / rated_capacity) × 100
```

**Ví dụ:**
- Rack `RACK-01-A`: rated_capacity = 10 kW, allocated_load = 7 kW
- → available_capacity = 3 kW
- → utilization_pct = 70%

Kết quả được lưu vào `node_variables` với `source = "computed"`.

### Bước 4: Tổng hợp cho các node cấp cao hơn

Hệ thống xử lý tuần tự 13 loại node, mỗi loại có cấu hình riêng:

#### Chuỗi Điện (Power Chain)
| Loại Node | Biến tải (từ Rack) | Biến capacity (từ CSV) |
|-----------|-------------------|----------------------|
| RPP | `allocated_load` | `rated_capacity` |
| Room PDU | `allocated_load` | `rated_capacity` |
| UPS | `allocated_load` | `rated_capacity` |

#### Chuỗi Làm Mát Khí (Air Cooling)
| Loại Node | Biến tải (từ Rack) | Biến capacity |
|-----------|-------------------|--------------|
| Air Zone | `allocated_air_load` | `design_capacity` |
| Air Cooling Unit | `allocated_air_load` | `rated_capacity` |

#### Chuỗi Làm Mát Lỏng (Liquid Cooling)
| Loại Node | Biến tải (từ Rack) | Biến capacity |
|-----------|-------------------|--------------|
| Liquid Loop | `allocated_liquid_load` | `design_capacity` |
| CDU | `allocated_liquid_load` | `rated_capacity` |
| RDHx | `allocated_liquid_load` | `rated_capacity` |
| DTC | `allocated_liquid_load` | `rated_capacity` |

#### Nhóm Không Gian (Spatial/Whitespace)
| Loại Node | Biến tải | Biến capacity |
|-----------|---------|--------------|
| Capacity Cell | `allocated_load` | `design_capacity` |
| Room Bundle | `allocated_load` | `design_capacity` |
| UPS Bundle | `allocated_load` | `design_capacity` |
| Row | `allocated_load` | *(không có)* |

### Thuật toán tổng hợp cho mỗi node

```
1. Tìm tất cả node thuộc loại đang xử lý (VD: tất cả RPP)
2. Với mỗi node:
   a. Tìm tất cả Rack con qua quan hệ spatial (đệ quy, tối đa 5 cấp)
   b. Tổng hợp: totalLoad = SUM(descendantRack[loadVariable])
   c. Lưu: node.allocated_load = totalLoad [source: computed]
   d. Nếu node có capacity riêng (HasOwnCap = true):
      - available_capacity = capacity - totalLoad
      - utilization_pct = (totalLoad / capacity) × 100
      - Lưu cả 2 giá trị [source: computed]
```

### Truy vấn đệ quy tìm Rack con

Sử dụng `WITH RECURSIVE` trong PostgreSQL để duyệt cây spatial topology:

```sql
WITH RECURSIVE descendants AS (
    -- Cấp 1: con trực tiếp
    SELECT bn.id, bn.node_id, bn.node_type, 1 as level
    FROM blueprint_edges be
    JOIN blueprint_nodes bn ON bn.id = be.to_node_id
    JOIN blueprint_types bt ON bt.id = be.blueprint_type_id
    WHERE be.from_node_id IN (nodeIDs)
      AND bt.slug = 'spatial-topology'

    UNION ALL

    -- Đệ quy: con của con, tối đa 5 cấp
    SELECT bn.id, bn.node_id, bn.node_type, d.level + 1
    FROM descendants d
    JOIN blueprint_edges be ON be.from_node_id = d.id
    JOIN blueprint_nodes bn ON bn.id = be.to_node_id
    JOIN blueprint_types bt ON bt.id = be.blueprint_type_id
    WHERE bt.slug = 'spatial-topology' AND d.level < 5
)
SELECT DISTINCT ON (id) * FROM descendants
WHERE node_type IN ('Rack')
```

---

## Ví Dụ Cụ Thể

### Tình huống: Tính utilization cho RPP-01

```
RPP-01 (rated_capacity = 100 kW từ CSV)
├── Row-01
│   ├── RACK-01-A (allocated_load = 7 kW)
│   ├── RACK-01-B (allocated_load = 5 kW)
│   └── RACK-01-C (allocated_load = 8 kW)
└── Row-02
    ├── RACK-02-A (allocated_load = 6 kW)
    └── RACK-02-B (allocated_load = 4 kW)
```

**Bước tính:**
1. Tìm tất cả Rack con của RPP-01 qua spatial topology → 5 Rack
2. `totalLoad = 7 + 5 + 8 + 6 + 4 = 30 kW`
3. Lưu `RPP-01.allocated_load = 30` (source: computed)
4. `available_capacity = 100 - 30 = 70 kW` (source: computed)
5. `utilization_pct = (30 / 100) × 100 = 30%` (source: computed)

### Tình huống: Tính cho Air Zone

```
AirZone-01 (design_capacity = 200 kW từ CSV)
├── RACK-01-A (allocated_air_load = 5 kW)  ← Chú ý: dùng allocated_AIR_load
├── RACK-01-B (allocated_air_load = 3 kW)
└── RACK-02-A (allocated_air_load = 4 kW)
```

**Kết quả:**
- `totalLoad = 5 + 3 + 4 = 12 kW` (sum của `allocated_air_load`, không phải `allocated_load`)
- `utilization_pct = (12 / 200) × 100 = 6%`

**Điểm quan trọng:** Mỗi chuỗi topology dùng biến tải khác nhau:
- Điện: `allocated_load` (tổng tải IT)
- Làm mát khí: `allocated_air_load` (chỉ phần khí)
- Làm mát lỏng: `allocated_liquid_load` (chỉ phần lỏng)

---

## Giai Đoạn 3: API Truy Vấn

### Endpoints

| Method | Path | Mô tả |
|--------|------|-------|
| `POST` | `/api/capacity/ingest` | Kích hoạt nhập CSV + tính toán |
| `GET` | `/api/capacity/nodes/:nodeId` | Lấy chỉ số cho 1 node |
| `GET` | `/api/capacity/nodes?type=Rack&min_utilization=80` | Danh sách phân trang |
| `GET` | `/api/capacity/summary` | Thống kê tổng hợp |

### Tích hợp với Dependency Trace

Khi gọi `GET /api/trace/full/:nodeId`, hệ thống:
1. Tính dependency + impact trace
2. Thu thập tất cả `nodeID` trong kết quả trace
3. Batch load capacity data: `GetCapacityMapForNodes([allNodeIDs])`
4. Gắn vào response: `Capacity: {nodeID → {varName → value}}`

Frontend DAG hiển thị thanh utilization trên mỗi node dựa trên dữ liệu này.

---

## Luồng Dữ Liệu Tổng Thể

```
CSV File (35 cột, ~657 node)
    │
    ▼
[Parse CSV] ── chuyển đổi cột → biến chuẩn theo loại node
    │
    ▼
[Upsert] ── node_variables (source: csv_import)
    │
    ▼
[Xóa computed cũ] ── DELETE WHERE source = 'computed'
    │
    ▼
[Tính Rack] ── available = rated - allocated, util% = allocated/rated × 100
    │
    ▼
[Tổng hợp 13 loại node] ── Với mỗi loại:
    │   ├── Tìm Rack con (đệ quy spatial, max 5 cấp)
    │   ├── SUM(loadVar) từ Rack con
    │   └── Tính available + utilization%
    │
    ▼
[Upsert computed] ── node_variables (source: computed)
    │
    ▼
[API] ── Phục vụ frontend DAG + capacity endpoints
```

---

## Mã Màu Utilization (Frontend)

| Mức sử dụng | Màu | Ý nghĩa |
|-------------|-----|---------|
| < 60% | Xanh lá `#22C55E` | Bình thường |
| 60% - 80% | Vàng `#EAB308` | Cảnh báo |
| > 80% | Đỏ `#EF4444` | Quá tải |

---

## Ghi Chú Kỹ Thuật

1. **Sparse data model**: Dùng key-value thay vì 35 cột vì hầu hết node chỉ dùng 3-5 biến
2. **Idempotent**: Có thể chạy lại ingestion bất kỳ lúc nào nhờ `ON CONFLICT DO UPDATE`
3. **Row không có capacity**: Loại node `Row` chỉ theo dõi tải, không có giới hạn capacity (HasOwnCap = false)
4. **Batch processing**: Upsert theo batch 100 record để tối ưu hiệu năng DB
5. **Spatial topology**: Mọi tổng hợp đều dựa trên quan hệ `spatial-topology`, không phải electrical hay cooling topology
