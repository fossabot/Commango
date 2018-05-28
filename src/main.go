/*
* @Author: Ximidar
* @Date:   2018-05-27 17:44:35
* @Last Modified by:   Ximidar
* @Last Modified time: 2018-05-27 23:15:49
*/

package main

import(
    "github.com/Commango/src/comm"
    "fmt"
    "os"
    "io"
    "errors"
    "time"
) 

func main() {

    comm := commango.New_Comm()

    err := comm.Init_Comm("/dev/ttyACM0", 115200, 8, 1)

    if err != nil{
        fmt.Println(err)
        fmt.Println("Cannot assign options")
        os.Exit(2)
    }

    err = comm.Open_Comm()

    if err != nil{
        fmt.Println("Cannot open port")
    }
    defer comm.Close_Comm()

    for count := 0; count < 10; count += 1{
        message := fmt.Sprintf("Hello at time: %v\n", time.Now())
        _, err = comm.Write_Comm_String(message)
    
        out, err := ReadLine(comm.Port)

        if err != nil{
            fmt.Println(err)
        }else{
            //fmt.Println(out)
            fmt.Println(string(out))
        }

        
    }
    
    

    
}

func ReadLine(r io.Reader) (out []byte, err error){
    read_line := false
    for read_line == false{
        read, err := ReadWithTimeout(r, 1)

        if err != nil{
            fmt.Println("Readline errored out", err)
            return nil, err
        }

        if read[0] == 10{
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
