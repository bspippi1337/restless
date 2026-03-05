package app

import "github.com/bspippi1337/restless/internal/council"

var GlobalBlackboard = council.NewBlackboard()

func PublishFinding(engine, kind, target, evidence string, confidence float64) {
	GlobalBlackboard.Publish(council.Finding{
		Engine:     engine,
		Kind:       kind,
		Target:     target,
		Evidence:   evidence,
		Confidence: confidence,
	})
}
