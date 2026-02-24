package main

import (
    "github.com/bspippi1337/restless/internal/ui/cli"
)

func main() {
    root := cli.NewRootCmd()
    root.Execute()
}
