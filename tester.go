package tester

import (
	"encoding/json"
	"fmt"
	hpack "github.com/ami-GS/GoHPACK"
	"html/template"
	"net/http"
	"strings"
	//"time"

	"appengine"
)

type jsonobject struct {
	Cases       []Case
	Draft       int
	Description string
}

type Case struct {
	Seqno             int
	Header_table_size uint32
	Wire              string
	Headers           []map[string]string
}

func init() {
	http.HandleFunc("/", root)
	http.HandleFunc("/verify", verify)
}

func ConvertHeader(headers []map[string]string) (dist []hpack.Header) {
	for _, dict := range headers {
		for k, v := range dict {
			dist = append(dist, hpack.Header{k, v})
		}
	}
	return
}

func root(w http.ResponseWriter, r *http.Request) {
	if err := inputHPACK.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var inputHPACK = template.Must(template.New("hoge").Parse(`
<html>
  <head>
    <title>HPACK tester</title>
  </head>
  <body>
    <p>input header json or wire</p>
    <form action="/verify" method="post">
      <div><textarea raws="10" cols="60" name="content"></textarea></div>
      <div><input type="checkbox" value="static" name="content">
      <input type="checkbox" value="dynamic" name="content">
      <input type="checkbox" value="huffman" name="content"></div>
      <div><input type="submit" value="verify"></div>
    </form>
  </body>
</html>
`))

var table = hpack.InitTable()

func verify(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		panic(err)
		return
	}
	isStatic := false
	isDynamic := false
	isHuffman := false
	var rawData string
	c := appengine.NewContext(r)
	c.Debugf("r.Form %s", r.Form["content"])
	for _, content := range r.Form["content"] {
		if string(content) == "static" {
			isStatic = true
		} else if string(content) == "dynamic" {
			isDynamic = true
		} else if string(content) == "huffman" {
			isHuffman = true
		} else {
			rawData = string(content)
		}
	}
	if strings.Contains(rawData, "[") {
		var jsontype jsonobject
		data := []byte(rawData)
		json.Unmarshal(data, &jsontype)
		for i, seq := range jsontype.Cases {
			headers := ConvertHeader(seq.Headers)
			wire := hpack.Encode(headers, isStatic, isDynamic, isHuffman, &table, -1)
			fmt.Fprintf(w, "seqno:%v %s\n\n", i, wire)
		}
	} else {
		headers := hpack.Decode(rawData, &table)
		fmt.Fprintf(w, "%v\n", headers)
	}
}
