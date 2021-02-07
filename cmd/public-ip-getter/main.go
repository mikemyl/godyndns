package main

import (
	"github.com/sestus/godyndns"
	"net/http"
)

func main() {
	godyndns.GetPublicIP(&http.Client{})
}
