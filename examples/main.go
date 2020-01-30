package main

import (
	"fmt"
	"os"

	"github.com/go-ini/ini"
	"github.com/minsunchina/hsfsm"
)

func main() {

	events := []hsfsm.Event{
		{Name: "NotReady", Src: []string{"Running"}, Dst: "Error"},
		{Name: "Ready", Src: []string{"Initializing", "Error"}, Dst: "Running"},
		{Name: "Stop", Src: []string{"Initializing", "Running", "Error"}, Dst: "Stopped"},
		{Name: "Delete", Src: []string{"Stopped"}, Dst: "Deleted"},
	}

	conf, err := ini.LoadSources(ini.LoadOptions{AllowBooleanKeys: true}, os.Getenv("HOME")+"/.my.cnf")
	if err != nil {
		fmt.Println(err)
		return
	}

	var uri string
	for _, s := range conf.Sections() {
		fmt.Println("section", s.Name())
		if s.Key("user").String() == "root" {
			uri = fmt.Sprintf("root:%s@tcp(127.0.0.1:3306)/fsm", s.Key("password"))
		}
	}

	fsm, err := hsfsm.NewFSM(
		&hsfsm.DataSourceConfig{
			URI:   uri,
			Table: "task",
			ID:    "2",
			Field: "state",
		},
		"Initializing", events,
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	state, err := fsm.Current()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(state)
	fsm.Event("Ready")
	state, err = fsm.Current()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(state)
}
