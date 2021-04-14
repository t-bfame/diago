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

type Dashboard struct {
	BaseUrl string `json:"url"`
	Panels []struct {
		Id int `json:"id"`
		Title string `json:"title"`
	} `json:"panels"`
	VarNames []string `json:"fields"`
}

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

type DashCreate struct {
	Dashboard interface{} `json:"dashboard"`
	FolderId int `json:"folderId"`
	Message string `json:"message"`
	Overwrite bool `json:"overwrite"`
}

func (d *Dashboard) ToJSON() ([]byte, error) {
	return json.Marshal(*d)
}

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

func NewDashboard() (*Dashboard, error) {
	js := []byte(c.Diago.GrafanaDashboardConfig)

	conf := new(DashConf)
	json.Unmarshal(js, conf)

	conf = checkDashBoard(conf.Uid)
	if conf == nil {
		err := createDashboard(js)
		if err != nil {
			return nil, err
		}
		conf = checkDashBoard(conf.Uid)
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
