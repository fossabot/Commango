/*
* @Author: Ximidar
* @Date:   2018-05-27 17:44:35
* @Last Modified by:   Ximidar
* @Last Modified time: 2018-06-02 14:51:12
 */

package main

import (
	"errors"
	"fmt"
	"github.com/Commango/src/comm"
	"io"
	"os"
	"time"
)

var COMM_OPEN bool

func main() {

	comm := commango.New_Comm()

	err := comm.Init_Comm("/dev/ttyACM0", 115200, 8, 1)

	if err != nil {
		fmt.Println(err)
		fmt.Println("Cannot assign options")
		os.Exit(2)
	}

	err = comm.Open_Comm()
    COMM_OPEN = true

	if err != nil {
		fmt.Println("Cannot open port")
	}


    go Read_Forever(comm.Port)

	go Write_Forever(comm)

    fmt.Println("Sleeping")
    time.Sleep(5 * time.Second)
    COMM_OPEN = false
    comm.Close_Comm()
    fmt.Println("Finished")

}

func Read_Forever(r io.Reader){

    for COMM_OPEN{
        out, err := ReadLine(r)
        if err != nil{
            if err != io.EOF{
                fmt.Println(err)
            } 
        } else {
            fmt.Println(string(out))
        }
    }
    
}

func Write_Forever(comm *commango.Comm){
    count := 0
    for COMM_OPEN{
        message := fmt.Sprintf("Hello at count: %v\n", count)
        _, err := comm.Write_Comm_String(message)
        if err != nil {
            fmt.Println(err)
        } 
        count += 1
        time.Sleep(25 * time.Millisecond)
    }
}

func ReadLine(r io.Reader) (out []byte, err error) {
	read_line := false
	for read_line == false {
		read, err := ReadWithTimeout(r, 1)

		if err != nil {
			//fmt.Println("Readline errored out", err)
			return nil, err
		}

		if read[0] == 10 {
			out = append(out, read[0])
			read_line = true
			return out, nil
		}

		out = append(out, read[0])

	}

	return

}

func ReadWithTimeout(r io.Reader, n int) ([]byte, error) {
	buf := make([]byte, n)
	done := make(chan error)
	readAndCallBack := func() {
		_, err := io.ReadAtLeast(r, buf, n)
		done <- err
	}

	go readAndCallBack()

	timeout := make(chan bool)
	sleepAndCallBack := func() { time.Sleep(2e9); timeout <- true }
	go sleepAndCallBack()

	select {
	case err := <-done:
		return buf, err
	case <-timeout:
		return nil, errors.New("Timed out.")
	}

	return nil, errors.New("Can't get here.")
}
