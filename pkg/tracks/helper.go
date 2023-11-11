package tracks

import (
	"fmt"
	"strings"
)

// method to convert from seconds to minutes:seconds:milliseconds
func secondsToMinutes(seconds float64) string {
	minutes := int(seconds / 60)
	seconds = seconds - float64(minutes*60)
	milliseconds := int((seconds - float64(int(seconds))) * 1000)
	return fmt.Sprintf("%02d:%02d.%03d", minutes, int(seconds), milliseconds)
}

// method to convert to seconds and 3 milliseconds
func toSectorTime(t float64) string {
	return fmt.Sprintf("%.3f", t)
}

func getDriverCodeName(name string) string {
	// this function reads a name with possible surname and will return the first letter of the name and the first 3 letters of the surname
	// if the name is empty, it will return an empty string
	if name == "" {
		return ""
	}
	// split the name into words
	words := strings.Split(name, " ")
	// get the first letter of the first word
	code := string(words[0][0])
	// if there is a second word, get the first 2 letters of it
	if len(words) > 1 {
		code += words[1][:2]
	} else {
		// if there is no second word, get the first 2 letters of the first word
		if len(words[0]) > 2 {
			code += words[0][1:3]
		} else {
			// if the first word is only 1 letter long, get the first 3 letters of the first word
			code += words[0]
		}
	}
	return strings.ToUpper(code)
}
