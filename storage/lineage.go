package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/OrionVesper/tfplan-pretty-output/terraform"
)

const (
    MaxPlansPerLineage    = 3
    MaxLineagesPerProject = 3
)


type LineagePlan struct {
	ID        string            
	Lineage   string            
	Number    int               
	Timestamp time.Time         
	Summary   terraform.Summary 
}


type Lineage struct {
	Name        string
	Plans       []LineagePlan 
	LatestStamp time.Time
}

type LineageStore struct {
	projectID string
	root      string
	workdir   string
}

func NewLineageStore(projectID, workdir string) (*LineageStore, error) {
	if projectID == "" {
		return nil, fmt.Errorf("projectID is required")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not determine home directory: %w", err)
	}
	root := filepath.Join(home, ".tfplan-pretty-output", "projects", projectID)
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("could not create %s: %w", root, err)
	}
	return &LineageStore{projectID: projectID, root: root, workdir: workdir}, nil
}

func (s *LineageStore) Root() string { return s.root }

func (s *LineageStore) ListLineages() ([]Lineage, error) {
	entries, err := os.ReadDir(s.root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("could not read %s: %w", s.root, err)
	}

	var lineages []Lineage
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		plans, err := s.readLineagePlans(name)
		if err != nil {
			return nil, err
		}
		if len(plans) == 0 {
			continue
		}
		lineages = append(lineages, Lineage{
			Name:        name,
			Plans:       plans,
			LatestStamp: plans[0].Timestamp,
		})
	}
	return lineages, nil
}

func (s *LineageStore) readLineagePlans(lineageName string) ([]LineagePlan, error) {
	dir := filepath.Join(s.root, lineageName)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("could not read %s: %w", dir, err)
	}

	type pair struct {
		hasTFPlan, hasJSON bool
		mod                time.Time
	}
	by := map[int]*pair{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ln, num, ok := parsePlanFilename(e.Name())
		if !ok || ln != lineageName {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		p, exists := by[num]
		if !exists {
			p = &pair{}
			by[num] = p
		}
		if strings.HasSuffix(e.Name(), ".tfplan") {
			p.hasTFPlan = true
		} else if strings.HasSuffix(e.Name(), ".json") {
			p.hasJSON = true
		}
		if info.ModTime().After(p.mod) {
			p.mod = info.ModTime()
		}
	}

	var plans []LineagePlan
	for num, p := range by {
		if !p.hasTFPlan || !p.hasJSON {
			continue
		}
		plans = append(plans, LineagePlan{
			ID:        fmt.Sprintf("%s-%d", lineageName, num),
			Lineage:   lineageName,
			Number:    num,
			Timestamp: p.mod,
		})
	}
	sort.Slice(plans, func(i, j int) bool {
		return plans[i].Number > plans[j].Number
	})
	return plans, nil
}

func (s *LineageStore) NextNumber(lineageName string) (int, error) {
	plans, err := s.readLineagePlans(lineageName)
	if err != nil {
		return 0, err
	}
	if len(plans) == 0 {
		return 1, nil
	}
	return plans[0].Number + 1, nil
}

func (s *LineageStore) SavePlan(
	lineageName string,
	plan *terraform.Plan,
	jsonBytes []byte,
	planFileSrc string,
) (string, error) {
	lineages, err := s.ListLineages()
	if err != nil {
		return "", err
	}
	exists := false
	for _, l := range lineages {
		if l.Name == lineageName {
			exists = true
			break
		}
	}
	if !exists && len(lineages) >= MaxLineagesPerProject {
		if err := s.evictLRULineage(lineages); err != nil {
			return "", err
		}
	}

	dir := filepath.Join(s.root, lineageName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("could not create %s: %w", dir, err)
	}

	num, err := s.NextNumber(lineageName)
	if err != nil {
		return "", err
	}
	id := fmt.Sprintf("%s-%d", lineageName, num)

	jsonPath := filepath.Join(dir, id+".json")
	planPath := filepath.Join(dir, id+".tfplan")

	if err := os.WriteFile(jsonPath, jsonBytes, 0o644); err != nil {
		return "", fmt.Errorf("could not write %s: %w", jsonPath, err)
	}

	if err := os.Rename(planFileSrc, planPath); err != nil {
		if cerr := copyFile(planFileSrc, planPath); cerr != nil {
			return "", fmt.Errorf("could not persist plan file: %w", cerr)
		}
		_ = os.Remove(planFileSrc)
	}

	if info, err := os.Stat(jsonPath); err != nil || info.Size() == 0 {
		return "", fmt.Errorf("plan JSON not written at %s", jsonPath)
	}
	if info, err := os.Stat(planPath); err != nil || info.Size() == 0 {
		return "", fmt.Errorf("plan file not written at %s", planPath)
	}

	if err := s.PruneLineage(lineageName); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not prune lineage %s: %v\n", lineageName, err)
	}

	if s.workdir != "" {
		if err := s.UpdateSymlink(lineageName, planPath); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not update symlink ./%s: %v\n", lineageName, err)
		}
	}

	return id, nil
}

func (s *LineageStore) PruneLineage(lineageName string) error {
	plans, err := s.readLineagePlans(lineageName)
	if err != nil {
		return err
	}
	if len(plans) <= MaxPlansPerLineage {
		return nil
	}
	dir := filepath.Join(s.root, lineageName)
	for _, p := range plans[MaxPlansPerLineage:] {
		_ = os.Remove(filepath.Join(dir, p.ID+".tfplan"))
		_ = os.Remove(filepath.Join(dir, p.ID+".json"))
	}
	return nil
}

func (s *LineageStore) DeleteLineage(lineageName string) error {
	dir := filepath.Join(s.root, lineageName)
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("could not remove %s: %w", dir, err)
	}
	if s.workdir != "" {
		sym := filepath.Join(s.workdir, lineageName)
		if fi, err := os.Lstat(sym); err == nil {
			if fi.Mode()&os.ModeSymlink != 0 {
				_ = os.Remove(sym)
			}
		}
	}
	return nil
}




func (s *LineageStore) DeleteSpecificPlan(lineageName string, number int) error {
	dir := filepath.Join(s.root, lineageName)
	id := fmt.Sprintf("%s-%d", lineageName, number)
	tfp := filepath.Join(dir, id+".tfplan")
	jsonp := filepath.Join(dir, id+".json")

	if _, err := os.Stat(tfp); err != nil {
		return fmt.Errorf("plan file not found: %s", tfp)
	}

	if err := os.Remove(tfp); err != nil {
		return fmt.Errorf("could not remove %s: %w", tfp, err)
	}
	if err := os.Remove(jsonp); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("could not remove %s: %w", jsonp, err)
	}

	
	plans, err := s.readLineagePlans(lineageName)
	if err != nil {
		return err
	}
	if len(plans) == 0 {
		
		if s.workdir != "" {
			sym := filepath.Join(s.workdir, lineageName)
			if fi, lerr := os.Lstat(sym); lerr == nil {
				if fi.Mode()&os.ModeSymlink != 0 {
					_ = os.Remove(sym)
				}
			}
		}
		return nil
	}
	
	latest := plans[0]
	target := filepath.Join(dir, latest.ID+".tfplan")
	return s.UpdateSymlink(lineageName, target)
}

func (s *LineageStore) evictLRULineage(lineages []Lineage) error {
	if len(lineages) == 0 {
		return nil
	}
	oldest := lineages[0]
	for _, l := range lineages[1:] {
		if l.LatestStamp.Before(oldest.LatestStamp) {
			oldest = l
		}
	}
	return s.DeleteLineage(oldest.Name)
}

func (s *LineageStore) FindPlan(ref string) (*LineagePlan, error) {
	lineageName, number, isSpecific := splitPlanRef(ref)
	plans, err := s.readLineagePlans(lineageName)
	if err != nil {
		return nil, err
	}
	if len(plans) == 0 {
		return nil, fmt.Errorf("lineage %q not found", lineageName)
	}
	if !isSpecific {
		latest := plans[0]
		return &latest, nil
	}
	for _, p := range plans {
		if p.Number == number {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("plan %q not found (kept: top %d of %s)",
		ref, MaxPlansPerLineage, lineageName)
}

func (s *LineageStore) FindPlanFile(ref string) (string, error) {
	p, err := s.FindPlan(ref)
	if err != nil {
		return "", err
	}
	return filepath.Join(s.root, p.Lineage, p.ID+".tfplan"), nil
}

func (s *LineageStore) FindPlanJSON(ref string) (string, error) {
	p, err := s.FindPlan(ref)
	if err != nil {
		return "", err
	}
	return filepath.Join(s.root, p.Lineage, p.ID+".json"), nil
}

func (s *LineageStore) UpdateSymlink(lineageName, targetPath string) error {
	if s.workdir == "" {
		return nil
	}
	sym := filepath.Join(s.workdir, lineageName)
	if fi, err := os.Lstat(sym); err == nil {
		if fi.Mode()&os.ModeSymlink == 0 {
			return fmt.Errorf("%s exists and is not a symlink — refusing to overwrite", sym)
		}
		if err := os.Remove(sym); err != nil {
			return fmt.Errorf("could not remove old symlink %s: %w", sym, err)
		}
	}
	if err := os.Symlink(targetPath, sym); err != nil {
		return fmt.Errorf("could not create symlink %s -> %s: %w", sym, targetPath, err)
	}
	return nil
}

func parsePlanFilename(filename string) (lineage string, number int, ok bool) {
	base := filename
	switch {
	case strings.HasSuffix(base, ".tfplan"):
		base = strings.TrimSuffix(base, ".tfplan")
	case strings.HasSuffix(base, ".json"):
		base = strings.TrimSuffix(base, ".json")
	default:
		return "", 0, false
	}
	idx := strings.LastIndex(base, "-")
	if idx <= 0 || idx == len(base)-1 {
		return "", 0, false
	}
	suffix := base[idx+1:]
	n, err := strconv.Atoi(suffix)
	if err != nil || n < 1 {
		return "", 0, false
	}
	return base[:idx], n, true
}

func splitPlanRef(ref string) (lineage string, number int, isSpecific bool) {
	idx := strings.LastIndex(ref, "-")
	if idx <= 0 || idx == len(ref)-1 {
		return ref, 0, false
	}
	suffix := ref[idx+1:]
	n, err := strconv.Atoi(suffix)
	if err != nil || n < 1 {
		return ref, 0, false
	}
	return ref[:idx], n, true
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
