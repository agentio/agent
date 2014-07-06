package agent

import (
	"code.google.com/p/go-uuid/uuid"
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"os"
	"os/exec"
	"time"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func getMongoSession() (mongoSession *mgo.Session) {
        var dialInfo mgo.DialInfo
        dialInfo.Addrs = []string{"127.0.0.1"}
        dialInfo.Username = os.Getenv("MONGODB_USERNAME")
        dialInfo.Password = os.Getenv("MONGODB_PASSWORD")
        mongoSession, err := mgo.DialWithInfo(&dialInfo)
	check(err)
	mongoSession.SetMode(mgo.Monotonic, true)
	return mongoSession
}

func md5HashWithSalt(input, salt string) string {
	hasher := hmac.New(md5.New, []byte(salt))
	hasher.Write([]byte(input))
	return hex.EncodeToString(hasher.Sum(nil))
}

func authorizeUser(username string, password string) (user User, err error) {

	mongoSession := getMongoSession()
        defer mongoSession.Close()

	saltedPassword := md5HashWithSalt(password, PasswordSalt)
	usersCollection := mongoSession.DB("accounts").C("users")
	err = usersCollection.Find(bson.M{"username": username, "password": saltedPassword}).One(&user)
	return user, err
}

func getApps(apps *[]App) (err error) {

	mongoSession := getMongoSession()
        defer mongoSession.Close()

	appsCollection := mongoSession.DB("control").C("apps")
	return appsCollection.Find(nil).All(apps)
}

func createApp(app map[string]interface{}) (appid bson.ObjectId, err error) {

	mongoSession := getMongoSession()
        defer mongoSession.Close()

	appsCollection := mongoSession.DB("control").C("apps")
	newId := bson.NewObjectId()
	app["_id"] = newId
	err = appsCollection.Insert(app)
	return newId, err
}

func deleteApp(app App) {

	mongoSession := getMongoSession()
        defer mongoSession.Close()

	oid := app.Id
	appsCollection := mongoSession.DB("control").C("apps")
	err := appsCollection.Remove(bson.M{"_id": oid})
	check(err)
}

func deleteAllApps() (err error) {

	mongoSession := getMongoSession()
        defer mongoSession.Close()

	appsCollection := mongoSession.DB("control").C("apps")
	_, err = appsCollection.RemoveAll(bson.M{})
	return err
}

func getApp(appid string, app *App) (err error) {

	mongoSession := getMongoSession()
        defer mongoSession.Close()

	appsCollection := mongoSession.DB("control").C("apps")
        if bson.IsObjectIdHex(appid) {
		oid := bson.ObjectIdHex(appid)
		return appsCollection.Find(bson.M{"_id": oid}).One(&app)
        } else {
		return appsCollection.Find(bson.M{"name": appid}).One(&app)
        }
}

func addAppVersion(
	app App,
	appfilename string,
	appfiledata []byte) Version {

	mongoSession := getMongoSession()
        defer mongoSession.Close()

	versionid := string(uuid.New())

	db := mongoSession.DB("control")
	file, err := db.GridFS("appfiles").Create(versionid)
	check(err)
	n, err := file.Write(appfiledata)
	check(err)
	err = file.Close()
	check(err)
	fmt.Printf("%d bytes written\n", n)

	version := Version{
		Version:  versionid,
		Filename: appfilename,
		Created:  time.Now(),
	}

	var versions []Version
	tempversions := app.Versions
	if tempversions == nil {
		versions = []Version{version}
	} else {
		versions = append(tempversions, version)
	}

	appsCollection := mongoSession.DB("control").C("apps")
	update := map[string]interface{}{"versions": versions}
	appsCollection.Update(bson.M{"_id": app.Id}, bson.M{"$set": update})

	return version
}

func getBusyPorts() map[uint32]uint32 {

	mongoSession := getMongoSession()
        defer mongoSession.Close()

	busyPorts := map[uint32]uint32{}
	appsCollection := mongoSession.DB("control").C("apps")
	var apps []App
	err := appsCollection.Find(nil).All(&apps)
	check(err)
	for _, app := range apps {
		workers := app.Workers
		for _, worker := range workers {
			port := worker.Port
			busyPorts[port] = 1
		}
	}
	return busyPorts
}

func deployAppVersion(
	app App,
	versionid string) bool {

	appname := app.Name

	// get the specified app version
	var version Version
	found := false
	versions := app.Versions
	for _, thisversion := range versions {
		thisversionid := thisversion.Version
		if thisversionid == versionid {
			version = thisversion
			found = true
		}
	}
	if !found {
		fmt.Printf("can't find version %v\n", version)

		return false // no version
	}
	fmt.Printf("deploying %v\n", version)

	// get the busy ports
	busyPorts := getBusyPorts()
	var port uint32
	port = 9000
	fmt.Printf("busy ports %v\n", busyPorts)

	workers := make([]Worker, 0)

	mongoSession := getMongoSession()
        defer mongoSession.Close()

	// for each worker
	var i uint32
	for i = 0; i < app.Capacity; i++ {
		//   create a worker id
		workerid := string(uuid.New())
		//   create the container directory
		workerpath := fmt.Sprintf("%v/workers/%v", ControlPath, workerid)
		os.Mkdir(workerpath, os.ModeDir+os.ModePerm)
		//   extract the version data and write it to the container directory

		db := mongoSession.DB("control")
		file, err := db.GridFS("appfiles").Open(versionid)
		check(err)
		data, err := ioutil.ReadAll(file)
		ioutil.WriteFile(fmt.Sprintf("%v/%v.zip", workerpath, appname), data, 0644)
		//   unzip the version data
		cwd, _ := os.Getwd()
		os.Chdir(workerpath)
		zipname := fmt.Sprintf("%v.zip", appname)
		out, err := exec.Command("/usr/bin/unzip", zipname).Output()
		fmt.Printf("output: %v\n", string(out))
		check(err)
		os.Chdir(cwd)
		//   create a var subdirectory of the worker container
		varpath := fmt.Sprintf("%v/var", workerpath)
		os.Mkdir(varpath, os.ModeDir+os.ModePerm)
		//   touch var/stdout.log and var/stderr.log

		//   (set command (+ "chmod -R ugo+rX " path))
		//   (set command (+ "chown -R control " path "/var"))
		//   (set command (+ "chmod -R ug+w " path "/var"))
		os.Chmod(workerpath, 0755)
		os.Chmod(varpath, 0775)
		chown := exec.Command("chown", "-R", "control", varpath)
		chown.Run()

		//   assign a port to the app
		port++
		for {
			_, present := busyPorts[port]
			if present {
				port++
			} else {
				break
			}
		}
		fmt.Printf("PORT %v\n", port)

		var worker Worker
		worker.Port = port
		worker.Host = "localhost"
		worker.Container = workerid
		worker.Version = versionid

                if (true) {
			// generate the docker configuration
			GenerateDockerConfiguration(app, worker)

			// build the docker image

			// stop any running docker images

			// launch the docker image

                } else {
			// generate the upstart configuration
			GenerateUpstartConfiguration(app, worker)

			// stop the app in case it was already running
			stopCommand := exec.Command("/sbin/initctl", "stop", fmt.Sprintf("agentio-worker-%v", port))
			stopCommand.Run()
	
			// load the upstart configuration to start the app
			startCommand := exec.Command("/sbin/initctl", "start", fmt.Sprintf("agentio-worker-%v", port))
			startCommand.Run()
                }

		// add the worker info to the workers list
		workers = append(workers, worker)
	}
	// add the workers to the app document (and save it)
	app.Workers = workers

	fmt.Printf("APP: %+v\n", app)

	appsCollection := mongoSession.DB("control").C("apps")
	update := map[string]interface{}{"workers": workers}
	appsCollection.Update(bson.M{"_id": app.Id}, bson.M{"$set": update})

	// regenerate nginx config and restart nginx
	RestartNginx()

	return true
}

func deleteAppVersions(app App) bool {

	mongoSession := getMongoSession()
        defer mongoSession.Close()

	db := mongoSession.DB("control")

	versions := app.Versions
	for _, thisversion := range versions {
		thisversionid := thisversion.Version
		err := db.GridFS("appfiles").Remove(thisversionid)
		check(err)
	}

	newversions := make([]Version, 0)

	appsCollection := mongoSession.DB("control").C("apps")
	update := map[string]interface{}{"versions": newversions}
	appsCollection.Update(bson.M{"_id": app.Id}, bson.M{"$set": update})

	return true
}

func deleteAppVersion(
	app App,
	versionid string) bool {

	mongoSession := getMongoSession()
        defer mongoSession.Close()

	db := mongoSession.DB("control")

	newversions := make([]string, 0)

	// get the specified app version
	versions := app.Versions
	for _, thisversion := range versions {
		thisversionid := thisversion.Version
		if thisversionid == versionid {
			fmt.Printf("deleting %v\n", thisversion)
			err := db.GridFS("appfiles").Remove(versionid)
			check(err)
		} else {
			newversions = append(newversions, thisversionid)
		}
	}

	appsCollection := mongoSession.DB("control").C("apps")
	update := map[string]interface{}{"versions": newversions}
	appsCollection.Update(bson.M{"_id": app.Id}, bson.M{"$set": update})

	return true
}

func deployApp(app App) bool {
	// deploy the most recent version
	if app.Versions != nil {
		versions := app.Versions
		if len(versions) > 0 {
			last := len(versions) - 1
			version := versions[last]
			versionid := version.Version
			return deployAppVersion(app, versionid)
		} else {
			return false
		}
	}
	return false
}

func stopApp(app App) bool {
	if len(app.Workers) == 0 {
		return false
	}
	workers := app.Workers
	for _, item := range workers {
		worker := item

		port := worker.Port

		// unload the upstart configuration to stop the app
		stop := exec.Command("/sbin/initctl", "stop", fmt.Sprintf("agentio-worker-%v", port))
		stop.Run()

		workerid := worker.Container

		// delete the upstart file
		RemoveUpstartConfiguration(app, worker)

		// remove the worker container
		workerpath := fmt.Sprintf("%v/workers/%v", ControlPath, workerid)
		os.RemoveAll(workerpath)
	}

	mongoSession := getMongoSession()
        defer mongoSession.Close()

	appsCollection := mongoSession.DB("control").C("apps")
	update := map[string]interface{}{"workers": make([]Worker, 0)}
	appsCollection.Update(bson.M{"_id": app.Id}, bson.M{"$set": update})

	// regenerate nginx config and restart nginx
	RestartNginx()

	return true
}

func stopAppVersion(
	app App,
	versionid string) bool {

	if len(app.Workers) == 0 {
		return false
	}
	deploymentversion := app.Workers[0].Version
	if deploymentversion == versionid {
		return stopApp(app)
	} else {
		return false
	}
}
