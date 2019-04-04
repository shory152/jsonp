package jsonp

import (
	"bytes"
	"fmt"
)

// json token
const (
	jsonTkUnknow   int = iota // invalid token
	jsonTkObjBegin            // '{'
	jsonTkObjEnd              // '}'
	jsonTkAryBegin            // '['
	jsonTkAryEnd              // ']'
	jsonTkColon               // ':'
	jsonTkComma               // ','
	jsonTkString              // "xxx"
	jsonTkNumber              // 123.456, -0.1234, 1.2e-3
	jsonTkTrue                // true
	jsonTkFalse               // false
	jsonTkNull                // null
)

type JsonToken struct {
	ID     int
	OffSet int
	Val    string
}

type JsonScanner func() (JsonToken, error)

func (s JsonScanner) Scan() (JsonToken, error) {
	return s()
}

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
	var val bytes.Buffer
	val.Grow(32)

	return JsonScanner(func() (JsonToken, error) {
		lastTk = JsonToken{}
		if lastErr != nil {
			return lastTk, lastErr
		}
		val.Reset()

		for {
			off := int(jr.Size()) - jr.Len()
			if c, _, err := jr.ReadRune(); err != nil {
				lastErr = err
				break
			} else {
				if isBlank(c) {
					continue
				}
				switch c {
				case '{':
					lastTk = JsonToken{jsonTkObjBegin, off, "{"}
					return lastTk, nil
				case '}':
					lastTk = JsonToken{jsonTkObjEnd, off, "}"}
					return lastTk, nil
				case '[':
					lastTk = JsonToken{jsonTkAryBegin, off, "["}
					return lastTk, nil
				case ']':
					lastTk = JsonToken{jsonTkAryEnd, off, "]"}
					return lastTk, nil
				case ':':
					lastTk = JsonToken{jsonTkColon, off, ":"}
					return lastTk, nil
				case ',':
					lastTk = JsonToken{jsonTkComma, off, ","}
					return lastTk, nil
				case '"': // string "xxx"
					jr.UnreadRune()
					lastTk, lastErr = scanString(jr, &val)
					return lastTk, lastErr
				case 'f': // false
					jr.UnreadRune()
					lastTk, lastErr = scanFalse(jr)
					return lastTk, lastErr
				case 't': // true
					jr.UnreadRune()
					lastTk, lastErr = scanTrue(jr)
					return lastTk, lastErr
				case 'n': // null
					jr.UnreadRune()
					lastTk, lastErr = scanNull(jr)
					return lastTk, lastErr
				case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9': // number
					jr.UnreadRune()
					lastTk, lastErr = scanNumber(jr, &val)
					return lastTk, lastErr
				}

				lastErr = syntaxErr(jr)
				return lastTk, lastErr
			}
		} // for

		return lastTk, lastErr
	})
}

// scan "abc\n\\akdk ' \"dfkad"
func scanString(jr *bytes.Reader, str *bytes.Buffer) (tk JsonToken, err error) {
	off := int(jr.Size()) - jr.Len()

	if c, _, e := jr.ReadRune(); e != nil {
		return tk, e
	} else if c == '"' {
		//str.WriteRune(c)
		goto s1
	} else {
		return tk, syntaxErr(jr)
	}

s1:
	if c, _, e := jr.ReadRune(); e != nil {
		return tk, e
	} else {
		switch c {
		case '"':
			//str.WriteRune(c)
			tk = JsonToken{jsonTkString, off, str.String()}
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
		default:
			return tk, syntaxErr(jr)
		}
	}

	return
}

func scanNumber(jr *bytes.Reader, str *bytes.Buffer) (tk JsonToken, err error) {
	off := int(jr.Size()) - jr.Len()

	if c, _, e := jr.ReadRune(); e != nil {
		return tk, e
	} else {
		switch c {
		case '-':
			str.WriteRune(c)
			goto n2
		case '0':
			str.WriteRune(c)
			goto n3
		case '1', '2', '3', '4', '5', '6', '7', '8', '9':
			str.WriteRune(c)
			goto n9
		default:
			return tk, syntaxErr(jr)
		}
	}
n2:
	if c, _, e := jr.ReadRune(); e != nil {
		return tk, e
	} else {
		switch c {
		case '0':
			str.WriteRune(c)
			goto n3
		case '1', '2', '3', '4', '5', '6', '7', '8', '9':
			str.WriteRune(c)
			goto n9
		default:
			return tk, syntaxErr(jr)
		}
	}
n3:
	if c, _, e := jr.ReadRune(); e != nil {
		return tk, e
	} else {
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			return tk, syntaxErr(jr)
		case '.':
			str.WriteRune(c)
			goto n4
		default: // output 0
			jr.UnreadRune()
			goto nn
		}
	}
n4:
	if c, _, e := jr.ReadRune(); e != nil {
		return tk, e
	} else {
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			str.WriteRune(c)
			goto n5
		default:
			return tk, syntaxErr(jr)
		}
	}
n5:
	if c, _, e := jr.ReadRune(); e != nil {
		return tk, e
	} else {
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			str.WriteRune(c)
			goto n5
		case 'e', 'E':
			str.WriteRune(c)
			goto n6
		default:
			jr.UnreadRune()
			goto nn
		}
	}
n6:
	if c, _, e := jr.ReadRune(); e != nil {
		return tk, e
	} else {
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			str.WriteRune(c)
			goto n8
		case '+', '-':
			str.WriteRune(c)
			goto n7
		default:
			return tk, syntaxErr(jr)
		}
	}
n7:
	if c, _, e := jr.ReadRune(); e != nil {
		return tk, e
	} else {
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			str.WriteRune(c)
			goto n8
		default:
			return tk, syntaxErr(jr)
		}
	}
n8:
	if c, _, e := jr.ReadRune(); e != nil {
		return tk, e
	} else {
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			str.WriteRune(c)
			goto n8
		default:
			jr.UnreadRune()
			goto nn
		}
	}
n9:
	if c, _, e := jr.ReadRune(); e != nil {
		return tk, e
	} else {
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			str.WriteRune(c)
			goto n9
		case '.':
			str.WriteRune(c)
			goto n4
		default:
			jr.UnreadRune()
			goto nn
		}
	}

nn: // return number
	tk = JsonToken{jsonTkNumber, off, str.String()}
	return tk, nil
}

func scanTrue(jr *bytes.Reader) (tk JsonToken, err error) {
	off := int(jr.Size()) - jr.Len()
	if err = scanGivenStr(jr, "true"); err != nil {
		return
	}
	return JsonToken{jsonTkTrue, off, "true"}, nil
}

func scanFalse(jr *bytes.Reader) (tk JsonToken, err error) {
	off := int(jr.Size()) - jr.Len()
	if err = scanGivenStr(jr, "false"); err != nil {
		return
	}
	return JsonToken{jsonTkFalse, off, "false"}, nil
}

func scanNull(jr *bytes.Reader) (tk JsonToken, err error) {
	off := int(jr.Size()) - jr.Len()
	if err = scanGivenStr(jr, "null"); err != nil {
		return
	}
	return JsonToken{jsonTkNull, off, "null"}, nil
}

func scanGivenStr(jr *bytes.Reader, givestr string) error {
	for _, givec := range givestr {
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
	jr.UnreadRune()
	var bef bytes.Buffer
	for i := 0; i < 16; i++ {
		if c, _, err := jr.ReadRune(); err == nil {
			bef.WriteRune(c)
		}
	}
	return fmt.Errorf("json syntax error: at %d, %s", pos, bef.String())
}
