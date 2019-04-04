package jsonp

import (
	"bytes"
	"errors"
)

type jsonScan struct {
	jr      *bytes.Reader
	lastErr error
	lastTk  JsonToken
	val     bytes.Buffer // tmp use for store char
	q       [4]JsonToken
	qh      int
	qt      int
}

func ScanJson3(json string) *jsonScan {
	jscan := &jsonScan{}
	jscan.jr = bytes.NewReader([]byte(json))
	jscan.val.Grow(32)
	return jscan
}

var errNoToken error = errors.New("no token has been scanned")

func (js *jsonScan) LastTk() (JsonToken, error) {
	if js.lastErr != nil {
		return js.lastTk, js.lastErr
	}
	if js.lastTk.ID == jsonTkUnknow {
		return js.lastTk, errNoToken
	}
	return js.lastTk, nil
}

func (js *jsonScan) UnScan() (err error) {
	if js.qh == js.qt {
		return errors.New("no previous token")
	}
	js.qh--
	if js.qh < 0 {
		js.qh = len(js.q) - 1
	}
	return
}

func (js *jsonScan) Scan() (JsonToken, error) {
	if js.lastErr != nil {
		return js.lastTk, js.lastErr
	}

	var off int
	js.val.Reset()

s_start:
	off = int(js.jr.Size()) - js.jr.Len()
	if c, _, err := js.jr.ReadRune(); err != nil {
		js.lastErr = err
		goto s_return
	} else {
		switch c {
		case ' ', '\t', '\n', '\r':
			goto s_start
		}

		switch c {
		case '{':
			js.lastTk = JsonToken{jsonTkObjBegin, off, "{"}
			goto s_return
		case '}':
			js.lastTk = JsonToken{jsonTkObjEnd, off, "}"}
			goto s_return
		case '[':
			js.lastTk = JsonToken{jsonTkAryBegin, off, "["}
			goto s_return
		case ']':
			js.lastTk = JsonToken{jsonTkAryEnd, off, "]"}
			goto s_return
		case ':':
			js.lastTk = JsonToken{jsonTkColon, off, ":"}
			goto s_return
		case ',':
			js.lastTk = JsonToken{jsonTkComma, off, ","}
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
			js.val.WriteRune(c)
			goto s_n2
		case '0':
			js.val.WriteRune(c)
			goto s_n3
		case '1', '2', '3', '4', '5', '6', '7', '8', '9': // number
			js.val.WriteRune(c)
			goto s_n9
		}

		js.lastErr = syntaxErr(js.jr)
		goto s_return
	}
	panic("can not reach here")

s_false:
	for _, gc := range falseR {
		if c, _, err := js.jr.ReadRune(); err != nil {
			js.lastErr = err
			goto s_return
		} else if c != gc {
			js.lastErr = syntaxErr(js.jr)
			goto s_return
		}
	}
	js.lastTk = JsonToken{jsonTkFalse, off, "false"}
	goto s_return
	panic("can not reach here")

s_true:
	for _, gc := range trueR {
		if c, _, err := js.jr.ReadRune(); err != nil {
			js.lastErr = err
			goto s_return
		} else if c != gc {
			js.lastErr = syntaxErr(js.jr)
			goto s_return
		}
	}
	js.lastTk = JsonToken{jsonTkTrue, off, "true"}
	goto s_return
	panic("can not reach here")

s_null:
	for _, gc := range nullR {
		if c, _, err := js.jr.ReadRune(); err != nil {
			js.lastErr = err
			goto s_return
		} else if c != gc {
			js.lastErr = syntaxErr(js.jr)
			goto s_return
		}
	}
	js.lastTk = JsonToken{jsonTkNull, off, "null"}
	goto s_return
	panic("can not reach here")

s_s1:
	if c, _, err := js.jr.ReadRune(); err != nil {
		js.lastErr = err
		goto s_return
	} else {
		switch c {
		case '"':
			//val.WriteRune(c)
			js.lastTk = JsonToken{jsonTkString, off, js.val.String()}
			goto s_return
		case '\\':
			js.val.WriteRune(c)
			goto s_s2
		default:
			js.val.WriteRune(c)
			goto s_s1
		}
	}
	panic("can not reach here")

s_s2:
	if c, _, err := js.jr.ReadRune(); err != nil {
		js.lastErr = err
		goto s_return
	} else {
		switch c {
		case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
			js.val.WriteRune(c)
			goto s_s1
		case 'u':
			js.val.WriteRune(c)
			for i := 0; i < 4; i++ {
				if c, _, err := js.jr.ReadRune(); err != nil {
					js.lastErr = err
					goto s_return
				} else if !isHexDigit(c) {
					js.lastErr = syntaxErr(js.jr)
					goto s_return
				} else {
					js.val.WriteRune(c)
				}
			}
			goto s_s1
		default:
			js.lastErr = syntaxErr(js.jr)
			goto s_return
		}
	}
	panic("can not reach here")

s_n2:
	if c, _, err := js.jr.ReadRune(); err != nil {
		js.lastErr = err
		goto s_return
	} else {
		switch c {
		case '0':
			js.val.WriteRune(c)
			goto s_n3
		case '1', '2', '3', '4', '5', '6', '7', '8', '9':
			js.val.WriteRune(c)
			goto s_n9
		default:
			js.lastErr = syntaxErr(js.jr)
			goto s_return
		}
	}
	panic("can not reach here")

s_n3:
	if c, _, err := js.jr.ReadRune(); err != nil {
		js.lastErr = err
		goto s_return
	} else {
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			js.lastErr = syntaxErr(js.jr)
			goto s_return
		case '.':
			js.val.WriteRune(c)
			goto s_n4
		default: // output 0
			js.jr.UnreadRune()
			goto s_nn
		}
	}
	panic("can not reach here")

s_n4:
	if c, _, err := js.jr.ReadRune(); err != nil {
		js.lastErr = err
		goto s_return
	} else {
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			js.val.WriteRune(c)
			goto s_n5
		default:
			js.lastErr = syntaxErr(js.jr)
			goto s_return
		}
	}
	panic("can not reach here")

s_n5:
	if c, _, err := js.jr.ReadRune(); err != nil {
		js.lastErr = err
		goto s_return
	} else {
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			js.val.WriteRune(c)
			goto s_n5
		case 'e', 'E':
			js.val.WriteRune(c)
			goto s_n6
		default:
			js.jr.UnreadRune()
			goto s_nn
		}
	}
	panic("can not reach here")

s_n6:
	if c, _, err := js.jr.ReadRune(); err != nil {
		js.lastErr = err
		goto s_return
	} else {
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			js.val.WriteRune(c)
			goto s_n8
		case '+', '-':
			js.val.WriteRune(c)
			goto s_n7
		default:
			js.lastErr = syntaxErr(js.jr)
			goto s_return
		}
	}
	panic("can not reach here")

s_n7:
	if c, _, err := js.jr.ReadRune(); err != nil {
		js.lastErr = err
		goto s_return
	} else {
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			js.val.WriteRune(c)
			goto s_n8
		default:
			js.lastErr = syntaxErr(js.jr)
			goto s_return
		}
	}
	panic("can not reach here")

s_n8:
	if c, _, err := js.jr.ReadRune(); err != nil {
		js.lastErr = err
		goto s_return
	} else {
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			js.val.WriteRune(c)
			goto s_n8
		default:
			js.jr.UnreadRune()
			goto s_nn
		}
	}
	panic("can not reach here")

s_n9:
	if c, _, err := js.jr.ReadRune(); err != nil {
		js.lastErr = err
		goto s_return
	} else {
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			js.val.WriteRune(c)
			goto s_n9
		case '.':
			js.val.WriteRune(c)
			goto s_n4
		default:
			js.jr.UnreadRune()
			goto s_nn
		}
	}
	panic("can not reach here")

s_nn:
	js.lastTk = JsonToken{jsonTkNumber, off, js.val.String()}
	goto s_return

s_return:
	return js.lastTk, js.lastErr
}
