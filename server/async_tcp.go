package server

import (
	"log"
	"net"
	"syscall"

	"github.com/NamanG22/Redis/config"
	"github.com/NamanG22/Redis/core"
)

var con_clients int = 0

func RunAsyncTCPServer() error{
	log.Println("Starting async TCP server on port", config.Host, config.Port)

	max_clients := 20000

	events := make([]syscall.Kevent_t, max_clients)

	// Create a socket listener (SetNonblock below; macOS does not accept O_NONBLOCK in type)
	serverFD, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		return err
	}
	defer syscall.Close(serverFD) 	

	if err = syscall.SetNonblock(serverFD, true); err != nil {
		return err
	}

	ip4 := net.ParseIP(config.Host)
	if err = syscall.Bind(serverFD, &syscall.SockaddrInet4{Port: config.Port, Addr: [4]byte{ip4[0], ip4[1], ip4[2], ip4[3]}}); err != nil {
		return err
	}

	if err = syscall.Listen(serverFD, max_clients); err != nil {
		return err
	}

	kqueueFD, err := syscall.Kqueue()
	if err != nil {
		log.Fatal(err)
	}
	defer syscall.Close(kqueueFD)

	var socketServerEvent syscall.Kevent_t = syscall.Kevent_t{
		Ident:  uint64(serverFD),
		Filter: syscall.EVFILT_READ,
		Flags:  syscall.EV_ADD,
	}

	if _, err = syscall.Kevent(kqueueFD, []syscall.Kevent_t{socketServerEvent}, nil, nil); err != nil {
		return err
	}

	for {
		nevents, err := syscall.Kevent(kqueueFD, nil, events, nil)
		if err != nil {
			continue
		}

		for i := 0; i < nevents; i++ {
			if int(events[i].Ident) == serverFD {
				// accept the incoming connection from a client
				fd, _, err := syscall.Accept(serverFD)
				if err != nil {
					log.Println("err",err)
					continue
				}

				//increase the number of concurrent clients count
				con_clients++
				log.Println("new client connected", con_clients)
				syscall.SetNonblock(fd, true)

				// add this new TCP connection to be monitored
				var socketClientEvent syscall.Kevent_t = syscall.Kevent_t{
					Ident:  uint64(fd),
					Filter: syscall.EVFILT_READ,
					Flags:  syscall.EV_ADD,
				}

				// register the new file descriptor to be monitored
				if _, err = syscall.Kevent(kqueueFD, []syscall.Kevent_t{socketClientEvent}, nil, nil); err != nil {
					log.Fatal(err)
				}
			} else{
				comm :=  core.FDComm{
					FD: int(events[i].Ident)}
				cmd, err := readCommand(comm)
				if err != nil {
					syscall.Close(int(events[i].Ident))
					con_clients--
					log.Println("client disconnected", con_clients)
					continue
				}
				respond(comm, cmd)
			}
		}
	}
}