package main

import (
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	networkingv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/networking/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	dnsPort = 53
)

var namespaces = map[string]pulumi.StringMap{
	"home-control": {
		"istio-injection": pulumi.String("enabled"),
		"type":            pulumi.String("infrastructure"),
		"managed-by":      pulumi.String("pulumi"),
	},
	"argocd": {
		"type":       pulumi.String("infrastructure"),
		"managed-by": pulumi.String("pulumi"),
	},
	"ingress-nginx": {
		"type":       pulumi.String("infrastructure"),
		"managed-by": pulumi.String("pulumi"),
	},
	"monitoring": {
		"type":       pulumi.String("infrastructure"),
		"managed-by": pulumi.String("pulumi"),
	},
	"logging": {
		"type":       pulumi.String("infrastructure"),
		"managed-by": pulumi.String("pulumi"),
	},
	"kafka": {
		"type":       pulumi.String("infrastructure"),
		"managed-by": pulumi.String("pulumi"),
	},
	"redis": {
		"type":       pulumi.String("infrastructure"),
		"managed-by": pulumi.String("pulumi"),
	},
	"postgres": {
		"type":       pulumi.String("infrastructure"),
		"managed-by": pulumi.String("pulumi"),
	},
}

type NetworkPolicyConfig struct {
	name      string
	namespace pulumi.StringInput
	app       string
	ports     []int
	protocols []string
}

func createNamespaces(ctx *pulumi.Context) error {
	for name, labels := range namespaces {
		_, err := corev1.NewNamespace(ctx, name, &corev1.NamespaceArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name:   pulumi.String(name),
				Labels: labels,
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func createNetworkPolicy(ctx *pulumi.Context, config NetworkPolicyConfig) error {
	_, err := networkingv1.NewNetworkPolicy(ctx, config.name, &networkingv1.NetworkPolicyArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Namespace: config.namespace,
			Labels: pulumi.StringMap{
				"managed-by": pulumi.String("pulumi"),
			},
		},
		Spec: &networkingv1.NetworkPolicySpecArgs{
			PodSelector: &metav1.LabelSelectorArgs{},
			PolicyTypes: pulumi.StringArray{
				pulumi.String("Egress"),
			},
			Egress: networkingv1.NetworkPolicyEgressRuleArray{
				&networkingv1.NetworkPolicyEgressRuleArgs{
					To: networkingv1.NetworkPolicyPeerArray{
						&networkingv1.NetworkPolicyPeerArgs{
							PodSelector: &metav1.LabelSelectorArgs{
								MatchLabels: pulumi.StringMap{
									"app": pulumi.String(config.app),
								},
							},
						},
					},
					Ports: createNetworkPolicyPorts(config.ports, config.protocols),
				},

				&networkingv1.NetworkPolicyEgressRuleArgs{
					To: networkingv1.NetworkPolicyPeerArray{
						&networkingv1.NetworkPolicyPeerArgs{
							IpBlock: &networkingv1.IPBlockArgs{
								Cidr: pulumi.String("0.0.0.0/0"),
							},
						},
					},
				},
			},
		},
	})
	return err
}

func createNetworkPolicyPorts(ports []int, protocols []string) networkingv1.NetworkPolicyPortArray {
	var policyPorts networkingv1.NetworkPolicyPortArray
	for _, port := range ports {
		for _, protocol := range protocols {
			policyPorts = append(policyPorts, &networkingv1.NetworkPolicyPortArgs{
				Port:     pulumi.Int(port),
				Protocol: pulumi.String(protocol),
			})
		}
	}
	return policyPorts
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		if err := createNamespaces(ctx); err != nil {
			return err
		}

		err := createNetworkPolicy(ctx, NetworkPolicyConfig{
			name:      "global-dns-access",
			namespace: pulumi.String("home-control"),
			app:       "coredns",
			ports:     []int{dnsPort},
			protocols: []string{"UDP", "TCP"},
		})

		return err
	})
}
