package agent

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func (c Connection) performRequestIntoBuffer(req *http.Request) (buffer []byte, err error) {
	client := &http.Client{}
	encoding := base64.StdEncoding.EncodeToString([]byte(c.Credentials))
	authorization := fmt.Sprintf("Basic %v", encoding)
	req.Header.Add("Authorization", authorization)
	response, err := client.Do(req)
	check(err)
	buffer, err = ioutil.ReadAll(response.Body)
	check(err)
	defer response.Body.Close()
	if response.StatusCode != 200 {
		fmt.Printf("status code %v\n", response.StatusCode)
		fmt.Println(string(buffer))
		panic("goodbye")
	}
	return
}

func (c Connection) performRequestIntoJSON(result interface{}, req *http.Request) (err error) {
	buffer, err := c.performRequestIntoBuffer(req)
	err = json.Unmarshal(buffer, result)
	return
}

func (c Connection) GetApps(result *[]App) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%v/control/apps", c.Service), nil)
	check(err)
	buffer, err := c.performRequestIntoBuffer(req)
	err = json.Unmarshal(buffer, result)
	check(err)
}

func (c Connection) CreateApp(result *map[string]interface{}, app App) {
	jsonData, err := json.Marshal(app)
	post_data := bytes.NewReader(jsonData)
	req, err := http.NewRequest("POST", fmt.Sprintf("%v/control/apps", c.Service), post_data)
	check(err)
	req.Header.Add("Content-Type", "application/json")
	c.performRequestIntoJSON(result, req)
}

func (c Connection) DeleteApps(result *map[string]interface{}) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%v/control/apps", c.Service), nil)
	check(err)
	c.performRequestIntoJSON(result, req)
}

func (c Connection) GetApp(result *App, appid string) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%v/control/apps/%v", c.Service, appid), nil)
	check(err)
	buffer, err := c.performRequestIntoBuffer(req)
	err = json.Unmarshal(buffer, result)
	check(err)
}

func (c Connection) DeleteApp(result *map[string]interface{}, appid string) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%v/control/apps/%v", c.Service, appid), nil)
	check(err)
	err = c.performRequestIntoJSON(result, req)
	check(err)
}

func (c Connection) DeleteAppVersion(result *map[string]interface{}, appid string, versionid string) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%v/control/apps/%v/versions/%v", c.Service, appid, versionid), nil)
	check(err)
	c.performRequestIntoJSON(result, req)
}

func (c Connection) CreateAppVersion(result *map[string]interface{}, appid string, version []byte) {
	post_data := bytes.NewReader(version)
	req, err := http.NewRequest("POST", fmt.Sprintf("%v/control/apps/%v/versions", c.Service, appid), post_data)
	check(err)
	c.performRequestIntoJSON(result, req)
}

func (c Connection) StartAppVersion(result *map[string]interface{}, appid string, versionid string) {
	data := url.Values{}
	data.Set("command", "start")
	req, err := http.NewRequest("POST", fmt.Sprintf("%v/control/apps/%v/versions/%v", c.Service, appid, versionid),
		bytes.NewBufferString(data.Encode()))
	check(err)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	c.performRequestIntoJSON(result, req)
}

func (c Connection) StopAppVersion(result *map[string]interface{}, appid string, versionid string) {
	data := url.Values{}
	data.Set("command", "stop")
	req, err := http.NewRequest("POST", fmt.Sprintf("%v/control/apps/%v/versions/%v", c.Service, appid, versionid),
		bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	check(err)
	c.performRequestIntoJSON(result, req)
}

func (c Connection) StartApp(result *map[string]interface{}, appid string) {
	data := url.Values{}
	data.Set("command", "start")
	req, err := http.NewRequest("POST", fmt.Sprintf("%v/control/apps/%v", c.Service, appid),
		bytes.NewBufferString(data.Encode()))
	check(err)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	c.performRequestIntoJSON(result, req)
}

func (c Connection) StopApp(result *map[string]interface{}, appid string) {
	data := url.Values{}
	data.Set("command", "stop")
	req, err := http.NewRequest("POST", fmt.Sprintf("%v/control/apps/%v", c.Service, appid),
		bytes.NewBufferString(data.Encode()))
	check(err)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	c.performRequestIntoJSON(result, req)
}
