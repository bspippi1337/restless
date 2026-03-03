package openapi

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bspippi1337/restless/internal/modules/openapi/cache"
	"github.com/bspippi1337/restless/internal/modules/openapi/guard/model"
	greport "github.com/bspippi1337/restless/internal/modules/openapi/guard/report"
	gruntime "github.com/bspippi1337/restless/internal/modules/openapi/guard/runtime"
)

// MaybeValidateResponse validates JSON responses against cached OpenAPI contracts.
// Safe no-op if no spec is discoverable.
func MaybeValidateResponse(
	ctx context.Context,
	baseURL string,
	method string,
	pathTemplate string,
	status int,
	contentType string,
	body []byte,
) {
	if ctx == nil {
		ctx = context.Background()
	}

	ct := strings.ToLower(contentType)

	if !strings.Contains(ct, "json") {
		return
	}
	if len(body) == 0 {
		return
	}

	doc, specRef, ok := cache.Get(ctx, baseURL)
	if !ok || doc == nil {
		return
	}

	v := gruntime.NewValidator(doc)
	findings, err := v.ValidateResponse(ctx, method, pathTemplate, status, contentType, body)
	if err != nil || len(findings) == 0 {
		return
	}

	res := model.GuardResult{
		TargetBaseURL: baseURL,
		SpecRef:       specRef,
		StartedAt:     time.Now(),
		FinishedAt:    time.Now(),
		Findings:      findings,
	}

	res.CDI = gruntime.ComputeCDI(findings, gruntime.DefaultWeights())

	fmt.Printf("OpenAPI contract drift detected (CDI %.3f)\n", res.CDI)
	fmt.Print(greport.PrintHuman(res))
}
