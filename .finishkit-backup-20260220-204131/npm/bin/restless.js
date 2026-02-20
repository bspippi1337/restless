#!/usr/bin/env node
"use strict";

const path = require("path");
const { spawnSync } = require("child_process");

const exe = process.platform === "win32" ? "restless.exe" : "restless";
const binPath = path.join(__dirname, "..", "install", "bin", exe);

const r = spawnSync(binPath, process.argv.slice(2), { stdio: "inherit" });
process.exit(r.status ?? 1);
