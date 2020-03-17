package eci

import (
	"encoding/json"
	"fmt"
	"github.com/Unknwon/goconfig"
	eci "github.com/aliyun/alibabacloud-sdk/eci-20180808/golang/client"
	util "github.com/aliyun/tea-util/golang/service"
	"os"
)


var client2 *eci.Client

/**
获取配置信息
 */
func init() {
	var cfg *goconfig.ConfigFile
	config, err := goconfig.LoadConfigFile("./eci/config.conf") //加载配置文件
	if err != nil {
		fmt.Println("get config file error:", err.Error())
		os.Exit(-1)
	}
	cfg = config

	accessKey, _ = cfg.GetValue("eci_conf_test", "access_key")
	secretKey, _ = cfg.GetValue("eci_conf_test", "secret_key")
	regionId, _ = cfg.GetValue("eci_conf_test", "region_id")

	var regionInfo map[string](map[string](string));
	value, _ := cfg.GetValue("eci_conf_test", "region_info")

	json.Unmarshal([]byte(value), &regionInfo)

	zoneId = regionInfo[regionId]["zoneId"]
	securityGroupId = regionInfo[regionId]["securityGroupId"]
	vSwitchId = regionInfo[regionId]["vSwitchId"]

	fmt.Printf("init success[ access_key:%s, secret_key:%s, region_id:%s, zoneId:%s, vSwitchId:%s, securityGroupId:%s]\n",
		accessKey, secretKey, regionId, zoneId, vSwitchId, securityGroupId)

	//init eci client
	// init config
	var eci_config = new(eci.Config).SetAccessKeyId(accessKey).
		SetAccessKeySecret(secretKey).
		SetRegionId("cn-hangzhou").
		SetEndpoint("eci.aliyuncs.com").
		SetType("access_key")

	// init client
	client2, err = eci.NewClient(eci_config)
	if err != nil {
		panic(err)
	}


}

func TestEci_v2() {

	createContainerGroup_v2()


}

func createContainerGroup_v2()  {
	// init runtimeObject
	runtimeObject := new(util.RuntimeOptions).SetAutoretry(false).
		SetMaxIdleConns(3)

	// init request
	request := new(eci.CreateContainerGroupRequest)
	request.SetRegionId(regionId)
	request.SetSecurityGroupId(securityGroupId)
	request.SetVSwitchId(vSwitchId)
	request.SetContainerGroupName("eci-test")

	createContainerRequestContainers := make([]*eci.CreateContainerGroupRequestContainer, 1)

	createContainerRequestContainer := new(eci.CreateContainerGroupRequestContainer)
	createContainerRequestContainer.SetName("nginx-liu")
	createContainerRequestContainer.SetImage("nginx")
	createContainerRequestContainer.SetCpu(2.0)
	createContainerRequestContainer.SetMemory(4.0)

	createContainerRequestContainers[0] = createContainerRequestContainer

	request.SetContainer(createContainerRequestContainers)

	// call api
	resp, err := client2.CreateContainerGroup(request, runtimeObject)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(resp)
}