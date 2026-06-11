package cmd

func hasDanglingDashDash(args []string) bool {
    for i, a := range args {
        if a == "--" {
            return i == len(args)-1
        }
    }
    return false
}