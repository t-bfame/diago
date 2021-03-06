package dashboard

import (
	"errors"
	"bytes"
	"io/ioutil"
	"encoding/json"
	"net/http"

	c "github.com/t-bfame/diago/config"

	log "github.com/sirupsen/logrus"
)

// Represents fields to form ipanels for embedding the grafana dashboard
type Dashboard struct {
	BaseUrl string `json:"url"`
	Panels []struct {
		Id int `json:"id"`
		Title string `json:"title"`
	} `json:"panels"`
	VarNames []string `json:"fields"`
}

// Represents the configuration of a Grafana dashboard
type DashConf struct {
	Uid string `json:"uid"`
	Panels []struct {
		Id int `json:"id"`
		Title string `json:"title"`
	} `json:"panels"`
	Templating struct {
		List []struct {
			Name string `json:"name"`
		}
	}
}

// Represents fields for the json payload to create the dashboard
type DashCreate struct {
	Dashboard interface{} `json:"dashboard"`
	FolderId int `json:"folderId"`
	Message string `json:"message"`
	Overwrite bool `json:"overwrite"`
}

// returns the json version of Dashboard struct
func (d *Dashboard) ToJSON() ([]byte, error) {
	return json.Marshal(*d)
}

// Checks whether a dashboard with the given UID exsists in Grafaana
func checkDashBoard(uid string) *DashConf {
	req, err := http.NewRequest("GET", c.Diago.GrafanaBasePath + "/api/dashboards/uid/" + uid, nil)

	req.Header.Set("Authorization", "Bearer " + c.Diago.GrafanaAPIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if (err != nil) || (resp.StatusCode < 200 || resp.StatusCode > 299) {
		log.WithField("uid", uid).Error("Unable to find grafana dashboard with uid");
		return nil
	}

	body, _ := ioutil.ReadAll(resp.Body)
	var config struct {
		Dashboard DashConf `json:"dashboard"`
	}

	json.Unmarshal([]byte(body), &config)

	return &config.Dashboard
}

// Creates a new Dashboard with the provided configurations
func createDashboard(body []byte) error {
	var dashboard interface{}
	json.Unmarshal(body, &dashboard)

	d := DashCreate{
		dashboard,
		0,
		"Diago api call: Creating Diago Dashboard",
		true,
	}

	blob, _ := json.Marshal(d)
	buff := bytes.NewBuffer(blob)

	req, err := http.NewRequest("POST", c.Diago.GrafanaBasePath + "/api/dashboards/db", buff)
	req.Header.Set("Authorization", "Bearer " + c.Diago.GrafanaAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if (err != nil) || (resp.StatusCode < 200 || resp.StatusCode > 299) {
		log.Error("Unable to create grafana dashboard");
		return errors.New("Unable to create grafana dashboard")
	}

	return nil
}

// Creates a new Grafana dashboard config
// If the diago grafana dashboard doesnt exist, then a new one is created
// using the Grafana API, esle the existing onces configurations are fetched
func NewDashboard() (*Dashboard, error) {
	js := []byte(c.Diago.GrafanaDashboardConfig)

	conf := new(DashConf)
	json.Unmarshal(js, conf)

	uid := conf.Uid
	conf = checkDashBoard(uid)
	if conf == nil {
		err := createDashboard(js)
		if err != nil {
			return nil, err
		}
		conf = checkDashBoard(uid)
	}

	var varArray []string
	for _, s := range conf.Templating.List {
		varArray = append(varArray,"var-" + s.Name)
	}

	return &Dashboard{
		c.Diago.GrafanaBasePath + "/d-solo/" + conf.Uid,
		conf.Panels,
		varArray,
	}, nil
}
