/*
* @Author: Ximidar
* @Date:   2018-07-28 11:10:37
* @Last Modified by:   Ximidar
* @Last Modified time: 2018-10-01 01:08:25
 */

package nats_conn

import (
	"encoding/json"
	"fmt"
	"github.com/nats-io/go-nats"
	"github.com/ximidar/Flotilla/Commango/comm"
	ms "github.com/ximidar/Flotilla/data_structures"
	"log"
	_ "os"
	"strconv"
	_ "strings"
)

const (
	// address name
	NAME = "commango."

	// reply subs
	LIST_PORTS       = NAME + "list_ports"
	INIT_COMM        = NAME + "init_comm"
	CONNECT_COMM     = NAME + "connect_comm"
	DISTCONNECT_COMM = NAME + "disconnect_comm"
	WRITE_COMM       = NAME + "write_comm"
	GET_STATUS       = NAME + "get_status"

	// pubs
	READ_LINE = NAME + "read_line"
	WRITE_LINE = NAME + "write_line"
	STATUS_UPDATE = NAME + "status_update"
)

type NatsConn struct {
	NC *nats.Conn

	//comm connection
	Comm *commango.Comm
}

func New_NatsConn() *NatsConn {

	gnats := new(NatsConn)
	err := error(nil)
	gnats.NC, err = nats.Connect(nats.DefaultURL)

	if err != nil {
		log.Fatalf("Can't connect: %v\n", err)
	}

	gnats.Comm = commango.New_Comm(gnats.Read_Line_Emitter, gnats.Write_Line_Emitter, gnats.Publish_Status)
	gnats.create_req_replies()

	return gnats
}

func (gnats *NatsConn) Serve() {
	select {} //TODO make this select detect shutdown signals
}

func (gnats *NatsConn) create_req_replies() error {
	// req replies
	gnats.NC.Subscribe(LIST_PORTS, gnats.list_ports)
	gnats.NC.Subscribe(INIT_COMM, gnats.init_comm)
	gnats.NC.Subscribe(CONNECT_COMM, gnats.connect_comm)
	gnats.NC.Subscribe(DISTCONNECT_COMM, gnats.disconnect_comm)
	gnats.NC.Subscribe(WRITE_COMM, gnats.write_comm)
	gnats.NC.Subscribe(GET_STATUS, gnats.get_status)
	return nil
}

func (gnats *NatsConn) list_ports(msg *nats.Msg) {
	reply := new(ms.Reply_JSON)
	ports, err := gnats.Comm.Get_Available_Ports()
	if err != nil {
		reply.Success = false
		reply.Message = []byte(err.Error())
	} else {
		reply.Success = true
		reply.Message, err = json.Marshal(ports)
	}

	m_reply, err := json.Marshal(reply)

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	gnats.NC.Publish(msg.Reply, m_reply)
}

func (gnats *NatsConn) get_status(msg *nats.Msg){
	reply := new(ms.Reply_JSON)
	status := gnats.Comm.Get_Comm_Status()
	reply.Success = true
	mdata, _ := json.Marshal(status)
	reply.Message = mdata
	rep_bytes, _ := json.Marshal(reply)
	gnats.NC.Publish(msg.Reply, rep_bytes)
}

func (gnats *NatsConn) Publish_Status(status *ms.Comm_Status){
	reply := new(ms.Reply_JSON)
	reply.Success = true
	mdata, _ := json.Marshal(status)
	reply.Message = mdata
	rep_bytes, _ := json.Marshal(reply)
	gnats.NC.Publish(STATUS_UPDATE, rep_bytes)

}

func (gnats *NatsConn) init_comm(msg *nats.Msg) {

	reply := new(ms.Reply_String)
	u_init_data := new(ms.Init_Comm)
	err := json.Unmarshal(msg.Data, &u_init_data)

	// error out if we cannot unmarshal the data
	if err != nil {
		reply.Success = false
		reply.Message = "could not unmarshal data"
		rep_byte, _ := json.Marshal(reply)
		gnats.NC.Publish(msg.Reply, rep_byte)
		return
	}

	// error out if we cannot initialize the comm
	err = gnats.Comm.Init_Comm(u_init_data.Port, u_init_data.Baud)
	if err != nil {
		reply.Success = false
		reply.Message = "Could Not Initialize Comm: " + err.Error()
		rep_byte, _ := json.Marshal(reply)
		gnats.NC.Publish(msg.Reply, rep_byte)
		return
	}

	// Create success response and send
	reply.Success = true
	reply.Message = "Comm Initialized"
	m_reply, err := json.Marshal(reply)
	if err != nil {
		panic(err)
	} // There should be no reason it can't marshal
	gnats.NC.Publish(msg.Reply, m_reply)

}

func (gnats *NatsConn) connect_comm(msg *nats.Msg) {
	err := gnats.Comm.Open_Comm()
	reply := new(ms.Reply_String)
	if err != nil {
		reply.Success = false
		reply.Message = err.Error()
		rep, _ := json.Marshal(reply)
		gnats.NC.Publish(msg.Reply, rep)
		return
	}
	reply.Success = true
	reply.Message = "Connected"
	m_reply, err := json.Marshal(reply)
	gnats.NC.Publish(msg.Reply, m_reply)
}

func (gnats *NatsConn) disconnect_comm(msg *nats.Msg) {
	err := gnats.Comm.Close_Comm()
	reply := new(ms.Reply_String)
	if err != nil {
		reply.Success = false
		reply.Message = err.Error()
		rep, _ := json.Marshal(reply)
		gnats.NC.Publish(msg.Reply, rep)
		return
	}
	reply.Success = true
	reply.Message = "Disconnected"
	m_reply, _ := json.Marshal(reply)
	gnats.NC.Publish(msg.Reply, m_reply)
}

func (gnats *NatsConn) write_comm(msg *nats.Msg) {
	bytes_written, err := gnats.Comm.Write_Comm(string(msg.Data))
	reply := new(ms.Reply_String)

	if err != nil {
		reply.Success = false
		reply.Message = err.Error()
		rep, _ := json.Marshal(reply)
		gnats.NC.Publish(msg.Reply, rep)
	}

	reply.Success = true
	reply.Message = strconv.Itoa(bytes_written)
	m_reply, _ := json.Marshal(reply)

	gnats.NC.Publish(msg.Reply, m_reply)
}

func (gnats *NatsConn) Read_Line_Emitter(line string) {
	gnats.NC.Publish(READ_LINE, []byte(line))
}

func (gnats *NatsConn) Write_Line_Emitter(line string) {
	gnats.NC.Publish(WRITE_LINE, []byte(line))
}

