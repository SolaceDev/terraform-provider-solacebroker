---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "solacebroker_msg_vpn_authentication_kerberos_realm Resource - solacebroker"
subcategory: ""
description: |-
  Kerberos Realm.
  A SEMP client authorized with a minimum access scope/level of "vpn/read-only" is required to perform this operation.
  This has been available since SEMP API version 2.40.
  The import identifier for this resource is {msg_vpn_name}/{kerberos_realm_name}, where {&lt;attribute&gt;} represents the value of the attribute and it must be URL-encoded.
---

# solacebroker_msg_vpn_authentication_kerberos_realm (Resource)

Kerberos Realm.



A SEMP client authorized with a minimum access scope/level of "vpn/read-only" is required to perform this operation.

This has been available since SEMP API version 2.40.

The import identifier for this resource is `{msg_vpn_name}/{kerberos_realm_name}`, where {&lt;attribute&gt;} represents the value of the attribute and it must be URL-encoded.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `kerberos_realm_name` (String) The Realm Name. Must start with "@", typically all uppercase.
- `msg_vpn_name` (String) The name of the Message VPN.

### Optional

- `enabled` (Boolean) Enable or disable the Realm. Changes to this attribute are synchronized to HA mates and replication sites via config-sync. The default value is `false`.
- `kdc_address` (String) Address (FQDN or IP) and optional port of the Key Distribution Center for principals in this Realm. Modifying this attribute while the object (or the relevant part of the object) is administratively enabled may be service impacting as enabled will be temporarily set to false to apply the change. Changes to this attribute are synchronized to HA mates and replication sites via config-sync. The default value is `""`.
