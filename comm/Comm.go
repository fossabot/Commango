/*
* @Author: matt
* @Date:   2018-05-25 15:58:30
* @Last Modified by:   Ximidar
* @Last Modified time: 2018-10-01 01:08:50
 */

package commango

import (
	_ "encoding/hex"
	"errors"
	"fmt"
	"go.bug.st/serial.v1" //https://godoc.org/go.bug.st/serial.v1
	"go.bug.st/serial.v1/enumerator"
	"io"
	"log"
	"strings"
	"strconv"
	"time"
	ms "github.com/ximidar/Flotilla/data_structures"
)

type Read_Line_Callback func(string)
type Emit_Write_Callback func(string)
type Emit_Status_Callback func(*ms.Comm_Status)

type Comm struct {
	options         *serial.Mode
	Available_Ports []string
	Detailed_Ports  []*enumerator.PortDetails
	Port_Path       string
	Port            serial.Port
	connected       bool
	Read_Stream     chan string
	Byte_Stream		chan byte
	Error_Stream    chan error

	Emit_Read Read_Line_Callback
	Emit_Write Emit_Write_Callback
	Emit_Status Emit_Status_Callback

	finished_reading bool
}

func New_Comm(read_line_callback Read_Line_Callback, 
			  write_line_callback Emit_Write_Callback, 
			  emit_status_callback Emit_Status_Callback) *Comm {
	comm := new(Comm)
	comm.Port_Path = "None"
	comm.connected = false
	comm.Emit_Read = read_line_callback
	comm.Emit_Write = write_line_callback
	comm.Emit_Status = emit_status_callback
	comm.options = &serial.Mode{
		BaudRate: 115200,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}
	comm.Read_Stream = make(chan string, 10)
	comm.Byte_Stream = make(chan byte, 10)
	comm.Error_Stream = make(chan error, 10)
	return comm
}

func (comm *Comm) Connected() (bool){
	return comm.connected
}

func (comm *Comm) Set_Connected(value bool) {
	comm.connected = value
	status := comm.Get_Comm_Status()
	comm.Emit_Status(status)
}

func (comm *Comm) Init_Comm(port_path string, baud int) error {

	comm.Port_Path = port_path
	comm.options = &serial.Mode{
		BaudRate: baud,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}
	comm.Print_Options()
	return nil
}

func (comm Comm) Print_Options() {
	fmt.Println("Comm Options:")
	fmt.Println("|  Port Path:", comm.Port_Path)
	fmt.Println("|  Serial Options:")
	fmt.Println("|  |  Baud Rate:", comm.options.BaudRate)
	fmt.Println("|  |  Parity:", comm.options.Parity)
	fmt.Println("|  |  Data Bits:", comm.options.DataBits)
	fmt.Println("|  |  Stop Bits:", comm.options.StopBits)
}

func (comm Comm) Get_Comm_Status() (*ms.Comm_Status){
	cs := new(ms.Comm_Status)

	cs.Port = comm.Port_Path
	cs.Baud = strconv.Itoa(comm.options.BaudRate)
	cs.Connected = comm.Connected()

	return cs
}

func (comm *Comm) Get_Available_Ports() ([]string, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, err
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
func (comm Comm) PreCheck() (ready bool) {
	ready = true
	if comm.Port_Path == "" {
		ready = false
	}
	return
}

// This initiates a reset for most microcontrollers
func (comm *Comm) cycle_dtr(){
	comm.Port.SetDTR(false)
	time.Sleep(100 * time.Millisecond)
	comm.Port.SetDTR(true)
	time.Sleep(100 * time.Millisecond)
	comm.Port.SetDTR(false)
}

func (comm *Comm) Open_Comm() error {

	// recover in case this all goes tits up
	defer func(){
		if recovery := recover(); recovery != nil{
			fmt.Println("Could not Open Port.", recovery)
			comm.Set_Connected(false)
		}
	}()

	// Do a Precheck before starting
	if !comm.PreCheck() {
		fmt.Println("Precheck Failed!")
		return errors.New("Precheck Failed, Please initialize the comm before trying to open it.")
	}

	if comm.Connected(){
		fmt.Println("Comm is already connected")
		return errors.New("Comm is already Connected")
	}

	var err error
	fmt.Printf("Attempting to open port with address %v\n", comm.Port_Path)
	comm.Port, err = serial.Open(comm.Port_Path, comm.options)
	if err != nil {
		fmt.Println("Error Could not open port", err)
		return err
	}
	comm.cycle_dtr()
	// Sleep to allow the port to start up
	time.Sleep(20 * time.Millisecond)
	comm.Set_Connected(true)
	fmt.Println("Port Opened")

	// Start up a reader
	//go comm.Read_Forever()
	go comm.Read_OK_Forever()
	return nil
}

func (comm *Comm) Close_Comm() error {
	// recover in case this all goes tits up
	defer func(){
		if recovery := recover(); recovery != nil{
			fmt.Println("Could not close port.", recovery)
		}
	}()

	if comm.Connected(){
		fmt.Printf("Attempting to close port with address %s\n", comm.Port_Path)
		err := comm.Port.Close()
		if err != nil {
			fmt.Println("Could not close port")
			return err
		}
		fmt.Println("Port Closed")
	}

	comm.Set_Connected(false)
	return nil
}

func (comm *Comm) Write_Comm(message string) (len_written int, err error) {
	len_written = -1
	//prepare string for writing
	// Check that message has an \n after it
	if !strings.HasSuffix(message, "\n") {
		message += "\n"
	}
	// Turn message into a bytestring
	byte_message := []byte(message)
	expected_write := len(byte_message)

	// Write the comm if we are connected
	if comm.Connected() {
		len_written, err = comm.Port.Write(byte_message)
		if err != nil {
			return
		}
		if len_written != expected_write {
			fmt.Println("Didn't write expected amount of bytes")
			fmt.Printf("Written: %v Expected: %v", len_written, expected_write)
			comm.Emit_Write(message + " Error on this line")
			return len_written, errors.New("Expected Bytes did not match written")
		}
	} else {
		err = errors.New("Cannot Write Comm")
		return
	}

	// Setup log message only if we wrote successfuly
	log_message := strings.Replace(message, "\n", "", -1)
	log_message = fmt.Sprintf("SENT: %v", log_message)
	fmt.Println(log_message)
	comm.Emit_Write(message)
	return
}


// This function will read the serial port until it receives ok
func (comm *Comm) Read_OK(){
	var buf []byte
	for comm.connected {
		select{
		case read := <- comm.Byte_Stream :
			//add the bytes to the buffer
			buf = append(buf, read)
		case <- time.After(10 * time.Millisecond):
			if len(buf) == 0 {continue}

			if comm.Check_for_OK(buf) {
				comm.Read_Stream <- string(buf)
				buf = []byte{}
			}
		case <- comm.Error_Stream:
			return //If we are erroring out then we can't read anything
		}
	}
}

// Check for ok or start
func (comm *Comm) Check_for_OK(buf []byte) (bool){
	buf_string := string(buf)

	if strings.Contains(buf_string, "ok"){
		fmt.Println("Got OK!")
		return true
	} else if strings.Contains(buf_string, "start"){
		fmt.Println("Got Start!")
		return true
	}

	fmt.Println("Got Nothing!")
	return false
}

// This function will continuously read the output from the Serial line
// It will send these bytes through a channel. 
func (comm *Comm) Stream_Bytes(){
	for comm.connected{
		read, err := comm.ReadBytes(1)
		if err != nil{
			comm.Error_Stream <- err
			fmt.Println("Stream Bytes Erroring out")
			return
		}
		comm.Byte_Stream <- read[0]
	}
}

func (comm *Comm) ReadLine() (out []byte, err error) {
	read_line := false
	for read_line == false {
		read, err := comm.ReadBytes(1)

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

func (comm *Comm) ReadBytes(n int) ([]byte, error) {
	buf := make([]byte, n)
	bytes_read, err := comm.Port.Read(buf)
	if bytes_read != n {
		log.Println(fmt.Sprintf("Read: %v Expexted: %v", bytes_read, n))
		err = errors.New(fmt.Sprintf("Read: %v Expexted: %v", bytes_read, n))
	}
	return buf, err
}

func (comm *Comm) Read_Forever() {

	for comm.connected {
		out, err := comm.ReadLine()
		if err != nil {
			if err != io.EOF {
				fmt.Printf("%v", err)
			}
		} else {
			string_out := string(out)
			comm.Emit_Read(string_out)
			string_out = fmt.Sprintf("RECV: %v", string_out)
			fmt.Print(string_out)
		}
	}
	fmt.Println("Stopping the reading")
}

func (comm *Comm) Read_OK_Forever() (err error){
	go comm.Stream_Bytes()
	go comm.Read_OK()

	for comm.connected {
		select{
		case full_string := <- comm.Read_Stream:
			comm.Emit_Read(full_string)
			fmt.Printf("RECV: %v\n", full_string)
		case err = <- comm.Error_Stream:
			fmt.Println("Read Ok Erroring out:", err.Error())
			return err
		}
	}

	fmt.Println("Stopped Reading")
	return nil


}
