package json_parser

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

const (
	stringDelim  = "\""
	kvDelim      = ":"
	start        = "{"
	end          = "}"
	arrayStart   = "["
	arrayEnd     = "]"
	elementDelim = ","
	space        = " "
	newLine      = "\n"
)

func isValidJsonFile(fileName string) bool {
	content := getContent(fileName)
	tokens := tokenize(content)
	return len(tokens) >= 2 &&
		tokens[0] == start &&
		tokens[len(tokens)-1] == end &&
		isValidKeyValuePair(tokens)
}

func getContent(fileName string) string {
	content := ""
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return ""
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		content += scanner.Text()
	}
	return content
}

func isSpecialToken(token string) bool {
	values := []string{kvDelim, start, end, elementDelim, arrayStart, arrayEnd}
	for _, val := range values {
		if token == val {
			return true
		}
	}
	return false
}

func isValidKey(token string) bool {
	if isSpecialToken(token) {
		return false
	}
	return token[0] == '"' && token[len(token)-1] == '"'
}

func isValidValue(token string) bool {
	if isSpecialToken(token) {
		return false
	}
	// case string
	if token[0] == '"' && token[len(token)-1] == '"' {
		return true
	}
	// case number
	if unicode.IsDigit(rune(token[0])) {
		if string(token[0]) == "0" {
			return false
		}
		_, err := strconv.Atoi(token)
		return err == nil
	}

	// case array
	if string(token[0]) == arrayStart && string(token[len(token)-1]) == arrayEnd {
		// split array by values
		values := strings.Split(token[1:len(token)-1], elementDelim)
		for _, value := range values {
			if value != "" && !(isValidValue(value)) {
				return false
			}
		}
		return true
	}

	return isBoolean(token) || isNull(token)
}

func isBoolean(token string) bool {
	return token == "true" || token == "false"
}

func isNull(token string) bool {
	return token == "null"
}

func isValidKeyValuePair(tokens []string) bool {
	if len(tokens) == 2 {
		return true
	}
	for i := 1; i < len(tokens)-1; i++ {
		// validate key
		if (i%4 == 1 && !isValidKey(tokens[i])) ||

			// validate delimiter
			(i%4 == 2 && (tokens[i] != kvDelim)) ||

			// validate value
			// check if is an inner object
			(i%4 == 3 && (string(tokens[i][0]) == start) && !isValidKeyValuePair(tokenize(tokens[i]))) ||
			(i%4 == 3 && (string(tokens[i][0]) != start) && !isValidValue(tokens[i])) ||

			// validate element delimiter
			(i%4 == 0 && tokens[i] != elementDelim) ||
			// edge case where last token is a comma
			(i%4 == 0 && (i+4 > len(tokens)-1) && tokens[i] == elementDelim) {
			return false
		}
	}
	return true
}

func tokenize(content string) []string {
	var tokens = make([]string, 0, 100)
	parsingString := false
	firstStart := true
	parsingInnerObject := false
	parsingArray := false

	token := ""
	for _, ch := range content {
		char := string(ch)

		if parsingInnerObject == true {
			if char == end {
				parsingInnerObject = false
				token += char
				tokens = append(tokens, token)
				token = ""
			} else if char == space || char == newLine {
				continue
			} else {
				token += char
			}
		} else if parsingArray {
			if char == arrayEnd {
				parsingArray = false
				token += char
				tokens = append(tokens, token)
				token = ""
			} else {
				token += char
			}
		} else if char == stringDelim && !parsingString {
			parsingString = true
			token += stringDelim
			continue
			// append to string
		} else if parsingString && char != stringDelim {
			token += char
			// finished parsing string
		} else if char == stringDelim && parsingString {
			parsingString = false
			token += stringDelim
			tokens = append(tokens, token)
			token = ""
		} else if char == space || char == newLine {
			continue
		} else if !isSpecialToken(char) {
			token += char
		} else if isSpecialToken(char) {
			// if first start
			if char == start && firstStart {
				tokens = append(tokens, char)
				firstStart = false
			} else if char == start && !firstStart {
				parsingInnerObject = true
				token += char
			} else if char == arrayStart {
				parsingArray = true
				token += char
			} else if token != "" {
				tokens = append(tokens, token)
				token = ""
				tokens = append(tokens, char)
			} else {
				tokens = append(tokens, char)
			}
		}
	}
	return tokens
}
