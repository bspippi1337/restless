<!-- README.md -->

<p align="center">
  <img src="assets/restless-terminal.svg" alt="Restless Live Terminal" width="900" />
</p>

<h1 align="center">Restless ⚡</h1>

<p align="center">
  <b>Universal API blade: discover, probe, simulate, and export.</b><br/>
  A CLI/TUI for exploring APIs fast, with an interactive browser demo.
</p>

<p align="center">
  <a href="https://github.com/bspippi1337/restless/releases"><img alt="Release" src="https://img.shields.io/github/v/release/bspippi1337/restless?sort=semver"/></a>
  <a href="https://github.com/bspippi1337/restless/actions/workflows/ci.yml"><img alt="CI" src="https://github.com/bspippi1337/restless/actions/workflows/ci.yml/badge.svg"/></a>
  <a href="LICENSE"><img alt="License" src="https://img.shields.io/badge/license-MIT-blue.svg"/></a>
</p>

---

## Live Interactive Demo (in browser)

[![Launch Live Demo](https://img.shields.io/badge/Launch-Live_Demo-238636?style=for-the-badge)](https://bspippi1337.github.io/restless/demo/)

Click the button above to run Restless directly in your browser. :contentReference[oaicite:2]{index=2}

---

## Install

### Debian / Ubuntu (APT)

```bash
echo "deb [trusted=yes] https://bspippi1337.github.io/restless/ ./" | sudo tee /etc/apt/sources.list.d/restless.list
sudo apt update
sudo apt install restless
