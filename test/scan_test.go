package test

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/shory152/jsonp"
)

//Version: 1.7.10
var jsonstr string = `

{
	"name":	"Jack (\"Bee\") Nimble",
	"format":	{
		"type":	"rect",
		"width":	1920,
		"height":	1080,
		"interlace":	false,
		"frame rate":	24
	}
}
["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"]
[[0, -1, 0], [1, 0, 0], [0, 0, 1]]
{
	"Image":	{
		"Width":	800,
		"Height":	600,
		"Title":	"View from 15th Floor",
		"Thumbnail":	{
			"Url":	"http:/*www.example.com/image/481989943",
			"Height":	125,
			"Width":	"100"
		},
		"IDs":	[116, 943, 234, 38793]
	}
}
[{
		"precision":	"zip",
		"Latitude":	37.7668,
		"Longitude":	-122.3959,
		"Address":	"",
		"City":	"SAN FRANCISCO",
		"State":	"CA",
		"Zip":	"94107",
		"Country":	"US"
	}, {
		"precision":	"zip",
		"Latitude":	37.371991,
		"Longitude":	-122.026,
		"Address":	"",
		"City":	"SUNNYVALE",
		"State":	"CA",
		"Zip":	"94085",
		"Country":	"US"
	}]
{
	"number":	null
}
[true, false, null, "ad''\\\t\n\rf\uabcda", -123.45e-3]
`

func TestScan(t *testing.T) {
	scanner := jsonp.ScanJson(jsonstr)
	var buf1 bytes.Buffer
	for {
		if tk, err := scanner(); err != nil {
			if err != io.EOF {
				t.Error(err)
			}
			break
		} else {
			buf1.WriteString(fmt.Sprint(tk))
		}
	}

	scanner = jsonp.ScanJson2(jsonstr)
	var buf2 bytes.Buffer
	for {
		if tk, err := scanner(); err != nil {
			if err != io.EOF {
				t.Error(err)
			}
			break
		} else {
			buf2.WriteString(fmt.Sprint(tk))
		}
	}

	if buf1.String() != buf2.String() {
		t.Error("not match")
	}
	t.Log(buf1.String())
	t.Log(buf2.String())

	scanner = jsonp.ScanJson2(jsonstr)
	var buf3 bytes.Buffer
	for {
		if tk, err := scanner.Scan(); err != nil {
			if err != io.EOF {
				t.Error(err)
			}
			break
		} else {
			buf3.WriteString(fmt.Sprint(tk))
		}
	}
	t.Log(buf3.String())
	if buf1.String() != buf3.String() {
		t.Error("not match")
	}

	s3 := jsonp.ScanJson3(jsonstr)
	buf3.Reset()
	for {
		if tk, err := s3.Scan(); err != nil {
			if err != io.EOF {
				t.Error(err)
			}
			break
		} else {
			buf3.WriteString(fmt.Sprint(tk))
		}
	}
	t.Log(buf3.String())
	if buf1.String() != buf3.String() {
		t.Error("not match")
	}
}

func TestParse(t *testing.T) {
	doc, err := jsonp.ParseJson(jsonstr)
	if err != nil {
		t.Error(err)
	} else {
		jsonp.ShowJson(doc, nil)
	}
}

func BenchmarkScan(b *testing.B) {
	for i := 0; i < b.N; i++ {
		scanner := jsonp.ScanJson(jsonstr)
		for {
			if tk, err := scanner(); err != nil {
				if err != io.EOF {
					b.Error(err)
				}
				break
			} else {
				_ = tk
			}
		}
	}
}
func BenchmarkScan2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		scanner := jsonp.ScanJson2(jsonstr)
		for {
			if tk, err := scanner(); err != nil {
				if err != io.EOF {
					b.Error(err)
				}
				break
			} else {
				_ = tk
			}
		}
	}
}
func BenchmarkScan22(b *testing.B) {
	for i := 0; i < b.N; i++ {
		scanner := jsonp.ScanJson2(jsonstr)
		for {
			if tk, err := scanner.Scan(); err != nil {
				if err != io.EOF {
					b.Error(err)
				}
				break
			} else {
				_ = tk
			}
		}
	}
}
func BenchmarkScan3(b *testing.B) {
	//b.N = 100000
	for i := 0; i < b.N; i++ {
		scanner := jsonp.ScanJson3(jsonstr)
		for {
			if tk, err := scanner.Scan(); err != nil {
				if err != io.EOF {
					b.Error(err)
				}
				break
			} else {
				_ = tk
			}
		}
	}
}

func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if jdoc, err := jsonp.ParseJson(jsonstr); err != nil {
			b.Error(err)
		} else {
			_ = jdoc
		}
	}
}
