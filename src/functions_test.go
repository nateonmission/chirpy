package main

import (
	"testing"
)

func TestCensorChirp(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "clean chirp unchanged",
			in:   "hello world",
			want: "hello world",
		},
		{
			name: "censors dirty word",
			in:   "this is kerfuffle",
			want: "this is ****",
		},
		{
			name: "case insensitive",
			in:   "this is FORNAX",
			want: "this is ****",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := censorChirp(tt.in)
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}