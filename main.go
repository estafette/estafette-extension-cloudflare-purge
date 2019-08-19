package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/cloudflare/cloudflare-go"
)

var (
	app       string
	version   string
	branch    string
	revision  string
	buildDate string
	goVersion = runtime.Version()
)

var (
	// flags
	paramsJSON      = kingpin.Flag("params", "Extension parameters, created from custom properties.").Envar("ESTAFETTE_EXTENSION_CUSTOM_PROPERTIES").Required().String()
	credentialsJSON = kingpin.Flag("credentials", "Cloudflare credentials configured at service level, passed in to this trusted extension.").Envar("ESTAFETTE_CREDENTIALS_CLOUDFLARE").Required().String()
)

func main() {

	// parse command line parameters
	kingpin.Parse()

	// log to stdout and hide timestamp
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	// log startup message
	logInfo("Starting %v version %v...", app, version)

	logInfo("Unmarshalling injected credentials...")
	var credentials []CloudflareCredentials
	err := json.Unmarshal([]byte(*credentialsJSON), &credentials)
	if err != nil {
		log.Fatal("Failed unmarshalling injected credentials: ", err)
	}

	if len(credentials) == 0 {
		log.Fatalf("No credentials of type cloudflare have been passed in.")
	}
	credential := credentials[0]

	var params Params

	logInfo("Unmarshalling parameters / custom properties...")
	err = json.Unmarshal([]byte(*paramsJSON), &params)
	if err != nil {
		log.Fatal("Failed unmarshalling parameters: ", err)
	}

	logInfo("Validating required parameters...")
	valid, errors := params.ValidateRequiredProperties()
	if !valid {
		log.Fatal("Not all valid fields are set: ", errors)
	}

	cloudflareAPI, err := cloudflare.New(credential.AdditionalProperties.APIKey, credential.AdditionalProperties.APIEmail)
	if err != nil {
		log.Fatal("Failed creating Cloudflare client: ", err)
	}

	for _, host := range params.Hosts {

		// get tld from host
		hostParts := strings.Split(host, ".")
		n := 2

		// get zone
		var id string
		for {
			domainParts := hostParts[len(hostParts)-n:]
			domain := strings.Join(domainParts, ".")

			id, err = cloudflareAPI.ZoneIDByName(domain)
			if err == nil || n == len(hostParts) {
				break
			}

			// take an extra part of the hostname to check if a zone exists
			n++
		}
		if err != nil {
			log.Fatalf("Can't find zone for host %v", host)
		}

		logInfo("Purging Cloudflare cache for host %v", host)
		response, err := cloudflareAPI.PurgeCache(id, cloudflare.PurgeCacheRequest{
			Hosts: []string{host},
		})
		if err != nil {
			log.Fatalf("Failed purging cache for host %v: %v", host, err)
		}
		if !response.Success {
			log.Fatalf("Failed purging cache for host %v: %v", host, response.Errors)
		}
		logInfo("Succesfully purged Cloudflare cache for host %v", host)
	}

	logInfo("Succesfully purged Cloudflare cache for all hosts")
}

func logInfo(message string, args ...interface{}) {
	formattedMessage := fmt.Sprintf(message, args...)
	log.Printf("%v\n\n", formattedMessage)
}
