// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

service "ec2" {

  sdk {
    id             = "EC2"
    client_version = [2]
  }

  names {
    provider_name_upper = "EC2"
    human_friendly      = "EC2 (Elastic Compute Cloud)"
  }

  client {
    go_v1_client_typename = "EC2"
    skip_client_generate  = true
  }

  endpoint_info {
    endpoint_api_call        = "DescribeVpcs"
  }

  resource_prefix {
    actual  = "aws_(ami|availability_zone|ec2_(availability|capacity|fleet|host|instance|public_ipv4_pool|serial|spot|tag)|eip|instance|key_pair|launch_template|placement_group|spot)"
    correct = "aws_ec2_"
  }

  sub_service "ec2ebs" {

    cli_v2_command {
          aws_cli_v2_command                    = ""
          aws_cli_v2_command_no_dashes          = ""
    }

    go_packages {
          v1_package                   = ""
          v2_package                   = ""
    }

    sdk {
      id             = ""
      client_version = null
    }

    names {
        provider_name_upper = "EC2EBS"
      human_friendly      = "EBS (EC2)"
    }

    resource_prefix {
      actual  = "aws_(ebs_|volume_attach|snapshot_create)"
      correct = "aws_ec2ebs_"
    }

    split_package            = "ec2"
    file_prefix              = "ebs_"
    doc_prefix               = ["ebs_", "volume_attachment", "snapshot_"]
    brand                    = "Amazon"
    exclude                  = true
      allowed_subcategory      = true
    note                     = "Part of EC2"
  }

  sub_service "ec2outposts" {

    cli_v2_command {
          aws_cli_v2_command           = ""
          aws_cli_v2_command_no_dashes          = ""
      }

    go_packages {
          v1_package                   = ""
          v2_package                   = ""
      }

    sdk {
      id             = ""
      client_version = null
    }

    names {
        provider_name_upper = "EC2Outposts"
      human_friendly      = "Outposts (EC2)"
    }

    resource_prefix {
      actual  = "aws_ec2_(coip_pool|local_gateway)"
      correct = "aws_ec2outposts_"
    }

    split_package            = "ec2"
    file_prefix              = "outposts_"
    doc_prefix               = ["ec2_coip_pool", "ec2_local_gateway"]
    brand                    = "AWS"
    exclude                  = true
      allowed_subcategory      = true
    note                     = "Part of EC2"
  }

  sub_service "transitgateway" {

  cli_v2_command {
        aws_cli_v2_command           = ""
        aws_cli_v2_command_no_dashes           = ""
    }

  go_packages {
        v1_package                   = ""
        v2_package                   = ""
    }

    sdk {
      id             = ""
      client_version = null
    }

    names {
        provider_name_upper = "TransitGateway"
      human_friendly      = "Transit Gateway"
    }

    resource_prefix {
      actual  = "aws_ec2_transit_gateway"
      correct = "aws_transitgateway_"
    }

    split_package            = "ec2"
    file_prefix              = "transitgateway_"
    doc_prefix               = ["ec2_transit_gateway"]
    brand                    = "AWS"
    exclude                  = true
      allowed_subcategory      = true
    note                     = "Part of EC2"
  }

  sub_service "verifiedaccess" {

    cli_v2_command {
          aws_cli_v2_command           = ""
          aws_cli_v2_command_no_dashes           = ""
      }

    go_packages {
          v1_package                   = ""
          v2_package                   = ""
      }

    sdk {
      id             = ""
      client_version = null
    }

    names {
        provider_name_upper = "VerifiedAccess"
      human_friendly      = "Verified Access"
    }

    resource_prefix {
      actual  = "aws_verifiedaccess"
      correct = "aws_verifiedaccess_"
    }
    
    split_package            = "ec2"
    file_prefix              = "verifiedaccess_"
    doc_prefix               = ["verifiedaccess_"]
    brand                    = "AWS"
    exclude                  = true
      allowed_subcategory      = true
    note                     = "Part of EC2"
  }

  sub_service "vpc" {

    cli_v2_command {
          aws_cli_v2_command           = ""
          aws_cli_v2_command_no_dashes          = ""
      }

    go_packages {
          v1_package                   = ""
          v2_package                   = ""
      }

    sdk {
      id             = ""
      client_version = null
    }

    names {
        provider_name_upper = "VPC"
      human_friendly      = "VPC (Virtual Private Cloud)"
    }

    resource_prefix {
      actual  = "aws_((default_)?(network_acl|route_table|security_group|subnet|vpc(?!_ipam))|ec2_(managed|network|subnet|traffic)|egress_only_internet|flow_log|internet_gateway|main_route_table_association|nat_gateway|network_interface|prefix_list|route\\b)"
      correct = "aws_vpc_"
    }

    split_package            = "ec2"
    file_prefix              = "vpc_"
    doc_prefix               = ["default_network_", "default_route_", "default_security_", "default_subnet", "default_vpc", "ec2_managed_", "ec2_network_", "ec2_subnet_", "ec2_traffic_", "egress_only_", "flow_log", "internet_gateway", "main_route_", "nat_", "network_", "prefix_list", "route_", "route\\.", "security_group", "subnet", "vpc_dhcp_", "vpc_endpoint", "vpc_ipv", "vpc_network_performance", "vpc_peering_", "vpc_security_group_", "vpc\\.", "vpcs\\."]
    brand                    = "Amazon"
    exclude                  = true
      allowed_subcategory      = true
    note                     = "Part of EC2"
  }

  sub_service "ipam" {

    cli_v2_command {
          aws_cli_v2_command                    = ""
          aws_cli_v2_command_no_dashes          = ""
    }

    go_packages {
          v1_package                   = ""
          v2_package                   = ""
    }
      
    sdk {
      id             = ""
      client_version = null
    }

    names {
        provider_name_upper = "IPAM"
      human_friendly      = "VPC IPAM (IP Address Manager)"
    }

    resource_prefix {
      actual  = "aws_vpc_ipam"
      correct = "aws_ipam_"
    }
    split_package            = "ec2"
    file_prefix              = "ipam_"
    doc_prefix               = ["vpc_ipam"]
    brand                    = "Amazon"
    exclude                  = true
      allowed_subcategory      = true
    note                     = "Part of EC2"
  }

  sub_service "vpnclient" {

    cli_v2_command {
          aws_cli_v2_command                    = ""
          aws_cli_v2_command_no_dashes          = ""
    }

    go_packages {
          v1_package                   = ""
          v2_package                   = ""
    }

    sdk {
      id             = ""
      client_version = null
    }

    names {
        provider_name_upper = "ClientVPN"
      human_friendly      = "VPN (Client)"
    }

    resource_prefix {
      actual  = "aws_ec2_client_vpn"
      correct = "aws_vpnclient_"
    }
    split_package            = "ec2"
    file_prefix              = "vpnclient_"
    doc_prefix               = ["ec2_client_vpn_"]
    brand                    = "AWS"
    exclude                  = true
      allowed_subcategory      = true
    note                     = "Part of EC2"
  }

  sub_service "vpnsite" {

    cli_v2_command {
          aws_cli_v2_command           = ""
          aws_cli_v2_command_no_dashes          = ""
    }

    go_packages {
          v1_package                   = ""
          v2_package                   = ""
    }

    sdk {
      id             = ""
      client_version = null
    }

    names {
        provider_name_upper = "SiteVPN"
      human_friendly      = "VPN (Site-to-Site)"
    }

    resource_prefix {
      actual  = "aws_(customer_gateway|vpn_)"
      correct = "aws_vpnsite_"
    }

    split_package            = "ec2"
    file_prefix              = "vpnsite_"
    doc_prefix               = ["customer_gateway", "vpn_"]
    brand                    = "AWS"
    exclude                  = true
      allowed_subcategory      = true
    note                     = "Part of EC2"
  }

  sub_service "wavelength" {

cli_v2_command {
       aws_cli_v2_command                     = ""
       aws_cli_v2_command_no_dashes           = ""
   }

go_packages {
       v1_package                   = ""
       v2_package                   = ""
   }
    sdk {
      id             = ""
      client_version = null
    }

    names {
        provider_name_upper = "Wavelength"
      human_friendly      = "Wavelength"
    }

    resource_prefix {
      actual  = "aws_ec2_carrier_gateway"
      correct = "aws_wavelength_"
    }

    split_package            = "ec2"
    file_prefix              = "wavelength_"
    doc_prefix               = ["ec2_carrier_"]
    brand                    = "AWS"
    exclude                  = true
      allowed_subcategory      = true
    note                     = "Part of EC2"
  }

  provider_package_correct = "ec2"
  split_package            = "ec2"
  file_prefix              = "ec2_"
  doc_prefix               = ["ami", "availability_zone", "ec2_availability_", "ec2_capacity_", "ec2_fleet", "ec2_host", "ec2_image_", "ec2_instance_", "ec2_public_ipv4_pool", "ec2_serial_", "ec2_spot_", "ec2_tag", "eip", "instance", "key_pair", "launch_template", "placement_group", "spot_"]
  brand                    = "Amazon"
}
