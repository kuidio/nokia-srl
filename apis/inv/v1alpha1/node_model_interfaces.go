package v1alpha1

import (
	"fmt"

	infrabev1alpha1 "github.com/kuidio/kuid/apis/backend/infra/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *NodeModel) GetEndPoints(cr *infrabev1alpha1.Node) ([]*infrabev1alpha1.Endpoint, error) {

	nodeGroupNodeID := infrabev1alpha1.String2NodeGroupNodeID(cr.GetName())
	if nodeGroupNodeID == nil {
		return nil, fmt.Errorf("cannot get node id from node, wrong nodeName, got: %s", cr.GetName())
	}

	endpoints := make([]*infrabev1alpha1.Endpoint, 0, len(r.Spec.Interfaces))
	for _, itfce := range r.Spec.Interfaces {
		var speed string
		if itfce.Speed != nil {
			speed = *itfce.Speed
		}

		nodeGroupEndpointID := infrabev1alpha1.NodeGroupEndpointID{
			NodeGroup: nodeGroupNodeID.NodeGroup,
			EndpointID: infrabev1alpha1.EndpointID{
				Endpoint: itfce.Name,
				NodeID:   nodeGroupNodeID.NodeID,
			},
		}
		endpoints = append(endpoints,
			infrabev1alpha1.BuildEndpoint(
				v1.ObjectMeta{
					Namespace: cr.GetNamespace(),
					Name:      nodeGroupEndpointID.KuidString(),
				},
				&infrabev1alpha1.EndpointSpec{
					Provider: cr.Spec.Provider,
					NodeGroupEndpointID: nodeGroupEndpointID,
					Speed:               &speed,
					VLANTagging:         itfce.VLANTagging,
				},
				nil,
			),
		)
	}
	return endpoints, nil
}
