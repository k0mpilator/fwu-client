package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/rs/zerolog/log"
)

const BUFFERSIZE = 1024

//connect to server
func connSrv() (conn net.Conn, err error) {
	//connect to server socket
	conn, err = net.Dial("tcp", "192.168.88.11:49999")
	if err != nil {
		fmt.Printf("[  ERR ] Server not found\r\n")
	}
	return conn, err
}

//read firmware version /etc/version
func compareFwVer() ([]byte, error) {
	valfw, err := ioutil.ReadFile("/etc/version")
	if err != nil {
		return nil, err
	}
	v := string(valfw)
	fmt.Printf("[CLIENT] Board fw: %s\r", v)
	return valfw, nil
}

//progress bar
func progressBar(fileSize int64, connection net.Conn, newFile *os.File, fileName string) {

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

	fmt.Printf("[  OK  ] Received file completely!\r\n")
	fmt.Printf("[  OK  ] %s with %v bytes downloaded\r\n", fileName, count)
}

//execute and run bash script
func execBash() {
	output := exec.Command("./fwu.sh")
	_, err := output.Output()
	if err != nil {
		log.Error().Err(err).Msg("")
	}
}

func main() {

	var a, b []byte

	connection, err := connSrv()
	if err != nil {
		os.Exit(3)
	}
	fmt.Printf("[  OK  ] Connected to server...\r\n")
	//defer connection.Close()

	bufferFileName := make([]byte, 64)
	bufferFileSize := make([]byte, 10)

	connection.Read(bufferFileSize)
	fileSize, err := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)
	if err != nil {
		log.Error().Err(err).Msg("")
	}

	connection.Read(bufferFileName)
	fileName := strings.Trim(string(bufferFileName), ":")

	var re = regexp.MustCompile(`(?m)\d{14}`)
	var str = string(fileName)

	strarr := re.FindAllString(str, -1)
	singstr := strings.Join(strarr, " ")
	bfwNum := []byte(singstr)
	//compare FW version
	a = bfwNum
	s := string(a)
	fmt.Printf("[  SRV ] Current firmware: %s\r\n", s)
	b, _ = compareFwVer()
	if bytes.Compare(a, b) < 0 {
		fmt.Printf("[CLIENT] Update not required\r\n")
	} else {
		fmt.Printf("[CLIENT] Update required\r\n")
		fmt.Printf("[  SRV ] Firmware: %s, size: %d Kb\r\n", fileName, fileSize/1024)
		newFile, err := os.Create(fileName)
		if err != nil {
			log.Error().Err(err).Msg("")
		}
		progressBar(fileSize, connection, newFile, fileName)
		defer newFile.Close()
		fmt.Println("Don`t power off board...")
		fmt.Println("System will reboot after few second...")
		execBash()
	}
}
