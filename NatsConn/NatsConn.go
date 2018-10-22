/*
* @Author: Ximidar
* @Date:   2018-07-28 11:10:37
* @Last Modified by:   Ximidar
* @Last Modified time: 2018-10-18 15:24:12
 */

package NatsConn

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/nats-io/go-nats"
	"github.com/ximidar/Flotilla/Commango/comm"
	DS "github.com/ximidar/Flotilla/DataStructures"
	CS "github.com/ximidar/Flotilla/DataStructures/CommStructures"
)

// NatsConn is the Comm interface to the Nats Server
type NatsConn struct {
	NC *nats.Conn

	//comm connection
	Comm *commango.Comm
}

// NewNatsConn will construct a NatsConn object
func NewNatsConn() *NatsConn {

	gnats := new(NatsConn)
	var err error
	gnats.NC, err = nats.Connect(nats.DefaultURL)

	if err != nil {
		log.Fatalf("Can't connect: %v\n", err)
	}

	gnats.Comm = commango.NewComm(gnats.ReadLineEmitter, gnats.WriteLineEmitter, gnats.PublishStatus)
	gnats.createReqReplies()

	return gnats
}

// Serve will keep the program open
func (gnats *NatsConn) Serve() {
	select {} //TODO make this select detect shutdown signals
}

func (gnats *NatsConn) createReqReplies() (err error) {
	// req replies
	_, err = gnats.NC.Subscribe(CS.ListPorts, gnats.listPorts)
	_, err = gnats.NC.Subscribe(CS.InitializeComm, gnats.initComm)
	_, err = gnats.NC.Subscribe(CS.ConnectComm, gnats.connectComm)
	_, err = gnats.NC.Subscribe(CS.DisconnectComm, gnats.disconnectComm)
	_, err = gnats.NC.Subscribe(CS.WriteComm, gnats.writeComm)
	_, err = gnats.NC.Subscribe(CS.GetStatus, gnats.getStatus)

	return err
}

func (gnats *NatsConn) listPorts(msg *nats.Msg) {
	reply := new(DS.ReplyJSON)
	ports, err := gnats.Comm.GetAvailablePorts()
	if err != nil {
		reply.Success = false
		reply.Message = []byte(err.Error())
	} else {
		reply.Success = true
		reply.Message, err = json.Marshal(ports)
	}

	mReply, err := json.Marshal(reply)

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	gnats.NC.Publish(msg.Reply, mReply)
}

func (gnats *NatsConn) getStatus(msg *nats.Msg) {
	reply := new(DS.ReplyJSON)
	status := gnats.Comm.GetCommStatus()
	reply.Success = true
	mdata, _ := json.Marshal(status)
	reply.Message = mdata
	repBytes, _ := json.Marshal(reply)
	gnats.NC.Publish(msg.Reply, repBytes)
}

// PublishStatus will publish the Comm status to the Nats Server
func (gnats *NatsConn) PublishStatus(status *CS.CommStatus) {
	reply := new(DS.ReplyJSON)
	reply.Success = true
	mdata, _ := json.Marshal(status)
	reply.Message = mdata
	repBytes, _ := json.Marshal(reply)
	gnats.NC.Publish(CS.StatusUpdate, repBytes)

}

func (gnats *NatsConn) initComm(msg *nats.Msg) {

	reply := new(DS.ReplyString)
	uInitData := new(CS.InitComm)
	err := json.Unmarshal(msg.Data, &uInitData)

	// error out if we cannot unmarshal the data
	if err != nil {
		reply.Success = false
		reply.Message = "could not unmarshal data"
		repByte, _ := json.Marshal(reply)
		gnats.NC.Publish(msg.Reply, repByte)
		return
	}

	// error out if we cannot initialize the comm
	err = gnats.Comm.InitComm(uInitData.Port, uInitData.Baud)
	if err != nil {
		reply.Success = false
		reply.Message = "Could Not Initialize Comm: " + err.Error()
		repByte, _ := json.Marshal(reply)
		gnats.NC.Publish(msg.Reply, repByte)
		return
	}

	// Create success response and send
	reply.Success = true
	reply.Message = "Comm Initialized"
	mReply, err := json.Marshal(reply)
	if err != nil {
		panic(err)
	} // There should be no reason it can't marshal
	gnats.NC.Publish(msg.Reply, mReply)

}

func (gnats *NatsConn) connectComm(msg *nats.Msg) {
	err := gnats.Comm.OpenComm()
	reply := new(DS.ReplyString)
	if err != nil {
		reply.Success = false
		reply.Message = err.Error()
		rep, _ := json.Marshal(reply)
		gnats.NC.Publish(msg.Reply, rep)
		return
	}
	reply.Success = true
	reply.Message = "Connected"
	mReply, err := json.Marshal(reply)
	gnats.NC.Publish(msg.Reply, mReply)
}

func (gnats *NatsConn) disconnectComm(msg *nats.Msg) {
	err := gnats.Comm.CloseComm()
	reply := new(DS.ReplyString)
	if err != nil {
		reply.Success = false
		reply.Message = err.Error()
		rep, _ := json.Marshal(reply)
		gnats.NC.Publish(msg.Reply, rep)
		return
	}
	reply.Success = true
	reply.Message = "Disconnected"
	mReply, _ := json.Marshal(reply)
	gnats.NC.Publish(msg.Reply, mReply)
}

func (gnats *NatsConn) writeComm(msg *nats.Msg) {
	bytesWritten, err := gnats.Comm.WriteComm(string(msg.Data))
	reply := new(DS.ReplyString)

	if err != nil {
		reply.Success = false
		reply.Message = err.Error()
		rep, _ := json.Marshal(reply)
		gnats.NC.Publish(msg.Reply, rep)
	}

	reply.Success = true
	reply.Message = strconv.Itoa(bytesWritten)
	mReply, _ := json.Marshal(reply)

	gnats.NC.Publish(msg.Reply, mReply)
}

// ReadLineEmmitter will publish any Read lines from Comm
func (gnats *NatsConn) ReadLineEmitter(line string) {
	gnats.NC.Publish(CS.ReadLine, []byte(line))
}

// WriteLineEmitter will publish any Written lines to Comm
func (gnats *NatsConn) WriteLineEmitter(line string) {
	gnats.NC.Publish(CS.WriteLine, []byte(line))
}
