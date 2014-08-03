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
	"labix.org/v2/mgo/bson"
	"os"
	"os/exec"
	"path"
	"runtime"
	"text/template"
	"time"
)

func mydir() string {
	_, filename, _, _ := runtime.Caller(1)
	return path.Dir(filename)
}

func hexValue(args bson.ObjectId) string {
	return args.Hex()
}

func appHasAPath(app App) bool {
	// the app must have a path
	paths := app.Paths
	if len(paths) == 0 {
		return false
	}
	return true
}

func appHasADomain(app App) bool {
	// the app must have a domain
	domains := app.Domains
	if len(domains) == 0 {
		return false
	}
	return true
}

func appHasWorkers(app App) bool {
	// the app must have a deployment
	if len(app.Workers) == 0 {
		return false
	}
	return true
}

func agentHasSSL() bool {
    if _, err := os.Stat("/home/control/etc/agent.key"); err == nil {
        if _, err := os.Stat("/home/control/etc/agent.crt"); err == nil {
            return true
        }
    }
    return false
}

func GenerateNginxConfiguration(filename string, apps []App) {
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
		"Apps":        apps,
	}

	t := template.New("nginx configuration")
	t = t.Funcs(template.FuncMap{
		"hex":           hexValue,
		"appHasAPath":   appHasAPath,
		"appHasADomain": appHasADomain,
		"appHasWorkers": appHasWorkers,
		"agentHasSSL":   agentHasSSL,
	})

	t, err = t.ParseFiles(path.Join(mydir(), "nginx.txt"))
	err = t.ExecuteTemplate(fo, "nginx", args)
	if err != nil {
		panic(err)
	}
}

func fetchApps() []App {
	mongoSession := getMongoSession()
	defer mongoSession.Close()
	appsCollection := mongoSession.DB("control").C("apps")
	var apps []App
	err := appsCollection.Find(bson.M{}).All(&apps)
	if err != nil {
		panic(err)
	}
	return apps
}

func RestartNginx() {
	os.Remove(PathForNginxConfiguration)

	apps := fetchApps()

	GenerateNginxConfiguration(PathForNginxConfiguration, apps)
	// system call to reload nginx
	reload := exec.Command(PathForNginx,
		"-s", "reload",
		"-c", PathForNginxConfiguration,
		"-p", fmt.Sprintf("%v/nginx/", ControlPath))
	reload.Run()
}

func PrimeNginx() {
	os.Remove(PathForNginxConfiguration)
	var apps []App
	GenerateNginxConfiguration(PathForNginxConfiguration, apps)
	/*
	   ;; control redirect
	   ((&a href:(+ "/control") "OK, Continue")
	    writeToFile:"#{CONTROLPATH}/public/restart.html" atomically:NO))
	*/
}
