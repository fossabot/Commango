/*
* @Author: Ximidar
* @Date:   2018-05-27 17:44:35
* @Last Modified by:   Ximidar
* @Last Modified time: 2018-09-15 17:51:08
 */
package main

import (
	"fmt"
	"github.com/ximidar/Commango/src/nats_conn"
	"os"
)

var COMM_OPEN bool

func main() {
	gnats := nats_conn.New_NatsConn()
	gnats.Serve()

	fmt.Println("Finished")
	os.Exit(0)

}
