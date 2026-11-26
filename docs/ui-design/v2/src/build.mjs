// Builds every SaleScale v2 static mockup into ../pages.
// Each module in ./pages exports `group` (index heading) and a default array of { file, title, html }.
// Usage: node docs/ui-design/v2/src/build.mjs
import { writeFileSync, mkdirSync, readdirSync, rmSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";
import { head, logo, ic } from "./lib.mjs";

const here = dirname(fileURLToPath(import.meta.url));
const outDir = join(here, "..", "pages");
rmSync(outDir, { recursive: true, force: true });
mkdirSync(outDir, { recursive: true });

const groups = [];
const seen = new Set();
for (const name of readdirSync(join(here, "pages")).filter((f) => f.endsWith(".mjs")).sort()) {
  const mod = await import(pathToFileURL(join(here, "pages", name)).href);
  for (const page of mod.default) {
    if (seen.has(page.file)) throw new Error(`duplicate page ${page.file} in ${name}`);
    seen.add(page.file);
    writeFileSync(join(outDir, page.file), page.html);
  }
  groups.push({ group: mod.group, pages: mod.default });
}

const index = `${head("SaleScale UI v2", "blue", "Mục lục các màn hình thiết kế SaleScale v2.")}
<body>
<main class="container" style="padding:32px 16px;max-width:960px">
  ${logo("UI v2")}
  <h1 style="font-size:24px;margin:24px 0 4px">Bộ màn hình SaleScale v2</h1>
  <p class="muted" style="margin:0 0 20px">${seen.size} màn hình. Ảnh sản phẩm là placeholder; số liệu là dữ liệu mẫu. Bản cũ nằm ở <code>docs/ui-design/pages</code>.</p>
  ${groups
    .map(
      (g) => `<section class="panel" style="margin-bottom:16px"><div class="panel-head"><h2 class="panel-title">${g.group}</h2><span class="muted">${g.pages.length} trang</span></div>
  ${g.pages.map((p) => `<a href="${p.file}" style="display:flex;justify-content:space-between;gap:16px;padding:10px 16px;border-top:1px solid var(--line)"><span>${p.title}<span class="muted" style="font-size:12px"> · ${p.file}</span></span><span class="link-more">Mở${ic.chevron}</span></a>`).join("")}</section>`
    )
    .join("\n  ")}
</main>
</body>
</html>
`;
writeFileSync(join(outDir, "index.html"), index);
console.log(`built ${seen.size} pages + index`);
