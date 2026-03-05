#!/bin/bash
set -e
mkdir -p assets/brand
cp *.svg assets/brand/
cp BRAND_GUIDE.md assets/brand/
git add assets/brand
git commit -m "Add Restless brand pack"
echo "Brand installed."
