package skini

/*
Parser -- responsible for parsing input line by line.
Parser recognizes each line's expression type and asks
Reflector to add values to corresponding target field.
*/

import (
	"fmt"
	"reflect"
	"regexp"
)

//------------------------------------------------------------
// Regex
//------------------------------------------------------------

// Looks like: [section]
var reLikeSection = regexp.MustCompile(`^\[[a-zA-Z0-9\.\|\s]+\]$`)

// Looks like map ? [map.*]
var reLikeMap = regexp.MustCompile(`^\[\s*map\..*\s*\]$`)

// Looks like: key [+]= value
var reLikeKeyValue = regexp.MustCompile(`^[^=]+\s+\+?=\s*.*$`)

// key [+]= value
var reKeyValue = regexp.MustCompile(`^(?P<key>[^=]+)\s+\+?=\s*(?P<value>.*)$`)

// key += value
var reKeyValuePlus = regexp.MustCompile(`^(?P<key>[^=]+)\s+\+=\s*(?P<value>.*)$`)

// Is this a map ? [map.*]
var reIsMap = regexp.MustCompile(`^\[\s*map\..*\s*\]$`)

// Map elements: [map.name | keyname ]
var reMap = regexp.MustCompile(`^\[\s*map\.(?P<name>[a-zA-Z0-9_\.]+)(?P<suffix>\s*\|\s*(?P<key>[a-zA-Z0-9_\-\.\*]*))?\s*\]$`)

// [section]
var reSection = regexp.MustCompile(`^\[\s*(?P<key>[a-zA-Z0-9\.]+)\s*\]$`)

//------------------------------------------------------------
// Parser structures
//------------------------------------------------------------

// Expression types
const (
	ExprNone    int = 0
	ExprSection int = 1 << iota
	ExprMap
	ExprList
	ExprKeyVal
	ExprVal
)

// Expresson values
type exprValues struct {
	name  string
	value string
}

// Parser state
type parserState struct {
	// Captures
	capSection string
	capMap     string
	capSubmap  string
	capList    string
}

//------------------------------------------------------------
// Line by line parser
//------------------------------------------------------------

func parseLine(target *reflect.Value, lineA, lineB string, state *parserState) (err error) {
	// Be ready to catch panic and report which line caused it
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("[PANIC] for:\n")
			fmt.Printf("  State.section = %v, .map = %v, .submap = %v, .list = %v\n",
				state.capSection, state.capMap, state.capSubmap, state.capList)
			fmt.Printf("  lineA: %s\n  lineB: %s\n", lineA, lineB)
			fmt.Printf("Panic type: %v\n", err)
			//fmt.Printf("Panic type: \v%v\n", debug.Stack())
			panic("[skini] Aborted, paniced during parse")
		}
	}()

	// Skip skippables (comments, etc.)
	if isSkip(lineA) {
		return
	}

	// Parse line as an expression
	typ, vals, err := parseExpr(lineA, lineB)
	if err != nil {
		return
	}

	/* Uncomment for capture debugging:
	var debugCapture = ""
	if state.capSection != "" {
		debugCapture = state.capSection
	} else if state.capMap != "" {
		debugCapture = state.capMap
	}
	*/

	// Exception:
	// If we're in a map then lists are not supported
	// and it's okay for a key to have empty value
	if state.capMap != "" && typ == ExprList {
		typ = ExprKeyVal
	}

	switch typ {

	case ExprSection:
		//fmt.Printf("\t==> SECTION, name = %s\n", vals.name)
		state.capSection = vals.name
		state.capMap, state.capSubmap, state.capList = "", "", ""

	case ExprMap:
		//fmt.Printf("\t==> MAP, name = %s, value = %s\n", vals.name, vals.value)
		state.capMap, state.capSubmap = vals.name, vals.value
		state.capSection, state.capList = "", ""

	case ExprList:
		//fmt.Printf("\t\t[%s] LIST, name = %s\n", debugCapture, vals.name)
		state.capList = vals.name

	// K = V
	case ExprKeyVal:
		//fmt.Printf("\t\t[%s] KeyVal, name = %s, value = %s\n", debugCapture, vals.name, vals.value)
		state.capList = ""

		if state.capMap != "" {
			// KV in Map: either map[s]s or map[s]map[s]s
			err = addMapItem(target, state.capMap, state.capSubmap, vals.name, vals.value)
		} else {
			// KV in Section: simple field
			err = setField(target, state.capSection, vals.name, vals.value)
		}

	// K = ...Vi
	case ExprVal:
		//fmt.Printf("\t\t\t[%s] Val, value = %s\n", state.capList, vals.value)
		if state.capMap != "" {
			// V in Map: not yet supported
			fmt.Printf("[SKINI] SKIPPING: Not supported: list in map: [%s] value = %s\n", state.capList, vals.value)
		} else {
			// V in Section: slice item, either top level or section
			err = addSliceItem(target, state.capSection, state.capList, vals.value)
		}

	default:
		err = fmt.Errorf("\tOOPS! Parser doesn't know how to handle this line: %s\n", lineA)
	}

	return
}

//------------------------------------------------------------
// Expression parser
//------------------------------------------------------------

func parseExpr(lineA, lineB string) (typ int, values *exprValues, err error) {

	// Map ? Must check before section
	if name, key, ok := isMap(lineA); ok {
		typ = ExprMap
		values = &exprValues{name, key}
		return
	}

	// Section ?
	if name, ok := isSection(lineA); ok {
		typ = ExprSection
		values = &exprValues{name, ""}
		return
	}

	// List ?
	if name, ok := isList(lineA, lineB); ok {
		typ = ExprList
		values = &exprValues{name, ""}
		return
	}

	// KV ?
	if name, value, ok := isKeyValue(lineA); ok {
		typ = ExprKeyVal
		values = &exprValues{name, value}
		return
	}

	// V ?
	if value, ok := isValue(lineA); ok {
		typ = ExprVal
		values = &exprValues{"", value}
		return
	}

	err = fmt.Errorf("unrecognized expression: %s\n", lineA)
	return
}

//------------------------------------------------------------
// Expression type recognizers
//------------------------------------------------------------

// Is skippable ?
func isSkip(line string) bool {
	if line == "" || line[0] == '#' || line[0] == ';' {
		return true
	}
	return false
}

// Is anything looking like a section ?
func isLikeSection(line string) bool {
	if line == "" || line[0] != '[' {
		return false
	}

	match := reLikeSection.FindStringSubmatch(line)
	if match == nil {
		return false
	}
	return true
}

// Is anything looking like a map ?
func isLikeMap(line string) bool {
	if line == "" || line[0] != '[' {
		return false
	}

	match := reLikeMap.FindStringSubmatch(line)
	if match == nil {
		return false
	}
	return true
}

// Is anything looking like a section ?
func isLikeKeyValue(line string) bool {
	match := reLikeKeyValue.FindStringSubmatch(line)
	if match == nil {
		return false
	}
	return true
}

// Is map definition ?
func isMap(line string) (name, key string, ok bool) {
	match := reIsMap.FindStringSubmatch(line)
	if match == nil {
		return
	}
	match = reMap.FindStringSubmatch(line)
	if match == nil {
		return
	}

	for i, capt := range reMap.SubexpNames() {
		switch capt {
		case "name":
			name = match[i]
		case "key":
			key = match[i]
		}
	}

	// At least name must be present
	if name != "" {
		ok = true
	}

	return
}

// Is section definition ? Section is considered after
// map test failed.
func isSection(line string) (key string, ok bool) {
	if line == "" || line[0] != '[' {
		return
	}

	match := reSection.FindStringSubmatch(line)
	if match == nil {
		return
	}

	for i, capt := range reSection.SubexpNames() {
		switch capt {
		case "key":
			return match[i], true
		}
	}
	return
}

// Is Key = Value ?
func isKeyValue(line string) (key, value string, ok bool) {
	match := reKeyValue.FindStringSubmatch(line)
	if match == nil {
		return
	}

	for i, capt := range reKeyValue.SubexpNames() {
		switch capt {
		case "key":
			key = match[i]
		case "value":
			value = match[i]
		}
	}

	if key != "" {
		ok = true
	}
	return
}

// Is Key += Value ?
func isKeyValuePlus(line string) bool {
	match := reKeyValuePlus.FindStringSubmatch(line)
	if match == nil {
		return false
	}
	return true
}

// Is Key Value List: Key = "" followed by Value ?
func isList(lineA, lineB string) (key string, ok bool) {
	var value string

	// Line A
	// Must look like "key = val"
	key, value, ok = isKeyValue(lineA)
	if !ok {
		return
	}

	// Line B
	// To qualify as a list, the value must be empty
	if value == "" {
		if value, ok = isValue(lineB); ok {
			return
		}
	} else {
		ok = false
	}

	return
}

// Is Value ?
func isValue(line string) (value string, ok bool) {
	// Value is anything that doesn't look like:
	// - section
	// - section map
	// - key value pair

	if _, ok = isSection(line); ok {
		return line, false
	}

	if _, _, ok = isMap(line); ok {
		return line, false
	}

	if _, _, ok = isKeyValue(line); ok {
		return line, false
	}

	return line, true
}
