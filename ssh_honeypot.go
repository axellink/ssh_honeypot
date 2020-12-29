package main

import (
	"fmt"
	"net"
	"io/ioutil"
	"strings"
	"golang.org/x/crypto/ssh"
	"flag"
	"regexp"
	"os"
	"strconv"
)

func LoadHostKey (file string) (ssh.Signer, error){
	keyBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil,err
	}

	key, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		return nil,err
	}

	return key,nil
}

func PortBindServer(port string, config ssh.ServerConfig, check chan bool){
	listener, err := net.Listen("tcp",":" + port)
	if err != nil {
		fmt.Println("Cannot bind to port " + port + " :", err)
		check <- false
		return
	}

	fmt.Println("Server listening on port " + port)
	check <- true

	for {
		nConn, err := listener.Accept()
		if err != nil {
			fmt.Println("Cannot accept connection :", err)
		}
		go func(nConn net.Conn, config ssh.ServerConfig){
			defer nConn.Close()
			ssh.NewServerConn(nConn, &config)
		}(nConn, config)
	}
}

func portsArray(commaPorts string) ([]string, error){
	controlRegex := "^([0-9]{1,5},)*[0-9]{1,5}$"
	if match,_ := regexp.MatchString(controlRegex, commaPorts); !match {
		return nil, fmt.Errorf("Malformed ports option")
	}
	ports := strings.Split(commaPorts,",")
	for _,p := range ports {
		if intp,_ := strconv.Atoi(p); (intp>65535 || intp == 0) {
			return nil, fmt.Errorf("Port %s not in range", p)
		}
	}
	return ports, nil
}

func loadCLI() ([]string, ssh.Signer, *os.File){
	var hostkeyFile string
	var commaPorts string
	var infoFilename string

	flag.StringVar(&hostkeyFile, "k", "/etc/ssh/ssh_host_rsa_key", "Host key file path")
	flag.StringVar(&commaPorts, "p", "22", "Comma separated ports list")
	flag.StringVar(&infoFilename, "f", "/dev/null", "File where writing logging info")
	flag.Parse()

	ports,err := portsArray(commaPorts)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	hostkey,err := LoadHostKey(hostkeyFile)
	if err != nil {
		fmt.Println("Could not load host key", hostkeyFile, ":", err)
		os.Exit(2)
	}

	infoFile, err := os.OpenFile(infoFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY,0644)
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}
	return ports,hostkey, infoFile
}

func ParseIPAddr(addr string) (string,string){
	var ipaddr string
	var port string
	if string(addr[0]) == "[" { //IPv6 Address
		ipaddr = strings.Split(string(addr[1:]),"]")[0]
		port = string(strings.Split(addr,"]")[1][1:])
	}else{
		ipaddr = strings.Split(addr,":")[0]
		port = strings.Split(addr,":")[1]
	}
	return ipaddr, port
}

func main() {
	ports,hostkey,infoFile := loadCLI()
	defer infoFile.Close()

	infoChannel := make(chan string)
	config := ssh.ServerConfig {
		PasswordCallback: func (c ssh.ConnMetadata, pass []byte) (*ssh.Permissions,error) {
			localAddr,localPort := ParseIPAddr(c.LocalAddr().String())
			remoteAddr,_ := ParseIPAddr(c.RemoteAddr().String())
			info := fmt.Sprintf("%s#;#%s#;#%s#;#%s#;#%s", remoteAddr, localAddr, localPort, c.User(), pass)
			infoChannel <- info
			return nil, fmt.Errorf("Incorrect password")
		},
	}
	config.AddHostKey(hostkey)

	oneSucceed := false
	oneFailed := false
	for _,p := range ports {
		check := make(chan bool, 1)
		go PortBindServer(p,config, check)
		if <- check {
			oneSucceed = true
		}else{
			oneFailed = true
		}
	}

	if ! oneSucceed {
		fmt.Println("Fatal : No server bound")
		os.Exit(4)
	}else if oneFailed {
		fmt.Println("Warning : some servers are not bound to port")
	}

	for info := range infoChannel {
		infoFile.Write([]byte(info + "\n"))
	}
}
