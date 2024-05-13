package istio

import (
	"errors"
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/api/v1alpha1/podtypes"
	"github.com/kartverket/skiperator/pkg/util"
	"golang.org/x/exp/slices"
	networkingv1beta1api "istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

func setCloudSqlRule(accessPolicy *podtypes.AccessPolicy, object client.Object) (*podtypes.AccessPolicy, error) {
	application, ok := object.(*skiperatorv1alpha1.Application)
	if !ok {
		return accessPolicy, nil
	}

	if !util.IsCloudSqlProxyEnabled(application.Spec.GCP) {
		return accessPolicy, nil
	}

	if application.Spec.GCP.CloudSQLProxy.IP == "" {
		return nil, errors.New("cloud sql proxy IP is not set")
	}

	externalRule := &podtypes.ExternalRule{
		Host:  fmt.Sprintf("%x.cloudsql", util.GenerateHashFromName(application.GetName())),
		Ip:    application.Spec.GCP.CloudSQLProxy.IP,
		Ports: []podtypes.ExternalPort{{Name: "cloudsqlproxy", Port: 3307, Protocol: "TCP"}},
	}

	if accessPolicy == nil {
		accessPolicy = &podtypes.AccessPolicy{}
	}

	(*accessPolicy).Outbound.External = append((*accessPolicy).Outbound.External, *externalRule)

	return accessPolicy, nil
}

func GetServiceEntries(accessPolicy *podtypes.AccessPolicy, object client.Object) ([]networkingv1beta1.ServiceEntry, error) {
	var serviceEntries []networkingv1beta1.ServiceEntry

	accessPolicy, err := setCloudSqlRule(accessPolicy, object)
	if err != nil {
		return nil, err
	}

	if accessPolicy != nil {
		for _, rule := range (*accessPolicy).Outbound.External {
			serviceEntryName := fmt.Sprintf("%s-egress-%x", object.GetName(), util.GenerateHashFromName(rule.Host))

			objectKind := object.GetObjectKind().GroupVersionKind().Kind

			switch object.(type) {
			case *skiperatorv1alpha1.Application:
				break
			default:
				serviceEntryName = fmt.Sprintf("%v-%v", strings.ToLower(objectKind), serviceEntryName)
			}

			resolution, addresses, endpoints := getIpData(rule.Ip)

			ports, err := getPorts(rule.Ports, rule.Ip)
			if err != nil {
				return nil, err
			}

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
					Ports:      ports,
				},
			}

			serviceEntries = append(serviceEntries, serviceEntry)
		}
	}

	return serviceEntries, nil
}

func GetServiceEntriesToDelete(serviceEntriesInNamespace []*networkingv1beta1.ServiceEntry, ownerName string, currentEgresses []networkingv1beta1.ServiceEntry) []networkingv1beta1.ServiceEntry {
	var serviceEntriesToDelete []networkingv1beta1.ServiceEntry

	for _, serviceEntry := range serviceEntriesInNamespace {

		ownerIndex := slices.IndexFunc(serviceEntry.GetOwnerReferences(), func(ownerReference metav1.OwnerReference) bool {
			return ownerReference.Name == ownerName
		})
		serviceEntryOwnedByThisApplication := ownerIndex != -1
		if !serviceEntryOwnedByThisApplication {
			continue
		}

		serviceEntryInCurrentEgresses := slices.IndexFunc(currentEgresses, func(inSpecEntry networkingv1beta1.ServiceEntry) bool {
			return inSpecEntry.Name == serviceEntry.Name
		})

		serviceEntryInOwnerSpec := serviceEntryInCurrentEgresses != -1
		if serviceEntryInOwnerSpec {
			continue
		}

		serviceEntriesToDelete = append(serviceEntriesToDelete, *serviceEntry)
	}

	return serviceEntriesToDelete
}

func getPorts(externalPorts []podtypes.ExternalPort, ruleIP string) ([]*networkingv1beta1api.ServicePort, error) {
	var ports []*networkingv1beta1api.ServicePort

	if len(externalPorts) == 0 {
		ports = append(ports, &networkingv1beta1api.ServicePort{
			Name:     "https",
			Number:   uint32(443),
			Protocol: "HTTPS",
		})

		return ports, nil
	}

	for _, port := range externalPorts {
		if ruleIP == "" && port.Protocol == "TCP" {
			errorMessage := fmt.Sprintf("static IP must be set for TCP port, found IP: %v", ruleIP)
			return nil, errors.New(errorMessage)
		}

		ports = append(ports, &networkingv1beta1api.ServicePort{
			Name:     port.Name,
			Number:   uint32(port.Port),
			Protocol: port.Protocol,
		})

	}

	return ports, nil
}

func getIpData(ip string) (networkingv1beta1api.ServiceEntry_Resolution, []string, []*networkingv1beta1api.WorkloadEntry) {
	if ip == "" {
		return networkingv1beta1api.ServiceEntry_DNS, nil, nil
	}

	return networkingv1beta1api.ServiceEntry_STATIC, []string{ip}, []*networkingv1beta1api.WorkloadEntry{{Address: ip}}
}
