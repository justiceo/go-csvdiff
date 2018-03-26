package csvdiff

import (
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func defaultOptions(cols ...string) *Options {
	return &Options{
		KeyColumns: cols,
		LazyQuotes: true,
		HasHeader:  true,
		Separator:  ',',
	}
}

func TestFromReader(t *testing.T) {
	for _, v := range []struct {
		name    string
		reader  io.Reader
		opt     *Options
		want    csvData
		wantErr bool
	}{
		{
			name:    "nil reader",
			opt:     defaultOptions(),
			wantErr: true,
		},
		{
			name:    "nil options",
			reader:  strings.NewReader("hello world"),
			wantErr: true,
		},
		{
			name:    "nil reader and options",
			wantErr: true,
		},
		{
			name:   "simple csv",
			reader: strings.NewReader("a,b\n1,2"),
			opt:    defaultOptions("a"),
			want: csvData{
				headers: []string{"a", "b"},
				headerMap: map[string]int{
					"a": 0,
					"b": 1,
				},
				records: map[string][]string{
					"1": []string{"1", "2"},
				},
			},
		},
		{
			name:    "valid csv with invalid key column errors",
			reader:  strings.NewReader("a,b\n1,2"),
			opt:     defaultOptions("c"),
			wantErr: true,
		},
		{
			name:   "empty csv",
			reader: strings.NewReader(""),
			opt:    &Options{HasHeader: false, Separator: ','},
			want:   csvData{},
		},
		{
			name:   "empty csv work with HasHeader set to true",
			reader: strings.NewReader(""),
			opt:    &Options{HasHeader: true, Separator: ','},
			want:   csvData{},
		},
		{
			name:   "header only csv is valid",
			reader: strings.NewReader("a,b"),
			opt:    defaultOptions("b"),
			want: csvData{
				headers: []string{"a", "b"},
				headerMap: map[string]int{
					"a": 0,
					"b": 1,
				},
				records: map[string][]string{},
			},
		},
		{
			name:   "custom separator works",
			reader: strings.NewReader("a;b;c"),
			opt:    &Options{KeyColumns: []string{"b", "c"}, HasHeader: true, Separator: ';'},
			want: csvData{
				headers: []string{"b", "c"},
				headerMap: map[string]int{
					"b": 0,
					"c": 1,
				},
				records: map[string][]string{},
			},
		},
		{
			name:    "bad csv errors",
			reader:  strings.NewReader("a;b\";c"),
			opt:     &Options{LazyQuotes: false, Separator: ';'},
			wantErr: true,
		},

		/*{
			name:   "lazy quotes forgives",
			reader: strings.NewReader("a;ba;c"),
			opt:    &Options{KeyColumns: []string{"a"}, LazyQuotes: false, HasHeader: true, Separator: ';'},
			want: csvData{
				headers: []string{"a", "b", "c", "d"},
				headerMap: map[string]int{
					"a": 0,
					"b": 1,
					"c": 3,
				},
				records: map[string][]string{},
			},
			wantErr: true,
		},*/
	} {
		v.want.opt = v.opt
		got, err := fromReader(v.reader, v.opt)

		if err != nil && !v.wantErr {
			t.Errorf("%s: unexpected error: %v", v.name, err)
			continue
		}

		if diff := cmp.Diff(v.want, got, cmpopts.IgnoreUnexported(csvData{})); diff != "" {
			t.Errorf("%s: got: %v, want: %v\n", v.name, got, v.want)
		}
	}
}

func TestCompare(t *testing.T) {
	for _, v := range []struct {
		name     string
		from, to io.Reader
		opt      *Options
		want     DiffRef
		wantErr  bool
	}{
		{
			name: "no diff",
			from: strings.NewReader("a,b\n1,2"),
			to:   strings.NewReader("a,b\n1,2"),
			opt:  defaultOptions("a"),
			want: DiffRef{
				Changed: map[string][]string{},
			},
		},
		{
			name: "row add",
			from: strings.NewReader("a,b\n1,2"),
			to:   strings.NewReader("a,b\n1,2\n3,4"),
			opt:  defaultOptions("a"),
			want: DiffRef{
				AddedRow: []string{"3"},
				Changed:  map[string][]string{},
			},
		},
		{
			name: "row removed",
			from: strings.NewReader("a,b\n1,2"),
			to:   strings.NewReader("a,b\n"),
			opt:  defaultOptions("a"),
			want: DiffRef{
				RemovedRow: []string{"1"},
				Changed:    map[string][]string{},
			},
		},
		{
			name: "row changed",
			from: strings.NewReader("a,b\n1,2"),
			to:   strings.NewReader("a,b\n1,3"),
			opt:  defaultOptions("a"),
			want: DiffRef{
				Changed: map[string][]string{
					"1": []string{"b"},
				},
			},
		},
		{
			name: "rows rearranged",
			from: strings.NewReader("a,b\n1,2\n3,4"),
			to:   strings.NewReader("a,b\n3,4\n1,2"),
			opt:  defaultOptions("a"),
			want: DiffRef{
				Changed: map[string][]string{},
			},
		},
		{
			name: "columns rearranged",
			from: strings.NewReader("a,b\n1,2\n3,4"),
			to:   strings.NewReader("b,a\n2,1\n4,3"),
			opt:  defaultOptions("a"),
			want: DiffRef{
				Changed: map[string][]string{},
			},
		},
		{
			name: "row and column rearranged",
			from: strings.NewReader("a,b\n1,2\n3,4"),
			to:   strings.NewReader("b,a\n4,3\n2,1"),
			opt:  defaultOptions("a"),
			want: DiffRef{
				Changed: map[string][]string{},
			},
		},
		{
			name: "column added",
			from: strings.NewReader("a,b\n1,2\n3,4"),
			to:   strings.NewReader("a,b,c\n1,2,5\n3,4,6"),
			opt:  defaultOptions("a"),
			want: DiffRef{
				Changed: map[string][]string{
					"1": []string{"c"},
					"3": []string{"c"},
				},
			},
		},
		{
			name: "column removed",
			from: strings.NewReader("a,b\n1,2\n3,4"),
			to:   strings.NewReader("a\n1\n3"),
			opt:  defaultOptions("a"),
			want: DiffRef{
				Changed: map[string][]string{
					"1": []string{"b"},
					"3": []string{"b"},
				},
			},
		},
		{
			name: "precision massages floats",
			from: strings.NewReader("a,b\n1,2.456789"),
			to:   strings.NewReader("a,b\n1,2.487651"),
			opt:  &Options{Precision: 1, Separator: ',', HasHeader: true, KeyColumns: []string{"a"}},
			want: DiffRef{
				Changed: map[string][]string{},
			},
		},
		{
			name: "compare is case sensitive",
			from: strings.NewReader("a,b\nHELLO,2.456789"),
			to:   strings.NewReader("a,b\nhello,2.487651"),
			opt:  &Options{IgnoreCase: false, Separator: ',', HasHeader: true, KeyColumns: []string{"a"}},
			want: DiffRef{
				AddedRow:   []string{"hello"},
				RemovedRow: []string{"HELLO"},
				Changed:    map[string][]string{},
			},
		},
		{
			name: "ignore case works on the keys",
			from: strings.NewReader("a,b\nHELLO,ddd"),
			to:   strings.NewReader("a,b\nhello,ddd"),
			opt:  &Options{IgnoreCase: true, Separator: ',', HasHeader: true, KeyColumns: []string{"a"}},
			want: DiffRef{
				Changed: map[string][]string{},
			},
		},
		{
			name: "ignore case works with precision",
			from: strings.NewReader("a,b\nHELLO,2.456789"),
			to:   strings.NewReader("a,b\nhello,2.487651"),
			opt:  &Options{Precision: 1, IgnoreCase: true, Separator: ',', HasHeader: true, KeyColumns: []string{"a"}},
			want: DiffRef{
				Changed: map[string][]string{},
			},
		},

		// fail path

		{
			name:    "bad from_csv errors",
			from:    strings.NewReader("a,b\",c"),
			opt:     &Options{LazyQuotes: false, Separator: ','},
			wantErr: true,
		},
		{
			name:    "bad to_csv errors",
			from:    strings.NewReader("a,b,c"),
			to:      strings.NewReader("a,b\",c"),
			opt:     &Options{LazyQuotes: false, Separator: ','},
			wantErr: true,
		},
	} {
		got, err := Compare(v.from, v.to, v.opt)

		if err != nil && !v.wantErr {
			t.Errorf("%s: unexpected error: %v", v.name, err)
			continue
		}

		if diff := cmp.Diff(v.want, got, cmpopts.IgnoreUnexported(csvData{})); diff != "" {
			t.Errorf("%s: got: %v, want: %v\n", v.name, got, v.want)
		}
	}
}
