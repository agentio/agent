//
//  Copyright 2014 Radtastical Inc.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.
//
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
