package agent

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
        "sort"
	"strings"
)

func dumpRequest(r *http.Request) {
	fmt.Println("Form")
	for key, value := range r.Form {
		fmt.Println("Key:", key, "Value:", value)
	}
	fmt.Println("Headers")
	for key, value := range r.Header {
		fmt.Println("Key:", key, "Value:", value)
	}
}

// generate an appropriate response
func respondWithResult(w http.ResponseWriter, result interface{}) {
	jsonData, err := json.Marshal(result)
	if err != nil {
		fmt.Println(err)
	}
	w.Write(jsonData)
}

func authorize(r *http.Request) (user User, err error) {
	authorization := r.Header["Authorization"]
	if len(authorization) == 1 {
		fields := strings.Fields(authorization[0])
		authorizationType := strings.ToLower(fields[0])
		authorizationToken := fields[1]
		if authorizationType == "basic" {
			data, err := base64.StdEncoding.DecodeString(authorizationToken)
			if err != nil {
				fmt.Println("error:", err)
				return user, err
			}
			credentials := string(data)
			pair := strings.SplitN(credentials, ":", 2)
			user, err := authorizeUser(pair[0], pair[1])
			return user, err
		} else {
			return user, errors.New(fmt.Sprintf("Unsupported authorization type: %s", authorizationType))
		}
	} else {
		return user, errors.New("No authorization header")
	}
	return user, err
}

func authorizedHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := authorize(r)
		if err != nil {
			w.WriteHeader(401)
			w.Write([]byte("Unauthorized"))
			return
		}
		fn(w, r)
	}
}

/*
 HANDLERS
*/

// get a list of apps
func getAppsHandler(w http.ResponseWriter, r *http.Request) {
	var apps []App
	err := getApps(&apps)
	check(err)
	respondWithResult(w, apps)
}

// create an app
func postAppsHandler(w http.ResponseWriter, r *http.Request) {
	buffer, err := ioutil.ReadAll(r.Body)
	var app map[string]interface{}
	err = json.Unmarshal(buffer, &app)
	if err != nil {
		panic(err)
	}
	appid, err := createApp(app)
	check(err)
	result := map[string]interface{}{
		"message": "OK",
		"appid":   appid,
	}
	respondWithResult(w, result)
}

// delete all apps
func deleteAppsHandler(w http.ResponseWriter, r *http.Request) {
	err := deleteAllApps()
	check(err)
	result := map[string]interface{}{
		"message": "OK",
		//"updated": info.Updated,
		//"removed": info.Removed,
		//		    UpsertedId interface{} // Upserted _id field, when not explicitly provided
	}
	respondWithResult(w, result)
}

// get an app
func getAppHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appid := vars["appid"]
	var app App
	err := getApp(appid, &app)
	check(err)
	respondWithResult(w, app)
}

// post a command to an app
func postAppHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("post app handler")

	err := r.ParseForm()
	if err != nil {
		//handle error http.Error() for example
		panic(err)
	}
	command := r.Form["command"][0]
	fmt.Printf("COMMAND %v\n", command)

	vars := mux.Vars(r)
	appid := vars["appid"]
	var app App
	err = getApp(appid, &app)
	check(err)

	if command == "start" {
		var result map[string]interface{}
		if deployApp(app) {
			result = map[string]interface{}{
				"message": "OK",
			}
		} else {
			result = map[string]interface{}{
				"message": "error: unable to deploy app",
			}
		}
		respondWithResult(w, result)
	} else if command == "stop" {
		var result map[string]interface{}
		if stopApp(app) {
			result = map[string]interface{}{
				"message": "OK",
			}
		} else {
			result = map[string]interface{}{
				"message": "error: unable to stop app",
			}
		}
		respondWithResult(w, result)
	}
}

// delete an app
func deleteAppHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appid := vars["appid"]
	var app App
	err := getApp(appid, &app)
	check(err)
	deleteAppVersions(app)
	deleteApp(app)
	result := map[string]interface{}{
		"message": "OK",
	}
	respondWithResult(w, result)
}

// get a list of app versions
func getAppVersionsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appid := vars["appid"]
	var app App
	err := getApp(appid, &app)
	check(err)
	respondWithResult(w, app.Versions)
}

// add an app version
func postAppVersionsHandler(w http.ResponseWriter, r *http.Request) {
	appfiledata, err := ioutil.ReadAll(r.Body)
	check(err)

	vars := mux.Vars(r)
	appid := vars["appid"]

	var app App
	err = getApp(appid, &app)
	check(err)

	appfilename := fmt.Sprintf("%v.zip", app.Name)

	version := addAppVersion(app, appfilename, appfiledata)

	result := map[string]interface{}{
		"message": "OK",
		"version": version.Version,
	}
	respondWithResult(w, result)
}

// delete all app versions
func deleteAppVersionsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appid := vars["appid"]
	var app App
	err := getApp(appid, &app)
	check(err)
	if deleteAppVersions(app) {
		result := map[string]interface{}{
			"message": "OK",
		}
		respondWithResult(w, result)
	} else {
		result := map[string]interface{}{
			"message": "error: unable to delete versions",
		}
		respondWithResult(w, result)
	}
}

// get information about an app version
func getAppVersionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	versionid := vars["versionid"]

	mongoSession := getMongoSession()
        defer mongoSession.Close()

	db := mongoSession.DB("control")
	file, err := db.GridFS("appfiles").Open(versionid)
	check(err)
	data, err := ioutil.ReadAll(file)
	w.Write(data)
}

// send a command to an app version
func postAppVersionHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		//handle error http.Error() for example
		return
	}
	command := r.Form["command"][0]

	vars := mux.Vars(r)
	appid := vars["appid"]
	versionid := vars["versionid"]

	var app App
	err = getApp(appid, &app)
	check(err)

	fmt.Printf("posted command %v\n", command)

	if command == "start" {
		var result map[string]interface{}
		if deployAppVersion(app, versionid) {
			result = map[string]interface{}{
				"message": "OK",
			}
		} else {
			result = map[string]interface{}{
				"message": "error: unable to deploy app",
			}
		}
		respondWithResult(w, result)
	} else if command == "stop" {
		var result map[string]interface{}
		if stopAppVersion(app, versionid) {
			result = map[string]interface{}{
				"message": "OK",
			}
		} else {
			result = map[string]interface{}{
				"message": "error: unable to stop app",
			}
		}
		respondWithResult(w, result)
	}
}

// delete an app version
func deleteAppVersionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appid := vars["appid"]
	versionid := vars["versionid"]

	var app App
	err := getApp(appid, &app)
	check(err)

	if deleteAppVersion(app, versionid) {
		result := map[string]interface{}{
			"message": "OK",
		}
		respondWithResult(w, result)
	} else {
		result := map[string]interface{}{
			"message": "error: unable to delete version",
		}
		respondWithResult(w, result)
	}
}

// get the logfile for a worker
func getWorkerLogHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workerid := vars["workerid"]
	var log string
	err := getLogForWorker(workerid, &log)
	check(err)
	w.Write([]byte(log))
}

func getPortsHandler(w http.ResponseWriter, r *http.Request) {
 	busyPorts := getBusyPorts()

	var keys []int
	for k := range busyPorts {
    		keys = append(keys, int(k))
	}
	sort.Ints(keys)

	jsonData, err := json.Marshal(keys)
        check(err)
        w.Write(jsonData)
}

func hostnameHandler(w http.ResponseWriter, r *http.Request) {
	_, err := authorize(r)
	if err != nil {
		w.WriteHeader(401)
		return
	}
	if r.Method == "GET" {
		b, err := ioutil.ReadFile("/etc/hosts")
		if err != nil {
			panic(err)
		}
		w.Write(b)
	} else if r.Method == "POST" {
		r.ParseForm()
		hostname := r.Form["hostname"][0]
		w.Write([]byte(hostname))
	}
}

var API = []struct {
	path        string
	method      string
	handler     func(http.ResponseWriter, *http.Request)
	description string
}{
	{"/control/apps", "GET", getAppsHandler, "get list of apps"},
	{"/control/apps", "POST", postAppsHandler, "create an app or send a command to all apps (start, stop)"},
	{"/control/apps", "DELETE", deleteAppsHandler, "delete all apps"},
	{"/control/apps/{appid}", "GET", getAppHandler, "get an app"},
	{"/control/apps/{appid}", "POST", postAppHandler, "send a command to an app (start, stop)"},
	{"/control/apps/{appid}", "DELETE", deleteAppHandler, "delete an app"},
	{"/control/apps/{appid}/versions", "GET", getAppVersionsHandler, "get versions of an app"},
	{"/control/apps/{appid}/versions", "POST", postAppVersionsHandler, "create a version or send a command to all versions (?)"},
	{"/control/apps/{appid}/versions", "DELETE", deleteAppVersionsHandler, "delete all versions of an app"},
	{"/control/apps/{appid}/versions/{versionid}", "GET", getAppVersionHandler, "get a version of an app"},
	{"/control/apps/{appid}/versions/{versionid}", "POST", postAppVersionHandler, "send a command to a version of an app (start, stop)"},
	{"/control/apps/{appid}/versions/{versionid}", "DELETE", deleteAppVersionHandler, "delete a version of an app"},
	{"/control/workers/{workerid}/log", "GET", getWorkerLogHandler, "get logfile for a worker"},
	{"/control/ports", "GET", getPortsHandler, "get busy ports"},
}

func ControlAPI() {
	r := mux.NewRouter()
	r.HandleFunc("/control/hostname", hostnameHandler).Methods("GET", "POST")

	for _, endpoint := range API {
		// extract the pieces
		path := endpoint.path
		method := endpoint.method
		handler := endpoint.handler
		description := endpoint.description
		// register the handler
		fmt.Printf("adding %v %v # %v\n", method, path, description)
		r.HandleFunc(path, authorizedHandler(handler)).Methods(method)
	}
	http.Handle("/", r)

	log.Fatal(http.ListenAndServe(":2010", nil))
}
