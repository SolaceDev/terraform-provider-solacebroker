---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "solacebroker_msg_vpn_distributed_cache_cluster_topic Data Source - solacebroker"
subcategory: ""
description: |-
  The Cache Instances that belong to the containing Cache Cluster will cache any messages published to topics that match a Topic Subscription.
  A SEMP client authorized with a minimum access scope/level of "vpn/read-only" is required to perform this operation.
  This has been available since SEMP API version 2.11.
---

# solacebroker_msg_vpn_distributed_cache_cluster_topic (Data Source)

The Cache Instances that belong to the containing Cache Cluster will cache any messages published to topics that match a Topic Subscription.



A SEMP client authorized with a minimum access scope/level of "vpn/read-only" is required to perform this operation.

This has been available since SEMP API version 2.11.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `cache_name` (String) The name of the Distributed Cache.
- `cluster_name` (String) The name of the Cache Cluster.
- `msg_vpn_name` (String) The name of the Message VPN.
- `topic` (String) The value of the Topic in the form a/b/c.
