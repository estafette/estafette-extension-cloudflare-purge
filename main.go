package main

import (
	"encoding/json"
	"runtime"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/cloudflare/cloudflare-go"
	foundation "github.com/estafette/estafette-foundation"
	"github.com/rs/zerolog/log"
)

var (
	appgroup  string
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

	// init log format from envvar ESTAFETTE_LOG_FORMAT
	foundation.InitLoggingFromEnv(appgroup, app, version, branch, revision, buildDate)

	log.Info().Msg("Unmarshalling injected credentials...")
	var credentials []CloudflareCredentials
	err := json.Unmarshal([]byte(*credentialsJSON), &credentials)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed unmarshalling injected credentials")
	}

	if len(credentials) == 0 {
		log.Fatal().Msg("No credentials of type cloudflare have been passed in.")
	}
	credential := credentials[0]

	var params Params

	log.Info().Msg("Unmarshalling parameters / custom properties...")
	err = json.Unmarshal([]byte(*paramsJSON), &params)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed unmarshalling parameters")
	}

	log.Info().Msg("Validating required parameters...")
	valid, errors := params.ValidateRequiredProperties()
	if !valid {
		log.Fatal().Msgf("Not all valid fields are set: %v", errors)
	}

	cloudflareAPI, err := cloudflare.New(credential.AdditionalProperties.APIKey, credential.AdditionalProperties.APIEmail)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed creating Cloudflare client")
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
			log.Fatal().Err(err).Msgf("Can't find zone for host %v", host)
		}

		log.Info().Msgf("Purging Cloudflare cache for host %v", host)
		response, err := cloudflareAPI.PurgeCache(id, cloudflare.PurgeCacheRequest{
			Hosts: []string{host},
		})
		if err != nil {
			log.Fatal().Err(err).Msgf("Failed purging cache for host %v", host)
		}
		if !response.Success {
			log.Fatal().Msgf("Failed purging cache for host %v: %v", host, response.Errors)
		}
		log.Info().Msgf("Succesfully purged Cloudflare cache for host %v", host)
	}

	log.Info().Msg("Succesfully purged Cloudflare cache for all hosts")
}
