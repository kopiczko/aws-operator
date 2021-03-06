package clustertpr

import (
	"github.com/giantswarm/clustertpr/calico"
	"github.com/giantswarm/clustertpr/cluster"
	"github.com/giantswarm/clustertpr/customer"
	"github.com/giantswarm/clustertpr/docker"
	"github.com/giantswarm/clustertpr/etcd"
	"github.com/giantswarm/clustertpr/flannel"
	"github.com/giantswarm/clustertpr/kubernetes"
	"github.com/giantswarm/clustertpr/node"
	"github.com/giantswarm/clustertpr/operator"
	"github.com/giantswarm/clustertpr/vault"
)

type Cluster struct {
	Calico     calico.Calico         `json:"calico" yaml:"calico"`
	Cluster    cluster.Cluster       `json:"cluster" yaml:"cluster"`
	Customer   customer.Customer     `json:"customer" yaml:"customer"`
	Docker     docker.Docker         `json:"docker" yaml:"docker"`
	Etcd       etcd.Etcd             `json:"etcd" yaml:"etcd"`
	Flannel    flannel.Flannel       `json:"flannel" yaml:"flannel"`
	Kubernetes kubernetes.Kubernetes `json:"kubernetes" yaml:"kubernetes"`
	Masters    []node.Node           `json:"masters" yaml:"masters"`
	Operator   operator.Operator     `json:"operator" yaml:"operator"`
	Vault      vault.Vault           `json:"vault" yaml:"vault"`
	Workers    []node.Node           `json:"workers" yaml:"workers"`
}
