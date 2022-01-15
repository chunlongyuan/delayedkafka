package xassert

import "testing"

func Test_isEmpty(t *testing.T) {
	tests := []struct {
		name string
		args interface{}
		want bool
	}{
		{args: "", want: true},
		{args: "a", want: false},
		{args: "ab", want: false},
		{args: "abc", want: false},
		{args: "1", want: false},
		{args: "12", want: false},
		{args: "123", want: false},
		{args: 1, want: false},
		{args: map[string]struct{}{}, want: true},
		{args: map[string]struct{}{"a": {}}, want: false},
		{args: []string{"1"}, want: false},
		{args: []string{}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isEmpty(tt.args); got != tt.want {
				t.Errorf("isEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}
