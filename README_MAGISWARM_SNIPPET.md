
## Magiswarm

One command recon session:

```bash
./build/restless magiswarm https://api.github.com
```

Tune it:

```bash
./build/restless magiswarm https://api.github.com   --concurrency 12   --max-requests 400   --timeout 4s   --out dist   --wordlist wordlists/api_paths.txt
```

Add headers (auth):

```bash
./build/restless magiswarm https://example.com   --header "Authorization: Bearer $TOKEN"
```

Outputs:

- `dist/magiswarm_<host>_<timestamp>.json`
- `dist/magiswarm_<host>_<timestamp>.topology.txt`
