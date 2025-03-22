package oscmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func DownLoadWget(url string, f func(data string)) error {
	// Create a cancellable context.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Pass command arguments as separate strings.
	cmd := exec.CommandContext(ctx, "wget", "-P", "out", url)
	// Run the command in its own process group.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %v", err)
	}
	cmd.Stderr = os.Stderr

	// Set up signal handling: we want to handle SIGINT only once.
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		// Wait for a signal.
		sig := <-signalChan
		// Stop receiving further signals.
		signal.Stop(signalChan)
		if cmd.Process != nil {
			f("kill")
			// Send the signal to the process group (note the negative PID).
			syscall.Kill(-cmd.Process.Pid, sig.(syscall.Signal))
			cancel()
		}
	}()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %v", err)
	}

	scanner := bufio.NewScanner(stdoutPipe)
	for scanner.Scan() {
		line := scanner.Text()
		f(line)
	}

	return cmd.Wait()
}
