package fsm

import (
	"testing"
)

func Test_eventQuery(t *testing.T) {
	type args struct {
		fsm   *FSM
		event string
	}
	tests := []struct {
		name      string
		args      args
		wantQuery string
	}{
		{
			name: "Ready",
			args: args{
				fsm: &FSM{
					Config: &Config{
						DBConfig: &DBConfig{
							Table: "task",
							Field: "state",
						},
						Init: "Initializing",
						Events: map[string]Event{
							"Ready": {
								Name: "Ready",
								Src:  []string{"Initializing", "Error"},
								Dst:  "Running",
							},
						},
					},
					ID: "1",
				},
				event: "Ready",
			},
			wantQuery: "INSERT INTO task (id, state) VALUES ('1', 'Initializing') " +
				"ON DUPLICATE KEY UPDATE state = CASE " +
				"WHEN state = 'Initializing' THEN 'Running' " +
				"WHEN state = 'Error' THEN 'Running' " +
				"ELSE state " +
				"END",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotQuery := eventQuery(tt.args.fsm, tt.args.event); gotQuery != tt.wantQuery {
				t.Errorf("eventQuery() = %v, want %v", gotQuery, tt.wantQuery)
			}
		})
	}
}

func Test_currentQuery(t *testing.T) {
	type args struct {
		fsm *FSM
	}
	tests := []struct {
		name      string
		args      args
		wantQuery string
	}{
		{
			name: "1",
			args: args{
				fsm: &FSM{
					Config: &Config{
						DBConfig: &DBConfig{
							Table: "task",
							Field: "state",
						},
					},
					ID: "1",
				},
			},
			wantQuery: "SELECT state FROM task WHERE id = '1'",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotQuery := currentQuery(tt.args.fsm); gotQuery != tt.wantQuery {
				t.Errorf("currentQuery() = %v, want %v", gotQuery, tt.wantQuery)
			}
		})
	}
}

func Test_initQuery(t *testing.T) {
	type args struct {
		fsm *FSM
	}
	tests := []struct {
		name      string
		args      args
		wantQuery string
	}{
		{
			name: "1",
			args: args{
				fsm: &FSM{
					Config: &Config{
						DBConfig: &DBConfig{
							Table: "task",
							Field: "state",
						},
						Init: "Initializing",
					},
					ID: "1",
				},
			},
			wantQuery: "INSERT INTO task (id, state) VALUES ('1', 'Initializing')",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotQuery := initQuery(tt.args.fsm); gotQuery != tt.wantQuery {
				t.Errorf("initQuery() = %v, want %v", gotQuery, tt.wantQuery)
			}
		})
	}
}
