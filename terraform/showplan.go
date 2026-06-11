package terraform

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"sync"
)

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiRe.ReplaceAllString(s, "")
}

func ShowPlanText(workdir, planFilePath string) (string, error) {
	cmd := exec.Command("terraform", "show", planFilePath)
	cmd.Dir = workdir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("terraform show failed: %w\nstderr: %s", err, stderr.String())
	}
	return stdout.String(), nil
}

func ExtractResourceSection(showText, address string) string {
	if address == "" {
		return "(no detail available for this resource)\n"
	}
	lines := strings.Split(showText, "\n")
	startIdx := -1
	for i, raw := range lines {
		clean := strings.TrimSpace(stripANSI(raw))
		if !strings.HasPrefix(clean, "# ") {
			continue
		}
		if !strings.Contains(clean, address) {
			continue
		}
		if strings.Contains(clean, " will be ") || strings.Contains(clean, " must be ") {
			idx := strings.Index(clean, address)
			willIdx := strings.Index(clean, " will be ")
			mustIdx := strings.Index(clean, " must be ")
			markerIdx := willIdx
			if markerIdx == -1 || (mustIdx != -1 && mustIdx < markerIdx) {
				markerIdx = mustIdx
			}
			if idx >= 0 && idx < markerIdx {
				startIdx = i
				break
			}
		}
	}
	if startIdx < 0 {
		return "(no detail available for this resource)\n"
	}
	endIdx := len(lines)
	for j := startIdx + 1; j < len(lines); j++ {
		clean := strings.TrimSpace(stripANSI(lines[j]))
		if strings.HasPrefix(clean, "# ") &&
			(strings.Contains(clean, " will be ") || strings.Contains(clean, " must be ")) {
			endIdx = j
			break
		}
		if strings.HasPrefix(clean, "Plan:") {
			endIdx = j
			break
		}
	}
	return strings.Join(lines[startIdx:endIdx], "\n")
}

var planSummaryRe = regexp.MustCompile(`Plan:\s+(\d+)\s+to add,\s+(\d+)\s+to change,\s+(\d+)\s+to destroy`)

func ExtractPlanSummary(showText string) (creates, changes, destroys int) {
	clean := stripANSI(showText)
	m := planSummaryRe.FindStringSubmatch(clean)
	if m == nil {
		return 0, 0, 0
	}
	fmt.Sscanf(m[1], "%d", &creates)
	fmt.Sscanf(m[2], "%d", &changes)
	fmt.Sscanf(m[3], "%d", &destroys)
	return
}

func ExtractOutputChangesSection(showText string) string {
	lines := strings.Split(showText, "\n")
	startIdx := -1
	for i, raw := range lines {
		clean := strings.TrimSpace(stripANSI(raw))
		if strings.HasPrefix(clean, "Changes to Outputs:") {
			startIdx = i
			break
		}
	}
	if startIdx < 0 {
		return ""
	}
	endIdx := len(lines)
	for j := startIdx + 1; j < len(lines); j++ {
		clean := strings.TrimSpace(stripANSI(lines[j]))
		if strings.HasPrefix(clean, "─") || strings.HasPrefix(clean, "---") {
			endIdx = j
			break
		}
	}
	return strings.Join(lines[startIdx:endIdx], "\n")
}

func ExtractWarningsSection(showText string) string {
	lines := strings.Split(showText, "\n")
	var out []string
	i := 0
	for i < len(lines) {
		clean := strings.TrimSpace(stripANSI(lines[i]))
		if strings.HasPrefix(clean, "╷") {
			isWarn := false
			for k := i + 1; k < len(lines) && k < i+5; k++ {
				c := strings.TrimSpace(stripANSI(lines[k]))
				if strings.HasPrefix(c, "│ Warning:") {
					isWarn = true
					break
				}
				if strings.HasPrefix(c, "│") {
					continue
				}
				break
			}
			if isWarn {
				j := i
				for j < len(lines) {
					c := strings.TrimSpace(stripANSI(lines[j]))
					out = append(out, lines[j])
					if strings.HasPrefix(c, "╵") {
						i = j + 1
						break
					}
					j++
				}
				if j >= len(lines) {
					i = len(lines)
				}
				out = append(out, "")
				continue
			}
		}
		if strings.HasPrefix(clean, "Warning:") {
			out = append(out, lines[i])
			j := i + 1
			for j < len(lines) {
				c := stripANSI(lines[j])
				if c == "" || (len(c) > 0 && (c[0] == ' ' || c[0] == '\t')) {
					out = append(out, lines[j])
					j++
					continue
				}
				break
			}
			i = j
			out = append(out, "")
			continue
		}
		i++
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

func ExtractNotesSection(showText string) string {
	lines := strings.Split(showText, "\n")
	var out []string
	i := 0
	for i < len(lines) {
		clean := strings.TrimSpace(stripANSI(lines[i]))
		if strings.HasPrefix(clean, "╷") {
			isNote := false
			for k := i + 1; k < len(lines) && k < i+5; k++ {
				c := strings.TrimSpace(stripANSI(lines[k]))
				if strings.HasPrefix(c, "│ Note:") {
					isNote = true
					break
				}
				if strings.HasPrefix(c, "│") {
					continue
				}
				break
			}
			if isNote {
				j := i
				for j < len(lines) {
					c := strings.TrimSpace(stripANSI(lines[j]))
					out = append(out, lines[j])
					if strings.HasPrefix(c, "╵") {
						i = j + 1
						break
					}
					j++
				}
				if j >= len(lines) {
					i = len(lines)
				}
				out = append(out, "")
				continue
			}
		}
		if strings.HasPrefix(clean, "Note:") {
			out = append(out, lines[i])
			j := i + 1
			for j < len(lines) {
				c := stripANSI(lines[j])
				if c == "" || (len(c) > 0 && (c[0] == ' ' || c[0] == '\t')) {
					out = append(out, lines[j])
					j++
					continue
				}
				break
			}
			i = j
			out = append(out, "")
			continue
		}
		i++
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

func CountWarnings(s string) int {
	return strings.Count(stripANSI(s), "Warning:")
}

func CountNotes(s string) int {
	c := 0
	for _, line := range strings.Split(stripANSI(s), "\n") {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "Note:") || strings.HasPrefix(t, "│ Note:") {
			c++
		}
	}
	return c
}

var outputTopLevelRe = regexp.MustCompile(`^\s{2,4}[+~\-]\s+\w`)

func CountOutputChanges(outputSection string) int {
	if outputSection == "" {
		return 0
	}
	c := 0
	for _, line := range strings.Split(stripANSI(outputSection), "\n") {
		if outputTopLevelRe.MatchString(line) {
			c++
		}
	}
	return c
}

func DominantOutputSymbol(outputSection string) string {
	if outputSection == "" {
		return "+"
	}
	plus, tilde, minus := 0, 0, 0
	for _, line := range strings.Split(stripANSI(outputSection), "\n") {
		if !outputTopLevelRe.MatchString(line) {
			continue
		}
		trim := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(trim, "+") {
			plus++
		} else if strings.HasPrefix(trim, "~") {
			tilde++
		} else if strings.HasPrefix(trim, "-") {
			minus++
		}
	}
	if plus >= tilde && plus >= minus {
		return "+"
	}
	if tilde >= minus {
		return "~"
	}
	return "-"
}

func ActionLabelFor(rc *ResourceChange) string {
	if rc == nil {
		return ""
	}
	actions := rc.Change.Actions
	switch actions.ActionType() {
	case ActionCreate:
		return "will be created"
	case ActionUpdate:
		return "will be updated in-place"
	case ActionDelete:
		return "will be destroyed"
	case ActionReplace:
		return "must be replaced"
	case ActionNoOp:
		for _, a := range actions {
			if a == "read" {
				return "will be read during apply"
			}
		}
		return ""
	}
	return ""
}

type SectionCache struct {
	mu sync.Mutex
	m  map[string]string
}

func NewSectionCache() *SectionCache { return &SectionCache{m: make(map[string]string)} }

func (c *SectionCache) GetOrLoad(workdir, planFilePath string) (string, error) {
	c.mu.Lock()
	if v, ok := c.m[planFilePath]; ok {
		c.mu.Unlock()
		return v, nil
	}
	c.mu.Unlock()
	text, err := ShowPlanText(workdir, planFilePath)
	if err != nil {
		return "", err
	}
	c.mu.Lock()
	c.m[planFilePath] = text
	c.mu.Unlock()
	return text, nil
}

func (c *SectionCache) Set(key, value string) {
	c.mu.Lock()
	c.m[key] = value
	c.mu.Unlock()
}
