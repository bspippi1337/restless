package main

import "github.com/bspippi1337/restless/internal/teacher"

import (
	"context"
	"flag"
	"fmt"
	"github.com/bspippi1337/restless/internal/version"
	"net/http"
	"os"

	"github.com/bspippi1337/restless/internal/core/app"
	"github.com/bspippi1337/restless/internal/core/types"
	"github.com/bspippi1337/restless/internal/modules/bench"
	"github.com/bspippi1337/restless/internal/modules/export"
	"github.com/bspippi1337/restless/internal/modules/openapi"
	"github.com/bspippi1337/restless/internal/modules/session"
)

func main() {

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "teacher":
			teacher.Run()
			return
		case "--version":
			fmt.Println("restless", version.Version)
			return
		case "openapi":
			handleOpenAPI(os.Args[2:])
			return
		case "profile":
			handleProfile(os.Args[2:])
			return
		}
	}

	runRequestMode(os.Args[1:])
}

func runRequestMode(args []string) {
	fs := flag.NewFlagSet("request", flag.ExitOnError)

	method := fs.String("X", "GET", "HTTP method")
	url := fs.String("url", "", "Request URL")
	body := fs.String("d", "", "Body string")

	fs.Parse(args)

	if *url == "" {
		fmt.Println("missing -url")
		fs.Usage()
		os.Exit(1)
	}

	sess := session.New()
	mods := []app.Module{
		sess,
		openapi.New(),
		export.New(),
		bench.New(),
	}

	a, err := app.New(mods)
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}

	req := types.Request{
		Method:  *method,
		URL:     *url,
		Headers: http.Header{},
		Body:    []byte(*body),
	}

	resp, err := a.RunOnce(context.Background(), req)
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}

	fmt.Printf("status: %d (dur=%dms)\n", resp.StatusCode, resp.DurationMs)
	fmt.Println(string(resp.Body))
}
