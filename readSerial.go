package main

import (
	"github.com/tarm/serial"
	"io/ioutil"
	"fmt"
	"log"
	"net/http"
	"strings"
)

//scans for connected serial devices
//and returns an array with their names
func ScanPorts() ([]string, error) {
	names := make([]string, 0)

	//read folders in directory
	files, err := ioutil.ReadDir("/dev/serial/by-id/")

	//check no errors, ie no serial devices
	if err != nil {
		return names, err
	}

	//add each folder name to names array
	for _, f := range files {
		names = append(names, f.Name())
		fmt.Println(f.Name())
	}
	//returns the names
	return names, err

}

//TODO: change to variable baud Rate
//takes a name and opens serial connection
func OpenSerial(fileName string) (*serial.Port, error) {

	c := &serial.Config{Name: "/dev/serial/by-id/" + fileName, Baud: 115200}
	s, err := serial.OpenPort(c)
	if err != nil {
		return s, err
	}

	//test byte write, to make sure connection working
	_, err = s.Write([]byte("test"))
	if err != nil {
		return s, err
	}

	return s, err

}

func StreamSerialWeb(s *serial.Port, w http.ResponseWriter) () {

	var buffer []byte
	//2kbyte buffer
	buffer = make([]byte, 2056)

	//read serial port into buffer
	len, err := s.Read(buffer)

	//fmt.Println(len)
	//make sure no errors or log fatal
	if err != nil {
		log.Fatal(err)
	}

	//trim buffer used to remove all
	//the unused bytes
	trimBuffer := make([]byte, len)
	trimBuffer = buffer[:len]

	//convert byte to string
	stin := string(trimBuffer)
	//token
	newToken := string(0x00A) + string(0x00A) + "data:"
	//create a new string with all the "\n" replaced by "\n\ndata:"
	//that way if a new line character is sent and the event stream is done
	//the data after is not lost but rather still sent in a new event
	newStin := strings.Replace(stin, string(0x00A), newToken, 100)
	//convert string to bytes
	pushBuffer := []byte(newStin)
	//write bytes to web writes
	w.Write(pushBuffer)
}

func StreamSerial(s *serial.Port) () {
	var buffer []byte
	buffer = make([]byte, 1024)
	for ; ; {
		_, err := s.Read(buffer)
		if err != nil {
			log.Fatal(err)
		}

		for i, _ := range (buffer) {

			fmt.Printf("%c", buffer[i])
			buffer[i] = 0x0
		}

	}
}

func StreamSerialTo(s *serial.Port) ([]byte) {
	var buffer []byte
	var bufferOut []byte
	var dataSize = 38

	buffer = make([]byte, dataSize)
	bufferOut = make([]byte, dataSize)
	_, err := s.Read(buffer)
	if err != nil {
		log.Fatal(err)
	}
	bufferOut = buffer

	return bufferOut

}
