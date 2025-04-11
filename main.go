package main

import (
	"bufio"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"maxiputz/motisConfigServer/download"
	motisconfigfile "maxiputz/motisConfigServer/motisConfigFile"
	"maxiputz/motisConfigServer/scrapper"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"syscall"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

type Response struct {
	Region   scrapper.Region       `json:"region"`
	Releases []scrapper.Release    `json:"releases"`
	GTFSUrl  []scrapper.Transitous `json:"gtfsUrl"`
	OS       string                `json:"os"`
	Arch     string                `json:"arch"`
}

type SocketChunk struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type SocketChunkString struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

var writeMutex sync.Mutex

//go:embed "ui/dist/*"
var folderPath embed.FS

//go:embed "assets/*"
var assatsPath embed.FS

func main() {

	regions, releases, transitous, err := scrapper.GetAllAssetes()

	downLoadCallback := func(name string, prgress string) {}
	motisImportCallback := func(data string) {}

	hostOS := runtime.GOOS
	if hostOS == "darwin" {
		hostOS = "macos"
	}

	if err != nil {
		panic(err)
	}

	app := fiber.New()
	app.Use(cors.New())

	app.Use("/", filesystem.New(filesystem.Config{
		Root:       http.FS(folderPath),
		PathPrefix: "ui/dist", // Path prefix inside the embedded FS
		Browse:     true,      // Allow directory browsing (optional)
	}))

	// Start the server

	app.Get("/init", func(c *fiber.Ctx) error {

		return c.JSON(Response{
			Region:   regions,
			Releases: releases,
			GTFSUrl:  transitous,
			OS:       hostOS,
			Arch:     runtime.GOARCH,
		})
	})

	app.Get("/ws/", websocket.New(func(c *websocket.Conn) {

		downLoadCallback = func(name, prgress string) {
			writeMutex.Lock()
			defer writeMutex.Unlock()
			c.WriteJSON(SocketChunk{
				Name: name,
				Data: prgress,
			})
		}
		motisImportCallback = func(data string) {
			fmt.Printf("data in ws: %v\n", data)
			c.WriteJSON(SocketChunkString{
				Name: "terminaldata",
				Data: data,
			})
		}

		for {
			if _, _, err = c.ReadMessage(); err != nil {
				log.Println("read:", err)
				break
			}
		}

	}))

	app.Post("/startDownload", func(c *fiber.Ctx) error {
		reqData := download.RequestDownload{}
		if err := c.BodyParser(&reqData); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		fmt.Printf("reqData: %+v\n", reqData)

		// Process the data as needed and then respon
		go func() {
			download.DownloadAll(reqData, func(fileName string, downloaded, total int64) {
				fmt.Printf("fileName: %v\n", fileName)
				fmt.Printf("downloaded: %v\n", downloaded)
				fmt.Printf("total: %v\n", total)
				fmt.Printf("(downloaded / total): %v\n", (float64(downloaded) / float64(total) * 100))

				progress := strconv.FormatFloat((float64(downloaded)/float64(total))*100, 'f', 2, 64)
				downLoadCallback(fileName, progress)
			})

			feeds, _ := findGtfsInOut()
			osmFile, _ := findOsmInOut()
			fmt.Printf("\"config is stared\": %v\n", "config is stared")
			runMotisCondfig(feeds, osmFile)
			fmt.Printf("config is writte you can run on your host pc ./motis import \n")
			fmt.Printf("after the import is run through you can run ./motis serve \n")

			reqestDataJson, err := json.MarshalIndent(reqData, " ", "    ")

			if err != nil {
				fmt.Printf("\"reqData not marshalled\": %v\n", "reqData not marshalled")
				panic(1)
			}

			os.WriteFile("./out/downloadUrls.json", reqestDataJson, 0664)

			os.Exit(0)
			runMotisImportCallback(motisImportCallback)

		}()
		return c.SendString("sending data")
	})

	app.Get("/import", func(c *fiber.Ctx) error {
		DownLoadWget("https://github.com/motis-project/motis/releases/download/v2.0.43/motis-macos-arm64.tar.bz2", func(data string) {
			fmt.Printf("data: %v\n", data)
		})
		runMotisImportCallback(motisImportCallback)
		return c.SendString("import is started")
	})

	fmt.Printf("\"👉 Open http://localhost:3001 in your browser to finish setup!\": %v\n", "👉 Open http://localhost:3001 in your browser to finish setup!")
	app.Listen(":3001")
	fmt.Printf("\"👉 Open http://localhost:3001 in your browser to finish setup!\": %v\n", "👉 Open http://localhost:3001 in your browser to finish setup!")

}

func findGtfsInOut() ([]string, error) {
	entries, err := os.ReadDir("out")
	if err != nil {
		return nil, fmt.Errorf("failed to read out directory: %w", err)
	}

	var results []string
	for _, entry := range entries {
		// Only consider files (skip directories)
		if !entry.IsDir() {
			name := entry.Name()
			matched, err := filepath.Match("*.gtfs.zip", name)
			if err != nil {
				return nil, fmt.Errorf("error matching pattern: %w", err)
			}
			if matched {
				results = append(results, name)
			}
		}
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no GTFS files found")
	}
	return results, nil
}

func findOsmInOut() (string, error) {
	entries, err := os.ReadDir("out")
	if err != nil {
		return "", fmt.Errorf("failed to read out directory: %w", err)
	}

	result := ""
	for _, entry := range entries {
		// Only consider files (skip directories)

		if !entry.IsDir() {
			name := entry.Name()
			matched, err := filepath.Match("*.osm.pbf", name)
			if err != nil {
				return "", fmt.Errorf("error matching pattern: %w", err)
			}
			if matched {
				result = name
				break
			}
		}
	}
	if len(result) == 0 {
		return "", fmt.Errorf("no GTFS files found")
	}
	return result, nil
}
func runMotisCondfig(feeds []string, osmFile string) {

	motisconfigfile.GenerateMotisConfig(osmFile, feeds, "out/")

	return
	args := append([]string{"config", osmFile}, feeds...)
	fmt.Printf("\"motis\": %v\n", "motis")
	fmt.Printf("args: %v\n", args)
	cmd := exec.Command("./motis", args...)
	cmd.Dir = "out" // Set the working directory to "out"
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error running './motis config': %v\nOutput: %s\n", err, output)
		return
	}
	fmt.Printf("Successfully wrote config:\n%s\n", output)
}

func runMotisImport() error {
	cmd := exec.Command("./motis", "import")
	cmd.Dir = "out" // Set the working directory to "out"
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error running './motis import': %v\nOutput: %s\n", err, output)
		return fmt.Errorf("failed to get stdout pipe: %v", err)
	}
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %v", err)
	}
	cmd.Stderr = os.Stderr
	fmt.Printf("Successfully import config:\n%s\n", output)

	return nil
}

func runMotisImportCallback(fn func(data string)) error {
	cmd := exec.Command("./motis", "import")
	cmd.Dir = "out" // Set the working directory to "out"

	fmt.Println("motis callback fun: starting command")

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %v", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %v", err)
	}
	fmt.Println("motis callback fun: command started")

	// Start a goroutine to log stderr
	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			fmt.Printf("stderr: %v\n", scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			fmt.Printf("error reading stderr: %v\n", err)
		}
	}()

	// Read stdout line by line.
	scanner := bufio.NewScanner(stdoutPipe)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Printf("stdout: %v\n", line)
		fn(line)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading stdout: %v", err)
	}

	// Wait for the command to finish.
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("command execution failed: %v", err)
	}

	return nil
}

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
