package discovery

import (
	"github.com/grupawp/tensorflow-deploy/app"
	"github.com/grupawp/tensorflow-deploy/exterr"
)

const (
	UnknownPackage   = "unknown"
	PlaintextPackage = "plaintext"
	DNSPackage       = "dns"
)

var (
	unknownPackageDiscoveryErrorCode = 1001
	ErrUnknownPackageDiscovery       = exterr.NewErrorWithMessage("package discovery is unknown").WithComponent(app.ComponentDiscovery).WithCode(unknownPackageDiscoveryErrorCode)
)

var packages = []string{
	DNSPackage,
	PlaintextPackage,
}

// Package returns package for given string
func Package(packageID string) (string, error) {
	for _, v := range packages {
		if v == packageID {
			return v, nil
		}
	}
	return UnknownPackage, ErrUnknownPackageDiscovery
}

// ListDiscoveries returns list of Services Discovery
func ListDiscoveries() []string {
	return packages
}
