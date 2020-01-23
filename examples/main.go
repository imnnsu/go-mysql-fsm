package main

import (
	"fmt"

	"github.com/minsunchina/hsfsm"
)

func main() {

	events := []hsfsm.Event{
		{Name: "NotReady", Src: []string{"Running"}, Dst: "Error"},
		{Name: "Ready", Src: []string{"Initializing", "Error"}, Dst: "Running"},
		{Name: "Stop", Src: []string{"Initializing", "Running", "Error"}, Dst: "Stopped"},
		{Name: "Delete", Src: []string{"Stopped"}, Dst: "Deleted"},
	}

	fsm, _ := hsfsm.NewFSM(
		&hsfsm.DataSourceConfig{
			URI:   "root@tcp(127.0.0.1:3306)/fsm",
			Table: "task",
			ID:    "2",
			Field: "state",
		},
		"Initializing", events,
	)

	state, _ := fsm.Current()
	fmt.Println(state)
	fsm.Event("Ready")
	state, _ = fsm.Current()
	fmt.Println(state)
}
