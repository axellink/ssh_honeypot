# Introduction #
Hi everyone, here is a little Golang file that makes use of crypto/ssh to create a false SSH server that will always deny access (it doesn't even open terminal) while logging source address and login/password used.

Please keep in mind that this is an afternoon project to familiarize myself with Golang, it's my first writing after I learned it with Gotour

# Dependencies #
It requires just crypto/ssh
`go get golang.org/x/crypto/ssh`

# Build #
Just Go build it
`go build ssh_honeypot.go`

# Usage #
Just launch it providing an host key file, a comma seperated list of ports and where to log connection as CSV :
`./ssh_honeypot -f log.csv -p 22,2222 -k host_key`

Here is the help :
```
./ssh_honeypot -h                               
Usage of ./ssh_honeypot:
  -f string
        File where writing logging info (default "/dev/null")
  -k string
        Host key file path (default "/etc/ssh/ssh_host_rsa_key")
  -p string
        Comma separated ports list (default "22")
```

# Enhancement #
If I find time and motivation, I will make a better use of log facilities. This project is not meaned to accept modification from others since it's for personal skill upgrade, and well there are plenty of ssh honeypot around that are way better than this one
