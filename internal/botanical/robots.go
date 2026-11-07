package botanical

import (
    "bufio"
    "fmt"
    "regexp"
    "strconv"
    "strings"
    "time"
)

type robotRule struct {
    path string
    allow bool
}
type robotGroup struct {
    agents []string
    rules []robotRule
    delay time.Duration
}
type robotsPolicy struct {
    groups []robotGroup
}
func parseRobots(data []byte) (robotsPolicy, error) {
    if len(data) > 500<<10 { return robotsPolicy{}, fmt.Errorf("robots.txt exceeds conservative 500 KiB limit") }
    var p robotsPolicy
    current := robotGroup{}
    hasRules := false
    flush := func() {
        if len(current.agents) > 0 { p.groups = append(p.groups, current) }
        current = robotGroup{}
        hasRules = false
    }
    scanner := bufio.NewScanner(strings.NewReader(string(data)))
    scanner.Buffer(make([]byte, 4096), 600<<10)
    for scanner.Scan() {
        line := strings.TrimSpace(strings.SplitN(scanner.Text(), "#", 2)[0])
        if line == "" { continue }
        key, value, ok := strings.Cut(line, ":")
        if !ok { continue }
        key, value = strings.ToLower(strings.TrimSpace(key)), strings.TrimSpace(value)
        switch key {
        case "user-agent":
            if hasRules { flush() }
            if value != "" { current.agents = append(current.agents, strings.ToLower(value)) }
        case "allow", "disallow":
            if len(current.agents) == 0 { continue }
            hasRules = true
            if value != "" { current.rules = append(current.rules, robotRule{value, key == "allow"}) }
        case "crawl-delay":
            if len(current.agents) == 0 { continue }
            hasRules = true
            if seconds, err := strconv.ParseFloat(value, 64); err == nil && seconds > 0 && seconds < 3600 {
                current.delay = time.Duration(seconds*float64(time.Second))
            }
        }
    }
    if err := scanner.Err(); err != nil { return robotsPolicy{}, err }
    flush()
    return p, nil
}
func (p robotsPolicy) selected(agent string) ([]robotRule, time.Duration) {
    agent = strings.ToLower(agent)
    matchLength := -1
    for _, g := range p.groups {
        for _, a := range g.agents {
            n := 0
            if a == "*" { n = 0 } else if strings.Contains(agent, a) { n = len(a) } else { continue }
            if n > matchLength { matchLength = n }
        }
    }
    var rules []robotRule
    var delay time.Duration
    for _, g := range p.groups {
        for _, a := range g.agents {
            n := -1
            if a == "*" { n = 0 } else if strings.Contains(agent, a) { n = len(a) }
            if n == matchLength {
                rules = append(rules, g.rules...)
                if g.delay > delay { delay = g.delay }
                break
            }
        }
    }
    return rules, delay
}
func (p robotsPolicy) allowed(agent, path string) bool {
    if path == "/robots.txt" { return true }
    rules, _ := p.selected(agent)
    best, allowed := -1, true
    for _, rule := range rules {
        pat := rule.path
        anchor := strings.HasSuffix(pat, "$")
        pat = strings.TrimSuffix(pat, "$")
        chunks := strings.Split(pat, "*")
        var expression strings.Builder
        expression.WriteString("^")
        for i, piece := range chunks {
            if i > 0 { expression.WriteString(".*") }
            expression.WriteString(regexp.QuoteMeta(piece))
        }
        if anchor { expression.WriteString("$") }
        r, err := regexp.Compile(expression.String())
        if err != nil || !r.MatchString(path) { continue }
        specificity := len(strings.ReplaceAll(pat, "*", ""))
        if specificity > best || (specificity == best && rule.allow) {
            best, allowed = specificity, rule.allow
        }
    }
    return allowed
}
