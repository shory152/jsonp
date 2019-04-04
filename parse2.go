package jsonp

import "io"

var (
	jcTrue  JsonVal = JsonVal{jsonValTrue, JsonTrue(true)}
	jcFalse JsonVal = JsonVal{jsonValFalse, JsonFalse(false)}
	jcNull  JsonVal = JsonVal{jsonValNull, JsonNull("null")}
)

func ParseJson2(jsons string) (jdoc JsonDoc, oerr error) {
	sc := ScanJson3(jsons)
	stack := make([]*JsonVal, 64)
	stop := -1

	stop++
	stack[stop] = &JsonVal{jsonElemDoc, &jdoc}
	goto s_doc

s_doc:
	if tk, err := sc.Scan(); err != nil {
		if err != io.EOF {
			oerr = err
		}
		goto s_return
	} else {
		switch tk.ID {
		case jsonTkObjBegin:
			obj := JsonVal{jsonValObj, nil}
			doc := stack[stop].Val.(*JsonDoc)
			*doc = append(*doc, obj)
			stop++
			stack[stop] = &obj
			goto s_obj
		case jsonTkAryBegin:
			goto s_ary
		default:
			oerr = syntaxErr2(tk)
			goto s_return
		}
	}
	panic("s_jdoc end")

s_obj:
	if tk, err := sc.Scan(); err != nil {
		oerr = err
		goto s_return
	} else {
		switch tk.ID {
		case jsonTkObjEnd:
			stop--
			goto s_next
		case jsonTkString:
			rkv := JsonKV{Key: JsonStr(tk.Val)}
			obj := stack[stop].Val.(*JsonObj)
			*obj = append(*obj, rkv)
			// read :
			if tk, err := sc.Scan(); err != nil {
				oerr = err
				goto s_return
			} else if tk.ID != jsonTkColon {
				oerr = syntaxErr2(tk)
				goto s_return
			}

			kv := JsonVal{jsonElemKV, rkv}
			stop++
			stack[stop] = &kv
			goto s_val
		default:
			oerr = syntaxErr2(tk)
			goto s_return
		}
	}
	panic("s_obj")

s_ary:

s_val:
	if tk, err := sc.Scan(); err != nil {
		oerr = err
		goto s_return
	} else {
		switch tk.ID {
		case jsonTkNull, jsonTkTrue, jsonTkFalse, jsonTkString, jsonTkNumber:
			jv := tk2val(tk)
			if stack[stop].ValTyp == jsonElemKV {
				kv := stack[stop].Val.(*JsonKV)
				kv.Val = jv
				goto s_next
			} else if stack[stop].ValTyp == jsonValAry {
				ary := stack[stop].Val.(*JsonAry)
				*ary = append(*ary, jv)
				goto s_next
			}

		case jsonTkObjBegin:
		case jsonTkAryBegin:
		default:
			oerr = syntaxErr2(tk)
			goto s_return
		}
	}

s_next:
	switch stack[stop].ValTyp {
	case jsonValObj:
		goto s_obj
	case jsonValAry:
		goto s_ary
	case jsonElemVal:
		goto s_val
	case jsonElemDoc:
		goto s_doc
	}
	panic("s_next")

s_return:
	return
}
