package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	DefaultPlansPerLineage = 3
	DefaultMaxLineages     = 3
	DefaultVerboseDepth    = "compact"
)

type Config struct {
	PlansPerLineage int    `yaml:"plansPerLineage"`
	MaxLineages     int    `yaml:"maxLineages"`
	VerboseDepth    string `yaml:"verboseDepth"`
}

func Defaults() Config {
	return Config{
		PlansPerLineage: DefaultPlansPerLineage,
		MaxLineages:     DefaultMaxLineages,
		VerboseDepth:    DefaultVerboseDepth,
	}
}

func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}
	return filepath.Join(home, ".tfplan-pretty-output", "config.yaml"), nil
}

func Load() Config {
	defaults := Defaults()

	path, err := ConfigPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not determine config path: %v\n", err)
		return defaults
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaults
		}
		fmt.Fprintf(os.Stderr, "warning: could not read %s: %v (using defaults)\n", path, err)
		return defaults
	}

	var raw Config
	if err := yaml.Unmarshal(data, &raw); err != nil {
		fmt.Fprintf(os.Stderr, "warning: malformed %s: %v (using defaults)\n", path, err)
		return defaults
	}

	out := defaults
	if raw.PlansPerLineage != 0 {
		if raw.PlansPerLineage < 1 || raw.PlansPerLineage > 100 {
			fmt.Fprintf(os.Stderr,
				"warning: %s: plansPerLineage must be between 1 and 100 (got %d; using default %d)\n",
				path, raw.PlansPerLineage, defaults.PlansPerLineage)
		} else {
			out.PlansPerLineage = raw.PlansPerLineage
		}
	}
	if raw.MaxLineages != 0 {
		if raw.MaxLineages < 1 || raw.MaxLineages > 100 {
			fmt.Fprintf(os.Stderr,
				"warning: %s: maxLineages must be between 1 and 100 (got %d; using default %d)\n",
				path, raw.MaxLineages, defaults.MaxLineages)
		} else {
			out.MaxLineages = raw.MaxLineages
		}
	}
	if raw.VerboseDepth != "" {
		if raw.VerboseDepth != "compact" && raw.VerboseDepth != "full" {
			fmt.Fprintf(os.Stderr,
				"warning: %s: verboseDepth must be 'compact' or 'full' (got %q; using default %q)\n",
				path, raw.VerboseDepth, defaults.VerboseDepth)
		} else {
			out.VerboseDepth = raw.VerboseDepth
		}
	}

	return out
}
