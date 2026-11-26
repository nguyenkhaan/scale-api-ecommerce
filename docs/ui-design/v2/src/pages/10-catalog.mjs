import { vnd, pct, compact, img, ic, P, quotaBar, flashItem, productCard, head, header, footer, countdown, mobileFrame } from "../lib.mjs";

export const group = "Khách hàng · Sản phẩm & flash sale";

const CSS = ["catalog.css"];

// ---------- local helpers (lib.mjs is shared and read-only) ----------
const icCheck = `<svg class="icon-sm" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><circle cx="8" cy="8" r="6.2"/><path d="m5.4 8.2 1.8 1.8 3.4-3.6"/></svg>`;
const icAlert = `<svg class="icon-sm" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" aria-hidden="true"><circle cx="8" cy="8" r="6.2"/><path d="M8 4.6v4.2M8 11.2v.1"/></svg>`;
const icRetry = `<svg class="icon-sm" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M13 8a5 5 0 1 1-1.5-3.6M13 2.5v2.8h-2.8"/></svg>`;
const icBack = `<svg class="icon-sm" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M10 3.5 5.5 8l4.5 4.5"/></svg>`;
const icCal = `<svg class="icon-sm" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" aria-hidden="true"><rect x="2.5" y="3.5" width="11" height="10" rx="1.5"/><path d="M2.5 6.5h11M5.5 2v3M10.5 2v3"/></svg>`;
const icPause = `<svg class="icon-sm" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" aria-hidden="true"><circle cx="8" cy="8" r="6.2"/><path d="M6.5 5.5v5M9.5 5.5v5"/></svg>`;

/** lib header() plus: no cart badge when cartCount is null, optional search value, optional active sub-nav link. */
function hdr({ cartCount = 3, loggedIn = true, q = "", flashActive = false } = {}) {
  let h = header(cartCount ?? 0, loggedIn);
  if (cartCount === null) {
    h = h.replace(`<span class="cart-count">0</span>`, "").replace(`aria-label="Giỏ hàng, 0 sản phẩm"`, `aria-label="Giỏ hàng"`);
  }
  if (cartCount >= 10) h = h.replace(`<span class="cart-count">`, `<span class="cart-count cat-count-2d">`);
  if (q) h = h.replace(`<input id="q" name="q"`, `<input id="q" name="q" value="${q}"`);
  if (flashActive) h = h.replace(`<a class="flash-link" href="FlashSales.html">`, `<a class="flash-link" href="FlashSales.html" aria-current="page">`);
  return h;
}

/** Customer page with a custom header (customerPage() always renders the default header). */
const page = ({ title, desc, body, headerHtml = hdr(), mobileBar = "", mainClass = "container page" }) => `${head(title, "blue", desc, CSS)}
<body${mobileBar ? ' class="has-mobile-bar"' : ""}>
${headerHtml}
<main class="${mainClass}">
${body}
</main>
${footer}
${mobileBar ? `<div class="mobile-bar">${mobileBar}</div>` : ""}
</body>
</html>
`;


// ======================================================================
// Main (trang chủ / danh sách sản phẩm) – state variants of the pilot
// ======================================================================
const cats = [
  ["cat-sale", "Đang flash sale", 6],
  ["cat-fashion", "Thời trang", 38],
  ["cat-electronics", "Điện tử", 52],
  ["cat-home", "Gia dụng", 27],
  ["cat-beauty", "Làm đẹp", 19],
  ["cat-accessory", "Phụ kiện", 24],
];

const mainSide = () => `<aside class="side" aria-label="Danh mục">
    <div class="panel">
      <h3>Danh mục</h3>
      <ul class="side-list">${cats.map(([i, n, c]) => `<li><a href="#"><img class="cat-ico" src="${img(i)}" alt="">${n}<span class="count">${c}</span></a></li>`).join("")}</ul>
    </div>
    <div class="panel">
      <h3>Lịch flash sale hôm nay</h3>
      <ul class="side-list">
        <li><a href="FlashSaleDetail-ended.html"><span>12:00 – 14:00</span><span class="count">Đã kết thúc</span></a></li>
        <li><a href="FlashSaleDetail.html" class="active"><span>20:00 – 22:00</span><span class="count" style="color:var(--sale);font-weight:600">Đang diễn ra</span></a></li>
        <li><a href="FlashSaleDetail-upcoming.html"><span>Mai 09:00</span><span class="count">Sắp diễn ra</span></a></li>
      </ul>
    </div>
  </aside>`;

const mainPromo = () => `<section class="promo-row" aria-label="Chương trình nổi bật">
      <a class="promo promo-sale" href="FlashSaleDetail.html">
        <div>
          <span class="badge" style="background:rgba(255,255,255,.2);color:#fff">${ic.bolt}Đang diễn ra · 20:00 – 22:00</span>
          <h2 style="margin-top:12px">Flash Sale 10.10 giảm đến 45%</h2>
          <p>6 sản phẩm, số suất có hạn. Giá được giữ khi bạn đặt đơn thành công.</p>
        </div>
        <div class="promo-stats"><div><b class="num">01:24:30</b><span>Kết thúc sau</span></div><div><b class="num">124/275</b><span>Suất đã bán</span></div></div>
        <img class="art" src="${img("flash-hero")}" alt="">
      </a>
      <div class="promo-side">
        <a class="promo-mini" href="FlashSaleDetail-upcoming.html"><div class="thumb"><img src="${img("airfryer")}" alt=""></div><div><span class="badge badge-brand">Sắp diễn ra</span><h3 style="margin-top:6px">Gia dụng cuối tuần</h3><p>Bắt đầu 09:00 ngày mai · 8 sản phẩm</p></div></a>
        <a class="promo-mini" href="Main.html"><div class="thumb"><img src="${img("sunscreen")}" alt=""></div><div><span class="badge badge-neutral">Mới về</span><h3 style="margin-top:6px">Chăm sóc da mùa nắng</h3><p>19 sản phẩm làm đẹp</p></div></a>
      </div>
    </section>`;

const mainFlash = () => `<section class="flash" aria-labelledby="flash-title">
      <div class="flash-head">
        <h2 id="flash-title" class="flash-title">${ic.bolt}Giá sốc hôm nay</h2>
        <div class="flash-meta">Kết thúc sau ${countdown("01", "24", "30")}</div>
        <span class="clock-note">${ic.info}Thời gian chính thức theo hệ thống</span>
        <a class="link-more" href="FlashSaleDetail.html">Xem chương trình${ic.chevron}</a>
      </div>
      <div class="flash-row">${[P.tee, P.phones, P.power, P.fryer, P.sun, P.cable].map(flashItem).join("")}</div>
    </section>`;

const catalogBar = (heading, extra = "") => `<section class="panel" aria-labelledby="all-title">
      <div class="catalog-bar">
        <div class="catalog-top"><h2 id="all-title">${heading}</h2>${extra}<label class="sort">Sắp xếp<select class="select"><option>Phổ biến</option><option>Bán chạy</option><option>Giá thấp đến cao</option><option>Giá cao đến thấp</option></select></label></div>
        <nav class="cat-chips" aria-label="Lọc theo danh mục">
          <a class="chip on" href="Main.html" aria-current="true">Tất cả</a>
          <a class="chip" href="#">${ic.bolt}Đang flash sale</a>
          <a class="chip" href="#">Thời trang</a>
          <a class="chip" href="#">Điện tử</a>
          <a class="chip" href="#">Gia dụng</a>
          <a class="chip" href="#">Làm đẹp</a>
        </nav>
      </div>
    </section>`;

const skelCard = () => `<div class="card cat-skel-card" aria-hidden="true">
  <div class="thumb cat-skel"></div>
  <div class="card-body">
    <span class="cat-skel cat-skel-line" style="width:40%"></span>
    <span class="cat-skel cat-skel-line"></span>
    <span class="cat-skel cat-skel-line" style="width:75%"></span>
    <span class="cat-skel cat-skel-line" style="width:55%;height:18px"></span>
    <span class="cat-skel cat-skel-line" style="width:35%"></span>
  </div>
</div>`;

function mainState(state) {
  let heading = "Tất cả sản phẩm";
  let extra = "";
  let content = "";
  let q = "";
  let title = "";
  let desc = "";
  if (state === "empty") {
    q = "áo khoác";
    title = "Không có kết quả – SaleScale";
    desc = "Danh sách sản phẩm khi tìm kiếm không có kết quả.";
    heading = `Kết quả cho “áo khoác”`;
    extra = `<span class="muted cat-count">0 sản phẩm</span>`;
    content = `<section class="panel cat-state" aria-labelledby="empty-title">
      <img src="${img("cat-empty-search")}" alt="" width="160" height="120">
      <h3 id="empty-title">Không tìm thấy sản phẩm phù hợp</h3>
      <p>Không có sản phẩm nào có tên chứa “áo khoác”. Thử từ khoá khác hoặc xem tất cả sản phẩm.</p>
      <a class="btn btn-primary" href="Main.html">Xem tất cả sản phẩm</a>
    </section>`;
  } else if (state === "error") {
    title = "Lỗi tải sản phẩm – SaleScale";
    desc = "Danh sách sản phẩm khi không tải được dữ liệu.";
    content = `<section class="panel cat-state" role="alert" aria-labelledby="err-title">
      <img src="${img("cat-offline")}" alt="" width="160" height="120">
      <h3 id="err-title">Không tải được danh sách sản phẩm</h3>
      <p>Mất kết nối mạng hoặc máy chủ chưa phản hồi. Kiểm tra kết nối rồi thử lại.</p>
      <a class="btn btn-primary" href="Main.html">${icRetry}Thử lại</a>
    </section>`;
  } else {
    title = "Đang tải sản phẩm – SaleScale";
    desc = "Danh sách sản phẩm đang tải, hiển thị khung chờ.";
    extra = `<span class="muted cat-count" role="status">Đang tải…</span>`;
    content = `<div class="grid-products" aria-busy="true" aria-label="Đang tải danh sách sản phẩm">${Array.from({ length: 10 }, skelCard).join("")}</div>`;
  }
  return page({
    title,
    desc,
    headerHtml: hdr({ q }),
    mainClass: "container home page",
    body: `  ${mainSide()}
  <div class="home-main">
    ${state === "empty" ? "" : mainPromo()}
    ${mainFlash()}
    ${catalogBar(heading, extra)}
    ${content}
  </div>`,
  });
}

// ======================================================================
// ProductDetail – state variants of the pilot
// ======================================================================
function productDetail(state) {
  const p = P.tee;
  const left = p.quota[1] - p.quota[0];
  const loggedIn = state !== "login";
  const cartCount = state === "added" ? 4 : state === "cart-full" ? 20 : state === "login" ? null : 3;
  const qty = state === "over-limit" ? 3 : 1;
  const overLimit = state === "over-limit";
  const similar = [P.linen, P.sandal, P.phones, P.bottle, P.clean, P.lamp, P.cable];

  const titles = {
    added: ["Đã thêm vào giỏ – Áo thun cotton basic – SaleScale", "Chi tiết sản phẩm sau khi thêm vào giỏ thành công."],
    "cart-full": ["Giỏ đã đủ 20 sản phẩm – Áo thun cotton basic – SaleScale", "Chi tiết sản phẩm khi giỏ hàng đã đủ 20 SKU."],
    login: ["Cần đăng nhập – Áo thun cotton basic – SaleScale", "Chi tiết sản phẩm khi khách chưa đăng nhập bấm thêm vào giỏ."],
    "over-limit": ["Vượt số lượng tối đa – Áo thun cotton basic – SaleScale", "Chi tiết sản phẩm khi số lượng vượt giới hạn flash sale mỗi tài khoản."],
  };

  const toast =
    state === "added"
      ? `<div class="cat-toast" role="status">
    <div class="thumb"><img src="${img("tshirt-black")}" alt=""></div>
    <div class="cat-toast-text"><b>${icCheck}Đã thêm vào giỏ</b><span>Áo thun basic · Đen / M · 1 sản phẩm · <span class="cat-sale-text">Giá flash sale</span></span><small>Giỏ hàng chưa giữ hàng. Suất sale chỉ được giữ khi đặt đơn thành công.</small></div>
    <a class="btn btn-primary" href="Cart.html">Xem giỏ hàng</a>
  </div>`
      : "";

  const buyMsg = {
    "cart-full": `<div class="notice cat-notice-danger" role="alert">${icAlert}<div><b>Chưa thêm được: giỏ hàng đã đủ 20 sản phẩm</b><p>Giỏ chứa tối đa 20 SKU khác nhau. Xoá bớt sản phẩm trong giỏ rồi thêm lại.</p><a class="cat-notice-link" href="Cart.html">Mở giỏ hàng${ic.chevron}</a></div></div>`,
    login: `<div class="notice notice-info" role="alert">${ic.user}<div><b>Bạn cần đăng nhập để thêm vào giỏ</b><p>Sau khi đăng nhập, bạn sẽ quay lại đúng trang này với lựa chọn Đen / M, giá flash sale, số lượng 1.</p><a class="btn btn-primary btn-sm" href="Auth.html">Đăng nhập</a></div></div>`,
  }[state] || "";

  const addBtn = overLimit
    ? `<button class="btn btn-sale btn-lg btn-block" type="button" disabled aria-describedby="qty-err">Thêm vào giỏ · Giá flash sale</button>`
    : state === "login"
      ? `<a class="btn btn-sale btn-lg btn-block" href="Auth.html">Thêm vào giỏ · Giá flash sale</a>`
      : `<a class="btn btn-sale btn-lg btn-block" href="ProductDetail-added.html">Thêm vào giỏ · Giá flash sale</a>`;

  const mobileNote = {
    added: `<small class="cat-ok">Đã thêm vào giỏ · Đen / M</small>`,
    "cart-full": `<small class="cat-err">Giỏ đã đủ 20 sản phẩm</small>`,
    login: `<small>Cần đăng nhập để thêm vào giỏ</small>`,
    "over-limit": `<small class="cat-err">Tối đa ${p.limit} sản phẩm/tài khoản</small>`,
  }[state];
  const mobileBtn = {
    added: `<a class="btn btn-primary" href="Cart.html">Xem giỏ hàng</a>`,
    "cart-full": `<a class="btn btn-primary" href="Cart.html">Mở giỏ hàng</a>`,
    login: `<a class="btn btn-primary" href="Auth.html">Đăng nhập</a>`,
    "over-limit": `<button class="btn btn-sale" type="button" disabled>Thêm vào giỏ</button>`,
  }[state];

  return page({
    title: titles[state][0],
    desc: titles[state][1],
    headerHtml: hdr({ cartCount, loggedIn }),
    mobileBar: `<div class="grow"><span class="price">${vnd(p.sale * qty)}</span>${mobileNote}</div>${mobileBtn}`,
    body: `  ${toast}
  <nav class="breadcrumb" aria-label="Đường dẫn"><a href="Main.html">Trang chủ</a>${ic.chevron}<a href="Main.html?cat=thoi-trang">Thời trang</a>${ic.chevron}<span aria-current="page">Áo thun basic</span></nav>
  <div class="pd">
    <section class="panel pd-gallery" aria-label="Hình sản phẩm">
      <div class="thumb"><img src="${img("tshirt-black")}" alt="Áo thun basic màu đen"><span class="badge badge-sale">-${pct(p.sale, p.base)}%</span></div>
      <div class="pd-thumbs">
        <button type="button" aria-pressed="true" aria-label="Ảnh 1"><img src="${img("tshirt-black")}" alt=""></button>
        <button type="button" aria-pressed="false" aria-label="Ảnh 2"><img src="${img("tshirt-white")}" alt=""></button>
        <button type="button" aria-pressed="false" aria-label="Ảnh 3"><img src="${img("tshirt-navy")}" alt=""></button>
        <button type="button" aria-pressed="false" aria-label="Ảnh 4"><img src="${img("tshirt-fold")}" alt=""></button>
        <button type="button" aria-pressed="false" aria-label="Ảnh 5"><img src="${img("tshirt-detail")}" alt=""></button>
      </div>
      <div class="pd-gallery-foot">${ic.info}Ảnh minh họa, màu thực tế có thể chênh lệch nhẹ.</div>
    </section>

    <div class="pd-info">
      <section class="panel panel-pad">
        <div class="pd-brand">Thương hiệu: <a href="#">${p.brand}</a> · Mã SKU: <span class="num">AT-DEN-M</span></div>
        <h1 class="pd-title">${p.name}</h1>
        <div class="pd-stats"><b>${p.rating}</b><span class="stars" style="--r:${p.rating}" aria-hidden="true"></span><span>(${p.reviews.toLocaleString("vi-VN").replace(/,/g, ".")} đánh giá)</span><span class="sep"></span><span>Đã bán <b>${compact(p.sold)}</b></span></div>
      </section>

      <section class="pd-flash" aria-labelledby="pd-flash-title">
        <div class="pd-flash-head">
          <span id="pd-flash-title" class="flash-title">${ic.bolt}Flash Sale 10.10</span>
          <span style="display:inline-flex;align-items:center;gap:8px;font-size:13px">Kết thúc sau ${countdown("01", "24", "30")}</span>
        </div>
        <div class="pd-flash-body">
          <div class="pd-price"><span class="price">${vnd(p.sale)}</span><span class="price-old" style="font-size:14px">${vnd(p.base)}</span><span class="discount">-${pct(p.sale, p.base)}%</span></div>
          <div class="pd-quota">${quotaBar(p.quota)}<span>Tối đa <b>${p.limit}</b> sản phẩm/tài khoản</span></div>
          <span class="clock-note">${ic.info}20:00 – 22:00, 10/10/2026 · thời gian chính thức theo hệ thống</span>
        </div>
      </section>

      <section class="panel panel-pad" aria-label="Chọn phân loại">
        <div class="opt-group">
          <span class="opt-label">Màu: <b>Đen</b></span>
          <div class="opts">
            <button class="opt" type="button" aria-pressed="true"><img src="${img("tshirt-black")}" alt="">Đen</button>
            <button class="opt" type="button" aria-pressed="false"><img src="${img("tshirt-white")}" alt="">Trắng</button>
            <button class="opt" type="button" aria-pressed="false"><img src="${img("tshirt-navy")}" alt="">Xanh than</button>
          </div>
        </div>
        <div class="opt-group">
          <span class="opt-label">Kích cỡ: <b>M</b></span>
          <div class="opts">
            <button class="opt" type="button" aria-pressed="false">S</button>
            <button class="opt" type="button" aria-pressed="true">M</button>
            <button class="opt" type="button" aria-pressed="false">L</button>
            <button class="opt" type="button" disabled title="Tạm hết hàng">XL</button>
          </div>
        </div>
      </section>

      <section class="panel panel-pad">
        <h2 class="pd-section-title">Thông tin chi tiết</h2>
        <table class="spec"><tbody>
          <tr><th scope="row">Thương hiệu</th><td>${p.brand}</td></tr>
          <tr><th scope="row">Danh mục</th><td>Thời trang › Áo thun</td></tr>
          <tr><th scope="row">Chất liệu</th><td>95% cotton, 5% spandex</td></tr>
          <tr><th scope="row">Kiểu dáng</th><td>Regular fit, cổ tròn bo viền</td></tr>
          <tr><th scope="row">Phân loại</th><td>3 màu · 4 kích cỡ (12 SKU)</td></tr>
        </tbody></table>
      </section>

      <section class="panel panel-pad pd-desc">
        <h2 class="pd-section-title">Mô tả sản phẩm</h2>
        <p>Áo thun cotton co giãn nhẹ, form regular, cổ tròn bo viền. Phù hợp mặc hằng ngày, đi làm hay đi chơi.</p>
        <ul><li>Vải dày vừa, thấm hút mồ hôi tốt.</li><li>Giặt máy ở nhiệt độ thường, không dùng chất tẩy.</li><li>Phơi trong bóng râm để giữ màu.</li></ul>
      </section>
    </div>

    <div class="buybox-col">
      <aside class="panel buybox" aria-label="Mua hàng">
        <div class="buy-variant"><div class="thumb"><img src="${img("tshirt-black")}" alt=""></div><div><b>Đen, M</b><small>Mã SKU AT-DEN-M · Còn hàng</small></div></div>
        <fieldset class="source" style="border:0;margin:0;padding:0">
          <legend class="opt-label" style="margin-bottom:8px">Mua theo</legend>
          <label><input type="radio" name="src" checked><span>Giá flash sale<small>Còn ${left} suất · tối đa ${p.limit} sản phẩm/tài khoản</small></span><span class="amt" style="color:var(--price)">${vnd(p.sale)}</span></label>
          <label><input type="radio" name="src"><span>Giá thường<small>Không áp dụng ưu đãi</small></span><span class="amt">${vnd(p.base)}</span></label>
        </fieldset>
        <div class="cat-qty-wrap">
          <div class="buy-qty"><label class="opt-label" for="qty">Số lượng</label><div class="qty${overLimit ? " cat-qty-invalid" : ""}"><button type="button" aria-label="Giảm"${qty === 1 ? " disabled" : ""}>−</button><input id="qty" value="${qty}" inputmode="numeric"${overLimit ? ' aria-invalid="true" aria-describedby="qty-err"' : ""}><button type="button" aria-label="Tăng"${overLimit ? " disabled" : ""}>+</button></div></div>
          ${overLimit ? `<p class="cat-field-err" id="qty-err" role="alert">${icAlert}<span>Vượt giới hạn ${p.limit} sản phẩm/tài khoản. Giảm số lượng hoặc chọn mua theo giá thường.</span></p>` : ""}
        </div>
        <div class="buy-total"><span>Tạm tính</span><span class="price">${vnd(p.sale * qty)}</span></div>
        <div class="buy-actions">
          ${addBtn}
          ${state === "added" ? `<a class="btn btn-outline-brand btn-block" href="Cart.html">Xem giỏ hàng (4)</a>` : `<a class="btn btn-outline-brand btn-block" href="Cart.html">Xem giỏ hàng</a>`}
        </div>
        ${buyMsg}
        <div class="trust">
          <div>${ic.lock}<span>Thêm vào giỏ <b>chưa giữ hàng</b>. Suất sale chỉ được giữ khi đặt đơn thành công.</span></div>
          <div>${ic.clock}<span>Sau khi đặt, bạn có 5 phút để thanh toán.</span></div>
          <div>${ic.shield}<span>Giá cuối cùng do hệ thống xác nhận ở bước xem trước đơn.</span></div>
        </div>
      </aside>
    </div>
  </div>

  <section class="panel" style="margin-top:16px" aria-labelledby="sim-title">
    <div class="panel-head"><h2 id="sim-title" class="panel-title">Sản phẩm tương tự</h2><a class="link-more" href="Main.html">Xem thêm${ic.chevron}</a></div>
    <div class="hscroll">${similar.map(productCard).join("")}</div>
  </section>`,
  });
}

// ======================================================================
// Flash sale campaigns list (FlashSales)
// ======================================================================
const campaigns = {
  live: [{ name: "Flash Sale 10.10", start: "20:00, 10/10/2026", end: "22:00, 10/10/2026", win: "20:00 – 22:00", date: "10/10", href: "FlashSaleDetail.html", items: [P.tee, P.phones, P.power, P.fryer], more: 2, sold: "124/275 suất đã bán", cd: ["01", "24", "30"] }],
  upcoming: [
    { name: "Gia dụng cuối tuần", start: "09:00, 11/10/2026", end: "11:00, 11/10/2026", win: "09:00 – 11:00", date: "11/10", href: "FlashSaleDetail-upcoming.html", items: [P.fryer, P.bottle, P.lamp], more: 5, sold: "8 sản phẩm · 160 suất", cd: ["12", "24", "30"] },
    { name: "Làm đẹp giữa tháng", start: "20:00, 15/10/2026", end: "22:00, 15/10/2026", win: "20:00 – 22:00", date: "15/10", href: "FlashSaleDetail-upcoming.html", items: [P.sun, P.clean], more: 0, sold: "2 sản phẩm · 80 suất" },
  ],
  ended: [
    { name: "Giờ vàng buổi trưa", start: "12:00, 10/10/2026", end: "14:00, 10/10/2026", win: "12:00 – 14:00", date: "10/10", href: "FlashSaleDetail-ended.html", items: [P.cable, P.case, P.sandal], more: 1, sold: "Đã kết thúc lúc 14:00, 10/10/2026" },
    { name: "Siêu sale điện tử", start: "20:00, 09/10/2026", end: "22:00, 09/10/2026", win: "20:00 – 22:00", date: "09/10", href: "FlashSaleDetail-stopped.html", items: [P.phones, P.power], more: 0, sold: "Đã dừng nhận đơn lúc 21:10, 09/10/2026", stopped: true },
  ],
};

function campaignRow(c, kind) {
  const badge =
    kind === "live"
      ? `<span class="badge badge-sale">${ic.bolt}Đang diễn ra</span>`
      : kind === "upcoming"
        ? `<span class="badge badge-brand">${ic.clock}Sắp diễn ra</span>`
        : c.stopped
          ? `<span class="badge badge-warning">${icPause}Đã dừng</span>`
          : `<span class="badge badge-neutral">Đã kết thúc</span>`;
  const right =
    kind === "live"
      ? `<div class="cat-camp-timer"><span class="muted">Kết thúc sau</span>${countdown(...c.cd)}</div>`
      : kind === "upcoming" && c.cd
        ? `<div class="cat-camp-timer"><span class="muted">Mở sau</span>${countdown(...c.cd)}</div>`
        : kind === "upcoming"
          ? `<div class="cat-camp-timer"><span class="muted">Mở lúc ${c.start}</span></div>`
          : "";
  const cta = kind === "live" ? `<a class="btn btn-sale" href="${c.href}">Xem chương trình</a>` : `<a class="btn ${kind === "upcoming" ? "btn-outline-brand" : ""}" href="${c.href}">Xem chương trình</a>`;
  return `<article class="panel cat-camp cat-camp-${kind}" aria-labelledby="camp-${c.date.replace("/", "")}-${kind}">
    <div class="cat-camp-when"><b class="num">${c.win}</b><span>${icCal}${c.date}</span></div>
    <div class="cat-camp-main">
      <div class="cat-camp-title"><h3 id="camp-${c.date.replace("/", "")}-${kind}">${c.name}</h3>${badge}</div>
      <div class="cat-camp-meta"><span>Bắt đầu: <b>${c.start}</b></span><span>Kết thúc: <b>${c.end}</b></span></div>
      <div class="cat-camp-items">${c.items.map((p) => `<div class="thumb"><img src="${img(p.img)}" alt=""></div>`).join("")}${c.more ? `<span class="cat-camp-more">+${c.more}</span>` : ""}<span class="cat-camp-sold">${c.sold}</span></div>
    </div>
    <div class="cat-camp-side">${right}${cta}</div>
  </article>`;
}

function flashSalesPage() {
  const section = (kind, label, list) => `<section class="cat-camp-group" aria-labelledby="g-${kind}">
    <h2 class="cat-group-title" id="g-${kind}">${label}<span class="count">${list.length}</span></h2>
    ${list.map((c) => campaignRow(c, kind)).join("\n    ")}
  </section>`;
  return page({
    title: "Chương trình flash sale – SaleScale",
    desc: "Danh sách chương trình flash sale đang diễn ra, sắp diễn ra và đã kết thúc.",
    headerHtml: hdr({ flashActive: true }),
    body: `  <nav class="breadcrumb" aria-label="Đường dẫn"><a href="Main.html">Trang chủ</a>${ic.chevron}<span aria-current="page">Flash sale</span></nav>
  <div class="panel cat-list-head">
    <div class="cat-list-title"><span class="flash-title">${ic.bolt}Flash sale</span><h1>Chương trình</h1></div>
    <nav class="cat-filter" aria-label="Lọc chương trình">
      <a href="FlashSales.html" aria-current="true">Tất cả<span class="count">5</span></a>
      <a href="#g-upcoming">Sắp diễn ra<span class="count">2</span></a>
      <a href="#g-live">Đang diễn ra<span class="count">1</span></a>
      <a href="#g-ended">Đã kết thúc<span class="count">2</span></a>
    </nav>
  </div>
  <div class="notice notice-info cat-gap">${ic.info}<span>Giờ bắt đầu, kết thúc và đồng hồ đếm ngược theo <b>giờ hệ thống</b>, không theo đồng hồ máy bạn. Suất sale chỉ được giữ khi đặt đơn thành công.</span></div>
  <div class="cat-camp-list">
  ${section("live", "Đang diễn ra", campaigns.live)}
  ${section("upcoming", "Sắp diễn ra", campaigns.upcoming)}
  ${section("ended", "Đã kết thúc", campaigns.ended)}
  </div>`,
  });
}

// ======================================================================
// Flash sale campaign detail (FlashSaleDetail*)
// ======================================================================
const skus = [
  { p: P.tee, variant: "Đen / M", sku: "AT-DEN-M", img: "tshirt-black", quota: [32, 50] },
  { p: P.tee, variant: "Đen / L", sku: "AT-DEN-L", img: "tshirt-black", quota: [44, 50] },
  { p: P.tee, variant: "Trắng / M", sku: "AT-TRANG-M", img: "tshirt-white", quota: [18, 50] },
  { p: P.phones, variant: "Đen", sku: "TN-X2-DEN", img: "headphones", quota: [17, 20] },
];

function skuCard({ p, variant, sku, img: im, quota }, state) {
  const live = state === "live";
  const closed = state === "ended" || state === "stopped";
  const priceBlock = closed
    ? `<span class="cat-closed-sale">Giá flash sale <s>${vnd(p.sale)}</s></span><span class="cat-regular">Giá thường hiện tại</span><span class="price price-plain">${vnd(p.base)}</span>`
    : `<div class="price-row"><span class="price">${vnd(p.sale)}</span><span class="discount">-${pct(p.sale, p.base)}%</span></div><span class="price-old">${vnd(p.base)}</span>`;
  const quotaBlock = live ? quotaBar(quota) : state === "upcoming" ? `<span class="cat-quota-total">${quota[1]} suất mở bán</span>` : "";
  const actions = live
    ? `<a class="btn btn-sale btn-block" href="ProductDetail.html">Mua ngay</a>`
    : state === "upcoming"
      ? `<button class="btn btn-block" type="button" disabled>${ic.clock}Mở bán lúc 20:00</button><a class="btn btn-ghost btn-block btn-sm" href="ProductDetail.html">Xem sản phẩm</a>`
      : `<a class="btn btn-outline-brand btn-block" href="ProductDetail.html">Xem sản phẩm</a>`;
  return `<article class="panel cat-sku${closed ? " cat-sku-closed" : ""}">
    <a class="thumb" href="ProductDetail.html" tabindex="-1" aria-hidden="true"><img src="${img(im)}" alt="" loading="lazy">${closed ? "" : `<span class="badge badge-sale">-${pct(p.sale, p.base)}%</span>`}</a>
    <div class="cat-sku-body">
      <h3 class="card-name"><a href="ProductDetail.html">${p.short}</a></h3>
      <span class="cat-sku-meta">${variant} · Mã <span class="num">${sku}</span></span>
      ${priceBlock}
      ${quotaBlock}
      <span class="limit-note">Tối đa ${p.limit} sản phẩm/tài khoản</span>
      <div class="cat-sku-actions">${actions}</div>
    </div>
  </article>`;
}

function flashSaleDetail(state) {
  const meta = {
    live: { badge: `<span class="badge cat-badge-on-sale">${ic.bolt}Đang diễn ra</span>`, timer: `<div class="cat-hero-timer"><span>Kết thúc sau</span>${countdown("01", "24", "30")}</div>`, title: "Flash Sale 10.10 – SaleScale", desc: "Chương trình flash sale đang diễn ra và các SKU tham gia." },
    upcoming: { badge: `<span class="badge cat-badge-on-sale">${ic.clock}Sắp diễn ra</span>`, timer: `<div class="cat-hero-timer"><span>Mở sau</span>${countdown("00", "24", "30")}</div>`, title: "Sắp diễn ra – Flash Sale 10.10 – SaleScale", desc: "Chương trình flash sale chưa mở bán, xem trước sản phẩm tham gia." },
    ended: { badge: `<span class="badge cat-badge-on-sale">Đã kết thúc</span>`, timer: "", title: "Đã kết thúc – Flash Sale 10.10 – SaleScale", desc: "Chương trình flash sale đã kết thúc, sản phẩm mua theo giá thường." },
    stopped: { badge: `<span class="badge cat-badge-on-sale">${icPause}Đã dừng</span>`, timer: "", title: "Đã dừng – Flash Sale 10.10 – SaleScale", desc: "Chương trình flash sale đã dừng nhận đơn, sản phẩm mua theo giá thường." },
  }[state];
  const notice = {
    upcoming: `<div class="notice notice-info" role="status">${ic.clock}<div><b>Chưa mở bán</b><p>Chương trình mở lúc 20:00, 10/10/2026. Bạn có thể xem trước sản phẩm, nút mua sẽ mở khi chương trình bắt đầu.</p></div></div>`,
    ended: `<div class="notice cat-notice-neutral" role="status">${ic.info}<div><b>Chương trình đã kết thúc</b><p>Flash Sale 10.10 đã kết thúc lúc 22:00, 10/10/2026. Sản phẩm vẫn mua được theo giá thường.</p></div></div>`,
    stopped: `<div class="notice notice-warning" role="status">${icPause}<div><b>Chương trình đã dừng nhận đơn</b><p>Flash Sale 10.10 đã dừng lúc 21:10, 10/10/2026. Sản phẩm vẫn mua được theo giá thường.</p></div></div>`,
  }[state] || "";
  const closed = state === "ended" || state === "stopped";
  return page({
    title: meta.title,
    desc: meta.desc,
    headerHtml: hdr({ flashActive: true }),
    body: `  <nav class="breadcrumb" aria-label="Đường dẫn"><a href="FlashSales.html" class="cat-back">${icBack}Tất cả chương trình</a></nav>
  <section class="cat-hero${closed ? " cat-hero-closed" : ""}" aria-labelledby="fs-title">
    <div class="cat-hero-main">
      <div class="cat-hero-top"><span class="flash-title">${ic.bolt}Flash sale</span>${meta.badge}</div>
      <h1 id="fs-title">Flash Sale 10.10</h1>
      <p class="cat-hero-win">${icCal}Khung giờ: <b class="num">20:00 – 22:00, 10/10/2026</b></p>
    </div>
    <div class="cat-hero-side">
      ${meta.timer}
      <span class="cat-hero-note">${ic.info}Thời gian chính thức theo hệ thống</span>
    </div>
    <img class="cat-hero-art" src="${img("flash-hero")}" alt="">
  </section>
  ${notice ? `<div class="cat-gap">${notice}</div>` : ""}
  <div class="cat-detail cat-gap">
    <section aria-labelledby="sku-title">
      <div class="cat-sku-head"><h2 id="sku-title">Sản phẩm tham gia <span class="muted">(4 SKU)</span></h2></div>
      <div class="cat-sku-grid">${skus.map((s) => skuCard(s, state)).join("")}</div>
    </section>
    <aside class="panel cat-rules" aria-labelledby="rules-title">
      <h2 id="rules-title" class="pd-section-title">Thông tin chương trình</h2>
      <table class="spec"><tbody>
        <tr><th scope="row">Khung giờ</th><td class="num">20:00 – 22:00, 10/10/2026</td></tr>
        <tr><th scope="row">Trạng thái</th><td>${{ live: "Đang diễn ra", upcoming: "Sắp diễn ra", ended: "Đã kết thúc", stopped: "Đã dừng lúc 21:10" }[state]}</td></tr>
        <tr><th scope="row">Sản phẩm</th><td>4 SKU</td></tr>
        <tr><th scope="row">Giới hạn</th><td>Theo từng SKU, mỗi tài khoản</td></tr>
      </tbody></table>
      <div class="trust">
        ${closed ? `<div>${ic.info}<span>Giá flash sale không còn áp dụng. Sản phẩm vẫn mua được theo <b>giá thường</b>.</span></div>` : ""}
        <div>${ic.lock}<span>Thêm vào giỏ <b>chưa giữ suất</b>. Suất sale chỉ được giữ khi đặt đơn thành công.</span></div>
        <div>${ic.shield}<span>Giá cuối cùng do hệ thống xác nhận ở bước xem trước đơn.</span></div>
        <div>${ic.clock}<span>Giờ mở, đóng và đồng hồ đếm ngược theo giờ hệ thống.</span></div>
      </div>
    </aside>
  </div>`,
  });
}

// ======================================================================
// CustomerHeader – sheet of header states
// ======================================================================
const esc = (s) => s.replace(/&/g, "&amp;").replace(/"/g, "&quot;");

function customerHeaderSheet() {
  const phoneDoc = (h) => `<!doctype html><html lang="vi" data-theme="blue"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Be+Vietnam+Pro:wght@400;500;600;700;800&display=swap"><link rel="stylesheet" href="../assets/salescale.css"></head><body style="background:#f5f5fa">${h}</body></html>`;
  const uniq = (h, id) => h.replace('for="q"', `for="q-${id}"`).replace('<input id="q"', `<input id="q-${id}"`);
  const variant = (id, label, note, h0) => { const h = uniq(h0, id); return `<section class="cat-sheet-item" aria-labelledby="${id}">
    <div class="cat-sheet-label"><h2 id="${id}">${label}</h2><p>${note}</p></div>
    <div class="cat-sheet-frame">${h}</div>
  </section>`; };
  const phone = (id, label, h) => `<figure class="cat-phone">
      <iframe srcdoc="${esc(phoneDoc(h))}" title="${label}" width="390" height="160"></iframe>
      <figcaption id="${id}">${label}</figcaption>
    </figure>`;
  return `${head("Header trang khách – SaleScale", "blue", "Các trạng thái của header trang khách: đã đăng nhập, chưa đăng nhập, số lượng giỏ và khổ điện thoại.", CSS)}
<body class="cat-sheet">
<main class="container">
  <div class="cat-sheet-intro">
    <h1>Header trang khách</h1>
    <p>Logo về <a href="Main.html">Trang chủ</a> · ô tìm theo tên sản phẩm · <a href="MyOrders.html">Đơn hàng</a> · tài khoản (<a href="Auth.html">Đăng nhập</a>) · <a href="Cart.html">Giỏ hàng</a> có số SKU · dải điều hướng với lối vào <a href="FlashSales.html">Flash sale</a>.</p>
  </div>
  ${variant("h1", "Đã đăng nhập · giỏ có 3 sản phẩm", "Trạng thái mặc định. Tên khách hiện ở mục tài khoản; số trên giỏ là số SKU khác nhau trong giỏ.", hdr())}
  ${variant("h2", "Chưa đăng nhập", "Mục tài khoản thành “Đăng nhập”. Không hiện số trên giỏ; bấm giỏ sẽ yêu cầu đăng nhập.", hdr({ cartCount: null, loggedIn: false }))}
  ${variant("h3", "Giỏ đã đủ 20 sản phẩm", "Giỏ chứa tối đa 20 SKU khác nhau. Khi đủ, nút thêm vào giỏ ở trang sản phẩm báo lỗi.", hdr({ cartCount: 20 }))}
  ${variant("h4", "Đang tìm kiếm · đang ở trang Flash sale", "Ô tìm giữ từ khoá vừa tìm; mục Flash sale được đánh dấu là trang hiện tại.", hdr({ q: "áo khoác", flashActive: true }))}
  <section class="cat-sheet-item" aria-labelledby="h5">
    <div class="cat-sheet-label"><h2 id="h5">Khổ điện thoại 390px</h2><p>Logo và biểu tượng trên một hàng, ô tìm xuống hàng riêng, nhãn chữ ẩn đi; dải danh mục cuộn ngang.</p></div>
    <div class="cat-phones">
      ${phone("h5a", "Đã đăng nhập · 3 sản phẩm", hdr())}
      ${phone("h5b", "Chưa đăng nhập", hdr({ cartCount: null, loggedIn: false }))}
    </div>
  </section>
</main>
</body>
</html>
`;
}

export default [
  { file: "Main-empty.html", title: "Trang chủ · không có kết quả", html: mainState("empty") },
  { file: "Main-error.html", title: "Trang chủ · lỗi tải", html: mainState("error") },
  { file: "Main-loading.html", title: "Trang chủ · đang tải", html: mainState("loading") },
  { file: "Main-mobile.html", title: "Trang chủ · điện thoại", html: mobileFrame("Main.html", "Trang chủ") },
  { file: "ProductDetail-added.html", title: "Chi tiết · đã thêm vào giỏ", html: productDetail("added") },
  { file: "ProductDetail-cart-full.html", title: "Chi tiết · giỏ đủ 20 SKU", html: productDetail("cart-full") },
  { file: "ProductDetail-login.html", title: "Chi tiết · cần đăng nhập", html: productDetail("login") },
  { file: "ProductDetail-mobile.html", title: "Chi tiết · điện thoại", html: mobileFrame("ProductDetail.html", "Chi tiết sản phẩm") },
  { file: "ProductDetail-over-limit.html", title: "Chi tiết · vượt số lượng", html: productDetail("over-limit") },
  { file: "FlashSales.html", title: "Danh sách flash sale", html: flashSalesPage() },
  { file: "FlashSales-mobile.html", title: "Danh sách flash sale · điện thoại", html: mobileFrame("FlashSales.html", "Danh sách flash sale") },
  { file: "FlashSaleDetail.html", title: "Chương trình · đang diễn ra", html: flashSaleDetail("live") },
  { file: "FlashSaleDetail-ended.html", title: "Chương trình · đã kết thúc", html: flashSaleDetail("ended") },
  { file: "FlashSaleDetail-mobile.html", title: "Chương trình · điện thoại", html: mobileFrame("FlashSaleDetail.html", "Chương trình flash sale") },
  { file: "FlashSaleDetail-stopped.html", title: "Chương trình · đã dừng", html: flashSaleDetail("stopped") },
  { file: "FlashSaleDetail-upcoming.html", title: "Chương trình · sắp diễn ra", html: flashSaleDetail("upcoming") },
  { file: "CustomerHeader.html", title: "Header trang khách", html: customerHeaderSheet() },
];
