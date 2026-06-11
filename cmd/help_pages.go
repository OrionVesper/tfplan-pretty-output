package cmd

const helpPlan = `USAGE:
  tfplan-pretty-output plan <name> [terraform-flags...] [-- view | auto-clean]

ARGUMENTS:
  <name>          Lineage name (mandatory). Lowercase letters, digits, and
                  hyphens. 2-30 characters.

OPTIONS (after --):
  view            Open the interactive viewer after capturing
  auto-clean      Run plan without saving (CI/CD / ephemeral mode)

NOTES:
  - The -out flag is not allowed; tfplan-pretty-output manages plan files.
  - Each lineage keeps the 3 most recent plans (older ones auto-pruned).
  - A symlink ./<name> in your workdir points to the latest plan.

EXAMPLES:
  tfplan-pretty-output plan prod
  tfplan-pretty-output plan staging -refresh-only
  tfplan-pretty-output plan prod -- view
  tfplan-pretty-output plan prod -- auto-clean

──────── terraform plan flags ────────
`

const helpApply = `USAGE:
  tfplan-pretty-output apply <name>

ARGUMENTS:
  <name>          Lineage name to apply the latest plan from

NOTES:
  - Only applies the LATEST plan in the named lineage.
  - Applying a specific older plan (e.g. 'apply prod-2') is not supported -
    old plans can drift from current code and corrupt state.

EXAMPLES:
  tfplan-pretty-output apply prod
  tfplan-pretty-output apply staging

──────── terraform apply flags ────────
`

const helpView = `USAGE:
  tfplan-pretty-output view <name | name-N | list>

ARGUMENTS:
  <name>          Open the latest plan in this lineage (interactive TUI)
  <name>-<N>      Open a specific plan in the lineage
  list            Print all saved plans grouped by lineage

TUI KEY BINDINGS:
  Navigation:
    Up / Down       Move between resources
    g / G           Jump to top / bottom
    Enter           Expand / collapse current row
    Shift + Up/Down Scroll inside an expansion
    Esc             Close expansion / clear search / quit
    q / Ctrl+C      Quit

  Modes:
    v               Switch to verbose mode (all expanded)
    Esc             Return to list view (from verbose / help)

  Search:
    /<text>         Search - switches to verbose with matches highlighted
    /<number>       Jump to row N and auto-expand (list view)
    n               Next match (verbose mode)
    N               Previous match (verbose mode)

  Help:
    ?               Show key bindings in the TUI

EXAMPLES:
  tfplan-pretty-output view prod
  tfplan-pretty-output view prod-98
  tfplan-pretty-output view list
`

const helpDiff = `USAGE:
  tfplan-pretty-output diff <plan-1> <plan-2>

ARGUMENTS:
  <plan-1>        First plan reference (e.g. prod, prod-97)
  <plan-2>        Second plan reference (e.g. prod-98)

DESCRIPTION:
  Compares two saved plans and prints a categorized summary:
  - Added resources (in plan-2 but not plan-1)
  - Removed resources (in plan-1 but not plan-2)
  - Changed actions (e.g. update -> replace)
  - Unchanged count

EXAMPLES:
  tfplan-pretty-output diff prod-97 prod-98
  tfplan-pretty-output diff staging prod
`

const helpShow = `USAGE:
  tfplan-pretty-output show <name | name-N>

ARGUMENTS:
  <name>          Show the latest plan in this lineage (full text)
  <name>-<N>      Show a specific plan

DESCRIPTION:
  Prints the full text output of 'terraform show' for the given plan.
  Pipe-friendly - works great with less, grep, etc.

EXAMPLES:
  tfplan-pretty-output show prod
  tfplan-pretty-output show prod-98 | less
  tfplan-pretty-output show prod | grep "destroy"
`

const helpRemove = `USAGE:
  tfplan-pretty-output remove <name | name-N | --all> [-y]

ARGUMENTS:
  <name>          Delete the entire lineage (all plans + workdir symlink)
  <name>-<N>      Delete a single specific plan
  --all           Delete EVERY lineage in this project

FLAGS:
  -y, --yes       Skip confirmation prompts (for scripts and automation)

EXAMPLES:
  tfplan-pretty-output remove prod-2
  tfplan-pretty-output remove prod
  tfplan-pretty-output remove --all
  tfplan-pretty-output remove prod -y
`

const helpDoctor = `USAGE:
  tfplan-pretty-output doctor

DESCRIPTION:
  Runs a series of health checks against your tfplan-pretty-output setup.
  Reports problems with actionable suggestions.

CHECKS:
  - Terraform binary exists on PATH
  - Terraform version
  - Working directory writable
  - Storage root (~/.tfplan-pretty-output/) writable
  - Project ID file exists or will be created
  - Lineages found in this project
  - Broken lineages (e.g. .tfplan without matching .json)

EXAMPLES:
  tfplan-pretty-output doctor
`

const helpHelp = `USAGE:
  tfplan-pretty-output help [command]

ARGUMENTS:
  (no argument)   Show the overview of all available commands
  <command>       Show detailed help for the named command

EXAMPLES:
  tfplan-pretty-output help
  tfplan-pretty-output help plan
  tfplan-pretty-output help view
`