// Package csvdiff implements a robust csv differ in go
package csvdiff

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Options represents configuration for csv diff.
type Options struct {
	IgnoreColumns, KeyColumns         []string
	IgnoreCase, LazyQuotes, HasHeader bool
	Separator                         rune
	Precision                         int
	Comparator                        Comparator
}

// Comparator returns true if values in cell "a" and "b" are considered equal or false otherwise.
type Comparator func(a, b string) bool

// DiffRef contains references to keys of changed records between two csvs
type DiffRef struct {
	AddedRow, RemovedRow []string
	Changed              map[string][]string
}

type csvData struct {
	headers   []string
	headerMap map[string]int
	records   map[string][]string
	opt       *Options
}

func (o Options) getComparator() Comparator {
	if o.Comparator != nil {
		return o.Comparator
	}

	return func(a, b string) bool {
		// attempt a numeric field comparison first
		afloat, aErr := strconv.ParseFloat(a, 64)
		bfloat, bErr := strconv.ParseFloat(b, 64)
		if aErr == nil && bErr == nil && o.Precision > 0 {
			p := "%." + strconv.Itoa(o.Precision) + "f"
			return fmt.Sprintf(p, afloat) == fmt.Sprintf(p, bfloat)
		}

		if o.IgnoreCase {
			return strings.ToLower(a) == strings.ToLower(b)
		}
		return a == b
	}
}

// OrderedSet returns a set from a slice
func OrderedSet(slices ...[]string) []string {
	set := []string{}
	m := make(map[string]bool)
	for _, slice := range slices {
		for _, s := range slice {
			if _, exists := m[s]; !exists {
				set = append(set, s)
				m[s] = true
			}
		}
	}
	return set
}

func fromReader(f io.Reader, opt *Options) (csvData, error) {
	if f == nil || opt == nil {
		return csvData{}, fmt.Errorf("reader and options cannot be nil")
	}

	c := csvData{
		headers:   []string{},
		headerMap: make(map[string]int),
		records:   make(map[string][]string),
		opt:       opt,
	}

	r := csv.NewReader(f)
	r.Comma = opt.Separator
	r.LazyQuotes = opt.LazyQuotes

	records, err := r.ReadAll()
	if err != nil {
		return c, err
	}

	if opt.HasHeader && len(records) > 0 {
		c.headers = records[0]
		for i, col := range c.headers {
			c.headerMap[col] = i
		}
		records = records[1:]
	}

	// TODO(justiceo): KeyColumns should be required for this implementation
	for _, record := range records {
		key, err := c.getKey(opt.KeyColumns, record)
		if err != nil {
			return c, err
		}
		if opt.IgnoreCase {
			key = strings.ToLower(key)
		}
		c.records[key] = record
	}
	return c, nil
}

func (c csvData) getKey(colNames []string, record []string) (string, error) {
	key := []string{}
	for _, col := range colNames {
		i, ok := c.headerMap[col]
		if !ok {
			return "", fmt.Errorf("Invalid column header: %s. Available headers: %v", col, c.headers)
		}
		key = append(key, record[i])
	}
	return strings.Join(key, ","), nil
}

// diffSingleRecordUnordered
func (c csvData) diffSingleRecord(key string, other csvData) []string {
	changes := []string{}
	headers := OrderedSet(c.headers, other.headers)
	for _, h := range headers {
		ci, ok1 := c.headerMap[h]
		oi, ok2 := other.headerMap[h]
		if ok1 && ok2 && c.opt.getComparator()(c.records[key][ci], other.records[key][oi]) {
			continue
		}
		changes = append(changes, h)
	}
	return changes
}

func (c csvData) diffAllRecords(other csvData) DiffRef {
	d := DiffRef{Changed: map[string][]string{}}
	for k := range c.records {
		if _, ok := other.records[k]; !ok {
			d.RemovedRow = append(d.RemovedRow, k)
		} else if changes := c.diffSingleRecord(k, other); len(changes) != 0 {
			d.Changed[k] = changes
		}
	}
	for k := range other.records {
		if _, ok := c.records[k]; ok {
			continue
		}
		d.AddedRow = append(d.AddedRow, k)
	}
	return d
}

// Compare returns the DiffRef between two csv files
func Compare(fromCSV, toCSV io.Reader, opt *Options) (DiffRef, error) {
	f, err := fromReader(fromCSV, opt)
	if err != nil {
		return DiffRef{}, fmt.Errorf("error parsing from_csv: %v", err)
	}
	t, err := fromReader(toCSV, opt)
	if err != nil {
		return DiffRef{}, fmt.Errorf("error parsing to_csv: %v", err)
	}
	return f.diffAllRecords(t), nil
}
