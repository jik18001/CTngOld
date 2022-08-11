package GZip

/*
Code Ownership:
Millenia - Created all Functions
*/

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
)

//Functions for Gzip Decompression

func _() {
	original := "os.File{}" //String placeholder for file holding CRV

	reader := bytes.NewReader([]byte(original))

	gzipreader, e1 := gzip.NewReader(reader) //new gzip reader

	if e1 != nil {
		fmt.Println(e1)
	}

	output, e2 := ioutil.ReadAll(gzipreader)
	if e2 != nil {
		fmt.Println(e2)
	}

	result := string(output)

	println(result) //Print Decompressed Data

}
