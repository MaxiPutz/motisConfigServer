package motisconfigfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GenerateConfigCommand(osmPath string, gtfsFiles []string, outputPath string) {
	commandStr := "./motis config " + osmPath + strings.Join(gtfsFiles, " ")

	os.WriteFile(outputPath+"runMotisConifg.sh", []byte(commandStr), 0664)
}

func GenerateMotisConfig(osmPath string, gtfsFiles []string, outputPath string) error {

	GenerateConfigCommand(osmPath, gtfsFiles, outputPath)

	config := strings.Builder{}

	// Base OSM and tile config
	config.WriteString(fmt.Sprintf(`osm: %s
tiles:
  profile: tiles-profiles/full.lua
  db_size: 274877906944
  flush_threshold: 100000
timetable:
  first_day: TODAY
  num_days: 365
  railviz: true
  with_shapes: true
  adjust_footpaths: true
  merge_dupes_intra_src: false
  merge_dupes_inter_src: false
  link_stop_distance: 100
  update_interval: 60
  http_timeout: 30
  incremental_rt_update: false
  use_osm_stop_coordinates: false
  extend_missing_footpaths: false
  max_footpath_length: 15
  max_matching_distance: 25
  datasets:
`, filepath.Base(osmPath)))

	// GTFS dataset entries
	for _, file := range gtfsFiles {
		key := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
		config.WriteString(fmt.Sprintf("    %s:\n", key))
		config.WriteString(fmt.Sprintf("      path: %s\n", filepath.Base(file)))
		config.WriteString("      default_bikes_allowed: false\n")
	}

	// Remaining static options
	config.WriteString(`street_routing: true
osr_footpath: false
geocoding: true
reverse_geocoding: true
`)

	// Write to file
	return os.WriteFile(outputPath+"config.yml", []byte(config.String()), 0644)
}
