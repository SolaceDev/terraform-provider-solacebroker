---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "solacebroker_msg_vpn_client_username_attribute Data Source - solacebroker"
subcategory: ""
description: |-
  A ClientUsername Attribute is a key+value pair that can be used to locate a client username, for example when using client certificate mapping.
  A SEMP client authorized with a minimum access scope/level of "vpn/read-only" is required to perform this operation.
  This has been available since SEMP API version 2.27.
---

# solacebroker_msg_vpn_client_username_attribute (Data Source)

A ClientUsername Attribute is a key+value pair that can be used to locate a client username, for example when using client certificate mapping.



A SEMP client authorized with a minimum access scope/level of "vpn/read-only" is required to perform this operation.

This has been available since SEMP API version 2.27.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `attribute_name` (String) The name of the Attribute.
- `attribute_value` (String) The value of the Attribute.
- `client_username` (String) The name of the Client Username.
- `msg_vpn_name` (String) The name of the Message VPN.
