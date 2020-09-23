package main

import (
	"context"

	"github.com/grupawp/tensorflow-deploy/app"
	"github.com/grupawp/tensorflow-deploy/discovery"
	"github.com/grupawp/tensorflow-deploy/discovery/dns"
	"github.com/grupawp/tensorflow-deploy/discovery/plaintext"
	"github.com/grupawp/tensorflow-deploy/serving"
)

// NewServiceDiscovery creates instance of ServiceDiscovery depends on serviceID
func NewServiceDiscovery(packageID string, conf app.ConfigDiscovery) (serving.Discoverer, error) {
	pacID, err := discovery.Package(packageID)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	switch pacID {
	case discovery.PlaintextPackage:
		return plaintext.NewPlaintext(ctx, *conf.Plaintext.HostsPath)
	case discovery.DNSPackage:
		return dns.NewDNS(ctx, *conf.DNS.ServiceSuffix, *conf.DNS.DefaultInstancePort)
	}

	return nil, discovery.ErrUnknownPackageDiscovery
}
