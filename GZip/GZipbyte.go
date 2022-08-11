package GZip

/*
Code Ownership:
Millenia - Created all Functions
*/

import (
	"bytes"
	"compress/gzip"
)

func Compress(input []byte) []byte {
	var buf bytes.Buffer
	compr := gzip.NewWriter(&buf)
	compr.Write(input)
	compr.Close()
	output := buf.Bytes()

	//print(output)
	return output

}
