package terraform

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ExitError struct {
	ExitCode   int
	Subcommand string
}

func (e *ExitError) Error() string {
	return fmt.Sprintf("terraform %s exited with code %d", e.Subcommand, e.ExitCode)
}

func IsExitError(err error) (*ExitError, bool) {
	var ee *ExitError
	if errors.As(err, &ee) {
		return ee, true
	}
	return nil, false
}

type RunnerOptions struct {
	WorkDir string
	Binary  string
}

type Runner struct {
	opts   RunnerOptions
	binary string
}

func NewRunner(opts RunnerOptions) (*Runner, error) {
	if opts.WorkDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("could not determine working directory: %w", err)
		}
		opts.WorkDir = wd
	}
	bin := opts.Binary
	if bin == "" {
		bin = "terraform"
	}
	return &Runner{opts: opts, binary: bin}, nil
}

func (r *Runner) Passthrough(subcommand string, stdout io.Writer, extraArgs []string) error {
	args := append([]string{subcommand}, extraArgs...)
	rewriter := NewCommandRewriter(stdout)
	cmd := r.newCmdWithArgs(rewriter, args...)
	err := RunWithSignalHandling(cmd)
	if fl, ok := rewriter.(interface{ Flush() error }); ok {
		_ = fl.Flush()
	}
	if err == nil {
		return nil
	}
	var ex *exec.ExitError
	if errors.As(err, &ex) {
		return &ExitError{ExitCode: ex.ExitCode(), Subcommand: subcommand}
	}
	return fmt.Errorf("could not execute terraform %s: %w", subcommand, err)
}

func (r *Runner) Plan(stdout io.Writer, extraArgs []string) (*Plan, []byte, error) {
	tmp, err := os.CreateTemp("", "tfplan-pretty-output-*.tfplan")
	if err != nil {
		return nil, nil, fmt.Errorf("could not create temp plan file: %w", err)
	}
	tmpPath := tmp.Name()
	_ = tmp.Close()
	_ = os.Remove(tmpPath)
	defer os.Remove(tmpPath)

	planArgs := append([]string{"plan", "-out=" + tmpPath}, extraArgs...)

	var buf bytes.Buffer
    
    rewriter := NewCommandRewriter(&buf)
    planCmd := r.newCmdWithArgs(rewriter, planArgs...)
    
    planErr := planCmd.Run()
    
    if flusher, ok := rewriter.(interface{ Flush() error }); ok {
        _ = flusher.Flush()
    }

    filtered := filterTerraformOutput(buf.String())

    fmt.Fprint(stdout, filtered)
    
	if fl, ok := rewriter.(interface{ Flush() error }); ok {
		_ = fl.Flush()
	}
	if planErr != nil {
		var ex *exec.ExitError
		if errors.As(planErr, &ex) {
			return nil, nil, &ExitError{ExitCode: ex.ExitCode(), Subcommand: "plan"}
		}
		return nil, nil, fmt.Errorf("could not execute terraform plan: %w", planErr)
	}
	if _, statErr := os.Stat(tmpPath); statErr != nil {
		return nil, nil, fmt.Errorf("plan succeeded but no plan file was produced at %s", tmpPath)
	}

	var jsonBuf bytes.Buffer
	showCmd := exec.Command(r.binary, "show", "-json", tmpPath)
	showCmd.Dir = r.opts.WorkDir
	showCmd.Env = os.Environ()
	showCmd.Stdin = os.Stdin
	showCmd.Stdout = &jsonBuf
	showCmd.Stderr = stdout
	if err := showCmd.Run(); err != nil {
		return nil, nil, fmt.Errorf("could not capture plan JSON via 'terraform show -json %s': %w", filepath.Base(tmpPath), err)
	}
	jsonBytes := jsonBuf.Bytes()
	plan, err := ParsePlanJSON(jsonBytes)
	if err != nil {
		return nil, nil, err
	}
	return plan, jsonBytes, nil
}

func (r *Runner) newCmdWithArgs(stdout io.Writer, args ...string) *exec.Cmd {
	cmd := exec.Command(r.binary, args...)
	cmd.Dir = r.opts.WorkDir
	cmd.Env = os.Environ()
	cmd.Stdin = os.Stdin
	cmd.Stdout = stdout
	cmd.Stderr = stdout
	return cmd
}

func filterTerraformOutput(out string) string {
    lines := strings.Split(out, "\n")

    var filtered []string

    skip := false

    for _, line := range lines {

        if strings.Contains(line, "Saved the plan to:") {
            skip = true
            continue
        }

        if skip && strings.TrimSpace(line) == "" {
            skip = false
            continue
        }

        if strings.Contains(line, "To perform exactly these actions, run") {
            skip = true
            continue
        }

        if !skip {
            filtered = append(filtered, line)
        }
    }

    return strings.Join(filtered, "\n")
}
