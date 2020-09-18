package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
)

func netListen() {
	log.Print("Opening network port 6743")
	ln, err := net.Listen("tcp", ":6743")
	if err != nil {
		log.Fatal("Unable to open port for listening")
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Error on incoming connection: %s", err.Error())
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	log.Print("new incoming net conn")

	for {
		msg, err := bufio.NewReader(conn).ReadBytes('\n')
		if err != nil {
			log.Print(err.Error())
			break
		}
		log.Printf("incoming message: %s", msg)
		go parseClientMessage(conn, msg)
	}

	conn.Close()
}

func netSendMessage(conn net.Conn, msg Message) error {
	raw, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	raw = append(raw, "\n"...)
	return netSendBytes(conn, raw)
}

func netSendBytes(conn net.Conn, message []byte) (err error) {
	if conn != nil {
		log.Printf("Sending message to client %s: %s", conn.RemoteAddr().String(), message)
		_, err = conn.Write(message)
	}

	return err
}
