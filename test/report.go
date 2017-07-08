package test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

type Report struct {
	Time     string `json:"time"`
	Sent     int    `json:"sent"`
	Received int    `json:"received"`
}

func (r Report) String() string {
	return fmt.Sprintf("%s, Sent: %d, Received: %d", r.Time, r.Sent, r.Received)
}

func WriteReport(file string, reports []*Report) error {
	json, err := json.Marshal(reports)
	if err != nil {
		return err
	}

	dir := "report"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, 0755); err != nil {
			return err
		}
	}

	return ioutil.WriteFile(path.Join(dir, file), json, 0644)
}
