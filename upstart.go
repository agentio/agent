package agent

import (
	"fmt"
	"os"
	"path"
	"text/template"
	"time"
)

func GenerateUpstartConfiguration(app App, worker Worker) {

	var filename string
	filename = fmt.Sprintf("%v/agentio-worker-%v.conf", PathForUpstartConfiguration, worker.Port)

	// open output file
	fo, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	// close fo on exit and check for its returned error
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	args := map[string]interface{}{
		"CONTROLPATH": ControlPath,
		"TIME":        time.Now(),
		"HOSTNAME":    HostName,
		"App":         app,
		"Worker":      worker,
	}

	t := template.New("upstart configuration")
	t, err = t.ParseFiles(path.Join(mydir(), "upstart.txt"))
	err = t.ExecuteTemplate(fo, "upstart", args)
	if err != nil {
		panic(err)
	}

}

func RemoveUpstartConfiguration(app App, worker Worker) {
	var filename string
	filename = fmt.Sprintf("%v/agentio-worker-%v.conf", PathForUpstartConfiguration, worker.Port)
	os.Remove(filename)
}

func RegenerateUpstartConfigurations() {
	apps := fetchApps()
	for _, app := range apps {
		for _, worker := range app.Workers {
			GenerateUpstartConfiguration(app, worker)
		}
	}
}
