import { vnd, pct, img, ic, logo, P, quotaBar, flashItem, productCard, head, countdown, customerPage, mobileFrame, steps } from "../lib.mjs";

export const group = "Tài khoản & hệ thống";

const CSS = ["auth.css"];

// ---------- local icons (lib.mjs has no eye / alert / check icons) ----------
const icx = {
  eye: `<svg class="icon-sm" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M1.5 8s2.5-4.75 6.5-4.75S14.5 8 14.5 8s-2.5 4.75-6.5 4.75S1.5 8 1.5 8Z"/><circle cx="8" cy="8" r="2"/></svg>`,
  alert: `<svg class="icon-sm" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M8 2.25 14.5 13.5h-13Z M8 6.5v3 M8 11.6v.2"/></svg>`,
  check: `<svg class="icon-sm" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="m3.5 8.5 3 3 6-7"/></svg>`,
  copy: `<svg class="icon-sm" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round" aria-hidden="true" style="width:12px;height:12px"><path d="M5.5 5.5h7.5v7.5H5.5Z M3 10.5V3h7.5"/></svg>`,
  // large state icons (30px via .auth-state-ico svg)
  receipt: `<svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.1" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M3.5 1.75h9v12.5l-2.25-1.5L8 14.25l-2.25-1.5L3.5 14.25Z M6 5.5h4 M6 8.5h4"/></svg>`,
  offline: `<svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" aria-hidden="true"><path d="M2 2l12 12 M8 12.5v.25 M5.75 10.25a3.25 3.25 0 0 1 3.5-.75 M3.25 7.75a7 7 0 0 1 2.75-1.6 M10.25 6.2a7 7 0 0 1 2.5 1.55"/></svg>`,
  lockBig: `<svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M4 7.25h8v6.5H4Z M5.75 7.25V5.5a2.25 2.25 0 0 1 4.5 0v1.75"/></svg>`,
  ban: `<svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" aria-hidden="true"><circle cx="8" cy="8" r="6.25"/><path d="M3.6 3.6l8.8 8.8"/></svg>`,
  notFound: `<svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" aria-hidden="true"><circle cx="7" cy="7" r="4.75"/><path d="M10.5 10.5 14 14 M5.5 5.5l3 3 M8.5 5.5l-3 3"/></svg>`,
  hourglass: `<svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M4.5 1.75h7 M4.5 14.25h7 M5.25 1.75v1.5L8 8l2.75-4.75v-1.5 M5.25 14.25v-1.5L8 8l2.75 4.75v1.5"/></svg>`,
  warnBig: `<svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M8 2.25 14.5 13.5h-13Z M8 6.5v3 M8 11.6v.2"/></svg>`,
};

const notice = (kind, role, title, desc, extra = "") => {
  const cls = { info: "notice-info", warning: "notice-warning", danger: "auth-notice-danger", success: "auth-notice-success", neutral: "auth-notice-neutral" }[kind];
  const icon = kind === "danger" ? icx.alert : kind === "success" ? icx.check : kind === "warning" ? ic.clock : ic.info;
  return `<div class="notice ${cls}" role="${role}">${icon}<div class="auth-notice-text"><b>${title}</b>${desc ? `<span>${desc}</span>` : ""}${extra}</div></div>`;
};

// ---------- Auth pages ----------
/**
 * mode: "login" | "register"
 * alert: html placed above the fields
 * email/pass: prefilled values; errors: {email, pass}; submitting; locked
 */
function authPage({ title, desc, mode = "login", alert = "", email = "", pass = "", errors = {}, submitting = false, locked = false }) {
  const login = mode === "login";
  const dis = submitting ? " disabled" : "";
  const emailErr = errors.email ? `<span class="auth-err" id="auth-email-err">${icx.alert}${errors.email}</span>` : "";
  const passHint = errors.pass
    ? `<span class="auth-err" id="auth-pass-hint">${icx.alert}${errors.pass}</span>`
    : login
      ? ""
      : `<span class="auth-hint" id="auth-pass-hint">${ic.info}Tối thiểu 8 ký tự</span>`;
  const label = login ? "Đăng nhập" : "Tạo tài khoản";
  const submit = locked
    ? `<button class="btn btn-lg btn-block" type="submit" disabled aria-busy="false">${ic.clock}Thử lại sau 04:59</button>`
    : submitting
      ? `<button class="btn btn-primary btn-lg btn-block" type="submit" disabled aria-busy="true"><span class="auth-spinner" aria-hidden="true"></span>Đang đăng nhập…</button>`
      : `<button class="btn btn-primary btn-lg btn-block" type="submit" aria-busy="false">${label}</button>`;

  const body = `<div class="auth-main">
  <section class="panel auth-card" aria-labelledby="auth-title">
    <div class="auth-form-col">
      <nav class="auth-tabs" aria-label="Chọn đăng nhập hoặc đăng ký">
        <a href="Auth.html"${login ? ' aria-current="page"' : ""}>Đăng nhập</a>
        <a href="Auth-register.html"${login ? "" : ' aria-current="page"'}>Đăng ký</a>
      </nav>
      <div class="auth-head">
        <h1 id="auth-title">${label}</h1>
        <p>${login ? "Dùng email và mật khẩu SaleScale của bạn." : "Chỉ cần email và mật khẩu để bắt đầu mua flash sale."}</p>
      </div>
      ${alert}
      <form class="auth-form" action="${login ? "Main.html" : "Main.html"}" novalidate>
        <div class="auth-field">
          <label for="auth-email">Email</label>
          <input class="input" id="auth-email" type="email" autocomplete="email" placeholder="ten@email.com" value="${email}" aria-invalid="${errors.email ? "true" : "false"}"${errors.email ? ' aria-describedby="auth-email-err"' : ""}${dis}>
          ${emailErr}
        </div>
        <div class="auth-field">
          <label for="auth-pass">Mật khẩu</label>
          <div class="auth-pass${errors.pass ? " is-invalid" : ""}${submitting ? " is-disabled" : ""}">
            <input id="auth-pass" type="password" autocomplete="${login ? "current-password" : "new-password"}" value="${pass}" aria-invalid="${errors.pass ? "true" : "false"}"${passHint ? ' aria-describedby="auth-pass-hint"' : ""}${dis}>
            <button type="button" aria-pressed="false" aria-label="Hiện mật khẩu"${dis}>${icx.eye}Hiện</button>
          </div>
          ${passHint}
        </div>
        ${submit}
      </form>
      <p class="auth-switch">${login ? `Chưa có tài khoản? <a href="Auth-register.html">Đăng ký</a>` : `Đã có tài khoản? <a href="Auth.html">Đăng nhập</a>`}</p>
    </div>
    <aside class="auth-aside" aria-label="Vì sao nên có tài khoản">
      <img src="${img("auth-hero")}" alt="">
      <h2>Săn flash sale mỗi ngày cùng SaleScale</h2>
      <ul>
        <li>${ic.bolt}<span>Flash sale hằng ngày với số suất giới hạn</span></li>
        <li>${ic.shield}<span>Giá và suất do hệ thống xác nhận khi đặt đơn</span></li>
        <li>${ic.clock}<span>Thanh toán trong 5 phút sau khi đặt; đơn quá hạn tự hủy và trả suất</span></li>
      </ul>
      <span class="clock-note">${ic.info}Thời gian hiển thị theo giờ hệ thống</span>
    </aside>
  </section>
</div>`;
  return customerPage({ title, desc, body, cartCount: 0, loggedIn: false, extraCss: CSS, mainClass: "page" });
}

const FILLED = { email: "minhanh@example.com", pass: "matkhau2026" };

const authPages = [
  {
    file: "Auth.html",
    title: "Đăng nhập",
    html: authPage({ title: "SaleScale – Đăng nhập", desc: "Đăng nhập SaleScale bằng email và mật khẩu." }),
  },
  {
    file: "Auth-email-used.html",
    title: "Đăng ký – email đã dùng",
    html: authPage({
      title: "Đăng ký – email đã được dùng",
      desc: "Đăng ký thất bại vì email đã có tài khoản.",
      mode: "register",
      ...FILLED,
      alert: notice("danger", "alert", "Email này đã được dùng", "Hãy đăng nhập bằng email này hoặc dùng email khác để đăng ký.", `<a href="Auth.html">Đăng nhập bằng email này</a>`),
    }),
  },
  {
    file: "Auth-locked.html",
    title: "Đăng nhập – bị khóa tạm",
    html: authPage({
      title: "Đăng nhập – thử quá nhiều lần",
      desc: "Đăng nhập bị tạm khóa sau nhiều lần thử sai.",
      ...FILLED,
      locked: true,
      alert: `<div class="notice auth-notice-danger" role="alert">${ic.clock}<div class="auth-notice-text"><b>Bạn đã thử đăng nhập quá nhiều lần</b><span>Vui lòng thử lại sau <span class="num" role="timer">04:59</span>.</span><small>Thời gian chính thức theo hệ thống</small></div></div>`,
    }),
  },
  { file: "Auth-mobile.html", title: "Đăng nhập – điện thoại", html: mobileFrame("Auth.html", "Đăng nhập") },
  {
    file: "Auth-redirect.html",
    title: "Đăng nhập – yêu cầu giữa chừng",
    html: authPage({
      title: "Đăng nhập – yêu cầu giữa chừng",
      desc: "Khách cần đăng nhập để thêm sản phẩm vào giỏ, sau đó quay lại trang đang xem.",
      alert: `<div class="notice notice-warning" role="status">${ic.lock}<div class="auth-notice-text"><b>Đăng nhập để thêm sản phẩm vào giỏ</b><span>Sau khi đăng nhập, bạn sẽ quay lại trang Áo thun basic với lựa chọn đang dở.</span>
        <div class="auth-pending"><span class="thumb"><img src="${img("tshirt-black")}" alt=""></span><span><b>Áo thun basic</b><small>Đen, M · 1 sản phẩm</small></span></div></div></div>`,
    }),
  },
  {
    file: "Auth-register.html",
    title: "Đăng ký",
    html: authPage({ title: "Đăng ký", desc: "Tạo tài khoản SaleScale bằng email và mật khẩu.", mode: "register" }),
  },
  {
    file: "Auth-register-errors.html",
    title: "Đăng ký – lỗi theo trường",
    html: authPage({
      title: "Đăng ký – lỗi theo trường",
      desc: "Đăng ký với lỗi định dạng email và mật khẩu quá ngắn.",
      mode: "register",
      email: "minhanh@",
      pass: "abc12",
      errors: { email: "Email chưa đúng định dạng, ví dụ: ten@email.com", pass: "Mật khẩu cần tối thiểu 8 ký tự (hiện có 5)" },
    }),
  },
  {
    file: "Auth-submitting.html",
    title: "Đăng nhập – đang gửi",
    html: authPage({ title: "Đăng nhập – đang gửi", desc: "Đang gửi yêu cầu đăng nhập, form bị khóa.", ...FILLED, submitting: true }),
  },
  {
    file: "Auth-wrong.html",
    title: "Đăng nhập – sai thông tin",
    html: authPage({
      title: "Đăng nhập – sai thông tin",
      desc: "Đăng nhập thất bại do sai email hoặc mật khẩu.",
      ...FILLED,
      alert: notice("danger", "alert", "Email hoặc mật khẩu không đúng", "Kiểm tra lại thông tin rồi thử lần nữa."),
    }),
  },
];

// ---------- Doc shell (Components / CommonStates) ----------
const docTop = (label) => `<header class="auth-doc-top"><div class="container">${logo("Hệ thống giao diện")}<span class="badge badge-brand">UI v2 · ${label}</span>
  <div class="auth-doc-meta"><a href="Components.html">Thành phần</a><a href="CommonStates.html">Trạng thái chung</a><a href="Main.html">Trang chủ mẫu</a></div></div></header>`;

const cls = (...names) => `<div class="auth-cls">${names.map((n) => `<code${n.startsWith(".auth-") ? ' class="is-local"' : ""}>${n}</code>`).join("")}</div>`;

const sec = (id, title, desc, classes, body) => `<section class="panel auth-sec" id="${id}" aria-labelledby="${id}-t">
  <header><h2 id="${id}-t">${title}</h2>${desc ? `<p>${desc}</p>` : ""}${classes}</header>
  <div class="auth-sec-body">${body}</div>
</section>`;

const sub = (title, body, note = "") => `<div class="auth-sub"><h3>${title}</h3>${body}${note ? `<p>${note}</p>` : ""}</div>`;

// ---------- Components ----------
function componentsPage() {
  const toc = [
    ["c-color", "Màu & token"],
    ["c-btn", "Nút"],
    ["c-input", "Ô nhập & số lượng"],
    ["c-badge", "Badge trạng thái"],
    ["c-price", "Khối giá & nhãn nguồn giá"],
    ["c-time", "Đếm ngược & suất"],
    ["c-nav", "Chip, tab, phân trang"],
    ["c-card", "Thẻ sản phẩm & flash sale"],
    ["c-table", "Bảng quản trị"],
    ["c-load", "Tải thêm, rỗng, skeleton"],
    ["c-line", "Lỗi tại dòng"],
    ["c-notice", "Thông báo & banner lỗi"],
    ["c-overlay", "Toast & hộp xác nhận"],
  ];
  const swatches = [
    ["--brand", "#0b6cf0", "Thương hiệu, nút chính"],
    ["--brand-soft", "#e8f1fe", "Nền chọn, badge brand"],
    ["--price / --sale", "#ff424e", "Chỉ cho giá, flash sale"],
    ["--success", "#00a86b", "Thành công"],
    ["--warning", "#b26b00", "Chờ, cảnh báo"],
    ["--danger", "#d0011b", "Lỗi, thất bại"],
    ["--ink", "#27272a", "Chữ chính"],
    ["--bg", "#f5f5fa", "Nền trang"],
  ];
  const t = P.tee;
  const body = `${head("SaleScale – Thành phần dùng chung", "blue", "Bảng thành phần giao diện v2 dùng chung, kèm tên lớp CSS để chuyển sang React.", CSS)}
<body>
${docTop("Thành phần")}
<div class="container auth-doc">
  <nav class="panel auth-toc" aria-label="Mục lục thành phần"><h2>Mục lục</h2><ol>${toc.map(([id, n]) => `<li><a href="#${id}">${n}</a></li>`).join("")}</ol></nav>
  <main class="auth-doc-main">
    <div class="auth-doc-intro">
      <h1>Thành phần dùng chung</h1>
      <p>Mỗi khối ghi tên lớp CSS thật. Lớp <code class="auth-code">màu xanh</code> nằm trong <b>salescale.css</b>; lớp <code class="auth-code" style="color:#6b4a00;background:var(--warning-soft)">màu vàng</code> (tiền tố <b>.auth-</b>) nằm trong <b>auth.css</b> và nên được chuyển vào bộ dùng chung khi làm React.</p>
    </div>

    ${sec("c-color", "Màu & token", "Đỏ chỉ dùng cho giá, giảm giá và flash sale. Trạng thái dùng nền nhạt.", cls("var(--brand)", "var(--price)", "var(--radius) 8px", "Be Vietnam Pro"),
      `<div class="auth-swatches">${swatches.map(([n, c, u]) => `<div class="auth-swatch"><i style="background:${c}"></i><span><b>${n}</b><small>${c}</small><span class="muted" style="font-size:11px">${u}</span></span></div>`).join("")}</div>`)}

    ${sec("c-btn", "Nút", "Nút chính màu xanh; nút đỏ chỉ cho hành động mua trong flash sale.", cls(".btn", ".btn-primary", ".btn-sale", ".btn-outline-brand", ".btn-ghost", ".btn-sm", ".btn-lg", ".btn-block", ".auth-spinner"),
      `${sub("Biến thể", `<div class="auth-row"><button class="btn btn-primary" type="button">Nút chính</button><button class="btn" type="button">Nút phụ</button><button class="btn btn-outline-brand" type="button">Viền thương hiệu</button><button class="btn btn-ghost" type="button">Nút chữ</button><button class="btn btn-sale" type="button">${ic.bolt}Mua giá sốc</button></div>`)}
      ${sub("Kích cỡ & trạng thái", `<div class="auth-row"><button class="btn btn-primary btn-sm" type="button">Nhỏ 32px</button><button class="btn btn-primary" type="button">Thường 40px</button><button class="btn btn-primary btn-lg" type="button">Lớn 48px</button><button class="btn" type="button" disabled>Vô hiệu</button><button class="btn btn-primary" type="button" disabled aria-busy="true"><span class="auth-spinner" aria-hidden="true"></span>Đang gửi…</button></div>`, "Khi đang gửi: khóa nút, đặt aria-busy=\"true\", giữ nguyên chiều rộng.")}`)}

    ${sec("c-input", "Ô nhập & số lượng", "", cls(".input", ".select", ".check", ".qty", ".auth-field", ".auth-pass", ".auth-hint", ".auth-err", '[aria-invalid="true"]'),
      `<div class="auth-cols">
        <div class="auth-field"><label for="cmp-in1">Thường</label><input class="input" id="cmp-in1" value="Nguyễn Minh Anh"></div>
        <div class="auth-field"><label for="cmp-in2">Đang focus</label><input class="input" id="cmp-in2" value="0901 234 567" style="border-color:var(--brand);box-shadow:0 0 0 3px var(--brand-soft)"></div>
        <div class="auth-field"><label for="cmp-in3">Lỗi</label><input class="input" id="cmp-in3" value="minhanh@" aria-invalid="true" aria-describedby="cmp-in3-e"><span class="auth-err" id="cmp-in3-e">${icx.alert}Email chưa đúng định dạng</span></div>
        <div class="auth-field"><label for="cmp-in4">Mật khẩu</label><div class="auth-pass"><input id="cmp-in4" type="password" value="matkhau2026"><button type="button" aria-pressed="false" aria-label="Hiện mật khẩu">${icx.eye}Hiện</button></div><span class="auth-hint">${ic.info}Tối thiểu 8 ký tự</span></div>
      </div>
      <div class="auth-cols">
        <div class="auth-field"><label for="cmp-sel">Danh mục</label><select class="select" id="cmp-sel"><option>Tất cả danh mục</option><option>Thời trang</option></select></div>
        <div class="auth-field"><span style="font-weight:500">Số lượng</span><div class="auth-row"><div class="qty"><button type="button" aria-label="Giảm" disabled>−</button><input aria-label="Số lượng" value="1" inputmode="numeric"><button type="button" aria-label="Tăng">+</button></div><span class="limit-note">Tối đa 2/tài khoản</span></div></div>
        <div class="auth-field"><span style="font-weight:500">Hộp chọn</span><label class="auth-row" style="gap:8px"><input class="check" type="checkbox" checked>Chọn tất cả (3 sản phẩm)</label></div>
      </div>`)}

    ${sec("c-badge", "Badge trạng thái", "Badge nền nhạt, chữ đậm vừa. Một trạng thái luôn dùng một màu ở mọi màn hình.", cls(".badge", ".badge-success", ".badge-warning", ".badge-danger", ".badge-neutral", ".badge-brand", ".badge-sale", ".badge-sale-soft", ".status.on", ".status.off"),
      `${sub("Trạng thái đơn", `<div class="auth-row"><span class="badge badge-warning">Chờ thanh toán</span><span class="badge badge-success">Đã xác nhận</span><span class="badge badge-danger">Thanh toán thất bại</span><span class="badge badge-neutral">Đã hủy</span><span class="badge badge-neutral">Hết hạn</span></div>`, "“Đã xác nhận” nghĩa là đơn đã được chốt, không phải đã giao hàng.")}
      ${sub("Phiên thanh toán", `<div class="auth-row"><span class="badge badge-warning">Đang chờ</span><span class="badge badge-neutral">Chưa xác định</span><span class="badge badge-success">Thành công</span><span class="badge badge-danger">Thất bại</span></div>`, "“Chưa xác định” không bao giờ dùng màu xanh lá: kết quả chưa rõ không phải thành công.")}
      ${sub("Trạng thái flash sale", `<div class="auth-row"><span class="badge badge-neutral">Nháp</span><span class="badge badge-brand">Đã lên lịch</span><span class="badge badge-brand">Sắp diễn ra</span><span class="badge badge-sale">${ic.bolt}Đang diễn ra</span><span class="badge badge-warning">Đã dừng</span><span class="badge badge-neutral">Đã kết thúc</span></div>`, "Trang quản trị hiện “Đã lên lịch”, trang khách hiện “Sắp diễn ra” cho cùng một trạng thái.")}
      ${sub("Trạng thái bán (quản trị)", `<div class="auth-row" style="gap:16px"><span class="status on">Đang bán</span><span class="status off">Ngừng bán</span></div>`)}`)}

    ${sec("c-price", "Khối giá & nhãn nguồn giá", "", cls(".price", ".price-plain", ".price-old", ".discount", ".price-row", ".badge-sale-soft", ".badge-neutral"),
      `${sub("Nhãn nguồn giá", `<div class="auth-row"><span class="badge badge-sale-soft">${ic.bolt}Flash sale</span><span class="badge badge-neutral">Giá thường</span></div>`)}
      <div class="auth-cols">
        <div class="auth-pricebox"><span>Giá thường</span><div class="price-row"><span class="price price-plain">${vnd(250000)}</span></div></div>
        <div class="auth-pricebox"><span>Có ưu đãi đang chạy</span><div class="price-row"><span class="price">${vnd(t.sale)}</span><span class="discount">-${pct(t.sale, t.base)}%</span></div><span class="price-old">${vnd(t.base)}</span><div class="auth-row" style="gap:6px"><span class="badge badge-sale-soft">${ic.bolt}Flash sale</span><span class="clock-note">Kết thúc 22:00</span></div></div>
      </div>`)}

    ${sec("c-time", "Đếm ngược & thanh suất", "Đếm ngược chỉ để hiển thị, không đọc từng giây cho trình đọc màn hình.", cls(".countdown", ".clock-note", ".quota", ".quota.soldout", ".limit-note", ".steps"),
      `<div class="auth-cols">
        ${sub("Đồng hồ flash sale", `<div class="auth-row"><span class="muted">Kết thúc sau</span>${countdown("01", "24", "30")}</div><span class="clock-note">${ic.info}Thời gian chính thức theo hệ thống</span>`)}
        ${sub("Hạn thanh toán", `<span class="auth-paytimer">Còn <b role="timer">04:59</b> để thanh toán</span><span class="clock-note">${ic.info}Thời gian chính thức theo hệ thống</span>`)}
      </div>
      ${sub("Thanh suất (điểm nhận diện)", `<div class="auth-cols">${quotaBar([32, 50])}${quotaBar([18, 20])}${quotaBar([50, 50])}</div>`, "Ba trạng thái: đã bán x/y, sắp hết (≥ 85%), đã hết suất.")}
      ${sub("Bước đặt hàng", `<div style="max-width:420px;border:1px solid var(--line);border-radius:var(--radius)">${steps(1)}</div>`)}`)}

    ${sec("c-nav", "Chip, tab, phân trang", "", cls(".chip", ".chip.on", ".tabs", ".tab.active", ".pager", ".breadcrumb"),
      `${sub("Chip lọc", `<div class="auth-row"><a class="chip on" href="#" aria-pressed="true">Tất cả</a><a class="chip" href="#">${ic.bolt}Đang flash sale</a><a class="chip" href="#">Thời trang</a><a class="chip" href="#">Điện tử</a></div>`)}
      ${sub("Tab", `<div style="border:1px solid var(--line);border-radius:var(--radius);overflow-x:auto"><div class="tabs" role="tablist" style="padding:0 8px"><a class="tab active" role="tab" aria-selected="true" href="#">Tất cả<span class="count">48</span></a><a class="tab" role="tab" aria-selected="false" href="#">Đang bán<span class="count">46</span></a><a class="tab" role="tab" aria-selected="false" href="#">Ngừng bán<span class="count">2</span></a></div></div>`)}
      ${sub("Phân trang (quản trị)", `<div class="pager" style="border:1px solid var(--line);border-radius:var(--radius)"><span>Hiển thị 1–8 / 48</span><div class="btns"><button class="btn btn-sm" type="button" disabled aria-label="Trang trước">‹</button><button class="btn btn-sm on" type="button" aria-current="page">1</button><button class="btn btn-sm" type="button">2</button><button class="btn btn-sm" type="button">3</button><button class="btn btn-sm" type="button" aria-label="Trang sau">›</button></div></div>`)}
      ${sub("Đường dẫn", `<nav class="breadcrumb" aria-label="Đường dẫn mẫu" style="padding:0"><a href="Main.html">Trang chủ</a>${ic.chevron}<a href="#">Thời trang</a>${ic.chevron}<span aria-current="page">Áo thun basic</span></nav>`)}`)}

    ${sec("c-card", "Thẻ sản phẩm & ô flash sale", "Dùng helper productCard() và flashItem() trong lib.mjs.", cls(".card", ".thumb", ".card-body", ".card-brand", ".card-name", ".rating", ".card-foot", ".flash-item", ".soldout-veil"),
      `${sub("Thẻ sản phẩm: flash sale, hết suất, giá thường", `<div class="auth-demo-cards">${[P.tee, P.sun, P.bottle, P.lamp, P.cable].map(productCard).join("")}</div>`)}
      ${sub("Ô flash sale", `<div class="auth-demo-flash">${[P.phones, P.power, P.sun].map(flashItem).join("")}</div>`)}`)}

    ${sec("c-table", "Bảng quản trị", "Bảng chỉ dùng đường kẻ ngang; số căn phải.", cls(".table-wrap", ".table", ".prod-cell", ".stock", ".stock-bar", ".status", ".row-actions"),
      `<div class="table-wrap" style="border:1px solid var(--line);border-radius:var(--radius)"><table class="table">
        <thead><tr><th scope="col">Sản phẩm</th><th scope="col" class="r">Giá</th><th scope="col">Tồn kho</th><th scope="col">Trạng thái</th><th scope="col"><span class="sr-only">Thao tác</span></th></tr></thead>
        <tbody>
          <tr><td><div class="prod-cell"><div class="thumb"><img src="${img(t.img)}" alt=""></div><div><b>${t.short}</b><small>${t.brand} · 12 SKU</small></div></div></td><td class="r">129.000 – 199.000 ₫</td><td><div class="stock"><span class="num"><b>420</b> khả dụng · 86 đang giữ</span><div class="stock-bar"><i class="avail" style="width:83%"></i><i class="held" style="width:17%"></i></div></div></td><td><span class="status on">Đang bán</span></td><td><div class="row-actions"><a href="ProductForm.html">Sửa</a><button type="button" class="danger">Ngừng bán</button></div></td></tr>
          <tr><td><div class="prod-cell"><div class="thumb"><img src="${img(P.case.img)}" alt=""></div><div><b>${P.case.short}</b><small>${P.case.brand} · 3 SKU</small></div></div></td><td class="r">59.000 ₫</td><td><div class="stock"><span class="num"><b>0</b> khả dụng</span><div class="stock-bar"></div><small>Hết hàng</small></div></td><td><span class="status off">Ngừng bán</span></td><td><div class="row-actions"><a href="ProductForm.html">Sửa</a><button type="button">Mở bán lại</button></div></td></tr>
        </tbody></table></div>`)}

    ${sec("c-load", "Tải thêm, trạng thái rỗng, skeleton", "", cls(".load-more", ".auth-state", ".auth-state-ico", ".auth-skel", ".auth-skel-line", ".auth-skel-thumb"),
      `${sub("Nút “Tải thêm”", `<div class="auth-row"><button class="btn btn-outline-brand" type="button">Tải thêm</button><button class="btn btn-outline-brand" type="button" disabled aria-busy="true"><span class="auth-spinner" aria-hidden="true"></span>Đang tải…</button></div>`, "Danh sách dài tải theo con trỏ (cursor), không có số trang.")}
      <div class="auth-cols">
        ${sub("Trạng thái rỗng", `<div class="auth-state" style="border:1px dashed var(--line-strong);border-radius:var(--radius)"><span class="auth-state-ico">${icx.receipt}</span><h3>Chưa có dữ liệu</h3><p>Mô tả ngắn điều gì sẽ hiện ở đây.</p><button class="btn btn-outline-brand" type="button">Hành động gợi ý</button></div>`)}
        ${sub("Skeleton", `<div role="status" aria-label="Đang tải" class="auth-sk-rows" style="border:1px solid var(--line);border-radius:var(--radius);padding:16px">
          ${[1, 2, 3].map(() => `<div class="auth-sk-row" aria-hidden="true"><span class="auth-skel auth-skel-thumb"></span><span style="display:grid;gap:8px"><span class="auth-skel auth-skel-line" style="width:80%"></span><span class="auth-skel auth-skel-line" style="width:45%"></span></span></div>`).join("")}
          <span class="auth-skel auth-skel-line" aria-hidden="true" style="height:36px;width:50%;justify-self:center"></span>
        </div>`, "Hiệu ứng lấp lánh tắt khi người dùng chọn giảm chuyển động.")}
      </div>`)}

    ${sec("c-line", "Lỗi tại dòng", "Hiện ngay dưới dòng sản phẩm bị ảnh hưởng ở bước xem trước đơn, kèm một hành động sửa.", cls(".auth-lineerr", ".badge"),
      `<div class="auth-lineerr"><ul>${[
        ["Chỉ còn 3 sản phẩm", "Không đủ hàng"],
        ["Đã hết suất giá flash sale", "Hết suất ưu đãi"],
        ["Vượt giới hạn 2 sản phẩm/tài khoản", "Vượt giới hạn mua"],
        ["Chương trình chưa mở hoặc đã kết thúc", "Chương trình không hoạt động"],
        ["Sản phẩm đã ngừng bán", "Ngừng bán"],
        [`Giá đã đổi thành ${vnd(275000)}`, "Đổi giá"],
      ].map(([m, c]) => `<li role="alert"><span class="badge">${c}</span><b>${m}</b><a class="btn btn-sm" href="Cart.html">Sửa trong giỏ</a></li>`).join("")}</ul></div>`)}

    ${sec("c-notice", "Thông báo & banner lỗi", "", cls(".notice", ".notice-info", ".notice-warning", ".auth-notice-danger", ".auth-notice-success", ".auth-notice-neutral", ".auth-reqid", ".auth-copy", ".auth-timer"),
      `${sub("Thông báo nhắc nghiệp vụ", `<div class="notice notice-info">${ic.lock}<span>Sản phẩm trong giỏ <b>chưa được giữ</b>. Hàng và suất flash sale chỉ được giữ khi bạn đặt đơn thành công.</span></div>`)}
      ${sub("Banner lỗi kèm mã yêu cầu", `<div class="notice auth-notice-danger" role="alert">${icx.alert}<div class="auth-notice-text"><b>Hệ thống tạm thời không xử lý được</b><span>Vui lòng thử lại sau ít phút.</span><span class="auth-reqid" style="justify-content:flex-start">Mã yêu cầu: <code>req_8f2c1a7b9d</code><button class="auth-copy" type="button" aria-label="Sao chép mã yêu cầu">${icx.copy}Sao chép</button></span></div></div>`)}
      ${sub("Khối thông báo trạng thái đơn", `<div style="display:grid;gap:8px">
        ${notice("info", "status", "Chờ thanh toán", "Hoàn tất thanh toán trước 20:46.")}
        ${notice("success", "status", "Thanh toán thành công", "Đơn đã chuyển sang Đã xác nhận.")}
        ${notice("danger", "status", "Thanh toán thất bại", "Đơn đã đóng. Bấm Đặt lại để tạo đơn mới.")}
        ${notice("warning", "status", "Đang kiểm tra kết quả", "Chưa thể kết luận, không cần thao tác lại.")}
        ${notice("neutral", "status", "Đơn đã hết hạn", "Không thể thanh toán đơn này nữa.")}
      </div>`)}
      ${sub("Chờ thử lại khi quá tải", `<div class="notice notice-warning" role="status" style="align-items:center;flex-wrap:wrap">${ic.clock}<div class="auth-notice-text" style="flex:1"><b>Hệ thống đang quá tải</b><span>Thử lại sau <span class="auth-timer" role="timer">00:15</span></span></div><button class="btn btn-sm" type="button" disabled>Thử lại sau 00:15</button></div>`)}`)}

    ${sec("c-overlay", "Toast & hộp xác nhận", "", cls(".auth-toast", ".auth-scrim", ".auth-dialog", ".btn-danger"),
      `<div class="auth-cols">
        ${sub("Toast", `<div class="auth-toast" role="status"><span class="thumb"><img src="${img("tshirt-black")}" alt=""></span><div><b>${icx.check}Đã thêm vào giỏ</b><small>Áo thun basic · Đen / M · 1 sản phẩm · Giá flash sale</small></div><a href="Cart.html">Xem giỏ</a></div>`)}
        ${sub("Hộp xác nhận", `<div class="auth-scrim"><div class="auth-dialog" role="dialog" aria-modal="false" aria-labelledby="cmp-dlg" aria-describedby="cmp-dlg-d"><h3 id="cmp-dlg">Hủy đơn SS-1010-0042?</h3><p id="cmp-dlg-d">Hàng đang giữ cho đơn này sẽ được trả lại ngay. Bạn không thể hoàn tác thao tác này.</p><div class="auth-dialog-actions"><button class="btn" type="button">Giữ đơn</button><button class="btn btn-danger" type="button">Xác nhận hủy đơn</button></div></div></div>`)}
      </div>`)}
  </main>
</div>
</body>
</html>
`;
  return body;
}

// ---------- CommonStates ----------
function commonStatesPage() {
  const state = (id, label, role, inner) => `<section class="panel" aria-labelledby="${id}"><header><h2 id="${id}">${label}</h2></header>${inner.replace("ROLE", role)}</section>`;
  const loading = `<div class="auth-loading" role="status" aria-label="Đang tải nội dung">
    <span class="auth-skel auth-skel-title" aria-hidden="true"></span>
    <div class="auth-sk-grid" aria-hidden="true">${[1, 2, 3, 4].map(() => `<div class="auth-sk-card"><span class="auth-skel auth-skel-thumb" style="border-radius:0"></span><span class="auth-skel auth-skel-line"></span><span class="auth-skel auth-skel-line" style="width:55%"></span></div>`).join("")}</div>
    <span class="auth-skel auth-skel-line" aria-hidden="true" style="width:70%"></span>
    <span class="sr-only">Đang tải nội dung…</span>
  </div>`;
  const s = (tone, icon, title, desc, action, extra = "") =>
    `<div class="auth-state" role="ROLE"><span class="auth-state-ico${tone ? ` is-${tone}` : ""}">${icon}</span><h3>${title}</h3>${desc ? `<p>${desc}</p>` : ""}${extra}${action}</div>`;
  const body = `<div class="auth-page-head"><div><h1>Trạng thái chung cho mọi màn hình</h1><p>Mỗi khối thay cho phần nội dung chính của trang; header và footer giữ nguyên.</p></div><a class="link-more" href="Components.html">Xem bảng thành phần${ic.chevron}</a></div>
<div class="auth-states">
  ${state("cs-loading", "Đang tải", "status", loading)}
  ${state("cs-empty", "Rỗng", "status", s("", icx.receipt, "Chưa có dữ liệu", "Ví dụ: “Bạn chưa có đơn hàng nào”, “Chưa có chương trình sắp diễn ra”.", `<a class="btn btn-primary" href="Main.html">Xem sản phẩm</a>`))}
  ${state("cs-network", "Lỗi mạng", "alert", s("danger", icx.offline, "Mất kết nối mạng", "Kiểm tra kết nối Internet rồi thử lại.", `<button class="btn btn-primary" type="button">Thử lại</button>`))}
  ${state("cs-login", "Chưa đăng nhập", "status", s("warning", icx.lockBig, "Bạn cần đăng nhập để tiếp tục", "Đang chuyển tới trang đăng nhập… Sau khi đăng nhập, bạn sẽ quay lại đúng trang này.", `<a class="btn btn-primary" href="Auth-redirect.html">Đăng nhập ngay</a>`, `<span class="auth-spinner" aria-hidden="true" style="color:var(--brand)"></span>`))}
  ${state("cs-forbidden", "Không có quyền", "alert", s("", icx.ban, "Bạn không có quyền truy cập trang này", "Tài khoản hiện tại không được phép xem nội dung này.", `<a class="btn btn-primary" href="Main.html">Về trang chủ</a>`))}
  ${state("cs-notfound", "Không tìm thấy", "alert", s("brand", icx.notFound, "Không tìm thấy nội dung", "Sản phẩm, đơn hàng hoặc trang này không còn tồn tại.", `<a class="btn btn-primary" href="Main.html">Về trang chủ</a>`))}
  ${state("cs-overload", "Quá tải", "status", s("warning", icx.hourglass, "Hệ thống đang quá tải", `Thử lại sau <span class="auth-timer" role="timer">00:15</span>`, `<button class="btn" type="button" disabled>Thử lại sau 00:15</button>`, `<small>Thời gian chờ do hệ thống trả về</small>`))}
  ${state("cs-system", "Lỗi hệ thống", "alert", s("danger", icx.warnBig, "Hệ thống tạm thời không xử lý được", "Vui lòng thử lại sau ít phút.", `<button class="btn btn-primary" type="button">Thử lại</button>`, `<span class="auth-reqid">Mã yêu cầu: <code>req_3c71e0b52a</code><button class="auth-copy" type="button" aria-label="Sao chép mã yêu cầu">${icx.copy}Sao chép</button></span>`))}
</div>`;
  return customerPage({ title: "SaleScale – Trạng thái chung", desc: "Các trạng thái dùng chung: đang tải, rỗng, lỗi mạng, chưa đăng nhập, không có quyền, không tìm thấy, quá tải, lỗi hệ thống.", body, extraCss: CSS });
}

export default [
  ...authPages,
  { file: "Components.html", title: "Thành phần dùng chung", html: componentsPage() },
  { file: "CommonStates.html", title: "Trạng thái chung", html: commonStatesPage() },
];
