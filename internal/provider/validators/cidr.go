package validators

import (
	"context"
	"fmt"
	"net"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type cidrValidator struct {
	maxPrefix      int
	requirePrivate bool
}

// CIDR returns a validator that checks the value is a valid CIDR using net.ParseCIDR.
// It does not enforce any prefix length restriction.
func CIDR() validator.String {
	return &cidrValidator{}
}

// CIDRWithMaxPrefix returns a validator that checks the value is a valid CIDR
// and that the prefix length is at most maxPrefix (e.g. 21 means /21 or larger network).
func CIDRWithMaxPrefix(maxPrefix int) validator.String {
	return &cidrValidator{maxPrefix: maxPrefix}
}

// PrivateCIDRWithMaxPrefix returns a validator that checks the value is a valid CIDR,
// that the prefix length is at most maxPrefix, and that the IP is in a private
// RFC 1918 range (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16).
func PrivateCIDRWithMaxPrefix(maxPrefix int) validator.String {
	return &cidrValidator{maxPrefix: maxPrefix, requirePrivate: true}
}

var rfc1918Networks = func() []*net.IPNet {
	cidrs := []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}
	nets := make([]*net.IPNet, len(cidrs))
	for i, c := range cidrs {
		_, n, _ := net.ParseCIDR(c)
		nets[i] = n
	}
	return nets
}()

func (v *cidrValidator) Description(_ context.Context) string {
	switch {
	case v.requirePrivate && v.maxPrefix > 0:
		return fmt.Sprintf("must be a valid RFC 1918 private CIDR with at most /%d prefix", v.maxPrefix)
	case v.maxPrefix > 0:
		return fmt.Sprintf("must be a valid CIDR with at most /%d prefix", v.maxPrefix)
	default:
		return "must be a valid CIDR"
	}
}

func (v *cidrValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v *cidrValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()

	_, cidr, err := net.ParseCIDR(value)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid CIDR",
			fmt.Sprintf("%q is not a valid CIDR notation.", value),
		)
		return
	}

	if v.maxPrefix > 0 {
		prefix, _ := cidr.Mask.Size()
		if prefix > v.maxPrefix {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid CIDR",
				fmt.Sprintf("%q has a /%d prefix, at least /%d is required.", value, prefix, v.maxPrefix),
			)
			return
		}
	}

	if v.requirePrivate {
		ip := cidr.IP
		private := false
		for _, n := range rfc1918Networks {
			if n.Contains(ip) {
				private = true
				break
			}
		}
		if !private {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid CIDR",
				fmt.Sprintf("%q is not in a private RFC 1918 range (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16).", value),
			)
		}
	}
}
