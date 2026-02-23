# Restless

## 🚀 Live Interactive Demo

<iframe src="https://bspippi1337.github.io/restless/demo/"
        width="100%"
        height="520"
        style="border:none;border-radius:8px;">
</iframe>

If it does not render inside GitHub, open directly:

https://bspippi1337.github.io/restless/demo/

---

## Install (Debian)

```bash
echo "deb [trusted=yes] https://bspippi1337.github.io/restless/ ./" | sudo tee /etc/apt/sources.list.d/restless.list
sudo apt update
sudo apt install restless
```

---

## Example

```bash
restless probe https://api.example.com
restless simulate https://api.example.com
restless smart https://api.example.com
```
