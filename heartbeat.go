package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type Device struct {
	IP string `json:"ip,omitempty"`
}

var devices []Device
var ipAddress []string
var ipAddressCount int
var newipAddressesCount int

func GetDeviceEndpoint(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	for _, item := range devices {
		if item.IP == params["ip"] {
			json.NewEncoder(w).Encode(item)
			return
		}
	}
	json.NewEncoder(w).Encode(&Device{})
}

func GetdevicesEndpoint(w http.ResponseWriter, req *http.Request) {
	json.NewEncoder(w).Encode(devices)
}

func GetHeartbeat(w http.ResponseWriter, req *http.Request) {
	// Ticker for checking the heartbeat of the present snmp IP addresses
	ticker1 := time.NewTicker(time.Second * 1)
	go func() {
		for range ticker1.C {
			check(ipAddress)
		}
	}()
	// ticker for checking new snmp devices
	ticker := time.NewTicker(time.Second * 30)
	go func() {
		for range ticker.C {
			snmpDevices()
		}
	}()

	time.Sleep(time.Minute * 60)
	ticker.Stop()

}

func check(ipAddress []string) {
	fmt.Println("--------------------------------------------")
	for i := 0; i < ipAddressCount; {
		if len(ipAddress[i]) > 5 {
			result, _ := exec.Command("/bin/sh", "-c", "ping -c 1 "+ipAddress[i]+" | grep rtt ").Output()
			str1 := string(result)
			if len(str1) < 5 {
				ip := ipAddress[i]
				fmt.Printf("the ip %s is down \n", ip)
				if strings.Compare(ip, "10.132.32.211") == 0 || strings.Compare(ip, "10.132.32.243") == 0 {
					fmt.Println("Removing ", ip)
					j := i
					copy(ipAddress[j:], ipAddress[j+1:]) // Shift a[i+1:] left one index
					ipAddress[len(ipAddress)-1] = ""     // Erase last element (write zero value)
					ipAddress = ipAddress[:len(ipAddress)-1]
					ipAddressCount--
				} else {
					i++
				}
			} else {
				fmt.Printf("the ip %s is up and running \n", ipAddress[i])
				i++
			}
		} else {
			i++
		}
	}
}

func snmpDevices() []string {
	result, _ := exec.Command("/bin/sh", "-c", "echo tcs123| sudo -S nmap -sU 10.132.32.198-255 -p 161 --open -oG - | awk '/161\\/open.*/{print $2}'").Output()
	str1 := string(result)
	ipAddress1 := strings.Split(str1, "\n")
	newipAddressesCount = len(ipAddress1)
	if newipAddressesCount > ipAddressCount {
		newlyAdded1 := make([]string, len(ipAddress1))
		copy(newlyAdded1, ipAddress1)
		newlyAdded := compare(newlyAdded1, ipAddress)
		fmt.Println(" newly added: ", newlyAdded)
		// for i := 0; i < newipAddressesCount; i++ {
		// 	if (ipAddress1[i] == "10.132.32.211") || (ipAddress1[i] == "10.132.32.243") {
		// 		fmt.Println("removing ", ipAddress1[i])
		// 		fmt.Println("the element index is ", i)
		// 		copy(ipAddress1[i:], ipAddress1[i+1:]) // Shift a[i+1:] left one index
		// 		ipAddress1[len(ipAddress1)-1] = ""     // Erase last element (write zero value)
		// 		ipAddress1 = ipAddress1[:len(ipAddress1)-1]
		// 		newipAddressesCount--
		// 	}
		// }
	}
	ipAddress = ipAddress1
	return ipAddress1
}

func compare(a, b []string) []string {
	for i := len(a) - 1; i >= 0; i-- {
		for _, vD := range b {
			if a[i] == vD {
				a = append(a[:i], a[i+1:]...)
				break
			}
		}
	}
	return a
}

func main() {
	router := mux.NewRouter()
	out, _ := exec.Command("/bin/sh", "-c", "echo tcs123| sudo -S nmap -sU 10.132.32.198-255 -p 161 --open -oG - | awk '/161\\/open.*/{print $2}'").Output()
	str1 := string(out)
	strs := strings.Split(str1, "\n")
	ipAddress = strs
	ipAddressCount = len(ipAddress)

	for index := range strs {
		devices = append(devices, Device{IP: strs[index]})
	}
	fmt.Println("Serving at http://localhost:12345/devices ...")
	fmt.Println(ipAddress)
	router.HandleFunc("/devices", GetdevicesEndpoint).Methods("GET")
	router.HandleFunc("/devices/{ip}", GetDeviceEndpoint).Methods("GET")
	router.HandleFunc("/heartbeat", GetHeartbeat).Methods("GET")

	log.Fatal(http.ListenAndServe(":12345", router))
}
