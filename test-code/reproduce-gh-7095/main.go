package main

import (
	"os"
	"fmt"
	"strings"
	"strconv"
	"regexp"
)

var (
	statusLineRE = regexp.MustCompile(`(\d+) blocks .*\[(\d+)/(\d+)\] \[([U_]+)\]`)
	ErrFileParse = fmt.Errorf("eror parse file")
)

type ParserFunc func(deviceLine, statusLine string) (active, total, down, size int64, err error)

func test_with_input(parser ParserFunc) error {
	inputs, err := os.ReadFile("./input.txt")
	if err != nil {
		fmt.Errorf("error read input %w", err)
		return err
	}

	_, err = test_mdoutput(string(inputs), parser)
	if err != nil {
		fmt.Errorf("error parse input %w", err)
		return err
	}

	return nil
}

// reproduce and validate issue https://github.com/harvester/harvester/issues/7095
// copied and changed from https://github.com/prometheus/procfs/blob/969849f4953c8e5d6c913b4ce677aa92df77c538/mdstat.go#L82
func test_mdoutput(mdStatData string, parser ParserFunc) ([]string, error) {
	lines := strings.Split(string(mdStatData), "\n")

	for i, line := range lines {
		fmt.Printf("input: %v %v\n", i, line)
		if strings.TrimSpace(line) == "" || line[0] == ' ' ||
			strings.HasPrefix(line, "Personalities") ||
			strings.HasPrefix(line, "unused") {
			continue
		}

		deviceFields := strings.Fields(line)
		if len(deviceFields) < 3 {
			return nil, fmt.Errorf("%w: Expected 3+ lines, got %q", ErrFileParse, line)
		}
		mdName := deviceFields[0] // mdx
		state := deviceFields[2]  // active or inactive

		if len(lines) <= i+3 {
			return nil, fmt.Errorf("%w: Too few lines for md device: %q", ErrFileParse, mdName)
		}

		// Failed disks have the suffix (F) & Spare disks have the suffix (S).
		fail := int64(strings.Count(line, "(F)"))
		spare := int64(strings.Count(line, "(S)"))
		active, total, down, size, err := parser(lines[i], lines[i+1])

		if err != nil {
			return nil, fmt.Errorf("%w: Cannot parse md device lines: %v: %w", ErrFileParse, active, err)
		}

		fmt.Printf("the output %v %v %v %v %v %v %v %v", fail, spare, active, total, down, size, mdName, state)

		syncLineIdx := i + 2
		if strings.Contains(lines[i+2], "bitmap") { // skip bitmap line
			syncLineIdx++
		}
	}		

	return nil, nil
}

func evalStatusLine_buggy(deviceLine, statusLine string) (active, total, down, size int64, err error) {
/*
	statusFields := strings.Fields(statusLine)
	if len(statusFields) < 1 {
		return 0, 0, 0, 0, fmt.Errorf("%w: Unexpected statusline %q: %w", ErrFileParse, statusLine, err)
	}

	sizeStr := statusFields[0]
*/

	// bug fix // https://github.com/prometheus/procfs/commit/b9b5ad9b6d5a3ba720f13d9e0ff0ff9eb8ad3a3d#diff-1661c40d89115402230dda8d0195c034b40fb46a3d60d05b9baa649b37277c8aL170
	// replace below line with above, and the panic is gone
	sizeStr := strings.Fields(statusLine)[0]
	
	size, err = strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("%w: Unexpected statusline %q: %w", ErrFileParse, statusLine, err)
	}

	if strings.Contains(deviceLine, "raid0") || strings.Contains(deviceLine, "linear") {
		// In the device deviceLine, only disks have a number associated with them in [].
		total = int64(strings.Count(deviceLine, "["))
		return total, total, 0, size, nil
	}

	if strings.Contains(deviceLine, "inactive") {
		return 0, 0, 0, size, nil
	}

	matches := statusLineRE.FindStringSubmatch(statusLine)
	if len(matches) != 5 {
		return 0, 0, 0, 0, fmt.Errorf("%w: Could not fild all substring matches %s: %w", ErrFileParse, statusLine, err)
	}

	total, err = strconv.ParseInt(matches[2], 10, 64)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("%w: Unexpected statusline %q: %w", ErrFileParse, statusLine, err)
	}

	active, err = strconv.ParseInt(matches[3], 10, 64)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("%w: Unexpected active %d: %w", ErrFileParse, active, err)
	}
	down = int64(strings.Count(matches[4], "_"))

	return active, total, down, size, nil
}

func evalStatusLine(deviceLine, statusLine string) (active, total, down, size int64, err error) {
	statusFields := strings.Fields(statusLine)
	if len(statusFields) < 1 {
		return 0, 0, 0, 0, fmt.Errorf("%w: Unexpected statusline %q: %w", ErrFileParse, statusLine, err)
	}

	sizeStr := statusFields[0]

	// bug fix // https://github.com/prometheus/procfs/commit/b9b5ad9b6d5a3ba720f13d9e0ff0ff9eb8ad3a3d#diff-1661c40d89115402230dda8d0195c034b40fb46a3d60d05b9baa649b37277c8aL170
	// replace below line with above, and the panic is gone
	//sizeStr := strings.Fields(statusLine)[0]
	
	size, err = strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("%w: Unexpected statusline %q: %w", ErrFileParse, statusLine, err)
	}

	if strings.Contains(deviceLine, "raid0") || strings.Contains(deviceLine, "linear") {
		// In the device deviceLine, only disks have a number associated with them in [].
		total = int64(strings.Count(deviceLine, "["))
		return total, total, 0, size, nil
	}

	if strings.Contains(deviceLine, "inactive") {
		return 0, 0, 0, size, nil
	}

	matches := statusLineRE.FindStringSubmatch(statusLine)
	if len(matches) != 5 {
		return 0, 0, 0, 0, fmt.Errorf("%w: Could not fild all substring matches %s: %w", ErrFileParse, statusLine, err)
	}

	total, err = strconv.ParseInt(matches[2], 10, 64)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("%w: Unexpected statusline %q: %w", ErrFileParse, statusLine, err)
	}

	active, err = strconv.ParseInt(matches[3], 10, 64)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("%w: Unexpected active %d: %w", ErrFileParse, active, err)
	}
	down = int64(strings.Count(matches[4], "_"))

	return active, total, down, size, nil
}

func main() {
	var err error
	if len(os.Args) > 1 {	
		err = test_with_input(evalStatusLine)
	} else {
		err = test_with_input(evalStatusLine_buggy)
	}
	if err != nil {
		fmt.Errorf("error processing data %w", err)
	}
}
// run `go run main.go` to reproduce the issue
// run `go run main.go 1` to validate the fix sovles the issue
