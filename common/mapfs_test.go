package common

import (
	"fmt"
	"testing"
)

var unsort = []string{
	"public/spa/web_modules/svelte/index.mjs",
	"public/spa/web_modules/svelte/internal/index.mjs",
	"public/spa/web_modules/svelte/transition/index.mjs",
	"public/spa/web_modules/svelte/easing/index.mjs",
	"public/spa/web_modules/svelte/store/index.mjs",
	"public/spa/web_modules/regexparam/dist/regexparam.mjs",
	"public/spa/web_modules/svelte/animate/index.mjs",
	"public/spa/web_modules/navaid/dist/navaid.mjs",
	"public/spa/web_modules/svelte/compiler.mjs",
}

var sorted = []string{
	"public/spa/web_modules/navaid/dist/navaid.mjs",
	"public/spa/web_modules/regexparam/dist/regexparam.mjs",
	"public/spa/web_modules/svelte/animate/index.mjs",
	"public/spa/web_modules/svelte/compiler.mjs",
	"public/spa/web_modules/svelte/easing/index.mjs",
	"public/spa/web_modules/svelte/index.mjs",
	"public/spa/web_modules/svelte/internal/index.mjs",
	"public/spa/web_modules/svelte/store/index.mjs",
	"public/spa/web_modules/svelte/transition/index.mjs",
}

func Test_binSearchIndex(t *testing.T) {
	type args struct {
		x string
		a []string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := binSearchIndex(tt.args.x, tt.args.a); got != tt.want {
				t.Errorf("binSearchIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_setEntry(t *testing.T) {
	type args struct {
		x       string
		entries []string
		ind     int
	}
	tests := []struct {
		name string
		args args
	}{

		{
			name: "four",
			args: args{
				x:       "e",
				ind:     4,
				entries: []string{"a", "b", "c", "d", "f", "g", "h"}},
		},
		{
			name: "seven",
			args: args{
				x:       "i",
				ind:     7,
				entries: []string{"a", "b", "c", "d", "f", "g", "h"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setEntry(tt.args.x, &tt.args.entries)
			if tt.args.entries[tt.args.ind] != tt.args.x {
				t.Errorf("setEntry() = %v, want %v", tt.args.entries[tt.args.ind], tt.args.x)
			}
		})
	}
}

func Test_getEntryIndex(t *testing.T) {
	type args struct {
		x       string
		entries *[]string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "zero",
			args: args{x: "public/spa/web_modules/abc/one.svelte", entries: &sorted},
			want: 0,
		},
		{
			name: "two",
			args: args{x: "public/spa/web_modules/svelte/animate/index.mjs", entries: &sorted},
			want: 2,
		},

		{
			name: "minus1" + fmt.Sprint(len(sorted)-1),
			args: args{x: "public/spa/web_modules/z/theend.svelte", entries: &sorted},
			want: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getEntryIndex(tt.args.x, tt.args.entries); got != tt.want {
				t.Errorf("getEntryIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_deleteEntry(t *testing.T) {

	entries := []string{"bar", "foo"}
	t.Run("deleteBar", func(t *testing.T) {
		deleteEntry("bar", &entries)

		if len(entries) != 1 || (entries[0]) != "foo" {
			t.Errorf("deleteEntry() = %v, want %v", entries, "[foo]")
		}
	})

}

func Test_searchPath(t *testing.T) {
	type args struct {
		path    string
		entries []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := searchPath(tt.args.path, &tt.args.entries); got != tt.want {
				t.Errorf("searchPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func insert(s ...string) *[]string {
	out := &[]string{}
	for _, s := range unsort {
		setEntry(s, out)
	}
	return out

}
func Test_isOrdered(t *testing.T) {

	got := insert(unsort...)

	for i, e := range *got {
		if sorted[i] != e {
			t.Errorf("insert() = %v, want %v", e, sorted[i])
			return
		}
	}

}
