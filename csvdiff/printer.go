package csvdiff

import (
	"encoding/json"
	"fmt"
	"io"
)

// AsSummary returns summary of the diff.
func AsSummary(fromCSV, toCSV io.Reader, opt *Options) (string, error) {
	f, err := fromReader(fromCSV, opt)
	if err != nil {
		return "", fmt.Errorf("error parsing from_csv: %v", err)
	}
	t, err := fromReader(toCSV, opt)
	if err != nil {
		return "", fmt.Errorf("error parsing to_csv: %v", err)
	}
	d := f.diffAllRecords(t)
	s := ""
	s += fmt.Sprintf("%d rows added.\n", len(d.AddedRow))
	s += fmt.Sprintf("%d rows changed.\n", len(d.Changed))
	s += fmt.Sprintf("%d rows removed.\n", len(d.RemovedRow))
	return s, nil
}

// AsJSON returns json represention of diffRef (with values).
func AsJSON(fromCSV, toCSV io.Reader, opt *Options) (string, error) {
	f, err := fromReader(fromCSV, opt)
	if err != nil {
		return "", fmt.Errorf("error parsing from_csv: %v", err)
	}
	t, err := fromReader(toCSV, opt)
	if err != nil {
		return "", fmt.Errorf("error parsing to_csv: %v", err)
	}
	d := f.diffAllRecords(t)
	return diffToJSON(d, f, t)
}

func diffToJSON(diff DiffRef, from, to csvData) (string, error) {
	j := make(map[string][]map[string]string)
	for _, v := range diff.AddedRow {
		j["Added"] = append(j["Added"], to.recordObj(v))
	}
	for _, v := range diff.RemovedRow {
		j["Removed"] = append(j["Removed"], from.recordObj(v))
	}
	for k, v := range diff.Changed {
		for _, col := range v {
			change := make(map[string]string)
			change["key"] = k
			change["field"] = col
			change["from"] = ""
			if fi, ok := from.headerMap[k]; ok {
				change["from"] = from.records[k][fi]
			}
			change["to"] = ""
			if ti, ok := to.headerMap[k]; ok {
				change["to"] = to.records[k][ti]
			}
			j["Changed"] = append(j["Changed"], change)
		}
	}

	json, err := json.MarshalIndent(j, "", "  ")
	if err != nil {
		return "", err
	}
	return string(json), nil
}

func (c csvData) recordObj(key string) map[string]string {
	res := make(map[string]string)
	for col, i := range c.headerMap {
		res[col] = c.records[key][i]
	}
	return res
}
