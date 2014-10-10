// Package skini is an improved INI file parser
package skini

/*
Skini -- parser of improved ini files
*/

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
)

// Parses file into provided template.
// Template can have only maps string to string.
func Parse(target interface{}, r io.Reader) (err error) {
	if elem, err := getElem(target); err == nil {
		return parseInput(&elem, r)
	}
	return
}

// Parses config file with given filename.
// Returns result as a map of string values.
func ParseFile(target interface{}, filename string) (err error) {
	elem, err := getElem(target)
	if err != nil {
		return
	}

	// Read config file
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error reading file: %s", filename)
	}
	defer file.Close()

	return parseInput(&elem, file)
}

// Read single field specified by key from input file.
func SeekFile(target interface{}, filename string, key string) (value string, err error) {
	elem, err := getElem(target)
	if err != nil {
		return
	}

	// Read config file
	file, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("error reading file: %s", filename)
	}
	defer file.Close()

	return seekInput(&elem, file, key)
}

// Find first relevant config file in given directory.
// Pattern is just like in terminal: config_*.ini, *.ini
// Key is the key inside the file that must be present and matched.
// Relevance of file is defined by provided function.
func ParseDir(target interface{}, dir string, pattern string, idkey string, matcher func(string) bool) (err error) {
	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("error scanning directory: %s, error = ", dir, err)
	}

	// Build pattern regex
	re, err := wildcardRegex(pattern)
	if err != nil {
		return err
	}

	// For each file in directory
	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}
		filename := fi.Name()
		if !re.MatchString(filename) {
			continue
		}

		// Seek file for idkey
		filename = path.Join(dir, filename)
		idvalue, err := SeekFile(target, filename, idkey)
		if err != nil {
			return fmt.Errorf("Error while seeking file '%s', error: %s", filename, err)
		}

		// Check if idkey value matches expected
		if matcher(idvalue) {
			return ParseFile(target, filename)
		}
	}
	return errors.New("No matching configuration file found")
}
