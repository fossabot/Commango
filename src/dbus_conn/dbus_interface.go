/*
* @Author: Ximidar
* @Date:   2018-07-28 11:10:37
* @Last Modified by:   Ximidar
* @Last Modified time: 2018-07-28 19:48:00
*/

package dbus_conn

import (
	"fmt"
	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"
	"os"
	"strings"

)

type DbusConn struct{
	Name string
	Program string
	FullName string
	FullNamePath string
	FullNameObjectPath dbus.ObjectPath

	SessionBus *dbus.Conn
}

func New_DbusConn() *DbusConn{
	dconn := new(DbusConn)

	dconn.Name = "commango"
	dconn.Program = "com.mango_core"
	dconn.FullName = dconn.Program + "." + dconn.Name
	dconn.FullNamePath = "/" + strings.Replace(dconn.FullName, ".", "/", -1)
	dconn.FullNameObjectPath = dbus.ObjectPath(dconn.FullNamePath)

	var err error
	dconn.SessionBus, err = dbus.SessionBus()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to session bus:", err)
		panic(err)
	}
	dconn.MakeName()
	return dconn
}

func (dconn *DbusConn) MakeName() (err error){
	fmt.Println(fmt.Sprintf("Full Name: %v", dconn.FullName))
	reply, err := dconn.SessionBus.RequestName(dconn.FullName, dbus.NameFlagDoNotQueue)

	if err != nil {
		panic(err)
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("name: %v already taken", dconn.FullName))
		panic(err)
	}
	return
}

func (dconn *DbusConn) ListNames() (names []string, err error){	

	err = dconn.SessionBus.BusObject().Call("org.freedesktop.DBus.ListNames", 0).Store(&names)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to get list of owned names:", err)
		panic(err)
		return
	}

	fmt.Println("Currently owned names on the session bus:")
	for _, v := range names {
		if v[0] != ':'{
			fmt.Println(v)
		}
	}

	return
}

// Start the services for controlling this program
func (dconn *DbusConn) Init_Services() (err error){
	err = dconn.Init_Connection_Info()
	return 
}	

func (dconn *DbusConn) Init_Connection_Info() (err error){
	return dconn.Make_Functions()
}

func (dconn *DbusConn) Make_Functions()(err error){
	connection_info_xml := `
<node>
	<interface name="` + dconn.FullName + `">
		<method name="Get_Connection_info">
			<arg direction="out" type="s"/>
		</method>
	</interface>` + introspect.IntrospectDataString + `
</node> `
	err = dconn.SessionBus.Export(dconn, dconn.FullNameObjectPath, dconn.FullName)
	err = dconn.SessionBus.Export(introspect.Introspectable(connection_info_xml), dconn.FullNameObjectPath, "org.freedesktop.DBus.Introspectable")

	if err != nil{
		fmt.Println("AHHHHHHHHHHHHHH")
		return err
	}
	return
}

func (dconn DbusConn) Get_Connection_info() (string, *dbus.Error){
	return "Hello This is a test!", nil
}