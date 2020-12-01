package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

type client chan <- string

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan string)
	counts   = make(chan struct{}, 100)
)

func broadcast() {
	clients := make(map[client]bool)
	for {
		select {
		case msg := <-messages:
			for ch := range clients {
				ch <- msg
			}
		case ch := <- entering :
			clients[ch] = true
		case ch := <- leaving :
			delete(clients, ch)
			close(ch)
		}
	}
}

func receive(conn net.Conn, ch <- chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}

func handleFunc(conn net.Conn) {
	//counts <- struct {}{}
	//defer func() {
	//	<- counts
	//} ()
	select {
		case counts <- struct {}{} :
		defer func () {
			<- counts
		} ()
	default :
		fmt.Fprintln(conn, "failed")
		return
	}

	ch := make(chan string)
	go receive(conn, ch)

	who := conn.RemoteAddr().String()
	ch <- " You are " + who
	messages <- who + " are arrived "
	entering <- ch

	input := bufio.NewScanner(conn)
	for input.Scan() {
		messages <- who + " : " + input.Text()
	}

	leaving <- ch
	messages <- who + " has left"
	conn.Close()
}

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:8000")
	if err != nil {
		log.Fatal(err)
	}
	go broadcast()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleFunc(conn)
	}
}