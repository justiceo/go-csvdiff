package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	csvdiff "github.com/justiceo/go-csvdiff/csvdiff"
)

type arrayFlags []string

func (arr *arrayFlags) String() string {
	return fmt.Sprintf("%v", *arr)
}

func (arr *arrayFlags) Set(value string) error {
	*arr = append(*arr, value)
	return nil
}

var (
	key        arrayFlags
	fromCSV    = flag.String("from_csv", "", "The base CSV file.")
	toCSV      = flag.String("to_csv", "", "CSV to compare to.")
	lazyQuotes = flag.Bool("lazy_quotes", false, "If true, a quote may appear in an unquoted field and a non-doubled quote may appear in a quoted field.")
	ignoreCase = flag.Bool("ignore_case", false, "Ignore case when comparing cell values. This also applies to key columns.")
	hasHeader  = flag.Bool("has_header", true, "If true, uses the first record as the csv header, otherwise uses a generated header.")
	precision  = flag.Int("precision", 0, "Number of decimal digits to approximate floating point values before comparison.")
	// separator  = flag.Rune("separator", ',', "Separator the field delimiter. It must be a valid rune and must not be \r, \n, or the Unicode replacement character (0xFFFD).")

)

func main() {
	flag.Var(&key, "key", "Column name to use as key.")
	flag.Parse()

	opt := &csvdiff.Options{
		KeyColumns: key,
		LazyQuotes: *lazyQuotes,
		IgnoreCase: *ignoreCase,
		Precision:  *precision,
		HasHeader:  *hasHeader,
		Separator:  ',',
	}

	fcsv, err := os.Open(*fromCSV)
	if err != nil {
		log.Fatalf("error opening from_csv: %v", err)
	}
	tcsv, err := os.Open(*toCSV)
	if err != nil {
		log.Fatalf("error opening to_csv: %v", err)
	}

	d, err := csvdiff.AsJSON(fcsv, tcsv, opt)
	if err != nil {
		log.Fatalf("error generating json diff: %v", err)
	}
	fmt.Println(d)
}
