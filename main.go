package main

import (
	"os"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/metricbeat/beater"

	// get system stats modules from metricbeat
	_ "github.com/elastic/beats/metricbeat/module/system"
	_ "github.com/elastic/beats/metricbeat/module/system/core"
	_ "github.com/elastic/beats/metricbeat/module/system/cpu"
	_ "github.com/elastic/beats/metricbeat/module/system/disk"
	_ "github.com/elastic/beats/metricbeat/module/system/filesystem"
	_ "github.com/elastic/beats/metricbeat/module/system/fsstat"
	_ "github.com/elastic/beats/metricbeat/module/system/memory"
	_ "github.com/elastic/beats/metricbeat/module/system/process"

	// get beats modules
	_ "github.com/urso/collectbeat/module/beats/generatorbeat"
)

// Name of this Beat.
var Name = "collectbeat"

func main() {
	if err := beat.Run(Name, "", beater.New()); err != nil {
		os.Exit(1)
	}
}
