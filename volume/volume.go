// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package volume

import (
	"moldbench/config"
	"moldbench/utils"

	"github.com/ablecloud-team/ablestack-mold-go/v2/cloudstack"
)

func ListVolumes(cs *cloudstack.CloudStackClient, domainId string) ([]*cloudstack.Volume, error) {
	result := make([]*cloudstack.Volume, 0)
	page := 1
	p := cs.Volume.NewListVolumesParams()
	p.SetDomainid(domainId)
	p.SetPagesize(config.PageSize)
	for {
		p.SetPage(page)
		resp, err := cs.Volume.ListVolumes(p)

		if err != nil {
			// log.Printf("Failed to list volume due to: %v", err)
			utils.HandleError(err)
			return result, err
		}
		result = append(result, resp.Volumes...)
		if len(result) < resp.Count {
			page++
		} else {
			break
		}
	}
	return result, nil
}

func CreateVolume(cs *cloudstack.CloudStackClient, domainId string, account string) (*cloudstack.CreateVolumeResponse, error) {
	volName := "Volume-" + utils.RandomString(10)
	p := cs.Volume.NewCreateVolumeParams()
	p.SetDomainid(domainId)
	p.SetName(volName)
	p.SetZoneid(config.ZoneId)
	p.SetDiskofferingid(config.DiskOfferingId)
	p.SetAccount(account)
	resp, err := cs.Volume.CreateVolume(p)
	if err != nil {
		// log.Printf("Failed to create volume due to: %v", err)
		utils.HandleError(err)
		return nil, err
	}
	return resp, nil
}

func DestroyVolume(cs *cloudstack.CloudStackClient, volumeId string) (*cloudstack.DestroyVolumeResponse, error) {
	p := cs.Volume.NewDestroyVolumeParams(volumeId)
	p.SetExpunge(true)
	resp, err := cs.Volume.DestroyVolume(p)
	if err != nil {
		// log.Printf("Failed to destroy volume with id  %s due to %v", volumeId, err)
		utils.HandleError(err)
		return nil, err
	}
	return resp, nil
}

func AttachVolume(cs *cloudstack.CloudStackClient, volumeId string, vmId string) (*cloudstack.AttachVolumeResponse, error) {
	p := cs.Volume.NewAttachVolumeParams(volumeId, vmId)
	resp, err := cs.Volume.AttachVolume(p)
	if err != nil {
		// log.Printf("Failed to attach volume due to: %v", err)
		utils.HandleError(err)
		return nil, err
	}
	return resp, nil
}

func DetachVolume(cs *cloudstack.CloudStackClient, volumeId string) (*cloudstack.DetachVolumeResponse, error) {
	p := cs.Volume.NewDetachVolumeParams()
	p.SetId(volumeId)
	resp, err := cs.Volume.DetachVolume(p)

	if err != nil {
		// log.Printf("Failed to detach volume due to: %v", err)
		utils.HandleError(err)
		return nil, err
	}
	return resp, nil
}

func CreateSnapshot(cs *cloudstack.CloudStackClient, volumeId string) (*cloudstack.CreateSnapshotResponse, error) {
	p := cs.Snapshot.NewCreateSnapshotParams(volumeId)
	resp, err := cs.Snapshot.CreateSnapshot(p)
	if err != nil {
		// log.Printf("Failed to create snapshot due to: %v", err)
		utils.HandleError(err)
		return nil, err
	}
	return resp, nil
}

func DeleteSnapshot(cs *cloudstack.CloudStackClient, snapshotId string) (*cloudstack.DeleteSnapshotResponse, error) {
	p := cs.Snapshot.NewDeleteSnapshotParams(snapshotId)
	resp, err := cs.Snapshot.DeleteSnapshot(p)
	if err != nil {
		// log.Printf("Failed to delete snapshot with id %s due to %v", snapshotId, err)
		utils.HandleError(err)
		return nil, err
	}
	return resp, nil
}
