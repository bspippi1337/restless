import json,sys

data=json.load(open(sys.argv[1]))

nodes=set()
edges=[]

host=data["target"]

for e in data["endpoints"]:

    p=e["path"].strip("/")

    parent=host

    for seg in p.split("/"):

        cur=parent+"/"+seg

        edges.append((parent,cur))

        nodes.add(parent)
        nodes.add(cur)

        parent=cur

print("<svg xmlns='http://www.w3.org/2000/svg' width='900' height='600'>")

y=30

for n in nodes:

    print(f"<text x='20' y='{y}' font-family='monospace'>{n}</text>")
    y+=20

print("</svg>")
