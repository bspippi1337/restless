# Restless

## 🚀 Live Interactive Demo

[![Launch Live Demo](https://img.shields.io/badge/Launch-Live_Demo-238636?style=for-the-badge)](https://bspippi1337.github.io/restless/demo/)

👉 Click the button above to run Restless in your browser.

---

## Install (Debian)

```bash
echo "deb [trusted=yes] https://bspippi1337.github.io/restless/ ./" | sudo tee /etc/apt/sources.list.d/restless.list
sudo apt update
sudo apt install restless
```

---

## Example Usage

```bash
restless probe https://api.example.com
restless simulate https://api.example.com
restless smart https://api.example.com
```

---

## Architecture

smartcmd → discover → engine → simulator → export → app
