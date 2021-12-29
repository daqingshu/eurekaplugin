package eurekaplugin

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chassis/go-chassis/core/registry"
	utiltags "github.com/go-chassis/go-chassis/pkg/util/tags"
	"github.com/go-chassis/openlog"
	"github.com/daqingshu/go-eureka-client/eureka"
)

const (
	// Eureka constant string
	Eureka = "eureka"
)

// Registrator struct represents file parameters
type Registrator struct {
	Name           string
	registryClient *eureka.Client
	opts           *registry.Options
}

func genSid(appID string, serviceName string, version string, env string) string {
	//sid := strings.Join([]string{appID, serviceName, version, env}, "-")
	sid := serviceName
	return sid
}

func getIpPort(ep string) (string, int) {
	ip := ""
	port := 80
	if strings.Count(ep, ":") > 0 {
		seps := strings.Split(ep, ":")
		ip = seps[0]
		port, _ = strconv.Atoi(seps[1])
	}
	return ip, port
}

func msiToInstVo(sid string, instance *registry.MicroServiceInstance) *eureka.InstanceVo {
	eps := registry.GetProtocolList(instance.EndpointsMap)
	openlog.GetLogger().Debug(fmt.Sprintf("epsmap = %v, eps=%v", instance.EndpointsMap, eps))
	u, _ := url.Parse(eps[0])
	port, _ := strconv.Atoi(u.Port())
	dataCenterInfo := &eureka.DataCenterInfo{
		Name:     "MyOwn",
		Class:    "com.netflix.appinfo.InstanceInfo$DefaultDataCenterInfo",
		Metadata: nil,
	}
	vo := &eureka.InstanceVo{
		Hostname:         u.Hostname(),
		App:              sid,
		IpAddr:           u.Hostname(),
		Status:           eureka.STATUS_UP,
		DataCenterInfo:   dataCenterInfo,
		Port:             &eureka.PositiveInt{Value: port, Enabled: false},
		SecurePort:       &eureka.PositiveInt{Value: port, Enabled: true},
		CountryId:        1,
		VipAddress:       strings.ToLower(sid),
		SecureVipAddress: strings.ToLower(sid),
	}
	return vo
}

// RegisterService register service
func (r *Registrator) RegisterService(ms *registry.MicroService) (string, error) {
	sid := genSid(ms.AppID, ms.ServiceName, ms.Version, ms.Environment)
	return sid, nil
}

// RegisterServiceInstance register service instance
func (r *Registrator) RegisterServiceInstance(sid string, instance *registry.MicroServiceInstance) (string, error) {
	api, _ := r.registryClient.Api()
	vo := msiToInstVo(sid, instance)
	instanceId, err := api.RegisterInstanceWithVo(vo)
	if err != nil {
		openlog.GetLogger().Error("register instance failed.")
		return "", err
	}
	return instanceId, nil
}

// RegisterServiceAndInstance register service and instance
func (r *Registrator) RegisterServiceAndInstance(ms *registry.MicroService, instance *registry.MicroServiceInstance) (string, string, error) {
	sid := genSid(ms.AppID, ms.ServiceName, ms.Version, ms.Environment)
	api, _ := r.registryClient.Api()
	vo := msiToInstVo(sid, instance)
	instanceId, err := api.RegisterInstanceWithVo(vo)
	if err != nil {
		openlog.GetLogger().Error("register instance failed.")
		return sid, "", err
	}
	return sid, instanceId, nil
}

// UnRegisterMicroServiceInstance unregister micro-service instances
func (r *Registrator) UnRegisterMicroServiceInstance(microServiceID, microServiceInstanceID string) error {
	api, _ := r.registryClient.Api()
	err := api.DeRegisterInstance(microServiceID, microServiceInstanceID)
	if err != nil {
		openlog.GetLogger().Error(fmt.Sprintf("unregister instance failed, microServiceID/instanceID = %s/%s.", microServiceID, microServiceInstanceID))
		return err
	}
	return nil
}

// Heartbeat check heartbeat of micro-service instance
func (r *Registrator) Heartbeat(microServiceID, microServiceInstanceID string) (bool, error) {
	api, _ := r.registryClient.Api()
	err := api.SendHeartbeat(microServiceID, microServiceInstanceID)
	if err != nil {
		openlog.GetLogger().Error(fmt.Sprintf("Heartbeat failed, microServiceID/instanceID: %s/%s. %s", microServiceID, microServiceInstanceID, err))
		return false, err
	}
	return true, nil
}

// AddDependencies add dependencies
func (r *Registrator) AddDependencies(request *registry.MicroServiceDependency) error {
	return nil
}

//AddSchemas add schema
func (r *Registrator) AddSchemas(microServiceID, schemaName, schemaInfo string) error {
	return nil
}

// UpdateMicroServiceInstanceStatus update micro-service instance status
func (r *Registrator) UpdateMicroServiceInstanceStatus(microServiceID, microServiceInstanceID, status string) error {
	api, _ := r.registryClient.Api()
	err := api.UpdateInstanceStatus(microServiceID, microServiceInstanceID, status)
	if err != nil {
		openlog.GetLogger().Error(fmt.Sprintf("UpdateMicroServiceInstanceStatus failed, microServiceID/instanceID = %s/%s, status=%s.", microServiceID, microServiceInstanceID, status))
		return err
	}
	openlog.GetLogger().Debug(fmt.Sprintf("UpdateMicroServiceInstanceStatus success, microServiceID/instanceID = %s/%s, status=%s.", microServiceID, microServiceInstanceID, status))
	return nil
}

// UpdateMicroServiceProperties update micro-service properities
func (r *Registrator) UpdateMicroServiceProperties(microServiceID string, properties map[string]string) error {
	return nil
}

// UpdateMicroServiceInstanceProperties update micro-service instance properities
func (r *Registrator) UpdateMicroServiceInstanceProperties(microServiceID, microServiceInstanceID string, properties map[string]string) error {
	api, _ := r.registryClient.Api()
	err := api.UpdateMeta(microServiceID, microServiceInstanceID, properties)
	if err != nil {
		openlog.GetLogger().Error(fmt.Sprintf("UpdateMicroServiceInstanceProperties failed, microServiceID/instanceID = %s/%s.", microServiceID, microServiceInstanceID))
		return err
	}
	openlog.GetLogger().Debug(fmt.Sprintf("UpdateMicroServiceInstanceProperties success, microServiceID/instanceID = %s/%s.", microServiceID, microServiceInstanceID))
	return nil
}

// Close close the file
func (r *Registrator) Close() error {
	return nil
}

// Discovery struct represents file service
type Discovery struct {
	Name           string
	registryClient *eureka.Client
	opts           *registry.Options
}

// GetMicroServiceID get micro-service id
func (r *Discovery) GetMicroServiceID(appID, microServiceName, version, env string) (string, error) {
	sid := genSid(appID, microServiceName, version, env)
	return sid, nil
}

// GetAllMicroServices get all microservices
func (r *Discovery) GetAllMicroServices() ([]*registry.MicroService, error) {
	var mss []*registry.MicroService
	api, _ := r.registryClient.Api()
	appVos, err := api.QueryAllInstances()
	if err != nil {
		openlog.GetLogger().Error("GetAllApplications failed: " + err.Error())
		return nil, err
	}
	for _, app := range appVos {
		mss = append(mss, &registry.MicroService{
			ServiceName: app.Name,
		})
	}
	return mss, nil
}

// GetMicroService get micro-service
func (r *Discovery) GetMicroService(microServiceID string) (*registry.MicroService, error) {
	api, _ := r.registryClient.Api()
	app, err := api.GetApplication(microServiceID)
	if err != nil {
		openlog.GetLogger().Error("GetMicroService failed: " + err.Error())
		return nil, err
	}
	return &registry.MicroService{
		AppID:       microServiceID,
		ServiceName: app.Name,
	}, nil
}


// GetMicroServiceInstances get micro-service instances
func (r *Discovery) GetMicroServiceInstances(consumerID, providerID string) ([]*registry.MicroServiceInstance, error) {
	api, _ := r.registryClient.Api()
	instanceVos, err := api.QueryAllInstanceByAppId(providerID)
	if err != nil {
		openlog.GetLogger().Error("GetMicroServiceInstances failed: " + err.Error())
		return nil, err
	}
	instances := filterInstances(instanceVos)
	return instances, nil
}


// FindMicroServiceInstances find micro-service instances
func (r *Discovery) FindMicroServiceInstances(consumerID, microServiceName string, tags utiltags.Tags) ([]*registry.MicroServiceInstance, error) {
	api, _ := r.registryClient.Api()
	providerInstances, err := api.QueryAllInstanceByAppId(microServiceName)
	if err != nil {
		return nil, fmt.Errorf("FindMicroServiceInstances failed, err: %s", err)
	}
	instances := filterInstances(providerInstances)
	return instances, nil
}

// AutoSync auto sync
func (r *Discovery) AutoSync() {
}

// Close close the file
func (r *Discovery) Close() error {
	return nil
}


// filterInstances filter instances
func filterInstances(providerInstances []eureka.InstanceVo) []*registry.MicroServiceInstance {
	instances := make([]*registry.MicroServiceInstance, 0)
	for _, ins := range providerInstances {
		msi := instanceVoToMicroServiceInstance(&ins)
		instances = append(instances, msi)
	}
	return instances
}

func instanceVoToMicroServiceInstance(instanceVo *eureka.InstanceVo) *registry.MicroServiceInstance {
	msi := &registry.MicroServiceInstance{
		InstanceID:      instanceVo.InstanceId,
		HostName:        instanceVo.Hostname,
		ServiceID:       instanceVo.App,
		DefaultProtocol: "https",
		DefaultEndpoint: instanceVo.IpAddr + ":" + strconv.Itoa(instanceVo.Port.Value),
		Status:          instanceVo.Status,
		EndpointsMap:    make(map[string]string),
		Metadata:        instanceVo.Metadata.Map,
		DataCenterInfo: &registry.DataCenterInfo{
			Name:          instanceVo.DataCenterInfo.Name,
			Region:        "",
			AvailableZone: instanceVo.DataCenterInfo.Class,
		},
	}
	return msi
}


// NewEurekaRegistry new eureka registry
func NewEurekaRegistry(options registry.Options) registry.Registrator {
	config := eureka.GetDefaultEurekaClientConfig()
	config.UseDnsForFetchingServiceUrls = false
	config.Region = "default-region-1"
	config.AvailabilityZones = map[string]string{
		config.Region: "default-zone-1",
	}
	serviceUrls := make([]string, 0)
	for _, addr := range options.Addrs {
		url := fmt.Sprintf("http://%s/eureka", addr)
		serviceUrls = append(serviceUrls, url)
	}
	openlog.GetLogger().Debug(fmt.Sprintf("eureka server addrs: %v", options.Addrs))
	config.ServiceUrl = map[string]string{
		config.AvailabilityZones[config.Region]: strings.Join(serviceUrls, ","),
	}
	openlog.GetLogger().Debug(fmt.Sprintf("eureka service url: %v", config.ServiceUrl))
	client := eureka.DefaultClient.Config(config)
	return &Registrator{
		Name:           Eureka,
		registryClient: client,
		opts:           &options,
	}
}


// NewEurekaRegistry new eureka discovery
func NewEurekaDiscovery(options registry.Options) registry.ServiceDiscovery {
	config := eureka.GetDefaultEurekaClientConfig()
	config.UseDnsForFetchingServiceUrls = false
	config.Region = "default-region-1"
	config.AvailabilityZones = map[string]string{
		config.Region: "default-zone-1",
	}
	serviceUrls := make([]string, 0)
	for _, addr := range options.Addrs {
		url := fmt.Sprintf("http://%s/eureka", addr)
		serviceUrls = append(serviceUrls, url)
	}
	openlog.GetLogger().Debug(fmt.Sprintf("eureka server addrs: %v", options.Addrs))
	config.ServiceUrl = map[string]string{
		config.AvailabilityZones[config.Region]: strings.Join(serviceUrls, ","),
	}
	openlog.GetLogger().Debug(fmt.Sprintf("eureka service url: %v", config.ServiceUrl))
	client := eureka.DefaultClient.Config(config)
	return &Discovery{
		Name:           Eureka,
		registryClient: client,
		opts:           &options,
	}
}

// Init initialize the plugin of service center registry
func Init() {
	registry.InstallRegistrator(Eureka, NewEurekaRegistry)
	registry.InstallServiceDiscovery(Eureka, NewEurekaDiscovery)
}

