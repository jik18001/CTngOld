# GZIP Implementation

## Contents: 
- `Decode.go`: defines the decode and encode functions 
- `GZipbyte.go`: contains the compress function 
- `GZipDecomp.go`: contains the decompression function

## Decode.go:
- Decode takes in bytes, decompresses them, then returns the decoded string if successful, error if not
- Encode takes in bytes, compresses them, then returns an encoded string if successful, error if not 

## GZipbyte.go: 
- When given a bytes it will compress them and return compressed bytes

## GZipDecomp.go:
- Decompresses the the compressed bytes to readable data

# Code Ownership
- Millenia - All functions in this package

##### Written By Isaac