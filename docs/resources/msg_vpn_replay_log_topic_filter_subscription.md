---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "solacebroker_msg_vpn_replay_log_topic_filter_subscription Resource - solacebroker"
subcategory: ""
description: |-
  One or more Subscriptions can be added to a replay-log so that only guaranteed messages published to matching topics are stored in the Replay Log.
  A SEMP client authorized with a minimum access scope/level of "vpn/read-only" is required to perform this operation.
  This has been available since SEMP API version 2.27.
  The import identifier for this resource is {msg_vpn_name}/{replay_log_name}/{topic_filter_subscription}, where {&lt;attribute&gt;} represents the value of the attribute and it must be URL-encoded.
---

# solacebroker_msg_vpn_replay_log_topic_filter_subscription (Resource)

One or more Subscriptions can be added to a replay-log so that only guaranteed messages published to matching topics are stored in the Replay Log.



A SEMP client authorized with a minimum access scope/level of "vpn/read-only" is required to perform this operation.

This has been available since SEMP API version 2.27.

The import identifier for this resource is `{msg_vpn_name}/{replay_log_name}/{topic_filter_subscription}`, where {&lt;attribute&gt;} represents the value of the attribute and it must be URL-encoded.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `msg_vpn_name` (String) The name of the Message VPN.
- `replay_log_name` (String) The name of the Replay Log.
- `topic_filter_subscription` (String) The topic of the Subscription.
