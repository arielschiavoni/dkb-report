package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	// Define command-line flags for input and output filenames and columns to keep.
	inputFile := flag.String("input", "", "Input CSV file")
	outputFile := flag.String("output", "output.csv", "Output CSV file")
	columnsToKeep := flag.String("columns", "2,5,6,8", "Comma-separated list of column indices (1-based) to keep")
	recordsToSkip := flag.Int("skip", 4, "Number of records to skip (default: 4)")

	flag.Parse()

	if *inputFile == "" {
		fmt.Println("Usage: dkb-report -input <inputfile.csv> -output <outputfile.csv> -columns <column_indices> -skip <rows to skip>")
		os.Exit(1)
	}

	// Open input and output files.
	input, err := os.Open(*inputFile)
	if err != nil {
		fmt.Printf("Error opening input file: %v\n", err)
		os.Exit(1)
	}
	defer input.Close()

	output, err := os.Create(*outputFile)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer output.Close()

	// Parse column indices to keep.
	columnIndices, err := parseColumnsToKeep(*columnsToKeep)
	if err != nil {
		fmt.Printf("Error parsing column indices: %v\n", err)
		os.Exit(1)
	}

	inputReader := bufio.NewReader(input)
	// Skip the specified number of records.
	if err := skipRecords(inputReader, *recordsToSkip); err != nil {
		fmt.Printf("Error skipping records: %v\n", err)
		os.Exit(1)
	}

	// // Create a custom CSV reader with a different separator.
	customDelimiter := ';'
	reader := csv.NewReader(inputReader)
	reader.Comma = customDelimiter
	reader.LazyQuotes = true
	writer := csv.NewWriter(output)

	// Process the CSV file, keeping only specified columns.
	fmt.Printf("Keeping only the following columns: %s\n", *columnsToKeep)
	for {
		record, err := reader.Read()

		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Printf("Error reading CSV: %v\n", err)
			continue // Skip the problematic line and continue processing
		}

		// keep specified columns from the record.
		newRecord := keepColumns(record, columnIndices)

		// Write the modified record to the output file.
		if err := writer.Write(newRecord); err != nil {
			fmt.Printf("Error writing CSV: %v\n", err)
			os.Exit(1)
		}
	}

	// Flush and close the CSV writer.
	writer.Flush()
	if err := writer.Error(); err != nil {
		fmt.Printf("Error flushing CSV writer: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Done!")
}

// skipRecords skips the specified number of records in the input file.
func skipRecords(reader *bufio.Reader, recordsToSkip int) error {
	fmt.Printf("Records to skip: %d\n", recordsToSkip)

	for i := 0; i < recordsToSkip; i++ {
		record, err := reader.ReadString('\n')

		fmt.Printf("Skipping record %s", record)
		if err != nil && err != io.EOF {
			return err
		}
	}

	return nil
}

// parseColumnsToKeep parses the comma-separated column indices and returns them as an array of integers.
func parseColumnsToKeep(input string) ([]int, error) {
	columns := []int{}
	indices := strings.Split(input, ",")

	for _, indexStr := range indices {
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			return nil, err
		}
		// substruct 1 due we need to go from column number to column index (0-based)
		columns = append(columns, index-1)
	}

	return columns, nil
}

// keepColumns keeps specified columns from a record.
func keepColumns(record []string, columnsToKeep []int) []string {
	newRecord := []string{}

	for i, value := range record {

		if contains(columnsToKeep, i) {
			// Define a regular expression ammountPattern to match only the value (leave out currency symbol!)
			ammountPattern := `(-?\d{1,3}(?:\.\d{3})*(?:,\d{1,2})?)(?: €)`

			// Compile the regex pattern
			regex := regexp.MustCompile(ammountPattern)

			// Find the first match
			matches := regex.FindStringSubmatch(value)

			// Extract the matched price value
			if len(matches) > 1 {
				amount := matches[1]
				newRecord = append(newRecord, amount)
			} else {
				newRecord = append(newRecord, value)
			}
		}
	}

	return newRecord
}

// contains checks if a value is present in a slice of integers.
func contains(slice []int, value int) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
