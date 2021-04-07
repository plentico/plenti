package common

import (
	"fmt"
	"math/rand"
	"testing"
)

func unsortSli() []string {
	out := []string{}
	for _, s := range sorted {
		out = append(out, s)
	}
	rand.Shuffle(len(out), func(i, j int) {
		out[i], out[j] = out[j], out[i]
	})
	return out
}

var sorted = []string{
	"public/spa/web_modules/navaid/dist/navaid.js",
	"public/spa/web_modules/navaid/dist/navaid.min.js",
	"public/spa/web_modules/navaid/dist/navaid.mjs",
	"public/spa/web_modules/regexparam/dist/regexparam.js",
	"public/spa/web_modules/regexparam/dist/regexparam.min.js",
	"public/spa/web_modules/regexparam/dist/regexparam.mjs",
	"public/spa/web_modules/svelte/compiler.js",
	"public/spa/web_modules/svelte/compiler.mjs",
	"public/spa/web_modules/svelte/index.js",
	"public/spa/web_modules/svelte/index.mjs",
	"public/spa/web_modules/svelte/register.js",
	"public/spa/web_modules/svelte/animate/index.js",
	"public/spa/web_modules/svelte/animate/index.mjs",
	"public/spa/web_modules/svelte/easing/index.js",
	"public/spa/web_modules/svelte/easing/index.mjs",
	"public/spa/web_modules/svelte/internal/index.js",
	"public/spa/web_modules/svelte/internal/index.mjs",
	"public/spa/web_modules/svelte/motion/index.js",
	"public/spa/web_modules/svelte/motion/index.mjs",
	"public/spa/web_modules/svelte/store/index.js",
	"public/spa/web_modules/svelte/store/index.mjs",
	"public/spa/web_modules/svelte/transition/index.js",
	"public/spa/web_modules/svelte/transition/index.mjs",
}

func insert() *[]string {
	out := &[]string{}
	for _, s := range unsortSli() {
		setEntry(s, out)
	}
	return out

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
				x:       "e.txt",
				ind:     4,
				entries: []string{"a.txt", "b.txt", "c.txt", "d.txt", "f.txt", "g.txt", "h.txt"}},
		},
		{
			name: "seven",
			args: args{
				x:       "i.txt",
				ind:     7,
				entries: []string{"a.txt", "b.txt", "c.txt", "d.txt", "f.txt", "g.txt", "h.txt"}},
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
			args: args{x: "public/spa/web_modules/aaaa.svelte", entries: &sorted},
			want: 0,
		},
		{
			name: "aDir",
			args: args{x: "public/spa/web_modules/svelte/animate", entries: &sorted},
			want: 11,
		},

		{
			name: "five",
			args: args{x: "public/spa/web_modules/regexparam/dist/regexparam.mjs", entries: &sorted},
			want: 5,
		},

		{
			name: "minus1" + fmt.Sprint(len(sorted)-1),
			args: args{x: "public/spa/web_modules/z/z/z/theend.svelte", entries: &sorted},
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

	entries := []string{"bar.txt", "foo.txt"}
	t.Run("deleteBar", func(t *testing.T) {
		deleteEntry("bar.txt", &entries)

		if len(entries) != 1 || (entries[0]) != "foo.txt" {
			t.Errorf("deleteEntry() = %v, want %v", entries, "[foo.txt]")
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

func Test_isOrdered(t *testing.T) {

	got := insert()

	for i, e := range *got {
		if sorted[i] != e {
			t.Errorf("insert() = %v, want %v", e, sorted[i])
			return
		}
	}

}
