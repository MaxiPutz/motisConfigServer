package oscmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

// UnpackDownloadedURIWithCallback extracts the file name from the given URI,
// builds a local path (assuming the file is in outDir), then unpacks that file.
// It uses the callback f to output debug messages.
func UnpackDownloadedURIWithCallback(uri, outDir string, f func(data string)) error {
	// Extract filename from the URI.
	segments := strings.Split(uri, "/")
	if len(segments) == 0 {
		return fmt.Errorf("invalid uri: %s", uri)
	}
	fileName := segments[len(segments)-1]
	localPath := fmt.Sprintf("%s/%s", outDir, fileName)

	// Determine the appropriate command arguments based on file extension.
	var cmd *exec.Cmd
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if strings.HasSuffix(localPath, ".tar.bz2") {
		cmd = exec.CommandContext(ctx, "tar", "xvjf", localPath, "-C", outDir)
	} else if strings.HasSuffix(localPath, ".tar.gz") {
		cmd = exec.CommandContext(ctx, "tar", "xvzf", localPath, "-C", outDir)
	} else if strings.HasSuffix(localPath, ".zip") {
		cmd = exec.CommandContext(ctx, "unzip", "-d", outDir, localPath)
	} else {
		return fmt.Errorf("unsupported file extension for %s", localPath)
	}

	// Run the command in its own process group.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Get a pipe for stdout.
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %v", err)
	}
	cmd.Stderr = os.Stderr

	// Set up signal handling: only handle the first SIGINT.
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		sig := <-signalChan
		// Stop further notifications.
		signal.Stop(signalChan)
		if cmd.Process != nil {
			f("kill")
			syscall.Kill(-cmd.Process.Pid, sig.(syscall.Signal))
			cancel()
		}
	}()

	f(fmt.Sprintf("Unpacking file: %s", localPath))
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
