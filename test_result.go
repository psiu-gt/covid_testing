// Type definitions for test results retrieved from Google Sheets
package main

type TestResult struct {
	Name           string
	WithWeek       bool
	TestDate       string
	LastTestResult string
}

func GetUntested(results *[]TestResult) *[]TestResult {
	var untested []TestResult
	for _, record := range *results {
		if !record.WithWeek {
			untested = append(untested, record)
		}
	}
	return &untested
}
