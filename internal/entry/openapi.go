package entry

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	gdiff "github.com/bspippi1337/restless/internal/modules/openapi/guard/diff"
	gloader "github.com/bspippi1337/restless/internal/modules/openapi/guard/loader"
	"github.com/bspippi1337/restless/internal/modules/openapi/guard/model"
	greport "github.com/bspippi1337/restless/internal/modules/openapi/guard/report"
	gruntime "github.com/bspippi1337/restless/internal/modules/openapi/guard/runtime"
)

func OpenAPI(args []string) error {
	if len(args) == 0 {
		printOpenAPIHelp()
		return nil
	}

	switch args[0] {
	case "guard":
		return runGuard(args[1:])
	case "diff":
		return runDiff(args[1:])
	default:
		printOpenAPIHelp()
		return nil
	}
}

func printOpenAPIHelp() {
	fmt.Println("restless openapi")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  restless openapi guard <METHOD> <pathTemplate> <status> <contentType> <jsonFile> --spec <specRef>")
	fmt.Println("  restless openapi diff <oldSpec> <newSpec>")
}

func runGuard(args []string) error {
	// guard <METHOD> <pathTemplate> <status> <contentType> <jsonFile> --spec <specRef>
	if len(args) < 6 {
		return fmt.Errorf("usage: restless openapi guard <METHOD> <pathTemplate> <status> <contentType> <jsonFile> --spec <specRef>")
	}

	method := strings.ToUpper(args[0])
	path := args[1]
	status := mustAtoi(args[2])
	contentType := args[3]
	jsonFile := args[4]

	specIndex := indexOf(args, "--spec")
	if specIndex == -1 || specIndex+1 >= len(args) {
		return fmt.Errorf("--spec required")
	}
	specRef := args[specIndex+1]

	body, err := os.ReadFile(jsonFile)
	if err != nil {
		return err
	}

	ctx := context.Background()
	doc, err := gloader.Load(ctx, specRef, gloader.LoadOptions{AllowRemoteRefs: true})
	if err != nil {
		return err
	}

	v := gruntime.NewValidator(doc)
	findings, err := v.ValidateResponse(ctx, method, path, status, contentType, body)
	if err != nil {
		return err
	}

	res := model.GuardResult{
		SpecRef:    specRef,
		StartedAt:  time.Now(),
		FinishedAt: time.Now(),
		Findings:   findings,
	}
	res.CDI = gruntime.ComputeCDI(findings, gruntime.DefaultWeights())

	fmt.Print(greport.PrintHuman(res))
	return nil
}

func runDiff(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: restless openapi diff <oldSpec> <newSpec>")
	}

	ctx := context.Background()
	oldDoc, err := gloader.Load(ctx, args[0], gloader.LoadOptions{AllowRemoteRefs: true})
	if err != nil {
		return err
	}
	newDoc, err := gloader.Load(ctx, args[1], gloader.LoadOptions{AllowRemoteRefs: true})
	if err != nil {
		return err
	}

	res, err := gdiff.Diff(ctx, oldDoc, newDoc)
	if err != nil {
		return err
	}

	fmt.Println("Recommended bump:", res.RecommendedBump)
	if len(res.Breaking) > 0 {
		fmt.Println("Breaking:")
		for _, s := range res.Breaking {
			fmt.Println("  -", s)
		}
	}
	if len(res.NonBreaking) > 0 {
		fmt.Println("Non-breaking:")
		for _, s := range res.NonBreaking {
			fmt.Println("  -", s)
		}
	}
	return nil
}

func mustAtoi(s string) int {
	// keep it minimal and dependency-free
	var n int
	_, _ = fmt.Sscanf(s, "%d", &n)
	return n
}

func indexOf(slice []string, target string) int {
	for i, v := range slice {
		if strings.TrimSpace(v) == target {
			return i
		}
	}
	return -1
}
