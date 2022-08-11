package GZip

/*
Code Ownership:
Millenia - Created all Functions
*/

//go-gzipb64 Github Repo

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func main() {

	//Input file or string to decode/encode here

	fmt.Print("Enter file name to encode/decode:")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if err != nil {
		fmt.Println("Cannot read the input:", err)
		return
	}

	var text []byte
	// fileContent, err := os.ReadFile(input)

	if err != nil {
		fmt.Println("\nTreating input as text")
		text = []byte(input)
	} else {
		fileContent := text
		fmt.Print("\n Treated the input as a file name and read the contents:\n\n", string(fileContent), "\n\n")
		text = fileContent
	}
	// Try to decode it
	op := "decode"
	result, err := Decode(text)
	if err != nil {
		// The only option left is to encode it!
		op = "encode"
		result, err = Encode(text)
		if err != nil {
			fmt.Println("Completely failed to do anything useful with that input")
			return
		}
	}

	fmt.Print("successful!\n\n")

	// output it
	os.Stdout.Write([]byte(result))
	fmt.Print("\n\n")

	// if len(fileContent) > 0 {
	fileOut := input + "." + op + "d"
	fmt.Print("Output written to " + fileOut + "\n\n")
	// os.WriteFile(fileOut, []byte(result), os.FileMode(os.O_RDWR|os.O_CREATE|os.O_TRUNC))
	// }
}

func Decode(text []byte) (result string, err error) {

	fmt.Print("Working to decode the content... ")

	// base64 decode it
	textDecoded := make([]byte, len(text))
	_, err = base64.RawStdEncoding.Decode(textDecoded, text)
	if err != nil {
		fmt.Print(err)
		return
	}

	// decompress it
	reader := bytes.NewReader(textDecoded)
	gzreader, err := gzip.NewReader(reader)
	gzreader.Multistream(false)
	if err != nil {
		fmt.Print("error at first stage: " + err.Error())
		return
	}

	resultBytes, err := ioutil.ReadAll(gzreader)
	if err != nil {
		fmt.Print("error at second stage: " + err.Error())
	}

	result = string(resultBytes)
	return
}

func Encode(text []byte) (result string, err error) {

	fmt.Print("\nWorking to encode the content... ")

	// Steps to compress it
	buf := new(bytes.Buffer)
	gz := gzip.NewWriter(buf)
	_, err = gz.Write(text)
	if err != nil {
		fmt.Print(err)
		return
	}
	gz.Close()

	// base64 to encode it
	result = base64.RawStdEncoding.EncodeToString(buf.Bytes())

	return
}

// Chilkat API
//Golang console utility
//GitHub, Inc.
