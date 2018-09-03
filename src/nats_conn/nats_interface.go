/*
* @Author: Ximidar
* @Date:   2018-07-28 11:10:37
* @Last Modified by:   Ximidar
* @Last Modified time: 2018-09-02 17:49:46
 */

package nats_conn

import (
	"fmt"
	"github.com/ximidar/Commango/src/comm"
	ms "github.com/ximidar/mango_structures"
	"github.com/nats-io/go-nats"
	"os"
	"strings"
	"encoding/json"
	
)

const(
	// address name
	NAME = "commango."

	// reply subs
	LIST_PORTS = NAME + "list_ports"
	INIT_COMM = NAME + "init_comm"
	CONNECT_COMM = NAME + "connect_comm"
	DISTCONNECT_COMM = NAME + "disconnect_comm"
	WRITE_COMM = NAME + "write_comm"

	// pubs
	READ_LINE = NAME + "read_line"
)

type NatsConn struct {

	NC *nats.Conn

	//comm connection
	Comm *commango.Comm
}

func New_NatsConn() *NatsConn {

	gnats := new(NatsConn)
	gnats.nc, err := nats.Connect(nats.DefaultURL)

	if err != nil {
		log.Fatalf("Can't connect: %v\n", err)
	}

	gnats.Comm = commango.New_Comm()

	return gnats
}


func (gnats *NatsConn) create_req_replies() (error){
	// req replies
	gnats.nc.Subscribe(LIST_PORTS, gnats.list_ports)
	gnats.nc.Subscribe(INIT_COMM, gnats.init_comm)
	gnats.nc.Subscribe(CONNECT_COMM, gnats.connect_comm)
	gnats.nc.Subscribe(DISTCONNECT_COMM, gnats.disconnect_comm)
	gnats.nc.Subscribe(WRITE_COMM, gnats.write_comm)
}

func (gnats *NatsConn) list_ports(msg *nats.Msg) {
	reply := new(ms.Reply)
	ports, err := gnats.Comm.Get_Available_Ports()
	if err != nil{
		reply.Success = false
		reply.Message = err.Error()
	} else {
		reply.Success = true
		reply.Message = json.Marshal(ports)
	}

	m_reply, err := json.Marshal(reply)

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	nc.Publish(msg.Reply, m_reply)
}

func (gnats *NatsConn) init_comm(msg *nats.Msg) {

	reply := new(ms.Reply)
	init_data := new(ms.Init_Comm)
	u_init_data, err := json.Unmarshal(msg.Data, &init_data)

	// error out if we cannot unmarshal the data
	if err != nil {
		reply.Success = false
		reply.Message = "could not unmarshal data"
		nc.Publish(msg.Reply, []byte(reply))
		return
	}

	// error out if we cannot initialize the comm
	err = gnats.Comm.Init_Comm(u_init_data.Port, u_init_data.Baud)
	if err != nil {
		reply.Success = false
		reply.Message = "Could Not Initialize Comm: " + err.Error()
		nc.Publish(msg.Reply, []byte(reply))
		return
	}

	// Create success response and send
	reply.Success = true
	reply.Message = "Comm Initialized"
	m_reply, err := json.Marshal(reply)
	if err != nil {panic(err)} // There should be no reason it can't marshal
	nc.Publish(msg.Reply, m_reply)
	
}

func (gnats *NatsConn) connect_comm(msg *nats.Msg) {
	err := gnats.Comm.Open_Comm()
	reply = new(Reply)
	if err != nil {
		reply.Success = false
		reply.Message = err.Error()
		nc.Publish(msg.Reply, []byte(reply))
		return
	} 
	reply.Success = true
	reply.Message = "Connected"
	m_reply, err := json.Marshal(reply)
	nc.Publish(msg.Reply, []byte(reply))
}

func (gnats *NatsConn) disconnect_comm(msg *nats.Msg) {
	err := gnats.Comm.Close_Comm()
	reply = new(Reply)
	if err != nil {
		reply.Success = false
		reply.Message = err.Error()
		nc.Publish(msg.Reply, []byte(reply))
		return
	} 
	reply.Success = true
	reply.Message = "Disconnected"
	m_reply, err := json.Marshal(reply)
	nc.Publish(msg.Reply, []byte(reply))
}

func (gnats *NatsConn) write_comm(msg *nats.Msg) {
	bytes_written, err := gnats.Comm.Write_Comm(string(msg.Data))
	nc.Publish(msg.Reply, []byte(bytes_written))
}


