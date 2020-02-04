package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/go-ini/ini"
	_ "github.com/go-sql-driver/mysql"
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

func transition(routineID string, fsm *fsm.FSM, entryID, event string) {
	state, err := fsm.Current(entryID)
	if err != nil {
		panic(err)
	}
	fmt.Printf("#%v | %-12v |\n", routineID, state)

	fmt.Printf("#%v |              | -> %v\n", routineID, event)
	fsm.Event(entryID, event)
	fmt.Printf("#%v |              | <- %v\n", routineID, event)

	state, err = fsm.Current(entryID)
	if err != nil {
		panic(err)
	}
	fmt.Printf("#%v | %-12v |\n", routineID, state)
}

func updateStateRoutine(routineID string, f *fsm.FSM, entryID string, events []string) {
	for _, event := range events {
		transition(routineID, f, entryID, event)
	}
}

func main() {

	uri, err := getURI()
	if err != nil {
		panic(err)
	}

	db, err := sql.Open("mysql", uri)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	events := []fsm.Event{
		{Name: "NotReady", Src: []string{"Running"}, Dst: "Error"},
		{Name: "Ready", Src: []string{"Initializing", "Error"}, Dst: "Running"},
		{Name: "Stop", Src: []string{"Initializing", "Running", "Error"}, Dst: "Stopped"},
		{Name: "Delete", Src: []string{"Stopped"}, Dst: "Deleted"},
	}

	f := fsm.NewFSM(db, "task", "state", "Initializing", events, false)

	entryID := "1"
	f.Initialize(entryID)
	go updateStateRoutine("1", f, entryID, []string{"Ready"})
	go updateStateRoutine("2", f, entryID, []string{"Ready", "NotReady"})
	go updateStateRoutine("3", f, entryID, []string{"Stop", "Delete"})

	time.Sleep(1 * time.Second)
}
