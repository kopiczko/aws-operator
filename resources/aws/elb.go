package aws

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	awsclient "github.com/giantswarm/aws-operator/client/aws"
	microerror "github.com/giantswarm/microkit/error"
)

// ELB is an Elastic Load Balancer
type ELB struct {
	Name             string
	dnsName          string
	hostedZoneID     string
	AZ               string
	SecurityGroup    string
	SubnetID         string
	Tags             []string
	Client           *elb.ELB
	LoadBalancerPort int
	InstancePort     int
}

func (lb *ELB) CreateIfNotExists() (bool, error) {
	if lb.Client == nil {
		return false, microerror.MaskAny(clientNotInitializedError)
	}

	if err := lb.CreateOrFail(); err != nil {
		if strings.Contains(err.Error(), awsclient.ELBConfigurationMismatch) {
			return false, microerror.MaskAny(err)
		}
		if strings.Contains(err.Error(), awsclient.ELBAlreadyExists) {
			return false, nil
		}

		return false, microerror.MaskAny(err)
	}

	return true, nil
}

func (lb *ELB) CreateOrFail() error {
	if lb.Client == nil {
		return microerror.MaskAny(clientNotInitializedError)
	}

	if _, err := lb.Client.CreateLoadBalancer(&elb.CreateLoadBalancerInput{
		LoadBalancerName: aws.String(lb.Name),
		Listeners: []*elb.Listener{
			{
				InstancePort:     aws.Int64(int64(lb.InstancePort)),
				LoadBalancerPort: aws.Int64(int64(lb.LoadBalancerPort)),
				// TCP because we want to do SSL passthrough, not termination
				Protocol: aws.String("TCP"),
			},
		},
		// we use the Subnet ID instead, since only one of either can be specified
		// AvailabilityZones: []*string{
		// 	aws.String(lb.AZ),
		// },
		SecurityGroups: []*string{
			aws.String(lb.SecurityGroup),
		},
		Subnets: []*string{
			aws.String(lb.SubnetID),
		},
	}); err != nil {
		return microerror.MaskAny(err)
	}

	// we have to populate some additional fields
	lbDescription, err := lb.findExisting()
	if err != nil {
		return microerror.MaskAny(err)
	}

	lb.setDNSFields(*lbDescription)

	return nil
}

func (lb ELB) Delete() error {
	if lb.Client == nil {
		return microerror.MaskAny(clientNotInitializedError)
	}

	if _, err := lb.Client.DeleteLoadBalancer(&elb.DeleteLoadBalancerInput{
		LoadBalancerName: aws.String(lb.Name),
	}); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

func (lb *ELB) RegisterInstances(instanceIDs []string) error {
	var instances []*elb.Instance

	for _, id := range instanceIDs {
		elbInstance := &elb.Instance{
			InstanceId: aws.String(id),
		}
		instances = append(instances, elbInstance)
	}

	if _, err := lb.Client.RegisterInstancesWithLoadBalancer(&elb.RegisterInstancesWithLoadBalancerInput{
		Instances:        instances,
		LoadBalancerName: aws.String(lb.Name),
	}); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

func (lb ELB) DNSName() string {
	return lb.dnsName
}

func (lb ELB) HostedZoneID() string {
	return lb.hostedZoneID
}

func (lb ELB) findExisting() (*elb.LoadBalancerDescription, error) {
	resp, err := lb.Client.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{
			aws.String(lb.Name),
		},
		PageSize: aws.Int64(1),
	})

	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	descriptions := resp.LoadBalancerDescriptions

	if len(descriptions) == 0 {
		return nil, NamedResourceNotFoundError{Name: lb.Name}
	}

	return descriptions[0], nil
}

// NewExistingELB retrieves and sets additional fields that deal with the ELB's location in the DNS,
// such as its FQDN and its Hosted Zone ID.
func NewExistingELB(name string, client *elb.ELB) (*ELB, error) {
	lb := ELB{
		Name:   name,
		Client: client,
	}

	lbDescription, err := lb.findExisting()
	if err != nil {
		return nil, err
	}

	lb.setDNSFields(*lbDescription)

	return &lb, nil
}

func (lb *ELB) setDNSFields(desc elb.LoadBalancerDescription) {
	lb.dnsName = *desc.DNSName
	lb.hostedZoneID = *desc.CanonicalHostedZoneNameID
}
