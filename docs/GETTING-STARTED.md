# Getting started

Install (APT):
```bash
echo "deb [trusted=yes] https://bspippi1337.github.io/restless/ ./" | sudo tee /etc/apt/sources.list.d/restless.list
sudo apt update
sudo apt install restless
```

Build from source:
```bash
go build -o restless ./cmd/restless
```

First run:
```bash
restless doctor
```

Discover + interact:
```bash
restless discover example.com --verify --budget-seconds 20 --budget-pages 8 --save-profile example
restless tui
```
