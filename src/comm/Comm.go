/*
* @Author: matt
* @Date:   2018-05-25 15:58:30
* @Last Modified by:   Ximidar
* @Last Modified time: 2018-06-03 00:09:12
 */

package commango

import (
	_ "encoding/hex"
	"errors"
	"fmt"
	"github.com/jacobsa/go-serial/serial"
	"io"
	"strings"
	"time"
)

type (
	Comm struct {
		options    serial.OpenOptions
		Last_Read  string
		Last_Write string

		finished_reading bool
		Port             io.ReadWriteCloser
	}
)

func New_Comm() *Comm {
	comm := new(Comm)
	return comm
}

func (comm *Comm) Init_Comm(port_path string, baud uint, data_bits uint, stop_bits uint) (err error) {
	//TODO explore other serial.OpenOptions
	comm.options = serial.OpenOptions{
		PortName:              port_path,
		BaudRate:              baud,
		DataBits:              data_bits,
		StopBits:              stop_bits,
		MinimumReadSize:       0,
		InterCharacterTimeout: 1000,
		ParityMode:            serial.PARITY_NONE,
	}

	return
}

func (comm *Comm) Open_Comm() (err error) {
	fmt.Printf("Opening port with address %v\n", comm.options.PortName)
	comm.Port, err = serial.Open(comm.options)
	if err != nil {
		fmt.Println("Error Could not open port", err)
	}
	// Sleep to allow the port to start up
	time.Sleep(20 * time.Millisecond)
	return
}

func (comm *Comm) Close_Comm() (err error) {
	fmt.Printf("Closing port with address %s\n", comm.options.PortName)
	comm.Port.Close()
	return
}

func (comm *Comm) Write_Comm_String(message string) (len_written int, err error) {
	log_message := strings.Replace(message, "\n", "", -1)
	log_message = fmt.Sprintf("SENT: %v", log_message)
	fmt.Println(log_message)
	byte_message := []byte(message)
	expected_write := len(byte_message)
	len_written, err = comm.Port.Write(byte_message)
	if err == nil {
		comm.Last_Write = message
	}
	if len_written != expected_write {
		fmt.Println("Didn't write expected amount of bytes")
		fmt.Printf("Written: %v Expected: %v", len_written, expected_write)
	}
	return
}

func (comm *Comm) Write_Comm_Array(message []string) (len_written int, err error) {
	joined_message := strings.Join(message, " ")
	len_written, err = comm.Write_Comm_String(joined_message)
	return
}

func (comm *Comm) ReadLine() (out []byte, err error) {
	read_line := false
	for read_line == false {
		read, err := comm.ReadWithTimeout(1)

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

func (comm *Comm) ReadWithTimeout(n int) ([]byte, error) {
	buf := make([]byte, n)
	done := make(chan error)
	readAndCallBack := func() {
		_, err := io.ReadAtLeast(comm.Port, buf, n)
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
