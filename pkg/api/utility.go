package api // "github.com/itszuvalex/mcdiscord/pkg/api"

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseNetLocation(loc string) (*NetLocation, error) {
	location := NetLocation{}
	ipport := strings.Split(loc, ":")
	if len(ipport) != 2 {
		return nil, fmt.Errorf("%s not a valid ip:port string", loc)
	}
	location.Address = ipport[0]
	port, err := strconv.Atoi(ipport[1])
	if err != nil {
		return nil, err
	}
	location.Port = port
	return &location, nil
}
