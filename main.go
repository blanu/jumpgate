package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
)

func main() {
	handlers := make([]context.CancelFunc, 0)
	defer func() {
		// If the main program exits, explicitly close all connections first
		for _, handler := range handlers {
			handler()
		}
	}()

	listener, listenError := net.Listen("tcp", "0.0.0.0:7777")
	if listenError != nil {
		os.Exit(10)
	}

	for {
		connection, acceptError := listener.Accept()
		if acceptError != nil {
			os.Exit(11)
		}

		go handleConnection(connection, &handlers)
	}
}

func handleConnection(connection net.Conn, handlers *[]context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	*handlers = append(*handlers, cancel)

	home, homeError := os.UserHomeDir()
	if homeError != nil {
		print("could not find home directory")
		home = "/root"
	}
	command := fmt.Sprintf("%s/go/bin/wormhole", home)

	var router *exec.Cmd
	router = exec.CommandContext(ctx, command)

	file, castError := connection.(*net.TCPConn).File()
	if castError != nil {
		return
	}
	router.ExtraFiles = []*os.File{file}
	startError := router.Start()
	if startError != nil {
		return
	}

	go func() {
		err := router.Wait()
		if err != nil {
			print(err)
			return
		}
	}()
}
