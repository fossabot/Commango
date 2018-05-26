/*
* @Author: matt
* @Date:   2018-05-25 15:58:30
* @Last Modified by:   matt
* @Last Modified time: 2018-05-25 16:53:42
*/

package main

import(
    "fmt"
    "github.com/jacobsa/go-serial/serial"
    "io"
)

type(
    Comm_Interface Interface{
        Open_Comm() (err error)
        Close_Comm() (err error)
        Write_Comm(message string) (err error)
        Write_Comm(message []string) (err error)
        Read_Comm() (message string, err error)
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

func (comm *Comm) Init_Comm(port_path string, baud int, data_bits int, stop_bits int, minimum_read int) (err error){
    //TODO explore other serial.OpenOptions
    comm.options = serial.OpenOptions{
        PortName: port_path,
        BaudRate: baud,
        DataBits: data_bits,
        StopBits: stop_bits,
        MinimumReadSize: minimum_read
    }

    return 
}

func (comm *Comm) Open_Comm() (err error){

    return
}

func (comm *Comm) Close_Comm() (err error){
    return
}

func (comm *Comm) Write_Comm(message string) (err error){
    return
}

func (comm *Comm) Write_Comm(message []string) (err error){
    return
}

func (comm *Comm) Read_Comm() (message string, err error){
    return
}