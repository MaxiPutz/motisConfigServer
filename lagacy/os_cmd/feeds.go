package oscmd

import (
	"fmt"
	"os"
)

func GetFeeds() []os.DirEntry {
	dirs, err := os.ReadDir("./feeds")
	if err != nil {
		dirs, err = os.ReadDir("../feeds")
		if err != nil {
			fmt.Printf("\"no feeds in the working directory found\": %v\n", "no feeds in the working directory found")
			panic(err)
		}
	}

	for _, dir := range dirs {
		fmt.Printf("dir.Name(): %v\n", dir.Name())
	}

	return dirs
}
