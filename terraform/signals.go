package terraform

import (
    "context"
    "fmt"
    "os"
    "os/exec"
    "os/signal"
    "syscall"
    "time"
)

const GracefulShutdownTimeout = 10 * time.Second

func RunWithSignalHandling(c *exec.Cmd) error {
    c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

    if err := c.Start(); err != nil {
        return err
    }

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    defer signal.Stop(sigCh)

    var interrupted bool
    go func() {
        select {
        case sig := <-sigCh:
            interrupted = true
            _ = syscall.Kill(-c.Process.Pid, sig.(syscall.Signal))

            timer := time.NewTimer(GracefulShutdownTimeout)
            select {
            case <-ctx.Done():
                timer.Stop()
            case <-timer.C:
                _ = syscall.Kill(-c.Process.Pid, syscall.SIGKILL)
            }
        case <-ctx.Done():
        }
    }()

    err := c.Wait()
    cancel()

    if interrupted {
        return &InterruptedError{cause: err}
    }
    return err
}

type InterruptedError struct {
    cause error
}

func (e *InterruptedError) Error() string {
    return "interrupted by user"
}

func (e *InterruptedError) Unwrap() error {
    return e.cause
}

func IsInterrupted(err error) bool {
    if err == nil {
        return false
    }
    _, ok := err.(*InterruptedError)
    return ok
}

func PrintInterruptMessage(kind string) {
    fmt.Println()
    fmt.Println("Gracefully stopping, might take a few seconds...")
}