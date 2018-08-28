/*
* @Author: Ximidar
* @Date:   2018-05-27 17:44:35
* @Last Modified by:   Ximidar
* @Last Modified time: 2018-07-29 17:16:42
 */
package main

import (
	"fmt"
	"github.com/ximidar/Commango/src/dbus_conn"
	"os"
)

var COMM_OPEN bool

func main() {
	dconn := dbus_conn.New_DbusConn()

	dconn.Init_Services()

	fmt.Println("Serving", dconn.FullName, dconn.FullNamePath)
	select {}

	fmt.Println("Finished")
	os.Exit(0)

}

