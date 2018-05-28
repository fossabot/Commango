/*
* @Author: matt
* @Date:   2018-05-25 15:58:30
* @Last Modified by:   Ximidar
* @Last Modified time: 2018-05-27 22:50:19
*/

package commango

import(
    "fmt"
    "github.com/jacobsa/go-serial/serial"
    "io"
    "strings"
    "encoding/hex"
    "time"
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
        
        finished_reading bool
        Port io.ReadWriteCloser

    }
)

func New_Comm() (*Comm){
    comm := new(Comm)
    return comm
}

func (comm *Comm) Init_Comm(port_path string, baud uint, data_bits uint, stop_bits uint) (err error){
    //TODO explore other serial.OpenOptions
    comm.options = serial.OpenOptions{
        PortName: port_path,
        BaudRate: baud,
        DataBits: data_bits,
        StopBits: stop_bits,
        MinimumReadSize: 0,
        InterCharacterTimeout: 1000,
        ParityMode: serial.PARITY_NONE,

    }

    return 
}

func (comm *Comm) Open_Comm() (err error){
    fmt.Printf("Opening port with address %v\n", comm.options.PortName)
    comm.Port, err = serial.Open(comm.options)
    if err != nil {
      fmt.Println("Error Could not open port", err)
    }
    // Sleep to allow the port to start up
    time.Sleep(3 * time.Second)
    return
}

func (comm *Comm) Close_Comm() (err error){
    fmt.Printf("Closing port with address %s\n", comm.options.PortName)
    comm.Port.Close()
    return
}

func (comm *Comm) Write_Comm_String(message string) (len_written int, err error){
    fmt.Println("Writing to Comm Port", message)
    byte_message := []byte(message)
    expected_write := len(byte_message)
    len_written, err = comm.Port.Write(byte_message)
    if err == nil{
        comm.Last_Write = message
    }
    if len_written != expected_write{
        fmt.Println("Didn't write expected amount of bytes")
        fmt.Printf("Written: %v Expected: %v", len_written, expected_write)
    }
    return
}

func (comm *Comm) Write_Comm_Array(message []string) (len_written int, err error){
    joined_message := strings.Join(message, " ")
    len_written, err = comm.Write_Comm_String(joined_message)
    return
}

func (comm *Comm) Read_Line() (message string, err error){
    fmt.Println("Trying to read from Port ", comm.options.PortName)
    var read_bytes []byte
    n := -1
    
    for n != 0{
        buf := make([]byte, 32)
        n, err = comm.Port.Read(buf)
        if err != nil{
            if err == io.EOF{
                fmt.Println("EOF found")
                if n > 0{
                    read_bytes = append(read_bytes, buf[:n]...)
                    message = hex.EncodeToString(read_bytes)
                    return
                }
            }else{
                fmt.Println(err)
                return
            }
            
        }

        read_bytes = append(read_bytes, buf[:n]...)


    }
    
    message = hex.EncodeToString(read_bytes)
    return
}

