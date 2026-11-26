// Shared building blocks for the SaleScale v2 static mockups.
// ---------- helpers ----------
export const vnd = (n) => n.toLocaleString("vi-VN").replace(/,/g, ".") + " ₫";
export const pct = (sale, base) => Math.round((1 - sale / base) * 100);
export const compact = (n) => (n >= 1000 ? (n / 1000).toFixed(1).replace(".", ",").replace(",0", "") + "k" : String(n));
export const img = (name) => `../assets/img/${name}.svg`;

export const ic = {
  search: `<svg class="icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" aria-hidden="true"><circle cx="9" cy="9" r="6"/><path d="M13.5 13.5 17 17"/></svg>`,
  cart: `<svg class="icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M2 3h2.2l2 10h9.3l1.8-7H5"/><circle cx="8" cy="16.5" r="1.3"/><circle cx="14.5" cy="16.5" r="1.3"/></svg>`,
  user: `<svg class="icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" aria-hidden="true"><circle cx="10" cy="7" r="3.5"/><path d="M3.5 17.5c0-3.6 2.9-5.6 6.5-5.6s6.5 2 6.5 5.6"/></svg>`,
  orders: `<svg class="icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><rect x="4" y="2.5" width="12" height="15" rx="2"/><path d="M7 7h6M7 10.5h6M7 14h3.5"/></svg>`,
  bolt: `<svg class="icon-sm" viewBox="0 0 16 16" fill="currentColor" aria-hidden="true"><path d="M9.2 1 3 9.2h4.3L6.4 15 13 6.6H8.6z"/></svg>`,
  clock: `<svg class="icon-sm" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" aria-hidden="true"><circle cx="8" cy="8" r="6.2"/><path d="M8 4.6V8l2.3 1.5"/></svg>`,
  info: `<svg class="icon-sm" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" aria-hidden="true"><circle cx="8" cy="8" r="6.2"/><path d="M8 7.3v3.8M8 4.9v.1"/></svg>`,
  chevron: `<svg class="icon-sm" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="m6 3.5 4.5 4.5L6 12.5"/></svg>`,
  pin: `<svg class="icon-sm" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true"><path d="M8 14.5s4.8-4.3 4.8-8.1A4.8 4.8 0 0 0 3.2 6.4c0 3.8 4.8 8.1 4.8 8.1z"/><circle cx="8" cy="6.4" r="1.7"/></svg>`,
  trash: `<svg class="icon-sm" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" aria-hidden="true"><path d="M2.5 4.5h11M6 4.5V3h4v1.5M4 4.5l.7 9h6.6l.7-9"/></svg>`,
  shield: `<svg class="icon-sm" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round" aria-hidden="true"><path d="M8 1.5 2.8 3.5v4c0 3.3 2.2 5.9 5.2 7 3-1.1 5.2-3.7 5.2-7v-4z"/><path d="m5.6 8 1.7 1.7 3.2-3.4"/></svg>`,
  lock: `<svg class="icon-sm" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true"><rect x="3" y="7" width="10" height="7.5" rx="1.5"/><path d="M5.3 7V5a2.7 2.7 0 0 1 5.4 0v2"/></svg>`,
  box: `<svg class="icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linejoin="round" aria-hidden="true"><path d="M10 2 3 5.5v9L10 18l7-3.5v-9z"/><path d="M3 5.5 10 9l7-3.5M10 9v9"/></svg>`,
  stack: `<svg class="icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linejoin="round" aria-hidden="true"><path d="m10 2.5 7.5 4-7.5 4-7.5-4z"/><path d="m2.5 10 7.5 4 7.5-4M2.5 13.5l7.5 4 7.5-4"/></svg>`,
  zap: `<svg class="icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linejoin="round" aria-hidden="true"><path d="M11.5 1.5 4 11h5.5l-1 7.5L16 9h-5.5z"/></svg>`,
  list: `<svg class="icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" aria-hidden="true"><path d="M7 5h10M7 10h10M7 15h10M3 5h.01M3 10h.01M3 15h.01"/></svg>`,
  card: `<svg class="icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.6" aria-hidden="true"><rect x="2" y="4.5" width="16" height="11" rx="2"/><path d="M2 8.5h16"/></svg>`,
  tag: `<svg class="icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linejoin="round" aria-hidden="true"><path d="M2.5 10.3V3h7.3l7.7 7.7-7.3 7.3z"/><circle cx="6.5" cy="7" r="1.2"/></svg>`,
  ext: `<svg class="icon-sm" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" aria-hidden="true"><path d="M9 2.5h4.5V7M13.5 2.5 7 9M11 9.5v3.5H3V5h3.5"/></svg>`,
  plus: `<svg class="icon-sm" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" aria-hidden="true"><path d="M8 3v10M3 8h10"/></svg>`,
};

const logoMark = `<span class="logo-mark" aria-hidden="true"><svg width="20" height="20" viewBox="0 0 20 20" fill="none"><path d="M14.5 5.2c-1-1.2-2.6-1.9-4.4-1.9-2.6 0-4.4 1.4-4.4 3.4 0 4.3 8.8 2.4 8.8 6.6 0 2-1.9 3.4-4.6 3.4-2 0-3.7-.8-4.7-2.1" stroke="#fff" stroke-width="2.3" stroke-linecap="round"/><path d="M12.6 1.5 10.4 6h3.2l-2.2 4.5" stroke="#ffe14d" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"/></svg></span>`;
export const logo = (sub = "Flash sale mỗi ngày") =>
  `<a class="logo" href="Main.html" aria-label="SaleScale – Trang chủ">${logoMark}<span class="logo-text"><b>SaleScale</b><small>${sub}</small></span></a>`;

// ---------- data (placeholder values for preview) ----------
export const P = {
  tee: { name: "Áo thun cotton basic cổ tròn, form regular", short: "Áo thun basic", brand: "Basico", img: "tshirt-black", cat: "Thời trang", base: 199000, sale: 129000, rating: 4.8, reviews: 1284, sold: 5320, quota: [32, 50], limit: 2 },
  phones: { name: "Tai nghe Bluetooth X2 chống ồn, pin 30 giờ", short: "Tai nghe Bluetooth X2", brand: "Sonora", img: "headphones", cat: "Điện tử", base: 890000, sale: 590000, rating: 4.6, reviews: 412, sold: 1870, quota: [18, 20], limit: 1 },
  bottle: { name: "Bình giữ nhiệt inox 304 dung tích 500ml", short: "Bình giữ nhiệt 500ml", brand: "Keepwell", img: "bottle", cat: "Gia dụng", base: 250000, rating: 4.9, reviews: 2210, sold: 9640 },
  power: { name: "Sạc dự phòng 10.000mAh sạc nhanh 20W", short: "Sạc dự phòng 10.000mAh", brand: "Voltix", img: "powerbank", cat: "Điện tử", base: 350000, sale: 299000, rating: 4.7, reviews: 836, sold: 3105, quota: [12, 40], limit: 2 },
  fryer: { name: "Nồi chiên không dầu 4,5L điều khiển cảm ứng", short: "Nồi chiên không dầu 4,5L", brand: "Kitchy", img: "airfryer", cat: "Gia dụng", base: 1290000, sale: 990000, rating: 4.8, reviews: 655, sold: 1402, quota: [7, 15], limit: 1 },
  linen: { name: "Áo sơ mi linen tay ngắn thoáng mát", short: "Áo sơ mi linen tay ngắn", brand: "Linea", img: "linen", cat: "Thời trang", base: 329000, rating: 4.5, reviews: 198, sold: 764 },
  cable: { name: "Cáp sạc USB-C to USB-C bọc dù dài 1m", short: "Cáp sạc USB-C 1m", brand: "Cablo", img: "cable", cat: "Điện tử", base: 79000, sale: 49000, rating: 4.7, reviews: 3012, sold: 12800, quota: [5, 100], limit: 3 },
  sun: { name: "Kem chống nắng SPF50 PA++++ 50ml", short: "Kem chống nắng SPF50 50ml", brand: "Sunveil", img: "sunscreen", cat: "Làm đẹp", base: 185000, sale: 139000, rating: 4.9, reviews: 4120, sold: 15300, quota: [50, 50], limit: 2 },
  lamp: { name: "Đèn bàn LED chống cận 3 chế độ sáng", short: "Đèn bàn LED chống cận", brand: "Lumio", img: "lamp", cat: "Gia dụng", base: 219000, rating: 4.6, reviews: 377, sold: 1120 },
  sandal: { name: "Dép quai ngang đế êm chống trượt", short: "Dép quai ngang đế êm", brand: "Softstep", img: "sandals", cat: "Thời trang", base: 99000, rating: 4.4, reviews: 920, sold: 4210 },
  case: { name: "Ốp lưng silicon chống sốc 4 góc", short: "Ốp lưng silicon chống sốc", brand: "Shieldy", img: "phonecase", cat: "Điện tử", base: 59000, rating: 4.3, reviews: 145, sold: 610 },
  clean: { name: "Sữa rửa mặt dịu nhẹ cho da nhạy cảm 150ml", short: "Sữa rửa mặt dịu nhẹ 150ml", brand: "Purely", img: "cleanser", cat: "Làm đẹp", base: 129000, rating: 4.8, reviews: 1760, sold: 6930 },
};

export const rating = (r, extra = "") =>
  `<span class="rating"><span class="stars" style="--r:${r}" aria-hidden="true"></span><span class="sr-only">${r} trên 5 sao</span>${extra}</span>`;

export const quotaBar = ([sold, total]) => {
  const soldOut = sold >= total;
  const w = Math.max(4, Math.round((sold / total) * 100));
  const label = soldOut ? "Đã hết suất" : sold / total >= 0.85 ? `<b>Sắp hết</b> · còn ${total - sold} suất` : `Đã bán <b>${sold}</b>/${total} suất`;
  return `<div class="quota${soldOut ? " soldout" : ""}"><div class="quota-track" role="img" aria-label="Đã bán ${sold} trên ${total} suất"><i style="width:${soldOut ? 100 : w}%"></i></div><span class="quota-label">${label}</span></div>`;
};

export const flashItem = (p) => {
  const soldOut = p.quota[0] >= p.quota[1];
  return `<a class="flash-item" href="ProductDetail.html">
  <div class="thumb"><img src="${img(p.img)}" alt="" loading="lazy"><span class="badge badge-sale">-${pct(p.sale, p.base)}%</span>${soldOut ? `<div class="soldout-veil"><span>Hết suất sale</span></div>` : ""}</div>
  <div class="price-row"><span class="price">${vnd(p.sale)}</span></div>
  <span class="price-old">${vnd(p.base)}</span>
  ${quotaBar(p.quota)}
  <span class="limit-note">Tối đa ${p.limit} sản phẩm/tài khoản</span>
</a>`;
};

export const productCard = (p) => {
  const onSale = !!p.sale;
  const soldOut = onSale && p.quota[0] >= p.quota[1];
  return `<a class="card" href="ProductDetail.html">
  <div class="thumb"><img src="${img(p.img)}" alt="" loading="lazy"></div>
  <div class="card-body">
    <div class="card-tags">${onSale && !soldOut ? `<span class="badge badge-sale">${ic.bolt}Flash sale</span>` : ""}${onSale && soldOut ? `<span class="badge badge-neutral">Hết suất sale</span>` : ""}</div>
    <span class="card-brand">${p.brand}</span>
    <h3 class="card-name">${p.name}</h3>
    ${rating(p.rating, `<span class="sep"></span><span>Đã bán ${compact(p.sold)}</span>`)}
    ${
      onSale && !soldOut
        ? `<div class="price-row"><span class="price">${vnd(p.sale)}</span><span class="discount">-${pct(p.sale, p.base)}%</span></div><span class="price-old">${vnd(p.base)}</span>`
        : `<div class="price-row"><span class="price price-plain">${vnd(p.base)}</span></div><span class="price-old" style="visibility:hidden">.</span>`
    }
    <div class="card-foot">${onSale && !soldOut ? `${ic.clock}Kết thúc 22:00 hôm nay` : `Giá thường · ${p.cat}`}</div>
  </div>
</a>`;
};

// ---------- shells ----------
export const head = (title, theme, desc, extraCss = []) => `<!doctype html>
<html lang="vi" data-theme="${theme}">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>${title}</title>
<meta name="description" content="${desc}">
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Be+Vietnam+Pro:ital,wght@0,400;0,500;0,600;0,700;0,800;1,800&display=swap">
<link rel="stylesheet" href="../assets/salescale.css">${extraCss.map((c) => `\n<link rel="stylesheet" href="../assets/${c}">`).join("")}
</head>`;

export const header = (cartCount = 3, loggedIn = true) => `<header class="site-header">
  <div class="container header-main">
    ${logo()}
    <form class="search" role="search" action="Main.html">
      <div class="search-box">${ic.search}<label class="sr-only" for="q">Tìm sản phẩm</label><input id="q" name="q" placeholder="Tìm áo thun, tai nghe, nồi chiên…"><button type="submit">Tìm kiếm</button></div>
      <div class="search-hints"><a href="#">áo thun</a><a href="#">tai nghe bluetooth</a><a href="#">nồi chiên không dầu</a><a href="#">kem chống nắng</a><a href="#">sạc dự phòng</a></div>
    </form>
    <nav class="header-actions" aria-label="Tài khoản">
      <a class="header-action" href="MyOrders.html">${ic.orders}<span class="label">Đơn hàng</span></a>
      <a class="header-action" href="Auth.html">${ic.user}<span class="label">${loggedIn ? "Minh Anh" : "Đăng nhập"}</span></a>
      <a class="header-action cart" href="Cart.html" aria-label="Giỏ hàng, ${cartCount} sản phẩm">${ic.cart}<span class="cart-count">${cartCount}</span><span class="label">Giỏ hàng</span></a>
    </nav>
  </div>
  <div class="header-sub"><div class="container">
    <a class="flash-link" href="FlashSales.html">${ic.bolt}Flash sale</a>
    <a href="Main.html?cat=thoi-trang">Thời trang</a><a href="Main.html?cat=dien-tu">Điện tử</a><a href="Main.html?cat=gia-dung">Gia dụng</a><a href="Main.html?cat=lam-dep">Làm đẹp</a>
    <span class="ship-to">${ic.pin}Giao đến: Q. 1, TP. Hồ Chí Minh</span>
  </div></div>
</header>`;

export const footer = `<footer class="site-footer">
  <div class="container">
    <div class="footer-grid">
      <div><h4>SaleScale</h4><ul><li>Flash sale hằng ngày với số suất giới hạn</li><li>Giá và suất do hệ thống xác nhận khi đặt đơn</li></ul></div>
      <div><h4>Mua sắm</h4><ul><li><a href="FlashSales.html">Chương trình flash sale</a></li><li><a href="Main.html">Tất cả sản phẩm</a></li><li><a href="Cart.html">Giỏ hàng</a></li></ul></div>
      <div><h4>Đơn hàng</h4><ul><li><a href="MyOrders.html">Đơn của tôi</a></li><li>Thanh toán trong 5 phút sau khi đặt</li><li>Đơn quá hạn tự hủy và trả suất</li></ul></div>
      <div><h4>Hỗ trợ</h4><ul><li>Hotline 1900 0000 (mô phỏng)</li><li>hotro@salescale.vn</li></ul></div>
    </div>
    <div class="footer-bottom"><span>© 2026 SaleScale · Đồ án mô phỏng hệ thống flash sale</span><span>Thời gian hiển thị theo giờ hệ thống</span></div>
  </div>
</footer>`;

export const countdown = (h, m, s) => `<span class="countdown num" role="timer" aria-label="Còn ${h} giờ ${m} phút ${s} giây"><span>${h}</span>:<span>${m}</span>:<span>${s}</span></span>`;

// ---------- admin shell ----------
const navLink = (icon, label, href, active, count) =>
  `<a class="nav-link${active ? " active" : ""}" href="${href}"${active ? ' aria-current="page"' : ""}>${icon}${label}${count ? `<span class="count">${count}</span>` : ""}</a>`;

/** active: "products" | "inventory" | "flashsales" | "orders" | "simulator" | "categories" */
export const adminShell = ({ title, desc, active, body, extraCss = [], searchPlaceholder = "Tìm sản phẩm, mã SKU…" }) => `${head(title, "blue", desc, extraCss)}
<body>
<div class="admin">
  <nav class="admin-nav" aria-label="Quản trị">
    ${logo("Quản trị")}
    <div class="nav-sec"><h4>Bán hàng</h4>
      ${navLink(ic.box, "Sản phẩm", "AdminProducts.html", active === "products", 48)}
      ${navLink(ic.stack, "Tồn kho", "Inventory.html", active === "inventory")}
      ${navLink(ic.zap, "Flash sale", "AdminFlashSales.html", active === "flashsales", 3)}
    </div>
    <div class="nav-sec"><h4>Vận hành</h4>
      ${navLink(ic.list, "Đơn hàng", "#", active === "orders")}
      ${navLink(ic.card, "Mô phỏng thanh toán", "#", active === "simulator")}
      ${navLink(ic.tag, "Danh mục", "#", active === "categories")}
    </div>
    <div class="nav-foot"><a href="Main.html" style="display:inline-flex;gap:6px;align-items:center;color:var(--brand)">${ic.ext}Xem trang khách hàng</a><span>Giờ hệ thống 20:35:30</span></div>
  </nav>
  <div>
    <header class="admin-top">
      <div class="search-box">${ic.search}<label class="sr-only" for="aq">Tìm trong quản trị</label><input id="aq" placeholder="${searchPlaceholder}"></div>
      <div class="admin-user"><span class="avatar" aria-hidden="true">AD</span><span>admin@salescale.vn</span><a class="btn btn-sm" href="Auth.html">Đăng xuất</a></div>
    </header>
    <main class="admin-body">
${body}
    </main>
  </div>
</div>
</body>
</html>
`;

// ---------- customer shell ----------
/** Full customer page: header, <main class="container page">body</main>, footer, optional mobile bar. */
export const customerPage = ({ title, desc, body, cartCount = 3, mobileBar = "", extraCss = [], mainClass = "container page", loggedIn = true }) => `${head(title, "blue", desc, extraCss)}
<body${mobileBar ? ' class="has-mobile-bar"' : ""}>
${header(cartCount, loggedIn)}
<main class="${mainClass}">
${body}
</main>
${footer}
${mobileBar ? `<div class="mobile-bar">${mobileBar}</div>` : ""}
</body>
</html>
`;

// ---------- phone preview ----------
/** A "-mobile" page: shows another v2 page inside a 390×844 phone frame. */
export const mobileFrame = (target, label) => `${head(`${label} – điện thoại`, "blue", `Xem trước ${label} ở khổ điện thoại 390px.`)}
<body style="background:#e9e9ef">
<main style="min-height:100vh;display:grid;place-items:center;gap:12px;padding:24px 16px">
  <p class="muted" style="margin:0;font-size:13px">Khổ điện thoại 390 × 844 · <a style="color:var(--brand)" href="${target}">Mở bản đầy đủ</a></p>
  <iframe src="${target}" title="${label} trên điện thoại" style="width:390px;max-width:100%;height:844px;border:0;border-radius:28px;background:#fff;box-shadow:0 0 0 10px #1f1f23,0 24px 60px rgba(0,0,0,.25)"></iframe>
</main>
</body>
</html>
`;

/** Common customer flow steps: 0 cart, 1 preview, 2 payment */
export const steps = (now) => `<ol class="steps">${["Giỏ hàng", "Xem trước đơn", "Thanh toán"].map((s, i) => `<li class="${i < now ? "done" : i === now ? "now" : ""}">${s}</li>`).join("")}</ol>`;
