package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	oscmd "github.com/maxiputz/transitous/os_cmd"
)

type ReqGTTF struct {
	Data []string `json:"data"`
}

type ProcessState struct {
	isMotisReadyToImport bool
	isOsmDownLoad        bool
	isFeedDownload       bool
	Feeds                []string
	OSMFile              string
	OnIsFinished         func(data string)
	OnIsMotisVerbose     func(data string)
}

func NewProcessState() ProcessState {
	return ProcessState{
		isMotisReadyToImport: false,
		isOsmDownLoad:        false,
		isFeedDownload:       false,
		OnIsFinished:         func(data string) { fmt.Printf("data: %v\n", data) },
		OnIsMotisVerbose:     func(data string) { fmt.Printf("data: %v\n", data) },
	}
}

func (ps *ProcessState) importMotis() error {
	if ps.isFeedDownload && ps.isMotisReadyToImport && ps.isOsmDownLoad {
		// cd to out as workin dir
		// call ./motis config
		cmd := exec.Command("./motis", "import")
		cmd.Dir = "out" // Set the working directory to "out"
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Error running './motis import': %v\nOutput: %s\n", err, output)
			return fmt.Errorf("failed to get stdout pipe: %v", err)
		}
		stdoutPipe, err := cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("failed to get stdout pipe: %v", err)
		}
		cmd.Stderr = os.Stderr
		fmt.Printf("Successfully wrote config:\n%s\n", output)

		scanner := bufio.NewScanner(stdoutPipe)

		for scanner.Scan() {
			line := scanner.Text()
			ps.OnIsMotisVerbose("import: " + line)
		}
	}
	return nil
}

func (ps *ProcessState) wirteConfigIfReady() {
	if ps.isFeedDownload && ps.isMotisReadyToImport && ps.isOsmDownLoad {
		// cd to out as workin dir
		// call ./motis config
		ps.Feeds, _ = findGtfsInOut()

		args := append([]string{"config", ps.OSMFile}, ps.Feeds...)
		cmd := exec.Command("./motis", args...)
		cmd.Dir = "out" // Set the working directory to "out"
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Error running './motis config': %v\nOutput: %s\n", err, output)
			return
		}
		fmt.Printf("Successfully wrote config:\n%s\n", output)
		ps.importMotis()
	}
}

func (ps *ProcessState) SetMotisReadyToImport() {
	ps.isMotisReadyToImport = true
	ps.wirteConfigIfReady()
}
func (ps *ProcessState) SetOsmDownLoaded() {
	ps.isOsmDownLoad = true
	ps.wirteConfigIfReady()
}
func (ps *ProcessState) SetFeedDownloaded() {
	ps.isFeedDownload = true
	ps.wirteConfigIfReady()
}

type Configure struct {
	Arch       string   `json:"arch"`
	OS         string   `json:"os"`
	OSM        string   `json:"osm"`
	Feeds      []string `json:"feeds"`
	MotisReady bool     `json:"motisReady"`
	OSReady    bool     `json:"osReady"`
	OSMReady   bool     `json:"osmReady"`
}

func NewConfigure() Configure {

	os := runtime.GOOS

	fmt.Printf("os: %v\n", os)
	if os == "darwin" {
		os = "macos"
	}

	return Configure{
		Arch:       runtime.GOARCH,
		OS:         os,
		MotisReady: false,
		OSReady:    false,
		OSMReady:   false,
	}
}

func main() {

	app := fiber.New()
	app.Use(cors.New())

	processState := NewProcessState()
	config := NewConfigure()

	python := oscmd.CeckPython()

	app.Get("/init", func(c *fiber.Ctx) error {
		return c.JSON(config)
	})

	app.Get("/configure/skip/download", func(c *fiber.Ctx) error {
		// Read query parameters
		feeds := c.Query("feeds")
		osmUrl := c.Query("osmUrl")
		motisUrl := c.Query("motisUrl")

		// Print parameters to the server console
		fmt.Println("Feeds:", feeds)
		fmt.Println("OSM URL:", osmUrl)
		fmt.Println("Motis URL:", motisUrl)

		dirs := strings.Split(feeds, ",")

		processState.OSMFile = strings.Split(osmUrl, "/")[len(strings.Split(osmUrl, "/"))-1]
		processState.Feeds = dirs
		go func() {
			config.Feeds, _ = findGtfsInOut()

			processState.SetFeedDownloaded()
		}()

		go func() {
			config.OSM = osmUrl
			processState.SetOsmDownLoaded()
		}()

		go func() {
			config.Arch = motisUrl
			config.OS = motisUrl
			processState.SetMotisReadyToImport()

		}()

		// Optionally, send them back as JSON
		return c.JSON(fiber.Map{
			"feeds":    feeds,
			"osmUrl":   osmUrl,
			"motisUrl": motisUrl,
		})
	})

	app.Get("/configure", func(c *fiber.Ctx) error {
		// Read query parameters
		feeds := c.Query("feeds")
		osmUrl := c.Query("osmUrl")
		motisUrl := c.Query("motisUrl")

		// Print parameters to the server console
		fmt.Println("Feeds:", feeds)
		fmt.Println("OSM URL:", osmUrl)
		fmt.Println("Motis URL:", motisUrl)

		dirs := strings.Split(feeds, ",")

		processState.OSMFile = strings.Split(osmUrl, "/")[len(strings.Split(osmUrl, "/"))-1]
		processState.Feeds = dirs
		go func() {
			fmt.Printf("dirs: %v\n", dirs)
			c5 := make(chan struct{}, 5)
			isCancel := false

			var wg sync.WaitGroup
			for _, e := range dirs {
				wg.Add(1)
				fmt.Printf("e: %v\n", e)
				c5 <- struct{}{}
				go func(e string) {
					if isCancel {
						return
					}
					oscmd.ExeFetch(python, e, func(data string) {
						if data == "kill" {
							isCancel = true
							time.Sleep(time.Millisecond * 300)
						}
						fmt.Printf("data: %v\n", data)
					})

					<-c5
					wg.Done()
				}(e)
			}

			wg.Wait()
			config.Feeds, _ = findGtfsInOut()

			processState.SetFeedDownloaded()
		}()

		go func() {
			err := oscmd.DownLoadWget(osmUrl, func(data string) {
				str := "osm: " + data
				fmt.Printf("str: %v\n", str)
			})
			if err != nil {
				// Log any error from the download command.
				log.Printf("DownLoadMotis error: %v\n", err)
			}
			config.Arch = motisUrl
			config.OS = motisUrl
			processState.SetOsmDownLoaded()
		}()

		go func() {
			err := oscmd.DownLoadWget(motisUrl, func(data string) {
				str := "motis: " + data
				fmt.Printf("str: %v\n", str)
			})
			if err != nil {
				// Log any error from the download command.
				log.Printf("DownLoadMotis error: %v\n", err)
			}
			err = oscmd.UnpackDownloadedURIWithCallback(motisUrl, "out", func(data string) {
				str := "motis: " + data
				fmt.Printf("str: %v\n", str)
			})
			if err != nil {
				// Log any error from the download command.
				log.Printf("DownLoadMotis error: %v\n", err)
			}
			processState.SetMotisReadyToImport()

		}()

		// Optionally, send them back as JSON
		return c.JSON(fiber.Map{
			"feeds":    feeds,
			"osmUrl":   osmUrl,
			"motisUrl": motisUrl,
		})
	})

	app.Get("/test", func(c *fiber.Ctx) error {

		return c.SendString("dereoida")
	})

	app.Get("/feeds", func(c *fiber.Ctx) error {
		feeds := oscmd.GetFeeds()

		name := []string{}

		for _, feed := range feeds {
			name = append(name, feed.Name())
		}

		return c.JSON(name)
	})

	app.Post("/startDownload", func(c *fiber.Ctx) error {
		fmt.Printf("\"in the post\": %v\n", "in the post")
		var reqData ReqGTTF
		if err := c.BodyParser(&reqData); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid request body")
		}

		dirs := reqData.Data
		c5 := make(chan struct{}, 5)
		isCancel := false
		for _, e := range dirs {
			fmt.Printf("e: %v\n", e)
			c5 <- struct{}{}
			go func(e string) {
				if isCancel {
					return
				}
				oscmd.ExeFetch(python, e, func(data string) {
					if data == "kill" {
						isCancel = true
						time.Sleep(time.Millisecond * 300)
					}
					fmt.Printf("data: %v\n", data)
				})

				<-c5
			}(e)
		}

		return c.SendString("Download has started")
	})

	app.Listen(":3000")
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
