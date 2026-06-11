package terraform

import (
	"bytes"
	"io"
	"regexp"
	"strings"
)

const rewriteToolName = "tfplan-pretty-output"

var knownSubcommands = []string{
	"init", "validate", "fmt", "plan", "apply", "destroy",
	"show", "output", "refresh", "providers", "version",
	"console", "import", "state", "workspace", "get",
	"force-unlock", "login", "logout", "test", "graph",
	"taint", "untaint",
}

var rewritePattern *regexp.Regexp

func init() {
	alts := strings.Join(knownSubcommands, "|")

	rewritePattern = regexp.MustCompile(
		`\bterraform([ \t]+(?:` + alts + `)\b|[ \t]+\[|[ \t]+-)`,
	)
}

type commandRewriter struct {
	out io.Writer
	buf bytes.Buffer
}

func NewCommandRewriter(w io.Writer) io.Writer { return &commandRewriter{out: w} }

func (r *commandRewriter) Write(p []byte) (int, error) {
	r.buf.Write(p)
	for {
		idx := bytes.IndexByte(r.buf.Bytes(), '\n')
		if idx < 0 {
			break
		}
		line := r.buf.Next(idx + 1)
		rewritten := rewriteLine(line)
		if _, err := r.out.Write(rewritten); err != nil {
			return 0, err
		}
	}
	return len(p), nil
}

func (r *commandRewriter) Flush() error {
	if r.buf.Len() == 0 {
		return nil
	}
	pending := r.buf.Bytes()
	r.buf.Reset()
	_, err := r.out.Write(rewriteLine(pending))
	return err
}

func (r *commandRewriter) Close() error {
	_ = r.Flush()
	if c, ok := r.out.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

func rewriteLine(line []byte) []byte {

	return rewritePattern.ReplaceAll(line, []byte(rewriteToolName+"$1"))
}
