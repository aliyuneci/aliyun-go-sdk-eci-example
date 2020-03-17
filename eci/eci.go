package eci

import (
	"encoding/json"
	"fmt"
	"github.com/Unknwon/goconfig"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/eci"
	"os"
	"time"
)

var accessKey string
var secretKey string
var regionId string
var zoneId string
var securityGroupId string
var vSwitchId string

var client *eci.Client

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
	client, err = eci.NewClientWithAccessKey(regionId, accessKey, secretKey)
	if err != nil {
		panic(err)
	}

}

func createContainerGroup() (string) {
	// Create Container Group
	createContainerRequest := eci.CreateCreateContainerGroupRequest()
	// required
	createContainerRequest.RegionId = regionId
	createContainerRequest.SecurityGroupId = securityGroupId
	createContainerRequest.VSwitchId = vSwitchId
	createContainerRequest.ContainerGroupName = "test-go-sdk"
	createContainerRequest.RestartPolicy = "Never"

	createContainerRequestContainer := make([]eci.CreateContainerGroupContainer, 1)
	createContainerRequestContainer[0].Image = "nginx"
	createContainerRequestContainer[0].Name = "nginx-liu"
	// option
	createContainerRequestContainer[0].Cpu = requests.NewFloat(2)
	createContainerRequestContainer[0].Memory = requests.NewFloat(4)
	createContainerRequestContainer[0].ImagePullPolicy = "IfNotPresent"

	createContainerRequest.Container = &createContainerRequestContainer

	//sdk-core默认的重试次数为3，在没有加幂等的条件下，资源创建的接口底层不需要自动重试
	client.GetConfig().MaxRetryTime = 0

	createContainerRequest.Method = "post"

	createContainerGroupResponse, err := client.CreateContainerGroup(createContainerRequest)
	if err != nil {
		panic(err)
	}

	containerGroupId := createContainerGroupResponse.ContainerGroupId

	fmt.Println(containerGroupId)

	return containerGroupId

}

func deleteContainerGroupById(containerGroupId string) {
	deleteContainerGroupRequest := eci.CreateDeleteContainerGroupRequest()
	deleteContainerGroupRequest.RegionId = regionId
	deleteContainerGroupRequest.ContainerGroupId = containerGroupId

	_, err := client.DeleteContainerGroup(deleteContainerGroupRequest)
	if err != nil {
		panic(err)
	}

	fmt.Println("DeleteContainerGroup ContainerGroupId :", containerGroupId)
}

func describeContainerGroup(containerGroupId string) (eci.DescribeContainerGroupsContainerGroup0)  {
	// Describe Container Groups
	describeContainerGroupsRequest := eci.CreateDescribeContainerGroupsRequest()
	describeContainerGroupsRequest.RegionId = regionId

	containerGroupIds := append([]string{}, containerGroupId)
	containerGroupIdsString, err := json.Marshal(containerGroupIds)
	describeContainerGroupsRequest.ContainerGroupIds = string(containerGroupIdsString)

	describeContainerGroupsResponse, err := client.DescribeContainerGroups(describeContainerGroupsRequest)
	if err != nil {
		panic(err)
	}

	describeContainerGroupNumber := len(describeContainerGroupsResponse.ContainerGroups)

	if describeContainerGroupsResponse.TotalCount != 1 && describeContainerGroupNumber != 1 {
		fmt.Println("Invalid ContainerGroups count", describeContainerGroupsResponse.TotalCount, describeContainerGroupNumber)
		panic("Invalid ContainerGroups count")
	}

	//fmt.Println("ContainerGroup status:", describeContainerGroupsResponse.ContainerGroups[0].Status, containerGroupId,)

	// container groups
	return describeContainerGroupsResponse.ContainerGroups[0]

}


func TestBatch() {
	containerGroupIds := make(chan string, 10)

	go func() {
		for i := 0; i < 1; i++ {
			containerGroupId := createContainerGroup()
			containerGroupIds <- containerGroupId
			time.Sleep(3 * time.Second)
		}
	}()

	go func() {
		for containerGroupId := range containerGroupIds {
			for i := 0; i < 20; i++ {
				status := describeContainerGroup(containerGroupId).Status
				if Running == ContainerGroupStatus(status) {
					break
				} else {
					time.Sleep(5 * time.Second)
				}
			}
			deleteContainerGroupById(containerGroupId)
		}
	}()

	//阻塞等待异步执行完，不然会提前退出。
	var input string
	fmt.Println("waiting for input to finish:")
	fmt.Scanln(&input)
	fmt.Println("test done!")
}

func TestEci()  {

	createContainerGroup()

}


