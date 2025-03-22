package scrapper

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Region represents a node in the tree.
type Region struct {
	Path       string   `json:"path"`
	IsLeaf     bool     `json:"isLeaf"`
	RegionName string   `json:"regionName"`
	OsmData    string   `json:"osmData"`
	RootPath   string   `json:"rootPath"`
	RawHtml    string   `json:"rawHtml"`
	Children   []Region `json:"children"`
}

// Geofabrik encapsulates the information needed to fetch a page.
type Geofabrik struct {
	Path     string
	Prefix   string
	Depth    int
	RootPath string
	BaseURL  string
}

// NewGeofabrik creates a new Geofabrik instance.
// If path is non-empty, it ensures that it starts with "/".
func NewGeofabrik(path, prefix string, depth int, rootPath string) *Geofabrik {
	if path != "" && !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return &Geofabrik{
		Path:     path,
		Prefix:   prefix,
		Depth:    depth,
		RootPath: rootPath,
		BaseURL:  "https://download.geofabrik.de",
	}
}

// getHTML fetches and parses the HTML document at BaseURL+Path.
func (g *Geofabrik) getHTML() (*goquery.Document, error) {
	url := g.BaseURL + g.Path
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching %s: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML: %v", err)
	}
	return doc, nil
}

// getChildRaw finds all rows that contain a subregion.
// It does so by selecting all td.subregion elements and retrieving their parent rows.
func (g *Geofabrik) getChildRaw() ([]*goquery.Selection, error) {
	doc, err := g.getHTML()
	if err != nil {
		return nil, err
	}
	var rows []*goquery.Selection
	doc.Find("td.subregion").Each(func(i int, s *goquery.Selection) {
		// Assuming the parent is a <tr> element.
		parent := s.Parent()
		if parent != nil {
			rows = append(rows, parent)
		}
	})
	return rows, nil
}

// getChildNodes extracts the data from each row into a Region object.
// It filters out rows where the second cell does not contain an anchor.
func (g *Geofabrik) getChildNodes() ([]Region, error) {
	rows, err := g.getChildRaw()
	if err != nil {
		return nil, err
	}
	var regions []Region
	for _, row := range rows {
		// Find all <td> elements in the row.
		cells := row.Find("td")
		if cells.Length() < 2 {
			continue
		}

		// Check that the second cell contains an anchor.
		secondCell := cells.Eq(1)
		if secondCell.Find("a").Length() == 0 {
			continue
		}

		// Get first cell anchor for path and region name.
		firstCell := cells.Eq(0)
		aTag := firstCell.Find("a").First()
		path, exists := aTag.Attr("href")
		if !exists {
			continue
		}
		regionName := strings.TrimSpace(aTag.Text())

		// Get OSM data from second cell anchor.
		osmData := ""
		osmATag := secondCell.Find("a").First()
		if href, ok := osmATag.Attr("href"); ok {
			osmData = href
		}

		// Determine if this is a leaf node by checking the last cell.
		lastCell := cells.Last()
		isLeaf := lastCell.Find("a").Length() > 0

		regions = append(regions, Region{
			Path:       path,
			IsLeaf:     isLeaf,
			RegionName: regionName,
			OsmData:    osmData,
			RootPath:   g.RootPath,
			RawHtml:    g.BaseURL,
			Children:   nil,
		})
	}
	return regions, nil
}

// GetTree recursively builds the tree of regions.
func (g *Geofabrik) GetTree() (Region, error) {
	childNodes, err := g.getChildNodes()
	if err != nil {
		return Region{}, err
	}

	for i, child := range childNodes {
		if !child.IsLeaf {
			childPath := child.Path
			// Adjust the path if the current depth is 2 or more.
			if g.Depth >= 2 {
				parts := strings.Split(strings.TrimPrefix(g.Path, "/"), "/")
				if len(parts) > 0 {
					childPath = parts[0] + "/" + child.Path
				}
			}
			// Create a new instance for the child node.
			node := NewGeofabrik(childPath, g.Prefix, g.Depth+1, g.Path)
			subtree, err := node.GetTree()
			if err != nil {
				log.Printf("Error fetching tree for %s: %v", child.Path, err)
				child.Children = []Region{}
			} else {
				child.Children = subtree.Children
			}
		} else {
			child.Children = []Region{}
		}
		// Update the slice with the possibly modified child.
		childNodes[i] = child
	}

	return Region{
		Path:       g.Path,
		IsLeaf:     false, // The current node is a directory.
		RegionName: "",    // You can set a name for the root if desired.
		OsmData:    "",
		RootPath:   g.RootPath,
		RawHtml:    g.BaseURL,
		Children:   childNodes,
	}, nil
}

func GetOsm() (Region, error) {
	root := NewGeofabrik("", "", 0, "")
	tree, err := root.GetTree()
	return tree, err
}

func UpdateGeofabrik() {
	// Create the root Geofabrik instance.
	root := NewGeofabrik("", "", 0, "")
	tree, err := root.GetTree()
	if err != nil {
		log.Fatalf("Error getting tree: %v", err)
	}
	fmt.Printf("Tree: %+v\n", tree)

	// Serialize tree to JSON
	data, err := json.MarshalIndent(tree, "", " ")
	if err != nil {
		log.Fatalf("Error marshalling tree to JSON: %v", err)
	}

	// Write new data to a temporary file.
	tmpFile := "assets/tmp/geofabrik.json"
	err = os.WriteFile(tmpFile, data, 0664)
	if err != nil {
		log.Fatalf("Error writing tmp file: %v", err)
	}
	_, err = os.Stat("assets/issueGeoFarbik")

	if err == nil {
		os.Remove("assets/issueGeoFarbik")
		log.Printf("Warning: Could not remove tmp file: %v", err)
	}

	// Define the main file path.
	mainFile := "assets/geofabrik.json"

	// Check if the main file exists.
	oldData, err := os.ReadFile(mainFile)
	if err != nil {

		fmt.Println("No changes detected. The main file remains unchanged.")
		if os.IsNotExist(err) {
			// No existing file: move the new file into place.
			err = os.Rename(tmpFile, mainFile)
			if err != nil {
				log.Fatalf("Error moving new file to main location: %v", err)
			}
			fmt.Println("No previous file found. New file moved to assets/geofabrik.json")
			return
		}
		log.Fatalf("Error reading existing main file: %v", err)
	}

	// Compare the existing file with the new data.
	if string(oldData) != string(data) {
		// Files are different, so archive the old file.
		archiveDir := "assets/archive"
		err = os.MkdirAll(archiveDir, 0755)
		if err != nil {

			log.Fatalf("Error creating archive directory: %v", err)
		}
		// Archive file name, for example: geofabrik-2025-03-18.json.
		archiveFile := filepath.Join(archiveDir, fmt.Sprintf("geofabrik-%s.json", time.Now().Format("2006-01-02")))
		err = os.Rename(mainFile, archiveFile)
		if err != nil {
			log.Fatalf("Error archiving old file: %v", err)
		}
		fmt.Printf("Old file archived to: %s\n", archiveFile)

		// Move the new file to the main location.
		err = os.Rename(tmpFile, mainFile)
		if err != nil {
			log.Fatalf("Error moving new file to main location: %v", err)
		}
		fmt.Println("New file moved to assets/geofabrik.json")

		// Create the issue file to notify the maintainer.
		issueFile := "assets/issueGeoFarbik"
		issueContent := []byte("Geofabrik tree file has changed. Please review the updates.")
		err = os.WriteFile(issueFile, issueContent, 0664)
		if err != nil {
			log.Fatalf("Error creating issue file: %v", err)
		}
		fmt.Printf("Issue file created: %s\n", issueFile)
	}
}
