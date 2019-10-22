package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/rs/zerolog/log"
)

const BUFFERSIZE = 1024

func main() {

	//connect to server socket
	connection, err := net.Dial("tcp", "192.168.88.10:8081")
	if err != nil {
		log.Error().Err(err).Msg("")
	}
	defer connection.Close()

	fmt.Println("Connected to server, start receiving the file name and file size")

	bufferFileName := make([]byte, 64)
	bufferFileSize := make([]byte, 10)

	connection.Read(bufferFileSize)
	fileSize, err := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)
	if err != nil {
		log.Error().Err(err).Msg("")
	}

	connection.Read(bufferFileName)
	fileName := strings.Trim(string(bufferFileName), ":")

	fmt.Println("Firmware:", fileName, ",", "size:", fileSize/1024, "Kb")

	newFile, err := os.Create(fileName)
	if err != nil {
		log.Error().Err(err).Msg("")
	}
	defer newFile.Close()

	var receivedBytes int64

	//progress bar
	var countSize int = int(fileSize / 1024)
	count := countSize
	// create and start new bar
	bar := pb.StartNew(count)

	for {
		if (fileSize - receivedBytes) < BUFFERSIZE {
			io.CopyN(newFile, connection, (fileSize - receivedBytes))
			connection.Read(make([]byte, (receivedBytes+BUFFERSIZE)-fileSize))
			bar.Finish()
			break
		}
		bar.Increment()
		time.Sleep(time.Millisecond)

		io.CopyN(newFile, connection, BUFFERSIZE)
		receivedBytes += BUFFERSIZE

	}

	fmt.Println("Received file completely!")
	fmt.Printf("%s with %v bytes downloaded\n\r", fileName, count)
}
