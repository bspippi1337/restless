# Restless – CLI-først API Discovery Tool

**Versjon:** v0.3.0 (februar 2026)  
**Repo:** https://github.com/bspippi1337/restless  
**Lisens:** MIT

## Introduksjon

Restless er et CLI-verktøy (med TUI) som automatisk oppdager og kartlegger REST API-endepunkter ut fra kun et domenenavn.  
Målet er å gjøre det lett å finne API-dokumentasjon, OpenAPI-spesifikasjoner, sitemaps og reelle endepunkter – uten å måtte lese alt manuelt.

**Hovedprinsipper:**
- Starter med bare et domenenavn (f.eks. `api.bankid.no`)
- Bruker dokumenterte kilder som frø (seeds)
- Sikker fuzzing basert på frø – ingen blind brute-force
- Verifiserer kun med GET/HEAD/OPTIONS som standard
- TUI for interaktiv bruk + JSON for scripting

## Installasjon

### Pre-built binær (anbefalt)

```bash
# Eksempel Linux amd64
curl -L https://github.com/bspippi1337/restless/releases/download/v0.3.0/restless_linux_amd64 -o restless
chmod +x restless
sudo mv restless /usr/local/bin/
restless --version
Fra kildekode
Bashgit clone https://github.com/bspippi1337/restless.git
cd restless
go mod tidy
make build
./bin/restless --version
Hurtigstart – de vanligste kommandoene
Bash# Start interaktiv TUI
restless

# Kjør discovery på et domene
restless discover bankid.no

# Med JSON-output (perfekt for scripting)
restless discover openai.com --json

# Mer tid og sider
restless discover example.com --seconds 60 --pages 20 --json > api-endpoints.json

# Kjør diagnostics & opprydding
restless doctor

# Se hjelp
restless help
Kommandoer i detalj
restless discover <domene> [flagg]
Bash# Enkel scanning
restless discover api.stripe.com

# Med begrensninger og JSON
restless discover api.example.com --seconds 30 --pages 8 --json

# Kun verifiserte endepunkter i ren tekst
restless discover bankid.no --json | jq -r '.endpoints[] | select(.verified==true) | .path'
restless doctor
Rydder gamle builds, logger, validerer oppsett og gir statusrapport.
Bashrestless doctor
Vanlige kombinasjoner
Bash# Finn alle endepunkter og lag en pen liste
restless discover openai.com --json | jq -r '.endpoints[] | "\(.method) \(.path) – \(.source)"'

# Kjør på flere domener
for d in stripe.com openai.com twilio.com; do
  restless discover "$d" --json > "endpoints-$d.json"
done
Sikkerhetsregler (viktig!)

Aldri POST/PUT/DELETE som standard
Kun GET/HEAD/OPTIONS brukes til verifisering
401/403 teller som "finnes" (men krever auth)
Tids- og sidebudsjett hindrer overbelastning
Seed-basert fuzzing – ikke blind dictionary attack

TUI-snarveier (når du kjører restless)

?       → åpne hjelp
Ctrl+D  → start discovery
Tab / Shift+Tab → bytt mellom faner
q / Esc → avslutt
