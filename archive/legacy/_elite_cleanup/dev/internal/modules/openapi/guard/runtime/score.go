package runtime

import "github.com/bspippi1337/restless/internal/modules/openapi/guard/model"

type ScoreWeights struct {
	ExtraField      float64
	MissingField    float64
	TypeMismatch    float64
	EnumViolation   float64
	SchemaViolation float64
}

func DefaultWeights() ScoreWeights {
	return ScoreWeights{
		ExtraField:      0.5,
		MissingField:    5.0,
		TypeMismatch:    4.0,
		EnumViolation:   2.0,
		SchemaViolation: 3.0,
	}
}

func ComputeCDI(findings []model.Finding, w ScoreWeights) float64 {
	if len(findings) == 0 {
		return 0
	}
	var sum float64
	for _, f := range findings {
		switch f.Kind {
		case model.KindExtraField:
			sum += w.ExtraField
		case model.KindMissingField:
			sum += w.MissingField
		case model.KindTypeMismatch:
			sum += w.TypeMismatch
		case model.KindEnumViolation:
			sum += w.EnumViolation
		default:
			sum += w.SchemaViolation
		}
	}
	den := float64(len(findings) + 4)
	return sum / den
}

func FailThreshold(findings []model.Finding, min model.FindingSeverity) bool {
	rank := func(s model.FindingSeverity) int {
		switch s {
		case model.SevInfo:
			return 0
		case model.SevLow:
			return 1
		case model.SevMedium:
			return 2
		case model.SevHigh:
			return 3
		case model.SevCritical:
			return 4
		default:
			return 2
		}
	}
	minR := rank(min)
	for _, f := range findings {
		if rank(f.Severity) >= minR {
			return true
		}
	}
	return false
}
