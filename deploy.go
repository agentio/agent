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
	if err != nil {
		return 
	}
	buffer, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return 
	}
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
	if err != nil {
		return err
	}
	err = json.Unmarshal(buffer, result)
	return err
}

func (c Connection) GetApps(result *[]App) (err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%v/control/apps", c.Service), nil)
	if err != nil {
		return err
	}
	buffer, err := c.performRequestIntoBuffer(req)
	if err != nil {
		return err
	}
	return json.Unmarshal(buffer, result)
}

func (c Connection) CreateApp(result *map[string]interface{}, app App) (err error) {
	jsonData, err := json.Marshal(app)
	post_data := bytes.NewReader(jsonData)
	req, err := http.NewRequest("POST", fmt.Sprintf("%v/control/apps", c.Service), post_data)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	return c.performRequestIntoJSON(result, req)
}

func (c Connection) DeleteApps(result *map[string]interface{}) (err error) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%v/control/apps", c.Service), nil)
	if err != nil {
		return err
	}
	return c.performRequestIntoJSON(result, req)
}

func (c Connection) GetApp(result *App, appid string) (err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%v/control/apps/%v", c.Service, appid), nil)
	if err != nil {
		return err
	}
	buffer, err := c.performRequestIntoBuffer(req)
	if err != nil {
		return err
	}
	return json.Unmarshal(buffer, result)
}

func (c Connection) DeleteApp(result *map[string]interface{}, appid string) (err error) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%v/control/apps/%v", c.Service, appid), nil)
	if err != nil {
		return err
	}
	return c.performRequestIntoJSON(result, req)
}

func (c Connection) DeleteAppVersion(result *map[string]interface{}, appid string, versionid string) (err error) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%v/control/apps/%v/versions/%v", c.Service, appid, versionid), nil)
	if err != nil {
		return err
	}
	return c.performRequestIntoJSON(result, req)
}

func (c Connection) CreateAppVersion(result *map[string]interface{}, appid string, version []byte) (err error) {
	post_data := bytes.NewReader(version)
	req, err := http.NewRequest("POST", fmt.Sprintf("%v/control/apps/%v/versions", c.Service, appid), post_data)
	if err != nil {
		return err
	}
	return c.performRequestIntoJSON(result, req)
}

func (c Connection) StartAppVersion(result *map[string]interface{}, appid string, versionid string) (err error) {
	data := url.Values{}
	data.Set("command", "start")
	req, err := http.NewRequest("POST", fmt.Sprintf("%v/control/apps/%v/versions/%v", c.Service, appid, versionid),
		bytes.NewBufferString(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return c.performRequestIntoJSON(result, req)
}

func (c Connection) StopAppVersion(result *map[string]interface{}, appid string, versionid string) (err error) {
	data := url.Values{}
	data.Set("command", "stop")
	req, err := http.NewRequest("POST", fmt.Sprintf("%v/control/apps/%v/versions/%v", c.Service, appid, versionid),
		bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		return err
	}
	return c.performRequestIntoJSON(result, req)
}

func (c Connection) StartApp(result *map[string]interface{}, appid string) (err error) {
	data := url.Values{}
	data.Set("command", "start")
	req, err := http.NewRequest("POST", fmt.Sprintf("%v/control/apps/%v", c.Service, appid),
		bytes.NewBufferString(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return c.performRequestIntoJSON(result, req)
}

func (c Connection) StopApp(result *map[string]interface{}, appid string) (err error) {
	data := url.Values{}
	data.Set("command", "stop")
	req, err := http.NewRequest("POST", fmt.Sprintf("%v/control/apps/%v", c.Service, appid),
		bytes.NewBufferString(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return c.performRequestIntoJSON(result, req)
}
