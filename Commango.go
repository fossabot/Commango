/*
* @Author: Ximidar
* @Date:   2018-05-27 17:44:35
* @Last Modified by:   Ximidar
* @Last Modified time: 2018-10-01 01:07:57
 */
package main

import (
	"fmt"
	"github.com/ximidar/Flotilla/Commango/nats_conn"
	"os"
)

var COMM_OPEN bool

func main() {
	gnats := nats_conn.New_NatsConn()
	gnats.Serve()

	fmt.Println("Finished")
	os.Exit(0)

}
