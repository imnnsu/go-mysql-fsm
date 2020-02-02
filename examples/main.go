package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/go-ini/ini"
	"github.com/minsunchina/go-mysql-fsm/fsm"
)

func getURI() (string, error) {
	conf, err := ini.LoadSources(ini.LoadOptions{AllowBooleanKeys: true}, os.Getenv("HOME")+"/.my.cnf")
	if err != nil {
		return "", err
	}

	for _, s := range conf.Sections() {
		if s.Key("user").String() == "root" {
			return fmt.Sprintf("root:%s@tcp(127.0.0.1:3306)/fsm", s.Key("password")), nil
		}
	}

	return "", errors.New("password for root not found")
}

func transition(fsm *fsm.FSM, event string) {
	src, err := fsm.Current()
	if err != nil {
		panic(err)
	}

	fsm.Event(event)

	tar, err := fsm.Current()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%-12v + %-10v -> %-12v\n", src, "["+event+"]", tar)
}

func main() {

	uri, err := getURI()
	if err != nil {
		panic(err)
	}

	events := []fsm.Event{
		{Name: "NotReady", Src: []string{"Running"}, Dst: "Error"},
		{Name: "Ready", Src: []string{"Initializing", "Error"}, Dst: "Running"},
		{Name: "Stop", Src: []string{"Initializing", "Running", "Error"}, Dst: "Stopped"},
		{Name: "Delete", Src: []string{"Stopped"}, Dst: "Deleted"},
	}

	f, err := fsm.NewFSM(
		&fsm.DataSourceConfig{
			URI:   uri,
			Table: "task",
			ID:    "1",
			Field: "state",
		}, "Initializing", events,
	)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	f.Init()
	transition(f, "Ready")
	transition(f, "NotReady")
	transition(f, "Ready")
	transition(f, "Stop")
	transition(f, "Delete")
}
