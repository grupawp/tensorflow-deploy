package dns

import (
	"context"
	"net"
	"strconv"

	"github.com/grupawp/tensorflow-deploy/app"
	"github.com/grupawp/tensorflow-deploy/exterr"
	"github.com/grupawp/tensorflow-deploy/logging"
)

const (
	srvProto = "tcp"
)

var (
	dnsLookupSRVNotFoundInstancesErrorCode = 1002
	dnsHostNotFoundInstancesErrorCode      = 1003

	errDNSLookupSRVNotFoundInstances = exterr.NewErrorWithMessage("instance not found").WithComponent(app.ComponentDiscovery).WithCode(dnsLookupSRVNotFoundInstancesErrorCode)
	errDNSHostNotFoundInstances      = exterr.NewErrorWithMessage("instance not found").WithComponent(app.ComponentDiscovery).WithCode(dnsHostNotFoundInstancesErrorCode)
)

// DNS
type DNS struct {
	serviceSuffix       string
	defaultInstancePort uint16
	resolver            *net.Resolver
}

// NewDNS
func NewDNS(ctx context.Context, serviceSuffix string, defaultInstancePort uint16) (*DNS, error) {
	return &DNS{
		serviceSuffix:       serviceSuffix,
		defaultInstancePort: defaultInstancePort,
		resolver:            net.DefaultResolver,
	}, nil
}

// Discover
func (d DNS) Discover(ctx context.Context, servableID app.ServableID) ([]string, error) {
	// SRV
	instances, err := d.lookupSRV(ctx, servableID.InstanceName())
	if err == nil {
		return instances, nil
	}
	logging.Debug(ctx, err.Error())

	// Host
	instances, err = d.lookupHost(ctx, servableID.InstanceHost(d.serviceSuffix), d.defaultInstancePort)
	if err == nil {
		return instances, nil
	}
	logging.Debug(ctx, err.Error())

	return nil, errDNSHostNotFoundInstances
}

func (d DNS) lookupSRV(ctx context.Context, service string) ([]string, error) {
	instances := make([]string, 0)

	_, addresses, err := d.resolver.LookupSRV(ctx, service, srvProto, d.serviceSuffix)
	if err != nil {
		return nil, exterr.WrapWithFrame(err)
	}
	if len(addresses) == 0 {
		return nil, errDNSLookupSRVNotFoundInstances
	}

	// GRPC BUG - grpclb problem TODO: description do konsultacji z Tomkiem
	for _, a := range addresses {
		addrInstances, err := d.lookupHost(ctx, a.Target, a.Port)
		if err != nil {
			return nil, exterr.WrapWithFrame(err)
		}

		instances = append(instances, addrInstances...)
	}

	return instances, nil
}

func (d DNS) lookupHost(ctx context.Context, host string, port uint16) ([]string, error) {
	instances := make([]string, 0)

	addrs, err := d.resolver.LookupHost(ctx, host)
	if err != nil {
		return nil, exterr.WrapWithFrame(err)
	}
	if len(addrs) == 0 {
		return nil, errDNSHostNotFoundInstances
	}

	for _, a := range addrs {
		instances = append(instances, net.JoinHostPort(a, strconv.FormatUint(uint64(port), 10)))
	}

	return instances, nil
}
