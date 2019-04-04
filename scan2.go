package jsonp

import "bytes"

var (
	falseR [4]rune = [...]rune{'a', 'l', 's', 'e'}
	trueR  [3]rune = [...]rune{'r', 'u', 'e'}
	nullR  [3]rune = [...]rune{'u', 'l', 'l'}
)

func ScanJson2(json string) JsonScanner {
	jr := bytes.NewReader([]byte(json))
	var lastTk JsonToken
	var lastErr error
	var off int
	var val bytes.Buffer
	val.Grow(32)

	return JsonScanner(func() (JsonToken, error) {
		lastTk = JsonToken{}
		if lastErr != nil {
			return lastTk, lastErr
		}
		val.Reset()

	s_start:
		off = int(jr.Size()) - jr.Len()
		if c, _, err := jr.ReadRune(); err != nil {
			lastErr = err
			goto s_return
		} else {
			switch c {
			case ' ', '\t', '\n', '\r':
				goto s_start
			}

			switch c {
			case '{':
				lastTk = JsonToken{jsonTkObjBegin, off, "{"}
				goto s_return
			case '}':
				lastTk = JsonToken{jsonTkObjEnd, off, "}"}
				goto s_return
			case '[':
				lastTk = JsonToken{jsonTkAryBegin, off, "["}
				goto s_return
			case ']':
				lastTk = JsonToken{jsonTkAryEnd, off, "]"}
				goto s_return
			case ':':
				lastTk = JsonToken{jsonTkColon, off, ":"}
				goto s_return
			case ',':
				lastTk = JsonToken{jsonTkComma, off, ","}
				goto s_return
			case '"': // string "xxx"
				goto s_s1
			case 'f': // false
				goto s_false
			case 't': // true
				goto s_true
			case 'n': // null
				goto s_null
			case '-':
				val.WriteRune(c)
				goto s_n2
			case '0':
				val.WriteRune(c)
				goto s_n3
			case '1', '2', '3', '4', '5', '6', '7', '8', '9': // number
				val.WriteRune(c)
				goto s_n9
			}

			lastErr = syntaxErr(jr)
			goto s_return
		}
		panic("can not reach here")

	s_false:
		for _, gc := range falseR {
			if c, _, err := jr.ReadRune(); err != nil {
				lastErr = err
				goto s_return
			} else if c != gc {
				lastErr = syntaxErr(jr)
				goto s_return
			}
		}
		lastTk = JsonToken{jsonTkFalse, off, "false"}
		goto s_return
		panic("can not reach here")

	s_true:
		for _, gc := range trueR {
			if c, _, err := jr.ReadRune(); err != nil {
				lastErr = err
				goto s_return
			} else if c != gc {
				lastErr = syntaxErr(jr)
				goto s_return
			}
		}
		lastTk = JsonToken{jsonTkTrue, off, "true"}
		goto s_return
		panic("can not reach here")

	s_null:
		for _, gc := range nullR {
			if c, _, err := jr.ReadRune(); err != nil {
				lastErr = err
				goto s_return
			} else if c != gc {
				lastErr = syntaxErr(jr)
				goto s_return
			}
		}
		lastTk = JsonToken{jsonTkNull, off, "null"}
		goto s_return
		panic("can not reach here")

	s_s1:
		if c, _, err := jr.ReadRune(); err != nil {
			lastErr = err
			goto s_return
		} else {
			switch c {
			case '"':
				//val.WriteRune(c)
				lastTk = JsonToken{jsonTkString, off, val.String()}
				goto s_return
			case '\\':
				val.WriteRune(c)
				goto s_s2
			default:
				val.WriteRune(c)
				goto s_s1
			}
		}
		panic("can not reach here")

	s_s2:
		if c, _, err := jr.ReadRune(); err != nil {
			lastErr = err
			goto s_return
		} else {
			switch c {
			case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
				val.WriteRune(c)
				goto s_s1
			case 'u':
				val.WriteRune(c)
				for i := 0; i < 4; i++ {
					if c, _, err := jr.ReadRune(); err != nil {
						lastErr = err
						goto s_return
					} else if !isHexDigit(c) {
						lastErr = syntaxErr(jr)
						goto s_return
					} else {
						val.WriteRune(c)
					}
				}
				goto s_s1
			default:
				lastErr = syntaxErr(jr)
				goto s_return
			}
		}
		panic("can not reach here")

	s_n2:
		if c, _, err := jr.ReadRune(); err != nil {
			lastErr = err
			goto s_return
		} else {
			switch c {
			case '0':
				val.WriteRune(c)
				goto s_n3
			case '1', '2', '3', '4', '5', '6', '7', '8', '9':
				val.WriteRune(c)
				goto s_n9
			default:
				lastErr = syntaxErr(jr)
				goto s_return
			}
		}
		panic("can not reach here")

	s_n3:
		if c, _, err := jr.ReadRune(); err != nil {
			lastErr = err
			goto s_return
		} else {
			switch c {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				lastErr = syntaxErr(jr)
				goto s_return
			case '.':
				val.WriteRune(c)
				goto s_n4
			default: // output 0
				jr.UnreadRune()
				goto s_nn
			}
		}
		panic("can not reach here")

	s_n4:
		if c, _, err := jr.ReadRune(); err != nil {
			lastErr = err
			goto s_return
		} else {
			switch c {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				val.WriteRune(c)
				goto s_n5
			default:
				lastErr = syntaxErr(jr)
				goto s_return
			}
		}
		panic("can not reach here")

	s_n5:
		if c, _, err := jr.ReadRune(); err != nil {
			lastErr = err
			goto s_return
		} else {
			switch c {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				val.WriteRune(c)
				goto s_n5
			case 'e', 'E':
				val.WriteRune(c)
				goto s_n6
			default:
				jr.UnreadRune()
				goto s_nn
			}
		}
		panic("can not reach here")

	s_n6:
		if c, _, err := jr.ReadRune(); err != nil {
			lastErr = err
			goto s_return
		} else {
			switch c {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				val.WriteRune(c)
				goto s_n8
			case '+', '-':
				val.WriteRune(c)
				goto s_n7
			default:
				lastErr = syntaxErr(jr)
				goto s_return
			}
		}
		panic("can not reach here")

	s_n7:
		if c, _, err := jr.ReadRune(); err != nil {
			lastErr = err
			goto s_return
		} else {
			switch c {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				val.WriteRune(c)
				goto s_n8
			default:
				lastErr = syntaxErr(jr)
				goto s_return
			}
		}
		panic("can not reach here")

	s_n8:
		if c, _, err := jr.ReadRune(); err != nil {
			lastErr = err
			goto s_return
		} else {
			switch c {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				val.WriteRune(c)
				goto s_n8
			default:
				jr.UnreadRune()
				goto s_nn
			}
		}
		panic("can not reach here")

	s_n9:
		if c, _, err := jr.ReadRune(); err != nil {
			lastErr = err
			goto s_return
		} else {
			switch c {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				val.WriteRune(c)
				goto s_n9
			case '.':
				val.WriteRune(c)
				goto s_n4
			default:
				jr.UnreadRune()
				goto s_nn
			}
		}
		panic("can not reach here")

	s_nn:
		lastTk = JsonToken{jsonTkNumber, off, val.String()}
		goto s_return

	s_return:
		return lastTk, lastErr
	})
}
