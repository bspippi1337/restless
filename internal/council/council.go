package council

import "fmt"

type Council struct {
	Blackboard *Blackboard
}

func NewCouncil(b *Blackboard) *Council {
	return &Council{Blackboard: b}
}

func (c *Council) Convene() {

	findings := c.Blackboard.List()

	byTarget := map[string][]Finding{}

	for _, f := range findings {
		byTarget[f.Target] = append(byTarget[f.Target], f)
	}

	for target, group := range byTarget {

		score := 0.0
		engines := map[string]bool{}

		for _, f := range group {
			score += f.Confidence
			engines[f.Engine] = true
		}

		if len(engines) > 1 && score > 1.2 {
			fmt.Printf("council consensus: %s (score %.2f)\n", target, score)
			status.IncConsensus()
		}
	}
}
