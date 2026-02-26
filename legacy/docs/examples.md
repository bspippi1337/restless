Restless – Practical Usage Examples
This document showcases practical, real-world usage scenarios combining Restless with other tools.
1. API Recon + jq + Documentation
Profile and store JSON
restless probe https://api.example.com� \ | jq . \ | tee api-profile.json
Export Markdown documentation
restless export --format=markdown probe https://api.example.com� API_REPORT.md
Use case:
API auditing
Reverse engineering
Internal documentation
2. Smart Mode + curl Generator
Analyze endpoint
restless smart https://api.example.com�
Generate runnable curl
restless export --format=curl post https://api.example.com� \ -d {name:pippi} test.sh
bash test.sh
Use case:
Rapid prototyping
Sharing reproducible API calls
3. Surface Mapping + grep
Find POST-enabled endpoints
restless probe https://api.example.com� | grep POST
Scan multiple endpoints
cat endpoints.txt | xargs -I{} restless probe {} \ | grep application/json
Use case:
Surface discovery
Capability filtering
4. Simulate + HAR Export
Build request interactively
restless simulate https://api.example.com�
Export to HAR
restless export --format=har > request.har
Open HAR in:
Chrome DevTools
Wireshark
Burp Suite
Use case:
Traffic inspection
Security review
5. CI Change Detection
Capture profile
restless probe https://api.example.com� > current.json
Compare against baseline
diff previous.json current.json
Example GitHub Actions snippet
run: restless probe https://api.example.com� > profile.json
run: diff expected.json profile.json
Use case:
Detect breaking API changes
Continuous validation
6. Smart + Fuzz (if enabled)
restless smart https://api.example.com� --fuzz | grep 500
Use case:
Stability testing
Error surface discovery
7. Generate Static Documentation
restless probe https://api.example.com� \ | restless export --format=markdown docs/api.md
Then:
mkdocs build
Use case:
Auto-generated API documentation
Static site publishing
8. Interactive CLI Selection (fzf)
restless probe https://api.example.com� \ | jq -r .methods[] \ | fzf
Use case:
Interactive method selection
CLI workflows
9. Logical Workflow Model
API ↓ Probe ↓ Profile ↓ Smart Decision ↓ Simulate ↓ Export ↓ Automate / Document
Restless is designed as an API workflow engine.
10. Philosophy
Restless excels when:
The API surface is unknown
You need structured discovery
You want reproducible exports
You integrate CLI tooling
You automate analysis
It is not merely a request sender. It is an API exploration layer.
