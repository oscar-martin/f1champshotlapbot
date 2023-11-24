package helper

import (
	"fmt"
	"hash/fnv"
	"strings"
)

// method to convert from seconds to minutes:seconds:milliseconds
func SecondsToMinutes(seconds float64) string {
	if seconds <= 0 {
		return "-"
	}
	minutes := int(seconds / 60)
	seconds = seconds - float64(minutes*60)
	milliseconds := int((seconds - float64(int(seconds))) * 1000)
	return fmt.Sprintf("%02d:%02d.%03d", minutes, int(seconds), milliseconds)
}

func SecondsToDiff(seconds float64) string {
	if seconds <= 0 {
		return "-"
	}
	diff := fmt.Sprintf("%.3fs", seconds)
	chars := len(diff)
	if chars < 9 {
		// add spaces to the left
		diff = strings.Repeat(" ", 9-chars) + diff
	}
	return diff
}

func SecondsToHoursAndMinutes(seconds float64) string {
	if seconds <= 0 {
		seconds = 0
	}
	hours := int(seconds / 3600)
	seconds = seconds - float64(hours*3600)
	minutes := int(seconds / 60)
	return fmt.Sprintf("%02dh %02dm", hours, minutes)
}

// method to convert to seconds and 3 milliseconds
func ToSectorTime(t float64) string {
	if t <= 0 {
		return "-"
	}
	return fmt.Sprintf("%.3f", t)
}

func GetDriverCodeName(name string) string {
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
		if len(words[1]) > 2 {
			code += words[1][:2]
		} else {
			// if the second word is only 1 letter long, get the first 3 letters of the first word
			code += words[1]
		}
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

// convert name to a hash with a limit of 15 characters
func ToID(name string) string {
	h := fnv.New32a()
	h.Write([]byte(name))
	return fmt.Sprint(h.Sum32())
}
