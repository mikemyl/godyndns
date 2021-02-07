GoDaddy DynDns
==============

[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/sestus/godyndns/master/LICENSE)
[![Build and Test](https://github.com/sestus/godyndns/workflows/Build%20and%20Test/badge.svg)](https://github.com/sestus/godyndns/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/sestus/godyndns)](https://goreportcard.com/report/github.com/sestus/godyndns)


godaddy-dyndns is a simple utility that is basically a DIY dynamic DNS. It checks the current public IP against a
[GoDaddy](https://godaddy.com) subdomain and,  if they don't match, it updates the subdomain to point to the new IP address.

Installation
------------

Grab the latest release binaries for your OS-architecture from the latest [release](https://github.com/sestus/godyndns/releases).


Usage
-----

```
$ ./godaddy-dyndns -h
godaddy-dyndns

Options:

  -h, --help                               display help information
      --api-key[=$GODADDY_API_KEY]         GoDaddy Api Key
      --secret-key[=$GODADDY_SECRET_KEY]   GoDaddy Secret Key
      --subdomain[=$GODADDY_SUBDOMAIN]     GoDaddy SubDomain to update. If the subdomain doesn't exist it creates it

```

Example
-------
```
./godaddy-dyndns --api-key=my_godaddy_api_key --secret-key=my_godaddy_secret_key --subdomain=mysubdomain.mikemylonakis.com
```