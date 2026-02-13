bash -lc '
set -e
cd ~/restless

mkdir -p scripts

cat > scripts/doctor.sh <<EOF
#!/usr/bin/env sh
set -eu
cd "\$(dirname "\$0")/.."

mkdir -p internal/core/model internal/core/discovery

cat > internal/core/model/types.go <<EOT
package model
import "time"
type EvidenceSource string
const (
  SourceOpenAPI EvidenceSource = "openapi"
  SourceSitemap EvidenceSource = "sitemap"
  SourceRobots  EvidenceSource = "robots"
  SourceHTML    EvidenceSource = "html"
  SourceProbe   EvidenceSource = "probe"
  SourceFuzzer  EvidenceSource = "fuzzer"
  SourceOther   EvidenceSource = "other"
)
type Evidence struct {
  Source EvidenceSource
  URL    string
  Note   string
  When   time.Time
  Score  float64
}
type Endpoint struct {
  Method string
  Path string
  FullURL string
  Evidences []Evidence
}
type Finding struct {
  BaseURL string
  Hosts []string
  DocURLs []string
  Endpoints []Endpoint
  Notes []string
  Confidence float64
}
EOT

cat > internal/core/discovery/types.go <<EOT
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
EOT

sed -i "s|github.com/bspippi1337/restless/internal/core/discovery|github.com/bspippi1337/restless/internal/core/model|g" internal/core/docparse/endpoints.go 2>/dev/null || true
sed -i "s/discovery\\.Endpoint/model.Endpoint/g" internal/core/docparse/endpoints.go 2>/dev/null || true
find . -name "*.go" -
