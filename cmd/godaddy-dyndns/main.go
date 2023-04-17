package main

import (
	"fmt"
	"github.com/labstack/gommon/color"
	"github.com/mkideal/cli"
	"github.com/sestus/godyndns"
	"log"
	"net/http"
	"os"
)

type argT struct {
	cli.Helper
	ApiKey    string `cli:"api-key" usage:"GoDaddy Api Key" dft:"$GODADDY_API_KEY"`
	SecretKey string `cli:"secret-key" usage:"GoDaddy Secret Key" dft:"$GODADDY_SECRET_KEY"`
	Domain    string `cli:"domain" usage:"GoDaddy Domain to update. If the domain doesn't exist it creates it" dft:"$GODADDY_DOMAIN"`
}

func (argv *argT) Validate(ctx *cli.Context) error {
	if argv.ApiKey == "" {
		return fmt.Errorf("GoDaddy api key not provided. Please specify one using --api-key=<api key> or by setting the GODADDY_API_KEY env var")
	}
	if argv.SecretKey == "" {
		return fmt.Errorf("GoDaddy secret key not provided. Please specify one using --secret-key=<secret key> or by setting the GODADDY_SECRET_KEY env var")
	}
	if argv.Domain == "" {
		return fmt.Errorf("GoDaddy domain key not provided. Please specify one using --domain=<domain> or by setting the GODADDY_DOMAIN env var")
	}
	return nil
}

func main() {
	boldTitle := color.Bold("godaddy-dyndns")
	desc := boldTitle + "\n\n" + "godaddy-dyndns is a simple utility that is basically a DIY dynamic DNS. It checks the current public IP" +
		" against a godaddy domain and,  if they don't match, it updates the domain with the new IP address. "
	os.Exit(cli.Run(new(argT), func(ctx *cli.Context) error {
		argv := ctx.Argv().(*argT)
		client := &http.Client{}
		ip, err := godyndns.GetPublicIP(client)
		if err != nil {
			log.Fatalf("Failed to get my public IP address : %e. Exiting..", err)
		}
		domainIp, err := godyndns.GetGodaddyARecordIP(client, argv.Domain, argv.ApiKey, argv.SecretKey)
		if err != nil {
			log.Fatalf("Failed to get the GoDaddy A Record : %s.", err)
		}
		if ip.Equal(domainIp) {
			log.Printf("%s is already pointing to %s. Won't update..", argv.Domain, domainIp)
			os.Exit(0)
		}
		log.Printf("%s is pointing to %s. Will update it to point to %s", argv.Domain, domainIp, ip)
		err = godyndns.UpdateGoDaddyARecord(client, argv.Domain, ip, argv.ApiKey, argv.SecretKey)
		if err != nil {
			log.Fatalf("Failed to Udpate GoDaddy A record : %e", err)
		}
		return nil
	}, desc))

}
