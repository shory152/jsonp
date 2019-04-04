package jsonp

import (
	"errors"
	"fmt"
	"io"
	"strconv"
)

// json value type
const (
	jsonValUnknow int = iota
	jsonValObj
	jsonValAry
	jsonValStr
	jsonValNum
	jsonValFalse
	jsonValTrue
	jsonValNull
)

const (
	jsonElemVal int = jsonValNull + 1 + iota
	jsonElemKV
	jsonElemDoc
)

type JsonVal struct {
	ValTyp int
	Val    interface{}
}
type JsonNull string
type JsonTrue bool
type JsonFalse bool
type JsonNum float64
type JsonStr string
type JsonAry []JsonVal
type JsonKV struct {
	Key JsonStr
	Val JsonVal
}
type JsonObj []JsonKV
type JsonDoc []JsonVal // only JsonObj, JsonAry

func tk2val(tk JsonToken) JsonVal {
	switch tk.ID {
	case jsonTkObjBegin, jsonTkObjEnd, jsonTkAryBegin, jsonTkAryEnd,
		jsonTkColon, jsonTkComma:
		panic("can not convert json token to json value")
	case jsonTkString:
		return JsonVal{jsonValStr, JsonStr(tk.Val)}
	case jsonTkNumber:
		num, _ := strconv.ParseFloat(tk.Val, 64)
		return JsonVal{jsonValNum, JsonNum(num)}
	case jsonTkTrue:
		return JsonVal{jsonValTrue, JsonTrue(true)}
	case jsonTkFalse:
		return JsonVal{jsonValFalse, JsonFalse(false)}
	case jsonTkNull:
		return JsonVal{jsonValNull, JsonNull(tk.Val)}
	default:
		panic("unknow json token")
	}
}

func ParseJson(jsons string) (jdoc JsonDoc, oerr error) {
	scanner := ScanJson2(jsons)
	for {
		if jv, err := parseObjOrAry(scanner); err != nil {
			if err == io.EOF {
				return
			}
			return jdoc, err
		} else {
			jdoc = append(jdoc, jv)
		}
	}

	return
}

func parseObjOrAry(s JsonScanner) (JsonVal, error) {
	if tk, err := s(); err != nil {
		return JsonVal{}, err
	} else if tk.ID == jsonTkObjBegin {
		return parseObj(s)
	} else if tk.ID == jsonTkAryBegin {
		return parseAry(s)
	} else {
		return JsonVal{}, syntaxErr2(tk)
	}
}

var errJobjClosed error = errors.New("}")
var errJaryClosed error = errors.New("]")

func parseObj(scanner JsonScanner) (jobj JsonVal, oerr error) {
	var obj JsonObj
	for {
		if jkv, err := parseKV(scanner); err != nil {
			if err == errJobjClosed {
				break
			}
			return jobj, err
		} else {
			obj = append(obj, jkv)
		}
		// skip ,
		if tk, err := scanner(); err != nil {
			return jobj, err
		} else if tk.ID == jsonTkComma {
			continue
		} else if tk.ID == jsonTkObjEnd {
			break
		} else {
			return jobj, syntaxErr2(tk)
		}
	}
	jobj = JsonVal{jsonValObj, obj}
	return
}
func parseKV(s JsonScanner) (jkv JsonKV, oerr error) {
	var k JsonStr
	var v JsonVal
	// key
	if tk, err := s(); err != nil {
		return jkv, err
	} else if tk.ID == jsonTkObjEnd {
		return jkv, errJobjClosed
	} else if tk.ID != jsonTkString {
		return jkv, syntaxErr2(tk)
	} else {
		k = JsonStr(tk.Val)
	}
	// :
	if tk, err := s(); err != nil {
		return jkv, err
	} else if tk.ID != jsonTkColon {
		return jkv, syntaxErr2(tk)
	}
	// val
	if v, oerr = parseVal(s); oerr != nil {
		return
	}

	return JsonKV{k, v}, nil
}

func parseVal(s JsonScanner) (jval JsonVal, oerr error) {
	if tk, err := s(); err != nil {
		return jval, err
	} else {
		switch tk.ID {
		case jsonTkObjBegin:
			return parseObj(s)
		case jsonTkAryBegin:
			return parseAry(s)
		case jsonTkString, jsonTkNumber, jsonTkTrue, jsonTkFalse, jsonTkNull:
			jval = tk2val(tk)
			return
		case jsonTkObjEnd:
			return jval, errJobjClosed
		case jsonTkAryEnd:
			return jval, errJaryClosed
		default:
			return jval, syntaxErr2(tk)
		}
	}
}

// v1,v2,...,vn]
func parseAry(s JsonScanner) (jary JsonVal, oerr error) {
	var ary JsonAry
	for {
		if jval, err := parseVal(s); err != nil {
			if err == errJaryClosed {
				break
			}
			return jary, err
		} else {
			ary = append(ary, jval)
		}
		// skip ,
		if tk, err := s(); err != nil {
			return jary, err
		} else if tk.ID == jsonTkComma {
			continue
		} else if tk.ID == jsonTkAryEnd {
			break
		} else {
			return jary, syntaxErr2(tk)
		}
	}
	jary = JsonVal{jsonValAry, ary}
	return
}

func syntaxErr2(tk JsonToken) error {
	return fmt.Errorf("json syntax error: at %d, %v", tk.OffSet, tk.Val)
}
