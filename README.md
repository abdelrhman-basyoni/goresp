# go-resp

A Go package for reading, writing, and serializing RESP (REDis Serialization Protocol) data.
it's about a go project that i made for a go package that I used in two separate projects  the go-redis-cli and godis
both are side projects where i rewrite Redis using Golang
## Key Features:

 - RESP Reader: Reads RESP data from an io.Reader and converts it into goresp.Value objects.

- RESP Writer: Converts goresp.Value objects to RESP format and writes them to an io.Writer.

- RESP Serializer:
SerializeCommand: Converts a string command into RESP-formatted bytes.

- SerializeReaderCommand: Converts a goresp.Value representing a command into RESP-formatted bytes.

- RESP Marshaler: Converts a goresp.Value object into a RESP-formatted byte representation.
  
# Installation

Bash
```
go get github.com/abdelrhman-basyoni/goresp
```
# Test
```
go test
```
Use code with caution.

