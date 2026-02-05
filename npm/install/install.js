"use strict";

const fs = require("fs");
const path = require("path");
const https = require("https");
const { spawnSync } = require("child_process");

const OWNER = "blckswan1337";
const REPO = "restless";

const pkg = JSON.parse(fs.readFileSync(path.join(__dirname, "..", "package.json"), "utf8"));
const version = pkg.version;
const tag = `v${version}`;

function mapPlatform() {
  const p = process.platform;
  if (p === "win32") return "windows";
  if (p === "darwin") return "darwin";
  if (p === "linux") return "linux";
  throw new Error(`Unsupported platform: ${p}`);
}

function mapArch() {
  const a = process.arch;
  if (a === "x64") return "amd64";
  if (a === "arm64") return "arm64";
  throw new Error(`Unsupported arch: ${a}`);
}

function assetName() {
  return `restless_${version}_${mapPlatform()}_${mapArch()}.tar.gz`;
}

function download(url, dest) {
  return new Promise((resolve, reject) => {
    https.get(url, { headers: { "User-Agent": "restless-npm-installer" } }, (res) => {
      if ([301,302,307,308].includes(res.statusCode)) return resolve(download(res.headers.location, dest));
      if (res.statusCode !== 200) return reject(new Error(`HTTP ${res.statusCode} for ${url}`));
      fs.mkdirSync(path.dirname(dest), { recursive: true });
      const file = fs.createWriteStream(dest);
      res.pipe(file);
      file.on("finish", () => file.close(resolve));
    }).on("error", reject);
  });
}

function extractTarGz(tarGzPath, outDir) {
  fs.mkdirSync(outDir, { recursive: true });
  const r = spawnSync("tar", ["-xzf", tarGzPath, "-C", outDir], { stdio: "inherit" });
  if (r.status !== 0) throw new Error("Failed to extract archive using tar.");
}

(async () => {
  const outDir = path.join(__dirname, "bin");
  const tmpDir = path.join(__dirname, "tmp");
  fs.mkdirSync(tmpDir, { recursive: true });

  const asset = assetName();
  const url = `https://github.com/${OWNER}/${REPO}/releases/download/${tag}/${asset}`;
  const archive = path.join(tmpDir, asset);

  console.log(`[restless] downloading ${url}`);
  await download(url, archive);

  console.log("[restless] extracting");
  extractTarGz(archive, outDir);

  if (process.platform !== "win32") {
    try { fs.chmodSync(path.join(outDir, "restless"), 0o755); } catch (_) {}
  }

  console.log("[restless] ready");
})().catch((e) => {
  console.error("[restless] install failed:", e.message);
  console.error("Hint: create a GitHub Release for tag", tag, "with matching assets.");
  process.exit(1);
});
