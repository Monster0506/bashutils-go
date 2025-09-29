package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var calCmd = &cobra.Command{
	Use:   "cal [month] [year]",
	Short: "Display a calendar",
	Long: `Display a calendar for a given month and year.
If no arguments are given, the current month is displayed.
If only the year is given, the full year is displayed.`,
	Args: cobra.RangeArgs(0, 2),
	Run: func(cmd *cobra.Command, args []string) {
		now := time.Now()
		month := int(now.Month())
		year := now.Year()

		if len(args) == 1 {
			// One arg → full year
			y, err := strconv.Atoi(args[0])
			if err != nil || y < 1 {
				fmt.Fprintf(os.Stderr, "cal: invalid year: %s\n", args[0])
				os.Exit(1)
			}
			printYear(y)
			return
		}

		if len(args) == 2 {
			// Two args → month + year
			m, err1 := strconv.Atoi(args[0])
			y, err2 := strconv.Atoi(args[1])
			if err1 != nil || err2 != nil || m < 1 || m > 12 || y < 1 {
				fmt.Fprintf(os.Stderr, "cal: invalid date: %s %s\n", args[0], args[1])
				os.Exit(1)
			}
			month = m
			year = y
		}

		printMonth(month, year)
	},
}

func init() {
	// Register the cal command
	rootCmd.AddCommand(calCmd)
}

// printMonth prints a single month calendar
func printMonth(month, year int) {
	monthName := time.Month(month).String()
	fmt.Printf("     %s %d\n", monthName, year)
	fmt.Println("Su Mo Tu We Th Fr Sa")

	// First day of the month
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	startWeekday := int(firstDay.Weekday()) // Sunday=0

	// Days in month
	daysInMonth := daysIn(month, year)

	// Print leading spaces
	for i := 0; i < startWeekday; i++ {
		fmt.Print("   ")
	}

	// Print days
	day := 1
	for i := startWeekday; day <= daysInMonth; i++ {
		fmt.Printf("%2d ", day)
		if i%7 == 6 {
			fmt.Println()
		}
		day++
	}
	fmt.Println()
}

// printYear prints the full year calendar
func printYear(year int) {
	fmt.Printf("                             %d\n\n", year)
	for quarter := 0; quarter < 4; quarter++ {
		// Print month names
		for m := 1; m <= 3; m++ {
			month := time.Month(quarter*3 + m)
			fmt.Printf("      %-9s", month.String())
			if m < 3 {
				fmt.Print("   ")
			}
		}
		fmt.Println()

		// Print weekday headers
		for m := 0; m < 3; m++ {
			fmt.Print("Su Mo Tu We Th Fr Sa  ")
		}
		fmt.Println()

		// Prepare month data
		monthDays := [3][]string{}
		for m := 0; m < 3; m++ {
			monthNum := quarter*3 + m + 1
			monthDays[m] = monthLines(monthNum, year)
		}

		// Print weeks side by side
		for week := 0; week < 6; week++ {
			for m := 0; m < 3; m++ {
				if week < len(monthDays[m]) {
					fmt.Print(monthDays[m][week])
				} else {
					fmt.Print(strings.Repeat(" ", 20))
				}
				if m < 2 {
					fmt.Print("  ")
				}
			}
			fmt.Println()
		}
		fmt.Println()
	}
}

// monthLines returns a slice of strings representing each week of a month
func monthLines(month, year int) []string {
	lines := []string{}
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	startWeekday := int(firstDay.Weekday())
	daysInMonth := daysIn(month, year)

	week := strings.Repeat("   ", startWeekday)
	day := 1
	for {
		for i := startWeekday; i < 7 && day <= daysInMonth; i++ {
			week += fmt.Sprintf("%2d ", day)
			day++
		}
		lines = append(lines, strings.TrimRight(week, " "))
		if day > daysInMonth {
			break
		}
		week = ""
		startWeekday = 0
	}
	// Pad each line to fixed width
	for i := range lines {
		lines[i] = fmt.Sprintf("%-20s", lines[i])
	}
	return lines
}

// daysIn returns the number of days in a given month/year
func daysIn(month, year int) int {
	// February leap year check
	if month == 2 {
		if (year%4 == 0 && year%100 != 0) || (year%400 == 0) {
			return 29
		}
		return 28
	}
	// Months with 31 days
	if month == 1 || month == 3 || month == 5 || month == 7 ||
		month == 8 || month == 10 || month == 12 {
		return 31
	}
	return 30
}
