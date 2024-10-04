package utils

import "testing"

func TestGenerate(t *testing.T) {
	type args struct {
		length int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Generate 10",
			args: args{length: 10},
		},
		{
			name: "Generate 5",
			args: args{length: 5},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Generate(tt.args.length); len(got) != tt.args.length {
				t.Errorf("Generate() length = %v, want %v", len(got), tt.args.length)
			}
		})
	}
}
