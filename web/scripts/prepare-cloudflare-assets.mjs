import { cp, mkdir, readFile, rm, writeFile } from "node:fs/promises";
import path from "node:path";

const sitePath = process.argv[2];

if (!sitePath || !/^[a-z0-9-]+$/.test(sitePath)) {
  throw new Error("Expected a URL-safe site path argument.");
}

const source = path.resolve("out");
const destinationRoot = path.resolve(".cloudflare/assets");
const destination = path.join(destinationRoot, sitePath);

await rm(destinationRoot, { recursive: true, force: true });
await mkdir(destination, { recursive: true });
await cp(source, destination, { recursive: true });

// Wrangler reads _redirects only from the assets root, and the site is served
// under /<sitePath>, so move the rules there and prefix both sides of each one.
const siteRedirects = path.join(destination, "_redirects");
const rules = await readFile(siteRedirects, "utf8").catch((error) => {
  if (error.code === "ENOENT") return null;
  throw error;
});
if (rules !== null) {
  const prefixed = rules.split("\n").map((line) => {
    const rule = line.trim();
    if (!rule || rule.startsWith("#")) return rule;
    const [from, to, ...rest] = rule.split(/\s+/);
    if (!from.startsWith("/") || !to) throw new Error(`Unsupported _redirects rule: ${rule}`);
    const target = to.startsWith("/") ? `/${sitePath}${to}` : to;
    return [`/${sitePath}${from}`, target, ...rest].join(" ");
  });
  await writeFile(path.join(destinationRoot, "_redirects"), prefixed.join("\n"));
  await rm(siteRedirects);
}
