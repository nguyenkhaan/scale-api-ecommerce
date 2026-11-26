// Renders a single page module into ../pages without touching other files.
// Usage: node docs/ui-design/v2/src/render-one.mjs <module path relative to src, e.g. pages/10-catalog.mjs>
import { writeFileSync, mkdirSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

const here = dirname(fileURLToPath(import.meta.url));
const outDir = join(here, "..", "pages");
mkdirSync(outDir, { recursive: true });
const mod = await import(pathToFileURL(join(here, process.argv[2])).href + `?t=${Date.now()}`);
for (const page of mod.default) writeFileSync(join(outDir, page.file), page.html);
console.log(`rendered ${mod.default.length} pages: ${mod.default.map((p) => p.file).join(", ")}`);
