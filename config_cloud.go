package evergreen

import (
	"github.com/evergreen-ci/evergreen/db"
	"github.com/pkg/errors"
	"gopkg.in/mgo.v2/bson"
)

// CloudProviders stores configuration settings for the supported cloud host providers.
type CloudProviders struct {
	AWS       AWSConfig       `bson:"aws" json:"aws" yaml:"aws"`
	Docker    DockerConfig    `bson:"docker" json:"docker" yaml:"docker"`
	GCE       GCEConfig       `bson:"gce" json:"gce" yaml:"gce"`
	OpenStack OpenStackConfig `bson:"openstack" json:"openstack" yaml:"openstack"`
	VSphere   VSphereConfig   `bson:"vsphere" json:"vsphere" yaml:"vsphere"`
}

func (c *CloudProviders) SectionId() string { return "providers" }

func (c *CloudProviders) Get() error {
	err := db.FindOneQ(ConfigCollection, db.Query(byId(c.SectionId())), c)
	if err != nil && err.Error() == errNotFound {
		*c = CloudProviders{}
		return nil
	}
	return errors.Wrapf(err, "error retrieving section %s", c.SectionId())
}

func (c *CloudProviders) Set() error {
	_, err := db.Upsert(ConfigCollection, byId(c.SectionId()), bson.M{
		"$set": bson.M{
			"aws":       c.AWS,
			"docker":    c.Docker,
			"gce":       c.GCE,
			"openstack": c.OpenStack,
			"vsphere":   c.VSphere,
		},
	})
	return errors.Wrapf(err, "error updating section %s", c.SectionId())
}

func (c *CloudProviders) ValidateAndDefault() error { return nil }

// AWSConfig stores auth info for Amazon Web Services.
type AWSConfig struct {
	Secret string `bson:"aws_secret" json:"aws_secret" yaml:"aws_secret"`
	Id     string `bson:"aws_id" json:"aws_id" yaml:"aws_id"`
}

// DockerConfig stores auth info for Docker.
type DockerConfig struct {
	APIVersion string `bson:"api_version" json:"api_version" yaml:"api_version"`
}

// OpenStackConfig stores auth info for Linaro using Identity V3. All fields required.
//
// The config is NOT compatible with Identity V2.
type OpenStackConfig struct {
	IdentityEndpoint string `bson:"identity_endpoint" json:"identity_endpoint" yaml:"identity_endpoint"`

	Username   string `bson:"username" json:"username" yaml:"username"`
	Password   string `bson:"password" json:"password" yaml:"password"`
	DomainName string `bson:"domain_name" json:"domain_name" yaml:"domain_name"`

	ProjectName string `bson:"project_name" json:"project_name" yaml:"project_name"`
	ProjectID   string `bson:"project_id" json:"project_id" yaml:"project_id"`

	Region string `bson:"region" json:"region" yaml:"region"`
}

// GCEConfig stores auth info for Google Compute Engine. Can be retrieved from:
// https://developers.google.com/identity/protocols/application-default-credentials
type GCEConfig struct {
	ClientEmail  string `bson:"client_email" json:"client_email" yaml:"client_email"`
	PrivateKey   string `bson:"private_key" json:"private_key" yaml:"private_key"`
	PrivateKeyID string `bson:"private_key_id" json:"private_key_id" yaml:"private_key_id"`
	TokenURI     string `bson:"token_uri" json:"token_uri" yaml:"token_uri"`
}

// VSphereConfig stores auth info for VMware vSphere. The config fields refer
// to your vCenter server, a centralized management tool for the vSphere suite.
type VSphereConfig struct {
	Host     string `bson:"host" json:"host" yaml:"host"`
	Username string `bson:"username" json:"username" yaml:"username"`
	Password string `bson:"password" json:"password" yaml:"password"`
}
