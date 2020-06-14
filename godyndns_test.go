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
		{"Should parse the IP from the body", args{mockHttpClient(200, "200 OK", "1.1.1.1")}, net.ParseIP("1.1.1.1")},
		{"Should return nil if a non valid IP is returned", args{mockHttpClient(200, "200 OK", "invalid IP")}, nil},
		{"Should return nil if a non 200 response is returned", args{mockHttpClient(400, "400 Bad request", "Bad request")}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := GetPublicIp(tt.args.client); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPublicIp() = %v, want %v", got, tt.want)
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
		name string
		args args
		want net.IP
		hasError bool
	}{
		{"Should parse the ip from the JSON", args{mockHttpClient(200, "200 OK", `[{"data":"1.1.1.1","name":"some.domain.com","ttl":600,"type":"A"}]`), "domainName", "apiKey", "secretKey"}, net.ParseIP("1.1.1.1"), false},
		{"Should return nil if non valid ip is returned", args{mockHttpClient(200, "200 OK", `[{"data":"invalid","name":"some.domain.com","ttl":600,"type":"A"}]`), "domainName", "apiKey", "secretKey"}, nil, true},
		{"Should return nil non 200 status code is returned", args{mockHttpClient(401, "401 Unauthorized", ``), "domainName", "apiKey", "secretKey"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetGodaddyARecordIp(tt.args.client, tt.args.domainName, tt.args.apiKey, tt.args.secretKey)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetGodaddyARecordIp() = %v, want %v", got, tt.want)
			}
			if (tt.hasError && err == nil) || (!tt.hasError && err != nil ) {
				t.Errorf("GetGodaddyARecordIp() err = %e , while wanted error : %v", err, tt.hasError)
			}
		})
	}
}

func TestUpdateGoDaddyARecord(t *testing.T) {
	type args struct {
		client     *http.Client
		domainName string
		publicIp   net.IP
		apiKey     string
		secretKey  string
	}
	tests := []struct {
		name string
		args args
		hasError bool
	}{
		{"Should return err if nil IP is given", args{mockHttpClient(0, "ignored", `[]`), "domainName", nil, "apiKey", "secretKey"}, true},
		{"Should return err if non 200 http status code", args{mockHttpClient(404, "404 Bad request", `[]`), "domainName", nil, "apiKey", "secretKey"}, true},
		{"Shouldn't return err if valid request", args{mockHttpClient(200, "200 OK", `ignored`), "domainName", net.ParseIP("1.1.1.1"), "apiKey", "secretKey"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := UpdateGoDaddyARecord(tt.args.client, tt.args.domainName, tt.args.publicIp, tt.args.apiKey, tt.args.secretKey)
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


type HttpTransportFunc func(req *http.Request) *http.Response

func (fn HttpTransportFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req), nil
}

func mockHttpClient(code int, status, body string) *http.Client {
	response := response(code, status, body)
	return &http.Client{
		Transport: HttpTransportFunc(func(req *http.Request) *http.Response {
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
	t.Errorf("Received %v (type %v), expected %v (type %v)", expected, reflect.TypeOf(expected), actual, reflect.TypeOf(actual))
}