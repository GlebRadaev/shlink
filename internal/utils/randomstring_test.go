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

func TestIsValidID(t *testing.T) {
	type args struct {
		id     string
		length int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Valid ID of length 10",
			args: args{id: Generate(10), length: 10},
			want: true,
		},
		{
			name: "Valid ID of length 5",
			args: args{id: Generate(5), length: 5},
			want: true,
		},
		{
			name: "Invalid ID: incorrect length",
			args: args{id: Generate(5), length: 10},
			want: false,
		},
		{
			name: "Invalid ID: contains special characters",
			args: args{id: "invalid@ID!", length: 10},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidID(tt.args.id, tt.args.length); got != tt.want {
				t.Errorf("IsValidID() = %v, want %v", got, tt.want)
			}
		})
	}
}
