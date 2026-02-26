mkdir -p restless_brand_pack && cd restless_brand_pack

cat > 01_brutal_construct.svg << 'EOF'
<svg viewBox="0 0 512 512" xmlns="http://www.w3.org/2000/svg">
  <rect width="512" height="512" fill="#0B0F14"/>
  <path d="M140 380 V120 H300 L230 220 H300 L200 380 Z" fill="#FFFFFF"/>
</svg>
EOF

cat > 02_precision_grid.svg << 'EOF'
<svg viewBox="0 0 512 512" xmlns="http://www.w3.org/2000/svg">
  <rect width="512" height="512" fill="#0B0F14"/>
  <path d="M150 380 V120 H300 L250 200 H300 L200 380"
        stroke="#FFFFFF"
        stroke-width="36"
        fill="none"
        stroke-linejoin="miter"/>
</svg>
EOF

cat > 03_controlled_energy.svg << 'EOF'
<svg viewBox="0 0 512 512" xmlns="http://www.w3.org/2000/svg">
  <rect width="512" height="512" fill="#0B0F14"/>
  <path d="M180 120 L320 120 L250 240 L320 240 L180 380 L260 240 L180 240 Z"
        fill="#FFFFFF"/>
</svg>
EOF

cat > BRAND_GUIDE.md << 'EOF'
# RESTLESS Brand Guide

Primary Logos:
- 01_brutal_construct.svg
- 02_precision_grid.svg
- 03_controlled_energy.svg

Colors:
#0B0F14
#FFFFFF
#00C8FF

Rules:
No gradients.
No shadows.
No distortion.
Maintain strong contrast.
EOF

cat > install_brand.sh << 'EOF'
#!/bin/bash
set -e
mkdir -p assets/brand
cp *.svg assets/brand/
cp BRAND_GUIDE.md assets/brand/
git add assets/brand
git commit -m "Add Restless brand pack"
echo "Brand installed."
EOF

chmod +x install_brand.sh
