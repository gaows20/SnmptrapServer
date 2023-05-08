package string_express

import (
	"bytes"
	"io/ioutil"
	"golang.org/x/text/encoding/simplifiedchinese"
	iconv "github.com/djimenez/iconv-go"
	"golang.org/x/text/transform"
)

func UTF82GB2312(s []byte)([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.HZGB2312.NewEncoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}


func GB23122ToUTF(s []byte)([]byte, error) {
	out := make([]byte, len(s))
	out = out[:]
	iconv.Convert(s, out, "gb2312", "utf-8")
	return out, nil
}
