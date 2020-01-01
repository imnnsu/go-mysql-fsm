// Package hsfsm implements a horizontally scalable finite state machine
// utilizing MySQL.
//
// Traditional finite state machines manage the states in memory. Package hsfsm
// updates and gets the states directly on the MySQL side, so that it avoids
// the in-memory problem. Thus it is able to be scaled horizontally with no
// difficulty.
package hsfsm

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// Event defines the name of the event, with its corresponding state
// transitions.
type Event struct {
	// Name is the name of the event.
	Name string

	// Src is a slice containing all the possible source states.
	Src []string

	// Dst is a destination state that source states transfer to when the event
	// happens.
	Dst string
}

// DataSourceConfig is a configuration used for MySQL setups.
type DataSourceConfig struct {
	// URI defines the MySQL address to connect.
	URI string

	// Table is the name of the table to be updated.
	Table string

	// ID is the identifier of the row to be updated.
	ID string

	// Field refers to the field in the row to be upated.
	Field string

	// Debug indicates whether debugging logs should be printed or not.
	Debug bool
}

// FSM defines a horizontally scalable finate state machine.
type FSM struct {
	db     *sql.DB
	table  string
	id     string
	field  string
	debug  bool
	init   string
	events map[string]Event
}

// NewFSM returns a new FSM from the configuration.
func NewFSM(dataSourceConfig *DataSourceConfig, init string, events []Event) (
	*FSM, error) {

	db, err := sql.Open("mysql", dataSourceConfig.URI)
	if err != nil {
		return nil, err
	}

	eventMap := make(map[string]Event)
	for _, event := range events {
		eventMap[event.Name] = event
	}

	return &FSM{
		db:     db,
		table:  dataSourceConfig.Table,
		id:     dataSourceConfig.ID,
		field:  dataSourceConfig.Field,
		debug:  dataSourceConfig.Debug,
		init:   init,
		events: eventMap,
	}, nil
}

// Event updates the state machine according to events defined.
//
// A new record is inserted into the database if the ID doesn't exist.
// Otherwise the row identified by the ID is updated.
func (fsm *FSM) Event(event string) error {

	if _, ok := fsm.events[event]; !ok {
		return errors.New("undefined event: " + event)
	}

	_, err := fsm.db.Exec(eventQuery(fsm, event))
	if err != nil {
		return err
	}

	return nil
}

// Current returns the current state of the state machine.
func (fsm *FSM) Current() (string, error) {

	var state string
	err := fsm.db.QueryRow(currentQuery(fsm)).Scan(&state)
	if err != nil {
		return "", err
	}

	return state, nil
}

func eventQuery(fsm *FSM, event string) (query string) {

	insertFormat := "INSERT INTO %s (id, %s) VALUES ('%s', '%s') "
	updateFormat := "ON DUPLICATE KEY UPDATE %s = CASE %s ELSE %s END"
	caseFormat := "WHEN state = '%s' THEN '%s'"

	dst := fsm.events[event].Dst
	var cases []string
	for _, src := range fsm.events[event].Src {
		cases = append(cases, fmt.Sprintf(caseFormat, src, dst))
	}
	query = fmt.Sprintf(insertFormat, fsm.table, fsm.field, fsm.id, fsm.init) +
		fmt.Sprintf(updateFormat, fsm.field, strings.Join(cases, " "),
			fsm.field)
	if fsm.debug {
		log.Println("[eventQuery]", query)
	}
	return
}

func currentQuery(fsm *FSM) (query string) {

	selectFormat := "SELECT %s FROM %s WHERE id = '%s'"

	query = fmt.Sprintf(selectFormat, fsm.field, fsm.table, fsm.id)
	if fsm.debug {
		log.Println("[currentQuery]", query)
	}
	return
}
