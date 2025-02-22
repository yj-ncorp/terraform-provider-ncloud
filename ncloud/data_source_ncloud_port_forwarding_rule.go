package ncloud

import (
	"log"
	"strconv"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/ncloud"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/server"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func init() {
	RegisterDataSource("ncloud_port_forwarding_rule", dataSourceNcloudPortForwardingRule())
}

func dataSourceNcloudPortForwardingRule() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNcloudPortForwardingRuleRead,

		Schema: map[string]*schema.Schema{
			// Deprecated
			"internet_line_type_code": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: ToDiagFunc(validation.StringInSlice([]string{"PUBLC", "GLBL"}, false)),
				Description:      "Internet line code. PUBLC(Public), GLBL(Global)",
				Deprecated:       "This parameter is no longer used.",
			},
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Region code. Get available values using the `data ncloud_regions`.",
			},
			"zone": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Zone code. Get available values using the `data ncloud_zones`.",
			},
			"server_instance_no": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Server instance number",
			},
			"port_forwarding_internal_port": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Port forwarding internal port.",
			},
			"port_forwarding_external_port": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Port forwarding external port.",
			},
			"port_forwarding_configuration_no": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Port forwarding configuration number.",
			},
			"port_forwarding_public_ip": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Port forwarding public ip",
			},
		},
	}
}

func dataSourceNcloudPortForwardingRuleRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*ProviderConfig)
	client := config.Client

	if config.SupportVPC {
		return NotSupportVpc("data source `ncloud_port_forwarding_rule`")
	}

	regionNo, err := parseRegionNoParameter(d)
	if err != nil {
		return err
	}
	zoneNo, err := parseZoneNoParameter(config, d)
	if err != nil {
		return err
	}
	reqParams := &server.GetPortForwardingRuleListRequest{
		RegionNo: regionNo,
		ZoneNo:   zoneNo,
	}

	logCommonRequest("GetPortForwardingRuleList", reqParams)
	resp, err := client.server.V2Api.GetPortForwardingRuleList(reqParams)
	if err != nil {
		logErrorResponse("GetPortForwardingRuleList", err, reqParams)
		return err
	}
	logResponse("GetPortForwardingRuleList", resp)

	allPortForwardingRules := resp.PortForwardingRuleList
	var filteredPortForwardingRuleList []*server.PortForwardingRule
	var portForwardingRule *server.PortForwardingRule

	filterServerInstanceNo, filterServerInstanceNoOk := d.GetOk("server_instance_no")
	filterInternalPort, filterInternalPortOk := d.GetOk("port_forwarding_internal_port")
	filterExternalPort, filterExternalPortOk := d.GetOk("port_forwarding_external_port")
	if filterServerInstanceNoOk || filterInternalPortOk || filterExternalPortOk {
		for _, portForwardingRule := range allPortForwardingRules {
			if filterServerInstanceNoOk && portForwardingRule.ServerInstance != nil && ncloud.StringValue(portForwardingRule.ServerInstance.ServerInstanceNo) == filterServerInstanceNo.(string) {
				filteredPortForwardingRuleList = append(filteredPortForwardingRuleList, portForwardingRule)
			} else if filterInternalPortOk && strconv.Itoa(int(ncloud.Int32Value(portForwardingRule.PortForwardingInternalPort))) == filterInternalPort.(string) {
				filteredPortForwardingRuleList = append(filteredPortForwardingRuleList, portForwardingRule)
			} else if filterExternalPortOk && strconv.Itoa(int(ncloud.Int32Value(portForwardingRule.PortForwardingExternalPort))) == filterExternalPort.(string) {
				filteredPortForwardingRuleList = append(filteredPortForwardingRuleList, portForwardingRule)
			}
		}
	} else {
		filteredPortForwardingRuleList = allPortForwardingRules[:]
	}

	if err := validateOneResult(len(filteredPortForwardingRuleList)); err != nil {
		return err
	}
	portForwardingRule = filteredPortForwardingRuleList[0]

	return portForwardingRuleAttributes(d, resp.PortForwardingRuleList[0].PortForwardingConfigurationNo, portForwardingRule)
}

func portForwardingRuleAttributes(d *schema.ResourceData, portForwardingConfigurationNo *string, rule *server.PortForwardingRule) error {
	log.Printf("rule: %+v", rule.PortForwardingExternalPort)
	d.SetId(ncloud.StringValue(portForwardingConfigurationNo))
	d.Set("port_forwarding_configuration_no", portForwardingConfigurationNo)
	d.Set("port_forwarding_public_ip", rule.ServerInstance.PortForwardingPublicIp)
	d.Set("server_instance_no", rule.ServerInstance.ServerInstanceNo)
	d.Set("port_forwarding_external_port", strconv.Itoa(int(ncloud.Int32Value(rule.PortForwardingExternalPort))))
	d.Set("port_forwarding_internal_port", strconv.Itoa(int(ncloud.Int32Value(rule.PortForwardingInternalPort))))

	return nil
}
