package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Portforward https://github.com/m13253/pofwd
func main() {

	go startForwarding("tcp", "127.0.0.1:8080", "tcp", "192.168.100.10:80")
	go startForwarding("tcp", "127.0.0.1:8081", "tcp", "192.168.100.10:80")

	<-make(chan bool)

}

// 源协议  源主机  目的协议 目的主机
func startForwarding(fromprotocol, fromaddress, toprotocol, toaddress string) error {
	if isPacketProtocol(fromprotocol) {
		return startForwardingPacket(fromprotocol, fromaddress, toprotocol, toaddress)
	}
	return startForwardingStream(fromprotocol, fromaddress, toprotocol, toaddress)
}

// 开始转发流
func startForwardingStream(fromprotocol, fromaddress, toprotocol, toaddress string) error {
	listener, _ := net.Listen(fromprotocol, fromaddress)
	go func() {
		for {
			connin, err := listener.Accept()
			if err != nil {
				if errnet, ok := err.(net.Error); ok {
					if errnet.Temporary() {
						log.Println(err)
						continue
					}
				}
				log.Fatalln(err)
			}
			go func() {
				connout, err := net.Dial(toprotocol, toaddress)
				if err != nil {
					log.Printf("%s %s <---> %s %s <===> %s ? <-!-> %s %s\n", connin.RemoteAddr().Network(), connin.RemoteAddr().String(), connin.LocalAddr().Network(), connin.LocalAddr().String(), toprotocol, toprotocol, toaddress)
					log.Println(err)
					connin.Close()
					return
				}
				log.Printf("%s %s <---> %s %s <===> %s %s <---> %s %s\n", connin.RemoteAddr().Network(), connin.RemoteAddr().String(), connin.LocalAddr().Network(), connin.LocalAddr().String(), connout.LocalAddr().Network(), connout.LocalAddr().String(), connout.RemoteAddr().Network(), connout.RemoteAddr().String())
				go func() {
					var err error
					var packetlen int
					buffer := make([]byte, 2048)
					if isPacketProtocol(toprotocol) {
						for {
							_, err = io.ReadFull(connin, buffer[:2])
							if err != nil {
								break
							}
							packetlen = (int(buffer[0]) << 8) | int(buffer[1])
							if packetlen > 2046 {
								err = &tooLargePacketError{
									Size: packetlen,
								}
								break
							}
							_, err = io.ReadFull(connin, buffer[2:2+packetlen])
							if err != nil {
								break
							}
							_, err = connout.Write(buffer[2 : 2+packetlen])
							if err != nil {
								break
							}
						}
					} else {
						for {
							packetlen, err = connin.Read(buffer)
							if err != nil {
								break
							}
							_, err = connout.Write(buffer[:packetlen])
							if err != nil {
								break
							}
						}
					}
					if err == io.EOF {
						log.Printf("%s %s <---> %s %s ==X=> %s %s <---> %s %s\n", connin.RemoteAddr().Network(), connin.RemoteAddr().String(), connin.LocalAddr().Network(), connin.LocalAddr().String(), connout.LocalAddr().Network(), connout.LocalAddr().String(), connout.RemoteAddr().Network(), connout.RemoteAddr().String())
					} else {
						log.Printf("%s %s <---> %s %s ==!=> %s %s <---> %s %s\n", connin.RemoteAddr().Network(), connin.RemoteAddr().String(), connin.LocalAddr().Network(), connin.LocalAddr().String(), connout.LocalAddr().Network(), connout.LocalAddr().String(), connout.RemoteAddr().Network(), connout.RemoteAddr().String())
						log.Println(err)
					}
					if connintcp, ok := connin.(*net.TCPConn); ok {
						connintcp.CloseRead()
					}
					if connouttcp, ok := connout.(*net.TCPConn); ok {
						connouttcp.CloseWrite()
					} else {
						connout.Close()
					}
				}()
				go func() {
					var err error
					var packetlen int
					buffer := make([]byte, 2048)
					if isPacketProtocol(toprotocol) {
						for {
							connout.SetReadDeadline(time.Now().Add(180 * time.Second))
							packetlen, err = connout.Read(buffer[2:])
							if err != nil {
								break
							}
							buffer[0], buffer[1] = byte(packetlen>>8), byte(packetlen)
							_, err = connin.Write(buffer[:2+packetlen])
							if err != nil {
								break
							}
						}
					} else {
						for {
							packetlen, err = connout.Read(buffer)
							if err != nil {
								break
							}
							_, err = connin.Write(buffer[:packetlen])
							if err != nil {
								break
							}
						}
					}
					if err == io.EOF {
						log.Printf("%s %s <---> %s %s <=X== %s %s <---> %s %s\n", connin.RemoteAddr().Network(), connin.RemoteAddr().String(), connin.LocalAddr().Network(), connin.LocalAddr().String(), connout.LocalAddr().Network(), connout.LocalAddr().String(), connout.RemoteAddr().Network(), connout.RemoteAddr().String())
					} else {
						log.Printf("%s %s <---> %s %s <=!== %s %s <---> %s %s\n", connin.RemoteAddr().Network(), connin.RemoteAddr().String(), connin.LocalAddr().Network(), connin.LocalAddr().String(), connout.LocalAddr().Network(), connout.LocalAddr().String(), connout.RemoteAddr().Network(), connout.RemoteAddr().String())
						log.Println(err)
					}
					if connouttcp, ok := connout.(*net.TCPConn); ok {
						connouttcp.CloseRead()
					}
					if connintcp, ok := connin.(*net.TCPConn); ok {
						connintcp.CloseWrite()
					} else {
						connin.Close()
					}
				}()
			}()
		}
	}()
	return nil
}

// 开始转发数据包
func startForwardingPacket(fromprotocol, fromaddress, toprotocol, toaddress string) error {
	connin, err := net.ListenPacket(fromprotocol, fromaddress)
	if err != nil {
		return err
	}
	log.Printf("serving on %s %s\n", connin.LocalAddr().Network(), connin.LocalAddr().String())
	go func() {
		type pipecache struct {
			Pipe  *io.PipeWriter
			Ready *uintptr
			TTL   time.Time
		}
		type hashableaddr struct {
			Network string
			String  string
		}
		pipes := make(map[hashableaddr]pipecache)
		pipeslock := new(sync.RWMutex)
		go func() {
			for {
				time.Sleep(59 * time.Second)
				now := time.Now()
				for k, v := range pipes {
					if v.TTL.Before(now) {
						pipeslock.Lock()
						delete(pipes, k)
						pipeslock.Unlock()
						v.Pipe.Close()
					}
				}
			}
		}()
		buffer := make([]byte, 2048)
		for {
			packetlen, addrin, err := connin.ReadFrom(buffer)
			if err != nil {
				log.Printf("%s ? <-!-> %s %s <===> %s ? <---> %s %s\n", connin.LocalAddr().Network(), connin.LocalAddr().Network(), connin.LocalAddr().String(), toprotocol, toprotocol, toaddress)
				if errnet, ok := err.(net.Error); ok {
					if errnet.Temporary() {
						log.Println(err)
						continue
					}
				}
				log.Fatalln(err)
			}
			pipeslock.RLock()
			if pipeout, ok := pipes[hashableaddr{
				Network: addrin.Network(),
				String:  addrin.String(),
			}]; ok {
				pipeslock.RUnlock()
				pipeout.TTL = time.Now().Add(180 * time.Second)
				if atomic.LoadUintptr(pipeout.Ready) != 0 {
					pipeout.Pipe.Write(buffer[:packetlen])
				}
			} else {
				pipeslock.RUnlock()
				firstpacket := make([]byte, packetlen)
				copy(firstpacket, buffer)
				go func(addrin net.Addr, firstpacket []byte) {
					connout, err := net.Dial(toprotocol, toaddress)
					if err != nil {
						log.Printf("%s %s <---> %s %s <===> %s ? <-!-> %s %s\n", addrin.Network(), addrin.String(), connin.LocalAddr().Network(), connin.LocalAddr().String(), toprotocol, toprotocol, toaddress)
						log.Println(err)
						return
					}
					log.Printf("%s %s <---> %s %s <===> %s %s <---> %s %s\n", addrin.Network(), addrin.String(), connin.LocalAddr().Network(), connin.LocalAddr().String(), connout.LocalAddr().Network(), connout.LocalAddr().String(), connout.RemoteAddr().Network(), connout.RemoteAddr().String())
					pipein, pipeout := io.Pipe()
					ready := new(uintptr)
					pipe := pipecache{
						Pipe:  pipeout,
						Ready: ready,
						TTL:   time.Now().Add(180 * time.Second),
					}
					pipeslock.Lock()
					pipes[hashableaddr{
						Network: addrin.Network(),
						String:  addrin.String(),
					}] = pipe
					pipeslock.Unlock()
					go func() {
						var err error
						var packetlen int
						buffer := make([]byte, 2048)
						if isPacketProtocol(toprotocol) {
							for {
								atomic.StoreUintptr(ready, 1)
								packetlen, err = pipein.Read(buffer)
								atomic.StoreUintptr(ready, 0)
								if err != nil {
									break
								}
								_, err = connout.Write(buffer[:packetlen])
								if err != nil {
									break
								}
							}
						} else {
							for {
								atomic.StoreUintptr(ready, 1)
								packetlen, err = pipein.Read(buffer[2:])
								atomic.StoreUintptr(ready, 0)
								if err != nil {
									break
								}
								buffer[0], buffer[1] = byte(packetlen>>8), byte(packetlen)
								_, err = connout.Write(buffer[:2+packetlen])
								if err != nil {
									break
								}
							}
						}
						if err == io.EOF {
							log.Printf("%s %s <---> %s %s ==X=> %s %s <---> %s %s\n", addrin.Network(), addrin.String(), connin.LocalAddr().Network(), connin.LocalAddr().String(), connout.LocalAddr().Network(), connout.LocalAddr().String(), connout.RemoteAddr().Network(), connout.RemoteAddr().String())
						} else {
							log.Printf("%s %s <---> %s %s ==!=> %s %s <---> %s %s\n", addrin.Network(), addrin.String(), connin.LocalAddr().Network(), connin.LocalAddr().String(), connout.LocalAddr().Network(), connout.LocalAddr().String(), connout.RemoteAddr().Network(), connout.RemoteAddr().String())
							log.Println(err)
						}
						pipeslock.Lock()
						delete(pipes, hashableaddr{
							Network: addrin.Network(),
							String:  addrin.String(),
						})
						pipeslock.Unlock()
						pipein.Close()
						if connouttcp, ok := connout.(*net.TCPConn); ok {
							connouttcp.CloseWrite()
						} else {
							connout.Close()
						}
					}()
					go func() {
						var err error
						var packetlen int
						buffer := make([]byte, 2048)
						if isPacketProtocol(toprotocol) {
							for {
								connout.SetReadDeadline(time.Now().Add(180 * time.Second))
								packetlen, err = connout.Read(buffer)
								if err != nil {
									break
								}
								_, err = connin.WriteTo(buffer[:packetlen], addrin)
								if err != nil {
									break
								}
							}
						} else {
							for {
								_, err = io.ReadFull(connout, buffer[:2])
								if err != nil {
									break
								}
								packetlen = (int(buffer[0]) << 8) | int(buffer[1])
								if packetlen > 2046 {
									err = &tooLargePacketError{
										Size: packetlen,
									}
									break
								}
								_, err = io.ReadFull(connout, buffer[2:2+packetlen])
								if err != nil {
									break
								}
								_, err = connin.WriteTo(buffer[2:2+packetlen], addrin)
								if err != nil {
									break
								}
							}
						}
						if err == io.EOF {
							log.Printf("%s %s <---> %s %s <=X== %s %s <---> %s %s\n", addrin.Network(), addrin.String(), connin.LocalAddr().Network(), connin.LocalAddr().String(), connout.LocalAddr().Network(), connout.LocalAddr().String(), connout.RemoteAddr().Network(), connout.RemoteAddr().String())
						} else {
							log.Printf("%s %s <---> %s %s <=!== %s %s <---> %s %s\n", addrin.Network(), addrin.String(), connin.LocalAddr().Network(), connin.LocalAddr().String(), connout.LocalAddr().Network(), connout.LocalAddr().String(), connout.RemoteAddr().Network(), connout.RemoteAddr().String())
							log.Println(err)
						}
						if connouttcp, ok := connout.(*net.TCPConn); ok {
							connouttcp.CloseRead()
						}
					}()
					pipeout.Write(firstpacket)
				}(addrin, firstpacket)
			}
		}
	}()
	return nil
}

func isPacketProtocol(protocolname string) bool {
	switch strings.ToLower(protocolname) {
	case "udp", "udp4", "udp6", "ip", "ip4", "ip6", "unixgram":
		return true
	default: // "tcp", "tcp4", "tcp6", "unix", "unixpacket"
		return false
	}
}

type tooLargePacketError struct {
	Size int
}

func (e *tooLargePacketError) Error() string {
	return fmt.Sprintf("packet too large (%d > 2046)", e.Size)
}
