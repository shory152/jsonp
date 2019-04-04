package jsonp

import (
	"fmt"
	"io"
	"os"
)

func ShowJson(jdoc JsonDoc, w io.Writer) {
	if w == nil {
		w = os.Stderr
	}
	for _, val := range jdoc {
		fprintJval(w, val, 0)
	}
}

func fprintJval(w io.Writer, val JsonVal, lvl int) {
	switch val.ValTyp {
	case jsonValObj:
		jobj := val.Val.(JsonObj)
		fmt.Fprintf(w, "{\n")
		for i, kv := range jobj {
			fmt.Fprintf(w, "\"%s\": ", kv.Key)
			fprintJval(w, kv.Val, lvl+1)
			if i+1 < len(jobj) {
				fmt.Fprintf(w, ",\n")
			}
		}
		fmt.Fprintf(w, "}\n")
	case jsonValAry:
		jary := val.Val.(JsonAry)
		fmt.Fprintf(w, "[")
		for i, v := range jary {
			fprintJval(w, v, lvl+1)
			if i+1 < len(jary) {
				fmt.Fprintf(w, ", ")
			}
		}
		fmt.Fprintf(w, "]\n")
	case jsonValStr:
		fmt.Fprintf(w, "\"%s\"", val.Val)
	default:
		fmt.Fprint(w, val.Val)
	}
}
