package terraform

import "sort"

type DiffEntry struct {
	Address string
	Before  ActionType
	After   ActionType
}

type PlanDiff struct {
	Added     []ResourceChange
	Removed   []ResourceChange
	Changed   []DiffEntry
	Unchanged int
}

func ComputePlanDiff(plan1, plan2 *Plan) *PlanDiff {
	d := &PlanDiff{}

	m1 := map[string]ResourceChange{}
	m2 := map[string]ResourceChange{}
	if plan1 != nil {
		for _, rc := range plan1.ResourceChanges {
			m1[rc.Address] = rc
		}
	}
	if plan2 != nil {
		for _, rc := range plan2.ResourceChanges {
			m2[rc.Address] = rc
		}
	}

	for addr, rc := range m1 {
		if _, ok := m2[addr]; !ok {
			d.Removed = append(d.Removed, rc)
		}
	}

	for addr, rc := range m2 {
		if _, ok := m1[addr]; !ok {
			d.Added = append(d.Added, rc)
		}
	}

	for addr, rc1 := range m1 {
		rc2, ok := m2[addr]
		if !ok {
			continue
		}
		a1 := rc1.Change.Actions.ActionType()
		a2 := rc2.Change.Actions.ActionType()
		if a1 == a2 {
			d.Unchanged++
		} else {
			d.Changed = append(d.Changed, DiffEntry{Address: addr, Before: a1, After: a2})
		}
	}

	sort.Slice(d.Added, func(i, j int) bool { return d.Added[i].Address < d.Added[j].Address })
	sort.Slice(d.Removed, func(i, j int) bool { return d.Removed[i].Address < d.Removed[j].Address })
	sort.Slice(d.Changed, func(i, j int) bool { return d.Changed[i].Address < d.Changed[j].Address })

	return d
}
