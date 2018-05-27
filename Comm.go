/*
* @Author: matt
* @Date:   2018-05-25 15:58:30
* @Last Modified by:   matt
* @Last Modified time: 2018-05-25 16:53:42
*/

package comm

import(
    "fmt"
    "github.com/jacobsa/go-serial/serial"
    "io"
    "strings"
    "encoding/hex"
)

type(
    Comm_Interface interface{
        Open_Comm() (err error)
        Close_Comm() (err error)
        Write_Comm_String(message string) (err error)
        Write_Comm_Array(message []string) (err error)
        //Read_Comm() (message string, err error)
        Read_Line() (message string, err error)
    }
    Comm struct{
        options serial.OpenOptions
        Last_Read string
        Last_Write string
        
        port io.ReadWriteCloser
    }
)

func New_Comm() (*Comm){
    comm := new(Comm)
    return comm
}

func (comm *Comm) Init_Comm(port_path string, baud uint, data_bits uint, stop_bits uint, minimum_read uint) (err error){
    //TODO explore other serial.OpenOptions
    comm.options = serial.OpenOptions{
        PortName: port_path,
        BaudRate: baud,
        DataBits: data_bits,
        StopBits: stop_bits,
        MinimumReadSize: minimum_read,
    }

    return 
}

func (comm *Comm) Open_Comm() (err error){
    fmt.Printf("Opening port with address %v\n", comm.options.PortName)
    comm.port, err = serial.Open(comm.options)
    if err != nil {
      fmt.Println("Error Could not open port", err)
    }
    return
}

func (comm *Comm) Close_Comm() (err error){
    fmt.Printf("Closing port with address %s\n", comm.options.PortName)
    comm.port.Close()
    return
}

func (comm *Comm) Write_Comm_String(message string) (err error){
    byte_message := []byte(message)
    _, err = comm.port.Write(byte_message)
    if err == nil{
        comm.Last_Write = message
    }

    return
}

func (comm *Comm) Write_Comm_Array(message []string) (err error){
    joined_message := strings.Join(message, " ")
    err = comm.Write_Comm_String(joined_message)
    return
}

func (comm *Comm) Read_Line() (message string, err error){
    var read_bytes []byte
    current_char := make([]byte, 1)
    _, err = comm.port.Read(current_char)
    if err != nil{
        fmt.Printf("Could not read from %s", comm.options.PortName)
        return
    }
    read_bytes = append(read_bytes, current_char[0])
    for hex.EncodeToString(current_char) != "\n"{
        current_char := make([]byte, 1)
        _, err = comm.port.Read(current_char)
        if err != nil{
            if err == io.EOF{
                fmt.Println("End of file reached")
                break
            }
        }
        read_bytes = append(read_bytes, current_char[0])
    }
    message = hex.EncodeToString(read_bytes)
    return
}