package envoyfilter

import (
	"encoding/json"
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/istio/envoyfilter/config_patch"
	"google.golang.org/protobuf/types/known/structpb"
	"istio.io/api/networking/v1alpha3"
	v1alpha4 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Generate(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	application, ok := r.GetSKIPObject().(*skiperatorv1alpha1.Application)
	if !ok {
		err := fmt.Errorf("failed to cast resource to application")
		ctxLog.Error(err, "Failed to generate auto login EnvoyFilter")
		return err
	}
	ctxLog.Debug("Attempting to auto login EnvoyFilter for application", "application", application.Name)

	oAuthClusterConfigPatchValueAsPbStruct, err := convertToPbStruct(config_patch.GetOAuthClusterConfigPatchValue("test"))
	if err != nil {
		ctxLog.Error(err, "failed to convert OAuth cluster config patch to protobuf")
	}
	oAuthSidecarConfigPatchValueAsPbStruct, err := convertToPbStruct(config_patch.GetOAuthSidecarConfigPatchValue())
	if err != nil {
		ctxLog.Error(err, "failed to convert OAuth sidecar config patch to protobuf")
	}

	autoLoginEnvoyFilter := v1alpha4.EnvoyFilter{
		ObjectMeta: metav1.ObjectMeta{
			Name:      application.Name + "-auto-login",
			Namespace: application.Namespace,
		},
		Spec: v1alpha3.EnvoyFilter{
			ConfigPatches: []*v1alpha3.EnvoyFilter_EnvoyConfigObjectPatch{
				{
					ApplyTo: v1alpha3.EnvoyFilter_CLUSTER,
					Match: &v1alpha3.EnvoyFilter_EnvoyConfigObjectMatch{
						ObjectTypes: &v1alpha3.EnvoyFilter_EnvoyConfigObjectMatch_Cluster{
							Cluster: &v1alpha3.EnvoyFilter_ClusterMatch{
								Service: "oauth",
							},
						},
					},
					Patch: &v1alpha3.EnvoyFilter_Patch{
						Operation: v1alpha3.EnvoyFilter_Patch_ADD,
						Value:     oAuthClusterConfigPatchValueAsPbStruct,
					},
				},
				{
					ApplyTo: v1alpha3.EnvoyFilter_HTTP_FILTER,
					Match: &v1alpha3.EnvoyFilter_EnvoyConfigObjectMatch{
						Context: v1alpha3.EnvoyFilter_SIDECAR_INBOUND,
						ObjectTypes: &v1alpha3.EnvoyFilter_EnvoyConfigObjectMatch_Listener{
							Listener: &v1alpha3.EnvoyFilter_ListenerMatch{
								FilterChain: &v1alpha3.EnvoyFilter_ListenerMatch_FilterChainMatch{
									Filter: &v1alpha3.EnvoyFilter_ListenerMatch_FilterMatch{
										Name: "envoy.filters.network.http_connection_manager",
										SubFilter: &v1alpha3.EnvoyFilter_ListenerMatch_SubFilterMatch{
											Name: "envoy.filters.jwt_authn",
										},
									},
								},
							},
						},
					},
					Patch: &v1alpha3.EnvoyFilter_Patch{
						Operation: v1alpha3.EnvoyFilter_Patch_INSERT_BEFORE,
						Value:     oAuthSidecarConfigPatchValueAsPbStruct,
					},
				},
			},
		},
	}

	r.AddResource(&autoLoginEnvoyFilter)
	ctxLog.Debug("Finished generating auto login EnvoyFilter for application", "application", application.Name)
	return nil
}

func convertToPbStruct[T any](v T) (*structpb.Struct, error) {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var rawMap map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &rawMap); err != nil {
		return nil, err
	}
	return structpb.NewStruct(rawMap)
}
