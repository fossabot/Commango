/*
* @Author: Ximidar
* @Date:   2018-05-27 17:44:35
* @Last Modified by:   Ximidar
* @Last Modified time: 2018-06-02 23:28:26
 */

package main

import (
    "fmt"
    "github.com/Commango/src/comm"
    "io"
    "os"
    "time"
    "strings"
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
    defer comm.Close_Comm()


    go Read_Forever(comm)

    go Write_Forever(comm)

    fmt.Println("Sleeping")
    time.Sleep(2 * time.Second)
    COMM_OPEN = false
    
    fmt.Println("Finished")

}

func Read_Forever(comm *commango.Comm){

    for COMM_OPEN{
        out, err := comm.ReadLine()
        if err != nil{
            if err != io.EOF{
                fmt.Printf("%v", err)
            } 
        } else {
            string_out := string(out)
            string_out = strings.Replace(string_out, "\n", "", -1)

            
            if !check_blank(out){
                string_out = fmt.Sprintf("RECV: %v", string_out)
                fmt.Println(string_out)
            }
            
        }
    }
    fmt.Println("Stopping the reading")
}

func check_blank(byte_slice []byte) bool {
    if len(byte_slice) > 2{
        return false
    }

    return true
    
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
    fmt.Println("Stopping the writing")
}

