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

type change struct {
	key, column string
}

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

	// get the indices of columns to use as key
	colIndices, err := c.getColIndices(opt.KeyColumns)
	if err != nil {
		return c, err
	}

	for _, record := range records {
		key := getKey(colIndices, record)
		c.records[key] = record
	}
	return c, nil
}

func getKey(colIndices []int, record []string) string {
	var keys []string
	for _, v := range colIndices {
		keys = append(keys, record[v])
	}
	return strings.Join(keys, ",")
}

func (c csvData) getColIndices(colNames []string) ([]int, error) {
	var indices []int
	for _, col := range colNames {
		i, ok := c.headerMap[col]
		if !ok {
			return nil, fmt.Errorf("Invalid column header: %s. Available headers: %v", col, c.headers)
		}
		indices = append(indices, i)
	}
	return indices, nil
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
	d := DiffRef{}
	for k := range c.records {
		if _, ok := other.records[k]; !ok {
			d.RemovedRow = append(d.RemovedRow, k)
		} else if changes := c.diffSingleRecord(k, other); changes != nil {
			d.Changed[k] = changes
		}
		delete(other.records, k)
	}
	for k := range other.records {
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
