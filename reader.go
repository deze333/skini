package skini

/*
Reader -- reads input line by line and
calls Parser to parse each line.
*/

import (
	"bufio"
	"fmt"
	"io"
    "reflect"
	"strings"
)

//------------------------------------------------------------
// Read and parses input line by line
//------------------------------------------------------------

// Parse whole input.
func parseInput(target *reflect.Value, r io.Reader) (err error) {
    var lineA, lineB string
	scanner := bufio.NewScanner(r)

    // Read first line
    if lineA, err = readNextLine(scanner); err != nil {
        return
    }
    if lineA == "" {
		return fmt.Errorf("error, file is empty")
    }

    // Initialize parser state
    pstate := &parserState{}

    // Read consecutive lines
    //i := 0
    for {
        if lineA == "" {
            break
        }

        if lineB, err = readNextLine(scanner); err != nil {
            return
        }

        // Special case of 'k += v', stick all lines together
        if isKeyValuePlus(lineA) {
            if lineA, lineB, err = appendLines(scanner, lineA, lineB); err != nil {
                return
            }
        }

        // DEBUG
        //fmt.Println(i, " : ", lineA)
        //i ++

        // Parse line
        if err = parseLine(target, lineA, lineB, pstate); err != nil {
            return
        }

        // Move to next scan ahead line
        lineA = lineB
    }
    return
}

// Append consecutive lines until next 'k = v' or [section].
func appendLines(scanner *bufio.Scanner, l1, l2 string) (lineA, lineB string, err error) {

/*
    fmt.Println()
    fmt.Println("l1 = ", l1)
    fmt.Println("l2 keyVal ?", isLikeKeyValue(l2), "=", l2)
    fmt.Println("l2 section ?", isLikeSection(l2), "=", l2)
    fmt.Println("l2 map ?", isLikeMap(l2), "=", l2)
*/

    // If next line is another value or section, return now
    if isLikeKeyValue(l2) || isLikeSection(l2) || isLikeMap(l2) {
        return l1, l2, nil
    }

    lines := []string{l1, " ", l2}
    for {
        // Read next line to check if join ends there or not
        if lineB, err = readNextLine(scanner); err != nil {
            return
        }

/*
    fmt.Println()
    fmt.Println("l1 = ", strings.Join(lines, ""))
    fmt.Println("lB keyVal ?", isLikeKeyValue(lineB), "=", lineB)
    fmt.Println("lB section ?", isLikeSection(lineB), "=", lineB)
    fmt.Println("lB map ?", isLikeMap(lineB), "=", lineB)
*/

        // Join ends if next line is:
        // EOF
        if lineB == "" {
            break
        }
        // 'k = v'
        if isLikeKeyValue(lineB) {
            break
        }
        // like any [section]
        if isLikeSection(lineB) {
            break
        }
        // like any [map.name]
        if isLikeMap(lineB) {
            break
        }

        // None of those, append
        lines = append(lines, " ", lineB)
    }
    return strings.Join(lines, ""), lineB, err
}

// Seek specified key.
func seekInput(target *reflect.Value, r io.Reader, key string) (value string, err error) {
    var lineA, lineB string
	scanner := bufio.NewScanner(r)

    // Read first line
    if lineA, err = readNextLine(scanner); err != nil {
        return
    }
    if lineA == "" {
		return "", fmt.Errorf("error, file is empty")
    }

    // Read consecutive lines
    for {
        if lineA == "" {
            break
        }

        if lineB, err = readNextLine(scanner); err != nil {
            return
        }

        // Check for key
        if strings.HasPrefix(lineA, key) {
            if _, value, ok := isKeyValue(lineA); ok {
                return value, nil
            } else {
                return "", fmt.Errorf("error, cannot parse: %s", value)
            }
        }

        // Move to next scan ahead line
        lineA = lineB
    }
    return
}

// Reads next not empty line from provide scanner. 
// Trims spaces and tabs from input line.
// Returns empty line on EOF.
// May return error if reading experienced one.
func readNextLine(scanner *bufio.Scanner) (line string, err error) {
    for {
        hasMore := scanner.Scan()
        line = strings.Trim(scanner.Text(), " \t")

        // Can't read any more ?
        if !hasMore {
            return "", scanner.Err()
        }

        if line != "" {
            return
        }
    }
}

