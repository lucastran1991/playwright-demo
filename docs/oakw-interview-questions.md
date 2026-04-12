# OAKw Evaluator — Deep Understanding Interview Questions

**Purpose:** Assess candidate's deep understanding of OAKw system mechanics, edge cases, and architectural reasoning.
**Format:** Questions in English, Answers in Vietnamese (Tiếng Việt có dấu).
**Difficulty:** Hard — only candidates who truly understand the system should pass.

---

## Q1. Why does OAKw use Capacity Cell Envelope instead of the sum of all Rack Circuit Capacities within the cell?

**Trả lời:**

Tổng Rack Circuit Capacity (RCC) thường **lớn hơn** Capacity Envelope vì hiện tượng oversubscription. Ví dụ: một CC có 12 rack, mỗi rack 40kW → tổng RCC = 480kW, nhưng CC Envelope chỉ có 400kW vì bị giới hạn bởi RPP panel capacity (power constraint) hoặc tổng AirZone + Liquid Loop cooling capacity (thermal constraint).

Envelope = MIN(Power_Capacity, Thermal_Capacity), trong đó:
- Power_Capacity = Σ(RPP_Panel_Capacity_design) của các RPP phục vụ CC
- Thermal_Capacity = Σ(AirZone_Cooling_Capacity) + Σ(LL_Cooling_Capacity)

Nếu dùng tổng RCC (480kW) thay vì Envelope (400kW), ta sẽ **định giá quá cao** — bán ra nhiều kW hơn thực tế có thể cam kết. Điều này giống như bán 480m² diện tích trong tòa nhà chỉ có 400m² sử dụng được. Envelope đảm bảo mỗi dollar OAKw đại diện cho kW **thực sự deliverable**.

---

## Q2. Two racks (Rack-A and Rack-B) are in the same Capacity Cell, both have RCC_design = 40kW, yet someone claims their OAKw values differ. Is this possible? Explain thoroughly.

**Trả lời:**

**Không thể xảy ra** nếu hai rack cùng CC và cùng RCC_design.

Công thức Rack.QV = CC.QV × (Rack.RCC / CC.Total_RCC). Vì Rack-A và Rack-B có cùng RCC = 40kW, tỷ lệ phân bổ giống nhau → Rack.QV giống nhau.

Tiếp theo, APR được tính ở cấp SBO, không phải cấp rack. Cả hai rack cùng thuộc 1 CC (cùng APR_CC), cùng 1 Room Bundle (cùng APR_RB), cùng 1 Room PDU Bundle (cùng APR_RPDU), cùng 1 UPS Bundle (cùng APR_UPS). Vậy Rack.OAKw = Rack.QV × (1+APR_CC) × (1+APR_RB) × (1+APR_RPDU) × (1+APR_UPS) — hoàn toàn giống nhau.

**Bẫy phổ biến:** Nhiều người sẽ trả lời "do quality khác nhau vì ở 2 row khác nhau". Sai — quality là thuộc tính của Row, ảnh hưởng đến Row.QV, nhưng khi đã tính xong CC.QV (= Σ Row.QV × CF), việc phân bổ về rack CHỈ dựa trên tỷ lệ RCC. Thông tin row-level đã được "gộp" và "trộn" trong CC.QV rồi.

---

## Q3. A CC has 4 rows: Row-A (Tier 3), Row-B (Tier 3), Row-C (Tier 1), Row-D (Tier 1). What is the Attribute Coherence Factor for density? Now remove Row-D. Does CF change?

**Trả lời:**

**4 rows:** Có 2 distinct tiers (Tier 3 và Tier 1) → CF_density = **0.98**.

**Bỏ Row-D, còn 3 rows:** Row-A (Tier 3), Row-B (Tier 3), Row-C (Tier 1) → vẫn 2 distinct tiers → CF_density = **0.98**. Không thay đổi.

**Điểm quan trọng mà ứng viên cần hiểu:** CF phụ thuộc vào **số lượng tier khác nhau tồn tại** trong CC, KHÔNG phải số row có tier khác nhau hay tỷ lệ row. 10 row Tier 3 + 1 row Tier 1 vẫn cho CF = 0.98 (2 tiers). Ngược lại, 3 row với Tier 1, Tier 3, Tier 5 cho CF = 0.92 (3 tiers) dù chỉ có 3 row.

Quy tắc:
- 1 tier duy nhất → CF = 1.00 (đồng nhất hoàn hảo)
- 2 tiers → CF = 0.98 (phạt nhẹ 2%)
- 3+ tiers → CF = 0.92 (phạt nặng 8%, tối đa)

---

## Q4. Why is the APR formula multiplicative (QV × (1+APR₁) × (1+APR₂) × ...) rather than additive (QV × (1 + APR₁ + APR₂ + ...))?

**Trả lời:**

Mỗi SBO family (CC, Room Bundle, RPDU Bundle, UPS Bundle) cung cấp **giá trị tùy chọn độc lập**. Optionality từ CC không phụ thuộc vào optionality từ UPS Bundle — chúng là các "layer" riêng biệt.

Phép nhân phản ánh **lãi kép (compound interest)**: mỗi layer premium được áp dụng lên giá trị đã được premium bởi layer trước. Đây là mô hình chính xác hơn vì:

- **Phép cộng** (additive): 4 APR = 10% → 1 + 0.4 = 1.40 (+40%). Giả định sai rằng mỗi layer có giá trị cố định bất kể các layer khác.
- **Phép nhân** (multiplicative): 4 APR = 10% → 1.1⁴ = 1.4641 (+46.41%). Phản ánh đúng rằng rack thuộc nhiều SBO đồng thời có giá trị cao hơn tổng từng SBO riêng lẻ — vì khách hàng có **nhiều cách mua** cùng lúc.

Ví dụ thực tế: Rack thuộc CC free + Room Bundle free → có thể bán như 1 CC riêng HOẶC 1 Room Bundle riêng HOẶC kết hợp. Mỗi cách bán NHÂN lên lựa chọn, không cộng.

Test case F16-04 xác nhận: QV=$10,000, APR=10% × 4 layers → $14,641 (nhân), KHÔNG phải $14,000 (cộng).

---

## Q5. CC-01 has 12 racks. PP-01 reserves ALL 12 racks. What is CC-01's RST? What is APR_CC? Many candidates get this wrong — explain why.

**Trả lời:**

RST = **Reserved** (KHÔNG phải Partially Reserved).
APR_CC = **5%** (giữ nguyên premium).

**Tại sao nhiều người trả lời sai:** Họ nghĩ "rack bị reserved → CC mất premium". Sai hoàn toàn. Quy tắc là:

| RST | Điều kiện | APR |
|---|---|---|
| Free | Tất cả rack free | 5% (giữ) |
| Reserved | Tất cả rack reserved cho **cùng 1 PP** | 5% (giữ) |
| Partially Reserved | Hỗn hợp free + reserved, HOẶC reserved cho nhiều PP khác nhau | **0%** (mất hết) |

Logic kinh doanh: khi CC **fully reserved** cho 1 khách hàng, tính module hóa được bảo toàn — khách hàng đó sở hữu toàn bộ CC, ranh giới điện/nhiệt không bị phá vỡ. CC vẫn là một đơn vị kinh doanh hoàn chỉnh.

Chỉ **Partially Reserved** mới phá vỡ tính module hóa: một phần CC thuộc khách A, phần còn lại free — khách B không thể mua nguyên CC nữa → mất optionality → mất premium.

Thêm edge case: 6 rack reserved cho PP-01, 6 rack reserved cho PP-02 → RST = **Partially Reserved** (dù tất cả rack đều reserved). Vì CC bị chia cho 2 khách hàng khác nhau → phá vỡ tính thống nhất.

---

## Q6. Rack-X is reserved in both PD-01 and PP-02. What is the RCST? What happens when someone tries to promote PD-01 to a Placement Plan?

**Trả lời:**

RCST = **Hard Conflict**. Quy tắc:
- Rack trong PD + PP → Hard Conflict (PP đã cam kết, xung đột nghiêm trọng)
- Rack trong PD + PD → Soft Conflict (cả hai chưa cam kết, xung đột nhẹ)
- Rack chỉ trong 1 PP hoặc 1 PD → None

Khi promote PD-01 → PP: **BỊ CHẶN**. Hệ thống không cho phép PD có rack bị Hard Conflict trở thành PP. Lý do: nếu promote, Rack-X sẽ thuộc 2 PP đồng thời — điều này vô nghĩa về mặt kinh doanh (không thể giao cùng 1 rack cho 2 khách hàng đã ký hợp đồng).

Để giải quyết, phải chọn 1 trong 2:
1. Release Rack-X khỏi PD-01 (thay rack khác)
2. Release Rack-X khỏi PP-02 (hủy cam kết với khách hàng kia — rất nghiêm trọng)

---

## Q7. CC-01 is in Room-1 only. CC-02 spans Room-1 and Room-2. Can the SBO Builder create a Composite CC from {CC-01, CC-02}? What is CC-02's total_sbo_count?

**Trả lời:**

**Không thể tạo Composite CC** từ {CC-01, CC-02}.

Quy tắc SBO Builder: Composite CC chỉ được tạo từ các CC **eligible** nằm trong **cùng 1 room**. CC-02 có room_count = 2 → **không eligible** cho room-based composite. Điều này đúng ngay cả khi CC-02 có chung Room-1 với CC-01.

CC-02 vẫn là Base SBO hợp lệ (tự nó = 1 SBO), nhưng:
- Không tham gia bất kỳ Composite CC nào
- total_sbo_count = **1** (chỉ có chính nó)

CC-01 có thể tạo composite với **CC khác** cũng nằm riêng trong Room-1 (room_count = 1), nhưng KHÔNG với CC-02.

Ý nghĩa kinh doanh: CC trải qua nhiều phòng đã phức tạp (entanglement penalty). Kết hợp nó với CC khác tạo thêm phức tạp không cần thiết — SBO Builder chặn điều này bằng eligibility rule.

---

## Q8. A facility has 3 CCs: CC-A (10 racks), CC-B (10 racks), CC-C (10 racks). All in Whitespace scenario (everything free). A sales team wants to reserve 10 racks for a customer. Compare the optionality impact of two strategies: (A) reserve all 10 racks from CC-A vs (B) reserve 5 from CC-A and 5 from CC-B.

**Trả lời:**

**Strategy A (10 racks từ CC-A):** CC-A fully reserved → RST = Reserved → APR_CC vẫn = 5%. CC-B và CC-C không bị ảnh hưởng. Tất cả 20 rack free còn lại giữ nguyên OAKw.

**Strategy B (5 từ CC-A + 5 từ CC-B):** CC-A partially reserved → APR_CC = 0%. CC-B cũng partially reserved → APR_CC = 0%. Tổng cộng **20 rack** bị ảnh hưởng (10 free rack trong CC-A + 10 free rack trong CC-B mất premium CC). Nếu Room Bundle cũng bị partially reserved → mất thêm APR_RB.

**So sánh Optionality Impact:**

| Metric | Strategy A | Strategy B |
|---|---|---|
| CC bị partial | 0 | 2 (CC-A, CC-B) |
| Free rack mất premium | 0 | 10 (5 trong CC-A + 5 trong CC-B) |
| APR_CC mất | 0% | 5% × 10 rack |
| Optionality Impact | Gần **$0** | **Âm lớn** (−$5,000+) |

**Kết luận:** Strategy A tốt hơn vượt trội. Đây chính là giá trị cốt lõi của OAKw — nó chứng minh bằng số liệu rằng **cách reserve quan trọng hơn số lượng reserve**. OAKw hướng dẫn sales team: "hãy lấp đầy nguyên CC thay vì cắt ngang nhiều CC".

---

## Q9. CC-X has total_sbo_count = 5. One of its Composite SBOs becomes partially reserved. What happens to the effective SBO count? How does this affect APR?

**Trả lời:**

total_sbo_count ban đầu = 5 (1 self + 4 composite SBOs).

Khi 1 Composite SBO bị partially reserved:
- Composite SBO đó **đóng góp 0** vào effective count (thay vì 1)
- Effective SBO count = 1 (self) + 3 (composite free/fully reserved) + 0 (partial) = **4**

Ảnh hưởng đến APR: Composite Boost giảm vì count giảm từ 5 → 4. Compound Membership Premium (CMP) phụ thuộc vào count — count thấp hơn → CMP thấp hơn → APR_CC giảm.

**Lưu ý tinh tế:**
1. Base SBO (CC-X bản thân) bị partially reserved → APR_CC = **0%** (mất hoàn toàn, không chỉ giảm)
2. Composite SBO bị partially reserved → chỉ **giảm count**, APR vẫn > 0 (trừ khi base SBO cũng partial)
3. Composite SBO **fully reserved** (không phải partial) → vẫn đóng góp 1 vào count

Đây là sự phân biệt quan trọng giữa **base SBO RST** (quyết định APR = 0 hay không) và **composite SBO RST** (chỉ ảnh hưởng count/boost).

---

## Q10. The Coherence Factor uses the SAME weights as Quality Premium (Density 35%, Power 20%, Cooling Red 15%, Cooling Mode 30%). Why would the designers choose the same weights? Could different weights make more sense?

**Trả lời:**

Cùng trọng số vì **triết lý thiết kế nhất quán**: nếu Density quan trọng nhất trong việc đánh giá chất lượng (35%), thì sự không đồng nhất về Density cũng quan trọng nhất trong việc phạt coherence (35%).

Tuy nhiên, đây là một **quyết định thiết kế có thể tranh luận**:

**Lý do dùng cùng weight:**
- Đơn giản, dễ hiểu, dễ giải thích cho stakeholder
- Tránh thêm 4 tham số cần calibrate (YAGNI)
- Nếu density quan trọng nhất khi đánh giá → nó cũng nên quan trọng nhất khi phạt

**Lý do có thể dùng weight khác:**
- Sự không đồng nhất về cooling mode (Air vs Liquid) có thể tạo vấn đề vận hành lớn hơn sự không đồng nhất về density → weight CF_cooling_mode nên cao hơn 30%
- Density không đồng nhất thực ra ít ảnh hưởng vận hành → weight CF_density có thể thấp hơn 35%
- Coherence Factor đo **rủi ro vận hành**, Quality Premium đo **giá trị thương mại** — hai concept khác nhau có thể cần weight khác nhau

**Trong hệ thống hiện tại**, dùng cùng weight là quyết định pragmatic đúng đắn: ít tham số hơn = ít bug hơn = dễ audit hơn. Tách weight khi và chỉ khi có dữ liệu thực chứng minh cần thiết.

---

## Q11. Explain why Row — not Rack, not CC — is the unit for quality attribute assignment. What would break if quality were assigned at Rack level? At CC level?

**Trả lời:**

**Tại sao Row:**
Row là đơn vị nhỏ nhất có thuộc tính chất lượng **tương đối đồng nhất**. Trong thực tế data center:
- Tất cả rack trong 1 row thường dùng cùng loại cooling (air hoặc liquid) — vì ống nước/ống gió đi dọc row
- Tất cả rack trong 1 row thường cùng mức power redundancy — vì chúng nối cùng RPP/PDU feeds
- Mật độ (density) trong row tương đối đồng đều — vì row được thiết kế cho 1 loại workload

**Nếu gán ở cấp Rack (quá chi tiết):**
- Mỗi rack cần 4 thuộc tính × hàng nghìn rack = **khối lượng dữ liệu khổng lồ** cần maintain
- Coherence Factor phải tính ở cấp rack thay vì row → số distinct tiers tăng nhanh → hầu hết CC sẽ có CF = 0.92 (3+ tiers) → mất khả năng phân biệt CC tốt và CC xấu
- Quality Premium sẽ khác nhau giữa các rack cùng row → công thức phức tạp hơn nhiều (cần weighted average thay vì row-level assignment)
- Vi phạm KISS — thêm granularity mà không thêm giá trị kinh doanh

**Nếu gán ở cấp CC (quá thô):**
- 1 CC có thể có row air-cooled và row liquid-cooled → gán "Air" hay "Liquid" cho cả CC? Không chính xác
- Mất khả năng tính Coherence Factor — CF đo sự khác biệt GIỮA các row. Nếu quality ở cấp CC thì chỉ có 1 giá trị → CF luôn = 1.0 → vô nghĩa
- CC-level quality che giấu sự không đồng nhất nội bộ → định giá sai

**Row là sweet spot**: đủ chi tiết để phản ánh sự khác biệt thực tế, đủ thô để quản lý được, và cho phép Coherence Factor hoạt động có ý nghĩa.

---

## Q12. In a scenario where a Placement Draft (PD-01) reserves 3 racks in CC-02, and an existing Placement Plan (PP-01) reserves 6 racks in CC-01: does CCV_PP01 change because of PD-01? Explain the scenario isolation principle.

**Trả lời:**

**CCV_PP01 KHÔNG thay đổi** khi thêm PD-01.

Nguyên tắc cách ly scenario (Scenario Isolation):

| Scenario | Bao gồm | CCV tính từ |
|---|---|---|
| S_w (Whitespace) | Không có reservation | Không có CCV |
| S_0 (Planned) | Chỉ các PP | CCV_PP tính trong S_0 |
| S_PD01 (Draft 01) | PP + PD-01 | CCV_PD01 tính trong S_PD01 |

CCV_PP01 luôn được tính trong **S_0** — scenario chỉ chứa PP. PD-01 nằm trong S_PD01 — một scenario riêng biệt. Hai scenario **không ảnh hưởng lẫn nhau**.

**Tại sao cách ly:**
- PP là cam kết chính thức — giá trị của nó phải ổn định, không bị dao động bởi các draft đang thương lượng
- PD là thử nghiệm ("what-if") — sales team muốn xem **nếu reserve thêm** thì portfolio thay đổi thế nào. Họ không muốn draft của họ làm thay đổi số liệu PP đã báo cáo cho management
- Nhiều PD có thể tồn tại đồng thời (PD-01, PD-02, PD-03) — mỗi cái là một "what-if" scenario riêng

**Tuy nhiên**, trong S_PD01, OAKw của các rack trong PP-01 **CÓ THỂ khác** so với S_0 nếu PD-01 làm thêm CC/Bundle bị partially reserved. Nhưng đây là OAKw trong context S_PD01, không phải CCV_PP01 chính thức.

---

## Q13. The Entanglement Discount has both penalties AND rewards. Give an example where a CC receives a reward (discount > 1.0). Why does the system reward this specific configuration?

**Trả lời:**

**Ví dụ reward:** Room-1 chứa 3 CC: CC-A, CC-B, CC-C. Tất cả nằm gọn trong Room-1 (room_count = 1).

→ Entanglement Discount cho mỗi CC = **> 1.0 (modularity boost)**.

**Tại sao thưởng:**
Khi nhiều CC cùng nằm trong 1 phòng, khách hàng có **nhiều lựa chọn mua sắm hơn** trong cùng không gian vật lý:
- Mua 1 CC riêng lẻ
- Mua 2 CC kết hợp (Composite SBO)
- Mua cả 3 CC (Composite SBO lớn hơn)

Tính tập trung này tạo ra **hiệu ứng modularity**: phòng đó trở thành "kệ hàng" với nhiều sản phẩm → dễ bán hơn → giá trị cao hơn.

**Bảng đầy đủ:**

| SBO Type | Bị phạt (discount < 1.0) khi | Được thưởng (discount > 1.0) khi |
|---|---|---|
| CC | CC trải qua >1 phòng (cross-room) | >1 CC cùng phòng (in-room modularity) |
| Room Bundle | RB trải qua >1 phòng | — |
| Room PDU Bundle | RPDU trải qua >1 phòng | >1 RPDU Bundle cùng phòng |
| UPS Bundle | — | >1 UPS Bundle cùng tầng (floor modularity) |

**Triết lý:** Hệ thống thưởng cho **tính tập trung** (concentrated options = good) và phạt cho **tính phân tán** (sprawled infrastructure = bad). Đây là bản chất của "entanglement" — càng nhiều phụ thuộc cross-boundary, càng khó quản lý, càng giảm giá trị.

---

## Q14. A facility has 100 racks, all free. All OAKw = $7,000. RPV = $700,000. You reserve 10 racks without breaking any CC containment (full CC reservation). Is RPV now $630,000? Trick question — explain all the nuances.

**Trả lời:**

**Câu trả lời nhanh: Đúng, RPV ≈ $630,000.** Nhưng cần phân tích kỹ vì có nhiều nuance.

**Tại sao gần đúng $630,000:**
- 10 rack reserved → 90 rack free
- CC được reserve nguyên vẹn → RST = Reserved (không phải Partial) → APR_CC vẫn = 5%
- Các CC khác không bị ảnh hưởng
- 90 free rack vẫn giữ OAKw = $7,000
- RPV = 90 × $7,000 = $630,000 ✓

**Nuance 1: Composite SBO có thể bị ảnh hưởng**
Nếu CC được reserve nằm trong Composite SBO với CC khác → Composite SBO đó bây giờ **fully reserved** (vì member CC fully reserved). Composite SBO fully reserved vẫn đóng góp 1 vào count → **không ảnh hưởng gì**. Nhưng nếu logic khác — cần kiểm tra.

**Nuance 2: Room Bundle / RPDU Bundle / UPS Bundle**
CC fully reserved → Room Bundle có thể bị partially reserved (nếu RB chứa CC khác vẫn free). Lúc này APR_RB = 0% cho TẤT CẢ rack trong RB đó → OAKw giảm → RPV < $630,000.

**Ví dụ:** Room Bundle chứa CC-A (10 rack, fully reserved) + CC-B (10 rack, free). RB RST = Partially Reserved → APR_RB = 0% cho 10 rack free trong CC-B. Mỗi rack CC-B mất 2% premium → OAKw giảm. RPV < $630,000.

**Nuance 3: Điều kiện chính xác RPV = $630,000**
Chỉ đúng nếu CC được reserve là CC **duy nhất** trong Room Bundle, Room PDU Bundle, và UPS Bundle của nó. Khi đó tất cả bundle đều fully reserved → không có partial → không ảnh hưởng rack khác.

**Kết luận:** Câu hỏi này test xem ứng viên có hiểu cascade effect từ CC → Bundle hay không. Câu trả lời "$630,000" chỉ đúng trong điều kiện đặc biệt.

---

## Q15. The OAKw system computes values per-scenario, meaning the same rack can have different OAKw in S_w, S_0, and S_PD. If you had to store and query these efficiently in PostgreSQL, what is the key architectural challenge? How would you avoid the N×M explosion problem?

**Trả lời:**

**Thách thức chính:** N racks × M scenarios = N×M giá trị OAKw. Nếu facility có 10,000 rack và 50 PD scenarios → 500,000 rows chỉ cho OAKw. Mỗi lần thêm 1 PD mới → tính lại OAKw cho toàn bộ 10,000 rack trong scenario đó.

**Vấn đề N×M explosion:**
- S_w: 10,000 giá trị (tính 1 lần, hiếm khi thay đổi)
- S_0: 10,000 giá trị (thay đổi khi PP thêm/bớt)
- S_PD01...S_PD50: 50 × 10,000 = 500,000 giá trị (mỗi draft là 1 scenario riêng)

**Giải pháp kiến trúc — tính theo layer, không lưu toàn bộ:**

1. **Steps 1-5 (non-scenario):** Tính 1 lần, lưu Rack.QV. Giá trị này **bất biến** qua tất cả scenario — chỉ phụ thuộc vào CC envelope và quality attributes. Lưu 10,000 rows.

2. **Steps 6-8 (scenario-dependent):** CHỈ lưu **delta** — RST per SBO per scenario. Bảng `sbo_reservation_status(scenario_id, sbo_id, rst)` nhỏ hơn nhiều (vài trăm SBO × vài chục scenario). OAKw được tính **on-the-fly** từ Rack.QV + APR(từ RST).

3. **Steps 9-10 (aggregation):** Tính RPV/CCV on-demand bằng query:
```sql
SELECT SUM(rack_qv * apr_multiplier) 
FROM racks JOIN scenario_aprs ON ... 
WHERE scenario_id = ? AND reservation_status = 'free'
```

**Tránh N×M bằng cách:**
- Lưu Rack.QV (N rows, scenario-independent)
- Lưu SBO RST per scenario (S × SBO_count rows, nhỏ)
- Tính APR multiplier tại query time từ RST
- Tính OAKw = QV × APRs tại query time
- KHÔNG pre-compute và lưu OAKw per rack per scenario

**Trade-off:** Query chậm hơn (tính APR on-the-fly) nhưng storage nhỏ hơn rất nhiều và thêm scenario mới không cần tính lại toàn bộ. Nếu cần tốc độ, cache S_w và S_0 (2 scenario phổ biến nhất), tính S_PD on-demand.
