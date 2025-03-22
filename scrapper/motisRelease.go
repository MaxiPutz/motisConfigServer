package scrapper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type Asset struct {
	Name                 string `json:"name"`
	Browser_download_url string `json:"browser_download_url"`
	Os                   string `json:"os"`
	Arch                 string `json:"arch"`
}

// Release represents the minimal JSON fields from the GitHub API.
type Release struct {
	TagName string  `json:"tag_name"`
	Name    string  `json:"name"`
	Assets  []Asset `json:"assets"`
}

func (a *Asset) parseOsAndAsset() {
	fmt.Printf("\"intheparser\": %v\n", "intheparser")
	isWindows := strings.Contains(a.Name, "windows")
	if isWindows {
		a.Os = "windows"
		a.Arch = "amd64"
		fmt.Printf("a: %v\n", a)
		return
	}
	isMacos := strings.Contains(a.Name, "macos")
	isArm := strings.Contains(a.Name, "arm")
	if isArm {
		a.Arch = "arm64"
	} else {
		a.Arch = "x86"
	}
	if isMacos {
		a.Os = "macos"
		return
	} else {
		a.Os = "linux"
		return
	}

}

// fetchReleases retrieves releases from GitHub for the given page.
func FetchReleases(page int) ([]Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/motis-project/motis/releases?per_page=100&page=%d", page)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	var releases []Release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}

	for i := range releases {
		for j := range releases[i].Assets {
			releases[i].Assets[j].parseOsAndAsset()
		}
	}

	return releases, nil
}

func FetchAll() ([]Release, error) {

	page := 1
	var allReleases []Release
	for {
		releases, err := FetchReleases(page)
		if err != nil {
			return nil, fmt.Errorf("error fetching page %d: %v", page, err)
		}
		if len(releases) == 0 {
			break
		}
		allReleases = append(allReleases, releases...)
		page++
	}
	err := writeInToAssets(allReleases)
	if err != nil {
		panic(err)
		return nil, err
	}
	return allReleases, nil
}

func PrintRelease(realses []Release) {
	fmt.Println("Releases:")
	for _, release := range realses {
		fmt.Printf("- %s (%s)\n", release.Name, release.TagName)
		for _, a := range release.Assets {
			fmt.Printf("a.Name: %v\n", a.Name)
			fmt.Printf("a.Os: %v\n", a.Os)
			fmt.Printf("a.Arch: %v\n", a.Arch)
		}
	}
}

func writeInToAssets(releases []Release) error {

	data, err := json.MarshalIndent(releases, "", "  ")

	fmt.Printf("data: %v\n", data)
	if err != nil {
		return err
	}

	err = os.WriteFile("assets/motis.json", data, 0664)
	if err != nil {
		return err
	}

	return nil
}
