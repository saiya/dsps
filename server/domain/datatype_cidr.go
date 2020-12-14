package domain

import (
	"encoding/json"
	"fmt"
	"net"

	"golang.org/x/xerrors"
)

// CIDR is IPNet wrapper struct
type CIDR struct {
	str string
	net *net.IPNet
}

// PrivateCIDRs is a list of private networks, link local, and loopback addresses.
var PrivateCIDRs []CIDR

func init() {
	for _, str := range []string{
		"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", // RFC 1918
		"fc00::/7",             // RFC 4193
		"127.0.0.0/8",          // RFC 1122
		"169.254.0.0/16",       // RFC 3927
		"::1/128", "fe80::/10", // RFC 4291
	} {
		cidr, err := ParseCIDR(str)
		if err != nil {
			panic(err)
		}
		PrivateCIDRs = append(PrivateCIDRs, cidr)
	}
}

// ParseCIDR parse given CIDR notation string
func ParseCIDR(str string) (CIDR, error) {
	_, net, err := net.ParseCIDR(str)
	if err != nil {
		return CIDR{}, fmt.Errorf("invalid CIDR notation: %w", err)
	}
	return CIDR{str: str, net: net}, nil
}

func (cidr CIDR) String() string {
	return cidr.str
}

// IPNet returns net.IPNet of this range.
func (cidr CIDR) IPNet() *net.IPNet {
	return cidr.net
}

// Contains returns true if given IP is valid and contained in this.
func (cidr CIDR) Contains(ip string) bool {
	p := net.ParseIP(ip)
	if p == nil {
		return false
	}
	return cidr.net.Contains(p)
}

// MarshalJSON method for configuration marshal/unmarshal
func (cidr CIDR) MarshalJSON() ([]byte, error) {
	return json.Marshal(cidr.String())
}

// UnmarshalJSON method for configuration marshal/unmarshal
func (cidr *CIDR) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case string:
		parsed, err := ParseCIDR(value)
		if err != nil {
			return err
		}
		*cidr = parsed
		return nil
	default:
		return xerrors.New("invalid CIDR notation")
	}
}
