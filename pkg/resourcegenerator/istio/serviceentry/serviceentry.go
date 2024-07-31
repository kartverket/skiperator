package serviceentry

import (
	"errors"
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/api/v1alpha1/podtypes"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/util"
	networkingv1beta1api "istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

func Generate(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()

	if r.GetType() == reconciliation.ApplicationType || r.GetType() == reconciliation.JobType {
		return getServiceEntries(r)
	} else {
		err := fmt.Errorf("unsupported type %s in service entry", r.GetType())
		ctxLog.Error(err, "Failed to generate service entry")
		return err
	}
}

func getServiceEntries(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	ctxLog.Debug("Attempting to generate service entries", "type", r.GetType())

	object := r.GetSKIPObject()
	accessPolicy := r.GetCommonSpec().AccessPolicy

	accessPolicy, err := setCloudSqlRule(accessPolicy, object)
	if err != nil {
		return err
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
				return err
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

			var obj client.Object = &serviceEntry
			r.AddResource(obj)
		}
	}

	ctxLog.Debug("Finished generating service entries for type", "type", r.GetType(), "name", object.GetName())
	return nil
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

	// The istio validation webhook will reject the service entry if the host is not a valid DNS name, such as an IP address.
	// So we generate something that will not crash with other apps in the same namespace.
	externalRule := &podtypes.ExternalRule{
		Host:  fmt.Sprintf("%s-%x.cloudsql", application.Name, util.GenerateHashFromName(application.Spec.Image)),
		Ip:    application.Spec.GCP.CloudSQLProxy.IP,
		Ports: []podtypes.ExternalPort{{Name: "cloudsqlproxy", Port: 3307, Protocol: "TCP"}},
	}

	if accessPolicy == nil {
		accessPolicy = &podtypes.AccessPolicy{}
	}

	(*accessPolicy).Outbound.External = append((*accessPolicy).Outbound.External, *externalRule)

	return accessPolicy, nil
}
