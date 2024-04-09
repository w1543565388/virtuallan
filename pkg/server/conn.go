package server

import (
	"fmt"
	"io"
	"math/rand"
	"net"

	"github.com/lucheng0127/virtuallan/pkg/utils"
	"github.com/songgao/water"
)

func HandleConn(conn net.Conn) {
	// Create tap
	config := new(water.Config)
	config.DeviceType = water.TAP
	config.Name = "tap0"

	iface, err := water.New(*config)
	if err != nil {
		fmt.Println(err)
	}

	if err = utils.AsignAddrToLink("tap0", "192.168.123.1/24", true); err != nil {
		fmt.Println(err)
	}

	go io.Copy(iface, conn)
	io.Copy(conn, iface)
}

type UClient struct {
	Conn  *net.UDPConn
	RAddr *net.UDPAddr
	Iface *water.Interface
}

var UPool map[string]*UClient

func init() {
	UPool = make(map[string]*UClient)
}

func RandStr(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func NewTap(br string) (*water.Interface, error) {
	config := new(water.Config)
	config.DeviceType = water.TAP
	config.Name = fmt.Sprintf("tap-%s", RandStr(4))

	iface, err := water.New(*config)
	if err != nil {
		return nil, err
	}

	err = utils.SetLinkMaster(config.Name, br)
	if err != nil {
		return nil, err
	}

	return iface, nil
}

func (svc *Server) GetClientForAddr(addr *net.UDPAddr, conn *net.UDPConn) (*UClient, error) {
	client, ok := UPool[addr.String()]
	if ok {
		return client, nil
	}

	iface, err := NewTap(svc.Bridge)
	if err != nil {
		return nil, err
	}

	client = new(UClient)
	client.Iface = iface
	client.RAddr = addr
	client.Conn = conn
	UPool[addr.String()] = client
	return client, nil
}
