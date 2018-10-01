package cmdopts

import (
	"github.com/sanity-io/litter"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateOptions(t *testing.T) {
	type Embedded struct {
		Field string `param:"field"`
	}

	tests := []struct {
		name    string
		src     interface{}
		want    []string
		wantErr bool
	}{
		{
			name: "map",
			src: map[string]int{
				"1": 1,
				"2": 15,
			},
			want:    []string{"--1", "1", "--2", "15"},
			wantErr: false,
		},
		{
			name:    "map-error",
			src:     map[int]string{},
			want:    nil,
			wantErr: true,
		},
		{
			name: "simple-struct",
			src: struct {
				A int    `param:"a"`
				B string `param:"b"`
			}{
				A: 15,
				B: "ab ba",
			},
			want:    []string{"--a", "15", "--b", "ab ba"},
			wantErr: false,
		},
		{
			name: "struct-embedded",
			src: struct {
				Embedded
				V string `param:"v"`
			}{
				Embedded: Embedded{
					Field: "cda",
				},
				V: "1!",
			},
			want:    []string{"--field", "cda", "--v", "1!"},
			wantErr: false,
		},
		{
			name: "struct-error",
			src: struct {
				A int
			}{
				A: 12,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Options(tt.src)
			if tt.wantErr {
				if err == nil {
					t.Errorf("error expected on input\n%s", litter.Sdump(tt.src))
				}
				return
			}
			if err != nil {
				t.Error(err)
				return
			}
			require.Equal(t, tt.want, got)
		})
	}
}
