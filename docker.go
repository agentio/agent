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

func GenerateDockerConfiguration(app App, worker Worker) {
	args := map[string]interface{}{
		"CONTROLPATH": ControlPath,
		"TIME":        time.Now(),
		"HOSTNAME":    HostName,
		"App":         app,
		"Worker":      worker,
	}
        var err error
	t := template.New("docker configuration")
	t, err = t.ParseFiles(path.Join(mydir(), "docker.txt"))
	if err != nil {
		panic(err)
	}

        filenames := []string{"Dockerfile","start","stop","rm","build","go-build"}

        for _,filename := range filenames {
		var filepath string
		filepath = fmt.Sprintf("/home/control/workers/%v/%v", worker.Container, filename)
		// open output file
		fo, err := os.Create(filepath)
		if err != nil {
			panic(err)
		}
		// close fo on exit and check for its returned error
		defer func() {
			if err := fo.Close(); err != nil {
				panic(err)
			}
		}()
		err = t.ExecuteTemplate(fo, filename, args)
		if err != nil {
			panic(err)
		}
                os.Chmod(filepath, 0777)
        }
}

