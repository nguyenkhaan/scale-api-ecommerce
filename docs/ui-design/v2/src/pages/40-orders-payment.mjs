import { vnd, img, ic, P, customerPage, mobileFrame, steps } from "../lib.mjs";

export const group = "Khách hàng · Đơn hàng & thanh toán";

const CSS = ["orders.css"];
const CART = 1;

// ---------- local helpers ----------
const mmss = (m, s, label = `Còn ${m} phút ${s} giây`) =>
  `<span class="ord-timer num" role="timer" aria-label="${label}">${m}:${s}</span>`;

const icx = {
  refresh: `<svg class="icon-sm" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M13.5 8a5.5 5.5 0 1 1-1.6-3.9M13.5 2.5V6H10"/></svg>`,
  check: `<svg class="icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><circle cx="10" cy="10" r="8"/><path d="m6.5 10.2 2.4 2.4 4.6-5"/></svg>`,
  cross: `<svg class="icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" aria-hidden="true"><circle cx="10" cy="10" r="8"/><path d="m7 7 6 6M13 7l-6 6"/></svg>`,
  warn: `<svg class="icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M10 2.8 1.8 17h16.4z"/><path d="M10 8v4M10 14.5v.1"/></svg>`,
  hourglass: `<svg class="icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" aria-hidden="true"><circle cx="10" cy="10" r="8"/><path d="M10 5.5V10l3 2"/></svg>`,
  search: `<svg class="icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" aria-hidden="true"><circle cx="9" cy="9" r="6"/><path d="M13.5 13.5 17 17M7 9h4"/></svg>`,
  ban: `<svg viewBox="0 0 48 48" width="48" height="48" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" aria-hidden="true"><rect x="11" y="21" width="26" height="19" rx="4"/><path d="M17 21v-5a7 7 0 0 1 14 0v5M24 29v4"/></svg>`,
  emptyBox: `<svg viewBox="0 0 96 96" width="96" height="96" fill="none" aria-hidden="true"><rect x="10" y="10" width="76" height="76" rx="38" fill="var(--brand-soft)"/><path d="M48 26 28 36v24l20 10 20-10V36z" stroke="var(--brand)" stroke-width="3" stroke-linejoin="round" fill="#fff"/><path d="m28 36 20 10 20-10M48 46v24" stroke="var(--brand)" stroke-width="3" stroke-linejoin="round"/></svg>`,
};

const clockNote = (text = "Thời gian chính thức theo hệ thống") => `<span class="clock-note">${ic.info}${text}</span>`;

const breadcrumb = (items) =>
  `<nav class="breadcrumb" aria-label="Đường dẫn"><a href="Main.html">Trang chủ</a>${items
    .map(([label, href]) => `${ic.chevron}${href ? `<a href="${href}">${label}</a>` : `<span aria-current="page">${label}</span>`}`)
    .join("")}</nav>`;

/** Tiki-like account sidebar. */
const accountSide = () => `<aside class="ord-side" aria-label="Tài khoản">
  <div class="ord-side-user"><span class="avatar" aria-hidden="true">MA</span><div><small>Tài khoản của</small><b>Nguyễn Minh Anh</b></div></div>
  <nav class="ord-side-nav" aria-label="Mục tài khoản">
    <a href="Auth.html">${ic.user}Thông tin tài khoản</a>
    <a class="active" href="MyOrders.html" aria-current="page">${ic.orders}Đơn của tôi</a>
    <a href="Cart.html">${ic.cart}Giỏ hàng</a>
  </nav>
</aside>`;

// ---------- data ----------
const ORDER_ID = "SS-1010-0042";
const LINES = [
  { p: P.tee, name: "Áo thun basic", meta: "Mã AT-DEN-M · Màu: Đen · Kích cỡ: M", qty: 2, unit: 129000 },
  { p: P.phones, name: "Tai nghe Bluetooth X2", meta: "Mã TN-X2-DEN · Màu: Đen", qty: 1, unit: 590000 },
];
const TOTAL = 848000;

// =====================================================================
// MyOrders
// =====================================================================
const STATUS = {
  pending: ["Chờ thanh toán", "badge-warning"],
  confirmed: ["Đã xác nhận", "badge-success"],
  failed: ["Thanh toán thất bại", "badge-danger"],
  expired: ["Hết hạn", "badge-neutral"],
  cancelled: ["Đã hủy", "badge-neutral"],
};

const ORDERS = [
  {
    id: "SS-1010-0042", at: "20:41, 10/10/2026", count: 3, total: 848000, status: "pending",
    items: [[P.tee, "Áo thun basic", "Đen / M", 2], [P.phones, "Tai nghe Bluetooth X2", "Đen", 1]],
  },
  { id: "SS-1010-0017", at: "20:12, 10/10/2026", count: 2, total: 258000, status: "confirmed", items: [[P.tee, "Áo thun basic", "Trắng / L", 2]] },
  {
    id: "SS-0909-0120", at: "20:31, 09/09/2026", count: 1, total: 299000, status: "expired", recon: true,
    items: [[P.power, "Sạc dự phòng 10.000mAh", "Đen", 1]],
  },
  { id: "SS-0909-0108", at: "20:05, 09/09/2026", count: 1, total: 590000, status: "failed", items: [[P.phones, "Tai nghe Bluetooth X2", "Đen", 1]] },
  { id: "SS-0909-0091", at: "20:02, 09/09/2026", count: 1, total: 129000, status: "expired", items: [[P.tee, "Áo thun basic", "Đen / M", 1]] },
  { id: "SS-0815-0003", at: "14:30, 15/08/2026", count: 1, total: 250000, status: "cancelled", items: [[P.bottle, "Bình giữ nhiệt 500ml", "Bạc", 1]] },
];

const orderTabs = (active = 0) => {
  const tabs = [
    ["Tất cả"],
    ["Chờ thanh toán", 1],
    ["Đã xác nhận"],
    ["Thanh toán thất bại"],
    ["Đã hủy"],
    ["Hết hạn"],
    ["Cần đối soát", 1],
  ];
  return `<nav class="tabs ord-tabs" aria-label="Lọc theo trạng thái">${tabs
    .map(([t, c], i) => `<a class="tab${i === active ? " active" : ""}" href="#"${i === active ? ' aria-current="page"' : ""}>${t}${c ? `<span class="count">${c}</span>` : ""}</a>`)
    .join("")}</nav>`;
};

const orderCard = (o) => {
  const [label, cls] = STATUS[o.status];
  const actions = [];
  if (o.status === "pending") actions.push(`<a class="btn btn-sale btn-sm" href="Payment.html">Thanh toán</a>`);
  if (["failed", "expired", "cancelled"].includes(o.status) && !o.recon) actions.push(`<a class="btn btn-outline-brand btn-sm" href="Cart.html">Mua lại</a>`);
  actions.push(`<a class="btn btn-sm" href="OrderDetail.html">Xem đơn</a>`);
  return `<article class="panel ord-card" aria-labelledby="o-${o.id}">
  <header class="ord-card-head">
    <div class="ord-card-id"><h2 id="o-${o.id}"><a href="OrderDetail.html">${o.id}</a></h2><span class="muted">Đặt lúc ${o.at} · ${o.count} sản phẩm</span></div>
    <div class="ord-card-status"><span class="badge ${cls}">${label}</span>${o.recon ? `<span class="badge badge-warning">Cần đối soát</span>` : ""}</div>
  </header>
  ${
    o.status === "pending"
      ? `<div class="ord-card-deadline">${ic.clock}<span>Còn ${mmss("03", "12")} để thanh toán</span>${clockNote()}</div>`
      : ""
  }
  ${
    o.recon
      ? `<div class="ord-card-deadline ord-card-deadline--recon">${ic.info}<span>Cổng thanh toán báo thành công sau hạn thanh toán. Đơn không được xác nhận lại; khoản tiền đang chờ cửa hàng đối soát.</span></div>`
      : ""
  }
  <ul class="ord-items">${o.items
    .map(([p, name, variant, qty]) => `<li><span class="thumb"><img src="${img(p.img)}" alt=""></span><div><b>${name}</b><small>Phân loại: ${variant}</small></div><span class="ord-qty num">× ${qty}</span></li>`)
    .join("")}</ul>
  <footer class="ord-card-foot">
    <div class="ord-card-total">Tổng tiền: <span class="price">${vnd(o.total)}</span></div>
    <div class="ord-card-actions">${actions.join("")}</div>
  </footer>
</article>`;
};

function myOrdersPage() {
  return customerPage({
    title: "Đơn của tôi – SaleScale",
    desc: "Danh sách đơn hàng của tôi, lọc theo trạng thái.",
    cartCount: CART,
    extraCss: CSS,
    body: `${breadcrumb([["Đơn của tôi"]])}
<div class="ord-layout">
  ${accountSide()}
  <div class="ord-main">
    <h1 class="ord-h1">Đơn của tôi</h1>
    <section class="panel">${orderTabs(0)}</section>
    <div class="ord-list">${ORDERS.map(orderCard).join("")}</div>
    <div class="ord-more"><button class="btn btn-outline-brand" type="button">Tải thêm</button></div>
  </div>
</div>`,
  });
}

function myOrdersEmptyPage() {
  return customerPage({
    title: "Đơn của tôi – trống – SaleScale",
    desc: "Đơn của tôi khi chưa có đơn hàng nào.",
    cartCount: CART,
    extraCss: CSS,
    body: `${breadcrumb([["Đơn của tôi"]])}
<div class="ord-layout">
  ${accountSide()}
  <div class="ord-main">
    <h1 class="ord-h1">Đơn của tôi</h1>
    <section class="panel">${orderTabs(0)}
      <div class="ord-empty" role="status">
        ${icx.emptyBox}
        <h2>Bạn chưa có đơn hàng nào</h2>
        <p>Đơn sẽ hiện ở đây sau khi bạn đặt hàng thành công.</p>
        <a class="btn btn-primary" href="Main.html">Xem sản phẩm</a>
      </div>
    </section>
  </div>
</div>`,
  });
}

// =====================================================================
// OrderDetail
// =====================================================================
const timeline = (state) => {
  const steps =
    state === "overdue"
      ? [
          ["done", "Đặt đơn thành công", "20:41, 10/10/2026 · hàng và suất sale được giữ"],
          ["bad", "Quá hạn thanh toán", "20:46, 10/10/2026"],
          ["now", "Giải phóng hàng", "Hệ thống đang xử lý"],
        ]
      : [
          ["done", "Đặt đơn thành công", "20:41, 10/10/2026 · hàng và suất sale được giữ"],
          ["now", "Chờ thanh toán", "Hạn 20:46, 10/10/2026"],
          ["", "Đã xác nhận", "Sau khi hệ thống nhận thanh toán"],
        ];
  return `<ol class="ord-timeline" aria-label="Tiến trình đơn">${steps
    .map(([s, t, d]) => `<li class="${s}"${s === "now" ? ' aria-current="step"' : ""}><b>${t}</b><small>${d}</small></li>`)
    .join("")}</ol>`;
};

const detailLines = () => `<section class="panel" aria-labelledby="od-lines">
  <div class="panel-head ord-lines-head"><h2 id="od-lines" class="panel-title">Sản phẩm</h2><span class="clock-note">${ic.lock}Thông tin được lưu tại thời điểm đặt đơn, không đổi theo giá hiện tại.</span></div>
  <ul class="ord-lines">${LINES.map(
    (l) => `<li class="ord-line">
    <span class="thumb"><img src="${img(l.p.img)}" alt=""></span>
    <div class="ord-line-info"><b>${l.name}</b><small>${l.meta}</small><span class="badge badge-sale-soft">${ic.bolt}Flash sale</span></div>
    <span class="ord-line-unit num">${l.qty} × ${vnd(l.unit)}</span>
    <span class="ord-line-total num">${vnd(l.qty * l.unit)}</span>
  </li>`
  ).join("")}</ul>
</section>`;

const detailAside = () => `<aside class="ord-aside" aria-label="Thông tin nhận hàng và thanh toán">
  <section class="panel panel-pad" aria-labelledby="od-rcv">
    <h2 id="od-rcv" class="ord-h2">Người nhận</h2>
    <address class="ord-addr"><b>Nguyễn Minh Anh</b><span class="num">0901 234 567</span><span>12 Nguyễn Trãi, Phường Bến Thành, Quận 1, TP. Hồ Chí Minh</span></address>
  </section>
  <section class="panel summary" aria-label="Tổng tiền">
    <div class="summary-row"><span>Tạm tính (3 sản phẩm)</span><span>${vnd(TOTAL)}</span></div>
    <div class="summary-total"><b>Tổng tiền</b><span><span class="price">${vnd(TOTAL)}</span><small>Giá đã chốt khi đặt đơn</small></span></div>
  </section>
</aside>`;

function orderDetailPage(state = "pending") {
  const overdue = state === "overdue";
  const actions = overdue
    ? ""
    : `<div class="ord-od-actions"><a class="btn btn-sale btn-lg" href="Payment.html">Thanh toán</a><a class="btn btn-lg" href="OrderDetail-cancel.html">Hủy đơn</a></div>`;
  const deadline = overdue
    ? `<div class="ord-deadline ord-deadline--over"><div>${ic.clock}<span>Hạn thanh toán: <b>20:46, 10/10/2026</b> · <b>đã quá hạn</b></span></div></div>`
    : `<div class="ord-deadline"><div>${ic.clock}<span>Hạn thanh toán: <b>20:46, 10/10/2026</b> · còn ${mmss("03", "12")}</span></div>${clockNote("Thời gian chính thức theo hệ thống · hạn cố định 5 phút kể từ lúc đặt")}</div>`;
  const overdueNotice = overdue
    ? `<div class="ord-callout ord-callout--warning" role="status">${icx.warn}<div><h2>Đã quá hạn thanh toán, hệ thống đang giải phóng hàng</h2><p>Đơn này không thể thanh toán nữa. Trạng thái đơn sẽ được cập nhật sau ít phút.</p></div></div>`
    : "";
  const dialog =
    state === "cancel"
      ? `<div class="ord-backdrop">
  <div class="ord-dialog" role="dialog" aria-modal="true" aria-labelledby="cancel-title" aria-describedby="cancel-desc">
    <span class="ord-dialog-ico" aria-hidden="true">${icx.warn}</span>
    <h2 id="cancel-title">Hủy đơn ${ORDER_ID}?</h2>
    <p id="cancel-desc">Hàng đang giữ cho đơn này sẽ được trả lại ngay. Bạn không thể hoàn tác thao tác này.</p>
    <div class="ord-dialog-actions"><a class="btn" href="OrderDetail.html" autofocus>Giữ đơn</a><button class="btn ord-btn-danger" type="button">Xác nhận hủy đơn</button></div>
  </div>
</div>`
      : "";
  const page = customerPage({
    title: `Đơn ${ORDER_ID} – SaleScale`,
    desc: "Chi tiết đơn: trạng thái, hạn thanh toán, sản phẩm đã chốt, người nhận và tổng tiền.",
    cartCount: CART,
    extraCss: CSS,
    mobileBar: overdue
      ? ""
      : `<div class="grow"><span class="price">${vnd(TOTAL)}</span><small>Còn 03:12 để thanh toán</small></div><a class="btn" href="OrderDetail-cancel.html">Hủy</a><a class="btn btn-sale" href="Payment.html">Thanh toán</a>`,
    body: `${breadcrumb([["Đơn của tôi", "MyOrders.html"], [`Đơn ${ORDER_ID}`]])}
<div class="ord-layout">
  ${accountSide()}
  <div class="ord-main">
    <section class="panel ord-od-head" aria-labelledby="od-title">
      <div class="ord-od-title">
        <div><h1 id="od-title">Đơn ${ORDER_ID}</h1><span class="muted">Đặt lúc 20:41, 10/10/2026</span></div>
        <span class="badge badge-warning ord-badge-lg">Chờ thanh toán</span>
      </div>
      ${deadline}
      ${timeline(state)}
      ${actions}
    </section>
    ${overdueNotice}
    <div class="ord-od-grid">
      ${detailLines()}
      ${detailAside()}
    </div>
  </div>
</div>`,
  });
  return dialog ? page.replace("</body>", `${dialog}\n</body>`) : page;
}

function orderForbiddenPage() {
  return customerPage({
    title: "Không có quyền xem đơn – SaleScale",
    desc: "Đơn hàng thuộc về một tài khoản khác.",
    cartCount: CART,
    extraCss: CSS,
    body: `${breadcrumb([["Đơn của tôi", "MyOrders.html"], ["Chi tiết đơn"]])}
<div class="ord-layout">
  ${accountSide()}
  <div class="ord-main">
    <section class="panel ord-empty" role="alert">
      <span class="ord-empty-ico ord-empty-ico--danger">${icx.ban}</span>
      <h1>Bạn không có quyền xem đơn này</h1>
      <p>Đơn này thuộc về một tài khoản khác. Kiểm tra lại đường dẫn hoặc mở danh sách đơn của bạn.</p>
      <a class="btn btn-primary" href="MyOrders.html">Về Đơn của tôi</a>
    </section>
  </div>
</div>`,
  });
}

// =====================================================================
// Payment
// =====================================================================
const PAY = {
  pending: {
    title: "SaleScale – Thanh toán mô phỏng",
    timer: true, session: ["Đang chờ", "badge-warning"], order: ["Chờ thanh toán", "badge-warning"], updated: "20:42:47",
    tone: "brand", role: "status", icon: icx.hourglass, head: "Đang chờ thanh toán",
    text: "Hoàn tất thanh toán trước 20:46. Quá hạn, đơn sẽ tự đóng và hàng được trả lại.",
    actions: `<button class="btn btn-sale btn-lg" type="button">Thanh toán ${vnd(TOTAL)}</button><a class="btn btn-lg" href="OrderDetail.html">Xem đơn</a>`,
    event: ["#1 Thành công", "trễ 2.000 ms · chữ ký hợp lệ", "Đã lên lịch 20:43:05", "sched"],
  },
  success: {
    title: "Thanh toán – thành công",
    session: ["Thành công", "badge-success"], order: ["Đã xác nhận", "badge-success"], updated: "20:43:08",
    tone: "success", role: "status", icon: icx.check, head: "Thanh toán thành công",
    text: `Hệ thống đã nhận ${vnd(TOTAL)}. Đơn ${ORDER_ID} đã chuyển sang <b>Đã xác nhận</b>, nghĩa là đơn đã được chốt.`,
    actions: `<a class="btn btn-primary btn-lg" href="OrderDetail.html">Xem đơn</a>`,
    event: ["#1 Thành công", "trễ 2.000 ms · chữ ký hợp lệ", "Đã gửi 20:43:07", "sent"],
  },
  failed: {
    title: "Thanh toán – thất bại",
    session: ["Thất bại", "badge-danger"], order: ["Thanh toán thất bại", "badge-danger"], updated: "20:43:20",
    tone: "danger", role: "alert", icon: icx.cross, head: "Thanh toán thất bại",
    text: `Cổng thanh toán từ chối giao dịch. Đơn ${ORDER_ID} đã đóng và hàng đã được trả lại. Bấm <b>Đặt lại</b> để tạo đơn mới.`,
    actions: `<a class="btn btn-primary btn-lg" href="Preview.html">Đặt lại</a><a class="btn btn-lg" href="OrderDetail.html">Xem đơn</a>`,
    event: ["#1 Thất bại", "trễ 0 ms · chữ ký hợp lệ", "Đã gửi 20:43:19", "sent"],
  },
  late: {
    title: "Thanh toán – thành công nhưng muộn",
    session: ["Thành công", "badge-success"], order: ["Hết hạn", "badge-neutral"], updated: "20:47:03", recon: true,
    tone: "warning", role: "alert", icon: icx.warn, head: "Đơn chưa được xác nhận vì đã hết thời gian giữ hàng.",
    text: "Cổng thanh toán báo thành công lúc 20:47, sau hạn 20:46. Vui lòng liên hệ cửa hàng để được hỗ trợ về khoản tiền này.",
    extra: `<p class="ord-callout-sub">${ic.info}<span>Thanh toán muộn không khôi phục đơn: đơn vẫn ở trạng thái <b>Hết hạn</b> và khoản tiền được chuyển sang <b>cần đối soát</b>.</span></p>`,
    actions: `<a class="btn btn-lg" href="OrderDetail.html">Xem đơn</a>`,
    event: ["#1 Thành công", "trễ 360.000 ms · chữ ký hợp lệ", "Đã gửi 20:47:02", "sent"],
  },
  unknown: {
    title: "Thanh toán – chưa xác định",
    timer: true, session: ["Chưa xác định", "badge-warning"], order: ["Chờ thanh toán", "badge-warning"], updated: "20:43:30",
    tone: "neutral", role: "status", icon: icx.search, head: "Đang kiểm tra, chưa thể kết luận",
    text: "Chưa nhận được kết quả chính thức từ cổng thanh toán. <b>Đừng thanh toán lại</b> để tránh bị trừ tiền hai lần. Trang sẽ tự cập nhật khi có kết quả.",
    actions: `<button class="btn btn-outline-brand btn-lg" type="button">${icx.refresh}Kiểm tra lại</button><a class="btn btn-lg" href="OrderDetail.html">Xem đơn</a>`,
    event: ["#1 Thành công", "không gửi (mô phỏng mất callback)", "Đã bỏ qua", "skip"],
  },
  expired: {
    title: "Thanh toán – đơn hết hạn",
    session: null, order: ["Hết hạn", "badge-neutral"], updated: "20:46:02",
    tone: "neutral", role: "status", icon: icx.hourglass, head: "Đơn đã hết hạn thanh toán",
    text: `Đơn ${ORDER_ID} đã hết thời gian giữ hàng lúc 20:46 nên không thể thanh toán nữa. Hàng đã được trả lại.`,
    actions: `<a class="btn btn-primary btn-lg" href="Cart.html">Mua lại</a><a class="btn btn-lg" href="OrderDetail.html">Xem đơn</a>`,
    event: null,
  },
};

const toolsPanel = (event) => `<aside class="panel ord-tools" aria-labelledby="tools-title">
  <div class="ord-tools-head"><h2 id="tools-title" class="ord-h2">Công cụ thử nghiệm</h2><span class="badge badge-warning">Chỉ dùng trong môi trường thử nghiệm</span></div>
  <p class="ord-tools-desc">Gửi sự kiện giả lập từ cổng thanh toán tới hệ thống để thử các kết quả.</p>
  <fieldset class="ord-fs"><legend>Kết quả</legend>
    <label><input type="radio" name="tool-result" checked>Thành công</label>
    <label><input type="radio" name="tool-result">Thất bại</label>
  </fieldset>
  <div class="ord-fields">
    <div class="ord-field"><label for="tool-delay">Độ trễ (ms)</label><input class="input" id="tool-delay" type="number" min="0" step="100" value="2000"></div>
    <div class="ord-field"><label for="tool-repeat">Số lần gửi lặp</label><input class="input" id="tool-repeat" type="number" min="1" value="1"></div>
  </div>
  <label class="ord-check"><input type="checkbox">Không gửi (mô phỏng mất callback)</label>
  <fieldset class="ord-fs"><legend>Chữ ký</legend>
    <label><input type="radio" name="tool-sign" checked>Hợp lệ</label>
    <label><input type="radio" name="tool-sign">Không hợp lệ</label>
  </fieldset>
  <details class="ord-adv"><summary>Nâng cao</summary>
    <div class="ord-fields">
      <div class="ord-field"><label for="tool-amount">Ghi đè số tiền</label><input class="input" id="tool-amount" type="number" placeholder="848000"></div>
      <div class="ord-field"><label for="tool-currency">Ghi đè tiền tệ</label><input class="input" id="tool-currency" type="text" placeholder="VND"></div>
    </div>
  </details>
  <button class="btn ord-btn-dark btn-block" type="button">Gửi sự kiện</button>
  <div class="ord-events"><h3>Sự kiện đã lên lịch</h3>
    ${
      event
        ? `<ul><li><b>${event[0]}</b><span class="muted">${event[1]}</span><span class="ord-ev ord-ev--${event[3]}">${event[2]}</span></li></ul>`
        : `<p class="muted">Chưa có sự kiện nào.</p>`
    }
  </div>
</aside>`;

function paymentPage(key) {
  const s = PAY[key];
  const pill = (label, [t, cls]) => `<span class="ord-pill">${label}: <span class="badge ${cls}">${t}</span></span>`;
  return customerPage({
    title: s.title,
    desc: "Trang thanh toán mô phỏng: số tiền, hạn thanh toán, phiên và trạng thái chính thức từ hệ thống.",
    cartCount: CART,
    extraCss: CSS,
    body: `${breadcrumb([["Đơn của tôi", "MyOrders.html"], [`Đơn ${ORDER_ID}`, "OrderDetail.html"], ["Thanh toán"]])}
<div class="ord-pay-top"><h1 class="ord-h1">Thanh toán mô phỏng</h1><div class="panel ord-pay-steps">${steps(2)}</div></div>
<div class="ord-pay">
  <div class="ord-pay-main">
    <section class="panel ord-session" aria-labelledby="ses-title">
      <h2 id="ses-title" class="ord-h2">Phiên thanh toán</h2>
      <div class="ord-amount"><span>Số tiền</span><b class="price num">${vnd(TOTAL)}</b><small>Thanh toán bằng VND</small></div>
      <dl class="ord-kv">
        <div><dt>Mã đơn</dt><dd><a href="OrderDetail.html">${ORDER_ID}</a></dd></div>
        <div><dt>Mã tham chiếu phiên</dt><dd class="ord-mono">PAY-7F3K9Q</dd></div>
        <div><dt>Hạn thanh toán</dt><dd><b>20:46, 10/10/2026</b>${s.timer ? ` <span class="ord-left">còn ${mmss("03", "12")}</span>` : ""}</dd></div>
      </dl>
      ${s.timer ? clockNote() : ""}
      <div class="ord-official">
        <span class="ord-official-label">Trạng thái chính thức từ hệ thống</span>
        ${s.session ? pill("Phiên", s.session) : ""}
        ${pill("Đơn", s.order)}
        ${s.recon ? `<span class="badge badge-warning">Cần đối soát</span>` : ""}
        <span class="ord-official-time">${icx.refresh}Tự làm mới mỗi 5 giây · Cập nhật lúc ${s.updated}</span>
      </div>
    </section>
    <div class="ord-callout ord-callout--${s.tone}" role="${s.role}">
      ${s.icon}
      <div>
        <h2>${s.head}</h2>
        <p>${s.text}</p>
        ${s.extra || ""}
        <div class="ord-callout-actions">${s.actions}</div>
      </div>
    </div>
  </div>
  ${toolsPanel(s.event)}
</div>`,
  });
}

export default [
  { file: "MyOrders.html", title: "Đơn của tôi", html: myOrdersPage() },
  { file: "MyOrders-empty.html", title: "Đơn của tôi – trống", html: myOrdersEmptyPage() },
  { file: "MyOrders-mobile.html", title: "Đơn của tôi – điện thoại", html: mobileFrame("MyOrders.html", "Đơn của tôi") },
  { file: "OrderDetail.html", title: "Chi tiết đơn", html: orderDetailPage("pending") },
  { file: "OrderDetail-cancel.html", title: "Chi tiết đơn – xác nhận hủy", html: orderDetailPage("cancel") },
  { file: "OrderDetail-forbidden.html", title: "Chi tiết đơn – không có quyền", html: orderForbiddenPage() },
  { file: "OrderDetail-mobile.html", title: "Chi tiết đơn – điện thoại", html: mobileFrame("OrderDetail.html", "Chi tiết đơn") },
  { file: "OrderDetail-overdue.html", title: "Chi tiết đơn – đã quá hạn", html: orderDetailPage("overdue") },
  { file: "Payment.html", title: "Thanh toán – đang chờ", html: paymentPage("pending") },
  { file: "Payment-expired.html", title: "Thanh toán – đơn hết hạn", html: paymentPage("expired") },
  { file: "Payment-failed.html", title: "Thanh toán – thất bại", html: paymentPage("failed") },
  { file: "Payment-late.html", title: "Thanh toán – thành công muộn", html: paymentPage("late") },
  { file: "Payment-mobile.html", title: "Thanh toán – điện thoại", html: mobileFrame("Payment.html", "Thanh toán") },
  { file: "Payment-success.html", title: "Thanh toán – thành công", html: paymentPage("success") },
  { file: "Payment-unknown.html", title: "Thanh toán – chưa xác định", html: paymentPage("unknown") },
];
