package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type weekStart int

const (
	startSunday weekStart = iota
	startMonday
)

const (
	calWidth     = 20
	calGap       = "  "
	calWeekRows  = 6
	calPerRow    = 3
	monthsInYear = 12
)

var calCmd = &cobra.Command{
	Use:   "cal [[month] year]",
	Short: "Display a calendar",
	Long: `Display a calendar for a given month and year.
If no arguments are given, the current month is displayed.
If only the year is given, the full year is displayed.`,
	Args: cobra.RangeArgs(0, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		monday, _ := cmd.Flags().GetBool("monday")
		wholeYear, _ := cmd.Flags().GetBool("year")
		three, _ := cmd.Flags().GetBool("three")

		ws := startSunday
		if monday {
			ws = startMonday
		}

		now := time.Now()
		month, year := int(now.Month()), now.Year()

		switch len(args) {
		case 1:
			y, err := strconv.Atoi(args[0])
			if err != nil || y < 1 {
				return fail(1, "cal: invalid year: %s", args[0])
			}
			year, wholeYear = y, true
		case 2:
			m, err1 := strconv.Atoi(args[0])
			y, err2 := strconv.Atoi(args[1])
			if err1 != nil || err2 != nil || m < 1 || m > monthsInYear || y < 1 {
				return fail(1, "cal: invalid date: %s %s", args[0], args[1])
			}
			month, year = m, y
		}

		switch {
		case three:
			printThreeMonths(month, year, ws)
		case wholeYear:
			printYear(year, ws)
		default:
			printBlocks([][]string{monthBlock(month, year, ws)})
		}
		return nil
	},
}

func init() {
	calCmd.Flags().BoolP("monday", "m", false, "start the week on Monday")
	calCmd.Flags().BoolP("year", "y", false, "show the whole year")
	calCmd.Flags().BoolP("three", "3", false, "show the previous, current and next month")
}

func monthBlock(month, year int, ws weekStart) []string {
	title := fmt.Sprintf("%s %d", time.Month(month), year)
	pad := max((calWidth-len(title))/2, 0)

	header := "Su Mo Tu We Th Fr Sa"
	if ws == startMonday {
		header = "Mo Tu We Th Fr Sa Su"
	}

	block := []string{strings.Repeat(" ", pad) + title, header}

	offset := int(time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local).Weekday())
	if ws == startMonday {
		offset = (offset + 6) % 7
	}

	cells := make([]string, 0, offset+daysIn(month, year))
	for i := 0; i < offset; i++ {
		cells = append(cells, "  ")
	}
	for day := 1; day <= daysIn(month, year); day++ {
		cells = append(cells, fmt.Sprintf("%2d", day))
	}

	for row := 0; row < calWeekRows; row++ {
		from := row * 7
		if from >= len(cells) {
			block = append(block, "")
			continue
		}
		block = append(block, strings.Join(cells[from:min(from+7, len(cells))], " "))
	}
	return block
}

func printBlocks(blocks [][]string) {
	rows := 0
	for _, b := range blocks {
		rows = max(rows, len(b))
	}

	var lines []string
	for r := 0; r < rows; r++ {
		cols := make([]string, len(blocks))
		for i, b := range blocks {
			cell := ""
			if r < len(b) {
				cell = b[r]
			}
			cols[i] = fmt.Sprintf("%-*s", calWidth, cell)
		}
		lines = append(lines, strings.TrimRight(strings.Join(cols, calGap), " "))
	}

	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	for _, l := range lines {
		fmt.Fprintln(stdout, l)
	}
}

func printThreeMonths(month, year int, ws weekStart) {
	var blocks [][]string
	for delta := -1; delta <= 1; delta++ {
		t := time.Date(year, time.Month(month+delta), 1, 0, 0, 0, 0, time.Local)
		blocks = append(blocks, monthBlock(int(t.Month()), t.Year(), ws))
	}
	printBlocks(blocks)
}

func printYear(year int, ws weekStart) {
	fmt.Fprintf(stdout, "%*d\n\n", (calWidth*calPerRow+len(calGap)*(calPerRow-1)+len(strconv.Itoa(year)))/2, year)

	for first := 1; first <= monthsInYear; first += calPerRow {
		var blocks [][]string
		for m := first; m < first+calPerRow; m++ {
			blocks = append(blocks, monthBlock(m, year, ws))
		}
		printBlocks(blocks)
		if first+calPerRow <= monthsInYear {
			fmt.Fprintln(stdout)
		}
	}
}

func daysIn(month, year int) int {
	if month == 2 {
		if (year%4 == 0 && year%100 != 0) || year%400 == 0 {
			return 29
		}
		return 28
	}

	switch month {
	case 4, 6, 9, 11:
		return 30
	}
	return 31
}
