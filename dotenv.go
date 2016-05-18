package configure

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// NewDotEnv returns an instance of the DotEnv checker. It takes a function
// which returns an io.Reader which will be called when the first value
// is recalled. The contents of the io.Reader MUST follow the DotEnv format.
func NewDotEnv(gen func() (io.Reader, error)) *DotEnv {
	return &DotEnv{
		gen: gen,
	}
}

// NewDotEnvFromFile returns an instance of the DotEnv checker. It reads its
// data from a file which its location has been specified through the path
// parameter
func NewDotEnvFromFile(path string) *DotEnv {
	return NewDotEnv(func() (io.Reader, error) {
		return os.Open(path)
	})
}

// DotEnv represents the DotEnv Checker. It reads an io.Reader and then pulls a value out of a map[string]interface{}.
type DotEnv struct {
	values map[string]interface{}
	gen    func() (io.Reader, error)
}

// Setup initializes the DotEnv Checker
func (h *DotEnv) Setup() error {
	r, err := h.gen()
	if err != nil {
		return err
	}

	h.values = make(map[string]interface{})
	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	for _, fullLine := range lines {
		if !isIgnoredLine(fullLine) {
			key, value, err := parseLine(fullLine)

			if err == nil {
				h.values[key] = value
			}
		}
	}

	return nil
}

func parseLine(line string) (key string, value string, err error) {
	if len(line) == 0 {
		err = errors.New("zero length string")
		return
	}

	// ditch the comments (but keep quoted hashes)
	if strings.Contains(line, "#") {
		segmentsBetweenHashes := strings.Split(line, "#")
		quotesAreOpen := false
		var segmentsToKeep []string
		for _, segment := range segmentsBetweenHashes {
			if strings.Count(segment, "\"") == 1 || strings.Count(segment, "'") == 1 {
				if quotesAreOpen {
					quotesAreOpen = false
					segmentsToKeep = append(segmentsToKeep, segment)
				} else {
					quotesAreOpen = true
				}
			}

			if len(segmentsToKeep) == 0 || quotesAreOpen {
				segmentsToKeep = append(segmentsToKeep, segment)
			}
		}

		line = strings.Join(segmentsToKeep, "#")
	}

	// now split key from value
	splitString := strings.SplitN(line, "=", 2)

	if len(splitString) != 2 {
		// try yaml mode!
		splitString = strings.SplitN(line, ":", 2)
	}

	if len(splitString) != 2 {
		err = errors.New("Can't separate key from value")
		return
	}

	// Parse the key
	key = splitString[0]
	if strings.HasPrefix(key, "export") {
		key = strings.TrimPrefix(key, "export")
	}
	key = strings.Trim(key, " ")

	// Parse the value
	value = splitString[1]
	// trim
	value = strings.Trim(value, " ")

	// check if we've got quoted values
	if strings.Count(value, "\"") == 2 || strings.Count(value, "'") == 2 {
		// pull the quotes off the edges
		value = strings.Trim(value, "\"'")

		// expand quotes
		value = strings.Replace(value, "\\\"", "\"", -1)
		// expand newlines
		value = strings.Replace(value, "\\n", "\n", -1)
	}

	return
}

func isIgnoredLine(line string) bool {
	trimmedLine := strings.Trim(line, " \n\t")
	return len(trimmedLine) == 0 || strings.HasPrefix(trimmedLine, "#")
}

func (h *DotEnv) value(name string) (interface{}, error) {
	val, ok := h.values[name]
	if !ok {
		return nil, errors.New("variable does not exist")
	}

	return val, nil
}

// Int returns an int if it exists within the DotEnv io.Reader
func (h *DotEnv) Int(name string) (int, error) {
	v, err := h.value(name)
	if err != nil {
		return 0, err
	}

	f, err := strconv.ParseFloat(v.(string), 64)

	if err != nil {
		i, err := strconv.ParseInt(v.(string), 10, 64)
		if err != nil {
			return v.(int), errors.New(fmt.Sprintf("%T unable", v))
		}

		return int(i), nil
	}
	return int(f), nil
}

// Bool returns a bool if it exists within the HCL io.Reader.
func (h *DotEnv) Bool(name string) (bool, error) {
	v, err := h.value(name)
	if err != nil {
		return false, err
	}

	b, err := strconv.ParseBool(v.(string))
	return b, err
}

// String returns a string if it exists within the HCL io.Reader.
func (h *DotEnv) String(name string) (string, error) {
	v, err := h.value(name)
	if err != nil {
		return "", err
	}

	return v.(string), nil
}
