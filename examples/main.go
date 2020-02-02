package main

import (
	"errors"
	"fmt"
	"os"
	"time"

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

	return "", errors.New("password for root is not found")
}

func transition(fsm *fsm.FSM, event, routineID string) {
	state, err := fsm.Current()
	if err != nil {
		panic(err)
	}
	fmt.Printf("#%v | %-12v |\n", routineID, state)

	fmt.Printf("#%v |              | -> %v\n", routineID, event)
	fsm.Event(event)
	fmt.Printf("#%v |              | <- %v\n", routineID, event)

	state, err = fsm.Current()
	if err != nil {
		panic(err)
	}
	fmt.Printf("#%v | %-12v |\n", routineID, state)
}

func updateStateRoutine(routineID string, f *fsm.FSM, events []string) {
	for _, event := range events {
		transition(f, event, routineID)
	}
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
	go updateStateRoutine("1", f, []string{"Ready"})
	go updateStateRoutine("2", f, []string{"Ready", "NotReady"})
	go updateStateRoutine("3", f, []string{"Stop", "Delete"})

	time.Sleep(1 * time.Second)
}
