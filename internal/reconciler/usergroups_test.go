package reconciler

import (
	"reflect"
	"testing"
)

func Test_contains(t *testing.T) {
	type args struct {
		list []string
		item string
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "example found",
			args: args{
				list: []string{"A", "B", "C"},
				item: "A",
			},
			want: true,
		},
		{
			name: "example not found",
			args: args{
				list: []string{"A", "B", "C"},
				item: "D",
			},
			want: false,
		},
		{
			name: "empty list",
			args: args{
				list: []string{},
				item: "D",
			},
			want: false,
		},
		{
			name: "nil list",
			args: args{
				item: "D",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := contains(tt.args.list, tt.args.item); got != tt.want {
				t.Errorf("contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_equal(t *testing.T) {
	type args struct {
		a []string
		b []string
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "equal empty slices",
			args: args{
				a: []string{},
				b: []string{},
			},
			want: true,
		},
		{
			name: "equal slices same order",
			args: args{
				a: []string{"A", "B", "C", "D"},
				b: []string{"A", "B", "C", "D"},
			},
			want: true,
		},
		{
			name: "equal slices different order 1",
			args: args{
				a: []string{"A", "B", "C", "D"},
				b: []string{"C", "B", "D", "A"},
			},
			want: true,
		},
		{
			name: "equal slices different order 2",
			args: args{
				a: []string{"D", "B", "C", "A"},
				b: []string{"C", "B", "D", "A"},
			},
			want: true,
		},
		{
			name: "not equal slices same length 1",
			args: args{
				a: []string{"A", "B", "C", "D"},
				b: []string{"Z", "B", "D", "A"},
			},
			want: false,
		},
		{
			name: "not equal slices same length 1",
			args: args{
				a: []string{"A", "B", "C", "Z"},
				b: []string{"C", "B", "D", "A"},
			},
			want: false,
		},
		{
			name: "not equal slices different length 1",
			args: args{
				a: []string{"A", "B", "C", "D"},
				b: []string{"A", "B", "C"},
			},
			want: false,
		},
		{
			name: "not equal slices different length 1",
			args: args{
				a: []string{"A", "B", "C"},
				b: []string{"A", "B", "C", "D"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := equal(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("equal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_remove(t *testing.T) {
	type args struct {
		list []string
		item string
	}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "success remove item",
			args: args{
				list: []string{"A", "B", "C", "D"},
				item: "C",
			},
			want: []string{"A", "B", "D"},
		},
		{
			name: "success remove last item",
			args: args{
				list: []string{"A"},
				item: "A",
			},
			want: []string{},
		},
		{
			name: "success remove empty",
			args: args{
				list: []string{"A", "B", "C", "D"},
				item: "",
			},
			want: []string{"A", "B", "C", "D"},
		},
		{
			name: "success remove empty from empty",
			args: args{
				list: []string{},
				item: "",
			},
			want: []string{},
		},
		{
			name: "success remove something from empty",
			args: args{
				list: []string{},
				item: "A",
			},
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := remove(tt.args.list, tt.args.item); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("remove() = %v, want %v", got, tt.want)
			}
		})
	}
}
