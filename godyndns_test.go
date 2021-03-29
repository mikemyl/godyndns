package godyndns

import (
	"bytes"
	"io/ioutil"
	"net"
	"net/http"
	"reflect"
	"testing"
)

func TestGetPublicIp(t *testing.T) {
	type args struct {
		client *http.Client
	}
	tests := []struct {
		name string
		args args
		want net.IP
	}{
		{"Should parse the IP from the body", args{mockHTTPClient(200, "200 OK", "1.1.1.1")}, net.ParseIP("1.1.1.1")},
		{"Should return nil if a non valid IP is returned", args{mockHTTPClient(200, "200 OK", "invalid IP")}, nil},
		{"Should return nil if a non 200 response is returned", args{mockHTTPClient(400, "400 Bad request", "Bad request")}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := GetPublicIP(tt.args.client); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPublicIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetGodaddyARecordIp(t *testing.T) {
	type args struct {
		client     *http.Client
		domainName string
		apiKey     string
		secretKey  string
	}
	tests := []struct {
		name     string
		args     args
		want     net.IP
		hasError bool
	}{
		{"Should parse the ip from the JSON", args{mockHTTPClient(200, "200 OK", `[{"data":"1.1.1.1","name":"some.domain.com","ttl":600,"type":"A"}]`), "some.domain.com", "apiKey", "secretKey"}, net.ParseIP("1.1.1.1"), false},
		{"Should return nil if invalid json returned", args{mockHTTPClient(200, "200 OK", `[foo]`), "some.domain.com", "apiKey", "secretKey"}, nil, true},
		{"Should return nil if empty json", args{mockHTTPClient(200, "200 OK", `[]`), "some.domain.com", "apiKey", "secretKey"}, nil, true},
		{"Should return nil if invalid subdomain given", args{mockHTTPClient(200, "200 OK", `[{"data":"1.1.1.1","name":"some.domain.com","ttl":600,"type":"A"}]`), "invalid", "apiKey", "secretKey"}, nil, true},
		{"Should return nil if non valid ip is returned", args{mockHTTPClient(200, "200 OK", `[{"data":"invalid","name":"some.domain.com","ttl":600,"type":"A"}]`), "some.domain.com", "apiKey", "secretKey"}, nil, true},
		{"Should return nil non 200 status code is returned", args{mockHTTPClient(401, "401 Unauthorized", ``), "some.domain.com", "apiKey", "secretKey"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetGodaddyARecordIP(tt.args.client, tt.args.domainName, tt.args.apiKey, tt.args.secretKey)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetGodaddyARecordIP() = %v, want %v", got, tt.want)
			}
			if (tt.hasError && err == nil) || (!tt.hasError && err != nil) {
				t.Errorf("GetGodaddyARecordIP() err = %e , while wanted error : %v", err, tt.hasError)
			}
		})
	}
}

func TestUpdateGoDaddyARecord(t *testing.T) {
	type args struct {
		client     *http.Client
		domainName string
		publicIP   net.IP
		apiKey     string
		secretKey  string
	}
	tests := []struct {
		name     string
		args     args
		hasError bool
	}{
		{"Should return err if nil IP is given", args{mockHTTPClient(0, "ignored", `[]`), "some.domain.com", nil, "apiKey", "secretKey"}, true},
		{"Should return err if non 200 http status code", args{mockHTTPClient(404, "404 Bad request", `[]`), "some.domain.com", nil, "apiKey", "secretKey"}, true},
		{"Should return err if invalid subdomain given", args{mockHTTPClient(200, "200 OK", `[]`), "invalid", net.ParseIP("1.1.1.1"), "apiKey", "secretKey"}, true},
		{"Shouldn't return err if valid request", args{mockHTTPClient(200, "200 OK", `ignored`), "some.domain.com", net.ParseIP("1.1.1.1"), "apiKey", "secretKey"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := UpdateGoDaddyARecord(tt.args.client, tt.args.domainName, tt.args.publicIP, tt.args.apiKey, tt.args.secretKey)
			if tt.hasError && err == nil {
				t.Errorf("Expected UpdateGoDaddyARecord() to return an error")
			}
			if !tt.hasError && err != nil {
				t.Errorf("UpdateGoDaddyARecord() returned an error")
			}
		})
	}
}

func Test_addHeaders(t *testing.T) {
	req := &http.Request{Header: http.Header{}}

	got := addHeaders(req, "foo", "bar")

	assertEqual(t, "application/json", got.Header.Get("accept"))
	assertEqual(t, "application/json", got.Header.Get("Content-Type"))
	assertEqual(t, "application/json", got.Header.Get("Content-Encoding"))
	assertEqual(t, "sso-key foo:bar", got.Header.Get("Authorization"))
}

func Test_constructUrl_AddsSchemeAndCreatesUrl(t *testing.T) {
	tests := []struct {
		name            string
		domainInput     string
		wantedSubDomain string
		wantedDomain    string
		wantError       bool
	}{
		{"Simple domain", "foo.bar.com", "foo", "bar.com", false},
		{"Domain starting with https://", "https://foo.bar.com", "foo", "bar.com", false},
		{"co.uk doamin", "https://foo.bar.co.uk", "foo", "bar.co.uk", false},
		{"No subdomain", "nosubdomain.io", "@", "nosubdomain.io", false},
		{"@ subdomain", "@.nosubdomain.io", "@", "nosubdomain.io", false},
		{"No domain", "invalid-domain", "", "", true},
		{"Invalid domain", "|#!$%", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := constructURL(tt.domainInput)
			if err != nil {
				if !tt.wantError {
					t.Errorf("construnctUrl returned an error unexpectedly")
				}
			} else {
				if tt.wantError {
					t.Errorf("wanted an error but didn't get one")
				}
				assertEqual(t, tt.wantedDomain, url.Domain+"."+url.TLD)
				assertEqual(t, tt.wantedSubDomain, url.Subdomain)
			}
		})
	}
}

type HTTPTransportFunc func(req *http.Request) *http.Response


func (fn HTTPTransportFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request), nil
}

func mockHTTPClient(code int, status, body string) *http.Client {
	response := response(code, status, body)
	return &http.Client{
		Transport: HTTPTransportFunc(func(req *http.Request) *http.Response {
			return response
		}),
	}
}

func response(code int, status, body string) *http.Response {
	return &http.Response{
		Status:     status,
		StatusCode: code,
		// Send response to be tested
		Body: ioutil.NopCloser(bytes.NewBufferString(body)),
		// Must be set to non-nil value or it panics
		Header: make(http.Header),
	}
}

func assertEqual(t *testing.T, expected interface{}, actual interface{}) {
	if expected == actual {
		return
	}
	t.Errorf("Received %v (type %v), expected %v (type %v)", actual, reflect.TypeOf(actual), expected, reflect.TypeOf(expected))
}
