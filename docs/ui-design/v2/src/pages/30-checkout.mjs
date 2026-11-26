import { vnd, img, ic, P, customerPage, mobileFrame, steps } from "../lib.mjs";

export const group = "Khách hàng · Giỏ hàng & đặt đơn";

const CSS = ["checkout.css"];

// Local icons (lib.mjs has no check / alert / refresh / spinner).
const icx = {
  check: `<svg class="icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="m4.5 10.5 3.5 3.5 7.5-8"/></svg>`,
  alert: `<svg class="icon-sm" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" aria-hidden="true"><circle cx="8" cy="8" r="6.2"/><path d="M8 4.6v4.2M8 11.2v.1"/></svg>`,
  refresh: `<svg class="icon-sm" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M13 3.5v3h-3"/><path d="M12.6 6.5A5 5 0 1 0 13 9.5"/></svg>`,
  copy: `<svg class="icon-sm" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true"><rect x="5.5" y="5.5" width="8" height="8" rx="1.5"/><path d="M10.5 5.5v-2a1 1 0 0 0-1-1h-6a1 1 0 0 0-1 1v6a1 1 0 0 0 1 1h2"/></svg>`,
  back: `<svg class="icon-sm" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M10 3.5 5.5 8l4.5 4.5"/></svg>`,
  user: `<svg class="icon-sm" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" aria-hidden="true"><circle cx="8" cy="5.5" r="2.8"/><path d="M2.8 14c0-2.9 2.3-4.5 5.2-4.5s5.2 1.6 5.2 4.5"/></svg>`,
  hourglass: `<svg class="icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M5 2.5h10M5 17.5h10M6 2.5c0 4 8 4.5 8 7.5s-8 3.5-8 7.5M14 2.5c0 4-8 4.5-8 7.5s8 3.5 8 7.5"/></svg>`,
  question: `<svg class="icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" aria-hidden="true"><circle cx="10" cy="10" r="7.5"/><path d="M7.8 7.6a2.3 2.3 0 0 1 4.5.6c0 1.6-2.3 1.9-2.3 3.4M10 14.2v.1"/></svg>`,
  stop: `<svg class="icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" aria-hidden="true"><circle cx="10" cy="10" r="7.5"/><path d="M10 6v5M10 13.8v.1"/></svg>`,
  traffic: `<svg class="icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M10 2.5 18 16.5H2z"/><path d="M10 8v4M10 14.3v.1"/></svg>`,
};
const spinner = `<span class="co-spin" aria-hidden="true"></span>`;

// ======================================================================
// Cart (same structure as the approved pilot Cart.html)
// ======================================================================

/** Copy of the pilot's cartLine (not exported) + `busy` and `error` states. */
function cartLine(p, opts) {
  const { variant, sku, qty, flash, checked = true, maxNote, off, busy, error } = opts;
  const unit = flash ? p.sale : p.base;
  const msgId = `${sku}-msg`;
  return `<div class="cart-line${off ? " line-off" : ""}${busy ? " co-busy" : ""}${error ? " co-line-error" : ""}"${busy ? ' aria-busy="true"' : ""}>
  <input class="check" type="checkbox" ${checked ? "checked" : ""} aria-label="Chọn ${p.short}"${busy ? " disabled" : ""}>
  <div class="line-product">
    <a class="thumb" href="ProductDetail.html"><img src="${img(p.img)}" alt=""></a>
    <div>
      <h3 class="line-name"><a href="ProductDetail.html">${p.name}</a></h3>
      <div class="line-meta">
        <span>Phân loại: ${variant} · Mã ${sku}</span>
        <div class="badges">${flash ? `<span class="badge badge-sale-soft">${ic.bolt}Giá flash sale</span>` : `<span class="badge badge-neutral">Giá thường</span>`}</div>
        <div class="line-mprice price-row" style="display:none"><span class="price" style="font-size:14px">${vnd(unit)}</span>${flash ? `<span class="price-old">${vnd(p.base)}</span>` : ""}</div>
      </div>
    </div>
  </div>
  <div class="line-price"><span class="co-cap">Giá tham khảo</span><span class="price">${vnd(unit)}</span>${flash ? `<span class="price-old">${vnd(p.base)}</span>` : ""}</div>
  <div class="line-qty"><div class="qty${error ? " co-qty-error" : ""}"><button type="button" aria-label="Giảm"${busy ? " disabled" : ""}>−</button><input aria-label="Số lượng ${p.short}" value="${qty}"${busy ? " disabled" : ""}${error ? ` aria-invalid="true" aria-describedby="${msgId}"` : ""}><button type="button" aria-label="Tăng"${maxNote || busy ? " disabled" : ""}>+</button></div>${
    busy ? `<span class="co-updating" role="status">${spinner}Đang cập nhật…</span>` : maxNote ? `<span class="limit-note">${maxNote}</span>` : ""
  }</div>
  <span class="line-total">${busy ? `<span class="co-skel co-skel-total" aria-hidden="true"></span><span class="sr-only">${vnd(unit * qty)}</span>` : vnd(unit * qty)}</span>
  <button class="icon-btn" type="button" aria-label="Xóa ${p.short}"${busy ? " disabled" : ""}>${ic.trash}</button>
  ${error ? `<div class="co-line-msg notice co-notice-danger" role="alert" id="${msgId}">${icx.alert}<span>${error}</span></div>` : ""}
</div>`;
}

function cartPage({ busyTee = false, phonesError = "" } = {}) {
  const subtotal = P.tee.sale * 2 + P.phones.sale * 1;
  const saving = (P.tee.base - P.tee.sale) * 2 + (P.phones.base - P.phones.sale);
  const body = `  <div class="cart-title"><h1>Giỏ hàng</h1><span class="muted">3/20 sản phẩm</span></div>
  <div class="cart-wrap">
    <div>
      <div class="notice notice-info">${ic.lock}<span>Sản phẩm trong giỏ <b>chưa được giữ</b>. Hàng và suất flash sale chỉ được giữ khi bạn đặt đơn thành công; giá có thể thay đổi trước bước đó.</span></div>
      <div class="panel" style="margin-top:12px">
        <div class="cart-cols"><label><input class="check" type="checkbox" aria-label="Chọn tất cả"></label><span>Tất cả (3 sản phẩm)</span><span>Đơn giá</span><span>Số lượng</span><span>Thành tiền</span><button class="icon-btn" type="button" aria-label="Xóa các dòng đã chọn">${ic.trash}</button></div>
      </div>
      <section class="panel cart-group" aria-labelledby="g1">
        <div class="cart-group-head"><input class="check" type="checkbox" checked aria-label="Chọn nhóm flash sale"><span id="g1" style="color:var(--sale);display:inline-flex;gap:6px;align-items:center">${ic.bolt}Flash Sale 10.10</span><span class="badge badge-sale-soft">Kết thúc 22:00</span><span class="clock-note">${ic.info}Theo giờ hệ thống</span></div>
        ${cartLine(P.tee, { variant: "Đen / M", sku: "AT-DEN-M", qty: 2, flash: true, maxNote: "Tối đa 2/tài khoản", busy: busyTee })}
        ${cartLine(P.phones, { variant: "Đen", sku: "TN-X2-DEN", qty: 1, flash: true, maxNote: "Tối đa 1/tài khoản", error: phonesError })}
      </section>
      <section class="panel cart-group" aria-labelledby="g2">
        <div class="cart-group-head"><input class="check" type="checkbox" aria-label="Chọn nhóm giá thường"><span id="g2">Giá thường</span></div>
        ${cartLine(P.bottle, { variant: "Bạc", sku: "BGN-500-BAC", qty: 1, flash: false, checked: false })}
      </section>
    </div>

    <aside class="cart-aside" aria-label="Tóm tắt">
      <div class="panel">${steps(0)}</div>
      <div class="panel summary">
        <div class="summary-row"><span>Tạm tính (2 dòng)</span><span>${vnd(subtotal + saving)}</span></div>
        <div class="summary-row"><span>Giảm giá flash sale</span><span style="color:var(--success)">−${vnd(saving)}</span></div>
        <div class="summary-total"><b>Tổng tham khảo</b><span><span class="price">${vnd(subtotal)}</span><small>Giá cuối cùng hiện ở bước xem trước đơn</small></span></div>
        <a class="btn btn-sale btn-lg btn-block" href="Preview.html">Xem trước đơn (2)</a>
      </div>
      <div class="notice notice-warning">${ic.info}<span>Mỗi lần đặt đơn tối đa 20 dòng. Dòng giá thường chưa chọn sẽ được giữ lại trong giỏ.</span></div>
    </aside>
  </div>`;
  return customerPage({
    title: busyTee ? "Giỏ hàng – đang cập nhật dòng – SaleScale" : phonesError ? "Giỏ hàng – lỗi tại dòng – SaleScale" : "Giỏ hàng – SaleScale",
    desc: busyTee ? "Giỏ hàng khi một dòng đang cập nhật số lượng." : "Giỏ hàng khi cập nhật số lượng một dòng bị từ chối.",
    body,
    cartCount: 3,
    extraCss: CSS,
    mobileBar: `<div class="grow"><span class="price">${vnd(subtotal)}</span><small>2 dòng · tổng tham khảo</small></div><a class="btn btn-sale" href="Preview.html">Xem trước đơn</a>`,
  });
}

function cartEmptyPage() {
  const body = `  <div class="cart-title"><h1>Giỏ hàng</h1><span class="muted">0/20 sản phẩm</span></div>
  <div class="cart-wrap">
    <div>
      <div class="notice notice-info">${ic.lock}<span>Sản phẩm trong giỏ <b>chưa được giữ</b>. Hàng chỉ được giữ khi bạn đặt đơn thành công.</span></div>
      <section class="panel co-empty" aria-labelledby="empty-title" style="margin-top:12px">
        <img class="co-empty-art" src="${img("co-empty-cart")}" alt="">
        <h2 id="empty-title">Giỏ hàng của bạn đang trống</h2>
        <p>Thêm sản phẩm từ danh sách hoặc từ chương trình flash sale.</p>
        <div class="co-actions co-actions-center">
          <a class="btn btn-primary" href="Main.html">Xem sản phẩm</a>
          <a class="btn btn-outline-brand" href="FlashSales.html">${ic.bolt}Xem flash sale</a>
        </div>
      </section>
    </div>
    <aside class="cart-aside" aria-label="Tóm tắt">
      <div class="panel">${steps(0)}</div>
      <div class="panel summary">
        <div class="summary-row"><span>Tạm tính (0 dòng)</span><span>${vnd(0)}</span></div>
        <div class="summary-total"><b>Tổng tham khảo</b><span><span class="price price-plain">${vnd(0)}</span><small>Chọn sản phẩm để xem trước đơn</small></span></div>
        <button class="btn btn-lg btn-block" type="button" disabled>Xem trước đơn</button>
      </div>
    </aside>
  </div>`;
  return customerPage({ title: "Giỏ hàng trống – SaleScale", desc: "Giỏ hàng chưa có sản phẩm nào.", body, cartCount: 0, extraCss: CSS });
}

// ======================================================================
// Preview (order review, Tiki checkout style)
// ======================================================================

const quoteTime = "20:40, 10/10/2026";

/**
 * A confirmed line. `newUnit` marks a changed price; `error` marks a rejected line.
 */
function previewLine(p, { variant, sku, qty, flash, newUnit, error }) {
  const unit = flash ? p.sale : p.base;
  const shown = newUnit ?? unit;
  return `<div class="co-pline${newUnit ? " co-pline-changed" : ""}${error ? " co-pline-error" : ""}" role="row">
  <div class="co-pprod" role="cell">
    <div class="thumb"><img src="${img(p.img)}" alt=""></div>
    <div>
      <b class="co-pname">${p.short}</b>
      <span class="co-pmeta">${variant} · Mã ${sku}</span>
      <span class="co-psrc">${flash ? `<span class="badge badge-sale-soft">${ic.bolt}Flash sale</span>` : `<span class="badge badge-neutral">Giá thường</span>`}${newUnit ? `<span class="badge badge-warning">Giá mới</span>` : ""}</span>
    </div>
  </div>
  <div class="co-punit" role="cell"><span class="co-cap">Đơn giá</span>${newUnit ? `<s class="price-old">${vnd(unit)}</s><span class="co-new">${vnd(newUnit)}</span>` : `<span>${vnd(unit)}</span>${flash ? `<s class="price-old">${vnd(p.base)}</s>` : ""}`}<span class="co-mqty">× ${qty}</span></div>
  <div class="co-pqty" role="cell"><span class="co-cap">SL</span><span class="num">${qty}</span></div>
  <div class="co-ptotal" role="cell"><span class="co-cap">Thành tiền</span>${vnd(shown * qty)}</div>
  ${error ? `<div class="co-line-msg notice co-notice-danger" role="alert">${icx.alert}<span><b>${error}</b> <a class="co-inline-link" href="Cart.html">Sửa trong giỏ</a></span></div>` : ""}
</div>`;
}

function previewGroup(id, flash, lines) {
  return `<div class="co-pgroup" role="rowgroup" aria-labelledby="${id}">
  <div class="co-pgroup-head" role="row"><span role="cell" id="${id}">${flash ? `<span class="co-flash-name">${ic.bolt}Flash Sale 10.10</span><span class="badge badge-sale-soft">Giá flash sale đã xác nhận</span>` : `<span>Giá thường</span>`}</span></div>
  ${lines.join("\n  ")}
</div>`;
}

/**
 * state: "ready" | "submitting" | "price-changed" | "rejected"
 */
function previewPage(state = "ready") {
  const submitting = state === "submitting";
  const changed = state === "price-changed";
  const rejected = state === "rejected";
  const dis = submitting ? " disabled" : "";

  // Data per state
  const lines = changed
    ? [
        ["co-gf", true, [previewLine(P.tee, { variant: "Đen / M", sku: "AT-DEN-M", qty: 2, flash: true })]],
        ["co-gr", false, [previewLine(P.bottle, { variant: "Bạc", sku: "BGN-500-BAC", qty: 1, flash: false, newUnit: 275000 })]],
      ]
    : [
        [
          "co-gf",
          true,
          [
            previewLine(P.tee, { variant: "Đen / M", sku: "AT-DEN-M", qty: 2, flash: true }),
            previewLine(P.phones, { variant: "Đen", sku: "TN-X2-DEN", qty: 1, flash: true, error: rejected ? "Đã hết suất giá flash sale" : "" }),
          ],
        ],
      ];
  const base = changed ? P.tee.base * 2 + P.bottle.base : P.tee.base * 2 + P.phones.base;
  const saving = changed ? (P.tee.base - P.tee.sale) * 2 : (P.tee.base - P.tee.sale) * 2 + (P.phones.base - P.phones.sale);
  const total = changed ? 533000 : 848000;
  const oldTotal = 508000;
  const baseNew = changed ? P.tee.base * 2 + 275000 : base;

  const topNotice = changed
    ? `<div class="notice notice-warning co-top-notice" role="status">${ic.info}<span><b>Giá đã thay đổi</b><br>Giá của Bình giữ nhiệt 500ml đổi từ <span class="co-nw">${vnd(250000)}</span> thành <span class="co-nw">${vnd(275000)}</span>. Tổng tiền mới là <span class="co-nw">${vnd(533000)}</span>. Vui lòng xác nhận giá mới trước khi đặt hàng.</span></div>`
    : rejected
      ? `<div class="notice co-notice-danger co-top-notice" role="alert">${icx.alert}<span><b>Chưa đặt được đơn hàng</b><br>Chưa có đơn nào được tạo. Giỏ hàng của bạn vẫn giữ nguyên. Có 1 dòng cần sửa, xem chi tiết ngay tại dòng bên dưới.</span></div>`
      : "";

  const confirmLabel = changed ? `Tôi đồng ý với giá mới <span class="co-amt">(${vnd(total)})</span>` : `Tôi xác nhận báo giá này <span class="co-amt">(${vnd(total)})</span>`;
  const needHint = changed ? "Bạn cần xác nhận giá mới trước khi đặt hàng." : "Bạn cần xác nhận báo giá trước khi đặt hàng.";
  // Ready / price-changed: unchecked → CSS :has() keeps "Đặt hàng" inactive until ticked (no JS).
  const confirmChecked = submitting || rejected;

  let placeBtn;
  if (submitting) placeBtn = `<button class="btn btn-sale btn-lg btn-block" type="button" disabled aria-busy="true">${spinner}Đang đặt hàng…</button><p class="co-hint co-hint-busy" role="status">Đang gửi yêu cầu. Vui lòng không đóng trang.</p>`;
  else if (rejected) placeBtn = `<button class="btn btn-lg btn-block" type="button" disabled>Đặt hàng</button><p class="co-hint">Sửa dòng bị lỗi trong giỏ rồi xem trước lại.</p>`;
  else placeBtn = `<a class="btn btn-sale btn-lg btn-block co-place" href="OrderResult.html">Đặt hàng</a><p class="co-hint co-need">${needHint}</p>`;

  const body = `  <nav class="breadcrumb" aria-label="Đường dẫn"><a href="Main.html">Trang chủ</a>${ic.chevron}<a href="Cart.html">Giỏ hàng</a>${ic.chevron}<span aria-current="page">Xem trước đơn hàng</span></nav>
  <div class="co-title"><h1>Xem trước đơn hàng</h1><span class="muted">Giá tại thời điểm xem trước ${quoteTime}</span></div>
  ${topNotice}
  <div class="cart-wrap co-wrap">
    <div class="co-main">
      <section class="panel" aria-labelledby="rcv-title">
        <div class="panel-head co-sec-head"><h2 id="rcv-title" class="panel-title">${icx.user}Người nhận</h2></div>
        <div class="co-form">
          <div class="co-field"><label for="rcv-name">Họ tên</label><input class="input" id="rcv-name" type="text" autocomplete="name" value="Nguyễn Minh Anh"${dis}></div>
          <div class="co-field"><label for="rcv-phone">Số điện thoại</label><input class="input" id="rcv-phone" type="tel" autocomplete="tel" value="0901 234 567"${dis}></div>
          <div class="co-field co-field-full"><label for="rcv-addr">Địa chỉ</label><textarea class="input co-textarea" id="rcv-addr" rows="2" autocomplete="street-address"${dis}>12 Nguyễn Trãi, Phường Bến Thành, Quận 1, TP. Hồ Chí Minh</textarea></div>
        </div>
      </section>

      <section class="panel" aria-labelledby="lines-title">
        <div class="panel-head co-sec-head"><h2 id="lines-title" class="panel-title">${ic.box}Sản phẩm (2 dòng)</h2><span class="co-sec-note">${ic.shield}Giá do hệ thống xác nhận lúc ${quoteTime}</span></div>
        <div class="co-ptable" role="table" aria-labelledby="lines-title">
          <div class="co-pcols" role="row"><span role="columnheader">Sản phẩm · nguồn giá</span><span role="columnheader">Đơn giá</span><span role="columnheader">SL</span><span role="columnheader">Thành tiền</span></div>
          ${lines.map(([id, f, ls]) => previewGroup(id, f, ls)).join("\n          ")}
        </div>
      </section>

      <div class="notice notice-info">${ic.lock}<span>Xem trước đơn <b>chưa giữ hàng</b>. Hàng và suất flash sale chỉ được giữ khi bạn bấm Đặt hàng và hệ thống tạo đơn thành công.</span></div>
    </div>

    <aside class="cart-aside" aria-label="Tổng đơn">
      <div class="panel">${steps(1)}</div>
      <div class="panel summary" aria-labelledby="sum-title">
        <h2 id="sum-title" class="panel-title">Tổng đơn</h2>
        <div class="summary-row"><span>Tạm tính (giá gốc)</span><span>${vnd(baseNew)}</span></div>
        <div class="summary-row"><span>Giảm giá flash sale</span><span style="color:var(--success)">−${vnd(saving)}</span></div>
        <div class="summary-row"><span>Phí vận chuyển</span><span>Chưa áp dụng</span></div>
        <div class="summary-total"><b>Tổng tiền</b><span>${changed ? `<s class="price-old co-old-total">${vnd(oldTotal)}</s>` : ""}<span class="price">${vnd(total)}</span><small>Đã gồm giá xác nhận ở bước này</small></span></div>
        <dl class="co-facts">
          <div><dt>${ic.clock}Giá tại thời điểm xem trước</dt><dd class="num">${quoteTime}</dd></div>
          <div><dt>${ic.lock}Giữ hàng sau khi đặt</dt><dd>5 phút</dd></div>
        </dl>
        <label class="co-confirm${changed ? " co-confirm-warn" : ""}"><input class="check" type="checkbox" id="co-ok"${confirmChecked ? " checked" : ""}${submitting || rejected ? " disabled" : ""}><span>${confirmLabel}</span></label>
        ${placeBtn}
        <a class="btn btn-ghost btn-block" href="Cart.html">${icx.back}Quay lại giỏ hàng</a>
      </div>
    </aside>
  </div>`;

  const mobileBar = submitting
    ? `<div class="grow"><span class="price">${vnd(total)}</span><small>Đang gửi yêu cầu…</small></div><button class="btn btn-sale" type="button" disabled aria-busy="true">${spinner}Đang đặt…</button>`
    : rejected
      ? `<div class="grow"><span class="price">${vnd(total)}</span><small>Có 1 dòng cần sửa</small></div><a class="btn btn-outline-brand" href="Cart.html">Sửa trong giỏ</a>`
      : `<div class="grow"><span class="price">${vnd(total)}</span><small>${changed ? "Xác nhận giá mới ở trên" : "Xác nhận báo giá ở trên"}</small></div><a class="btn btn-sale co-place" href="OrderResult.html">Đặt hàng</a>`;

  const titles = {
    ready: "Xem trước đơn hàng – SaleScale",
    submitting: "Xem trước – đang gửi – SaleScale",
    "price-changed": "Xem trước – giá thay đổi – SaleScale",
    rejected: "Xem trước – bị từ chối – SaleScale",
  };
  return customerPage({ title: titles[state], desc: "Xem trước đơn: người nhận, giá đã xác nhận và tổng tiền trước khi đặt hàng.", body, cartCount: 3, extraCss: CSS, mobileBar });
}

// ======================================================================
// Order result
// ======================================================================

function orderSuccessPage() {
  const lines = [
    [P.tee, "Đen / M · Mã AT-DEN-M", 2, P.tee.sale],
    [P.phones, "Đen · Mã TN-X2-DEN", 1, P.phones.sale],
  ];
  const body = `  <div class="co-success-wrap">
    <section class="panel co-success" aria-labelledby="res-title">
      <div class="co-success-head">
        <span class="co-res-icon co-res-success">${icx.check}</span>
        <div>
          <h1 id="res-title">Đặt đơn thành công</h1>
          <p>Mã đơn <b class="num">SS-1010-0042</b> · Đặt lúc <span class="num">20:41, 10/10/2026</span></p>
        </div>
        <span class="badge badge-warning co-status-badge">Chờ thanh toán</span>
      </div>
      <div class="co-deadline" role="timer" aria-label="Còn 4 phút 59 giây để thanh toán">
        <div><span class="co-cap">Hạn thanh toán</span><b class="num">20:46, 10/10/2026</b></div>
        <div class="co-deadline-left"><span class="co-cap">Còn</span><b class="num">04:59</b></div>
        <span class="clock-note">${ic.info}Thời gian chính thức theo hệ thống</span>
      </div>
      <div class="notice notice-warning">${ic.clock}<span>Hàng được giữ <b>5 phút</b> kể từ lúc tạo đơn. Mở lại trang hay thanh toán lại không gia hạn thời gian này.</span></div>
      <div class="co-actions">
        <a class="btn btn-sale btn-lg" href="Payment.html">${ic.card}Thanh toán ngay</a>
        <a class="btn btn-outline-brand btn-lg" href="OrderDetail.html">Xem đơn</a>
      </div>
    </section>

    <aside class="co-success-side" aria-label="Đơn hàng">
      <div class="panel">${steps(2)}</div>
      <section class="panel" aria-labelledby="ord-title">
        <div class="panel-head co-sec-head"><h2 id="ord-title" class="panel-title">${ic.orders}Đơn SS-1010-0042</h2><span class="badge badge-sale-soft">${ic.bolt}Flash Sale 10.10</span></div>
        <ul class="co-olist">
          ${lines
            .map(
              ([p, meta, q, u]) => `<li><div class="thumb"><img src="${img(p.img)}" alt=""></div><div><b>${p.short}</b><small>${meta}</small><small><span class="badge badge-sale-soft">Flash sale</span> <span class="num">${q} × ${vnd(u)}</span></small></div><span class="num co-oamt">${vnd(q * u)}</span></li>`
            )
            .join("\n          ")}
        </ul>
        <div class="summary co-osum"><div class="summary-total"><b>Tổng tiền</b><span class="price">${vnd(848000)}</span></div></div>
      </section>
    </aside>
  </div>`;
  return customerPage({
    title: "Kết quả đặt đơn – SaleScale",
    desc: "Đơn đã được tạo (201), đang chờ thanh toán trong 5 phút.",
    body,
    cartCount: 1,
    extraCss: CSS,
    mobileBar: `<div class="grow"><span class="price">${vnd(848000)}</span><small>Còn 04:59 để thanh toán</small></div><a class="btn btn-sale" href="Payment.html">Thanh toán ngay</a>`,
  });
}

/** Non-success result: a single honest status card inside the normal page frame. */
function resultState({ title, desc, tone, icon, role = "status", badge, heading, text, extra = "", actions }) {
  const body = `  <div class="co-res-wrap">
    <div class="panel">${steps(1)}</div>
    <section class="panel co-res co-res-${tone}" role="${role}" aria-labelledby="res-title">
      <span class="co-res-icon co-res-${tone}">${icon}</span>
      <span class="badge ${badge[0]}">${badge[1]}</span>
      <h1 id="res-title">${heading}</h1>
      <p class="co-res-text">${text}</p>
      ${extra}
      <div class="co-actions co-actions-center">${actions}</div>
    </section>
    <p class="co-res-foot">${ic.shield}Chỉ khi hệ thống xác nhận đã tạo đơn, bạn mới thấy mã đơn và hạn thanh toán.</p>
  </div>`;
  return customerPage({ title, desc, body, cartCount: 3, extraCss: CSS });
}

const orderProcessing = () =>
  resultState({
    title: "Kết quả – yêu cầu trước đang xử lý – SaleScale",
    desc: "Yêu cầu đặt đơn trước đó vẫn đang được xử lý.",
    tone: "info",
    icon: icx.hourglass,
    badge: ["badge-brand", "Đang xử lý · chưa có kết quả"],
    heading: "Yêu cầu trước đó đang được xử lý, vui lòng đợi.",
    text: "Bạn vừa gửi yêu cầu đặt đơn này. Kết quả sẽ hiện ngay khi hệ thống xử lý xong.",
    extra: `<div class="co-res-line">${spinner}<span>Đang chờ kết quả từ hệ thống</span></div>`,
    actions: `<button class="btn btn-primary" type="button">${icx.refresh}Kiểm tra lại</button>`,
  });

const orderUnknown = () =>
  resultState({
    title: "Kết quả – chưa rõ – SaleScale",
    desc: "Chưa nhận được kết quả đặt đơn do kết nối gián đoạn.",
    tone: "info",
    icon: icx.question,
    badge: ["badge-neutral", "Chưa rõ kết quả · chưa phải đặt thành công"],
    heading: "Đang kiểm tra kết quả đặt đơn…",
    text: "Kết nối bị gián đoạn nên chưa nhận được kết quả. Hệ thống đang tự kiểm tra lại, bạn không cần đặt lại đơn.",
    extra: `<div class="co-res-line">${spinner}<span>Tự kiểm tra lại sau <b class="num">5</b> giây</span></div>
      <div class="notice notice-info co-res-note">${ic.info}<span><b>Thử lại</b> gửi lại đúng yêu cầu cũ (cùng mã yêu cầu), nên không tạo thêm đơn trùng.</span></div>`,
    actions: `<button class="btn btn-primary" type="button">${icx.refresh}Kiểm tra kết quả</button><button class="btn btn-outline-brand" type="button">Thử lại</button>`,
  });

const orderOverload = () =>
  resultState({
    title: "Kết quả – quá tải – SaleScale",
    desc: "Hệ thống đang quá tải, chưa có đơn nào được tạo.",
    tone: "warning",
    icon: icx.traffic,
    badge: ["badge-warning", "Quá tải · chưa tạo đơn"],
    heading: "Hệ thống đang quá tải",
    text: "Chưa có đơn nào được tạo. Giỏ hàng của bạn vẫn giữ nguyên. Nút thử lại sẽ bật khi hết thời gian chờ.",
    extra: `<div class="co-wait"><span class="co-cap">Thử lại sau</span><span class="countdown num" role="timer" aria-label="Còn 15 giây"><span>00</span>:<span>15</span></span><small>Thời gian chờ do hệ thống trả về</small></div>`,
    actions: `<button class="btn btn-primary" type="button" disabled>Thử lại sau 00:15</button><a class="btn btn-outline-brand" href="Cart.html">Về giỏ hàng</a>`,
  });

const orderError = () =>
  resultState({
    title: "Kết quả – lỗi hệ thống – SaleScale",
    desc: "Lỗi hệ thống khi đặt đơn, chưa xác định được đơn đã tạo hay chưa.",
    tone: "danger",
    role: "alert",
    icon: icx.stop,
    badge: ["badge-danger", "Lỗi hệ thống · chưa xác định"],
    heading: "Hệ thống tạm thời không xử lý được",
    text: "Chưa xác định được đơn của bạn đã được tạo hay chưa. Bấm <b>Kiểm tra lại</b> để xem kết quả, không cần đặt lại.",
    extra: `<div class="co-reqid"><span>Mã yêu cầu: <code class="num">req_8f2c1a7b9d</code></span><button class="btn btn-sm" type="button">${icx.copy}Sao chép</button></div>`,
    actions: `<button class="btn btn-primary" type="button">${icx.refresh}Kiểm tra lại</button>`,
  });

export default [
  { file: "Cart-empty.html", title: "Giỏ hàng trống", html: cartEmptyPage() },
  { file: "Cart-mobile.html", title: "Giỏ hàng – điện thoại", html: mobileFrame("Cart.html", "Giỏ hàng") },
  { file: "Cart-row-error.html", title: "Giỏ hàng – lỗi tại dòng", html: cartPage({ phonesError: "Không cập nhật được số lượng: vượt giới hạn 1 sản phẩm/tài khoản. Số lượng vẫn là 1." }) },
  { file: "Cart-updating.html", title: "Giỏ hàng – đang cập nhật", html: cartPage({ busyTee: true }) },
  { file: "Preview.html", title: "Xem trước đơn", html: previewPage("ready") },
  { file: "Preview-mobile.html", title: "Xem trước – điện thoại", html: mobileFrame("Preview.html", "Xem trước đơn") },
  { file: "Preview-price-changed.html", title: "Xem trước – giá thay đổi", html: previewPage("price-changed") },
  { file: "Preview-rejected.html", title: "Xem trước – bị từ chối", html: previewPage("rejected") },
  { file: "Preview-submitting.html", title: "Xem trước – đang gửi", html: previewPage("submitting") },
  { file: "OrderResult.html", title: "Đặt đơn thành công (201)", html: orderSuccessPage() },
  { file: "OrderResult-error.html", title: "Kết quả – lỗi hệ thống", html: orderError() },
  { file: "OrderResult-mobile.html", title: "Kết quả – điện thoại", html: mobileFrame("OrderResult.html", "Kết quả đặt đơn") },
  { file: "OrderResult-overload.html", title: "Kết quả – quá tải", html: orderOverload() },
  { file: "OrderResult-processing.html", title: "Kết quả – đang xử lý", html: orderProcessing() },
  { file: "OrderResult-unknown.html", title: "Kết quả – chưa rõ", html: orderUnknown() },
];
