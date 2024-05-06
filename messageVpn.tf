
terraform {
  required_providers {
    solacebroker = {
      source = "registry.terraform.io/solaceproducts/solacebroker"
    }
  }
}

provider "solacebroker" {
  # username       = "admin"                 # This is a placeholder.
  # password       = "admin"                 # This is a placeholder.
  url            = "http://localhost:8080"
}


resource "solacebroker_broker" "miau" {
  guaranteed_msging_enabled                                            = true
  guaranteed_msging_event_cache_usage_threshold                        = {"clear_percent":60,"set_percent":80}
  guaranteed_msging_event_delivered_unacked_threshold                  = {"clear_percent":60,"set_percent":80}
  guaranteed_msging_event_disk_usage_threshold                         = {"clear_percent":60,"set_percent":80}
  guaranteed_msging_event_egress_flow_count_threshold                  = {"clear_percent":60,"set_percent":80}
  guaranteed_msging_event_endpoint_count_threshold                     = {"clear_percent":60,"set_percent":80}
  guaranteed_msging_event_ingress_flow_count_threshold                 = {"clear_percent":60,"set_percent":80}
  guaranteed_msging_event_msg_count_threshold                          = {"clear_percent":60,"set_percent":80}
  guaranteed_msging_event_msg_spool_file_count_threshold               = {"clear_percent":60,"set_percent":80}
  guaranteed_msging_event_msg_spool_usage_threshold                    = {"clear_percent":60,"set_percent":80}
  guaranteed_msging_event_transacted_session_count_threshold           = {"clear_percent":60,"set_percent":80}
  guaranteed_msging_event_transacted_session_resource_count_threshold  = {"clear_percent":60,"set_percent":80}
  guaranteed_msging_event_transaction_count_threshold                  = {"clear_percent":60,"set_percent":80}
  service_amqp_enabled                                                 = true
  service_event_connection_count_threshold                             = {"clear_percent":60,"set_percent":80}
  service_health_check_enabled                                         = true
  service_health_check_tls_enabled                                     = true
  service_health_check_tls_listen_port                                 = 5553
  service_mqtt_enabled                                                 = true
  service_rest_event_outgoing_connection_count_threshold               = {"clear_percent":60,"set_percent":80}
  service_rest_incoming_enabled                                        = true
  service_rest_outgoing_enabled                                        = true
  service_semp_cors_allow_any_host_enabled                             = false
  service_smf_event_connection_count_threshold                         = {"clear_percent":60,"set_percent":80}
  service_tls_event_connection_count_threshold                         = {"clear_percent":60,"set_percent":80}
  service_web_transport_enabled                                        = true
  tls_block_version11_enabled                                          = true

}
