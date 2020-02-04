[![Build Status](https://travis-ci.com/minsunchina/go-mysql-fsm.svg?branch=master)](https://travis-ci.com/minsunchina/go-mysql-fsm)
[![Coverage Status](https://coveralls.io/repos/github/minsunchina/go-mysql-fsm/badge.svg)](https://coveralls.io/github/minsunchina/go-mysql-fsm)
[![GoDoc](https://godoc.org/github.com/minsunchina/go-mysql-fsm/fsm?status.svg)](https://godoc.org/github.com/minsunchina/go-mysql-fsm/fsm)
[![Go Report Card](https://goreportcard.com/badge/github.com/minsunchina/go-mysql-fsm)](https://goreportcard.com/report/github.com/minsunchina/go-mysql-fsm)

# Horizontally Scalable FSM storing states in MySQL

The [finite state machine](https://en.wikipedia.org/wiki/Finite-state_machine) is a very classical model. Usually, a finite state machine is manipulated in-memory, so that it requires extra efforts when horizontally scaling. To solve this problem, `go-mysql-fsm/fsm` is developed, storing and updating the FSM states in MySQL. See [docs](docs/README.md) for more information.

## Setup the Database

A MySQL database and a table should be accessible from the environment. For example, we are using the table described below.

```text
mysql> describe fsm.task;
+-------+-------------+------+-----+---------+-------+
| Field | Type        | Null | Key | Default | Extra |
+-------+-------------+------+-----+---------+-------+
| id    | varchar(64) | NO   | PRI | NULL    |       |
| state | varchar(64) | NO   |     |         |       |
+-------+-------------+------+-----+---------+-------+
```

## Usage

Using [go-sql-driver/mysql](github.com/go-sql-driver/mysql) to setup the database connection.

```go
import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
)

db, err := sql.Open("mysql", "user:password@tcp(127.0.0.1:3306)/fsm")
if err != nil {
    panic(err)
}
defer db.Close()
```

> Please see [DSN (Data Source Name)](https://github.com/go-sql-driver/mysql#dsn-data-source-name) for reference of uri.

Create the Finite State Machine.

```go
import "github.com/minsunchina/go-mysql-fsm/fsm"

events := []fsm.Event{
    {Name: "NotReady", Src: []string{"Running"}, Dst: "Error"},
    {Name: "Ready", Src: []string{"Initializing", "Error"}, Dst: "Running"},
    {Name: "Stop", Src: []string{"Initializing", "Running", "Error"}, Dst: "Stopped"},
    {Name: "Delete", Src: []string{"Stopped"}, Dst: "Deleted"},
}

f := fsm.NewFSM(db, "task", "state", "Initializing", events, false)
```

Then we can set or get the states of multiple finite state machines with the same transition rule, identified by different ID's.

```go
f.Initialize("1")
f.Initialize("2")
state, _ := f.Current("1")
state, _ := f.Current("2")
f.Event("1", "Ready")
f.Event("2", "NotReady")
state, _ = f.Current("1")
state, _ = f.Current("2")
```

Please check the [example](examples/main.go) that updates the finite state machine in multiple go-routines.
