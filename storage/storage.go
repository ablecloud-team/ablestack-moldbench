package storage

import (
	"moldbench/utils"

	"github.com/ablecloud-team/ablestack-mold-go/v2/cloudstack"
)

func CreateStoragePool(cs *cloudstack.CloudStackClient, storage_name, storage_url, zoneid, pod_id, cluster_id string) (*cloudstack.CreateStoragePoolResponse, error) {
	p := cs.Pool.NewCreateStoragePoolParams(storage_name, storage_url, zoneid)
	p.SetHypervisor("KVM")
	p.SetProvider("DefaultPrimary")
	p.SetScope("CLUSTER")
	p.SetClusterid(cluster_id)
	p.SetPodid(pod_id)
	p.SetTags(storage_name)

	resp, err := cs.Pool.CreateStoragePool(p)
	if err != nil {
		// log.Printf("Failed to create network due to: %v", err)
		utils.HandleError(err)
		return nil, err
	}
	return resp, nil
}
