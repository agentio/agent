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

