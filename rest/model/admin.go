package model

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/evergreen-ci/evergreen"
	"github.com/mongodb/grip/send"
	"github.com/pkg/errors"
)

func NewConfigModel() *APIAdminSettings {
	return &APIAdminSettings{
		Alerts:            &APIAlertsConfig{},
		Amboy:             &APIAmboyConfig{},
		Api:               &APIapiConfig{},
		AuthConfig:        &APIAuthConfig{},
		ContainerPools:    &APIContainerPoolsConfig{},
		Credentials:       map[string]string{},
		Expansions:        map[string]string{},
		HostInit:          &APIHostInitConfig{},
		Jira:              &APIJiraConfig{},
		JIRANotifications: &APIJIRANotificationsConfig{},
		Keys:              map[string]string{},
		LoggerConfig:      &APILoggerConfig{},
		Notify:            &APINotifyConfig{},
		Plugins:           map[string]map[string]interface{}{},
		Providers:         &APICloudProviders{},
		RepoTracker:       &APIRepoTrackerConfig{},
		Scheduler:         &APISchedulerConfig{},
		ServiceFlags:      &APIServiceFlags{},
		Slack:             &APISlackConfig{},
		Splunk:            &APISplunkConnectionInfo{},
		Ui:                &APIUIConfig{},
	}
}

// APIAdminSettings is the structure of a response to the admin route
type APIAdminSettings struct {
	Alerts             *APIAlertsConfig                  `json:"alerts,omitempty"`
	Amboy              *APIAmboyConfig                   `json:"amboy,omitempty"`
	Api                *APIapiConfig                     `json:"api,omitempty"`
	ApiUrl             APIString                         `json:"api_url,omitempty"`
	AuthConfig         *APIAuthConfig                    `json:"auth,omitempty"`
	Banner             APIString                         `json:"banner,omitempty"`
	BannerTheme        APIString                         `json:"banner_theme,omitempty"`
	ClientBinariesDir  APIString                         `json:"client_binaries_dir,omitempty"`
	ConfigDir          APIString                         `json:"configdir,omitempty"`
	Credentials        map[string]string                 `json:"credentials,omitempty"`
	ContainerPools     *APIContainerPoolsConfig          `json:"container_pools,omitempty"`
	Expansions         map[string]string                 `json:"expansions,omitempty"`
	GithubPRCreatorOrg APIString                         `json:"github_pr_creator_org,omitempty"`
	HostInit           *APIHostInitConfig                `json:"hostinit,omitempty"`
	Jira               *APIJiraConfig                    `json:"jira,omitempty"`
	Keys               map[string]string                 `json:"keys,omitempty"`
	LoggerConfig       *APILoggerConfig                  `json:"logger_config,omitempty"`
	LogPath            APIString                         `json:"log_path,omitempty"`
	Notify             *APINotifyConfig                  `json:"notify,omitempty"`
	Plugins            map[string]map[string]interface{} `json:"plugins,omitempty"`
	PprofPort          APIString                         `json:"pprof_port,omitempty"`
	Providers          *APICloudProviders                `json:"providers,omitempty"`
	RepoTracker        *APIRepoTrackerConfig             `json:"repotracker,omitempty"`
	Scheduler          *APISchedulerConfig               `json:"scheduler,omitempty"`
	ServiceFlags       *APIServiceFlags                  `json:"service_flags,omitempty"`
	Slack              *APISlackConfig                   `json:"slack,omitempty"`
	Splunk             *APISplunkConnectionInfo          `json:"splunk,omitempty"`
	SuperUsers         []string                          `json:"superusers,omitempty"`
	Ui                 *APIUIConfig                      `json:"ui,omitempty"`
	JIRANotifications  *APIJIRANotificationsConfig       `json:"jira_notifications,omitempty"`
}

// BuildFromService builds a model from the service layer
func (as *APIAdminSettings) BuildFromService(h interface{}) error {
	switch v := h.(type) {
	case *evergreen.Settings:
		if v == nil {
			return errors.New("evergreen settings object is nil")
		}
		apiModelReflect := reflect.ValueOf(*as)
		dbModelReflect := reflect.ValueOf(*v)
		for i := 0; i < apiModelReflect.NumField(); i++ {
			propName := apiModelReflect.Type().Field(i).Name
			val := apiModelReflect.FieldByName(propName)
			if val.IsNil() {
				continue
			}

			// check to see if this property is an API model itself
			interfaceVal := val.Interface()
			model, ok := interfaceVal.(Model)
			if !ok {
				continue
			}
			// build the sub-model from the DB model. assumes that the 2 fields are named the same thing
			if err := model.BuildFromService(dbModelReflect.FieldByName(propName).Interface()); err != nil {
				return errors.Wrapf(err, "error converting model section %s", propName)
			}
		}
		as.ApiUrl = &v.ApiUrl
		as.Banner = &v.Banner
		tmp := string(v.BannerTheme)
		as.BannerTheme = &tmp
		as.ClientBinariesDir = &v.ClientBinariesDir
		as.ConfigDir = &v.ConfigDir
		as.GithubPRCreatorOrg = &v.GithubPRCreatorOrg
		as.LogPath = &v.LogPath
		as.Plugins = v.Plugins
		as.PprofPort = &v.PprofPort
		as.Credentials = v.Credentials
		as.Expansions = v.Expansions
		as.Keys = v.Keys
		as.SuperUsers = v.SuperUsers
	default:
		return errors.Errorf("%T is not a supported admin settings type", h)
	}
	return nil
}

// ToService returns a service model from an API model
func (as *APIAdminSettings) ToService() (interface{}, error) {
	settings := evergreen.Settings{
		Credentials: map[string]string{},
		Expansions:  map[string]string{},
		Keys:        map[string]string{},
		Plugins:     evergreen.PluginConfig{},
		SuperUsers:  as.SuperUsers,
	}
	if as.ApiUrl != nil {
		settings.ApiUrl = *as.ApiUrl
	}
	if as.Banner != nil {
		settings.Banner = *as.Banner
	}
	if as.BannerTheme != nil {
		settings.BannerTheme = evergreen.BannerTheme(*as.BannerTheme)
	}
	if as.ClientBinariesDir != nil {
		settings.ClientBinariesDir = *as.ClientBinariesDir
	}
	if as.ConfigDir != nil {
		settings.ConfigDir = *as.ConfigDir
	}
	if as.GithubPRCreatorOrg != nil {
		settings.GithubPRCreatorOrg = *as.GithubPRCreatorOrg
	}
	if as.LogPath != nil {
		settings.LogPath = *as.LogPath
	}
	if as.PprofPort != nil {
		settings.PprofPort = *as.PprofPort
	}
	apiModelReflect := reflect.ValueOf(*as)
	dbModelReflect := reflect.ValueOf(&settings).Elem()
	for i := 0; i < apiModelReflect.NumField(); i++ {
		propName := apiModelReflect.Type().Field(i).Name
		val := apiModelReflect.FieldByName(propName)
		if val.IsNil() {
			continue
		}

		// check to see if this property is an API model itself
		interfaceVal := val.Interface()
		model, ok := interfaceVal.(Model)
		if !ok {
			continue
		}
		// set the corresponding DB model field. assumes that the 2 fields are named the same thing
		i, err := model.ToService()
		if err != nil {
			return nil, errors.Wrapf(err, "error converting model section %s", propName)
		}
		valToSet := reflect.ValueOf(i)
		dbModelReflect.FieldByName(propName).Set(valToSet)
	}
	for k, v := range as.Credentials {
		settings.Credentials[k] = v
	}
	for k, v := range as.Expansions {
		settings.Expansions[k] = v
	}
	for k, v := range as.Keys {
		settings.Keys[k] = v
	}
	for k, v := range as.Plugins {
		settings.Plugins[k] = map[string]interface{}{}
		for k2, v2 := range v {
			settings.Plugins[k][k2] = v2
		}
	}
	return settings, nil
}

type APIAlertsConfig struct {
	SMTP APISMTPConfig `json:"smtp"`
}

func (a *APIAlertsConfig) BuildFromService(h interface{}) error {
	switch v := h.(type) {
	case evergreen.AlertsConfig:
		if err := a.SMTP.BuildFromService(v.SMTP); err != nil {
			return err
		}
	default:
		return errors.Errorf("%T is not a supported type", h)
	}
	return nil
}

func (a *APIAlertsConfig) ToService() (interface{}, error) {
	smtp, err := a.SMTP.ToService()
	if err != nil {
		return nil, err
	}
	return evergreen.AlertsConfig{
		SMTP: smtp.(evergreen.SMTPConfig),
	}, nil
}

type APISMTPConfig struct {
	Server     APIString   `json:"server"`
	Port       int         `json:"port"`
	UseSSL     bool        `json:"use_ssl"`
	Username   APIString   `json:"username"`
	Password   APIString   `json:"password"`
	From       APIString   `json:"from"`
	AdminEmail []APIString `json:"admin_email"`
}

func (a *APISMTPConfig) BuildFromService(h interface{}) error {
	switch v := h.(type) {
	case evergreen.SMTPConfig:
		a.Server = ToAPIString(v.Server)
		a.Port = v.Port
		a.UseSSL = v.UseSSL
		a.Username = ToAPIString(v.Username)
		a.Password = ToAPIString(v.Password)
		a.From = ToAPIString(v.From)
		for _, s := range v.AdminEmail {
			a.AdminEmail = append(a.AdminEmail, ToAPIString(s))
		}
	default:
		return errors.Errorf("%T is not a supported type", h)
	}
	return nil
}

func (a *APISMTPConfig) ToService() (interface{}, error) {
	if a == nil {
		return nil, nil
	}
	config := evergreen.SMTPConfig{
		Server:   FromAPIString(a.Server),
		Port:     a.Port,
		UseSSL:   a.UseSSL,
		Username: FromAPIString(a.Username),
		Password: FromAPIString(a.Password),
		From:     FromAPIString(a.From),
	}
	for _, s := range a.AdminEmail {
		config.AdminEmail = append(config.AdminEmail, FromAPIString(s))
	}
	return config, nil
}

type APIAmboyConfig struct {
	Name           APIString `json:"name"`
	DB             APIString `json:"database"`
	PoolSizeLocal  int       `json:"pool_size_local"`
	PoolSizeRemote int       `json:"pool_size_remote"`
	LocalStorage   int       `json:"local_storage_size"`
}

func (a *APIAmboyConfig) BuildFromService(h interface{}) error {
	switch v := h.(type) {
	case evergreen.AmboyConfig:
		a.Name = ToAPIString(v.Name)
		a.DB = ToAPIString(v.DB)
		a.PoolSizeLocal = v.PoolSizeLocal
		a.PoolSizeRemote = v.PoolSizeRemote
		a.LocalStorage = v.LocalStorage
	default:
		return errors.Errorf("%T is not a supported type", h)
	}
	return nil
}

func (a *APIAmboyConfig) ToService() (interface{}, error) {
	return evergreen.AmboyConfig{
		Name:           FromAPIString(a.Name),
		DB:             FromAPIString(a.DB),
		PoolSizeLocal:  a.PoolSizeLocal,
		PoolSizeRemote: a.PoolSizeRemote,
		LocalStorage:   a.LocalStorage,
	}, nil
}

type APIapiConfig struct {
	HttpListenAddr      APIString `json:"http_listen_addr"`
	GithubWebhookSecret APIString `json:"github_webhook_secret"`
}

func (a *APIapiConfig) BuildFromService(h interface{}) error {
	switch v := h.(type) {
	case evergreen.APIConfig:
		a.HttpListenAddr = ToAPIString(v.HttpListenAddr)
		a.GithubWebhookSecret = ToAPIString(v.GithubWebhookSecret)
	default:
		return errors.Errorf("%T is not a supported type", h)
	}
	return nil
}

func (a *APIapiConfig) ToService() (interface{}, error) {
	return evergreen.APIConfig{
		HttpListenAddr:      FromAPIString(a.HttpListenAddr),
		GithubWebhookSecret: FromAPIString(a.GithubWebhookSecret),
	}, nil
}

type APIAuthConfig struct {
	Crowd  *APICrowdConfig      `json:"crowd"`
	Naive  *APINaiveAuthConfig  `json:"naive"`
	Github *APIGithubAuthConfig `json:"github"`
}

func (a *APIAuthConfig) BuildFromService(h interface{}) error {
	switch v := h.(type) {
	case evergreen.AuthConfig:
		if v.Crowd != nil {
			a.Crowd = &APICrowdConfig{}
			if err := a.Crowd.BuildFromService(v.Crowd); err != nil {
				return err
			}
		}
		if v.Github != nil {
			a.Github = &APIGithubAuthConfig{}
			if err := a.Github.BuildFromService(v.Github); err != nil {
				return err
			}
		}
		if v.Naive != nil {
			a.Naive = &APINaiveAuthConfig{}
			if err := a.Naive.BuildFromService(v.Naive); err != nil {
				return err
			}
		}
	default:
		return errors.Errorf("%T is not a supported type", h)
	}
	return nil
}

func (a *APIAuthConfig) ToService() (interface{}, error) {
	var crowd *evergreen.CrowdConfig
	var naive *evergreen.NaiveAuthConfig
	var github *evergreen.GithubAuthConfig
	i, err := a.Crowd.ToService()
	if err != nil {
		return nil, err
	}
	if i != nil {
		crowd = i.(*evergreen.CrowdConfig)
	}
	i, err = a.Naive.ToService()
	if err != nil {
		return nil, err
	}
	if i != nil {
		naive = i.(*evergreen.NaiveAuthConfig)
	}
	i, err = a.Github.ToService()
	if err != nil {
		return nil, err
	}
	if i != nil {
		github = i.(*evergreen.GithubAuthConfig)
	}
	return evergreen.AuthConfig{
		Crowd:  crowd,
		Naive:  naive,
		Github: github,
	}, nil
}

type APICrowdConfig struct {
	Username APIString `json:"username"`
	Password APIString `json:"password"`
	Urlroot  APIString `json:"url_root"`
}

func (a *APICrowdConfig) BuildFromService(h interface{}) error {
	switch v := h.(type) {
	case *evergreen.CrowdConfig:
		if v == nil {
			return nil
		}
		a.Username = ToAPIString(v.Username)
		a.Password = ToAPIString(v.Password)
		a.Urlroot = ToAPIString(v.Urlroot)
	default:
		return errors.Errorf("%T is not a supported type", h)
	}
	return nil
}

func (a *APICrowdConfig) ToService() (interface{}, error) {
	if a == nil {
		return nil, nil
	}
	return &evergreen.CrowdConfig{
		Username: FromAPIString(a.Username),
		Password: FromAPIString(a.Password),
		Urlroot:  FromAPIString(a.Urlroot),
	}, nil
}

type APINaiveAuthConfig struct {
	Users []*APIAuthUser `json:"users"`
}

func (a *APINaiveAuthConfig) BuildFromService(h interface{}) error {
	switch v := h.(type) {
	case *evergreen.NaiveAuthConfig:
		if v == nil {
			return nil
		}
		for _, u := range v.Users {
			APIuser := &APIAuthUser{}
			if err := APIuser.BuildFromService(u); err != nil {
				return err
			}
			a.Users = append(a.Users, APIuser)
		}
	default:
		return errors.Errorf("%T is not a supported type", h)
	}
	return nil
}

func (a *APINaiveAuthConfig) ToService() (interface{}, error) {
	if a == nil {
		return nil, nil
	}
	config := evergreen.NaiveAuthConfig{}
	for _, u := range a.Users {
		i, err := u.ToService()
		if err != nil {
			return nil, err
		}
		user := i.(*evergreen.AuthUser)
		config.Users = append(config.Users, user)
	}
	return &config, nil
}

type APIAuthUser struct {
	Username    APIString `json:"username"`
	DisplayName APIString `json:"display_name"`
	Password    APIString `json:"password"`
	Email       APIString `json:"email"`
}

func (a *APIAuthUser) BuildFromService(h interface{}) error {
	switch v := h.(type) {
	case *evergreen.AuthUser:
		if v == nil {
			return nil
		}
		a.Username = ToAPIString(v.Username)
		a.Password = ToAPIString(v.Password)
		a.DisplayName = ToAPIString(v.DisplayName)
		a.Email = ToAPIString(v.Email)
	default:
		return errors.Errorf("%T is not a supported type", h)
	}
	return nil
}

func (a *APIAuthUser) ToService() (interface{}, error) {
	if a == nil {
		return nil, nil
	}
	return &evergreen.AuthUser{
		Username:    FromAPIString(a.Username),
		Password:    FromAPIString(a.Password),
		DisplayName: FromAPIString(a.DisplayName),
		Email:       FromAPIString(a.Email),
	}, nil
}

type APIGithubAuthConfig struct {
	ClientId     APIString   `json:"client_id"`
	ClientSecret APIString   `json:"client_secret"`
	Users        []APIString `json:"users"`
	Organization APIString   `json:"organization"`
}

func (a *APIGithubAuthConfig) BuildFromService(h interface{}) error {
	switch v := h.(type) {
	case *evergreen.GithubAuthConfig:
		if v == nil {
			return nil
		}
		a.ClientId = ToAPIString(v.ClientId)
		a.ClientSecret = ToAPIString(v.ClientSecret)
		a.Organization = ToAPIString(v.Organization)
		for _, u := range v.Users {
			a.Users = append(a.Users, ToAPIString(u))
		}
	default:
		return errors.Errorf("%T is not a supported type", h)
	}
	return nil
}

func (a *APIGithubAuthConfig) ToService() (interface{}, error) {
	if a == nil {
		return nil, nil
	}
	config := evergreen.GithubAuthConfig{
		ClientId:     FromAPIString(a.ClientId),
		ClientSecret: FromAPIString(a.ClientSecret),
		Organization: FromAPIString(a.Organization),
	}
	for _, u := range a.Users {
		config.Users = append(config.Users, FromAPIString(u))
	}
	return &config, nil
}

// APIBanner is a public structure representing the banner part of the admin settings
type APIBanner struct {
	Text  APIString `json:"banner"`
	Theme APIString `json:"theme"`
}

type APIHostInitConfig struct {
	SSHTimeoutSeconds int64 `json:"ssh_timeout_secs"`
}

func (a *APIHostInitConfig) BuildFromService(h interface{}) error {
	switch v := h.(type) {
	case evergreen.HostInitConfig:
		a.SSHTimeoutSeconds = v.SSHTimeoutSeconds
	default:
		return errors.Errorf("%T is not a supported type", h)
	}
	return nil
}

func (a *APIHostInitConfig) ToService() (interface{}, error) {
	return evergreen.HostInitConfig{
		SSHTimeoutSeconds: a.SSHTimeoutSeconds,
	}, nil
}

type APIJiraConfig struct {
	Host           APIString `json:"host"`
	Username       APIString `json:"username"`
	Password       APIString `json:"password"`
	DefaultProject APIString `json:"default_project"`
}

func (a *APIJiraConfig) BuildFromService(h interface{}) error {
	switch v := h.(type) {
	case evergreen.JiraConfig:
		a.Host = ToAPIString(v.Host)
		a.Username = ToAPIString(v.Username)
		a.Password = ToAPIString(v.Password)
		a.DefaultProject = ToAPIString(v.DefaultProject)
	default:
		return errors.Errorf("%T is not a supported type", h)
	}
	return nil
}

func (a *APIJiraConfig) ToService() (interface{}, error) {
	return evergreen.JiraConfig{
		Host:           FromAPIString(a.Host),
		Username:       FromAPIString(a.Username),
		Password:       FromAPIString(a.Password),
		DefaultProject: FromAPIString(a.DefaultProject),
	}, nil
}

type APILoggerConfig struct {
	Buffer         *APILogBuffering `json:"buffer"`
	DefaultLevel   APIString        `json:"default_level"`
	ThresholdLevel APIString        `json:"threshold_level"`
}

func (a *APILoggerConfig) BuildFromService(h interface{}) error {
	switch v := h.(type) {
	case evergreen.LoggerConfig:
		a.DefaultLevel = ToAPIString(v.DefaultLevel)
		a.ThresholdLevel = ToAPIString(v.ThresholdLevel)
		a.Buffer = &APILogBuffering{}
		if err := a.Buffer.BuildFromService(v.Buffer); err != nil {
			return err
		}
	default:
		return errors.Errorf("%T is not a supported type", h)
	}
	return nil
}

func (a *APILoggerConfig) ToService() (interface{}, error) {
	config := evergreen.LoggerConfig{
		DefaultLevel:   FromAPIString(a.DefaultLevel),
		ThresholdLevel: FromAPIString(a.ThresholdLevel),
	}
	i, err := a.Buffer.ToService()
	if err != nil {
		return nil, err
	}
	buffer := i.(evergreen.LogBuffering)
	config.Buffer = buffer
	return config, nil
}

type APILogBuffering struct {
	DurationSeconds int `json:"duration_seconds"`
	Count           int `json:"count"`
}

func (a *APILogBuffering) BuildFromService(h interface{}) error {
	switch v := h.(type) {
	case evergreen.LogBuffering:
		a.DurationSeconds = v.DurationSeconds
		a.Count = v.Count
	default:
		return errors.Errorf("%T is not a supported type", h)
	}
	return nil
}

func (a *APILogBuffering) ToService() (interface{}, error) {
	return evergreen.LogBuffering{
		DurationSeconds: a.DurationSeconds,
		Count:           a.Count,
	}, nil
}

type APINotifyConfig struct {
	BufferTargetPerInterval int           `json:"buffer_target_per_interval"`
	BufferIntervalSeconds   int           `json:"buffer_interval_seconds"`
	SMTP                    APISMTPConfig `json:"smtp"`
}

func (a *APINotifyConfig) BuildFromService(h interface{}) error {
	switch v := h.(type) {
	case evergreen.NotifyConfig:
		a.SMTP = APISMTPConfig{}
		if err := a.SMTP.BuildFromService(v.SMTP); err != nil {
			return err
		}
		a.BufferTargetPerInterval = v.BufferTargetPerInterval
		a.BufferIntervalSeconds = v.BufferIntervalSeconds
	default:
		return errors.Errorf("%T is not a supported type", h)
	}
	return nil
}

func (a *APINotifyConfig) ToService() (interface{}, error) {
	smtp, err := a.SMTP.ToService()
	if err != nil {
		return nil, err
	}
	return evergreen.NotifyConfig{
		BufferTargetPerInterval: a.BufferTargetPerInterval,
		BufferIntervalSeconds:   a.BufferIntervalSeconds,
		SMTP: smtp.(evergreen.SMTPConfig),
	}, nil
}

type APICloudProviders struct {
	AWS       *APIAWSConfig       `json:"aws"`
	Docker    *APIDockerConfig    `json:"docker"`
	GCE       *APIGCEConfig       `json:"gce"`
	OpenStack *APIOpenStackConfig `json:"openstack"`
	VSphere   *APIVSphereConfig   `json:"vsphere"`
}

func (a *APICloudProviders) BuildFromService(h interface{}) error {
	switch v := h.(type) {
	case evergreen.CloudProviders:
		a.AWS = &APIAWSConfig{}
		a.Docker = &APIDockerConfig{}
		a.GCE = &APIGCEConfig{}
		a.OpenStack = &APIOpenStackConfig{}
		a.VSphere = &APIVSphereConfig{}
		if err := a.AWS.BuildFromService(v.AWS); err != nil {
			return err
		}
		if err := a.Docker.BuildFromService(v.Docker); err != nil {
			return err
		}
		if err := a.GCE.BuildFromService(v.GCE); err != nil {
			return err
		}
		if err := a.OpenStack.BuildFromService(v.OpenStack); err != nil {
			return err
		}
		if err := a.VSphere.BuildFromService(v.VSphere); err != nil {
			return err
		}
	default:
		return errors.Errorf("%T is not a supported type", h)
	}
	return nil
}

func (a *APICloudProviders) ToService() (interface{}, error) {
	aws, err := a.AWS.ToService()
	if err != nil {
		return nil, err
	}
	docker, err := a.Docker.ToService()
	if err != nil {
		return nil, err
	}
	gce, err := a.GCE.ToService()
	if err != nil {
		return nil, err
	}
	openstack, err := a.OpenStack.ToService()
	if err != nil {
		return nil, err
	}
	vsphere, err := a.VSphere.ToService()
	if err != nil {
		return nil, err
	}
	return evergreen.CloudProviders{
		AWS:       aws.(evergreen.AWSConfig),
		Docker:    docker.(evergreen.DockerConfig),
		GCE:       gce.(evergreen.GCEConfig),
		OpenStack: openstack.(evergreen.OpenStackConfig),
		VSphere:   vsphere.(evergreen.VSphereConfig),
	}, nil
}

type APIContainerPoolsConfig struct {
	Pools []APIContainerPool `json:"pools"`
}

func (a *APIContainerPoolsConfig) BuildFromService(h interface{}) error {
	switch v := h.(type) {
	case evergreen.ContainerPoolsConfig:
		for _, pool := range v.Pools {
			APIpool := APIContainerPool{}
			if err := APIpool.BuildFromService(pool); err != nil {
				return err
			}
			a.Pools = append(a.Pools, APIpool)
		}
	default:
		return errors.Errorf("%T is not a supported type", h)
	}
	return nil
}

func (a *APIContainerPoolsConfig) ToService() (interface{}, error) {
	if a == nil {
		return nil, nil
	}
	config := evergreen.ContainerPoolsConfig{}
	for _, p := range a.Pools {
		i, err := p.ToService()
		if err != nil {
			return nil, err
		}
		pool := i.(evergreen.ContainerPool)
		config.Pools = append(config.Pools, pool)
	}
	return config, nil
}

type APIContainerPool struct {
	Distro        APIString `json:"distro"`
	Id            APIString `json:"id"`
	MaxContainers int       `json:"max_containers"`
	Port          uint16    `json:"port"`
}

func (a *APIContainerPool) BuildFromService(h interface{}) error {
	switch v := h.(type) {
	case evergreen.ContainerPool:
		a.Distro = ToAPIString(v.Distro)
		a.Id = ToAPIString(v.Id)
		a.MaxContainers = v.MaxContainers
		a.Port = v.Port
	default:
		return errors.Errorf("%T is not a supported type", h)
	}
	return nil
}

func (a *APIContainerPool) ToService() (interface{}, error) {
	return evergreen.ContainerPool{
		Distro:        FromAPIString(a.Distro),
		Id:            FromAPIString(a.Id),
		MaxContainers: a.MaxContainers,
		Port:          a.Port,
	}, nil
}

type APIAWSConfig struct {
	Secret APIString `json:"aws_secret"`
	Id     APIString `json:"aws_id"`
}

func (a *APIAWSConfig) BuildFromService(h interface{}) error {
	switch v := h.(type) {
	case evergreen.AWSConfig:
		a.Secret = ToAPIString(v.Secret)
		a.Id = ToAPIString(v.Id)
	default:
		return errors.Errorf("%T is not a supported type", h)
	}
	return nil
}

func (a *APIAWSConfig) ToService() (interface{}, error) {
	return evergreen.AWSConfig{
		Id:     FromAPIString(a.Id),
		Secret: FromAPIString(a.Secret),
	}, nil
}

type APIDockerConfig struct {
	APIVersion APIString `json:"api_version"`
}

func (a *APIDockerConfig) BuildFromService(h interface{}) error {
	switch v := h.(type) {
	case evergreen.DockerConfig:
		a.APIVersion = ToAPIString(v.APIVersion)
	default:
		return errors.Errorf("%T is not a supported type", h)
	}
	return nil
}

func (a *APIDockerConfig) ToService() (interface{}, error) {
	return evergreen.DockerConfig{
		APIVersion: FromAPIString(a.APIVersion),
	}, nil
}

type APIGCEConfig struct {
	ClientEmail  APIString `json:"client_email"`
	PrivateKey   APIString `json:"private_key"`
	PrivateKeyID APIString `json:"private_key_id"`
	TokenURI     APIString `json:"token_uri"`
}

func (a *APIGCEConfig) BuildFromService(h interface{}) error {
	switch v := h.(type) {
	case evergreen.GCEConfig:
		a.ClientEmail = ToAPIString(v.ClientEmail)
		a.PrivateKey = ToAPIString(v.PrivateKey)
		a.PrivateKeyID = ToAPIString(v.PrivateKeyID)
		a.TokenURI = ToAPIString(v.TokenURI)
	default:
		return errors.Errorf("%T is not a supported type", h)
	}
	return nil
}

func (a *APIGCEConfig) ToService() (interface{}, error) {
	return evergreen.GCEConfig{
		ClientEmail:  FromAPIString(a.ClientEmail),
		PrivateKey:   FromAPIString(a.PrivateKey),
		PrivateKeyID: FromAPIString(a.PrivateKeyID),
		TokenURI:     FromAPIString(a.TokenURI),
	}, nil
}

type APIOpenStackConfig struct {
	IdentityEndpoint APIString `json:"identity_endpoint"`

	Username   APIString `json:"username"`
	Password   APIString `json:"password"`
	DomainName APIString `json:"domain_name"`

	ProjectName APIString `json:"project_name"`
	ProjectID   APIString `json:"project_id"`

	Region APIString `json:"region"`
}

func (a *APIOpenStackConfig) BuildFromService(h interface{}) error {
	switch v := h.(type) {
	case evergreen.OpenStackConfig:
		a.IdentityEndpoint = ToAPIString(v.IdentityEndpoint)
		a.Username = ToAPIString(v.Username)
		a.Password = ToAPIString(v.Password)
		a.DomainName = ToAPIString(v.DomainName)
		a.ProjectName = ToAPIString(v.ProjectName)
		a.ProjectID = ToAPIString(v.ProjectID)
		a.Region = ToAPIString(v.Region)
	default:
		return errors.Errorf("%T is not a supported type", h)
	}
	return nil
}

func (a *APIOpenStackConfig) ToService() (interface{}, error) {
	return evergreen.OpenStackConfig{
		IdentityEndpoint: FromAPIString(a.IdentityEndpoint),
		Username:         FromAPIString(a.Username),
		Password:         FromAPIString(a.Password),
		DomainName:       FromAPIString(a.DomainName),
		ProjectID:        FromAPIString(a.ProjectID),
		ProjectName:      FromAPIString(a.ProjectName),
		Region:           FromAPIString(a.Region),
	}, nil
}

type APIVSphereConfig struct {
	Host     APIString `json:"host"`
	Username APIString `json:"username"`
	Password APIString `json:"password"`
}

func (a *APIVSphereConfig) BuildFromService(h interface{}) error {
	switch v := h.(type) {
	case evergreen.VSphereConfig:
		a.Host = ToAPIString(v.Host)
		a.Username = ToAPIString(v.Username)
		a.Password = ToAPIString(v.Password)
	default:
		return errors.Errorf("%T is not a supported type", h)
	}
	return nil
}

func (a *APIVSphereConfig) ToService() (interface{}, error) {
	return evergreen.VSphereConfig{
		Host:     FromAPIString(a.Host),
		Username: FromAPIString(a.Username),
		Password: FromAPIString(a.Password),
	}, nil
}

type APIRepoTrackerConfig struct {
	NumNewRepoRevisionsToFetch int `json:"revs_to_fetch"`
	MaxRepoRevisionsToSearch   int `json:"max_revs_to_search"`
	MaxConcurrentRequests      int `json:"max_con_requests"`
}

func (a *APIRepoTrackerConfig) BuildFromService(h interface{}) error {
	switch v := h.(type) {
	case evergreen.RepoTrackerConfig:
		a.NumNewRepoRevisionsToFetch = v.NumNewRepoRevisionsToFetch
		a.MaxConcurrentRequests = v.MaxConcurrentRequests
		a.MaxRepoRevisionsToSearch = v.MaxRepoRevisionsToSearch
	default:
		return errors.Errorf("%T is not a supported type", h)
	}
	return nil
}

func (a *APIRepoTrackerConfig) ToService() (interface{}, error) {
	return evergreen.RepoTrackerConfig{
		NumNewRepoRevisionsToFetch: a.NumNewRepoRevisionsToFetch,
		MaxConcurrentRequests:      a.MaxConcurrentRequests,
		MaxRepoRevisionsToSearch:   a.MaxRepoRevisionsToSearch,
	}, nil
}

type APISchedulerConfig struct {
	TaskFinder       APIString `json:"task_finder"`
	HostAllocator    APIString `json:"host_allocator"`
	FreeHostFraction float64   `json:"free_host_fraction"`
}

func (a *APISchedulerConfig) BuildFromService(h interface{}) error {
	switch v := h.(type) {
	case evergreen.SchedulerConfig:
		a.TaskFinder = ToAPIString(v.TaskFinder)
		a.HostAllocator = ToAPIString(v.HostAllocator)
		a.FreeHostFraction = v.FreeHostFraction
	default:
		return errors.Errorf("%T is not a supported type", h)
	}
	return nil
}

func (a *APISchedulerConfig) ToService() (interface{}, error) {
	return evergreen.SchedulerConfig{
		TaskFinder:       FromAPIString(a.TaskFinder),
		HostAllocator:    FromAPIString(a.HostAllocator),
		FreeHostFraction: a.FreeHostFraction,
	}, nil
}

// APIServiceFlags is a public structure representing the admin service flags
type APIServiceFlags struct {
	TaskDispatchDisabled         bool `json:"task_dispatch_disabled"`
	HostinitDisabled             bool `json:"hostinit_disabled"`
	MonitorDisabled              bool `json:"monitor_disabled"`
	AlertsDisabled               bool `json:"alerts_disabled"`
	TaskrunnerDisabled           bool `json:"taskrunner_disabled"`
	RepotrackerDisabled          bool `json:"repotracker_disabled"`
	SchedulerDisabled            bool `json:"scheduler_disabled"`
	GithubPRTestingDisabled      bool `json:"github_pr_testing_disabled"`
	RepotrackerPushEventDisabled bool `json:"repotracker_push_event_disabled"`
	CLIUpdatesDisabled           bool `json:"cli_updates_disabled"`
	BackgroundStatsDisabled      bool `json:"background_stats_disabled"`
	TaskLoggingDisabled          bool `json:"task_logging_disabled"`

	// Notifications Flags
	EventProcessingDisabled      bool `json:"event_processing_disabled"`
	JIRANotificationsDisabled    bool `json:"jira_notifications_disabled"`
	SlackNotificationsDisabled   bool `json:"slack_notifications_disabled"`
	EmailNotificationsDisabled   bool `json:"email_notifications_disabled"`
	WebhookNotificationsDisabled bool `json:"webhook_notifications_disabled"`
	GithubStatusAPIDisabled      bool `json:"github_status_api_disabled"`
}

type APISlackConfig struct {
	Options *APISlackOptions `json:"options"`
	Token   APIString        `json:"token"`
	Level   APIString        `json:"level"`
}

func (a *APISlackConfig) BuildFromService(h interface{}) error {
	switch v := h.(type) {
	case evergreen.SlackConfig:
		a.Token = ToAPIString(v.Token)
		a.Level = ToAPIString(v.Level)
		if v.Options != nil {
			a.Options = &APISlackOptions{}
			if err := a.Options.BuildFromService(*v.Options); err != nil { //nolint: vet
				return err
			}
		}
	default:
		return errors.Errorf("%T is not a supported type", h)
	}
	return nil
}

func (a *APISlackConfig) ToService() (interface{}, error) {
	i, err := a.Options.ToService()
	if err != nil {
		return nil, err
	}
	options := i.(send.SlackOptions) //nolint: vet
	return evergreen.SlackConfig{
		Token:   FromAPIString(a.Token),
		Level:   FromAPIString(a.Level),
		Options: &options,
	}, nil
}

type APISlackOptions struct {
	Channel       APIString       `json:"channel"`
	Hostname      APIString       `json:"hostname"`
	Name          APIString       `json:"name"`
	Username      APIString       `json:"username"`
	IconURL       APIString       `json:"icon_url"`
	BasicMetadata bool            `json:"add_basic_metadata"`
	Fields        bool            `json:"use_fields"`
	AllFields     bool            `json:"all_fields"`
	FieldsSet     map[string]bool `json:"fields"`
}

func (a *APISlackOptions) BuildFromService(h interface{}) error {
	switch v := h.(type) {
	case send.SlackOptions:
		a.Channel = ToAPIString(v.Channel)
		a.Hostname = ToAPIString(v.Hostname)
		a.Name = ToAPIString(v.Name)
		a.Username = ToAPIString(v.Username)
		a.IconURL = ToAPIString(v.IconURL)
		a.BasicMetadata = v.BasicMetadata
		a.Fields = v.Fields
		a.AllFields = v.AllFields
		a.FieldsSet = v.FieldsSet
	default:
		return errors.Errorf("%T is not a supported type", h)
	}
	return nil
}

func (a *APISlackOptions) ToService() (interface{}, error) {
	if a == nil {
		return send.SlackOptions{}, nil
	}
	return send.SlackOptions{
		Channel:       FromAPIString(a.Channel),
		Hostname:      FromAPIString(a.Hostname),
		Name:          FromAPIString(a.Name),
		Username:      FromAPIString(a.Username),
		IconURL:       FromAPIString(a.IconURL),
		BasicMetadata: a.BasicMetadata,
		Fields:        a.Fields,
		AllFields:     a.AllFields,
		FieldsSet:     a.FieldsSet,
	}, nil
}

type APISplunkConnectionInfo struct {
	ServerURL APIString `json:"url"`
	Token     APIString `json:"token"`
	Channel   APIString `json:"channel"`
}

func (a *APISplunkConnectionInfo) BuildFromService(h interface{}) error {
	switch v := h.(type) {
	case send.SplunkConnectionInfo:
		a.ServerURL = ToAPIString(v.ServerURL)
		a.Token = ToAPIString(v.Token)
		a.Channel = ToAPIString(v.Channel)
	default:
		return errors.Errorf("%T is not a supported type", h)
	}
	return nil
}

func (a *APISplunkConnectionInfo) ToService() (interface{}, error) {
	return send.SplunkConnectionInfo{
		ServerURL: FromAPIString(a.ServerURL),
		Token:     FromAPIString(a.Token),
		Channel:   FromAPIString(a.Channel),
	}, nil
}

type APIUIConfig struct {
	Url            APIString `json:"url"`
	HelpUrl        APIString `json:"help_url"`
	HttpListenAddr APIString `json:"http_listen_addr"`
	Secret         APIString `json:"secret"`
	DefaultProject APIString `json:"default_project"`
	CacheTemplates bool      `json:"cache_templates"`
	SecureCookies  bool      `json:"secure_cookies"`
	CsrfKey        APIString `json:"csrf_key"`
	CORSOrigin     APIString `json:"cors_origin"`
}

func (a *APIUIConfig) BuildFromService(h interface{}) error {
	switch v := h.(type) {
	case evergreen.UIConfig:
		a.Url = ToAPIString(v.Url)
		a.HelpUrl = ToAPIString(v.HelpUrl)
		a.HttpListenAddr = ToAPIString(v.HttpListenAddr)
		a.Secret = ToAPIString(v.Secret)
		a.DefaultProject = ToAPIString(v.DefaultProject)
		a.CacheTemplates = v.CacheTemplates
		a.SecureCookies = v.SecureCookies
		a.CsrfKey = ToAPIString(v.CsrfKey)
		a.CORSOrigin = ToAPIString(v.CORSOrigin)
	default:
		return errors.Errorf("%T is not a supported type", h)
	}
	return nil
}

func (a *APIUIConfig) ToService() (interface{}, error) {
	return evergreen.UIConfig{
		Url:            FromAPIString(a.Url),
		HelpUrl:        FromAPIString(a.HelpUrl),
		HttpListenAddr: FromAPIString(a.HttpListenAddr),
		Secret:         FromAPIString(a.Secret),
		DefaultProject: FromAPIString(a.DefaultProject),
		CacheTemplates: a.CacheTemplates,
		SecureCookies:  a.SecureCookies,
		CsrfKey:        FromAPIString(a.CsrfKey),
		CORSOrigin:     FromAPIString(a.CORSOrigin),
	}, nil
}

// RestartTasksResponse is the response model returned from the /admin/restart route
type RestartTasksResponse struct {
	TasksRestarted []string `json:"tasks_restarted"`
	TasksErrored   []string `json:"tasks_errored"`
}

// BuildFromService builds a model from the service layer
func (ab *APIBanner) BuildFromService(h interface{}) error {
	switch v := h.(type) {
	case APIBanner:
		ab.Text = v.Text
		ab.Theme = v.Theme
	default:
		return errors.Errorf("%T is not a supported admin banner type", h)
	}
	return nil
}

// ToService is not yet implemented
func (ab *APIBanner) ToService() (interface{}, error) {
	return ab, nil
}

// BuildFromService builds a model from the service layer
func (as *APIServiceFlags) BuildFromService(h interface{}) error {
	switch v := h.(type) {
	case evergreen.ServiceFlags:
		as.TaskDispatchDisabled = v.TaskDispatchDisabled
		as.HostinitDisabled = v.HostinitDisabled
		as.MonitorDisabled = v.MonitorDisabled
		as.AlertsDisabled = v.AlertsDisabled
		as.TaskrunnerDisabled = v.TaskrunnerDisabled
		as.RepotrackerDisabled = v.RepotrackerDisabled
		as.SchedulerDisabled = v.SchedulerDisabled
		as.GithubPRTestingDisabled = v.GithubPRTestingDisabled
		as.RepotrackerPushEventDisabled = v.RepotrackerPushEventDisabled
		as.CLIUpdatesDisabled = v.CLIUpdatesDisabled
		as.EventProcessingDisabled = v.EventProcessingDisabled
		as.JIRANotificationsDisabled = v.JIRANotificationsDisabled
		as.SlackNotificationsDisabled = v.SlackNotificationsDisabled
		as.EmailNotificationsDisabled = v.EmailNotificationsDisabled
		as.WebhookNotificationsDisabled = v.WebhookNotificationsDisabled
		as.GithubStatusAPIDisabled = v.GithubStatusAPIDisabled
		as.BackgroundStatsDisabled = v.BackgroundStatsDisabled
		as.TaskLoggingDisabled = v.TaskLoggingDisabled
	default:
		return errors.Errorf("%T is not a supported service flags type", h)
	}
	return nil
}

// ToService returns a service model from an API model
func (as *APIServiceFlags) ToService() (interface{}, error) {
	return evergreen.ServiceFlags{
		TaskDispatchDisabled:         as.TaskDispatchDisabled,
		HostinitDisabled:             as.HostinitDisabled,
		MonitorDisabled:              as.MonitorDisabled,
		AlertsDisabled:               as.AlertsDisabled,
		TaskrunnerDisabled:           as.TaskrunnerDisabled,
		RepotrackerDisabled:          as.RepotrackerDisabled,
		SchedulerDisabled:            as.SchedulerDisabled,
		GithubPRTestingDisabled:      as.GithubPRTestingDisabled,
		RepotrackerPushEventDisabled: as.RepotrackerPushEventDisabled,
		CLIUpdatesDisabled:           as.CLIUpdatesDisabled,
		EventProcessingDisabled:      as.EventProcessingDisabled,
		JIRANotificationsDisabled:    as.JIRANotificationsDisabled,
		SlackNotificationsDisabled:   as.SlackNotificationsDisabled,
		EmailNotificationsDisabled:   as.EmailNotificationsDisabled,
		WebhookNotificationsDisabled: as.WebhookNotificationsDisabled,
		GithubStatusAPIDisabled:      as.GithubStatusAPIDisabled,
		BackgroundStatsDisabled:      as.BackgroundStatsDisabled,
		TaskLoggingDisabled:          as.TaskLoggingDisabled,
	}, nil
}

// BuildFromService builds a model from the service layer
func (rtr *RestartTasksResponse) BuildFromService(h interface{}) error {
	switch v := h.(type) {
	case *RestartTasksResponse:
		rtr.TasksRestarted = v.TasksRestarted
		rtr.TasksErrored = v.TasksErrored
	default:
		return errors.Errorf("%T is the incorrect type for a restart task response", h)
	}
	return nil
}

// ToService is not implemented for /admin/restart
func (rtr *RestartTasksResponse) ToService() (interface{}, error) {
	return nil, errors.New("ToService not implemented for RestartTasksResponse")
}

func AdminDbToRestModel(in evergreen.ConfigSection) (Model, error) {
	id := in.SectionId()
	var out Model
	if id == evergreen.ConfigDocID {
		out = &APIAdminSettings{}
		err := out.BuildFromService(reflect.ValueOf(in).Interface())
		if err != nil {
			return nil, err
		}
	} else {
		structVal := reflect.ValueOf(*NewConfigModel())
		for i := 0; i < structVal.NumField(); i++ {
			// this assumes that the json tag is the same as the section ID
			tag := strings.Split(structVal.Type().Field(i).Tag.Get("json"), ",")[0]
			if tag != id {
				continue
			}

			propName := structVal.Type().Field(i).Name
			propVal := structVal.FieldByName(propName)
			propInterface := propVal.Interface()
			apiModel, ok := propInterface.(Model)
			if !ok {
				return nil, fmt.Errorf("unable to convert section %s to a Model interface", id)
			}
			out = apiModel
		}
		if out == nil {
			return nil, fmt.Errorf("section %s is not defined in the APIAdminSettings struct", id)
		}
		err := out.BuildFromService(reflect.Indirect(reflect.ValueOf(in)).Interface())
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}

type APIJIRANotificationsConfig struct {
	CustomFields map[string]map[string]string `json:"custom_fields,omitempty"`
}

func (j *APIJIRANotificationsConfig) BuildFromService(h interface{}) error {
	var config *evergreen.JIRANotificationsConfig
	switch v := h.(type) {
	case *evergreen.JIRANotificationsConfig:
		config = v
	case evergreen.JIRANotificationsConfig:
		config = &v
	default:
		return errors.Errorf("expected *evergreen.APIJIRANotificationsConfig, but got %T instead", h)
	}

	if len(config.CustomFields) == 0 {
		return nil
	}

	m, err := config.CustomFields.ToMap()
	if err != nil {
		return errors.Wrap(err, "failed to build jira custom field configuration")
	}

	j.CustomFields = m

	return nil
}
func (j *APIJIRANotificationsConfig) ToService() (interface{}, error) {
	if j.CustomFields == nil || len(j.CustomFields) == 0 {
		return evergreen.JIRANotificationsConfig{}, nil
	}
	config := evergreen.JIRANotificationsConfig{
		CustomFields: evergreen.JIRACustomFieldsByProject{},
	}

	config.CustomFields.FromMap(j.CustomFields)

	return config, nil
}
