#!/usr/bin/env python3
"""Export Restless Logo B to PNG/ICO/iconset.

Usage:
  python3 scripts/brand_export_icons.py
"""
from pathlib import Path
import shutil
from PIL import Image
import cairosvg

ROOT = Path(__file__).resolve().parents[1]
svg_path = ROOT / "assets/brand/logo/restless_logo_B.svg"
exports = ROOT / "assets/brand/logo/exports"
png_dir = exports / "png"
ico_dir = exports / "ico"
iconset_dir = exports / "macos.iconset"

logo_svg = svg_path.read_text(encoding="utf-8")

png_dir.mkdir(parents=True, exist_ok=True)
ico_dir.mkdir(parents=True, exist_ok=True)
iconset_dir.mkdir(parents=True, exist_ok=True)

sizes = [16, 32, 48, 64, 128, 256, 512, 1024]
for s in sizes:
    out_png = png_dir / f"restless_logo_B_{s}.png"
    cairosvg.svg2png(bytestring=logo_svg.encode("utf-8"), write_to=str(out_png), output_width=s, output_height=s)

ico_sizes = [16, 32, 48, 64, 128, 256]
images = [Image.open(png_dir / f"restless_logo_B_{s}.png").convert("RGBA") for s in ico_sizes]
images[0].save(str(ico_dir / "restless.ico"), format="ICO", sizes=[(s,s) for s in ico_sizes])

def copy_iconset(base_size):
    shutil.copyfile(png_dir / f"restless_logo_B_{base_size}.png", iconset_dir / f"icon_{base_size}x{base_size}.png")
    shutil.copyfile(png_dir / f"restless_logo_B_{base_size*2}.png", iconset_dir / f"icon_{base_size}x{base_size}@2x.png")

for bs in [16, 32, 128, 256, 512]:
    copy_iconset(bs)

print("✅ Exported:", exports)
