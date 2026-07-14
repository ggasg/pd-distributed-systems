// plan-dates.go — print the full study schedule given a start date for W01.
//
// Usage:
//   go run tools/plan-dates.go --start 2026-07-06
//
// W00 (Infrastructure Setup) falls the week before the given start date.
// W01 starts on the given date. Each subsequent week is +7 days.

package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

type week struct {
	num   int
	arc   string
	title string
}

var weeks = []week{
	{0, "Setup", "Infrastructure Setup"},
	{1, "Arc 1", "LSM-Trees and Storage Engines"},
	{2, "Arc 1", "Encoding and Wire Formats"},
	{3, "Arc 1", "MapReduce and Its Limits"},
	{4, "Arc 1", "Clocks, Causality, and Time"},
	{5, "Arc 2", "Stream Processing Primitives"},
	{6, "Arc 2", "Naiad and Timely Dataflow"},
	{7, "Arc 2", "Differential Dataflow"},
	{8, "Arc 2", "Query Execution"},
	{9, "Arc 2", "Rule-Based Query Planning in Scala"},
	{10, "Arc 2", "Aggregation Algebra: Monoids and Semigroups"},
	{11, "Arc 3", "ML Data Pipelines"},
	{12, "Arc 3", "Distributed Training"},
	{13, "Arc 3", "The Actor Model and Ray"},
	{14, "Arc 3", "GPU Memory and Compute"},
	{15, "Arc 3", "Attention and KV Cache"},
	{16, "Arc 3", "Fault Tolerance and Snapshots"},
	{17, "Arc 3", "Capstone"},
	{18, "Arc 4", "Kubernetes and Operators"},
	{19, "Arc 4", "Observability: Metrics, Tracing, Logging"},
}

const optionalCapstoneWeek = 20

func main() {
	startStr := flag.String("start", "", "Start date for W01 (YYYY-MM-DD, required)")
	flag.Parse()

	if *startStr == "" {
		fmt.Fprintln(os.Stderr, "error: --start is required (e.g. --start 2026-07-06)")
		os.Exit(1)
	}

	w01Start, err := time.Parse("2006-01-02", *startStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing date %q: %v\n", *startStr, err)
		os.Exit(1)
	}

	today := time.Now().Truncate(24 * time.Hour)

	fmt.Printf("%-5s  %-7s  %-12s  %-12s  %s\n", "Week", "Arc", "Start", "End", "Topic")
	fmt.Println("----------------------------------------------------------------------")

	for _, w := range weeks {
		var weekStart time.Time
		if w.num == 0 {
			weekStart = w01Start.AddDate(0, 0, -7)
		} else {
			weekStart = w01Start.AddDate(0, 0, (w.num-1)*7)
		}
		weekEnd := weekStart.AddDate(0, 0, 6)

		marker := ""
		if !today.Before(weekStart) && !today.After(weekEnd) {
			marker = " ← now"
		}

		fmt.Printf("W%-4d  %-7s  %-12s  %-12s  %s%s\n",
			w.num,
			w.arc,
			weekStart.Format("Jan 02 2006"),
			weekEnd.Format("Jan 02 2006"),
			w.title,
			marker,
		)
	}

	lastCore := weeks[len(weeks)-1].num
	optionalEnd := w01Start.AddDate(0, 0, (optionalCapstoneWeek-1)*7+6)

	fmt.Printf("\nTotal: %d core weeks  |  W01 starts %s  |  W%d ends %s\n",
		len(weeks),
		w01Start.Format("Jan 02 2006"),
		lastCore,
		w01Start.AddDate(0, 0, (lastCore-1)*7+6).Format("Jan 02 2006"),
	)
	fmt.Printf("Optional W%d (Grand Capstone) ends %s\n", optionalCapstoneWeek, optionalEnd.Format("Jan 02 2006"))
}
