# Restless

Deterministic reduction of complex system state.

---

## Standalone

```bash
restless explain POST:/orders --input failing.json
```

Reduce a failing API interaction to its minimal, verifiable cause.

---

## Unix-native

```bash
jq -c . failing.json | restless explain POST:/orders
```

Insert `restless` into your pipeline to turn normalized input into structured explanation.

---

## Install

```bash
git clone https://github.com/<you>/restless.git
cd restless
./install.sh
```

---

## Usage

```bash
restless --help
```

---

## Principles

- Deterministic output
- Minimal surface area
- Composable by design
- Readable by humans

---

## License

MIT
