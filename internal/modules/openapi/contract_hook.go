package openapi

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bspippi1337/restless/internal/modules/openapi/ai"
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
	path string,
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
	// 12.5: accept concrete paths by mapping to template if possible.
	if doc.Paths != nil && doc.Paths.Find(path) == nil {
		if tpl, ok := gruntime.MatchPathTemplate(doc, path); ok {
			path = tpl
		}
	}

	v := gruntime.NewValidator(doc)
	findings, err := v.ValidateResponse(ctx, method, path, status, contentType, body)
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
	// 13: not-quite-steady-state snapshot (strictly observational).
	_ = ai.UpdateFromGuard(baseURL, specRef, findings, res.CDI)

	fmt.Printf("OpenAPI contract drift detected (CDI %.3f)\n", res.CDI)
	fmt.Print(greport.PrintHuman(res))
}
