# Giải Thích Chi Tiết: OAK Value Calculator

## Mục Lục

1. [Khái Niệm Kinh Doanh](#1-khái-niệm-kinh-doanh)
2. [Thuật Ngữ](#2-thuật-ngữ)
3. [Tổng Quan 10 Bước](#3-tổng-quan-10-bước)
4. [Phase 1: Quality-Adjusted Row Value (Steps 1-3)](#4-phase-1-quality-adjusted-row-value)
5. [Phase 2: Capacity Cell → Rack Distribution (Steps 4-5)](#5-phase-2-capacity-cell--rack-distribution)
6. [Phase 3: SBO Premium — Scenario Dependent (Steps 6-8)](#6-phase-3-sbo-premium--scenario-dependent)
7. [Phase 4: Portfolio Summary (Steps 9-10)](#7-phase-4-portfolio-summary)
8. [Ví Dụ Tính Toán End-to-End](#8-ví-dụ-tính-toán-end-to-end)
9. [Luồng Dữ Liệu](#9-luồng-dữ-liệu)
10. [Phụ Lục: Bảng Hằng Số](#10-phụ-lục-bảng-hằng-số)

---

## 1. Khái Niệm Kinh Doanh

### Vấn đề

Tài nguyên khan hiếm nhất trong data center là **năng lượng (kW)**. Câu hỏi cốt lõi: **Làm sao biến mỗi kilowatt thành dollar cao nhất?**

Giá trị của 1 kW **không đồng đều** — nó phụ thuộc vào:

- **Chất lượng hạ tầng** xung quanh nó (mật độ, độ dự phòng, phương thức làm mát)
- **Tính cách ly** (tenant isolation — không bị ảnh hưởng bởi hàng xóm)
- **Tính tùy chọn** (optionality — có bao nhiêu cách bán được trong tương lai?)

### Tương đồng: Bất động sản thương mại

Tòa nhà thương mại bán theo **diện tích (m²)**. Cùng một tòa nhà, cùng tổng m²:

- Nếu bố trí tenant **đúng** → mỗi suite có ranh giới riêng → tenant tiếp theo vẫn được suite riêng → **giá thuê cao**
- Nếu bố trí tenant **sai** → phá vỡ ranh giới → tenant mới không có không gian riêng → **giá thuê thấp**

Data center tương tự, nhưng "ranh giới" không chỉ là vật lý mà còn là **điện** và **nhiệt**:

| Bất động sản | Data center |
|---|---|
| Diện tích (m²) | Kilowatt (kW) |
| Suite riêng | Capacity cell riêng |
| Tường ngăn | Cách ly điện (RPP/PDU) + nhiệt (air zone/liquid loop) |
| Quyền chọn bố trí | Supply Block Options (SBO) |

### OAK = Optionality Adjusted Kilowatt Value

**OAK** là giá trị dollar quy đổi cho mỗi kW, đã tính đến:

1. **Quality premium**: kW ở hạ tầng chất lượng cao hơn → giá trị cao hơn
2. **Coherence factor**: capacity cell đồng nhất nội bộ → giá trị cao hơn
3. **SBO premium**: rack thuộc nhiều SBO → nhiều cách bán → giá trị cao hơn
4. **Reservation impact**: đặt trước phá vỡ cấu trúc → premium bị mất

---

## 2. Thuật Ngữ

| Thuật ngữ | Viết tắt | Mô tả |
|---|---|---|
| Optionality Adjusted Kilowatt | OAK | Giá trị dollar/kW đã điều chỉnh theo tính tùy chọn |
| Supply Block Option | SBO | Một đơn vị bán hàng khả thi (capacity cell, room bundle, v.v.) |
| Standard Premium Rate | SPR | Phí premium cơ bản khi rack thuộc một SBO |
| Adjusted Premium Rate | APR | SPR đã điều chỉnh bởi composite boost + structural discount |
| Composite Boost | — | Thưởng thêm khi capacity cell thuộc nhiều SBO composite |
| Capacity Cell Envelope | — | Tổng kW thực sự có thể commit trong capacity cell (nhỏ hơn tổng rack circuit) |
| Rack Circuit Capacity | — | Công suất mạch điện tối đa của rack |
| Quality Premium | QP | Hệ số nhân dựa trên 4 chiều chất lượng |
| Coherence Factor | CF | Hệ số chiết khấu dựa trên mức đồng nhất nội bộ capacity cell |
| Whitespace Scenario | — | Scenario gốc: mọi thứ đều free, không có reservation |
| Plan/Draft Scenario | — | Scenario với reservation cụ thể |

---

## 3. Tổng Quan 10 Bước

```
╔═══════════════════════════════════════════════════════════════════════╗
║                     NON-SCENARIO DEPENDENT                          ║
║                     (chỉ dựa trên whitespace blueprint)             ║
║                                                                     ║
║  Step 1: Phân bổ kW từ capacity cell → rows                        ║
║  Step 2: Xác định quality profile per row (4 chiều)                ║
║  Step 3: Tính quality-adjusted row value ($)                       ║
║  Step 4: Tính capacity cell quality value (+ coherence factor)     ║
║  Step 5: Phân phối về từng rack                                    ║
║                                                                     ║
╠═══════════════════════════════════════════════════════════════════════╣
║                     SCENARIO DEPENDENT                              ║
║                     (phụ thuộc reservation status)                  ║
║                                                                     ║
║  Step 6: SPR — standard premium từ SBO membership                  ║
║  Step 7: APR — adjusted premium (composite boost + discounts)      ║
║  Step 8: Tính final Rack OAK value ($)                             ║
║  Step 9: Remaining Portfolio Value (tổng free rack OAK)            ║
║  Step 10: Captured Capacity Value (tổng reserved rack OAK)         ║
║                                                                     ║
╚═══════════════════════════════════════════════════════════════════════╝
```

---

## 4. Phase 1: Quality-Adjusted Row Value

### Step 1 — Phân bổ kW từ Capacity Cell → Row

**Vấn đề**: Tổng rack circuit capacity trong 1 capacity cell thường **lớn hơn** capacity cell envelope. Vì capacity cell bị ràng buộc bởi RPP, air zone, v.v.

**Không đúng**: Cộng tất cả rack circuit capacities lại → con số quá lớn.

**Đúng**: Lấy capacity cell envelope → phân bổ tỷ lệ cho từng row.

```
row_allocated_kw = capacity_cell_envelope × (row_total_rack_circuit / cell_total_rack_circuit)
```

**Ví dụ:**

```
Capacity Cell CC-01 (envelope = 500 kW)
├── Row-A: tổng rack circuit = 200 kW
├── Row-B: tổng rack circuit = 150 kW
└── Row-C: tổng rack circuit = 250 kW
     Tổng rack circuit = 600 kW (lớn hơn envelope!)

Row-A allocated = 500 × (200 / 600) = 166.67 kW
Row-B allocated = 500 × (150 / 600) = 125.00 kW
Row-C allocated = 500 × (250 / 600) = 208.33 kW
                               Tổng  = 500.00 kW ✓
```

**Tại sao Row?** Row là đơn vị nhỏ nhất có thuộc tính chất lượng **tương đối đồng nhất**. Hiếm khi 1 row vừa có liquid cooling vừa có air cooling, hoặc nửa row 60kW nửa row 10kW.

---

### Step 2 — Quality Profile per Row

Mỗi row được đánh giá theo **4 chiều chất lượng**:

#### 2a. Density Bracket (Trọng số: 35%)

Dựa trên **trung bình rack circuit capacity (design)** trong row.

| Tier | Rack Circuit Capacity | Premium |
|---|---|---|
| 1 | <= 15 kW | 1.00 (base) |
| 2 | 15 – 30 kW | 1.05 |
| 3 | 30 – 60 kW | 1.12 |
| 4 | 60 – 120 kW | 1.22 |
| 5 | >= 120 kW | 1.30 |

> Nếu row có rack với density khác nhau → lấy **trung bình** rồi xếp bracket.

#### 2b. Power Redundancy Tier (Trọng số: 20%)

| Cấu hình | Premium | Mô tả |
|---|---|---|
| N (single feed) | 1.00 | 1 đường điện, không dự phòng |
| N+1 | 1.03 | 1 đường + thiết bị upstream dự phòng |
| A/B feeds | 1.08 | 2 đường điện độc lập từ PDU/RPP riêng |
| 2N architecture | 1.14 | 2 hệ thống điện hoàn chỉnh, mỗi hệ chịu 100% tải |

#### 2c. Cooling Redundancy Tier (Trọng số: 15%)

| Cấu hình | Premium | Mô tả |
|---|---|---|
| N | 1.00 | Capacity = expected load, không dự phòng |
| N+1 | 1.03 | +1 unit cooling dự phòng |
| N+2 | 1.08 | +2 units cooling dự phòng |
| N+N (dual loop) | 1.14 | 2 vòng cooling chia tải |

#### 2d. Cooling Mode (Trọng số: 30%)

| Chế độ | Premium |
|---|---|
| Air Only | 1.00 |
| Air + RDHx | 1.03 |
| Air + DTC | 1.07 |
| Full Liquid DTC | 1.07 |

> **Custom Factor**: Có thể thêm factor tùy chỉnh ngoài 4 chiều trên.

---

### Step 3 — Quality-Adjusted Row Value

**Tính quality premium (QP)** = bình quân gia quyền của 4 premium:

```
QP = (density_premium × 0.35) + (power_redundancy_premium × 0.20)
   + (cooling_redundancy_premium × 0.15) + (cooling_mode_premium × 0.30)
```

**Tính quality-adjusted row value:**

```
row_quality_value = row_allocated_kw × BASE_KW_VALUE × QP
```

Trong đó `BASE_KW_VALUE = $180/kW` (giá trị cơ bản cho 1 kW không có premium).

**Ví dụ cho Row-A:**

```
Row-A: avg density = 45 kW → bracket High → 1.12
       power redundancy = 2N → 1.14
       cooling redundancy = N+1 → 1.02
       cooling mode = full liquid → 1.07

QP = (1.12 × 0.35) + (1.14 × 0.20) + (1.02 × 0.15) + (1.07 × 0.30)
   = 0.392 + 0.228 + 0.153 + 0.321
   = 1.094

row_quality_value = 166.67 × 180 × 1.094 = $32,819.96
```

---

## 5. Phase 2: Capacity Cell → Rack Distribution

### Step 4 — Capacity Cell Quality Value (+ Coherence Factor)

**Coherence Factor (CF):** Đo mức đồng nhất giữa các row trong 1 capacity cell theo 4 chiều chất lượng.

**Logic:** Capacity cell đồng nhất hấp dẫn hơn cho tenant. Nếu lẫn lộn → kém hấp dẫn → giảm giá trị.

#### Attribute Coherence Factor (per dimension)

| Phân bố tier giữa các row | Attribute CF |
|---|---|
| Tất cả row cùng tier | **1.00** |
| 2 tier khác nhau tồn tại | **0.98** |
| 3+ tier khác nhau tồn tại | **0.92** |

#### Tính CC Coherence Factor

```
CC_CF = Σ(Attribute_CF × Attribute_Weight)
      = (CF_density × 0.35) + (CF_power × 0.20) + (CF_cooling_red × 0.15) + (CF_cooling_mode × 0.30)
```

**Ví dụ:**

```
CC-01 có 3 rows:
  Row-A: density=Tier3(30-60), power=2N, cooling_red=N+1, cooling_mode=Full Liquid
  Row-B: density=Tier3(30-60), power=2N, cooling_red=N+1, cooling_mode=Full Liquid
  Row-C: density=Tier2(15-30), power=N+1, cooling_red=N+1, cooling_mode=Air Only

CF_density     = 0.98  (2 tiers: Tier3 + Tier2)
CF_power       = 0.98  (2 tiers: 2N + N+1)
CF_cooling_red = 1.00  (1 tier: N+1 — tất cả giống)
CF_cooling_mode = 0.98 (2 tiers: Full Liquid + Air Only)

CC_CF = (0.98 × 0.35) + (0.98 × 0.20) + (1.00 × 0.15) + (0.98 × 0.30)
      = 0.343 + 0.196 + 0.150 + 0.294
      = 0.983
```

**Capacity cell quality value:**

```
cell_quality_value = SUM(row_quality_values) × CF
```

---

### Step 5 — Phân phối về Rack

Phân phối capacity cell quality value về từng rack theo tỷ lệ rack circuit capacity:

```
rack_quality_value = cell_quality_value × (rack_circuit_capacity / cell_total_rack_circuit)
```

**Ví dụ:**

```
CC-01 cell_quality_value = $50,000 (sau coherence discount)
Total rack circuit = 600 kW

Rack-01 (circuit = 30 kW): 50,000 × (30/600) = $2,500
Rack-02 (circuit = 60 kW): 50,000 × (60/600) = $5,000
...
```

**Đến đây hoàn thành Steps 1-5 (non-scenario dependent)**. Mỗi rack có một `rack_quality_value ($)`.

---

## 6. Phase 3: SBO Premium — Scenario Dependent

> **Quan trọng:** Từ step 6 trở đi, kết quả **thay đổi theo scenario** vì reservation status khác nhau.

### Bối cảnh SBO

**SBO (Supply Block Option)** = một đơn vị bán hàng khả thi. Một rack có thể thuộc nhiều SBO:

```
Rack-01 thuộc:
├── Capacity Cell CC-01        (base SBO)
├── Room Bundle RB-01          (base SBO)
├── Composite: CC-01 + CC-02   (composite SBO)
└── UPS Bundle UB-01           (base SBO)
```

Càng thuộc nhiều SBO → càng nhiều cách bán → **giá trị cao hơn**.

### Step 6 — Standard Premium Rate (SPR)

Mỗi "gia đình" SBO có SPR riêng (tính bằng %):

| SBO Family | SPR |
|---|---|
| Capacity Cell | **5%** |
| Room Bundle | **2%** |
| Room PDU Bundle | **3%** |
| UPS Bundle | **3%** |

**Điều kiện SPR = 0:** Nếu base SBO bị **partially reserved** → premium hoàn toàn biến mất (APR = 0).

### Step 7 — Adjusted Premium Rate (APR)

```
APR = SPR × Entanglement_Discount × Composite_Boost × RST_Effect
```

#### 7a. Composite Boost (thưởng)

SBO Builder output `sbo_association_summary` chứa `total_sbo_count` per base SBO. Count này = self(1) + composite_sbo_count.

**Compound Membership Premium (CMP)**: Dựa trên total_sbo_count. Rack thuộc nhiều SBO trong cùng family → CMP cao hơn.

**Reservation ảnh hưởng:** Nếu composite SBO partially reserved → đóng góp 0 vào count. Nếu free hoặc fully reserved → đóng góp 1.

#### 7b. Entanglement Discount (phạt)

| SBO Type | Bị phạt khi | Được thưởng khi |
|---|---|---|
| Capacity Cell | CC span >1 Room (cross-room) | >1 CC trong cùng Room (in-room modularity) |
| Room Bundle | RB span >1 Room | — |
| Room PDU Bundle | RPDU span >1 Room | >1 RPDU Bundle trong cùng Room |
| UPS Bundle | — | >1 UPS Bundle trong cùng Floor |

#### 7c. RST Effect

Nếu base SBO **partially reserved** → **APR = 0** cho base SBO đó (mọi premium bị reset).

### Step 8 — Final Rack OAKw

```
Rack.OAKw = Rack.QV × (1 + APR_CC) × (1 + APR_RoomBundle) × (1 + APR_RPDUBundle) × (1 + APR_UPSBundle)
```

**Quan trọng:** Công thức là **nhân** (multiplicative), không phải cộng. Mỗi SBO family nhân riêng.

> Mỗi rack chỉ thuộc 1 base SBO per type (1 CC, 1 Room Bundle, 1 RPDU Bundle, 1 UPS Bundle). Composite SBO benefits đã embedded trong APR qua composite boost.

---

## 7. Phase 4: Portfolio Summary

### Step 9 — Remaining Portfolio Value

```
remaining_portfolio_value = SUM(rack_oak) cho tất cả rack FREE trong scenario
```

- Whitespace scenario: tất cả rack đều free → tổng toàn bộ
- Plan scenario: chỉ rack chưa reserved

### Step 10 — Captured Capacity Value

```
captured_capacity_value = SUM(rack_oak) cho tất cả rack RESERVED trong scenario
```

### Chỉ số bổ sung

```
average_oak_per_rack = remaining_portfolio_value / count(free_racks)
```

### Optionality Impact

```
Optionality_Impact = Σ(Rack.R_OAKw_before) - Σ(Rack.R_OAKw_after)
```

"Before" = scenario trước khi rack được placed. "After" = scenario sau khi placed. Đo lường chi phí optionality bị mất khi commit một placement.

---

## 8. Ví Dụ Tính Toán End-to-End

### Setup

```
Capacity Cell CC-01 (envelope = 400 kW)
├── Row-A (6 racks, total circuit = 240 kW)
│   ├── Rack-A1 (40 kW circuit)
│   ├── Rack-A2 (40 kW circuit)
│   ├── Rack-A3 (40 kW circuit)
│   ├── Rack-A4 (40 kW circuit)
│   ├── Rack-A5 (40 kW circuit)
│   └── Rack-A6 (40 kW circuit)
│   Attributes: density=Tier3/30-60kW(1.12), power=2N(1.14), cool_red=N+1(1.03), cool_mode=Full Liquid(1.07)
│
└── Row-B (6 racks, total circuit = 240 kW)
    ├── Rack-B1 (40 kW circuit)
    ├── Rack-B2 (40 kW circuit)
    ├── Rack-B3 (40 kW circuit)
    ├── Rack-B4 (40 kW circuit)
    ├── Rack-B5 (40 kW circuit)
    └── Rack-B6 (40 kW circuit)
    Attributes: density=Tier3/30-60kW(1.12), power=2N(1.14), cool_red=N+1(1.03), cool_mode=Full Liquid(1.07)

Total rack circuit = 480 kW (lớn hơn envelope 400 kW)
```

### Step 1: Phân bổ kW

```
Row-A allocated = 400 × (240/480) = 200 kW
Row-B allocated = 400 × (240/480) = 200 kW
```

### Step 2: Quality Profile

Cả 2 row giống nhau:
- Density: High → 1.12
- Power redundancy: 2N → 1.14
- Cooling redundancy: N+1 → 1.02
- Cooling mode: Full Liquid → 1.07

### Step 3: Quality-Adjusted Row Value

```
QP = (1.12 × 0.35) + (1.14 × 0.20) + (1.03 × 0.15) + (1.07 × 0.30)
   = 0.392 + 0.228 + 0.1545 + 0.321
   = 1.0955

Row-A value = 200 × 180 × 1.0955 = $39,438
Row-B value = 200 × 180 × 1.0955 = $39,438
```

### Step 4: Capacity Cell Quality Value

```
Coherence Factor:
  Density:      1/1 = 1.0 (cả 2 row đều High)
  Power:        1/1 = 1.0
  Cooling Red:  1/1 = 1.0
  Cooling Mode: 1/1 = 1.0

CF = (1.0 × 0.35) + (1.0 × 0.20) + (1.0 × 0.15) + (1.0 × 0.30) = 1.0

Cell quality value = (39,438 + 39,438) × 1.0 = $78,876
```

### Step 5: Phân phối về Rack

```
Mỗi rack: 78,876 × (40/480) = $6,573
```

12 rack × $6,573 = $78,876 ✓

### Steps 6-8 (nếu có SBO — Whitespace scenario)

```
CC-01 thuộc 1 base SBO (CC) + total_sbo_count = 2 (1 self + 1 composite)
  APR_CC = 5% (SPR, CC is free, not partially reserved)
  APR_RoomBundle = 2% (thuộc 1 room bundle)
  APR_RPDUBundle = 3%
  APR_UPSBundle = 3%

Rack.OAKw = 6,573 × (1+0.05) × (1+0.02) × (1+0.03) × (1+0.03)
          = 6,573 × 1.05 × 1.02 × 1.03 × 1.03
          = 6,573 × 1.1388
          = $7,485
```

### Steps 9-10

```
Whitespace: tất cả free, all SBOs intact
  RPV = 12 × $7,485 = $89,820
  CCV = $0

Plan scenario: 6 rack (Row-A) reserved → CC-01 partially reserved → APR_CC = 0
  Room Bundle also partially reserved → APR_RB = 0
  Rack.OAKw = 6,573 × (1+0) × (1+0) × (1+0.03) × (1+0.03) = 6,573 × 1.0609 = $6,973
  RPV (free racks) = 6 × $6,973 = $41,838
  CCV (reserved racks) = 6 × $6,973 = $41,838
```

> **Chú ý sự mất giá:** Khi reserve 6/12 rack, CC và Room Bundle bị partially reserved → APR_CC (5%) và APR_RB (2%) biến mất. Portfolio mất ~$7,000 giá trị optionality.

---

## 9. Luồng Dữ Liệu

```
                    ┌─────────────────────────────────────┐
                    │         INPUT DATA                   │
                    │  • Capacity cell envelope (kW)       │
                    │  • Rack circuit capacities (kW)      │
                    │  • Row quality attributes (4 dims)   │
                    │  • SBO associations (from SBO Builder)│
                    │  • Reservation status (per scenario)  │
                    └──────────────┬──────────────────────┘
                                   │
                    ┌──────────────▼──────────────────────┐
                    │  STEP 1: Cell → Row kW Allocation    │
                    │  Proportional by rack circuit cap    │
                    └──────────────┬──────────────────────┘
                                   │
                    ┌──────────────▼──────────────────────┐
                    │  STEP 2-3: Row Quality Premium       │
                    │  4 dimensions × weights → QP         │
                    │  row_value = kW × $180 × QP          │
                    └──────────────┬──────────────────────┘
                                   │
                    ┌──────────────▼──────────────────────┐
                    │  STEP 4: Coherence Factor            │
                    │  Cross-row uniformity discount        │
                    │  cell_value = Σ(row_values) × CF     │
                    └──────────────┬──────────────────────┘
                                   │
                    ┌──────────────▼──────────────────────┐
                    │  STEP 5: Cell → Rack Distribution    │
                    │  Proportional by rack circuit cap    │
                    │  → rack_quality_value ($)            │
                    └──────────────┬──────────────────────┘
                                   │
            ═══════════════════════╪═══════════════════════
            SCENARIO BOUNDARY      │
            ═══════════════════════╪═══════════════════════
                                   │
                    ┌──────────────▼──────────────────────┐
                    │  STEP 6-7: SBO Premium               │
                    │  SPR (membership) + APR (boost/disc)  │
                    │  Partially reserved → premium = 0    │
                    └──────────────┬──────────────────────┘
                                   │
                    ┌──────────────▼──────────────────────┐
                    │  STEP 8: Final Rack OAK              │
                    │  rack_oak = quality_value + Σ(APRs)  │
                    └──────────────┬──────────────────────┘
                                   │
                    ┌──────────────▼──────────────────────┐
                    │  STEP 9-10: Portfolio Summary         │
                    │  Remaining = Σ(free rack OAK)        │
                    │  Captured  = Σ(reserved rack OAK)    │
                    └─────────────────────────────────────┘
```

### So sánh với Load Capacity Calculator

| Khía cạnh | Load Capacity | OAK Value |
|---|---|---|
| Hướng tính | Bottom-up (rack → parent) | Top-down (cell → row → rack) rồi bottom-up |
| Đơn vị output | kW, % utilization | Dollar ($) |
| Phụ thuộc | Chỉ spatial hierarchy | + SBO + reservation + quality attributes |
| Scenario | Đơn (thực tế) | Đa scenario (whitespace, plan, draft) |
| Mục đích | "Còn bao nhiêu kW?" | "Mỗi kW đáng bao nhiêu tiền?" |

---

## 10. Phụ Lục: Bảng Hằng Số

### Base Value

| Hằng số | Giá trị | Mô tả |
|---|---|---|
| `BASE_KW_VALUE` | $180 | Giá trị cơ sở per kW (không premium) |

### Quality Premium Weights

| Dimension | Weight | Mô tả |
|---|---|---|
| Density | 0.35 | Mật độ rack |
| Power Redundancy | 0.20 | Dự phòng điện |
| Cooling Redundancy | 0.15 | Dự phòng làm mát |
| Cooling Mode | 0.30 | Phương thức làm mát |

### Density Premium Brackets

| Tier | Range | Premium |
|---|---|---|
| 1 | <= 15 kW | 1.00 |
| 2 | 15 – 30 kW | 1.05 |
| 3 | 30 – 60 kW | 1.12 |
| 4 | 60 – 120 kW | 1.22 |
| 5 | >= 120 kW | 1.30 |

### Power Redundancy Premium

| Config | Premium |
|---|---|
| N (single feed) | 1.00 |
| N+1 | 1.03 |
| A/B feeds | 1.08 |
| 2N architecture | 1.14 |

### Cooling Redundancy Premium

| Config | Premium |
|---|---|
| N | 1.00 |
| N+1 | 1.03 |
| N+2 | 1.08 |
| N+N (dual loop) | 1.14 |

### Cooling Mode Premium

| Mode | Premium |
|---|---|
| Air Only | 1.00 |
| Air + RDHx | 1.03 |
| Air + DTC | 1.07 |
| Full Liquid DTC | 1.07 |

### Coherence Factor

| Phân bố tier giữa rows | Attribute CF |
|---|---|
| Tất cả cùng tier | 1.00 |
| 2 tier khác nhau | 0.98 |
| 3+ tier khác nhau | 0.92 |

### SPR Values

| SBO Family | SPR |
|---|---|
| Capacity Cell | 5% |
| Room Bundle | 2% |
| Room PDU Bundle | 3% |
| UPS Bundle | 3% |

### APR Rules

| Điều kiện | Hiệu ứng |
|---|---|
| Base SBO partially reserved | **APR = 0** (mọi premium mất) |
| Composite SBO partially reserved | Đóng góp 0 vào composite count |
| Composite SBO free/fully reserved | Đóng góp 1 vào composite count |
| CC straddle >1 Room | Entanglement discount (penalty) |
| >1 CC trong cùng Room | In-room modularity (reward) |
| Room Bundle straddle >1 Room | Entanglement discount |
| >1 RPDU Bundle trong cùng Room | In-room modularity (reward) |
| >1 UPS Bundle trong cùng Floor | In-floor modularity (reward) |

---

### Câu Hỏi Đã Giải Quyết (từ Blueprint PDF v0.5)

1. ~~Density bracket thresholds~~ → 5 tiers: <=15, 15-30, 30-60, 60-120, >=120 kW
2. ~~Premium values~~ → Đầy đủ bảng cho cả 4 dimensions
3. ~~Coherence factor formula~~ → Tier-count based: 1 tier=1.0, 2 tiers=0.98, 3+=0.92
4. ~~SPR values~~ → CC=5%, RB=2%, RPDU=3%, UPS=3%
5. ~~Rack.OAKw formula~~ → Multiplicative: QV × (1+APR_CC) × (1+APR_RB) × (1+APR_RPDU) × (1+APR_UPS)

### Câu Hỏi Chưa Giải Quyết

1. **Composite boost function** — CMP (Compound Membership Premium) table has base premiums per count=1 but cells for count=2,3,n are empty in PDF. Need exact multiplier values.
2. **Entanglement discount exact values** — PDF describes penalty/reward rules but exact numeric discount factors not specified.
3. **Custom Factor** — Row quality profile mentions "Custom Factor" as 5th dimension. Weight and values undefined.
4. **Rack.CCAC vs Row.CCAC usage** — OAK step 1 uses Row.CCAC for row value. When distributing CC.QV to rack (step 6), uses Rack_Circuit_Capacity proportion. Confirm these are both using `_design` variant.
