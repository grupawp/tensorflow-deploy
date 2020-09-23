package lock

import (
	"testing"

	"github.com/grupawp/tensorflow-deploy/app"
)

func Test_lock(t *testing.T) {
	tx := &Lock{}
	modelID := app.ServableID{Team: "1", Project: "2", Name: "3"}
	if err := tx.Lock(modelID); err != nil {
		t.Errorf("lock error")
		return
	}

	if err := tx.Lock(modelID); err == nil {
		t.Errorf("lock doesn't work!")
	}

	tx.UnLock(modelID)

	if err := tx.Lock(modelID); err != nil {
		t.Errorf("lock error")
		return
	}

}

func Test_lock_Lock(t *testing.T) {
	type fields struct {
	}
	tests := []struct {
		name    string
		fields  fields
		args    app.ServableID
		wantErr bool
	}{
		{
			name:    "test1 - no errors",
			fields:  fields{},
			args:    app.ServableID{Team: "aa", Project: "bb", Name: "cc"},
			wantErr: false,
		},
		{
			name:    "test2 - too short args",
			fields:  fields{},
			args:    app.ServableID{Team: "", Project: "", Name: ""},
			wantErr: true,
		},
		{
			name:    "test3 - too long args",
			fields:  fields{},
			args:    app.ServableID{Team: "1234567890123456789012345678901234567890", Project: "123456789012345678901234567890", Name: "123456789012345678901234567890"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Lock{}
			if err := l.Lock(tt.args); (err != nil) != tt.wantErr {
				t.Errorf("lock.Lock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
