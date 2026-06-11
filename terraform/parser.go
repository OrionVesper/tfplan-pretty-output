package terraform

import (
	"encoding/json"
	"fmt"
)

func ParsePlanJSON(data []byte) (*Plan, error) {
	var p Plan
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("could not parse plan JSON: %w", err)
	}
	if p.FormatVersion == "" {
		return nil, fmt.Errorf("invalid plan: missing format_version (is this really 'terraform show -json' output?)")
	}
	return &p, nil
}

func BuildSummary(plan *Plan) Summary {
	var s Summary
	for _, rc := range plan.ResourceChanges {
		s.Total++
		switch rc.Change.Actions.ActionType() {
		case ActionCreate:
			s.Creates++
		case ActionUpdate:
			s.Updates++
		case ActionDelete:
			s.Deletes++
		case ActionReplace:
			s.Replaces++
		default:
			s.NoOps++
		}
	}
	return s
}
