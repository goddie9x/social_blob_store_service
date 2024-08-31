package config

import (
	"fmt"
	"log"

	"github.com/hudl/fargo"
)

type DiscoveryServerConnect struct {
	instance fargo.Instance
	conn     fargo.EurekaConnection
}

func (d *DiscoveryServerConnect) ConnectToEurekaDiscoveryServer(appConfig *Config) {
	d.instance = fargo.Instance{
		InstanceId:        fmt.Sprintf("%s:%s:%d", appConfig.IpAddr, appConfig.EurekaAppName, appConfig.Port),
		HostName:          appConfig.HostName,
		IPAddr:            appConfig.IpAddr,
		App:               appConfig.EurekaAppName,
		VipAddress:        appConfig.EurekaAppName,
		DataCenterInfo:    fargo.DataCenterInfo{Name: fargo.MyOwn, Class: "com.netflix.appinfo.InstanceInfo$DefaultDataCenterInfo"},
		HealthCheckUrl:    fmt.Sprintf("http://%s:%d/health", "localhost", appConfig.Port),
		StatusPageUrl:     fmt.Sprintf("http://%s:%d/status", "localhost", appConfig.Port),
		HomePageUrl:       fmt.Sprintf("http://%s:%d/", "localhost", appConfig.Port),
		LeaseInfo:         fargo.LeaseInfo{RenewalIntervalInSecs: 90, DurationInSecs: 120},
		Status:            fargo.UP,
		SecurePortEnabled: false,
		Port:              appConfig.Port,
		PortEnabled:       true,
	}
	var fargoConfig fargo.Config
	fargoConfig.Eureka.ServiceUrls = []string{appConfig.EurekaDiscoveryServerUrl}
	fargoConfig.Eureka.PollIntervalSeconds = 1
	d.conn = fargo.NewConnFromConfig(fargoConfig)
	d.conn.RegisterInstance(&d.instance)
}

func (d *DiscoveryServerConnect) DeregisterFromEurekaDiscoveryServer() {
	err := d.conn.DeregisterInstance(&d.instance)
	if err != nil {
		log.Fatalf("Failed to deregister instance: %v", err)
	}
}
