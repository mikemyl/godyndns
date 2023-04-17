package godyndns

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jpillora/go-tld"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
)

type domain struct {
	Data string `json:"data"`
	Name string `json:"name"`
	TTL  int    `json:"ttl"`
	Type string `json:"type"`
}

type domainUpdate struct {
	Data string `json:"data"`
	TTL  int    `json:"ttl"`
}

type domainUpdates []domainUpdate

const domainsPath = "https://api.godaddy.com/v1/domains"

// GetPublicIP Gets the public ip of the current host, assuming that it can reach the internet
// it accepts an *http.Client as a param mainly for testing purposes
func GetPublicIP(client *http.Client) (net.IP, error) {
	var ipResolvers = [3]string{"http://ipinfo.io/ip", "https://api.ipify.org?format=text", "https://checkip.amazonaws.com/api"}
	for _, url := range ipResolvers {
		ip, err := getPublicIPFrom(client, url)
		if err == nil {
			log.Printf("My public IP is:%s\n", ip)
			return ip, nil
		}
	}
	return nil, fmt.Errorf("couldn't get my public IP. Tried %v", ipResolvers)

}

// UpdateGoDaddyARecord updates the A record of a given GoDaddy domain if the public IP that it points to
// is different compared to the publicIP parameter. The domainName param needs to look like : subdomain.domain.com.
// The function will then make a REST api call on the domain.com and will update the domain with the publicIP
func UpdateGoDaddyARecord(client *http.Client, domainName string, publicIP net.IP, apiKey, secretKey string) error {
	if publicIP == nil {
		log.Println("Given publicIP is nll")
		return errors.New("given publicIP is nll")
	}
	domainURL, err := constructURL(domainName)
	if err != nil {
		log.Printf("Failed to update the A record as I couldn't extract the domain from %s\n", domainName)
		return err
	}

	url := fmt.Sprintf("%s/%s.%s/records/A/%s", domainsPath, domainURL.Domain, domainURL.TLD, domainURL.Subdomain)
	record, _ := json.Marshal(domainUpdates{domainUpdate{publicIP.String(), 600}})
	req, _ := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(record))
	addHeaders(req, apiKey, secretKey)
	_, err = doRequest(client, req)
	return err
}

// GetGodaddyARecordIP gets the A record associated with the domainName.  The domainName param needs to look like :
// subdomain.domain.com . Upon successful retrieval, it returns the IP address associated with that domain
func GetGodaddyARecordIP(client *http.Client, domainName string, apiKey, secretKey string) (net.IP, error) {
	domainURL, err := constructURL(domainName)
	if err != nil {
		log.Printf("Failed to get A record as I couldn't extract the domain from %s\n", domainName)
		return nil, err
	}
	targetURL := fmt.Sprintf("%s/%s.%s/records/A/%s", domainsPath, domainURL.Domain, domainURL.TLD, domainURL.Subdomain)
	req, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		log.Printf("Failed to get the record details for domain %s : %s", domainName, err)
		return nil, err
	}
	addHeaders(req, apiKey, secretKey)
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to execute request %s : %s", req.URL, err)
		return nil, err
	}
	if resp.StatusCode != 200 {
		log.Printf("%s to %s returned %s.\n", req.Method, req.URL, resp.Status)
		return nil, errors.New(resp.Status)
	}
	defer resp.Body.Close()
	var record []domain
	err = json.NewDecoder(resp.Body).Decode(&record)
	if err != nil {
		log.Printf("Failed to decode the response body. %s", err)
		return nil, err
	}
	if len(record) == 0 {
		log.Printf("Couldn't get info on the domain : %s. Do we own that domain?", domainName)
		return nil, errors.New("invalid domain")
	}
	ip := net.ParseIP(record[0].Data)
	if ip == nil {
		return ip, fmt.Errorf("couldn't parse %s to an IP address", record[0].Data)
	}
	return net.ParseIP(record[0].Data), nil
}

func constructURL(domain string) (*tld.URL, error) {
	u, err := tld.Parse(domain)
	if err != nil {
		log.Printf("Couldn't construct domain from %s : %s", domain, err)
		return nil, err
	}
	if !u.ICANN {
		u, err = tld.Parse("https://" + domain)
		if err != nil {
			if strings.Contains(err.Error(), "empty label in domain") && domain[0] == '@' {
				u, err = tld.Parse("https://" + domain[2:])
			}
			if err != nil {
				log.Printf("Couldn't construct domain from %s : %s", domain, err)
				return nil, err
			}
		}
	}
	if len(u.Subdomain) == 0 {
		u.Subdomain = "@"
	}
	return u, nil
}

func addHeaders(r *http.Request, apiKey, secretKey string) *http.Request {
	r.Header.Set("Authorization", fmt.Sprintf("sso-key %s:%s", apiKey, secretKey))
	r.Header.Set("accept", "application/json")
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Content-Encoding", "application/json")
	return r
}

func doRequest(client *http.Client, r *http.Request) (string, error) {
	resp, err := client.Do(r)
	if err != nil {
		log.Fatalf("Failed to execute request %s : %s", r.URL, err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to parse request response : %s", err)
		return "", err
	}
	if resp.StatusCode != 200 {
		log.Printf("%s to %s returned %s.", r.Method, r.URL, resp.Status)
		return "", errors.New(resp.Status)
	}
	return string(body), nil
}

func getPublicIPFrom(client *http.Client, url string) (net.IP, error) {
	log.Printf("Getting my public IP address from  %s ...\n", url)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatalf("Failed to reach %s to get my public IP address: %s", url, err)
	}
	ret, err := doRequest(client, req)
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(strings.TrimSuffix(ret, "\n"))
	if ip == nil {
		return nil, fmt.Errorf("couldn't parse %s to an IP address", strings.TrimSuffix(ret, "\n"))
	}
	return ip, nil
}
