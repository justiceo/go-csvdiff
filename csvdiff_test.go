package csvdiff

import (
	"io"
	"reflect"
	"strings"
	"testing"
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
				opt: defaultOptions("a"),
			},
		},
	} {

		got, err := fromReader(v.reader, v.opt)

		if err != nil && !v.wantErr {
			t.Errorf("%s: unexpected error: %v", v.name, err)
		}

		if !reflect.DeepEqual(v.want, got) {
			t.Errorf("%s: got: %v, want: %v\n", v.name, got, v.want)
		}
	}

}
