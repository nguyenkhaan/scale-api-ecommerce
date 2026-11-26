import { vnd, pct, compact, img, ic, logo, P, rating, quotaBar, flashItem, productCard, head, header, footer, countdown, adminShell, customerPage, mobileFrame, steps } from "../lib.mjs";

export const group = "Bản thử đã duyệt";

function mainPage(theme) {
  const flash = [P.tee, P.phones, P.power, P.fryer, P.sun, P.cable];
  const grid = [P.tee, P.phones, P.bottle, P.power, P.fryer, P.linen, P.cable, P.sun, P.lamp, P.sandal, P.case, P.clean, P.bottle, P.linen, P.lamp];
  const cats = [
    ["cat-sale", "Đang flash sale", 6, false],
    ["cat-fashion", "Thời trang", 38, true],
    ["cat-electronics", "Điện tử", 52, false],
    ["cat-home", "Gia dụng", 27, false],
    ["cat-beauty", "Làm đẹp", 19, false],
    ["cat-accessory", "Phụ kiện", 24, false],
  ];
  return `${head("SaleScale – Mua sắm flash sale", theme, "Trang chủ SaleScale: flash sale đang diễn ra và danh sách sản phẩm.")}
<body>
${header()}
<main class="container home page">
  <aside class="side" aria-label="Danh mục">
    <div class="panel">
      <h3>Danh mục</h3>
      <ul class="side-list">${cats.map(([i, n, c, a]) => `<li><a href="#"${a ? ' class="active" aria-current="true"' : ""}><img class="cat-ico" src="${img(i)}" alt="">${n}<span class="count">${c}</span></a></li>`).join("")}</ul>
    </div>
    <div class="panel">
      <h3>Lịch flash sale hôm nay</h3>
      <ul class="side-list">
        <li><a href="FlashSaleDetail.html"><span>12:00 – 14:00</span><span class="count">Đã kết thúc</span></a></li>
        <li><a href="FlashSaleDetail.html" class="active"><span>20:00 – 22:00</span><span class="count" style="color:var(--sale);font-weight:600">Đang diễn ra</span></a></li>
        <li><a href="FlashSaleDetail-upcoming.html"><span>Mai 09:00</span><span class="count">Sắp diễn ra</span></a></li>
      </ul>
    </div>
  </aside>

  <div class="home-main">
    <section class="promo-row" aria-label="Chương trình nổi bật">
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
    </section>

    <section class="flash" aria-labelledby="flash-title">
      <div class="flash-head">
        <h2 id="flash-title" class="flash-title">${ic.bolt}Giá sốc hôm nay</h2>
        <div class="flash-meta">Kết thúc sau ${countdown("01", "24", "30")}</div>
        <span class="clock-note">${ic.info}Theo giờ hệ thống</span>
        <a class="link-more" href="FlashSaleDetail.html">Xem tất cả${ic.chevron}</a>
      </div>
      <div class="flash-row">${flash.map(flashItem).join("")}</div>
    </section>

    <section class="panel" aria-labelledby="all-title">
      <div class="catalog-bar">
        <div class="catalog-top"><h2 id="all-title">Gợi ý cho bạn</h2><label class="sort">Sắp xếp<select class="select"><option>Phổ biến</option><option>Bán chạy</option><option>Giá thấp đến cao</option><option>Giá cao đến thấp</option></select></label></div>
        <a class="chip on" href="#" aria-pressed="true">Tất cả</a>
        <a class="chip" href="#">${ic.bolt}Đang flash sale</a>
        <a class="chip" href="#">Thời trang</a>
        <a class="chip" href="#">Điện tử</a>
        <a class="chip" href="#">Gia dụng</a>
        <a class="chip" href="#">Làm đẹp</a>
      </div>
    </section>
    <div class="grid-products">${grid.map(productCard).join("")}</div>
    <div class="load-more"><button class="btn btn-outline-brand" type="button">Xem thêm sản phẩm</button></div>
  </div>
</main>
${footer}
</body>
</html>
`;
}

function productDetailPage() {
  const p = P.tee;
  const similar = [P.linen, P.sandal, P.phones, P.bottle, P.clean, P.lamp, P.cable];
  return `${head("Áo thun cotton basic – SaleScale", "blue", "Chi tiết sản phẩm với flash sale, chọn phân loại và nguồn giá.")}
<body class="has-mobile-bar">
${header()}
<main class="container page">
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
          <label><input type="radio" name="src" checked><span>Giá flash sale<small>Còn ${p.quota[1] - p.quota[0]} suất · tối đa ${p.limit}/tài khoản</small></span><span class="amt" style="color:var(--price)">${vnd(p.sale)}</span></label>
          <label><input type="radio" name="src"><span>Giá thường<small>Không giới hạn suất sale</small></span><span class="amt">${vnd(p.base)}</span></label>
        </fieldset>
        <div class="buy-qty"><span class="opt-label">Số lượng</span><div class="qty"><button type="button" aria-label="Giảm" disabled>−</button><input aria-label="Số lượng" value="1" inputmode="numeric"><button type="button" aria-label="Tăng">+</button></div></div>
        <div class="buy-total"><span>Tạm tính</span><span class="price">${vnd(p.sale)}</span></div>
        <div class="buy-actions">
          <a class="btn btn-sale btn-lg btn-block" href="ProductDetail-added.html">Thêm vào giỏ</a>
          <a class="btn btn-outline-brand btn-block" href="Cart.html">Xem giỏ hàng</a>
        </div>
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
  </section>
</main>
${footer}
<div class="mobile-bar"><div class="grow"><span class="price">${vnd(p.sale)}</span><small>Flash sale · còn ${p.quota[1] - p.quota[0]} suất</small></div><a class="btn btn-sale" href="ProductDetail-added.html">Thêm vào giỏ</a></div>
</body>
</html>
`;
}

function cartLine(p, opts) {
  const { variant, sku, qty, flash, checked = true, maxNote, off } = opts;
  const unit = flash ? p.sale : p.base;
  return `<div class="cart-line${off ? " line-off" : ""}">
  <input class="check" type="checkbox" ${checked ? "checked" : ""} aria-label="Chọn ${p.short}">
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
  <div class="line-price"><span class="price">${vnd(unit)}</span>${flash ? `<span class="price-old">${vnd(p.base)}</span>` : ""}</div>
  <div class="line-qty"><div class="qty"><button type="button" aria-label="Giảm">−</button><input aria-label="Số lượng ${p.short}" value="${qty}"><button type="button" aria-label="Tăng"${maxNote ? " disabled" : ""}>+</button></div>${maxNote ? `<span class="limit-note">${maxNote}</span>` : ""}</div>
  <span class="line-total">${vnd(unit * qty)}</span>
  <button class="icon-btn" type="button" aria-label="Xóa ${p.short}">${ic.trash}</button>
</div>`;
}

function cartPage() {
  const subtotal = P.tee.sale * 2 + P.phones.sale * 1;
  const saving = (P.tee.base - P.tee.sale) * 2 + (P.phones.base - P.phones.sale);
  return `${head("Giỏ hàng – SaleScale", "blue", "Giỏ hàng nhiều SKU, nhóm theo flash sale và giá thường.")}
<body class="has-mobile-bar">
${header(3)}
<main class="container page">
  <div class="cart-title"><h1>Giỏ hàng</h1><span class="muted">3/20 sản phẩm</span></div>
  <div class="cart-wrap">
    <div>
      <div class="notice notice-info">${ic.lock}<span>Sản phẩm trong giỏ <b>chưa được giữ</b>. Hàng và suất flash sale chỉ được giữ khi bạn đặt đơn thành công; giá có thể thay đổi trước bước đó.</span></div>
      <div class="panel" style="margin-top:12px">
        <div class="cart-cols"><label><input class="check" type="checkbox" aria-label="Chọn tất cả"></label><span>Tất cả (3 sản phẩm)</span><span>Đơn giá</span><span>Số lượng</span><span>Thành tiền</span><button class="icon-btn" type="button" aria-label="Xóa các dòng đã chọn">${ic.trash}</button></div>
      </div>
      <section class="panel cart-group" aria-labelledby="g1">
        <div class="cart-group-head"><input class="check" type="checkbox" checked aria-label="Chọn nhóm flash sale"><span id="g1" style="color:var(--sale);display:inline-flex;gap:6px;align-items:center">${ic.bolt}Flash Sale 10.10</span><span class="badge badge-sale-soft">Kết thúc 22:00</span><span class="clock-note">${ic.info}Theo giờ hệ thống</span></div>
        ${cartLine(P.tee, { variant: "Đen / M", sku: "AT-DEN-M", qty: 2, flash: true, maxNote: "Tối đa 2/tài khoản" })}
        ${cartLine(P.phones, { variant: "Đen", sku: "TN-X2-DEN", qty: 1, flash: true, maxNote: "Tối đa 1/tài khoản" })}
      </section>
      <section class="panel cart-group" aria-labelledby="g2">
        <div class="cart-group-head"><input class="check" type="checkbox" aria-label="Chọn nhóm giá thường"><span id="g2">Giá thường</span></div>
        ${cartLine(P.bottle, { variant: "Bạc", sku: "BGN-500-BAC", qty: 1, flash: false, checked: false })}
      </section>
    </div>

    <aside class="cart-aside" aria-label="Tóm tắt">
      <div class="panel"><ol class="steps"><li class="now">Giỏ hàng</li><li>Xem trước đơn</li><li>Thanh toán</li></ol></div>
      <div class="panel summary">
        <div class="summary-row"><span>Tạm tính (2 dòng)</span><span>${vnd(subtotal + saving)}</span></div>
        <div class="summary-row"><span>Giảm giá flash sale</span><span style="color:var(--success)">−${vnd(saving)}</span></div>
        <div class="summary-total"><b>Tổng tham khảo</b><span><span class="price">${vnd(subtotal)}</span><small>Giá cuối cùng hiện ở bước xem trước đơn</small></span></div>
        <a class="btn btn-sale btn-lg btn-block" href="Preview.html">Xem trước đơn (2)</a>
      </div>
      <div class="notice notice-warning">${ic.info}<span>Mỗi lần đặt đơn tối đa 20 dòng. Dòng giá thường chưa chọn sẽ được giữ lại trong giỏ.</span></div>
    </aside>
  </div>
</main>
${footer}
<div class="mobile-bar"><div class="grow"><span class="price">${vnd(subtotal)}</span><small>2 dòng · tổng tham khảo</small></div><a class="btn btn-sale" href="Preview.html">Xem trước đơn</a></div>
</body>
</html>
`;
}

function adminProductsPage() {
  const rows = [
    [P.tee, "Thời trang", 12, "129.000 – 199.000 ₫", 420, 86, 5320, true],
    [P.phones, "Điện tử", 1, "590.000 – 890.000 ₫", 34, 12, 1870, true],
    [P.bottle, "Gia dụng", 2, "250.000 ₫", 312, 4, 9640, true],
    [P.power, "Điện tử", 1, "299.000 – 350.000 ₫", 96, 18, 3105, true],
    [P.fryer, "Gia dụng", 1, "990.000 – 1.290.000 ₫", 9, 6, 1402, true],
    [P.linen, "Thời trang", 8, "329.000 ₫", 188, 2, 764, true],
    [P.case, "Điện tử", 3, "59.000 ₫", 0, 0, 610, false],
    [P.sandal, "Thời trang", 6, "99.000 ₫", 240, 0, 4210, true],
  ];
  return adminShell({ title: "Sản phẩm – SaleScale Quản trị", desc: "Quản trị danh sách sản phẩm, tồn kho và trạng thái bán.", active: "products", body: `      <div class="admin-head">
        <div><h1>Sản phẩm</h1><p>Sản phẩm ngừng bán vẫn được giữ để tra cứu đơn cũ, không xóa hẳn.</p></div>
        <div class="actions"><a class="btn" href="Inventory.html">Điều chỉnh tồn</a><a class="btn btn-primary" href="ProductForm.html">${ic.plus}Thêm sản phẩm</a></div>
      </div>

      <section class="kpis" aria-label="Tổng quan">
        <div class="kpi"><span>Đang bán</span><b>46</b><small>trên 48 sản phẩm · 112 SKU</small></div>
        <div class="kpi"><span>Đang giữ cho đơn chờ thanh toán</span><b>128</b><small>sản phẩm · 41 đơn</small></div>
        <div class="kpi"><span>Sắp hết hàng</span><b>5</b><small class="warn">SKU còn dưới 10 khả dụng</small></div>
        <div class="kpi"><span>Trong flash sale đang diễn ra</span><b>6</b><small>Flash Sale 10.10 · kết thúc 22:00</small></div>
      </section>

      <section class="panel" aria-label="Danh sách sản phẩm">
        <div class="tabs" role="tablist">
          <a class="tab active" role="tab" aria-selected="true" href="#">Tất cả<span class="count">48</span></a>
          <a class="tab" role="tab" aria-selected="false" href="#">Đang bán<span class="count">46</span></a>
          <a class="tab" role="tab" aria-selected="false" href="#">Ngừng bán<span class="count">2</span></a>
          <a class="tab" role="tab" aria-selected="false" href="#">Hết hàng<span class="count">1</span></a>
        </div>
        <div class="filters">
          <label class="sr-only" for="f1">Tìm sản phẩm</label><input id="f1" class="input" placeholder="Tên sản phẩm hoặc mã SKU">
          <label class="sr-only" for="f2">Danh mục</label><select id="f2" class="select"><option>Tất cả danh mục</option><option>Thời trang</option><option>Điện tử</option><option>Gia dụng</option><option>Làm đẹp</option></select>
          <label class="sr-only" for="f3">Thương hiệu</label><select id="f3" class="select"><option>Tất cả thương hiệu</option></select>
          <span class="result">Hiển thị 1–8 / 48</span>
        </div>
        <div class="table-wrap">
          <table class="table">
            <thead><tr><th scope="col">Sản phẩm</th><th scope="col">Danh mục</th><th scope="col" class="r">SKU</th><th scope="col" class="r">Giá</th><th scope="col">Tồn kho</th><th scope="col" class="r">Đã bán</th><th scope="col">Trạng thái</th><th scope="col"><span class="sr-only">Thao tác</span></th></tr></thead>
            <tbody>
              ${rows
                .map(([p, cat, skus, price, avail, held, sold, on]) => {
                  const total = avail + held || 1;
                  return `<tr>
                <td><div class="prod-cell"><div class="thumb"><img src="${img(p.img)}" alt=""></div><div><b>${p.short}</b><small>${p.brand}${p.sale ? ` · <span style="color:var(--sale)">Trong flash sale</span>` : ""}</small></div></div></td>
                <td>${cat}</td>
                <td class="r">${skus}</td>
                <td class="r">${price}</td>
                <td><div class="stock"><span class="num"><b>${avail}</b> khả dụng${held ? ` · ${held} đang giữ` : ""}</span><div class="stock-bar"><i class="avail" style="width:${(avail / total) * 100}%"></i><i class="held" style="width:${(held / total) * 100}%"></i></div>${avail > 0 && avail < 10 ? `<small style="color:var(--warning)">Sắp hết</small>` : avail === 0 ? `<small>Hết hàng</small>` : ""}</div></td>
                <td class="r">${sold.toLocaleString("vi-VN").replace(/,/g, ".")}</td>
                <td><span class="status ${on ? "on" : "off"}">${on ? "Đang bán" : "Ngừng bán"}</span></td>
                <td><div class="row-actions"><a href="ProductForm.html">Sửa</a>${on ? `<button type="button" class="danger">Ngừng bán</button>` : `<button type="button">Mở bán lại</button>`}</div></td>
              </tr>`;
                })
                .join("")}
            </tbody>
          </table>
        </div>
        <div class="pager"><span>Mỗi trang 8 sản phẩm</span><div class="btns"><button class="btn btn-sm" type="button" disabled>‹</button><button class="btn btn-sm on" type="button" aria-current="page">1</button><button class="btn btn-sm" type="button">2</button><button class="btn btn-sm" type="button">3</button><span style="padding:6px">…</span><button class="btn btn-sm" type="button">6</button><button class="btn btn-sm" type="button" aria-label="Trang sau">›</button></div></div>
      </section>`,
  });
}


export default [
  { file: "Main.html", title: "Trang chủ", html: mainPage("blue") },
  { file: "ProductDetail.html", title: "Chi tiết sản phẩm", html: productDetailPage() },
  { file: "Cart.html", title: "Giỏ hàng", html: cartPage() },
  { file: "AdminProducts.html", title: "Quản trị sản phẩm", html: adminProductsPage() },
];
