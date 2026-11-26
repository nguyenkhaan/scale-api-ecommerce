import { vnd, pct, img, ic, adminShell, quotaBar, countdown } from "../lib.mjs";

export const group = "Quản trị";

const CSS = ["admin.css"];
const nf = (n) => n.toLocaleString("vi-VN").replace(/,/g, ".");
const signed = (n) => (n > 0 ? `+${nf(n)}` : n < 0 ? `−${nf(-n)}` : "0");

// Local icons that lib.mjs does not provide.
const icAlert = `<svg class="icon-sm" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M8 2.25 14.5 13.5h-13Z M8 6.5v3 M8 11.6v.2"/></svg>`;
const icReload = `<svg class="icon-sm" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M13.5 8a5.5 5.5 0 1 1-1.6-3.9 M13.5 2.5v3.5H10"/></svg>`;
const icPause = `<svg class="icon-sm" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" aria-hidden="true"><path d="M6 4v8 M10 4v8"/></svg>`;
const icClose = `<svg class="icon-sm" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" aria-hidden="true"><path d="m4 4 8 8M12 4l-8 8"/></svg>`;
const icImage = `<svg class="icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round" aria-hidden="true"><rect x="2.5" y="3.5" width="15" height="13" rx="2"/><circle cx="7" cy="8" r="1.5"/><path d="m3 15 4.5-4.5 3 3 2.5-2.5 4 4"/></svg>`;

/** Quota bar for a closed program (stopped/ended): same look, no "Sắp hết" urgency. */
const closedBar = ([sold, total]) =>
  `<div class="quota${sold >= total ? " soldout" : ""}"><div class="quota-track" role="img" aria-label="Đã bán ${sold} trên ${total} suất"><i style="width:${Math.max(4, Math.round((sold / total) * 100))}%"></i></div><span class="quota-label">${sold >= total ? "Đã hết suất" : `Đã bán <b>${sold}</b>/${total} suất · đã đóng`}</span></div>`;

const crumbs = (items) =>
  `<nav class="breadcrumb adm-crumbs" aria-label="Đường dẫn">${items
    .map(([label, href], i) => (i === items.length - 1 ? `<span aria-current="page">${label}</span>` : `<a href="${href}">${label}</a>${ic.chevron}`))
    .join("")}</nav>`;

const errText = (id, text) => `<p class="adm-err" id="${id}">${icAlert}<span>${text}</span></p>`;

/** Labelled form field with optional help text and error. */
const field = ({ id, label, req, value = "", disabled, error, help, type = "text", placeholder = "", extra = "", cls = "" }) => {
  const desc = [help ? `${id}-help` : "", error ? `${id}-err` : ""].filter(Boolean).join(" ");
  return `<div class="adm-field ${cls}">
            <label for="${id}">${label}${req ? ` <span class="adm-req" aria-hidden="true">*</span>` : ""}${extra}</label>
            <input id="${id}" class="input${error ? " adm-invalid" : ""}" type="${type}" value="${value}"${placeholder ? ` placeholder="${placeholder}"` : ""}${req ? " required" : ""}${disabled ? " disabled" : ""}${error ? ' aria-invalid="true"' : ""}${desc ? ` aria-describedby="${desc}"` : ""}>
            ${help ? `<p class="adm-help" id="${id}-help">${help}</p>` : ""}${error ? errText(`${id}-err`, error) : ""}
          </div>`;
};

const errorSummary = (title, text, links) => `<div class="adm-notice adm-notice-danger" role="alert" tabindex="-1">
        ${icAlert}<div><b>${title}</b><p>${text}</p><ul class="adm-errlist">${links.map(([href, t]) => `<li><a href="${href}">${t}</a></li>`).join("")}</ul></div>
      </div>`;

// =====================================================================
// Flash sale list
// =====================================================================
const FS_STATE = {
  draft: ["Nháp", "adm-st-draft", "badge-neutral"],
  scheduled: ["Đã lên lịch", "adm-st-sched", "badge-brand"],
  live: ["Đang diễn ra", "adm-st-live", "badge-sale-soft"],
  stopped: ["Đã dừng", "adm-st-stop", "badge-warning"],
  ended: ["Đã kết thúc", "adm-st-ended", "badge-neutral"],
};
const fsDot = (s) => `<span class="status ${FS_STATE[s][1]}">${FS_STATE[s][0]}</span>`;
const fsBadge = (s) => `<span class="badge ${FS_STATE[s][2]}">${FS_STATE[s][0]}</span>`;

function adminFlashSalesPage() {
  const rows = [
    ["Flash Sale 10.10", "live", "20:00, 10/10/2026", "22:00, 10/10/2026", 4, [78, 130], "FlashSaleForm-live.html", "Xem"],
    ["Flash Sale cuối tuần", "scheduled", "20:00, 11/10/2026", "22:00, 11/10/2026", 3, null, "FlashSaleForm-scheduled.html", "Xem"],
    ["Flash Sale 11.11", "draft", "20:00, 11/11/2026", "22:00, 11/11/2026", 2, null, "FlashSaleForm.html", "Sửa"],
    ["Flash Sale 9.10", "stopped", "20:00, 09/10/2026", "22:00, 09/10/2026", 3, [61, 90], "FlashSaleForm-ended.html", "Xem"],
    ["Flash Sale 9.9", "ended", "20:00, 09/09/2026", "22:00, 09/09/2026", 4, [116, 130], "FlashSaleForm-ended.html", "Xem"],
  ];
  const tabs = [["Tất cả", 5, true], ["Nháp", 1], ["Đã lên lịch", 1], ["Đang diễn ra", 1], ["Đã dừng", 1], ["Đã kết thúc", 1]];
  return adminShell({
    title: "Flash sale – SaleScale Quản trị",
    desc: "Danh sách chương trình flash sale theo trạng thái.",
    active: "flashsales",
    extraCss: CSS,
    searchPlaceholder: "Tìm chương trình flash sale…",
    body: `      <div class="admin-head">
        <div><h1>Flash sale</h1><p>Giá, quota và giới hạn mua bị khoá khi chương trình đã lên lịch. Thời gian theo giờ hệ thống.</p></div>
        <div class="actions"><a class="btn btn-primary" href="FlashSaleForm.html">${ic.plus}Tạo chương trình</a></div>
      </div>

      <section class="kpis" aria-label="Tổng quan">
        <div class="kpi"><span>Đang diễn ra</span><b>1</b><small>Flash Sale 10.10 · kết thúc 22:00</small></div>
        <div class="kpi"><span>Suất đã bán (đang diễn ra)</span><b>78<em class="adm-of">/130</em></b><small>11 suất đang giữ cho đơn chờ thanh toán</small></div>
        <div class="kpi"><span>Đã lên lịch</span><b>1</b><small>Bắt đầu 20:00, 11/10/2026</small></div>
        <div class="kpi"><span>Nháp</span><b>1</b><small class="warn">Chưa lên lịch · chưa hiển thị cho khách</small></div>
      </section>

      <section class="panel" aria-label="Danh sách chương trình">
        <div class="tabs adm-tabs" role="tablist">
          ${tabs.map(([t, n, on]) => `<a class="tab${on ? " active" : ""}" role="tab" aria-selected="${on ? "true" : "false"}" href="#">${t}<span class="count">${n}</span></a>`).join("\n          ")}
        </div>
        <div class="filters">
          <label class="sr-only" for="fs-q">Tìm chương trình</label><input id="fs-q" class="input" placeholder="Tên chương trình">
          <label class="sr-only" for="fs-m">Tháng</label><select id="fs-m" class="select"><option>Tất cả thời gian</option><option>Tháng 11/2026</option><option>Tháng 10/2026</option><option>Tháng 9/2026</option></select>
          <span class="result">Đang hiển thị 5 chương trình</span>
        </div>
        <div class="table-wrap">
          <table class="table adm-cards">
            <thead><tr><th scope="col">Tên chương trình</th><th scope="col">Trạng thái</th><th scope="col">Giờ bắt đầu</th><th scope="col">Giờ kết thúc</th><th scope="col" class="r">Số SKU</th><th scope="col">Suất đã bán</th><th scope="col"><span class="sr-only">Thao tác</span></th></tr></thead>
            <tbody>
              ${rows
                .map(
                  ([name, st, start, end, skus, q, href, act]) => `<tr>
                <td data-label="Chương trình"><a class="adm-strong-link" href="${href}">${name}</a></td>
                <td data-label="Trạng thái">${fsDot(st)}</td>
                <td data-label="Bắt đầu" class="num">${start}</td>
                <td data-label="Kết thúc" class="num">${end}</td>
                <td data-label="Số SKU" class="r">${skus}</td>
                <td data-label="Suất đã bán"><div class="adm-quota-cell">${q ? (st === "live" ? quotaBar(q) : closedBar(q)) : `<span class="muted">Chưa bắt đầu</span>`}</div></td>
                <td class="adm-cell-act adm-cell-abs"><div class="row-actions"><a href="${href}">${act}</a></div></td>
              </tr>`
                )
                .join("\n              ")}
            </tbody>
          </table>
        </div>
        <div class="pager"><span>Đang hiển thị 5 chương trình</span><button class="btn btn-sm" type="button">Tải thêm</button></div>
      </section>`,
  });
}

// =====================================================================
// Flash sale form (draft / scheduled / live / ended / errors)
// =====================================================================
const SKU = {
  "AT-DEN-M": { label: "Áo thun basic · Đen / M", base: 199000, img: "tshirt-black" },
  "AT-DEN-L": { label: "Áo thun basic · Đen / L", base: 199000, img: "tshirt-black" },
  "AT-TRANG-M": { label: "Áo thun basic · Trắng / M", base: 199000, img: "tshirt-white" },
  "AT-TRANG-L": { label: "Áo thun basic · Trắng / L", base: 199000, img: "tshirt-white" },
  "TN-X2-DEN": { label: "Tai nghe Bluetooth X2 · Đen", base: 890000, img: "headphones" },
  "BGN-500-BAC": { label: "Bình giữ nhiệt 500ml · Bạc", base: 250000, img: "bottle" },
  "SDP-10K-DEN": { label: "Sạc dự phòng 10.000mAh · Đen", base: 350000, img: "powerbank" },
};

const skuCell = (code, err) => {
  const s = SKU[code];
  return `<div class="prod-cell adm-skucell"><div class="thumb"><img src="${img(s.img)}" alt=""></div><div><b><code class="adm-sku">${code}</code> · ${s.label}</b><small>Giá thường ${vnd(s.base)}</small>${err || ""}</div></div>`;
};

const FS = {
  draft: {
    name: "Flash Sale 11.11", start: "20:00, 11/11/2026", end: "22:00, 11/11/2026",
    skus: [{ code: "AT-DEN-M", price: 129000, quota: 50, limit: 2 }, { code: "TN-X2-DEN", price: 590000, quota: 20, limit: 1 }],
  },
  errors: {
    name: "Flash Sale 11.11", start: "20:00, 11/11/2026", end: "19:00, 11/11/2026",
    skus: [
      { code: "AT-DEN-M", price: 199000, quota: 50, limit: 2, priceErr: "Giá flash sale phải nhỏ hơn giá thường (199.000 ₫)" },
      { code: "AT-DEN-M", price: 129000, quota: 10, limit: 2, rowErr: "SKU AT-DEN-M bị trùng trong chương trình này" },
      { code: "TN-X2-DEN", price: 590000, quota: 20, limit: 1, rowErr: "SKU đã thuộc “Flash Sale ngày độc thân” (19:00 – 21:00, 11/11/2026), thời gian bị chồng lấn" },
    ],
  },
  scheduled: {
    name: "Flash Sale cuối tuần", start: "20:00, 11/10/2026", end: "22:00, 11/10/2026",
    skus: [{ code: "AT-DEN-M", price: 129000, quota: 40, limit: 2 }, { code: "AT-TRANG-M", price: 129000, quota: 30, limit: 2 }, { code: "TN-X2-DEN", price: 590000, quota: 15, limit: 1 }],
  },
  live: {
    name: "Flash Sale 10.10", start: "20:00, 10/10/2026", end: "22:00, 10/10/2026",
    skus: [
      { code: "AT-DEN-M", price: 129000, quota: 50, limit: 2, held: 6, sold: 31 },
      { code: "AT-DEN-L", price: 129000, quota: 30, limit: 2, held: 2, sold: 18 },
      { code: "AT-TRANG-M", price: 129000, quota: 30, limit: 2, held: 0, sold: 12 },
      { code: "TN-X2-DEN", price: 590000, quota: 20, limit: 1, held: 3, sold: 17 },
    ],
  },
  ended: {
    name: "Flash Sale 9.9", start: "20:00, 09/09/2026", end: "22:00, 09/09/2026",
    skus: [
      { code: "AT-DEN-M", price: 129000, quota: 50, limit: 2, held: 0, sold: 50 },
      { code: "AT-DEN-L", price: 129000, quota: 30, limit: 2, held: 0, sold: 27 },
      { code: "AT-TRANG-M", price: 129000, quota: 30, limit: 2, held: 0, sold: 19 },
      { code: "TN-X2-DEN", price: 590000, quota: 20, limit: 1, held: 0, sold: 20 },
    ],
  },
};

function flashSaleFormPage(mode) {
  const d = FS[mode];
  const status = mode === "errors" ? "draft" : mode;
  const editable = status === "draft"; // SKU price / quota / limit
  const infoLocked = status === "live" || status === "ended"; // name / start / end
  const hasStats = status === "live" || status === "ended";
  const sum = (k) => d.skus.reduce((a, s) => a + (s[k] || 0), 0);

  const actions = {
    draft: `<button class="btn" type="button">Lưu nháp</button><button class="btn btn-primary" type="button">Lên lịch</button>`,
    scheduled: `<button class="btn" type="button">Hủy lịch</button><button class="btn btn-primary" type="button">Lưu thay đổi</button>`,
    live: `<button class="btn adm-btn-danger" type="button" aria-haspopup="dialog">${icPause}Dừng nhận đơn</button>`,
    ended: `<a class="btn" href="AdminFlashSales.html">Quay lại danh sách</a>`,
  }[status];

  const notice = {
    errors: errorSummary("Chưa lên lịch được chương trình", "Có 4 lỗi cần sửa: giờ kết thúc, giá flash sale, SKU bị trùng và SKU thuộc chương trình khác có thời gian chồng lấn.", [
      ["#fsf-end", "Giờ kết thúc phải sau giờ bắt đầu"],
      ["#fsf-r1-price", "AT-DEN-M: giá flash sale phải nhỏ hơn giá thường"],
      ["#fsf-r2", "AT-DEN-M bị trùng trong chương trình"],
      ["#fsf-r3", "TN-X2-DEN thuộc chương trình khác có thời gian chồng lấn"],
    ]),
    scheduled: `<div class="notice notice-warning adm-notice" role="status">${ic.lock}<div><b>Giá, quota và giới hạn mua đã bị khoá</b><p>Chương trình đã lên lịch nên các giá trị này không đổi được, để khách luôn thấy đúng giá đã công bố. Bấm <b>Hủy lịch</b> để đưa về nháp và chỉnh sửa.</p></div></div>`,
    ended: `<div class="notice notice-info adm-notice" role="status">${ic.info}<div><b>Chương trình đã kết thúc, chỉ xem</b><p>${d.name} kết thúc lúc 22:00, 09/09/2026. Không thể chỉnh sửa hay khởi động lại chương trình này.</p></div></div>`,
  }[mode] || "";

  const kpis = hasStats
    ? `<section class="kpis" aria-label="Tổng quan chương trình">
        <div class="kpi"><span>Đã bán</span><b>${sum("sold")}<em class="adm-of">/${sum("quota")} suất</em></b><small>${Math.round((sum("sold") / sum("quota")) * 100)}% tổng quota</small></div>
        <div class="kpi"><span>Đang giữ cho đơn chờ thanh toán</span><b>${sum("held")}</b><small>${status === "live" ? "Đơn đã giữ hàng vẫn được thanh toán tới hạn" : "Không còn suất đang giữ"}</small></div>
        <div class="kpi"><span>${status === "live" ? "Kết thúc sau" : "Đã kết thúc lúc"}</span>${status === "live" ? `<div class="adm-kpi-timer">${countdown("01", "24", "30")}</div><small>22:00, 10/10/2026 · theo giờ hệ thống</small>` : `<b>22:00</b><small>09/09/2026 · theo giờ hệ thống</small>`}</div>
        <div class="kpi"><span>SKU tham gia</span><b>${d.skus.length}</b><small>${d.skus.filter((s) => s.sold >= s.quota).length} SKU đã hết suất</small></div>
      </section>`
    : "";

  const lockNote = (text) => `<span class="adm-locknote">${ic.lock}${text}</span>`;

  const skuRows = d.skus
    .map((s, i) => {
      const n = i + 1;
      const s0 = SKU[s.code];
      const rowErr = s.rowErr ? errText(`fsf-r${n}-err`, s.rowErr) : "";
      const disc = s.price < s0.base ? `<small class="adm-disc">Giảm ${pct(s.price, s0.base)}%</small>` : "";
      if (!editable) {
        return `<tr>
                <td data-label="SKU">${skuCell(s.code)}</td>
                <td data-label="Giá flash sale"><span class="adm-fsprice">${vnd(s.price)}</span>${disc}</td>
                <td data-label="Quota" class="r">${s.quota}</td>
                <td data-label="Giới hạn mỗi tài khoản">${s.limit} sản phẩm</td>
                <td class="adm-cell-act adm-cell-abs"><span class="adm-lockcell" title="Đã khoá">${ic.lock}<span class="sr-only">Đã khoá</span></span></td>
              </tr>`;
      }
      const priceDesc = s.priceErr ? ` aria-describedby="fsf-r${n}-price-err"` : "";
      return `<tr id="fsf-r${n}"${s.rowErr || s.priceErr ? ' class="adm-row-err"' : ""}>
                <td data-label="SKU">${skuCell(s.code, rowErr)}</td>
                <td data-label="Giá flash sale (₫)"><div class="adm-money"><input id="fsf-r${n}-price" class="input num${s.priceErr ? " adm-invalid" : ""}" type="number" value="${s.price}" aria-label="Giá flash sale, ${s.code}"${s.priceErr ? ' aria-invalid="true"' : ""}${priceDesc}${s.rowErr ? ` aria-describedby="fsf-r${n}-err"` : ""}><span aria-hidden="true">₫</span></div>${s.priceErr ? errText(`fsf-r${n}-price-err`, s.priceErr) : disc}</td>
                <td data-label="Quota"><input class="input num adm-in-sm" type="number" value="${s.quota}" aria-label="Quota, ${s.code}"></td>
                <td data-label="Giới hạn mỗi tài khoản"><div class="adm-money"><input class="input num adm-in-sm" type="number" value="${s.limit}" aria-label="Giới hạn mỗi tài khoản, ${s.code}"><span aria-hidden="true">sp</span></div></td>
                <td class="adm-cell-act adm-cell-abs"><button class="adm-iconbtn" type="button" aria-label="Bỏ SKU ${s.code} khỏi chương trình">${ic.trash}</button></td>
              </tr>`;
    })
    .join("\n              ");

  const stats = hasStats
    ? `<div class="table-wrap"><table class="table adm-cards">
            <thead><tr><th scope="col">SKU</th><th scope="col" class="r">Quota</th><th scope="col" class="r">Đang giữ</th><th scope="col" class="r">Đã bán</th><th scope="col">Tiến độ</th></tr></thead>
            <tbody>
              ${d.skus
                .map(
                  (s) => `<tr><td data-label="SKU"><code class="adm-sku">${s.code}</code> · ${SKU[s.code].label}</td><td data-label="Quota" class="r">${s.quota}</td><td data-label="Đang giữ" class="r">${s.held}</td><td data-label="Đã bán" class="r"><b>${s.sold}</b></td><td data-label="Tiến độ"><div class="adm-quota-cell">${status === "live" ? quotaBar([s.sold, s.quota]) : closedBar([s.sold, s.quota])}</div></td></tr>`
                )
                .join("\n              ")}
            </tbody>
          </table></div>`
    : `<div class="adm-empty">${ic.list}<p><b>Chưa có dữ liệu.</b> Thống kê xuất hiện khi chương trình bắt đầu nhận đơn.</p></div>`;

  const dialog =
    status === "live"
      ? `
<div class="adm-scrim">
  <div class="adm-modal" role="dialog" aria-modal="true" aria-labelledby="stop-title" aria-describedby="stop-desc">
    <span class="adm-modal-icon" aria-hidden="true">${icPause}</span>
    <h2 id="stop-title">Dừng nhận đơn cho ${d.name}?</h2>
    <div id="stop-desc">
      <p>Khách sẽ không đặt được đơn mới với giá flash sale nữa. <b>Thao tác này không hoàn tác được.</b></p>
      <p class="adm-modal-note">${ic.info}Các đơn đã giữ hàng vẫn được thanh toán tới hạn.</p>
    </div>
    <div class="adm-modal-foot"><button class="btn" type="button" autofocus>Tiếp tục chạy</button><button class="btn adm-btn-danger-solid" type="button">Dừng nhận đơn</button></div>
  </div>
</div>`
      : "";

  const titles = { draft: "Chương trình flash sale", errors: "Chương trình – lỗi", scheduled: "Chương trình – đã lên lịch", live: "Chương trình – đang diễn ra", ended: "Chương trình – đã kết thúc" };
  const html = adminShell({
    title: `${titles[mode]} – SaleScale Quản trị`,
    desc: `Chương trình ${d.name}: thông tin, SKU tham gia và thống kê.`,
    active: "flashsales",
    extraCss: CSS,
    searchPlaceholder: "Tìm chương trình flash sale…",
    body: `      ${crumbs([["Flash sale", "AdminFlashSales.html"], [d.name]])}
      <div class="admin-head adm-head">
        <div><h1>${d.name} ${fsBadge(status)}</h1><p>${d.start} – ${d.end} · ${d.skus.length} SKU</p></div>
        <div class="actions">${actions}</div>
      </div>
      ${notice}
      ${kpis}
      <form class="adm-form" action="#" novalidate>
      <section class="panel" aria-labelledby="fsf-info">
        <div class="panel-head adm-ph"><h2 class="panel-title" id="fsf-info">Thông tin chương trình</h2>${infoLocked ? lockNote(status === "live" ? "Đang diễn ra · không sửa được" : "Đã kết thúc · chỉ xem") : ""}</div>
        <div class="adm-grid adm-grid-fs">
          ${field({ id: "fsf-name", label: "Tên chương trình", req: true, value: d.name, disabled: infoLocked })}
          ${field({ id: "fsf-start", label: "Giờ bắt đầu", req: true, value: d.start, disabled: infoLocked, help: infoLocked ? "" : "Giờ hệ thống, định dạng HH:mm, dd/MM/yyyy" })}
          ${field({ id: "fsf-end", label: "Giờ kết thúc", req: true, value: d.end, disabled: infoLocked, error: mode === "errors" ? "Giờ kết thúc phải sau giờ bắt đầu" : "" })}
        </div>
      </section>

      <section class="panel" aria-labelledby="fsf-skus">
        <div class="panel-head adm-ph"><h2 class="panel-title" id="fsf-skus">SKU tham gia <span class="adm-count">${d.skus.length}</span></h2>${!editable ? lockNote("Giá, quota và giới hạn mua đã khoá") : ""}</div>
        <ul class="adm-rules"><li>Giá flash sale phải nhỏ hơn giá thường.</li><li>Mỗi SKU chỉ xuất hiện một lần trong chương trình.</li></ul>
        <div class="table-wrap">
          <table class="table adm-cards adm-skutable">
            <thead><tr><th scope="col">SKU</th><th scope="col">Giá flash sale (₫)</th><th scope="col"${editable ? "" : ' class="r"'}>Quota</th><th scope="col">Giới hạn mỗi tài khoản</th><th scope="col"><span class="sr-only">Thao tác</span></th></tr></thead>
            <tbody>
              ${skuRows}
            </tbody>
          </table>
        </div>
        ${editable ? `<div class="adm-addrow"><button class="btn btn-sm btn-outline-brand" type="button">${ic.plus}Thêm SKU</button></div>` : ""}
      </section>
      </form>

      <section class="panel" aria-labelledby="fsf-stats">
        <div class="panel-head adm-ph"><h2 class="panel-title" id="fsf-stats">Thống kê theo SKU</h2>${hasStats ? `<span class="muted adm-small">Cập nhật theo giờ hệ thống 20:35:30</span>` : ""}</div>
        ${stats}
      </section>`,
  });
  return dialog ? html.replace("</body>", `${dialog}\n</body>`) : html;
}

// =====================================================================
// Inventory
// =====================================================================
const INV = [
  ["AT-DEN-M", 120, 6, true],
  ["AT-DEN-L", 80, 2, true],
  ["AT-TRANG-M", 95, 0, true],
  ["AT-TRANG-L", 0, 0, false],
  ["TN-X2-DEN", 40, 3, true],
  ["BGN-500-BAC", 60, 0, true],
  ["SDP-10K-DEN", 35, 1, true],
];

function inventoryPage(mode = "list") {
  const total = INV.reduce((a, r) => a + r[1], 0);
  const held = INV.reduce((a, r) => a + r[2], 0);
  const focus = mode !== "list";
  const rows = INV.map(([code, onHand, h, on]) => {
    const avail = onHand - h;
    const s = SKU[code];
    const sel = focus && code === "AT-DEN-M";
    return `<tr${sel ? ' class="adm-row-sel"' : ""}>
                <td data-label="Mã SKU"><code class="adm-sku">${code}</code></td>
                <td data-label="Sản phẩm"><div class="prod-cell adm-invprod"><div class="thumb"><img src="${img(s.img)}" alt=""></div><div><b>${s.label}</b>${on ? "" : `<span class="status off">Ngừng bán</span>`}</div></div></td>
                <td data-label="Hiện có" class="r">${onHand}</td>
                <td data-label="Đang giữ" class="r">${h ? `<span class="adm-held">${h}</span>` : "0"}</td>
                <td data-label="Khả dụng" class="r"><div class="stock adm-avail"><b class="num">${avail}</b><div class="stock-bar" aria-hidden="true"><i class="avail" style="width:${onHand ? (avail / onHand) * 100 : 0}%"></i><i class="held" style="width:${onHand ? (h / onHand) * 100 : 0}%"></i></div></div></td>
                <td class="adm-cell-act"><div class="row-actions"><a href="Inventory-adjust.html"${sel && mode !== "history" ? ' aria-current="true"' : ""}>Điều chỉnh</a><a href="Inventory-history.html"${sel && mode === "history" ? ' aria-current="true"' : ""}>Lịch sử</a></div></td>
              </tr>`;
  }).join("\n              ");

  const titles = { list: "Tồn kho", adjust: "Điều chỉnh tồn", "below-held": "Điều chỉnh tồn – dưới lượng đang giữ", conflict: "Điều chỉnh tồn – xung đột phiên bản", history: "Lịch sử điều chỉnh" };
  const html = adminShell({
    title: `${titles[mode]} – SaleScale Quản trị`,
    desc: "Tồn kho theo SKU: hiện có, đang giữ và khả dụng.",
    active: "inventory",
    extraCss: CSS,
    body: `      <div class="admin-head">
        <div><h1>Tồn kho</h1><p>Mọi điều chỉnh đều cần lý do và được ghi vào lịch sử của SKU.</p></div>
        <div class="actions"><a class="btn" href="AdminProducts.html">Về danh sách sản phẩm</a></div>
      </div>

      <div class="adm-formula" role="note">${ic.info}<span><b>Khả dụng</b> = <b>Hiện có</b> − <b>Đang giữ</b> <span class="muted">(hàng đã được giữ cho đơn chờ thanh toán)</span></span></div>

      <section class="kpis" aria-label="Tổng quan tồn kho">
        <div class="kpi"><span>Hiện có</span><b>${total}</b><small>trên ${INV.length} SKU</small></div>
        <div class="kpi"><span>Đang giữ</span><b>${held}</b><small>cho đơn chờ thanh toán</small></div>
        <div class="kpi"><span>Khả dụng</span><b>${total - held}</b><small>có thể bán ngay</small></div>
        <div class="kpi"><span>Hết hàng</span><b>1</b><small class="warn">AT-TRANG-L · đang ngừng bán</small></div>
      </section>

      <section class="panel" aria-label="Tồn kho theo SKU">
        <div class="tabs" role="tablist">
          <a class="tab active" role="tab" aria-selected="true" href="Inventory.html">Tất cả<span class="count">7</span></a>
          <a class="tab" role="tab" aria-selected="false" href="#">Đang giữ hàng<span class="count">4</span></a>
          <a class="tab" role="tab" aria-selected="false" href="#">Hết hàng<span class="count">1</span></a>
        </div>
        <div class="filters">
          <label class="sr-only" for="inv-q">Tìm SKU</label><input id="inv-q" class="input" placeholder="Mã SKU hoặc tên sản phẩm">
          <label class="sr-only" for="inv-c">Danh mục</label><select id="inv-c" class="select"><option>Tất cả danh mục</option><option>Thời trang</option><option>Điện tử</option><option>Gia dụng</option></select>
          <span class="result">Đang hiển thị 7 SKU</span>
        </div>
        <div class="table-wrap">
          <table class="table adm-cards">
            <thead><tr><th scope="col">Mã SKU</th><th scope="col">Sản phẩm</th><th scope="col" class="r">Hiện có</th><th scope="col" class="r">Đang giữ</th><th scope="col" class="r">Khả dụng</th><th scope="col"><span class="sr-only">Thao tác</span></th></tr></thead>
            <tbody>
              ${rows}
            </tbody>
          </table>
        </div>
        <div class="pager"><span>Đang hiển thị 7 SKU</span><button class="btn btn-sm" type="button">Tải thêm</button></div>
      </section>`,
  });
  const drawer = mode === "list" ? "" : mode === "history" ? historyDrawer() : adjustDrawer(mode);
  return drawer ? html.replace("</body>", `${drawer}\n</body>`) : html;
}

function adjustDrawer(mode) {
  const below = mode === "below-held";
  const conflict = mode === "conflict";
  const delta = below ? -118 : -20;
  const reason = below ? "Kiểm kê cuối ngày" : "Kiểm kê cuối ngày: thiếu 20 sản phẩm";
  const after = 120 + delta;
  const availAfter = after - 6;
  const deltaDesc = below ? "adj-delta-err" : "adj-delta-help";
  return `
<div class="adm-scrim adm-scrim-side">
  <aside class="adm-drawer" role="dialog" aria-modal="true" aria-labelledby="adj-title">
    <div class="adm-drawer-head">
      <div><h2 id="adj-title">Điều chỉnh tồn · <code class="adm-sku">AT-DEN-M</code></h2><p>Áo thun basic · Đen / M</p></div>
      <a class="adm-iconbtn" href="Inventory.html" aria-label="Đóng">${icClose}</a>
    </div>
    <div class="adm-drawer-body">
      ${
        conflict
          ? `<div class="adm-notice adm-notice-danger" role="alert">${icAlert}<div><b>Dữ liệu vừa bị thay đổi bởi người khác</b><p>Số tồn của SKU này đã đổi sau khi bạn mở hộp thoại. Tải lại để xem số mới nhất rồi điều chỉnh lại.</p><a class="btn btn-sm adm-btn-reload" href="Inventory-adjust.html">${icReload}Tải lại</a></div></div>`
          : ""
      }
      <dl class="adm-trio${conflict ? " adm-stale" : ""}">
        <div><dt>Hiện có</dt><dd>120</dd></div>
        <div><dt>Đang giữ</dt><dd class="adm-held">6</dd></div>
        <div><dt>Khả dụng</dt><dd>114</dd></div>
      </dl>
      <div class="adm-field">
        <label for="adj-delta">Số lượng thay đổi <span class="adm-req" aria-hidden="true">*</span></label>
        <input id="adj-delta" class="input num${below ? " adm-invalid" : ""}" type="number" value="${delta}" required${below ? ' aria-invalid="true"' : ""} aria-describedby="${deltaDesc}">
        ${below ? errText("adj-delta-err", "Không thể giảm xuống dưới lượng đang giữ (6). Bạn chỉ có thể trừ tối đa 114.") : `<p class="adm-help" id="adj-delta-help">Nhập số dương để cộng, số âm để trừ. Phải khác 0.</p>`}
      </div>
      <div class="adm-field">
        <label for="adj-reason">Lý do <span class="adm-req" aria-hidden="true">*</span></label>
        <textarea id="adj-reason" class="input adm-textarea" rows="2" required aria-describedby="adj-reason-help">${reason}</textarea>
        <p class="adm-help" id="adj-reason-help">Bắt buộc. Lý do được ghi vào lịch sử điều chỉnh của SKU.</p>
      </div>
      <div class="adm-after${below ? " bad" : ""}" role="status">
        <span class="adm-after-title">Sau điều chỉnh</span>
        <div><span>Hiện có</span><b class="num">120 → ${after}</b></div>
        <div><span>Khả dụng</span><b class="num">114 → ${availAfter < 0 ? `−${-availAfter}` : availAfter}</b></div>
      </div>
    </div>
    <div class="adm-drawer-foot"><a class="btn" href="Inventory.html">Hủy</a><button class="btn btn-primary" type="button"${below || conflict ? " disabled" : ""}>Xác nhận điều chỉnh</button></div>
  </aside>
</div>`;
}

function historyDrawer() {
  const rows = [
    ["18:02, 10/10/2026", "admin@salescale.vn", 100, 120, "Nhập thêm hàng trước Flash Sale 10.10"],
    ["09:15, 09/10/2026", "kho@salescale.vn", 105, 100, "Hàng lỗi đường may"],
    ["08:30, 01/10/2026", "admin@salescale.vn", 0, 105, "Tồn ban đầu khi tạo SKU"],
  ];
  return `
<div class="adm-scrim adm-scrim-side">
  <aside class="adm-drawer adm-drawer-wide" role="dialog" aria-modal="true" aria-labelledby="his-title">
    <div class="adm-drawer-head">
      <div><h2 id="his-title">Lịch sử điều chỉnh · <code class="adm-sku">AT-DEN-M</code></h2><p>Áo thun basic · Đen / M · hiện có 120</p></div>
      <a class="btn btn-sm" href="Inventory.html">Đóng</a>
    </div>
    <div class="adm-drawer-body adm-flush">
      <table class="table adm-cards adm-ledger">
        <thead><tr><th scope="col">Thời điểm</th><th scope="col">Người thực hiện</th><th scope="col" class="r">Trước</th><th scope="col" class="r">Sau</th><th scope="col" class="r">Thay đổi</th><th scope="col">Lý do</th></tr></thead>
        <tbody>
          ${rows
            .map(([t, who, b, a, why]) => {
              const dlt = a - b;
              return `<tr><td data-label="Thời điểm" class="num">${t}</td><td data-label="Người thực hiện">${who}</td><td data-label="Trước" class="r">${b}</td><td data-label="Sau" class="r">${a}</td><td data-label="Thay đổi" class="r"><b class="${dlt > 0 ? "adm-plus" : "adm-minus"}">${signed(dlt)}</b></td><td data-label="Lý do">${why}</td></tr>`;
            })
            .join("\n          ")}
        </tbody>
      </table>
    </div>
  </aside>
</div>`;
}

// =====================================================================
// Product form (create / edit / errors)
// =====================================================================
function productFormPage(mode) {
  const edit = mode === "edit";
  const errs = mode === "errors";
  const name = errs ? "" : "Áo thun basic";
  const skus = edit
    ? [["AT-DEN-M", "Đen", "M", 199000, 5, true], ["AT-DEN-L", "Đen", "L", 199000, 5, true], ["AT-TRANG-M", "Trắng", "M", 199000, 5, true], ["AT-TRANG-L", "Trắng", "L", 199000, 5, false]]
    : [["AT-DEN-M", "Đen", "M", 199000, 5, 120], ["AT-DEN-L", "Đen", "L", errs ? 0 : 199000, 5, 80], [errs ? "AT-DEN-M" : "AT-TRANG-M", "Trắng", "M", 199000, 5, 95]];
  const images = edit ? ["tshirt-black", "tshirt-white", "tshirt-fold", "tshirt-detail"] : errs ? [] : ["tshirt-black", "tshirt-fold"];

  const attr = (n, k, v) => `<div class="adm-attr"><input class="input" type="text" value="${k}" aria-label="Tên thuộc tính, dòng ${n}"><input class="input" type="text" value="${v}" aria-label="Giá trị thuộc tính, dòng ${n}"></div>`;

  const rows = skus
    .map(([code, color, size, price, max, last], i) => {
      const n = i + 1;
      const codeErr = errs && n === 3;
      const priceErr = errs && n === 2;
      const codeCell = edit
        ? `<code class="adm-sku">${code}</code><small class="adm-locknote">${ic.lock}Không đổi</small>`
        : `<input id="pf-r${n}-code" class="input adm-in-code${codeErr ? " adm-invalid" : ""}" type="text" value="${code}" aria-label="Mã SKU, dòng ${n}" required${codeErr ? ` aria-invalid="true" aria-describedby="pf-r${n}-code-err"` : ""}>${codeErr ? errText(`pf-r${n}-code-err`, "Mã SKU đã tồn tại") : ""}`;
      const lastCell = edit
        ? `<td data-label="Trạng thái"><div class="adm-skustatus"><span class="status ${last ? "on" : "off"}">${last ? "Đang bán" : "Ngừng bán"}</span><button class="adm-linkbtn${last ? " danger" : ""}" type="button">${last ? "Ngừng bán" : "Mở bán lại"}</button></div></td>`
        : `<td data-label="Tồn ban đầu"><input class="input num adm-in-sm" type="number" value="${last}" aria-label="Tồn ban đầu, dòng ${n}"></td>`;
      return `<tr id="pf-r${n}"${codeErr || priceErr ? ' class="adm-row-err"' : ""}>
                <td data-label="Mã SKU">${codeCell}</td>
                <td data-label="Thuộc tính"><div class="adm-attrs">${attr(n, "Màu", color)}${attr(n, "Kích cỡ", size)}<button class="adm-linkbtn" type="button">${ic.plus}Thêm thuộc tính</button></div></td>
                <td data-label="Giá thường (₫)"><div class="adm-money"><input id="pf-r${n}-price" class="input num${priceErr ? " adm-invalid" : ""}" type="number" value="${price}" aria-label="Giá thường, dòng ${n}" required${priceErr ? ` aria-invalid="true" aria-describedby="pf-r${n}-price-err"` : ""}><span aria-hidden="true">₫</span></div>${priceErr ? errText(`pf-r${n}-price-err`, "Giá phải lớn hơn 0") : ""}</td>
                <td data-label="Tối đa mỗi đơn"><input class="input num adm-in-sm" type="number" value="${max}" aria-label="Tối đa mỗi đơn, dòng ${n}"></td>
                ${lastCell}
                ${edit ? "" : `<td class="adm-cell-act adm-cell-abs"><button class="adm-iconbtn" type="button" aria-label="Bỏ dòng SKU ${n}">${ic.trash}</button></td>`}
              </tr>`;
    })
    .join("\n              ");

  const gallery = `<section class="panel adm-gallery-panel" aria-labelledby="pf-img-title">
          <div class="panel-head adm-ph"><h2 class="panel-title" id="pf-img-title">Hình ảnh <span class="adm-count">${images.length}</span></h2><span class="badge badge-brand">Mới ở v2</span></div>
          <div class="adm-pad">
            ${
              images.length
                ? `<ul class="adm-gallery" aria-label="Ảnh sản phẩm">${images
                    .map(
                      (im, i) => `<li class="adm-gimg${i === 0 ? " main" : ""}"><img src="${img(im)}" alt="Ảnh ${i + 1}">${i === 0 ? `<span class="adm-gtag">Ảnh chính</span>` : `<button class="adm-gbtn" type="button">Đặt làm ảnh chính</button>`}<button class="adm-gdel" type="button" aria-label="Xóa ảnh ${i + 1}">${icClose}</button></li>`
                    )
                    .join("")}<li><label class="adm-gadd" for="pf-file">${ic.plus}<span>Thêm ảnh</span></label></li></ul>`
                : `<div class="adm-gempty"><span aria-hidden="true">${icImage}</span><p><b>Chưa có ảnh</b>Có thể thêm nhiều ảnh; ảnh không bắt buộc.</p><label class="btn btn-sm btn-outline-brand" for="pf-file">${ic.plus}Thêm ảnh</label></div>`
            }
            <input id="pf-file" class="sr-only" type="file" accept="image/*" multiple>
            <div class="adm-field adm-urlrow">
              <label for="pf-img">Hoặc thêm ảnh từ URL</label>
              <div class="adm-inline"><input id="pf-img" class="input" type="url" placeholder="https://…" aria-describedby="pf-img-help"><button class="btn" type="button">Thêm</button></div>
              <p class="adm-help" id="pf-img-help">Ảnh chính hiển thị trên thẻ sản phẩm; các ảnh còn lại hiển thị ở trang chi tiết.</p>
            </div>
          </div>
        </section>`;

  const crumbLast = edit ? "Áo thun basic" : "Thêm sản phẩm";
  const titles = { create: "Thêm sản phẩm", edit: "Sửa sản phẩm", errors: "Thêm sản phẩm – lỗi" };
  return adminShell({
    title: `${titles[mode]} – SaleScale Quản trị`,
    desc: edit ? "Sửa thông tin, hình ảnh và SKU của sản phẩm." : "Tạo sản phẩm mới với thông tin, hình ảnh và danh sách SKU.",
    active: "products",
    extraCss: CSS,
    body: `      ${crumbs([["Sản phẩm", "AdminProducts.html"], [crumbLast]])}
      <div class="admin-head adm-head">
        <div><h1>${edit ? `Sửa sản phẩm <span class="badge badge-success">Đang bán</span>` : "Thêm sản phẩm"}</h1><p>${edit ? "Áo thun basic · 4 SKU · cập nhật lần cuối 18:02, 10/10/2026" : "Trường có dấu * là bắt buộc."}</p></div>
        ${edit ? `<div class="actions"><button class="btn adm-btn-danger" type="button">Ngừng bán sản phẩm</button></div>` : ""}
      </div>
      ${errs ? errorSummary("Chưa lưu được sản phẩm", "Có 3 lỗi cần sửa: thiếu tên sản phẩm, mã SKU bị trùng và giá không lớn hơn 0.", [["#pf-name", "Vui lòng nhập tên sản phẩm"], ["#pf-r3-code", "Dòng 3: mã SKU đã tồn tại"], ["#pf-r2-price", "Dòng 2: giá phải lớn hơn 0"]]) : ""}
      <form class="adm-form adm-pf" action="#" novalidate>
        <div class="adm-pf-top">
        <section class="panel" aria-labelledby="pf-info">
          <div class="panel-head adm-ph"><h2 class="panel-title" id="pf-info">Thông tin sản phẩm</h2></div>
          <div class="adm-grid adm-grid-2">
            ${field({ id: "pf-name", label: "Tên sản phẩm", req: true, value: name, cls: "span-2", placeholder: "Ví dụ: Áo thun basic", error: errs ? "Vui lòng nhập tên sản phẩm" : "" })}
            ${field({ id: "pf-brand", label: "Thương hiệu", value: errs ? "" : "Basico", placeholder: "Ví dụ: Basico", help: "Hiển thị trên thẻ sản phẩm.", extra: ` <span class="badge badge-brand">Mới ở v2</span>` })}
            <div class="adm-field">
              <label for="pf-cat">Danh mục <span class="adm-req" aria-hidden="true">*</span></label>
              <select id="pf-cat" class="select" required><option value="fashion" selected>Thời trang</option><option value="tech">Điện tử</option><option value="home">Gia dụng</option><option value="beauty">Làm đẹp</option></select>
            </div>
            <div class="adm-field span-2">
              <label for="pf-desc">Mô tả</label>
              <textarea id="pf-desc" class="input adm-textarea" rows="4">Áo thun cotton co giãn nhẹ, form regular, cổ tròn bo viền.</textarea>
            </div>
          </div>
        </section>
        ${gallery}
        </div>

        <section class="panel" aria-labelledby="pf-skus">
          <div class="panel-head adm-ph"><h2 class="panel-title" id="pf-skus">SKU <span class="adm-count">${skus.length}</span></h2></div>
          <ul class="adm-rules">${edit ? `<li>Mã SKU không đổi sau khi tạo.</li><li>Chỉnh tồn kho ở trang <a href="Inventory.html">Tồn kho</a>.</li>` : `<li>Mã SKU không được trùng.</li><li>Tồn ban đầu chỉ nhập khi tạo mới.</li>`}</ul>
          <div class="table-wrap">
            <table class="table adm-cards adm-pftable">
              <thead><tr><th scope="col">Mã SKU <span class="adm-req" aria-hidden="true">*</span></th><th scope="col">Thuộc tính</th><th scope="col">Giá thường (₫) <span class="adm-req" aria-hidden="true">*</span></th><th scope="col">Tối đa mỗi đơn</th><th scope="col">${edit ? "Trạng thái" : "Tồn ban đầu"}</th>${edit ? "" : `<th scope="col"><span class="sr-only">Thao tác</span></th>`}</tr></thead>
              <tbody>
              ${rows}
              </tbody>
            </table>
          </div>
          <div class="adm-addrow"><button class="btn btn-sm btn-outline-brand" type="button">${ic.plus}Thêm SKU</button></div>
        </section>

        <div class="adm-formbar">
          <span class="muted">Trường có dấu * là bắt buộc.</span>
          <div class="actions"><a class="btn" href="AdminProducts.html">Hủy</a><button class="btn btn-primary" type="submit">${edit ? "Lưu thay đổi" : "Lưu sản phẩm"}</button></div>
        </div>
      </form>`,
  });
}

export default [
  { file: "AdminFlashSales.html", title: "Danh sách flash sale", html: adminFlashSalesPage() },
  { file: "FlashSaleForm.html", title: "Chương trình – nháp", html: flashSaleFormPage("draft") },
  { file: "FlashSaleForm-ended.html", title: "Chương trình – đã kết thúc", html: flashSaleFormPage("ended") },
  { file: "FlashSaleForm-errors.html", title: "Chương trình – lỗi", html: flashSaleFormPage("errors") },
  { file: "FlashSaleForm-live.html", title: "Chương trình – đang diễn ra", html: flashSaleFormPage("live") },
  { file: "FlashSaleForm-scheduled.html", title: "Chương trình – đã lên lịch", html: flashSaleFormPage("scheduled") },
  { file: "Inventory.html", title: "Tồn kho", html: inventoryPage("list") },
  { file: "Inventory-adjust.html", title: "Tồn kho – điều chỉnh", html: inventoryPage("adjust") },
  { file: "Inventory-below-held.html", title: "Tồn kho – dưới lượng đang giữ", html: inventoryPage("below-held") },
  { file: "Inventory-conflict.html", title: "Tồn kho – xung đột", html: inventoryPage("conflict") },
  { file: "Inventory-history.html", title: "Tồn kho – lịch sử", html: inventoryPage("history") },
  { file: "ProductForm.html", title: "Thêm sản phẩm", html: productFormPage("create") },
  { file: "ProductForm-edit.html", title: "Sửa sản phẩm", html: productFormPage("edit") },
  { file: "ProductForm-errors.html", title: "Thêm sản phẩm – lỗi", html: productFormPage("errors") },
];
