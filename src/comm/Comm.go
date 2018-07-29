/*
* @Author: matt
* @Date:   2018-05-25 15:58:30
* @Last Modified by:   Ximidar
* @Last Modified time: 2018-07-28 22:50:21
 */

package commango

import (
	_ "encoding/hex"
	"errors"
	"fmt"
	"log"
	"go.bug.st/serial.v1" //https://godoc.org/go.bug.st/serial.v1
	"go.bug.st/serial.v1/enumerator"
	"github.com/godbus/dbus"
	_"io"
	"strings"
	"time"
)

type Comm struct {
	options    *serial.Mode
	Available_Ports []string
	Detailed_Ports []*enumerator.PortDetails
	Port_Path string
	Port             serial.Port


	Last_Read  string
	Last_Write string

	finished_reading bool
	
}

func New_Comm() *Comm {
	comm := new(Comm)
	return comm
}

func (comm *Comm) Init_Comm(port_path string, baud int) (*dbus.Error) {

	comm.Port_Path = port_path
	comm.options = &serial.Mode{
		BaudRate: baud,
		Parity: serial.EvenParity,
		DataBits: 7,
		StopBits: serial.OneStopBit, 
	}
	comm.Print_Options()
	return nil
}

func (comm Comm) Print_Options(){
	fmt.Println("Comm Options:")
	fmt.Println("|  Port Path:", comm.Port_Path)
	fmt.Println("|  Serial Options:")
	fmt.Println("|  |  Baud Rate:", comm.options.BaudRate)
	fmt.Println("|  |  Parity:", comm.options.Parity)
	fmt.Println("|  |  Data Bits:", comm.options.DataBits)
	fmt.Println("|  |  Stop Bits:", comm.options.StopBits)
}

func (comm *Comm) Get_Available_Ports() ([]string, *dbus.Error){
	ports, err := serial.GetPortsList()
	if err != nil {
		log.Fatal(err)
	}
	if len(ports) == 0 {
		ports = []string{string("none")}
	}
	comm.Available_Ports = ports
	return ports, nil
}

func (comm *Comm) Get_Detailed_Ports() {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		log.Fatal(err)
	}
	if len(ports) == 0 {
		fmt.Println("No serial ports found!")
		return
	}
	for _, port := range ports {
		fmt.Printf("Found port: %s\n", port.Name)
		if port.IsUSB {
			fmt.Printf("   USB ID     %s:%s\n", port.VID, port.PID)
			fmt.Printf("   USB serial %s\n", port.SerialNumber)
		}
	}

	comm.Detailed_Ports = ports
}

// Do a little error checking for good start
func (comm Comm) PreCheck() (ready bool){
	ready = true
	if comm.Port_Path == ""{
		ready = false
	}
	return
}

func (comm *Comm) Open_Comm() (*dbus.Error) {

	// Do a Precheck before starting
	if !comm.PreCheck(){
		fmt.Println("Precheck Failed!")
		return dbus.MakeFailedError(errors.New("Precheck Failed, Please initialize the comm before trying to open it."))
	}

	var err error
	fmt.Printf("Opening port with address %v\n", comm.Port_Path)
	comm.Port, err = serial.Open(comm.Port_Path, comm.options)
	if err != nil {
		fmt.Println("Error Could not open port", err)
		return dbus.MakeFailedError(err)
	}
	// Sleep to allow the port to start up
	time.Sleep(20 * time.Millisecond)
	return nil
}

func (comm *Comm) Close_Comm() (*dbus.Error) {
	fmt.Printf("Closing port with address %s\n", comm.Port_Path)
	comm.Port.Close()
	return nil
}

func (comm *Comm) Write_Comm_String(message string) (int, *dbus.Error) {
	log_message := strings.Replace(message, "\n", "", -1)
	log_message = fmt.Sprintf("SENT: %v", log_message)
	fmt.Println(log_message)
	byte_message := []byte(message)
	expected_write := len(byte_message)
	len_written, err := comm.Port.Write(byte_message)
	if err == nil {
		comm.Last_Write = message
	} else {
		return len_written, dbus.MakeFailedError(err)
	}
	if len_written != expected_write {
		fmt.Println("Didn't write expected amount of bytes")
		fmt.Printf("Written: %v Expected: %v", len_written, expected_write)
		return len_written, dbus.MakeFailedError(errors.New("Expected Bytes did not match written"))
	}
	return len_written, dbus.MakeFailedError(err)
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
		bytes_read, err := comm.Port.Read(buf)
		if bytes_read != n{
			log.Println(fmt.Sprintf("Read: %v Expexted: %v", bytes_read, n))
		}
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
