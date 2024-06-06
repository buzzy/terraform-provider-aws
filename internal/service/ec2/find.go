// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"strconv"

	aws_sdkv2 "github.com/aws/aws-sdk-go-v2/aws"
	ec2_sdkv2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	tfawserr_sdkv2 "github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindCOIPPools(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeCoipPoolsInput) ([]*ec2.CoipPool, error) {
	var output []*ec2.CoipPool

	err := conn.DescribeCoipPoolsPagesWithContext(ctx, input, func(page *ec2.DescribeCoipPoolsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.CoipPools {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidPoolIDNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindCOIPPool(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeCoipPoolsInput) (*ec2.CoipPool, error) {
	output, err := FindCOIPPools(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findEBSVolumes(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeVolumesInput) ([]*ec2.Volume, error) {
	var output []*ec2.Volume

	err := conn.DescribeVolumesPagesWithContext(ctx, input, func(page *ec2.DescribeVolumesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Volumes {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVolumeNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindEBSVolumeByID(ctx context.Context, conn *ec2_sdkv2.Client, id string) (*awstypes.Volume, error) {
	input := &ec2_sdkv2.DescribeVolumesInput{
		VolumeIds: []string{id},
	}

	output, err := findEBSVolumeV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.VolumeStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.VolumeId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindEBSVolumeV1(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeVolumesInput) (*ec2.Volume, error) {
	output, err := findEBSVolumes(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindEBSVolumeByIDV1(ctx context.Context, conn *ec2_sdkv2.Client, id string) (*awstypes.Volume, error) {
	input := &ec2_sdkv2.DescribeVolumesInput{
		VolumeIds: []string{id},
	}

	output, err := findEBSVolumeV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.VolumeStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.VolumeId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findEIPs(ctx context.Context, conn *ec2_sdkv2.Client, input *ec2_sdkv2.DescribeAddressesInput) ([]awstypes.Address, error) {
	output, err := conn.DescribeAddresses(ctx, input)

	if tfawserr_sdkv2.ErrCodeEquals(err, errCodeInvalidAddressNotFound, errCodeInvalidAllocationIDNotFound) ||
		tfawserr_sdkv2.ErrMessageContains(err, errCodeAuthFailure, "does not belong to you") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Addresses, nil
}

func findEIP(ctx context.Context, conn *ec2_sdkv2.Client, input *ec2_sdkv2.DescribeAddressesInput) (*awstypes.Address, error) {
	output, err := findEIPs(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findEIPByAllocationID(ctx context.Context, conn *ec2_sdkv2.Client, id string) (*awstypes.Address, error) {
	input := &ec2_sdkv2.DescribeAddressesInput{
		AllocationIds: []string{id},
	}

	output, err := findEIP(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws_sdkv2.ToString(output.AllocationId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findEIPByAssociationID(ctx context.Context, conn *ec2_sdkv2.Client, id string) (*awstypes.Address, error) {
	input := &ec2_sdkv2.DescribeAddressesInput{
		Filters: newAttributeFilterListV2(map[string]string{
			"association-id": id,
		}),
	}

	output, err := findEIP(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws_sdkv2.ToString(output.AssociationId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findEIPAttributes(ctx context.Context, conn *ec2_sdkv2.Client, input *ec2_sdkv2.DescribeAddressesAttributeInput) ([]awstypes.AddressAttribute, error) {
	var output []awstypes.AddressAttribute

	pages := ec2_sdkv2.NewDescribeAddressesAttributePaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.Addresses...)
	}

	return output, nil
}

func findEIPAttribute(ctx context.Context, conn *ec2_sdkv2.Client, input *ec2_sdkv2.DescribeAddressesAttributeInput) (*awstypes.AddressAttribute, error) {
	output, err := findEIPAttributes(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findEIPDomainNameAttributeByAllocationID(ctx context.Context, conn *ec2_sdkv2.Client, id string) (*awstypes.AddressAttribute, error) {
	input := &ec2_sdkv2.DescribeAddressesAttributeInput{
		AllocationIds: []string{id},
		Attribute:     awstypes.AddressAttributeNameDomainName,
	}

	output, err := findEIPAttribute(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws_sdkv2.ToString(output.AllocationId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindLocalGatewayRouteTables(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeLocalGatewayRouteTablesInput) ([]*ec2.LocalGatewayRouteTable, error) {
	var output []*ec2.LocalGatewayRouteTable

	err := conn.DescribeLocalGatewayRouteTablesPagesWithContext(ctx, input, func(page *ec2.DescribeLocalGatewayRouteTablesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.LocalGatewayRouteTables {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindLocalGatewayRouteTable(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeLocalGatewayRouteTablesInput) (*ec2.LocalGatewayRouteTable, error) {
	output, err := FindLocalGatewayRouteTables(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindLocalGatewayVirtualInterfaceGroups(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeLocalGatewayVirtualInterfaceGroupsInput) ([]*ec2.LocalGatewayVirtualInterfaceGroup, error) {
	var output []*ec2.LocalGatewayVirtualInterfaceGroup

	err := conn.DescribeLocalGatewayVirtualInterfaceGroupsPagesWithContext(ctx, input, func(page *ec2.DescribeLocalGatewayVirtualInterfaceGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.LocalGatewayVirtualInterfaceGroups {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindLocalGatewayVirtualInterfaceGroup(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeLocalGatewayVirtualInterfaceGroupsInput) (*ec2.LocalGatewayVirtualInterfaceGroup, error) {
	output, err := FindLocalGatewayVirtualInterfaceGroups(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindLocalGateways(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeLocalGatewaysInput) ([]*ec2.LocalGateway, error) {
	var output []*ec2.LocalGateway

	err := conn.DescribeLocalGatewaysPagesWithContext(ctx, input, func(page *ec2.DescribeLocalGatewaysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.LocalGateways {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindLocalGateway(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeLocalGatewaysInput) (*ec2.LocalGateway, error) {
	output, err := FindLocalGateways(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindNetworkACL(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeNetworkAclsInput) (*ec2.NetworkAcl, error) {
	output, err := FindNetworkACLs(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindNetworkACLs(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeNetworkAclsInput) ([]*ec2.NetworkAcl, error) {
	var output []*ec2.NetworkAcl

	err := conn.DescribeNetworkAclsPagesWithContext(ctx, input, func(page *ec2.DescribeNetworkAclsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.NetworkAcls {
			if v == nil {
				continue
			}

			output = append(output, v)
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidNetworkACLIDNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindNetworkACLByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.NetworkAcl, error) {
	input := &ec2.DescribeNetworkAclsInput{
		NetworkAclIds: aws.StringSlice([]string{id}),
	}

	output, err := FindNetworkACL(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.NetworkAclId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindNetworkACLAssociationByID(ctx context.Context, conn *ec2.EC2, associationID string) (*ec2.NetworkAclAssociation, error) {
	input := &ec2.DescribeNetworkAclsInput{
		Filters: newAttributeFilterList(map[string]string{
			"association.association-id": associationID,
		}),
	}

	output, err := FindNetworkACL(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	for _, v := range output.Associations {
		if aws.StringValue(v.NetworkAclAssociationId) == associationID {
			return v, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func FindNetworkACLAssociationBySubnetID(ctx context.Context, conn *ec2.EC2, subnetID string) (*ec2.NetworkAclAssociation, error) {
	input := &ec2.DescribeNetworkAclsInput{
		Filters: newAttributeFilterList(map[string]string{
			"association.subnet-id": subnetID,
		}),
	}

	output, err := FindNetworkACL(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	for _, v := range output.Associations {
		if aws.StringValue(v.SubnetId) == subnetID {
			return v, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func FindNetworkACLEntryByThreePartKey(ctx context.Context, conn *ec2.EC2, naclID string, egress bool, ruleNumber int) (*ec2.NetworkAclEntry, error) {
	input := &ec2.DescribeNetworkAclsInput{
		Filters: newAttributeFilterList(map[string]string{
			"entry.egress":      strconv.FormatBool(egress),
			"entry.rule-number": strconv.Itoa(ruleNumber),
		}),
		NetworkAclIds: aws.StringSlice([]string{naclID}),
	}

	output, err := FindNetworkACL(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	for _, v := range output.Entries {
		if aws.BoolValue(v.Egress) == egress && aws.Int64Value(v.RuleNumber) == int64(ruleNumber) {
			return v, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func FindNetworkInterface(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeNetworkInterfacesInput) (*ec2.NetworkInterface, error) {
	output, err := FindNetworkInterfaces(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindNetworkInterfaces(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeNetworkInterfacesInput) ([]*ec2.NetworkInterface, error) {
	var output []*ec2.NetworkInterface

	err := conn.DescribeNetworkInterfacesPagesWithContext(ctx, input, func(page *ec2.DescribeNetworkInterfacesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.NetworkInterfaces {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidNetworkInterfaceIDNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindNetworkInterfaceByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.NetworkInterface, error) {
	input := &ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: aws.StringSlice([]string{id}),
	}

	output, err := FindNetworkInterface(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.NetworkInterfaceId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindNetworkInterfaceByAttachmentID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.NetworkInterface, error) {
	input := &ec2.DescribeNetworkInterfacesInput{
		Filters: newAttributeFilterList(map[string]string{
			"attachment.attachment-id": id,
		}),
	}

	networkInterface, err := FindNetworkInterface(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if networkInterface == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return networkInterface, nil
}

func FindNetworkInterfaceAttachmentByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.NetworkInterfaceAttachment, error) {
	input := &ec2.DescribeNetworkInterfacesInput{
		Filters: newAttributeFilterList(map[string]string{
			"attachment.attachment-id": id,
		}),
	}

	networkInterface, err := FindNetworkInterface(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if networkInterface.Attachment == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return networkInterface.Attachment, nil
}

func FindNetworkInterfaceSecurityGroup(ctx context.Context, conn *ec2.EC2, networkInterfaceID string, securityGroupID string) (*ec2.GroupIdentifier, error) {
	networkInterface, err := FindNetworkInterfaceByID(ctx, conn, networkInterfaceID)

	if err != nil {
		return nil, err
	}

	for _, groupIdentifier := range networkInterface.Groups {
		if aws.StringValue(groupIdentifier.GroupId) == securityGroupID {
			return groupIdentifier, nil
		}
	}

	return nil, &retry.NotFoundError{
		LastError: fmt.Errorf("Network Interface (%s) Security Group (%s) not found", networkInterfaceID, securityGroupID),
	}
}

func FindNetworkInsightsAnalysis(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeNetworkInsightsAnalysesInput) (*ec2.NetworkInsightsAnalysis, error) {
	output, err := FindNetworkInsightsAnalyses(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindNetworkInsightsAnalyses(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeNetworkInsightsAnalysesInput) ([]*ec2.NetworkInsightsAnalysis, error) {
	var output []*ec2.NetworkInsightsAnalysis

	err := conn.DescribeNetworkInsightsAnalysesPagesWithContext(ctx, input, func(page *ec2.DescribeNetworkInsightsAnalysesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.NetworkInsightsAnalyses {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidNetworkInsightsAnalysisIdNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindNetworkInsightsAnalysisByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.NetworkInsightsAnalysis, error) {
	input := &ec2.DescribeNetworkInsightsAnalysesInput{
		NetworkInsightsAnalysisIds: aws.StringSlice([]string{id}),
	}

	output, err := FindNetworkInsightsAnalysis(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.NetworkInsightsAnalysisId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindNetworkInsightsPath(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeNetworkInsightsPathsInput) (*ec2.NetworkInsightsPath, error) {
	output, err := FindNetworkInsightsPaths(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindNetworkInsightsPaths(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeNetworkInsightsPathsInput) ([]*ec2.NetworkInsightsPath, error) {
	var output []*ec2.NetworkInsightsPath

	err := conn.DescribeNetworkInsightsPathsPagesWithContext(ctx, input, func(page *ec2.DescribeNetworkInsightsPathsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.NetworkInsightsPaths {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidNetworkInsightsPathIdNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindNetworkInsightsPathByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.NetworkInsightsPath, error) {
	input := &ec2.DescribeNetworkInsightsPathsInput{
		NetworkInsightsPathIds: aws.StringSlice([]string{id}),
	}

	output, err := FindNetworkInsightsPath(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.NetworkInsightsPathId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindSecurityGroupByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.SecurityGroup, error) {
	input := &ec2.DescribeSecurityGroupsInput{
		GroupIds: aws.StringSlice([]string{id}),
	}

	output, err := FindSecurityGroup(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.GroupId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

// FindSecurityGroupByNameAndVPCID looks up a security group by name, VPC ID. Returns a retry.NotFoundError if not found.
func FindSecurityGroupByNameAndVPCID(ctx context.Context, conn *ec2.EC2, name, vpcID string) (*ec2.SecurityGroup, error) {
	input := &ec2.DescribeSecurityGroupsInput{
		Filters: newAttributeFilterList(
			map[string]string{
				"group-name": name,
				"vpc-id":     vpcID,
			},
		),
	}
	return FindSecurityGroup(ctx, conn, input)
}

// FindSecurityGroupByNameAndVPCIDAndOwnerID looks up a security group by name, VPC ID and owner ID. Returns a retry.NotFoundError if not found.
func FindSecurityGroupByNameAndVPCIDAndOwnerID(ctx context.Context, conn *ec2.EC2, name, vpcID, ownerID string) (*ec2.SecurityGroup, error) {
	input := &ec2.DescribeSecurityGroupsInput{
		Filters: newAttributeFilterList(
			map[string]string{
				"group-name": name,
				"vpc-id":     vpcID,
				"owner-id":   ownerID,
			},
		),
	}
	return FindSecurityGroup(ctx, conn, input)
}

// FindSecurityGroup looks up a security group using an ec2.DescribeSecurityGroupsInput. Returns a retry.NotFoundError if not found.
func FindSecurityGroup(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeSecurityGroupsInput) (*ec2.SecurityGroup, error) {
	output, err := FindSecurityGroups(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindSecurityGroups(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeSecurityGroupsInput) ([]*ec2.SecurityGroup, error) {
	var output []*ec2.SecurityGroup

	err := conn.DescribeSecurityGroupsPagesWithContext(ctx, input, func(page *ec2.DescribeSecurityGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.SecurityGroups {
			if v == nil {
				continue
			}

			output = append(output, v)
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidGroupNotFound, errCodeInvalidSecurityGroupIDNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindSecurityGroupRule(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeSecurityGroupRulesInput) (*ec2.SecurityGroupRule, error) {
	output, err := FindSecurityGroupRules(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindSecurityGroupRules(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeSecurityGroupRulesInput) ([]*ec2.SecurityGroupRule, error) {
	var output []*ec2.SecurityGroupRule

	err := conn.DescribeSecurityGroupRulesPagesWithContext(ctx, input, func(page *ec2.DescribeSecurityGroupRulesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.SecurityGroupRules {
			if v == nil {
				continue
			}

			output = append(output, v)
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidSecurityGroupRuleIdNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindSecurityGroupRuleByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.SecurityGroupRule, error) {
	input := &ec2.DescribeSecurityGroupRulesInput{
		SecurityGroupRuleIds: aws.StringSlice([]string{id}),
	}

	output, err := FindSecurityGroupRule(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.SecurityGroupRuleId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindSecurityGroupEgressRuleByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.SecurityGroupRule, error) {
	output, err := FindSecurityGroupRuleByID(ctx, conn, id)

	if err != nil {
		return nil, err
	}

	if !aws.BoolValue(output.IsEgress) {
		return nil, &retry.NotFoundError{}
	}

	return output, nil
}

func FindSecurityGroupIngressRuleByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.SecurityGroupRule, error) {
	output, err := FindSecurityGroupRuleByID(ctx, conn, id)

	if err != nil {
		return nil, err
	}

	if aws.BoolValue(output.IsEgress) {
		return nil, &retry.NotFoundError{}
	}

	return output, nil
}

func FindSecurityGroupRulesBySecurityGroupID(ctx context.Context, conn *ec2.EC2, id string) ([]*ec2.SecurityGroupRule, error) {
	input := &ec2.DescribeSecurityGroupRulesInput{
		Filters: newAttributeFilterList(map[string]string{
			"group-id": id,
		}),
	}

	return FindSecurityGroupRules(ctx, conn, input)
}

func FindSubnetByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.Subnet, error) {
	input := &ec2.DescribeSubnetsInput{
		SubnetIds: aws.StringSlice([]string{id}),
	}

	output, err := FindSubnet(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.SubnetId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindSubnet(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeSubnetsInput) (*ec2.Subnet, error) {
	output, err := FindSubnets(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindSubnets(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeSubnetsInput) ([]*ec2.Subnet, error) {
	var output []*ec2.Subnet

	err := conn.DescribeSubnetsPagesWithContext(ctx, input, func(page *ec2.DescribeSubnetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Subnets {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidSubnetIDNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindSubnetCIDRReservationBySubnetIDAndReservationID(ctx context.Context, conn *ec2.EC2, subnetID, reservationID string) (*ec2.SubnetCidrReservation, error) {
	input := &ec2.GetSubnetCidrReservationsInput{
		SubnetId: aws.String(subnetID),
	}

	output, err := conn.GetSubnetCidrReservationsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidSubnetIDNotFound) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || (len(output.SubnetIpv4CidrReservations) == 0 && len(output.SubnetIpv6CidrReservations) == 0) {
		return nil, tfresource.NewEmptyResultError(input)
	}

	for _, r := range output.SubnetIpv4CidrReservations {
		if aws.StringValue(r.SubnetCidrReservationId) == reservationID {
			return r, nil
		}
	}
	for _, r := range output.SubnetIpv6CidrReservations {
		if aws.StringValue(r.SubnetCidrReservationId) == reservationID {
			return r, nil
		}
	}

	return nil, &retry.NotFoundError{
		LastError:   err,
		LastRequest: input,
	}
}

func FindSubnetIPv6CIDRBlockAssociationByID(ctx context.Context, conn *ec2.EC2, associationID string) (*ec2.SubnetIpv6CidrBlockAssociation, error) {
	input := &ec2.DescribeSubnetsInput{
		Filters: newAttributeFilterList(map[string]string{
			"ipv6-cidr-block-association.association-id": associationID,
		}),
	}

	output, err := FindSubnet(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	for _, association := range output.Ipv6CidrBlockAssociationSet {
		if aws.StringValue(association.AssociationId) == associationID {
			if state := aws.StringValue(association.Ipv6CidrBlockState.State); state == ec2.SubnetCidrBlockStateCodeDisassociated {
				return nil, &retry.NotFoundError{Message: state}
			}

			return association, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func FindVPCAttribute(ctx context.Context, conn *ec2.EC2, vpcID string, attribute string) (bool, error) {
	input := &ec2.DescribeVpcAttributeInput{
		Attribute: aws.String(attribute),
		VpcId:     aws.String(vpcID),
	}

	output, err := conn.DescribeVpcAttributeWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCIDNotFound) {
		return false, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return false, err
	}

	if output == nil {
		return false, tfresource.NewEmptyResultError(input)
	}

	var v *ec2.AttributeBooleanValue
	switch attribute {
	case ec2.VpcAttributeNameEnableDnsHostnames:
		v = output.EnableDnsHostnames
	case ec2.VpcAttributeNameEnableDnsSupport:
		v = output.EnableDnsSupport
	case ec2.VpcAttributeNameEnableNetworkAddressUsageMetrics:
		v = output.EnableNetworkAddressUsageMetrics
	default:
		return false, fmt.Errorf("unsupported VPC attribute: %s", attribute)
	}

	if v == nil {
		return false, tfresource.NewEmptyResultError(input)
	}

	return aws.BoolValue(v.Value), nil
}

func FindVPC(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeVpcsInput) (*ec2.Vpc, error) {
	output, err := FindVPCs(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindVPCs(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeVpcsInput) ([]*ec2.Vpc, error) {
	var output []*ec2.Vpc

	err := conn.DescribeVpcsPagesWithContext(ctx, input, func(page *ec2.DescribeVpcsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Vpcs {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCIDNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindVPCByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.Vpc, error) {
	input := &ec2.DescribeVpcsInput{
		VpcIds: aws.StringSlice([]string{id}),
	}

	output, err := FindVPC(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.VpcId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindVPCDHCPOptionsAssociation(ctx context.Context, conn *ec2.EC2, vpcID string, dhcpOptionsID string) error {
	vpc, err := FindVPCByID(ctx, conn, vpcID)

	if err != nil {
		return err
	}

	if aws.StringValue(vpc.DhcpOptionsId) != dhcpOptionsID {
		return &retry.NotFoundError{
			LastError: fmt.Errorf("EC2 VPC (%s) DHCP Options Set (%s) Association not found", vpcID, dhcpOptionsID),
		}
	}

	return nil
}

func FindVPCCIDRBlockAssociationByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.VpcCidrBlockAssociation, *ec2.Vpc, error) {
	input := &ec2.DescribeVpcsInput{
		Filters: newAttributeFilterList(map[string]string{
			"cidr-block-association.association-id": id,
		}),
	}

	vpc, err := FindVPC(ctx, conn, input)

	if err != nil {
		return nil, nil, err
	}

	for _, association := range vpc.CidrBlockAssociationSet {
		if aws.StringValue(association.AssociationId) == id {
			if state := aws.StringValue(association.CidrBlockState.State); state == ec2.VpcCidrBlockStateCodeDisassociated {
				return nil, nil, &retry.NotFoundError{Message: state}
			}

			return association, vpc, nil
		}
	}

	return nil, nil, &retry.NotFoundError{}
}

func FindVPCIPv6CIDRBlockAssociationByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.VpcIpv6CidrBlockAssociation, *ec2.Vpc, error) {
	input := &ec2.DescribeVpcsInput{
		Filters: newAttributeFilterList(map[string]string{
			"ipv6-cidr-block-association.association-id": id,
		}),
	}

	vpc, err := FindVPC(ctx, conn, input)

	if err != nil {
		return nil, nil, err
	}

	for _, association := range vpc.Ipv6CidrBlockAssociationSet {
		if aws.StringValue(association.AssociationId) == id {
			if state := aws.StringValue(association.Ipv6CidrBlockState.State); state == ec2.VpcCidrBlockStateCodeDisassociated {
				return nil, nil, &retry.NotFoundError{Message: state}
			}

			return association, vpc, nil
		}
	}

	return nil, nil, &retry.NotFoundError{}
}

func FindVPCDefaultNetworkACL(ctx context.Context, conn *ec2.EC2, id string) (*ec2.NetworkAcl, error) {
	input := &ec2.DescribeNetworkAclsInput{
		Filters: newAttributeFilterList(map[string]string{
			"default": "true",
			"vpc-id":  id,
		}),
	}

	return FindNetworkACL(ctx, conn, input)
}

func FindVPCDefaultSecurityGroup(ctx context.Context, conn *ec2.EC2, id string) (*ec2.SecurityGroup, error) {
	input := &ec2.DescribeSecurityGroupsInput{
		Filters: newAttributeFilterList(map[string]string{
			"group-name": DefaultSecurityGroupName,
			"vpc-id":     id,
		}),
	}

	return FindSecurityGroup(ctx, conn, input)
}

func FindVPCEndpoint(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeVpcEndpointsInput) (*ec2.VpcEndpoint, error) {
	output, err := FindVPCEndpoints(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindVPCEndpoints(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeVpcEndpointsInput) ([]*ec2.VpcEndpoint, error) {
	var output []*ec2.VpcEndpoint

	err := conn.DescribeVpcEndpointsPagesWithContext(ctx, input, func(page *ec2.DescribeVpcEndpointsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.VpcEndpoints {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCEndpointIdNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindVPCEndpointByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.VpcEndpoint, error) {
	input := &ec2.DescribeVpcEndpointsInput{
		VpcEndpointIds: aws.StringSlice([]string{id}),
	}

	output, err := FindVPCEndpoint(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == vpcEndpointStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.VpcEndpointId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindVPCConnectionNotification(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeVpcEndpointConnectionNotificationsInput) (*ec2.ConnectionNotification, error) {
	output, err := FindVPCConnectionNotifications(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindVPCConnectionNotifications(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeVpcEndpointConnectionNotificationsInput) ([]*ec2.ConnectionNotification, error) {
	var output []*ec2.ConnectionNotification

	err := conn.DescribeVpcEndpointConnectionNotificationsPagesWithContext(ctx, input, func(page *ec2.DescribeVpcEndpointConnectionNotificationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ConnectionNotificationSet {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidConnectionNotification) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindVPCConnectionNotificationByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.ConnectionNotification, error) {
	input := &ec2.DescribeVpcEndpointConnectionNotificationsInput{
		ConnectionNotificationId: aws.String(id),
	}

	output, err := FindVPCConnectionNotification(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.ConnectionNotificationId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindVPCEndpointServiceConfiguration(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeVpcEndpointServiceConfigurationsInput) (*ec2.ServiceConfiguration, error) {
	output, err := FindVPCEndpointServiceConfigurations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindVPCEndpointServiceConfigurations(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeVpcEndpointServiceConfigurationsInput) ([]*ec2.ServiceConfiguration, error) {
	var output []*ec2.ServiceConfiguration

	err := conn.DescribeVpcEndpointServiceConfigurationsPagesWithContext(ctx, input, func(page *ec2.DescribeVpcEndpointServiceConfigurationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ServiceConfigurations {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCEndpointServiceIdNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindVPCEndpointServiceConfigurationByServiceName(ctx context.Context, conn *ec2.EC2, name string) (*ec2.ServiceConfiguration, error) {
	input := &ec2.DescribeVpcEndpointServiceConfigurationsInput{
		Filters: newAttributeFilterList(map[string]string{
			"service-name": name,
		}),
	}

	return FindVPCEndpointServiceConfiguration(ctx, conn, input)
}

func FindVPCEndpointServiceConfigurationByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.ServiceConfiguration, error) {
	input := &ec2.DescribeVpcEndpointServiceConfigurationsInput{
		ServiceIds: aws.StringSlice([]string{id}),
	}

	output, err := FindVPCEndpointServiceConfiguration(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.ServiceState); state == ec2.ServiceStateDeleted || state == ec2.ServiceStateFailed {
		return nil, &retry.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.ServiceId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindVPCEndpointServicePermissions(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeVpcEndpointServicePermissionsInput) ([]*ec2.AllowedPrincipal, error) {
	var output []*ec2.AllowedPrincipal

	err := conn.DescribeVpcEndpointServicePermissionsPagesWithContext(ctx, input, func(page *ec2.DescribeVpcEndpointServicePermissionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.AllowedPrincipals {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCEndpointServiceIdNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindVPCEndpointServicePermissionsByServiceID(ctx context.Context, conn *ec2.EC2, id string) ([]*ec2.AllowedPrincipal, error) {
	input := &ec2.DescribeVpcEndpointServicePermissionsInput{
		ServiceId: aws.String(id),
	}

	return FindVPCEndpointServicePermissions(ctx, conn, input)
}

func FindVPCEndpointServicePermission(ctx context.Context, conn *ec2.EC2, serviceID, principalARN string) (*ec2.AllowedPrincipal, error) {
	// Applying a server-side filter on "principal" can lead to errors like
	// "An error occurred (InvalidFilter) when calling the DescribeVpcEndpointServicePermissions operation: The filter value arn:aws:iam::123456789012:role/developer contains unsupported characters".
	// Apply the filter client-side.
	input := &ec2.DescribeVpcEndpointServicePermissionsInput{
		ServiceId: aws.String(serviceID),
	}

	allowedPrincipals, err := FindVPCEndpointServicePermissions(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	allowedPrincipals = tfslices.Filter(allowedPrincipals, func(v *ec2.AllowedPrincipal) bool {
		return aws.StringValue(v.Principal) == principalARN
	})

	return tfresource.AssertSinglePtrResult(allowedPrincipals)
}

// FindVPCEndpointRouteTableAssociationExists returns NotFoundError if no association for the specified VPC endpoint and route table IDs is found.
func FindVPCEndpointRouteTableAssociationExists(ctx context.Context, conn *ec2.EC2, vpcEndpointID string, routeTableID string) error {
	vpcEndpoint, err := FindVPCEndpointByID(ctx, conn, vpcEndpointID)

	if err != nil {
		return err
	}

	for _, vpcEndpointRouteTableID := range vpcEndpoint.RouteTableIds {
		if aws.StringValue(vpcEndpointRouteTableID) == routeTableID {
			return nil
		}
	}

	return &retry.NotFoundError{
		LastError: fmt.Errorf("VPC Endpoint (%s) Route Table (%s) Association not found", vpcEndpointID, routeTableID),
	}
}

// FindVPCEndpointSecurityGroupAssociationExists returns NotFoundError if no association for the specified VPC endpoint and security group IDs is found.
func FindVPCEndpointSecurityGroupAssociationExists(ctx context.Context, conn *ec2.EC2, vpcEndpointID, securityGroupID string) error {
	vpcEndpoint, err := FindVPCEndpointByID(ctx, conn, vpcEndpointID)

	if err != nil {
		return err
	}

	for _, group := range vpcEndpoint.Groups {
		if aws.StringValue(group.GroupId) == securityGroupID {
			return nil
		}
	}

	return &retry.NotFoundError{
		LastError: fmt.Errorf("VPC Endpoint (%s) Security Group (%s) Association not found", vpcEndpointID, securityGroupID),
	}
}

// FindVPCEndpointSubnetAssociationExists returns NotFoundError if no association for the specified VPC endpoint and subnet IDs is found.
func FindVPCEndpointSubnetAssociationExists(ctx context.Context, conn *ec2.EC2, vpcEndpointID string, subnetID string) error {
	vpcEndpoint, err := FindVPCEndpointByID(ctx, conn, vpcEndpointID)

	if err != nil {
		return err
	}

	for _, vpcEndpointSubnetID := range vpcEndpoint.SubnetIds {
		if aws.StringValue(vpcEndpointSubnetID) == subnetID {
			return nil
		}
	}

	return &retry.NotFoundError{
		LastError: fmt.Errorf("VPC Endpoint (%s) Subnet (%s) Association not found", vpcEndpointID, subnetID),
	}
}

func FindVPCPeeringConnection(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeVpcPeeringConnectionsInput) (*ec2.VpcPeeringConnection, error) {
	output, err := FindVPCPeeringConnections(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output, func(v *ec2.VpcPeeringConnection) bool { return v.Status != nil })
}

func FindVPCPeeringConnections(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeVpcPeeringConnectionsInput) ([]*ec2.VpcPeeringConnection, error) {
	var output []*ec2.VpcPeeringConnection

	err := conn.DescribeVpcPeeringConnectionsPagesWithContext(ctx, input, func(page *ec2.DescribeVpcPeeringConnectionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.VpcPeeringConnections {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCPeeringConnectionIDNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindVPCPeeringConnectionByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.VpcPeeringConnection, error) {
	input := &ec2.DescribeVpcPeeringConnectionsInput{
		VpcPeeringConnectionIds: aws.StringSlice([]string{id}),
	}

	output, err := FindVPCPeeringConnection(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// See https://docs.aws.amazon.com/vpc/latest/peering/vpc-peering-basics.html#vpc-peering-lifecycle.
	switch statusCode := aws.StringValue(output.Status.Code); statusCode {
	case ec2.VpcPeeringConnectionStateReasonCodeDeleted,
		ec2.VpcPeeringConnectionStateReasonCodeExpired,
		ec2.VpcPeeringConnectionStateReasonCodeFailed,
		ec2.VpcPeeringConnectionStateReasonCodeRejected:
		return nil, &retry.NotFoundError{
			Message:     statusCode,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.VpcPeeringConnectionId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindTrafficMirrorFilter(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeTrafficMirrorFiltersInput) (*ec2.TrafficMirrorFilter, error) {
	output, err := FindTrafficMirrorFilters(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindTrafficMirrorFilters(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeTrafficMirrorFiltersInput) ([]*ec2.TrafficMirrorFilter, error) {
	var output []*ec2.TrafficMirrorFilter

	err := conn.DescribeTrafficMirrorFiltersPagesWithContext(ctx, input, func(page *ec2.DescribeTrafficMirrorFiltersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TrafficMirrorFilters {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTrafficMirrorFilterIdNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindTrafficMirrorFilterByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.TrafficMirrorFilter, error) {
	input := &ec2.DescribeTrafficMirrorFiltersInput{
		TrafficMirrorFilterIds: aws.StringSlice([]string{id}),
	}

	output, err := FindTrafficMirrorFilter(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.TrafficMirrorFilterId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindTrafficMirrorFilterRuleByTwoPartKey(ctx context.Context, conn *ec2.EC2, filterID, ruleID string) (*ec2.TrafficMirrorFilterRule, error) {
	output, err := FindTrafficMirrorFilterByID(ctx, conn, filterID)

	if err != nil {
		return nil, err
	}

	for _, v := range [][]*ec2.TrafficMirrorFilterRule{output.IngressFilterRules, output.EgressFilterRules} {
		for _, v := range v {
			if aws.StringValue(v.TrafficMirrorFilterRuleId) == ruleID {
				return v, nil
			}
		}
	}

	return nil, &retry.NotFoundError{}
}

func FindTrafficMirrorSession(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeTrafficMirrorSessionsInput) (*ec2.TrafficMirrorSession, error) {
	output, err := FindTrafficMirrorSessions(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindTrafficMirrorSessions(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeTrafficMirrorSessionsInput) ([]*ec2.TrafficMirrorSession, error) {
	var output []*ec2.TrafficMirrorSession

	err := conn.DescribeTrafficMirrorSessionsPagesWithContext(ctx, input, func(page *ec2.DescribeTrafficMirrorSessionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TrafficMirrorSessions {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTrafficMirrorSessionIdNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindTrafficMirrorSessionByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.TrafficMirrorSession, error) {
	input := &ec2.DescribeTrafficMirrorSessionsInput{
		TrafficMirrorSessionIds: aws.StringSlice([]string{id}),
	}

	output, err := FindTrafficMirrorSession(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.TrafficMirrorSessionId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindTrafficMirrorTarget(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeTrafficMirrorTargetsInput) (*ec2.TrafficMirrorTarget, error) {
	output, err := FindTrafficMirrorTargets(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindTrafficMirrorTargets(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeTrafficMirrorTargetsInput) ([]*ec2.TrafficMirrorTarget, error) {
	var output []*ec2.TrafficMirrorTarget

	err := conn.DescribeTrafficMirrorTargetsPagesWithContext(ctx, input, func(page *ec2.DescribeTrafficMirrorTargetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TrafficMirrorTargets {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTrafficMirrorTargetIdNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindTrafficMirrorTargetByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.TrafficMirrorTarget, error) {
	input := &ec2.DescribeTrafficMirrorTargetsInput{
		TrafficMirrorTargetIds: aws.StringSlice([]string{id}),
	}

	output, err := FindTrafficMirrorTarget(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.TrafficMirrorTargetId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindDHCPOptions(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeDhcpOptionsInput) (*ec2.DhcpOptions, error) {
	output, err := FindDHCPOptionses(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindDHCPOptionses(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeDhcpOptionsInput) ([]*ec2.DhcpOptions, error) {
	var output []*ec2.DhcpOptions

	err := conn.DescribeDhcpOptionsPagesWithContext(ctx, input, func(page *ec2.DescribeDhcpOptionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DhcpOptions {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidDHCPOptionIDNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindDHCPOptionsByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.DhcpOptions, error) {
	input := &ec2.DescribeDhcpOptionsInput{
		DhcpOptionsIds: aws.StringSlice([]string{id}),
	}

	output, err := FindDHCPOptions(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.DhcpOptionsId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindEgressOnlyInternetGateway(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeEgressOnlyInternetGatewaysInput) (*ec2.EgressOnlyInternetGateway, error) {
	output, err := FindEgressOnlyInternetGateways(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindEgressOnlyInternetGateways(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeEgressOnlyInternetGatewaysInput) ([]*ec2.EgressOnlyInternetGateway, error) {
	var output []*ec2.EgressOnlyInternetGateway

	err := conn.DescribeEgressOnlyInternetGatewaysPagesWithContext(ctx, input, func(page *ec2.DescribeEgressOnlyInternetGatewaysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.EgressOnlyInternetGateways {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindEgressOnlyInternetGatewayByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.EgressOnlyInternetGateway, error) {
	input := &ec2.DescribeEgressOnlyInternetGatewaysInput{
		EgressOnlyInternetGatewayIds: aws.StringSlice([]string{id}),
	}

	output, err := FindEgressOnlyInternetGateway(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.EgressOnlyInternetGatewayId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindFlowLogByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.FlowLog, error) {
	input := &ec2.DescribeFlowLogsInput{
		FlowLogIds: aws.StringSlice([]string{id}),
	}

	output, err := FindFlowLog(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.FlowLogId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindFlowLogs(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeFlowLogsInput) ([]*ec2.FlowLog, error) {
	var output []*ec2.FlowLog

	err := conn.DescribeFlowLogsPagesWithContext(ctx, input, func(page *ec2.DescribeFlowLogsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.FlowLogs {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindFlowLog(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeFlowLogsInput) (*ec2.FlowLog, error) {
	output, err := FindFlowLogs(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindInternetGateway(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeInternetGatewaysInput) (*ec2.InternetGateway, error) {
	output, err := FindInternetGateways(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindInternetGateways(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeInternetGatewaysInput) ([]*ec2.InternetGateway, error) {
	var output []*ec2.InternetGateway

	err := conn.DescribeInternetGatewaysPagesWithContext(ctx, input, func(page *ec2.DescribeInternetGatewaysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.InternetGateways {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidInternetGatewayIDNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindInternetGatewayByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.InternetGateway, error) {
	input := &ec2.DescribeInternetGatewaysInput{
		InternetGatewayIds: aws.StringSlice([]string{id}),
	}

	output, err := FindInternetGateway(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.InternetGatewayId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindInternetGatewayAttachment(ctx context.Context, conn *ec2.EC2, internetGatewayID, vpcID string) (*ec2.InternetGatewayAttachment, error) {
	internetGateway, err := FindInternetGatewayByID(ctx, conn, internetGatewayID)

	if err != nil {
		return nil, err
	}

	if len(internetGateway.Attachments) == 0 || internetGateway.Attachments[0] == nil {
		return nil, tfresource.NewEmptyResultError(internetGatewayID)
	}

	if count := len(internetGateway.Attachments); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, internetGatewayID)
	}

	attachment := internetGateway.Attachments[0]

	if aws.StringValue(attachment.VpcId) != vpcID {
		return nil, tfresource.NewEmptyResultError(vpcID)
	}

	return attachment, nil
}

func findKeyPair(ctx context.Context, conn *ec2_sdkv2.Client, input *ec2_sdkv2.DescribeKeyPairsInput) (*awstypes.KeyPairInfo, error) {
	output, err := findKeyPairs(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findKeyPairs(ctx context.Context, conn *ec2_sdkv2.Client, input *ec2_sdkv2.DescribeKeyPairsInput) ([]awstypes.KeyPairInfo, error) {
	output, err := conn.DescribeKeyPairs(ctx, input)

	if tfawserr_sdkv2.ErrCodeEquals(err, errCodeInvalidKeyPairNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output.KeyPairs, nil
}

func findKeyPairByName(ctx context.Context, conn *ec2_sdkv2.Client, name string) (*awstypes.KeyPairInfo, error) {
	input := &ec2_sdkv2.DescribeKeyPairsInput{
		KeyNames: []string{name},
	}

	output, err := findKeyPair(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws_sdkv2.ToString(output.KeyName) != name {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindManagedPrefixList(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeManagedPrefixListsInput) (*ec2.ManagedPrefixList, error) {
	output, err := FindManagedPrefixLists(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindManagedPrefixLists(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeManagedPrefixListsInput) ([]*ec2.ManagedPrefixList, error) {
	var output []*ec2.ManagedPrefixList

	err := conn.DescribeManagedPrefixListsPagesWithContext(ctx, input, func(page *ec2.DescribeManagedPrefixListsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.PrefixLists {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidPrefixListIDNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindManagedPrefixListByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.ManagedPrefixList, error) {
	input := &ec2.DescribeManagedPrefixListsInput{
		PrefixListIds: aws.StringSlice([]string{id}),
	}

	output, err := FindManagedPrefixList(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == ec2.PrefixListStateDeleteComplete {
		return nil, &retry.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.PrefixListId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindManagedPrefixListEntries(ctx context.Context, conn *ec2.EC2, input *ec2.GetManagedPrefixListEntriesInput) ([]*ec2.PrefixListEntry, error) {
	var output []*ec2.PrefixListEntry

	err := conn.GetManagedPrefixListEntriesPagesWithContext(ctx, input, func(page *ec2.GetManagedPrefixListEntriesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Entries {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidPrefixListIDNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindManagedPrefixListEntriesByID(ctx context.Context, conn *ec2.EC2, id string) ([]*ec2.PrefixListEntry, error) {
	input := &ec2.GetManagedPrefixListEntriesInput{
		PrefixListId: aws.String(id),
	}

	return FindManagedPrefixListEntries(ctx, conn, input)
}

func FindManagedPrefixListEntryByIDAndCIDR(ctx context.Context, conn *ec2.EC2, id, cidr string) (*ec2.PrefixListEntry, error) {
	prefixListEntries, err := FindManagedPrefixListEntriesByID(ctx, conn, id)

	if err != nil {
		return nil, err
	}

	for _, v := range prefixListEntries {
		if aws.StringValue(v.Cidr) == cidr {
			return v, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func FindNATGateway(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeNatGatewaysInput) (*ec2.NatGateway, error) {
	output, err := FindNATGateways(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindNATGateways(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeNatGatewaysInput) ([]*ec2.NatGateway, error) {
	var output []*ec2.NatGateway

	err := conn.DescribeNatGatewaysPagesWithContext(ctx, input, func(page *ec2.DescribeNatGatewaysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.NatGateways {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeNatGatewayNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindNATGatewayByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.NatGateway, error) {
	input := &ec2.DescribeNatGatewaysInput{
		NatGatewayIds: aws.StringSlice([]string{id}),
	}

	output, err := FindNATGateway(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == ec2.NatGatewayStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.NatGatewayId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindNATGatewayAddressByNATGatewayIDAndAllocationID(ctx context.Context, conn *ec2.EC2, natGatewayID, allocationID string) (*ec2.NatGatewayAddress, error) {
	output, err := FindNATGatewayByID(ctx, conn, natGatewayID)

	if err != nil {
		return nil, err
	}

	for _, v := range output.NatGatewayAddresses {
		if aws.StringValue(v.AllocationId) == allocationID {
			return v, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func FindNATGatewayAddressByNATGatewayIDAndPrivateIP(ctx context.Context, conn *ec2.EC2, natGatewayID, privateIP string) (*ec2.NatGatewayAddress, error) {
	output, err := FindNATGatewayByID(ctx, conn, natGatewayID)

	if err != nil {
		return nil, err
	}

	for _, v := range output.NatGatewayAddresses {
		if aws.StringValue(v.PrivateIp) == privateIP {
			return v, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func FindPrefixList(ctx context.Context, conn *ec2.EC2, input *ec2.DescribePrefixListsInput) (*ec2.PrefixList, error) {
	output, err := FindPrefixLists(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindPrefixLists(ctx context.Context, conn *ec2.EC2, input *ec2.DescribePrefixListsInput) ([]*ec2.PrefixList, error) {
	var output []*ec2.PrefixList

	err := conn.DescribePrefixListsPagesWithContext(ctx, input, func(page *ec2.DescribePrefixListsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.PrefixLists {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidPrefixListIdNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindPrefixListByName(ctx context.Context, conn *ec2.EC2, name string) (*ec2.PrefixList, error) {
	input := &ec2.DescribePrefixListsInput{
		Filters: newAttributeFilterList(map[string]string{
			"prefix-list-name": name,
		}),
	}

	return FindPrefixList(ctx, conn, input)
}

func FindVPCEndpointConnectionByServiceIDAndVPCEndpointID(ctx context.Context, conn *ec2.EC2, serviceID, vpcEndpointID string) (*ec2.VpcEndpointConnection, error) {
	input := &ec2.DescribeVpcEndpointConnectionsInput{
		Filters: newAttributeFilterList(map[string]string{
			"service-id": serviceID,
			// "InvalidFilter: The filter vpc-endpoint-id  is invalid"
			// "vpc-endpoint-id ": vpcEndpointID,
		}),
	}

	var output *ec2.VpcEndpointConnection

	err := conn.DescribeVpcEndpointConnectionsPagesWithContext(ctx, input, func(page *ec2.DescribeVpcEndpointConnectionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.VpcEndpointConnections {
			if aws.StringValue(v.VpcEndpointId) == vpcEndpointID {
				output = v

				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if vpcEndpointState := aws.StringValue(output.VpcEndpointState); vpcEndpointState == vpcEndpointStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     vpcEndpointState,
			LastRequest: input,
		}
	}

	return output, nil
}

func FindImportSnapshotTasks(ctx context.Context, conn *ec2_sdkv2.Client, input *ec2_sdkv2.DescribeImportSnapshotTasksInput) ([]awstypes.ImportSnapshotTask, error) {
	var output []awstypes.ImportSnapshotTask

	pages := ec2_sdkv2.NewDescribeImportSnapshotTasksPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			if tfawserr_sdkv2.ErrCodeEquals(err, errCodeInvalidConversionTaskIdMalformed, "not found") {
				return nil, &retry.NotFoundError{
					LastError:   err,
					LastRequest: input,
				}
			}
			return nil, err
		}

		output = append(output, page.ImportSnapshotTasks...)
	}

	return output, nil
}

func FindImportSnapshotTask(ctx context.Context, conn *ec2_sdkv2.Client, input *ec2_sdkv2.DescribeImportSnapshotTasksInput) (*awstypes.ImportSnapshotTask, error) {
	output, err := FindImportSnapshotTasks(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output, func(v *awstypes.ImportSnapshotTask) bool { return v.SnapshotTaskDetail != nil })
}

func FindImportSnapshotTaskByID(ctx context.Context, conn *ec2_sdkv2.Client, id string) (*awstypes.ImportSnapshotTask, error) {
	input := &ec2_sdkv2.DescribeImportSnapshotTasksInput{
		ImportTaskIds: []string{id},
	}

	output, err := FindImportSnapshotTask(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.ImportTaskId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindSnapshots(ctx context.Context, conn *ec2_sdkv2.Client, input *ec2_sdkv2.DescribeSnapshotsInput) ([]awstypes.Snapshot, error) {
	var output []awstypes.Snapshot

	pages := ec2_sdkv2.NewDescribeSnapshotsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			if tfawserr_sdkv2.ErrCodeEquals(err, errCodeInvalidSnapshotNotFound) {
				return nil, &retry.NotFoundError{
					LastError:   err,
					LastRequest: input,
				}
			}
			return nil, err
		}

		output = append(output, page.Snapshots...)
	}

	return output, nil
}

func FindSnapshot(ctx context.Context, conn *ec2_sdkv2.Client, input *ec2_sdkv2.DescribeSnapshotsInput) (*awstypes.Snapshot, error) {
	output, err := FindSnapshots(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func FindSnapshotByID(ctx context.Context, conn *ec2_sdkv2.Client, id string) (*awstypes.Snapshot, error) {
	input := &ec2_sdkv2.DescribeSnapshotsInput{
		SnapshotIds: []string{id},
	}

	output, err := FindSnapshot(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.SnapshotId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindSnapshotAttribute(ctx context.Context, conn *ec2_sdkv2.Client, input *ec2_sdkv2.DescribeSnapshotAttributeInput) (*ec2_sdkv2.DescribeSnapshotAttributeOutput, error) {
	output, err := conn.DescribeSnapshotAttribute(ctx, input)

	if tfawserr_sdkv2.ErrCodeEquals(err, errCodeInvalidSnapshotNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func FindCreateSnapshotCreateVolumePermissionByTwoPartKey(ctx context.Context, conn *ec2_sdkv2.Client, snapshotID, accountID string) (awstypes.CreateVolumePermission, error) {
	input := &ec2_sdkv2.DescribeSnapshotAttributeInput{
		Attribute:  awstypes.SnapshotAttributeNameCreateVolumePermission,
		SnapshotId: aws.String(snapshotID),
	}

	output, err := FindSnapshotAttribute(ctx, conn, input)

	if err != nil {
		return awstypes.CreateVolumePermission{}, err
	}

	for _, v := range output.CreateVolumePermissions {
		if aws.StringValue(v.UserId) == accountID {
			return v, nil
		}
	}

	return awstypes.CreateVolumePermission{}, &retry.NotFoundError{LastRequest: input}
}

func FindFindSnapshotTierStatuses(ctx context.Context, conn *ec2_sdkv2.Client, input *ec2_sdkv2.DescribeSnapshotTierStatusInput) ([]awstypes.SnapshotTierStatus, error) {
	var output []awstypes.SnapshotTierStatus

	pages := ec2_sdkv2.NewDescribeSnapshotTierStatusPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.SnapshotTierStatuses...)
	}

	return output, nil
}

func FindFindSnapshotTierStatus(ctx context.Context, conn *ec2_sdkv2.Client, input *ec2_sdkv2.DescribeSnapshotTierStatusInput) (*awstypes.SnapshotTierStatus, error) {
	output, err := FindFindSnapshotTierStatuses(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func FindSnapshotTierStatusBySnapshotID(ctx context.Context, conn *ec2_sdkv2.Client, id string) (*awstypes.SnapshotTierStatus, error) {
	input := &ec2_sdkv2.DescribeSnapshotTierStatusInput{
		Filters: newAttributeFilterListV2(map[string]string{
			"snapshot-id": id,
		}),
	}

	output, err := FindFindSnapshotTierStatus(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.SnapshotId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindNetworkPerformanceMetricSubscriptions(ctx context.Context, conn *ec2_sdkv2.Client, input *ec2_sdkv2.DescribeAwsNetworkPerformanceMetricSubscriptionsInput) ([]awstypes.Subscription, error) {
	var output []awstypes.Subscription

	pages := ec2_sdkv2.NewDescribeAwsNetworkPerformanceMetricSubscriptionsPaginator(conn, input, func(o *ec2_sdkv2.DescribeAwsNetworkPerformanceMetricSubscriptionsPaginatorOptions) {
		o.Limit = 100
	})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.Subscriptions...)
	}

	return output, nil
}

func FindNetworkPerformanceMetricSubscriptionByFourPartKey(ctx context.Context, conn *ec2_sdkv2.Client, source, destination, metric, statistic string) (*awstypes.Subscription, error) {
	input := &ec2_sdkv2.DescribeAwsNetworkPerformanceMetricSubscriptionsInput{}

	output, err := FindNetworkPerformanceMetricSubscriptions(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	for _, v := range output {
		if aws_sdkv2.ToString(v.Source) == source && aws_sdkv2.ToString(v.Destination) == destination && string(v.Metric) == metric && string(v.Statistic) == statistic {
			return &v, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func FindInstanceConnectEndpoint(ctx context.Context, conn *ec2_sdkv2.Client, input *ec2_sdkv2.DescribeInstanceConnectEndpointsInput) (*awstypes.Ec2InstanceConnectEndpoint, error) {
	output, err := FindInstanceConnectEndpoints(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func FindInstanceConnectEndpoints(ctx context.Context, conn *ec2_sdkv2.Client, input *ec2_sdkv2.DescribeInstanceConnectEndpointsInput) ([]awstypes.Ec2InstanceConnectEndpoint, error) {
	var output []awstypes.Ec2InstanceConnectEndpoint

	pages := ec2_sdkv2.NewDescribeInstanceConnectEndpointsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr_sdkv2.ErrCodeEquals(err, errCodeInvalidInstanceConnectEndpointIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.InstanceConnectEndpoints...)
	}

	return output, nil
}

func FindInstanceConnectEndpointByID(ctx context.Context, conn *ec2_sdkv2.Client, id string) (*awstypes.Ec2InstanceConnectEndpoint, error) {
	input := &ec2_sdkv2.DescribeInstanceConnectEndpointsInput{
		InstanceConnectEndpointIds: []string{id},
	}
	output, err := FindInstanceConnectEndpoint(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.Ec2InstanceConnectEndpointStateDeleteComplete {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws_sdkv2.ToString(output.InstanceConnectEndpointId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindVerifiedAccessGroupPolicyByID(ctx context.Context, conn *ec2_sdkv2.Client, id string) (*ec2_sdkv2.GetVerifiedAccessGroupPolicyOutput, error) {
	input := &ec2_sdkv2.GetVerifiedAccessGroupPolicyInput{
		VerifiedAccessGroupId: &id,
	}
	output, err := conn.GetVerifiedAccessGroupPolicy(ctx, input)

	if tfawserr_sdkv2.ErrCodeEquals(err, errCodeInvalidVerifiedAccessGroupIdNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func FindVerifiedAccessEndpointPolicyByID(ctx context.Context, conn *ec2_sdkv2.Client, id string) (*ec2_sdkv2.GetVerifiedAccessEndpointPolicyOutput, error) {
	input := &ec2_sdkv2.GetVerifiedAccessEndpointPolicyInput{
		VerifiedAccessEndpointId: &id,
	}
	output, err := conn.GetVerifiedAccessEndpointPolicy(ctx, input)

	if tfawserr_sdkv2.ErrCodeEquals(err, errCodeInvalidVerifiedAccessEndpointIdNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func FindVerifiedAccessGroup(ctx context.Context, conn *ec2_sdkv2.Client, input *ec2_sdkv2.DescribeVerifiedAccessGroupsInput) (*awstypes.VerifiedAccessGroup, error) {
	output, err := FindVerifiedAccessGroups(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func FindVerifiedAccessGroups(ctx context.Context, conn *ec2_sdkv2.Client, input *ec2_sdkv2.DescribeVerifiedAccessGroupsInput) ([]awstypes.VerifiedAccessGroup, error) {
	var output []awstypes.VerifiedAccessGroup

	pages := ec2_sdkv2.NewDescribeVerifiedAccessGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr_sdkv2.ErrCodeEquals(err, errCodeInvalidVerifiedAccessGroupIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.VerifiedAccessGroups...)
	}

	return output, nil
}

func FindVerifiedAccessGroupByID(ctx context.Context, conn *ec2_sdkv2.Client, id string) (*awstypes.VerifiedAccessGroup, error) {
	input := &ec2_sdkv2.DescribeVerifiedAccessGroupsInput{
		VerifiedAccessGroupIds: []string{id},
	}
	output, err := FindVerifiedAccessGroup(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws_sdkv2.ToString(output.VerifiedAccessGroupId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindVerifiedAccessInstance(ctx context.Context, conn *ec2_sdkv2.Client, input *ec2_sdkv2.DescribeVerifiedAccessInstancesInput) (*awstypes.VerifiedAccessInstance, error) {
	output, err := FindVerifiedAccessInstances(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func FindVerifiedAccessInstances(ctx context.Context, conn *ec2_sdkv2.Client, input *ec2_sdkv2.DescribeVerifiedAccessInstancesInput) ([]awstypes.VerifiedAccessInstance, error) {
	var output []awstypes.VerifiedAccessInstance

	pages := ec2_sdkv2.NewDescribeVerifiedAccessInstancesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr_sdkv2.ErrCodeEquals(err, errCodeInvalidVerifiedAccessInstanceIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.VerifiedAccessInstances...)
	}

	return output, nil
}

func FindVerifiedAccessInstanceByID(ctx context.Context, conn *ec2_sdkv2.Client, id string) (*awstypes.VerifiedAccessInstance, error) {
	input := &ec2_sdkv2.DescribeVerifiedAccessInstancesInput{
		VerifiedAccessInstanceIds: []string{id},
	}
	output, err := FindVerifiedAccessInstance(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws_sdkv2.ToString(output.VerifiedAccessInstanceId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindVerifiedAccessInstanceLoggingConfiguration(ctx context.Context, conn *ec2_sdkv2.Client, input *ec2_sdkv2.DescribeVerifiedAccessInstanceLoggingConfigurationsInput) (*awstypes.VerifiedAccessInstanceLoggingConfiguration, error) {
	output, err := FindVerifiedAccessInstanceLoggingConfigurations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func FindVerifiedAccessInstanceLoggingConfigurations(ctx context.Context, conn *ec2_sdkv2.Client, input *ec2_sdkv2.DescribeVerifiedAccessInstanceLoggingConfigurationsInput) ([]awstypes.VerifiedAccessInstanceLoggingConfiguration, error) {
	var output []awstypes.VerifiedAccessInstanceLoggingConfiguration

	pages := ec2_sdkv2.NewDescribeVerifiedAccessInstanceLoggingConfigurationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr_sdkv2.ErrCodeEquals(err, errCodeInvalidVerifiedAccessInstanceIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.LoggingConfigurations...)
	}

	return output, nil
}

func FindVerifiedAccessInstanceLoggingConfigurationByInstanceID(ctx context.Context, conn *ec2_sdkv2.Client, id string) (*awstypes.VerifiedAccessInstanceLoggingConfiguration, error) {
	input := &ec2_sdkv2.DescribeVerifiedAccessInstanceLoggingConfigurationsInput{
		VerifiedAccessInstanceIds: []string{id},
	}
	output, err := FindVerifiedAccessInstanceLoggingConfiguration(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws_sdkv2.ToString(output.VerifiedAccessInstanceId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindVerifiedAccessInstanceTrustProviderAttachmentExists(ctx context.Context, conn *ec2_sdkv2.Client, vaiID, vatpID string) error {
	output, err := FindVerifiedAccessInstanceByID(ctx, conn, vaiID)

	if err != nil {
		return err
	}

	for _, v := range output.VerifiedAccessTrustProviders {
		if aws_sdkv2.ToString(v.VerifiedAccessTrustProviderId) == vatpID {
			return nil
		}
	}

	return &retry.NotFoundError{
		LastError: fmt.Errorf("Verified Access Instance (%s) Trust Provider (%s) Attachment not found", vaiID, vatpID),
	}
}

func FindVerifiedAccessTrustProvider(ctx context.Context, conn *ec2_sdkv2.Client, input *ec2_sdkv2.DescribeVerifiedAccessTrustProvidersInput) (*awstypes.VerifiedAccessTrustProvider, error) {
	output, err := FindVerifiedAccessTrustProviders(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func FindVerifiedAccessTrustProviders(ctx context.Context, conn *ec2_sdkv2.Client, input *ec2_sdkv2.DescribeVerifiedAccessTrustProvidersInput) ([]awstypes.VerifiedAccessTrustProvider, error) {
	var output []awstypes.VerifiedAccessTrustProvider

	pages := ec2_sdkv2.NewDescribeVerifiedAccessTrustProvidersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr_sdkv2.ErrCodeEquals(err, errCodeInvalidVerifiedAccessTrustProviderIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.VerifiedAccessTrustProviders...)
	}

	return output, nil
}

func FindVerifiedAccessTrustProviderByID(ctx context.Context, conn *ec2_sdkv2.Client, id string) (*awstypes.VerifiedAccessTrustProvider, error) {
	input := &ec2_sdkv2.DescribeVerifiedAccessTrustProvidersInput{
		VerifiedAccessTrustProviderIds: []string{id},
	}
	output, err := FindVerifiedAccessTrustProvider(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws_sdkv2.ToString(output.VerifiedAccessTrustProviderId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindImageBlockPublicAccessState(ctx context.Context, conn *ec2_sdkv2.Client) (*string, error) {
	input := &ec2_sdkv2.GetImageBlockPublicAccessStateInput{}
	output, err := conn.GetImageBlockPublicAccessState(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || output.ImageBlockPublicAccessState == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ImageBlockPublicAccessState, nil
}

func FindVerifiedAccessEndpoint(ctx context.Context, conn *ec2_sdkv2.Client, input *ec2_sdkv2.DescribeVerifiedAccessEndpointsInput) (*awstypes.VerifiedAccessEndpoint, error) {
	output, err := FindVerifiedAccessEndpoints(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func FindVerifiedAccessEndpoints(ctx context.Context, conn *ec2_sdkv2.Client, input *ec2_sdkv2.DescribeVerifiedAccessEndpointsInput) ([]awstypes.VerifiedAccessEndpoint, error) {
	var output []awstypes.VerifiedAccessEndpoint

	pages := ec2_sdkv2.NewDescribeVerifiedAccessEndpointsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr_sdkv2.ErrCodeEquals(err, errCodeInvalidVerifiedAccessEndpointIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.VerifiedAccessEndpoints...)
	}

	return output, nil
}

func FindVerifiedAccessEndpointByID(ctx context.Context, conn *ec2_sdkv2.Client, id string) (*awstypes.VerifiedAccessEndpoint, error) {
	input := &ec2_sdkv2.DescribeVerifiedAccessEndpointsInput{
		VerifiedAccessEndpointIds: []string{id},
	}
	output, err := FindVerifiedAccessEndpoint(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := output.Status; status != nil && status.Code == awstypes.VerifiedAccessEndpointStatusCodeDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(status.Code),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws_sdkv2.ToString(output.VerifiedAccessEndpointId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findFastSnapshotRestore(ctx context.Context, conn *ec2_sdkv2.Client, input *ec2_sdkv2.DescribeFastSnapshotRestoresInput) (*awstypes.DescribeFastSnapshotRestoreSuccessItem, error) {
	output, err := findFastSnapshotRestores(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findFastSnapshotRestores(ctx context.Context, conn *ec2_sdkv2.Client, input *ec2_sdkv2.DescribeFastSnapshotRestoresInput) ([]awstypes.DescribeFastSnapshotRestoreSuccessItem, error) {
	var output []awstypes.DescribeFastSnapshotRestoreSuccessItem

	pages := ec2_sdkv2.NewDescribeFastSnapshotRestoresPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.FastSnapshotRestores...)
	}

	return output, nil
}

func findFastSnapshotRestoreByTwoPartKey(ctx context.Context, conn *ec2_sdkv2.Client, availabilityZone, snapshotID string) (*awstypes.DescribeFastSnapshotRestoreSuccessItem, error) {
	input := &ec2_sdkv2.DescribeFastSnapshotRestoresInput{
		Filters: newAttributeFilterListV2(map[string]string{
			"availability-zone": availabilityZone,
			"snapshot-id":       snapshotID,
		}),
	}

	output, err := findFastSnapshotRestore(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.FastSnapshotRestoreStateCodeDisabled {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	return output, nil
}
