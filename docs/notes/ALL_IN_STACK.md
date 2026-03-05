All-in Distribution Stack (Restless)
====================================

This bundle adds:
- Makefile targets: build/install/uninstall/release/docker/deb/aptrepo
- Dockerfile
- GitHub Actions: CI + Release + Docker publish to GHCR
- Debian packaging control + scripts to build .deb and a minimal APT repo
- Homebrew formula template

Apply
-----
Unzip into repo root, then:

  rm -f internal/core/discover/discover.go   # if you have legacy discovery engine
  gofmt -w . || true
  make doctor
  make build

Install locally
---------------
  sudo make install

Generate completion (already done by make install)
--------------------------------------------------
  make completion
  source dist/completion/restless.bash  # bash
  fpath+=(dist/completion); autoload -Uz compinit && compinit  # zsh

Docker
------
  make docker
  docker run --rm restless:$(git describe --tags --always --dirty) --help

Debian package
--------------
  make deb
  sudo dpkg -i dist/restless_*_amd64.deb

APT repo (unsigned, trusted=yes)
--------------------------------
  make aptrepo
  python3 -m http.server 8000 --directory .apt-repo
  echo 'deb [trusted=yes] http://127.0.0.1:8000 stable main' | sudo tee /etc/apt/sources.list.d/restless.list
  sudo apt update
  sudo apt install restless

GitHub Release
--------------
Push a tag vX.Y.Z and workflows will upload binaries and SHA256SUMS:
  git tag v0.1.0 && git push --tags

Docker publish to GHCR
----------------------
Tag push triggers GHCR publish:
  ghcr.io/<owner>/<repo>/restless:latest
