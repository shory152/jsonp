package jsonp

import (
	"bytes"
	"fmt"
)

// json token
const (
	jsonTkObjBegin int = iota // '{'
	jsonTkObjEnd              // '}'
	jsonTkAryBegin            // '['
	jsonTkAryEnd              // ']'
	jsonTkColon               // ':'
	jsonTkComma               // ','
	jsonTkString              // "xxx"
	jsonTkNumber              // 123.456
	jsonTkTrue                // true
	jsonTkFalse               // false
	jsonTkNull                // null
)

type JsonToken struct {
	ID  int
	Val string
}

type JsonScanner func() (JsonToken, error)

func isBlank(c rune) bool {
	switch c {
	case ' ', '\t', '\n', '\r':
		return true
	default:
		return false
	}
}

func isDigit(c rune) bool {
	switch c {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true
	default:
		return false
	}
}

func isHexDigit(c rune) bool {
	switch c {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true
	case 'a', 'b', 'c', 'd', 'e', 'f':
		return true
	case 'A', 'B', 'C', 'D', 'E', 'F':
		return true
	default:
		return false
	}
}

func ScanJson(json string) JsonScanner {
	jr := bytes.NewReader([]byte(json))
	var lastTk JsonToken
	var lastErr error

	return JsonScanner(func() (JsonToken, error) {
		if lastErr != nil {
			return lastTk, lastErr
		}

		for {
			if c, _, err := jr.ReadRune(); err != nil {
				lastErr = err
				break
			} else {
				if isBlank(c) {
					continue
				}
				switch c {
				case '{':
					lastTk = JsonToken{jsonTkObjBegin, "{"}
					return lastTk, nil
				case '}':
					lastTk = JsonToken{jsonTkObjEnd, "}"}
					return lastTk, nil
				case '[':
					lastTk = JsonToken{jsonTkAryBegin, "["}
					return lastTk, nil
				case ']':
					lastTk = JsonToken{jsonTkAryEnd, "]"}
					return lastTk, nil
				case ':':
					lastTk = JsonToken{jsonTkColon, ":"}
					return lastTk, nil
				case ';':
					lastTk = JsonToken{jsonTkComma, ";"}
					return lastTk, nil
				case '"':
					jr.UnreadRune()
					lastTk, lastErr = scanString(jr)
					return lastTk, lastErr
				case 'f':
					jr.UnreadRune()
					lastTk, lastErr = scanFalse(jr)
					return lastTk, lastErr
				case 't':
					jr.UnreadRune()
					lastTk, lastErr = scanTrue(jr)
					return lastTk, lastErr
				case 'n':
					jr.UnreadRune()
					lastTk, lastErr = scanNull(jr)
					return lastTk, lastErr
				}

				if isDigit(c) {
					jr.UnreadRune()
					lastTk, lastErr = scanNumber(jr)
					return lastTk, lastErr
				}

				lastErr = syntaxErr(jr)
				return lastTk, lastErr
			}
		} // for

		return lastTk, lastErr
	})
}

func scanString(jr *bytes.Reader) (tk JsonToken, err error) {
	var str bytes.Buffer
s1:
	if c, _, e := jr.ReadRune(); e != nil {
		return tk, e
	} else {
		switch c {
		case '"':
			str.WriteRune(c)
			tk = JsonToken{jsonTkString, str.String()}
			return
		case '\\':
			str.WriteRune(c)
			goto s2
		default:
			str.WriteRune(c)
			goto s1
		}
	}

s2:
	if c, _, e := jr.ReadRune(); e != nil {
		return tk, e
	} else {
		switch c {
		case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
			str.WriteRune(c)
			goto s1
		case 'u':
			str.WriteRune(c)
			for i := 0; i < 4; i++ {
				if c, _, e := jr.ReadRune(); e != nil {
					return tk, e
				} else if !isHexDigit(c) {
					return tk, syntaxErr(jr)
				} else {
					str.WriteRune(c)
				}
			}
			goto s1
		}
	}

	return
}

func scanNumber(jr *bytes.Reader) (tk JsonToken, err error) {
	return
}

func scanTrue(jr *bytes.Reader) (tk JsonToken, err error) {
	if err = scanGivenStr(jr, "true"); err != nil {
		return
	}
	return JsonToken{jsonTkTrue, "true"}, nil
}

func scanFalse(jr *bytes.Reader) (tk JsonToken, err error) {
	if err = scanGivenStr(jr, "false"); err != nil {
		return
	}
	return JsonToken{jsonTkFalse, "false"}, nil
}

func scanNull(jr *bytes.Reader) (tk JsonToken, err error) {
	if err = scanGivenStr(jr, "null"); err != nil {
		return
	}
	return JsonToken{jsonTkNull, "null"}, nil
}

func scanGivenStr(jr *bytes.Reader, givestr string) error {
	for givec := range givestr {
		if c, _, err := jr.ReadRune(); err != nil {
			return err
		} else if c != rune(givec) {
			return syntaxErr(jr)
		}
	}
	return nil
}

func syntaxErr(jr *bytes.Reader) error {
	pos := int(jr.Size()) - jr.Len()
	return fmt.Errorf("json syntax error: at %d", pos)
}
