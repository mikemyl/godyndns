GoDaddy DynDns
==============

[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/sestus/godyndns/master/LICENSE)
[![Build status](https://travis-ci.com/sestus/godyndns.svg?branch=master)](https://travis-ci.com/github/sestus/godyndns)


godaddy-dyndns is a simple utility that is basically a DIY dynamic DNS. It checks the current public IP against a
[GoDaddy](https://godaddy.com) subdomain and,  if they don't match, it updates the subdomain to point to the new IP address.

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