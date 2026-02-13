package discovery

import "github.com/bspippi1337/restless/internal/core/model"

type EvidenceSource = model.EvidenceSource
type Evidence = model.Evidence
type Endpoint = model.Endpoint
type Finding = model.Finding

const (
	SourceOpenAPI = model.SourceOpenAPI
	SourceSitemap = model.SourceSitemap
	SourceRobots  = model.SourceRobots
	SourceHTML    = model.SourceHTML
	SourceProbe   = model.SourceProbe
	SourceFuzzer  = model.SourceFuzzer
	SourceOther   = model.SourceOther
)
