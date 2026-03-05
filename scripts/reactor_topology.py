import json,sys

rep=json.load(open(sys.argv[1]))

host=rep["target"]
paths=[e["path"] for e in rep["endpoints"]]

print(host)

for p in paths:
    p=p.strip("/")
    if not p:
        continue

    depth=p.count("/")
    indent=" "*(depth*2)

    print(f"{indent}├── {p}")
