package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

const DataBufferSize = 4096

var wav = [4]string{"52", "49", "46", "46"}
var wavDataChunk = [4]string{"64", "61", "74", "61"}
var filesCreated = 0

var indexedMatchers = make(map[string]int)

func main() {
	initMatchers()

	inputFile, err := os.Open("sample.pak")
	checkErr(err)
	defer func() {
		if err := inputFile.Close(); err != nil {
			panic(err)
		}
	}()

	reader := bufio.NewReader(inputFile)

	pos := 0
	for {
		b, eof := reader.ReadByte()
		sByte := fmt.Sprintf("%x", b)
		if eof != nil {
			break
		}

		isWav := isWavFound(sByte)
		if isWav {
			fmt.Printf("Found wav at position %d\n", pos)
			header, dataLength, err := extractWavHeader(reader, &pos)
			checkErr(err)
			extractWav(reader, header, dataLength, &pos)
		}
		pos++
	}
}

func initMatchers() {
	indexedMatchers["wav"] = 0
}

func isWavFound(b string) bool {
	currentMatchedPos := indexedMatchers["wav"]

	if wav[currentMatchedPos] == b {
		indexedMatchers["wav"] += 1
	} else {
		indexedMatchers["wav"] = 0
	}

	if indexedMatchers["wav"] == len(wav) {
		indexedMatchers["wav"] = 0
		return true
	}
	return false
}

func extractWavHeader(reader *bufio.Reader, pos *int) ([]byte, uint, error) {
	currentMatchedPos := 0
	header := []byte{0x52, 0x49, 0x46, 0x46}
	for {
		b, eof := reader.ReadByte()
		sByte := fmt.Sprintf("%x", b)
		header = append(header, b)
		if wavDataChunk[currentMatchedPos] == sByte {
			currentMatchedPos++
		} else {
			currentMatchedPos = 0
		}

		if currentMatchedPos == len(wav) {
			dataSizeBytes := make([]byte, 4)
			_, _ = reader.Read(dataSizeBytes)
			dataLength := uint(dataSizeBytes[0]) |
				uint(dataSizeBytes[1])<<8 |
				uint(dataSizeBytes[2])<<16 |
				uint(dataSizeBytes[3])<<32

			return append(header, dataSizeBytes...), dataLength, nil
		}

		if eof != nil {
			return header, 0, eof
		}
		*pos++
	}
}

func extractWav(reader *bufio.Reader, header []byte, dataLength uint, pos *int) {
	fmt.Printf("Reading wav data chunk of length %d\n", dataLength)

	filename := generateFilename()
	writeToFile(header, filename)
	for i := int(dataLength); i > 0; i -= DataBufferSize {
		var data []byte
		if i < DataBufferSize {
			data = make([]byte, i)
		} else {
			data = make([]byte, DataBufferSize)
		}

		bytesRead, err := io.ReadFull(reader, data)
		writeToFile(data, filename)

		*pos += bytesRead
		if err != nil {
			break
		}
	}
}

func writeToFile(fileBytes []byte, dest string) {
	fmt.Printf("writing to file %s\n", dest)

	outputFile, err := os.OpenFile(dest, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	checkErr(err)

	defer func() {
		if err := outputFile.Close(); err != nil {
			panic(err)
		}
	}()

	outputByteLen, err := outputFile.Write(fileBytes)
	fmt.Printf("finished writing %d bytes to %s\n", outputByteLen, dest)
}

func generateFilename() string {
	filename := fmt.Sprintf("output%d.wav", filesCreated)
	filesCreated++
	return filename
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
