package main

//Imports
import (
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	config "github.com/spf13/viper"
)

//Runtime
func init() {
	config.AddConfigPath("$HOME/.config/dreamhost-dyndns")
	config.SetConfigName("config")
	config.ReadInConfig()
}

func main() {
	apiKey := config.GetString("APIKey")
	dnsName := config.GetString("DNSName")
	externalIP := getHTTP(config.GetString("IPLookupUri"))
	previousRecord := getPreviousRecord(apiKey, dnsName)
	log.Printf("Deduced external IP: %q", externalIP)
	log.Printf("Previous record: %+v", previousRecord)

	if externalIP == previousRecord[4] {
		log.Println("DNS Unchanged...")
	} else {
		log.Println("Updating DNS...")
		log.Printf("removal: %q", removeDNS(apiKey, previousRecord))
		log.Printf("add: %q", addDNS(apiKey, dnsName, externalIP))
	}
}

//Functions
func getHTTP(uri string) string {
	r, err := http.Get(uri)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer r.Body.Close()

	contents, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	return string(contents)
}

func newUUID() string {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return ""
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}

func getIndexInSlice(slice []string, text string) int {
	for p, v := range slice {
		if v == text {
			return p
		}
	}
	return -1
}

func getPreviousRecord(apiKey string, dnsEntry string) []string {
	hostname, _ := os.Hostname()

	uri := "https://api.dreamhost.com/?key=" + apiKey + "&unique_id=" + newUUID() + "&cmd=dns-list_records&ps=" + hostname
	records := strings.Fields(getHTTP(uri))
	index := getIndexInSlice(records, dnsEntry)

	if index == -1 {
		return []string{"0", "0", "0", "0", "0", "0"}
	}

	//'account_id', 'zone', 'record', 'type', 'value', 'comment', 'editable'
	return records[index-2 : index+6]
}

func removeDNS(apiKey string, previousRecord []string) []string {
	hostname, _ := os.Hostname()

	uri := "https://api.dreamhost.com/?key=" + apiKey + "&unique_id=" + newUUID() + "&cmd=dns-remove_record&ps=" + hostname + "&record=" + previousRecord[2] + "&type=" + previousRecord[3] + "&value=" + previousRecord[4]
	response := strings.Fields(getHTTP(uri))

	return response
}

func addDNS(apiKey string, dnsName string, externalIP string) []string {
	hostname, _ := os.Hostname()

	uri := "https://api.dreamhost.com/?key=" + apiKey + "&unique_id=" + newUUID() + "&cmd=dns-add_record&ps=" + hostname + "&record=" + dnsName + "&type=A&value=" + externalIP
	response := strings.Fields(getHTTP(uri))

	return response
}
