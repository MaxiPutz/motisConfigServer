package oscmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func CeckPython() string {

	python := "python"

	cmd := exec.Command(python)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		python = "python3"
	}

	cmd = exec.Command(python)

	err = cmd.Run()
	if err != nil {
		fmt.Printf("\"no python env found\": %v\n", "no python env found")
		panic(err)
	}

	return python
}

func ExeFetch(python, nationCodeFile string, f func(data string)) error {
	feedsDir := "./feeds"
	fetchDir := "./src/fetch.py"
	if _, err := os.Stat(feedsDir); err != nil {
		feedsDir = "../feeds"
		fetchDir = "../src/fetch.py"
		if _, err := os.Stat(feedsDir); err != nil {
			return fmt.Errorf("no feeds found: %v", err)
		}
	}

	fmt.Println(python, feedsDir, nationCodeFile)
	cmd := exec.Command(python, fetchDir, feedsDir+"/"+nationCodeFile)

	// Optional: set the command to run in its own process group.
	// This allows you to forward the signal to the entire group.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Get a pipe connected to the command's standard output.
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %v", err)
	}
	cmd.Stderr = os.Stderr

	// Set up signal handling: listen for SIGINT (Ctrl+C) and forward it.
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	defer signal.Stop(signalChan)

	go func() {
		sig := <-signalChan
		if cmd.Process != nil {
			f("kill")
			syscall.Kill(-cmd.Process.Pid, sig.(syscall.Signal))

		}
	}()

	// Start the command.
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %v", err)
	}

	// Read stdout data line by line and notify via the callback.
	scanner := bufio.NewScanner(stdoutPipe)
	for scanner.Scan() {
		line := scanner.Text()
		f(line)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading stdout: %v", err)
	}

	// Wait for the command to exit.
	return cmd.Wait()
}
