# Design: SaleScale frontend UI foundation

## 1. Ý định

Xây dựng nền tảng giao diện dùng chung cho frontend React/Vite của SaleScale, bám theo bộ HTML mẫu tại `docs/ui-design/pages` và giữ nguyên các hành vi backend đã triển khai. Mục tiêu trước mắt là tạo theme, customer shell, admin shell và các component dùng chung; sau đó áp dụng chúng cho các màn hình đang có trong frontend.

Bộ HTML mẫu là nguồn tham chiếu trực quan và trạng thái UI. OpenAPI cùng backend hiện tại là nguồn sự thật cho request, response, quyền truy cập và mã lỗi.

### Baseline hiện tại

Theo [Phase 03 — Flash sale và giỏ hàng](../../plans/phases/phase-03-sale-cart.md), phạm vi FE/BE hiện đã đi tới sale campaign, cart nhiều SKU và checkout preview cơ bản. Phase 03 ghi nhận backend 123 test, frontend 110 test và e2e qua API đạt 35/35 bước; checkpoint còn lại là chạy lại hành trình trên trình duyệt sau các đợt sửa lỗi. Vì vậy design này lấy Phase 03 làm mốc triển khai thực tế, không coi toàn bộ 76 HTML mẫu là phạm vi phải hoàn thành ngay.

## 2. Mục tiêu

- Đồng nhất giao diện customer và admin với mẫu SaleScale.
- Tách layout dùng chung khỏi page và feature để không lặp header, navigation, card, form và trạng thái dữ liệu.
- Chuẩn hóa responsive layout cho desktop, tablet và màn hình nhỏ.
- Giữ đúng các quy tắc nghiệp vụ hiện tại: giá/tổng tiền/tồn kho/trạng thái do backend quyết định; checkout preview không giữ hàng; lỗi hiển thị theo mã ổn định và lỗi theo từng dòng.
- Chuẩn hóa toàn bộ frontend dùng route API dưới tiền tố `/api`, không dùng `/api/v1`.
- Mỗi lát cắt UI phải được kiểm tra song song với backend thật hoặc test server đúng contract; không coi frontend mock là bằng chứng hoàn tất.
- Cho phép chỉnh backend ở mức tối thiểu khi cần để khớp route/payload/error contract hoặc làm integration test chạy được, nhưng không mở rộng nghiệp vụ.
- Chỉ triển khai tới các màn hình và chức năng đã có trong frontend hiện tại; không dựng UI giả cho các phase sau.

## 3. Không thuộc phạm vi

- Không mở rộng backend, schema, OpenAPI operation hoặc nghiệp vụ checkout/payment. Chỉ chấp nhận thay đổi backend tối thiểu nếu integration chứng minh contract hiện tại không thể dùng đúng.
- Không triển khai các màn hình chưa có luồng frontend hoàn chỉnh trong phase này: product detail, flash-sale detail, order result, payment, my orders, inventory, admin product form/list đầy đủ, reconciliation và notifications.
- Không chuyển toàn bộ 76 file HTML tĩnh thành route React ngay một lần.
- Không đưa dữ liệu mẫu tĩnh vào luồng thật để che API còn thiếu.
- Không mở rộng search thành global search; phase này chỉ hỗ trợ tìm theo tên trong catalog public vì backend đã có query contract.
- Không thay đổi các bất biến nghiệp vụ, error code hoặc cơ chế session/CSRF.

## 4. Chống mục tiêu

- Không biến customer shell thành dashboard quản trị.
- Không dùng màu, badge hoặc trạng thái chỉ để trang trí nếu không có ý nghĩa nghiệp vụ.
- Không dùng emoji làm icon cấu trúc; icon dùng SVG nhất quán và có accessible name khi là control.
- Không đặt logic gọi API, tính nghiệp vụ hoặc xử lý lỗi vào component presentational dùng chung.
- Không sửa backend để làm UI pass bằng cách nới validation, bỏ authorization, bỏ CSRF, thay đổi trạng thái nghiệp vụ hoặc che lỗi transaction.

## 5. Các hướng đã cân nhắc

### Hướng A — Foundation-first (được chọn)

Dựng token, customer/admin shell và component dùng chung trước, sau đó áp dụng cho các route hiện có: auth, catalog list, flash-sale list, cart/preview và admin flash-sale. Đây là hướng phù hợp nhất với yêu cầu đã thống nhất, giảm lặp code và không mở rộng phạm vi backend.

### Hướng B — Vertical slice

Làm hoàn chỉnh riêng luồng `Flash Sale → Cart → Preview` trước rồi mới quay lại admin. Hướng này cho kết quả trực quan nhanh nhưng dễ tạo component tạm hoặc phải chỉnh lại theme khi mở rộng sang auth và admin.

### Hướng C — Full migration

Chuyển toàn bộ 76 HTML screen/state thành React route ngay từ đầu. Hướng này có độ bao phủ cao nhưng vượt phạm vi phase hiện tại và sẽ tạo nhiều màn hình chưa có API/hành vi thật.

## 6. Theme và design tokens

Các giá trị dưới đây được rút ra từ bộ mẫu; mọi component dùng token semantic thay vì hard-code màu riêng theo page.

| Nhóm | Token đề xuất | Giá trị tham chiếu |
|---|---|---|
| Font | `font-family` | Inter, system-ui, Segoe UI, sans-serif |
| Nền ứng dụng | `color-bg-app` | `#f5f5f5` |
| Surface | `color-surface` | `#ffffff` |
| Text chính | `color-text` | `#222222` |
| Text phụ | `color-text-muted` | `#757575` |
| Border | `color-border` | `#e8e8e8` |
| Border input | `color-border-input` | `#bdbdbd` |
| Primary | `color-primary` | `#ee4d2d` |
| Primary text/link | `color-primary-strong` | `#d73211` |
| Success | `color-success` | `#26aa99` / `#00796b` |
| Warning | `color-warning` | `#ffbf00` / `#8a6100` |
| Danger | `color-danger` | `#d0011b` |
| Disabled surface | `color-disabled` | `#f5f5f5` |
| Control height | `control-md` | 40px |
| Touch/action height | `control-lg` | 44px |
| Card radius | `radius-card` | 4px |
| Control radius | `radius-control` | 2px |
| Spacing rhythm | `space-*` | 4/8/12/16/24/32px |
| Body type | `type-body` | 14px / 20px |
| Page title | `type-title` | 20px / 28px, weight 600 |
| Section title | `type-section` | 16px / 24px, weight 600 |
| Supporting text | `type-caption` | 12px / 16px |

Surface separation ưu tiên border 1px và chênh lệch nền nhẹ; shadow chỉ dùng rất nhẹ cho card/overlay. Trạng thái phải có cả màu, nhãn và icon/shape để không phụ thuộc duy nhất vào màu.

## 7. Layout dùng chung

### 7.1 Customer shell

Đặt trong `app/frontend/src/components/layout/CustomerShell.tsx` và các component con cùng thư mục.

- Header customer theo mẫu: thanh tiện ích trên nền primary, logo SaleScale, liên kết Flash Sale, tài khoản; hàng chính có ô tìm kiếm catalog và giỏ hàng.
- Nội dung nằm trong container tối đa khoảng 1232px, gutter 16px trên mobile và tăng lên ở desktop.
- Route hiện tại được biểu thị rõ bằng màu/weight/underline; link và icon có focus-visible.
- Trên mobile, header chuyển thành layout nhiều hàng hoặc menu có thể truy cập bằng bàn phím; không ép search và navigation trên cùng một hàng hẹp.
- Dùng cho `/`, `/login`, `/register`, `/products`, `/flash-sales`, `/cart`, `/account` và các customer page được triển khai trong phase này.

### 7.2 Admin shell

Đặt trong `app/frontend/src/components/layout/AdminShell.tsx` và các component con cùng thư mục.

- Sidebar trái khoảng 220px trên desktop, nền trắng, border phải, logo và nhãn “Quản trị”.
- Navigation chỉ hiển thị các route đã triển khai trong phase hiện tại; Flash sale là mục admin đang hoạt động. Products và Tồn kho không hiển thị dưới dạng link chết trước khi page/route tương ứng được triển khai. Item active dùng nền cam rất nhạt, text primary-strong và `aria-current="page"`.
- Header admin có link “Xem trang khách hàng”, user hiện tại và logout.
- Content padding khoảng 24px desktop, co về 16px trên màn hình nhỏ.
- Trên mobile, sidebar trở thành drawer/menu có nút mở đóng; không làm mất đường dẫn hoặc focus.
- Dùng cho `/admin` và `/admin/flash-sales` ở phạm vi hiện tại; route admin khác chỉ thêm khi feature/API tương ứng được phê duyệt.

### 7.3 Route composition

Chuyển router sang layout route để page chỉ trả về nội dung chính:

```text
CustomerShell
├── LoginPage
├── RegisterPage
├── ProductsPage
├── FlashSalesPage
├── CartPage
└── AccountPage

AdminShell
├── AdminPage
└── AdminFlashSalePage
```

`HealthPage` là màn hình chẩn đoán nội bộ; có thể giữ route riêng và dùng component trạng thái chung, không cần ép thành một customer page trong đợt visual parity đầu tiên.

## 8. Component dùng chung

Đặt trong `app/frontend/src/components/common/`. Danh sách ban đầu:

- `Button`: primary, secondary, danger, ghost, loading, disabled; giữ kích thước và focus state thống nhất.
- `IconButton`: control icon-only bắt buộc có `aria-label`.
- `Card` và `Section`: surface trắng, border, radius và padding theo token.
- `Badge` / `StatusBadge`: order, payment, flash-sale và price-source; mapping trạng thái tập trung.
- `PriceBlock`: giá thường, giá flash sale, giá cũ, nguồn giá và thời điểm kết thúc.
- `Countdown`: hiển thị thời gian hệ thống; không coi countdown phía client là nguồn quyết định nghiệp vụ.
- `QuantityStepper`: tăng/giảm số lượng, output hiện tại, giới hạn và trạng thái busy.
- `FormField`, `TextInput`, `SelectField`, `FieldError`, `ErrorSummary`: label, helper text, `aria-invalid`, `aria-describedby` và lỗi theo field.
- `Alert` / `ActionNotice`: success, warning, error, info; hỗ trợ `role="alert"` hoặc `aria-live` đúng ngữ cảnh.
- `LoadingState`, `EmptyState`, `ErrorState`, `ForbiddenState`, `NotFoundState`, `OverloadState`: dùng chung cho list/card/page.
- `Breadcrumbs`, `SearchField`, `Tabs`, `LoadMoreButton` và `DataTable` cho các page đã có hoặc sẽ được dùng ở phase kế tiếp.

Component dùng chung chỉ nhận dữ liệu và callback; chúng không tự gọi service. Các feature giữ logic query/mutation và chuyển dữ liệu vào component.

## 9. Phạm vi áp dụng trong phase hiện tại

Phạm vi visual implementation bám Phase 03: `FR-014–020`, `INV-09/10`, `TEST-09/10/20`, đồng thời hoàn thiện các behavior frontend nền tảng đã có contract P0 như catalog search. Các page phải giữ được hành vi sale lifecycle, cart version/ownership, lựa chọn nhiều dòng và preview không reserve; chi tiết đặt đơn chuyển sang Phase 04.

| Khu vực | Route/component hiện có | Điều chỉnh chính |
|---|---|---|
| Auth | `/login`, `/register`, `AuthFormView`, `AuthNav` | Dùng customer header, form field/error component, tab đăng nhập/đăng ký và trạng thái submitting/error theo mẫu |
| Catalog | `/products`, `ProductCard` | Chuyển từ card Tailwind hiện tại sang product grid/card theo mẫu; thêm search theo tên qua `GET /api/products?query=...`; giữ chọn SKU, số lượng, thêm giỏ |
| Flash sale customer | `/flash-sales`, `FlashSaleCampaignCard`, `Countdown` | Badge trạng thái, price block, countdown, CTA và loading/empty/error theo mẫu |
| Cart | `/cart`, `CartLineItem`, `CheckoutPreviewPanel` | Card/list 2 cột responsive, line error, selection, quantity stepper, price-change acknowledgement; không tự tạo order mới trong khi checkout chưa triển khai |
| Admin | `/admin`, `/admin/flash-sales`, form modal/card | Dùng admin shell, table/card style, status badge, form sections và action state theo mẫu |
| Health | `/health` | Giữ chức năng chẩn đoán; tái sử dụng `LoadingState`/`ErrorState` nếu phù hợp |

`/` nên trở thành alias của customer/public catalog trong phase này, thay vì luôn redirect authenticated user sang `/account`. Phần Flash Sale nổi bật trong `Main.html` được để phase sau vì destination detail chưa thuộc phạm vi; `ProductsPage` không nên bắt buộc đăng nhập vì `GET /api/products` là public, còn thao tác thêm giỏ mới yêu cầu login và giữ `next` path.

## 10. Chuẩn hóa API

Tất cả frontend service dùng path dưới `/api`:

| Feature | Path chuẩn |
|---|---|
| Auth | `/api/auth/register`, `/api/auth/login`, `/api/auth/me`, `/api/auth/logout` |
| Catalog | `/api/products`, `/api/products/:productId` |
| Flash sale | `/api/flash-sales`, `/api/flash-sales/:id` |
| Cart/preview | `/api/cart`, `/api/cart/items/:skuId`, `/api/checkout/preview` |
| Health | `/api/health/ready` |
| Admin flash sale | `/api/admin/flash-sales...` |

- Xóa toàn bộ `/api/v1` khỏi `flashsale.service.ts` và các test URL liên quan.
- Dùng một biến môi trường `VITE_API_BASE_URL` cho mọi service. Giá trị có thể là origin backend (`http://localhost:4000`) hoặc rỗng khi dùng Vite proxy cùng-origin; client thêm đúng một lần tiền tố `/api` trong cả hai trường hợp.
- Giữ `credentials: include`, session cookie và CSRF header theo contract.
- Không phân tích message text để quyết định UI; map theo `code`, `violations`, `line_errors`, status HTTP và `Retry-After`.
- Không coi redirect/payment state phía client là xác nhận thanh toán.

Search catalog dùng query parameter `query`, đồng bộ giá trị với URL để reload/back/deep-link giữ được kết quả. Khi query thay đổi, cursor pagination phải reset; lỗi search dùng cùng loading/empty/error state của catalog. Search ở header trên page khác chỉ điều hướng về `/products?query=...`, không tạo API search riêng.

### Chính sách thay đổi backend tối thiểu

Khi integration FE–BE phát hiện mismatch, xử lý theo thứ tự:

1. Xác nhận mismatch bằng OpenAPI, route backend, request thực tế và test hiện có.
2. Ưu tiên sửa frontend nếu backend đã đúng contract.
3. Nếu backend thật sự lệch contract, chỉ sửa phần nhỏ nhất: route prefix `/api`, response/request field, error mapping, CORS/config phục vụ môi trường dev hoặc header cần thiết cho session/CSRF.
4. Không đổi business rule, state transition, transaction boundary, inventory/quota invariant, idempotency semantics hoặc error code đã công bố.
5. Mọi patch BE phải có test backend liên quan trước khi dùng làm bằng chứng cho FE integration.

### Vòng kiểm thử FE–BE cho từng lát cắt

Mỗi lát cắt triển khai phải chạy theo vòng sau, không dồn đến cuối phase:

1. Chạy typecheck/lint/unit-render test của frontend.
2. Chạy backend test liên quan và khởi động API với database/test fixture phù hợp.
3. Chạy service/integration test qua HTTP thật từ frontend client, kiểm tra cookie, CSRF, status code, error code và payload.
4. Chạy lại hành trình người dùng trên browser ở desktop và mobile breakpoint khi lát cắt thay đổi layout hoặc interaction.
5. Đối chiếu dữ liệu backend sau thao tác; riêng cart/preview phải chứng minh không đổi inventory, quota hoặc usage.

## 11. Trạng thái và accessibility

Mỗi page/list/interaction cần có loading, empty, error và retry phù hợp. Các trạng thái trong mẫu phải được giữ khi áp dụng:

- Loading: `role="status"`, `aria-busy`, không làm layout nhảy mạnh.
- Error: nguyên nhân + hành động khôi phục; dùng `role="alert"` khi cần thông báo ngay.
- Empty: giải thích vì sao rỗng và CTA tiếp theo.
- Forbidden/not found/unauthenticated: không chỉ redirect im lặng; giữ message và đường quay lại.
- Overload/rate limit: hiển thị thời gian retry nếu backend cung cấp `Retry-After`.
- Form submit lỗi nhiều field: error summary có thể focus, đồng thời giữ lỗi inline.
- Control tối thiểu 40–44px, keyboard navigation, visible focus, label đầy đủ, `aria-current`, `aria-expanded`, `aria-pressed` khi có.
- Không dùng màu là tín hiệu duy nhất; icon/nhãn phải truyền đạt trạng thái.
- Hỗ trợ `prefers-reduced-motion`; countdown không đọc từng giây liên tục cho screen reader.

## 12. Ràng buộc triển khai

- Giữ React 19, Vite, Tailwind 4, React Router và TanStack Query hiện có; chưa thêm UI framework mới.
- Ưu tiên SVG icon nội bộ nhất quán với mẫu; không thêm dependency icon chỉ để thay thế vài icon cơ bản.
- Giữ cấu trúc feature hiện tại; chỉ tạo `src/components/layout` và `src/components/common` làm lớp dùng chung.
- Không ghi đè thư mục `docs/ui-design/`; đây là tài liệu tham chiếu người dùng cung cấp và hiện đang là file chưa track.
- Mọi thay đổi BE phải ở mức tối thiểu, có lý do integration, test hồi quy và không làm thay đổi bất biến hoặc behavior của Phase 03.
- Design này không tự phê duyệt thay đổi nghiệp vụ; route/payload/error contract chỉ được chỉnh khi cần thiết để khớp implementation và phải ghi nhận trong tài liệu/test liên quan.

## 13. Tiêu chí nghiệm thu design foundation

- Các route trong phạm vi dùng đúng một customer shell hoặc admin shell, không còn header/navigation riêng bị lặp ở page.
- Theme token phản ánh đúng palette, typography, radius, spacing và surface của HTML mẫu.
- Component common có variant/state đủ cho page hiện tại: default, hover, focus, disabled, loading, error, empty.
- Customer shell hiển thị đúng trên 390px, 768px và desktop; admin shell có navigation usable trên mobile.
- `/api/v1` không còn trong source/test frontend; mọi service dùng cùng cách ghép base URL.
- Catalog search gửi đúng `query`, reset cursor khi đổi từ khóa và giữ được URL state khi reload/back.
- Catalog public mở được khi chưa đăng nhập; add-to-cart vẫn yêu cầu auth và giữ đường dẫn quay lại.
- Các hành vi hiện tại của cart, preview, flash-sale và admin flash-sale không bị thay đổi ngoài visual/layout.
- Phase 03 vẫn đạt các requirement/test liên quan: `FR-014–020`, `INV-09/10`, `TEST-09/10/20`; preview không reserve và cart không làm đổi inventory/quota/usage.
- Chạy được `bun run typecheck`, `bun run lint`, `bun test` và `bun run build`; backend test liên quan, e2e API và browser smoke được chạy ở cùng checkpoint.
- Các test route/API được cập nhật theo chuẩn `/api`; nếu có patch BE, patch có test hồi quy tương ứng và được ghi rõ trong handoff.

## 14. Các quyết định được chốt sau builder-check

- **Topology dev:** không mở rộng phạm vi sang deployment mới. Frontend phải chạy được với backend origin hiện tại và với Vite proxy; cả hai đều gửi credentials/CSRF đúng contract.
- **Entry point:** `/` trỏ vào catalog public hiện có. Không thêm featured Flash Sale riêng cho đến khi detail route/API thuộc phase triển khai.
- **Admin navigation:** chỉ render navigation item có route và page thật trong phase hiện tại; Products/Tồn kho sẽ được thêm khi feature tương ứng được lập kế hoạch.
- **Phase boundary:** design implementation dừng ở Phase 03; order/payment/order history và các P1 page chỉ được mở sau khi Phase 04/05 có contract và plan riêng.
- **Integration-first:** mỗi lát cắt UI có checkpoint FE–BE riêng; không chờ hoàn tất toàn bộ theme mới kiểm tra API và browser flow.
- **Catalog search:** triển khai ở mức tối thiểu trên contract `GET /api/products?query=...`; không mở rộng thành global search hoặc search cho admin/flash-sale trong phase này.

## 15. Việc còn lại cho bước lập kế hoạch

Sau khi design này được duyệt, bước tiếp theo là tách kế hoạch triển khai thành các lát cắt nhỏ:

1. Chốt baseline Phase 03, route `/api` và test fixture/integration command.
2. Theme tokens và base styles; chạy FE checks và browser smoke tối thiểu.
3. Customer shell và common states/form controls; kiểm tra auth/catalog/flash-sale với BE thật.
4. Admin shell và admin navigation; kiểm tra admin flash-sale list/form/action với BE thật.
5. Refactor cart/preview sang shell và common components; chạy các case line error, price change, quota/limit và no-reservation.
6. Tổng kiểm tra Phase 03, chỉ sửa BE tối thiểu nếu integration còn mismatch, rồi chạy lại build/lint/typecheck/backend tests/e2e/browser.

Mỗi lát cắt cần nêu file ảnh hưởng, requirement/test liên quan và verification command trước khi bắt đầu chỉnh code.
