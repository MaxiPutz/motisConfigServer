package scrapper

import (
	"sync"
)

func GetAllAssetes() (Region, []Release, []Transitous, error) {
	region := Region{}
	relesa := []Release{}
	transitous := []Transitous{}
	var err error
	var wg sync.WaitGroup

	wg.Add(3)
	go func() {
		transitous, err = GetProcesGTFSLinks()
		wg.Done()
	}()
	go func() {
		relesa, err = FetchAll()
		wg.Done()
	}()
	go func() {
		region, err = GetOsm()
		wg.Done()
	}()

	wg.Wait()
	return region, relesa, transitous, err
}
