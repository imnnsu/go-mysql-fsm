# Update the FSM States in MySQL

## Files Structure

```text
    .
    |-- events
    |   |-- delete.dml
    |   |-- init.dml
    |   |-- not_ready.dml
    |   |-- ready.dml
    |   `-- stop.dml
    |-- task.ddl
    `-- user.dcl
```

## Usage of Scripts

The scripts under `mysql` provides an example of how to represent an FSM in MySQL and how to update the states with MySQL statements.

- `task.ddl` creates a database and the table.
- `user.dcl` grants authorities to a specific user.
- `events` holds all the *.dml files, containing statements that updates the FSM states, corresponding to each FSM event.
