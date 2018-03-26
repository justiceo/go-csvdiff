package csvdiff

import (
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestAsSummary(t *testing.T) {
	for _, v := range []struct {
		name     string
		from, to io.Reader
		opt      *Options
		want     string
		wantErr  bool
	}{
		{
			name: "no diff",
			from: strings.NewReader("a,b\n1,2\n3,4"),
			to:   strings.NewReader("a,b\n1,7\n5,6"),
			opt:  defaultOptions("a"),
			want: "1 rows added.\n1 rows changed.\n1 rows removed.\n",
		},
	} {

		got, err := AsSummary(v.from, v.to, v.opt)

		if err != nil && !v.wantErr {
			t.Errorf("%s: unexpected error: %v", v.name, err)
			continue
		}

		if diff := cmp.Diff(v.want, got, cmpopts.IgnoreUnexported(csvData{})); diff != "" {
			t.Errorf("%s: got: %v, want: %v\n", v.name, got, v.want)
		}
	}
}
