package download

import (
	"archive/tar"
	"compress/bzip2"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"sync"
)

// RequestDownload represents the incoming JSON payload.
type RequestDownload struct {
	GTFSURLs []string `json:"gtfsUrls"`
	OsmURL   string   `json:"osmUrl"`
	MotisUrl string   `json:"motisUrl"`
}

// extractFileName returns the base file name from a URL.
func extractFileName(url string) string {
	return path.Base(url)
}

// ProgressCallback is a function type called with progress updates.
// fileName: name of the file being downloaded.
// downloaded: number of bytes downloaded so far.
// total: total number of bytes to download (if known, otherwise 0).
type ProgressCallback func(fileName string, downloaded int64, total int64)

// ProgressWriter wraps an io.Writer and reports progress after each write.
type ProgressWriter struct {
	Writer   io.Writer
	FileName string
	Total    int64
	Current  int64
	Callback ProgressCallback
}

func (pw *ProgressWriter) Write(p []byte) (n int, err error) {
	n, err = pw.Writer.Write(p)
	if n > 0 {
		pw.Current += int64(n)
		if pw.Callback != nil {
			pw.Callback(pw.FileName, pw.Current, pw.Total)
		}
	}
	return n, err
}

// downloadFileWithProgress downloads a file from the given URL and writes it to outDir
// using the file's base name. It calls progressCallback with progress updates.
func downloadFileWithProgress(url, outDir string, progressCallback ProgressCallback) error {
	fileName := extractFileName(url)
	outPath := filepath.Join(outDir, fileName)

	// Create the output file.
	outFile, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", outPath, err)
	}
	defer outFile.Close()

	// Send HTTP GET request.
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to GET %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status downloading %s: %s", url, resp.Status)
	}

	// Determine total size if available.
	var totalSize int64 = 0
	if cl := resp.Header.Get("Content-Length"); cl != "" {
		totalSize, err = strconv.ParseInt(cl, 10, 64)
		if err != nil {
			totalSize = 0
		}
	}

	// Wrap the file writer with a progress writer.
	pw := &ProgressWriter{
		Writer:   outFile,
		FileName: fileName,
		Total:    totalSize,
		Callback: progressCallback,
	}

	// Copy the response body into the file via the progress writer.
	_, err = io.Copy(pw, resp.Body)
	if err != nil {
		return fmt.Errorf("failed copying data for %s: %w", url, err)
	}

	return nil
}

// extractTarBz2 extracts a .tar.bz2 archive (Motis) into outDir.
func extractTarBz2(filePath, outDir string) error {
	// Open the tar.bz2 file.
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer f.Close()

	// Create a bzip2 reader.
	bz2Reader := bzip2.NewReader(f)
	tarReader := tar.NewReader(bz2Reader)

	// Iterate over tar entries.
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive.
		}
		if err != nil {
			return fmt.Errorf("error reading tar header: %w", err)
		}

		target := filepath.Join(outDir, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory if needed.
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create dir %s: %w", target, err)
			}
		case tar.TypeReg:
			// Ensure the directory exists.
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("failed to create dir for file %s: %w", target, err)
			}
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", target, err)
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return fmt.Errorf("error writing file %s: %w", target, err)
			}
			outFile.Close()
		default:
			// Skip any other types.
			fmt.Printf("Skipping unknown type: %v in file %s\n", header.Typeflag, header.Name)
		}
	}

	return nil
}

// DownloadAll downloads all files (GTFS, Osm, and Motis) concurrently with a maximum
// of 5 simultaneous downloads. It calls progressCallback with progress updates for each file.
// After all downloads complete, it extracts the Motis archive.
func DownloadAll(req RequestDownload, progressCallback ProgressCallback) error {
	outDir := "out"
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("failed to create out folder: %w", err)
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 5) // Limit to 5 concurrent downloads.
	errorsChan := make(chan error, len(req.GTFSURLs)+2)

	// Helper function to run a single download task.
	downloadTask := func(url string, taskName string) {
		defer wg.Done()
		sem <- struct{}{}
		defer func() { <-sem }()
		fmt.Printf("Starting download for %s: %s\n", taskName, url)
		if err := downloadFileWithProgress(url, outDir, progressCallback); err != nil {
			errorsChan <- fmt.Errorf("%s download error for %s: %w", taskName, url, err)
			return
		}
		fmt.Printf("%s finished: %s\n", taskName, url)
	}

	// Download all GTFS URLs.
	for _, url := range req.GTFSURLs {
		wg.Add(1)
		go downloadTask(url, "GTFS")
	}

	// Download the Osm file.
	wg.Add(1)
	go downloadTask(req.OsmURL, "Osm")

	// Download the Motis file.
	wg.Add(1)
	go downloadTask(req.MotisUrl, "Motis")

	// Wait for all downloads to finish.
	wg.Wait()
	close(errorsChan)
	fmt.Println("All downloads finished.")
	for err := range errorsChan {
		if err != nil {
			return err
		}
	}

	// After downloading, extract the Motis archive.
	motisFileName := extractFileName(req.MotisUrl)
	motisFilePath := filepath.Join(outDir, motisFileName)
	fmt.Printf("Extracting Motis file: %s\n", motisFilePath)
	if err := extractTarBz2(motisFilePath, outDir); err != nil {
		return fmt.Errorf("failed extracting motis file: %w", err)
	}

	return nil
}
