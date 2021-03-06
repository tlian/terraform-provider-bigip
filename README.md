# Overview

A [Terraform](terraform.io) provider for F5 BigIP. Resources are currently available for LTM.

[![Build Status](https://travis-ci.org/f5devcentral/terraform-provider-bigip.svg?branch=master)](https://travis-ci.org/f5devcentral/terraform-provider-bigip)
[![Go Report Card](https://goreportcard.com/badge/github.com/f5devcentral/terraform-provider-bigip)](https://goreportcard.com/report/github.com/f5devcentral/terraform-provider-bigip)
[![license](https://img.shields.io/badge/license-Mozilla-red.svg?style=flat)](https://github.com/f5devcentral/terraform-provider-bigip/blob/master/LICENSE)
[![Join the chat at https://gitter.im/f5devcentral/terraform-provider-bigip](https://badges.gitter.im/f5devcentral/terraform-provider-bigip.svg)](https://gitter.im/f5devcentral/terraform-provider-bigip?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

# F5 Requirements

This provider uses the iControlREST API. Make sure that is installed and enabled on your F5 before proceeding. All the resources are validated with BIGIP-12.1.1.0.0.184.iso

# Provider Configuration

### Example
```
provider "bigip" {
  address = "${var.url}"
  username = "${var.username}"
  password = "${var.password}"
}
```

### Reference

`address` - (Required) Address of the device

`username` - (Required) Username for authentication

`password` - (Required) Password for authentication

`token_auth` - (Optional, Default=false) Enable to use a non-administrator user via TMOS or an external authentication source (LDAP, TACACS, etc)

`login_ref` - (Optional, Default="tmos") Login reference for token authentication (see BIG-IP REST docs for details)

# Resources

For resources should be named with their "full path". The full path is the combination of the partition + name of the resource.
For example `/Common/my-pool`.

## bigip_ltm_monitor

Configures a custom monitor for use by health checks.

### Example
```
resource "bigip_ltm_monitor" "monitor" {
  name = "/Common/terraform_monitor"
  parent = "/Common/http"
  send = "GET /some/path\r\n"
  timeout = "999"
  interval = "999"
  destination = "1.2.3.4:1234"
}
```

### Reference

`name` - (Required) Name of the monitor

`parent` - (Required) Existing LTM monitor to inherit from

`interval` - (Optional) Check interval in seconds

`timeout` - (Optional) Timeout in seconds

`send` - (Optional) Request string to send

`receive` - (Optional) Expected response string

`receive_disable` - (Optional)

`reverse` - (Optional)

`transparent` - (Optional)

`manual_resume` - (Optional)

`ip_dscp` - (Optional)

`time_until_up` - (Optional)

`destination` - (Optional) Specify an alias address for monitoring

## bigip_ltm_node

Manages a node configuration

### Example

```
resource "bigip_ltm_node" "node" {
  name = "/Common/terraform_node1"
  address = "10.10.10.10"
}
```

### Reference

`name` - (Required) Name of the node

`address` - (Required) IP or hostname of the node

`state` - (Optional) Default is "user-up" you can set to "user-down" if you want to disable

## bigip_ltm_pool

### Example

```
resource "bigip_ltm_pool" "pool" {
  name = "/Common/terraform-pool"
  load_balancing_mode = "round-robin"
  nodes = ["${bigip_ltm_node.node.name}:80"]
  monitors = ["${bigip_ltm_monitor.monitor.name}","${bigip_ltm_monitor.monitor2.name}"]
  allow_snat = "no"
  allow_nat = "no"
}
```

### Reference

`name` - (Required) Name of the pool

`nodes` - (Optional) Nodes to add to the pool. Format node_name:port. e.g. `node01:443`

`monitors` - (Optional) List of monitor names to associate with the pool

`allow_nat` - (Optional)

`allow_snat` - (Optional)

`load_balancing_mode` - (Optional, Default = round-robin)

`slow_ramp_time` - (Optional, Default = 10)

`service_down_action` - (Optional, Default = none)

`reselect_tries` - (Optional, Default = 0)

## bigip_ltm_virtual_server

Configures a Virtual Server

### Example

```
resource "bigip_ltm_virtual_server" "http" {
  name = "/Common/terraform_vs_http"
  destination = "10.12.12.12"
  port = 80
  pool = "/Common/the-default-pool"
}

# A Virtual server with SSL enabled
resource "bigip_ltm_virtual_server" "https" {
  name = "/Common/terraform_vs_https"
  destination = "${var.vip_ip}"
  port = 443
  pool = "${var.pool}"
  profiles = ["/Common/tcp","/Common/my-awesome-ssl-cert","/Common/http"]
  source_address_translation = "automap"
}

# A Virtual server with separate client and server profiles
resource "bigip_ltm_virtual_server" "https" {
  name = "/Common/terraform_vs_https"
  destination = "${var.vip_ip}"
  port = 443
  pool = "${var.pool}"
  client_profiles = ["/Common/tcp"]
  server_profiles = ["/Common/tcp-lan-optimized"]
  source_address_translation = "automap"
}
```

### Reference

`name` - (Required) Name of the virtual server

`port` - (Required) Listen port for the virtual server

`source` - (Optional) Source IP and mask

`destination` - (Required) Destination IP

`pool` - (Optional) Default pool name

`mask` - (Optional) Mask can either be in CIDR notation or decimal, i.e.: `24` or `255.255.255.0`. A CIDR mask of `0` is the same as `0.0.0.0`

`profiles` - (Optional) List of profiles associated both client and server contexts on the virtual server. This includes protocol, ssl, http, etc.

`client_profiles` - (Optional) List of client context profiles associated on the virtual server. Not mutually exclusive with `profiles` and `server_profiles`

`server_profiles` - (Optional) List of server context profiles associated on the virtual server. Not mutually exclusive with `profiles` and `client_profiles`

`irules` - (Optional) List of irules associated on the virtual server

`source_address_translation` - (Optional) Can be either omitted for `none` or the values `automap` or `snat`

`snatpool` - (Optional) Name of the snatpool to use. Requires source_address_translation to be set to 'snat'

`ip_protocol` - (Optional) Specify the IP protocol to use with the the virtual server (all, tcp, or udp are valid)

`policies` - (Optional) List of policies associated on the virtual server

`vlans` - (Optional) List of VLANs associated on the virtual server

## bigip_ltm_irule

Creates iRules

### Example

```
# Loading from a file is the preferred method
resource "bigip_ltm_irule" "rule" {
  name = "/Common/terraform_irule"
  irule = "${file("myirule.tcl")}"
}

resource "bigip_ltm_irule" "rule2" {
  name = "/Common/terraform_irule2"
  irule = <<EOF
when CLIENT_ACCEPTED {
     log local0. "test"
   }
EOF
}
```

### Reference

`name` - (Required) Name of the iRule

`irule` - (Required) Body of the iRule


## bigip_ltm_virtual_address

Configures a Virtual Address. NOTE: create/delete are not implemented
since the virtual addresses should be created/deleted automatically
with the corresponding virtual server.

### Example

```
resource "bigip_ltm_virtual_address" "vs_va" {

    name = "/Common/${bigip_ltm_virtual_server.vs.destination}"
    advertize_route = true
}
```

### Reference

`name` - (Required) Name of the virtual address

`description` - (Optional) Description of the virtual address

`advertize_route` - (Optional) Enabled dynamic routing of the address

`conn_limit` - (Optional, Default=0) Max number of connections for virtual address

`enabled` - (Optional, Default=true) Enable or disable the virtual address

`arp` - (Optional, Default=true) Enable or disable ARP for the virtual address

`auto_delete` - (Optional, Default=true) Automatically delete the virtual address with the virtual server

`icmp_echo` - (Optional, Default=true) Enable/Disable ICMP response to the virtual address

`traffic_group` - (Optional, Default=/Common/traffic-group-1) Specify the partition and traffic group

## bigip_ltm_policy

Configure [local traffic policies](https://support.f5.com/kb/en-us/solutions/public/15000/000/sol15085.html).
This is a fairly low level resource that does little to make actually using policies any simpler. A solid
understanding of how policies and their associated rules, actions and conditions
are managed through iControlREST is recommended.

### Example

```
resource "bigip_ltm_policy" "policy" {
  name = "/Common/my_policy"
  strategy = "/Common/first-match"
  requires = ["http"]
  controls = ["forwarding"]
  rule {
    name = "/Common/rule1"

    condition {
      httpUri = true
      startsWith = true
      values = ["/foo"]
    }

    condition {
      httpMethod = true
      values = ["GET"]
    }

    action {
      forward = true
      pool = "/Common/my_pool"
    }
  }
}
```

### Reference

`name` - (Required) Name of the policy

`strategy` - (Required) Strategy selection when more than one rule matches.

`requires` - (Required) Defines the types of conditions that you can use when configuring a rule.

`controls` - (Required) Defines the types of actions that you can use when configuring a rule.

`rule` - defines a single rule to add to the policy. Multiple rules can be defined for a single policy.

**Rules**

 Actions and Conditions support all fields available via the iControlREST API. You can see all of the
 available fields in the [iControlREST API documentation](https://devcentral.f5.com/d/icontrol-rest-api-reference-version-120).
 Each field in the actions and conditions objects is available. Pro tip: Create your policy via the GUI first then use
 the REST API to figure out how to configure the terraform resource.

 `name` (Required) - Name of the rule

 `action` - Defines a single action. Multiple actions can exist per rule.

 `condition` - Defines a single condition. Multiple conditions can exist per rule.

## bigip_ltm_persistence_profile_cookie

Configures a cookie persistence profile

### Example

```
resource "bigip_ltm_persistence_profile_cookie" "test_ppcookie" {
    name = "/Common/terraform_cookie"
    defaults_from = "/Common/cookie"
    match_across_pools = "enabled"
    match_across_services = "enabled"
    match_across_virtuals = "enabled"
    timeout = 3600
    override_conn_limit = "enabled"
    always_send = "enabled"
    cookie_encryption = "required"
    cookie_encryption_passphrase = "iam"
    cookie_name = "ham"
    expiration = "1:0:0"
    hash_length = 0

    lifecycle {
        ignore_changes = [ "cookie_encryption_passphrase" ]
    }
}
 

```

### Reference

`name` - (Required) Name of the virtual address

`defaults_from` - (Required) Parent cookie persistence profile

`match_across_pools` (Optional) (enabled or disabled) match across pools with given persistence record

`match_across_services` (Optional) (enabled or disabled) match across services with given persistence record

`match_across_virtuals` (Optional) (enabled or disabled) match across virtual servers with given persistence record

`mirror` (Optional) (enabled or disabled) mirror persistence record

`timeout` (Optional) (enabled or disabled) Timeout for persistence of the session in seconds

`override_conn_limit` (Optional) (enabled or disabled) Enable or dissable pool member connection limits are overridden for persisted clients. Per-virtual connection limits remain hard limits and are not overridden.

`always_send` (Optional) (enabled or disabled) always send cookies

`cookie_encryption` (Optional) (required, preferred, or disabled) To required, preferred, or disabled policy for cookie encryption

`cookie_encryption_passphrase` (Optional) (required, preferred, or disabled) Passphrase for encrypted cookies. The field is encrypted on the server and will always return differently then set.
If this is configured specify `ignore_changes` under the `lifecycle` block to ignore returned encrypted value.

`cookie_name` (Optional) Name of the cookie to track persistence

`expiration` (Optional) Expiration TTL for cookie specified in DAY:HOUR:MIN:SECONDS (Examples: 1:0:0:0 one day, 1:0:0 one hour, 30:0 thirty minutes) 

`hash_length` (Optional) (Integer) Length of hash to apply to cookie

`hash_offset` (Optional) (Integer) Number of characters to skip in the cookie for the hash

`httponly` (Optional) (enabled or disabled) Sending only over http

## bigip_ltm_persistence_profile_dstaddr

Configures a destination address persistence profile

### Example

```
resource "bigip_ltm_persistence_profile_dstaddr" "dstaddr" {
  name = "/Common/terraform_dstaddr"
  defaults_from = "/Common/dest_addr"
  match_across_pools = "enabled"
  match_across_services = "enabled"
  match_across_virtuals = "enabled"
  mirror = "enabled"
  timeout = 3600
  override_conn_limit = "enabled"
  mask = "255.255.255.0"
}
```

### Reference

`name` - (Required) Name of the virtual address

`defaults_from` - (Required) Parent cookie persistence profile

`match_across_pools` (Optional) (enabled or disabled) match across pools with given persistence record

`match_across_services` (Optional) (enabled or disabled) match across services with given persistence record

`match_across_virtuals` (Optional) (enabled or disabled) match across virtual servers with given persistence record

`mirror` (Optional) (enabled or disabled) mirror persistence record

`timeout` (Optional) (enabled or disabled) Timeout for persistence of the session in seconds

`override_conn_limit` (Optional) (enabled or disabled) Enable or dissable pool member connection limits are overridden for persisted clients. Per-virtual connection limits remain hard limits and are not overridden.

`mask` (Optional) Identify a range of source IP addresses to manage together as a single source address affinity persistent connection when connecting to the pool. Must be a valid IPv4 or IPv6 mask.

`hash_algorithm` (Optional) Specify the hash algorithm

## bigip_ltm_persistence_profile_srcaddr

Configures a source address persistence profile

### Example

```
resource "bigip_ltm_persistence_profile_srcaddr" "srcaddr" {
    name = "/Common/terraform_srcaddr"
    defaults_from = "/Common/source_addr"
    match_across_pools = "enabled"
    match_across_services = "enabled"
    match_across_virtuals = "enabled"
    mirror = "enabled"
    timeout = 3600
    override_conn_limit = "enabled"
    hash_algorithm = "carp"
    map_proxies = "enabled"
    mask = "255.255.255.255"
}
```

### Reference

`name` - (Required) Name of the virtual address

`defaults_from` - (Required) Parent cookie persistence profile

`match_across_pools` (Optional) (enabled or disabled) match across pools with given persistence record

`match_across_services` (Optional) (enabled or disabled) match across services with given persistence record

`match_across_virtuals` (Optional) (enabled or disabled) match across virtual servers with given persistence record

`mirror` (Optional) (enabled or disabled) mirror persistence record

`timeout` (Optional) (enabled or disabled) Timeout for persistence of the session in seconds

`override_conn_limit` (Optional) (enabled or disabled) Enable or dissable pool member connection limits are overridden for persisted clients. Per-virtual connection limits remain hard limits and are not overridden.

`hash_algorithm` (Optional) Specify the hash algorithm

`mask` (Optional) Identify a range of source IP addresses to manage together as a single source address affinity persistent connection when connecting to the pool. Must be a valid IPv4 or IPv6 mask.

`map_proxies` (Optional) (enabled or disabled) Directs all to the same single pool member

## bigip_ltm_persistence_profile_ssl

Configures an SSL persistence profile

### Example

```
resource "bigip_ltm_persistence_profile_ssl" "ppssl" {
    name = "/Common/terraform_ssl"
    defaults_from = "/Common/ssl"
    match_across_pools = "enabled"
    match_across_services = "enabled"
    match_across_virtuals = "enabled"
    mirror = "enabled"
    timeout = 3600
    override_conn_limit = "enabled"
}
```

### Reference

`name` - (Required) Name of the virtual address

`defaults_from` - (Required) Parent cookie persistence profile

`match_across_pools` (Optional) (enabled or disabled) match across pools with given persistence record

`match_across_services` (Optional) (enabled or disabled) match across services with given persistence record

`match_across_virtuals` (Optional) (enabled or disabled) match across virtual servers with given persistence record

`mirror` (Optional) (enabled or disabled) mirror persistence record

`timeout` (Optional) (enabled or disabled) Timeout for persistence of the session in seconds

`override_conn_limit` (Optional) (enabled or disabled) Enable or dissable pool member connection limits are overridden for persisted clients. Per-virtual connection limits remain hard limits and are not overridden.

# Building

Create the distributable packages like so:

```
make get-deps && make bin && make dist
```

See these pages for more information:

 * https://www.terraform.io/docs/internals/internal-plugins.html
 * https://github.com/hashicorp/terraform#developing-terraform

# Testing

Running the acceptance test suite requires an F5 to test against. Set `BIGIP_HOST`, `BIGIP_USER`
and `BIGIP_PASSWORD` to a device to run the tests against. By default tests will use the `Common`
partition for creating objects. You can change the partition by setting `BIGIP_TEST_PARTITION`.

```
BIGIP_HOST=f5.mycompany.com BIGIP_USER=foo BIGIP_PASSWORD=secret make testacc
```


Read [here](https://github.com/hashicorp/terraform/blob/master/.github/CONTRIBUTING.md#running-an-acceptance-test) for
more information about acceptance testing in Terraform.
