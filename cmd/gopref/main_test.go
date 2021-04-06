package main

import (
	"reflect"
	"testing"
)

func Test_predictionJob(t *testing.T) {
	tests := []struct {
		name    string
		prefix  string
		want    []string
		wantErr bool
	}{
		{
			name:    "errors",
			prefix:  "errors:",
			want:    []string{"errors:UCS-COMMON", "errors:home"},
			wantErr: false,
		},
		{
			name:    "github.com/s",
			prefix:  "github.com/sir",
			want:    []string{"github.com/sirkon/", "github.com/sirupsen/"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := predictionJob(tt.prefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("predictionJob() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("predictionJob() got = %v, want %v", got, tt.want)
			}
		})
	}
}
