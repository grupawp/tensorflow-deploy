package plaintext

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/grupawp/tensorflow-deploy/app"
	"github.com/grupawp/tensorflow-deploy/exterr"
)

type Plaintext struct {
	hostsPath string
}

var (
	regexWhitespace = regexp.MustCompile(`[[:space:]]`)

	plainTextNotFoundInstanceAddressErrorCode = 1004
	plainTextInvalidInstanceAddressErrorCode  = 1005
	plainTextInvalidInstanceNameErrorCode     = 1006
	plainTextInstanceNameIsEmptyErrorCode     = 1007

	messageErrorInvalidInstanceAddress = "instance's address is invalid"

	errPlainTextNotFoundInstanceAddress = exterr.NewErrorWithMessage("not found instance's address").WithComponent(app.ComponentDiscovery).WithCode(plainTextNotFoundInstanceAddressErrorCode)
	errPlainTextInvalidInstanceName     = exterr.NewErrorWithMessage("instance name is invalid").WithComponent(app.ComponentDiscovery).WithCode(plainTextInvalidInstanceNameErrorCode)
	errPlainTextInstanceNameIsEmpty     = exterr.NewErrorWithMessage("instance name is empty").WithComponent(app.ComponentDiscovery).WithCode(plainTextInstanceNameIsEmptyErrorCode)
)

// NewPlaintext creates instance of discovery-Plaintext
func NewPlaintext(ctx context.Context, hostsPath string) (*Plaintext, error) {
	return &Plaintext{
		hostsPath: hostsPath,
	}, nil
}

// Discover returns serving instances
func (pt *Plaintext) Discover(ctx context.Context, servableID app.ServableID) ([]string, error) {
	lines, err := pt.readLines()
	if err != nil {
		return nil, exterr.WrapWithFrame(err)
	}
	var instances []string

	for lineNo, line := range lines {
		if pt.skipLine(line) {
			continue
		}

		skip, err := pt.skipInstanceName(line, servableID.InstanceName())
		if err != nil {
			return nil, exterr.WrapWithFrame(err)
		}
		if skip {
			continue
		}
		currentInstances, err := pt.readInstances(ctx, lineNo, line)
		if err != nil {
			return nil, exterr.WrapWithFrame(err)
		}
		if len(currentInstances) == 0 {
			continue
		}
		instances = append(instances, currentInstances...)
	}
	return instances, nil
}

func (pt *Plaintext) readInstances(ctx context.Context, lineNo int, line string) ([]string, error) {
	lineNo++
	line = strings.TrimLeft(line, "\t\n\f\r ")
	var instances []string
	_, err := pt.extractInstanceName(line)
	if err != nil {
		return nil, exterr.WrapWithFrame(err)
	}

	entries, err := pt.extractInstances(line)
	if err != nil {
		return nil, exterr.WrapWithFrame(err)
	}

	for _, v := range entries {
		if !pt.isValidInstance(v) {
			return nil, exterr.NewErrorWithMessage(fmt.Sprintf("%s line %d, address %s", messageErrorInvalidInstanceAddress, lineNo, v)).
				WithComponent(app.ComponentDiscovery).WithCode(plainTextInvalidInstanceAddressErrorCode)
		}
		instances = append(instances, v)
	}

	return instances, nil
}

func (pt *Plaintext) readLines() ([]string, error) {
	file, err := os.Open(pt.hostsPath)
	if err != nil {
		return nil, err
	}

	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, nil
}

func (pt *Plaintext) skipLine(line string) bool {
	if len(line) == 0 {
		return true
	}
	re := regexp.MustCompile(`^[[:space:]]*[#;]+`)
	return re.MatchString(line)
}

func (pt *Plaintext) skipInstanceName(line, instanceName string) (bool, error) {
	in, err := pt.extractInstanceName(line)
	if err != nil {
		return false, exterr.WrapWithFrame(err)
	}
	if in != instanceName {
		return true, nil
	}
	return false, nil
}

func (pt *Plaintext) extractInstanceName(line string) (string, error) {
	if len(line) == 0 {
		return "", errPlainTextInstanceNameIsEmpty
	}
	instances := regexWhitespace.Split(line, -1)
	if len(instances) == 0 {
		return "", errPlainTextInstanceNameIsEmpty
	}
	if !pt.isValidInstanceName(instances[0]) {
		return "", errPlainTextInvalidInstanceName
	}
	return instances[0], nil
}

func (pt *Plaintext) extractInstances(line string) ([]string, error) {
	splitInstances := regexWhitespace.Split(line, -1)
	if len(splitInstances) <= 1 {
		return nil, errPlainTextNotFoundInstanceAddress
	}
	var instances []string
	for k, v := range splitInstances {
		if k == 0 || len(v) == 0 {
			continue
		}
		instances = append(instances, v)
	}
	return instances, nil
}

func (pt *Plaintext) isValidInstanceName(instanceName string) bool {
	re := regexp.MustCompile(fmt.Sprintf("^tfs-[[:alnum:]]{1,%d}-[[:alnum:]]{1,%d}", app.MaxTeamLength, app.MaxProjectLength))
	return re.MatchString(instanceName)
}

func (pt *Plaintext) isValidInstance(entry string) bool {
	re := regexp.MustCompile(`:`)
	pieces := re.Split(entry, -1)
	if len(pieces) != 2 {
		return false
	}

	// ip
	re = regexp.MustCompile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)
	if !re.MatchString(pieces[0]) {
		return false
	}

	// port
	re = regexp.MustCompile(`^0{1,}`)
	if re.MatchString(pieces[1]) {
		return false
	}

	port, err := strconv.Atoi(pieces[1])
	if err != nil {
		return false
	}

	if 0 < port && port < 65536 {
		return true
	}
	return false
}
