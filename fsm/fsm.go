// Package fsm implements a horizontally scalable finite state machine
// utilizing MySQL.
//
// Traditional finite state machines manage the states in memory. Package fsm
// updates and gets the states directly on the MySQL side, so that it avoids
// the in-memory problem. Thus it is able to be scaled horizontally with no
// difficulty.
package fsm

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
)

// Event holds the name of an event, with its corresponding state transitions.
type Event struct {
	// Name is the name of the event.
	Name string

	// Src is a slice containing all the possible source states.
	Src []string

	// Dst is a destination state that source states transfer to when the event
	// happens.
	Dst string
}

// DBConfig is the configuration of the MySQL database.
type DBConfig struct {
	// DB represents the connection to the database.
	DB *sql.DB
	// Table is the name of the table to be updated.
	Table string
	// Field refers to the field in the row to be upated.
	Field string
	// Debug deicdes whehter the debugging logs should be shown.
	Debug bool
}

// Config is the configuratoin of the finite state machine.
type Config struct {
	*DBConfig
	// Init is the initial state of the finite state machine.
	Init string
	// Events is a map from the event name to its transitions.
	Events map[string]Event
	// Debug indicates whether debugging logs should be printed or not.
}

// FSM is a finite state machine storing states in the MySQL database.
type FSM struct {
	*Config
	// ID is used in the MySQL database as the primary key.
	ID string
}

// NewConfig returns a new configuration for the finite state machine and its
// MySQL configurations.
func NewConfig(db *sql.DB, table, field, init string, events []Event) *Config {

	dbConfig := &DBConfig{
		DB:    db,
		Table: table,
		Field: field,
	}

	eventMap := make(map[string]Event)
	for _, event := range events {
		eventMap[event.Name] = event
	}

	return &Config{
		DBConfig: dbConfig,
		Init:     init,
		Events:   eventMap,
	}
}

// NewFSM returns a new FSM.
func NewFSM(config *Config, id string) *FSM {
	return &FSM{
		Config: config,
		ID:     id,
	}
}

// Initialize inserts into the table with the initial state.
func (fsm *FSM) Initialize() error {

	_, err := fsm.DB.Exec(initQuery(fsm))
	return err
}

// Current returns the current state of the state machine.
func (fsm *FSM) Current() (string, error) {

	var state string
	err := fsm.DB.QueryRow(currentQuery(fsm)).Scan(&state)
	if err != nil {
		return "", err
	}

	return state, nil
}

// Event updates the state machine according to events defined.
//
// A new record is inserted into the database if the ID doesn't exist.
// Otherwise the row identified by the ID is updated.
func (fsm *FSM) Event(event string) error {

	if _, ok := fsm.Events[event]; !ok {
		return errors.New("undefined event: " + event)
	}

	_, err := fsm.DB.Exec(eventQuery(fsm, event))
	if err != nil {
		return err
	}

	return nil
}

func eventQuery(fsm *FSM, event string) (query string) {

	insertFormat := "INSERT INTO %s (id, %s) VALUES ('%s', '%s') "
	updateFormat := "ON DUPLICATE KEY UPDATE %s = CASE %s ELSE %s END"
	caseFormat := "WHEN state = '%s' THEN '%s'"

	dst := fsm.Events[event].Dst
	var cases []string
	for _, src := range fsm.Events[event].Src {
		cases = append(cases, fmt.Sprintf(caseFormat, src, dst))
	}
	query = fmt.Sprintf(insertFormat, fsm.Table, fsm.Field, fsm.ID, fsm.Init) +
		fmt.Sprintf(updateFormat, fsm.Field, strings.Join(cases, " "),
			fsm.Field)
	if fsm.Debug {
		log.Println("[eventQuery]", query)
	}
	return
}

func currentQuery(fsm *FSM) (query string) {

	selectFormat := "SELECT %s FROM %s WHERE id = '%s'"

	query = fmt.Sprintf(selectFormat, fsm.Field, fsm.Table, fsm.ID)
	if fsm.Debug {
		log.Println("[currentQuery]", query)
	}
	return
}

func initQuery(fsm *FSM) (query string) {

	insertFormat := "INSERT INTO %s (id, %s) VALUES ('%s', '%s')"

	query = fmt.Sprintf(insertFormat, fsm.Table, fsm.Field, fsm.ID, fsm.Init)
	if fsm.Debug {
		log.Println("[initQuery]", query)
	}
	return
}
