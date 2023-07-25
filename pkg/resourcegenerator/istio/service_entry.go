package istio

import (
	"fmt"
	"github.com/kartverket/skiperator/api/v1alpha1/podtypes"
	"github.com/kartverket/skiperator/pkg/util"
	networkingv1beta1api "istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"regexp"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetServiceEntries(accessPolicy *podtypes.AccessPolicy, object client.Object) []networkingv1beta1.ServiceEntry {
	var serviceEntries []networkingv1beta1.ServiceEntry

	if accessPolicy != nil {
		for _, rule := range (*accessPolicy).Outbound.External {
			serviceEntryNamePrefix := fmt.Sprintf("%s-egress-%x", object.GetName(), util.GenerateHashFromName(rule.Host))
			serviceEntryName := util.ResourceNameWithHash(serviceEntryNamePrefix, object.GetObjectKind().GroupVersionKind().Kind)
			resolution, addresses, endpoints := getIpData(rule.Ip)

			serviceEntry := networkingv1beta1.ServiceEntry{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: object.GetNamespace(),
					Name:      serviceEntryName,
				},
				Spec: networkingv1beta1api.ServiceEntry{
					// Avoid leaking service entry to other namespaces
					ExportTo:   []string{".", "istio-system", "istio-gateways"},
					Hosts:      []string{rule.Host},
					Resolution: resolution,
					Addresses:  addresses,
					Endpoints:  endpoints,
					Ports:      getPorts(rule.Ports, rule.Ip),
				},
			}

			serviceEntries = append(serviceEntries, serviceEntry)
		}
	}

	return serviceEntries
}

func getPorts(externalPorts []podtypes.ExternalPort, ruleIP string) []*networkingv1beta1api.Port {
	var ports []*networkingv1beta1api.Port

	if len(externalPorts) == 0 {
		ports = append(ports, &networkingv1beta1api.Port{
			Name:     "https",
			Number:   uint32(443),
			Protocol: "HTTPS",
		})

		return ports
	}

	for _, port := range externalPorts {
		if ruleIP == "" && port.Protocol == "TCP" {
			continue
		}

		ports = append(ports, &networkingv1beta1api.Port{
			Name:     port.Name,
			Number:   uint32(port.Port),
			Protocol: port.Protocol,
		})

	}

	return ports
}

func getIpData(ip string) (networkingv1beta1api.ServiceEntry_Resolution, []string, []*networkingv1beta1api.WorkloadEntry) {
	if ip == "" {
		return networkingv1beta1api.ServiceEntry_DNS, nil, nil
	}

	return networkingv1beta1api.ServiceEntry_STATIC, []string{ip}, []*networkingv1beta1api.WorkloadEntry{{Address: ip}}
}

// Filter for service entries named like *-egress-*
func IsEgressServiceEntry(serviceEntry *networkingv1beta1.ServiceEntry) bool {
	match, _ := regexp.MatchString("^.*-egress-.*$", serviceEntry.Name)

	return match
}
