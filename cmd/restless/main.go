package main

import (
"fmt"
"os"

"github.com/bspippi1337/restless/internal"
)

func main() {
if len(os.Args) < 2 {
fmt.Println("Usage: restless <command>")
os.Exit(1)
}

switch os.Args[1] {
case "smart":
if len(os.Args) < 3 {
fmt.Println("Usage: restless smart <url>")
os.Exit(1)
}
err := internal.RunSmart(os.Args[2])
if err != nil {
fmt.Println("Error:", err)
os.Exit(1)
}
default:
fmt.Println("Unknown command:", os.Args[1])
os.Exit(1)
}
}
