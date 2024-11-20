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

package vm

import (
	"moldbench/config"
	"moldbench/utils"

	"github.com/ablecloud-team/ablestack-mold-go/v2/cloudstack"
)

func ListVMs(cs *cloudstack.CloudStackClient, domainId string) ([]*cloudstack.VirtualMachine, error) {
	result := make([]*cloudstack.VirtualMachine, 0)
	page := 1
	p := cs.VirtualMachine.NewListVirtualMachinesParams()
	p.SetDomainid(domainId)
	p.SetPagesize(config.PageSize)
	for {
		p.SetPage(page)
		resp, err := cs.VirtualMachine.ListVirtualMachines(p)
		if err != nil {
			// log.Printf("Failed to list vm due to %v", err)
			utils.HandleError(err)
			return result, err
		}
		result = append(result, resp.VirtualMachines...)
		if len(result) < resp.Count {
			page++
		} else {
			break
		}
	}
	return result, nil
}

func DeployVm(cs *cloudstack.CloudStackClient, domainId string, networkId string, account string) (*cloudstack.DeployVirtualMachineResponse, error) {
	vmName := "vm-" + utils.RandomString(10)
	p := cs.VirtualMachine.NewDeployVirtualMachineParams(config.ServiceOfferingId, config.TemplateId, vmName)
	p.SetDomainid(domainId)
	p.SetZoneid(config.ZoneId)
	p.SetNetworkids([]string{networkId})
	p.SetName(vmName)
	p.SetAccount(account)
	p.SetStartvm(config.StartVM)
	resp, err := cs.VirtualMachine.DeployVirtualMachine(p)
	if err != nil {
		// log.Printf("Failed to deploy vm due to: %v", err)
		utils.HandleError(err)
		return nil, err
	}
	return resp, nil
}

func DestroyVm_cs(cs *cloudstack.CloudStackClient, vmId string) error {
	deleteParams := cs.VirtualMachine.NewDestroyVirtualMachineParams(vmId)
	deleteParams.SetExpunge(true)
	_, err := cs.VirtualMachine.DestroyVirtualMachine(deleteParams)
	if err != nil {
		// log.Printf("Failed to destroy Vm with Id %s due to %v", vmId, err)
		utils.HandleError(err)
		return err
	}
	return nil
}

func DestroyVm(cs *cloudstack.CloudStackClient, vmId string) (*cloudstack.DestroyVirtualMachineResponse, error) {
	deleteParams := cs.VirtualMachine.NewDestroyVirtualMachineParams(vmId)
	deleteParams.SetExpunge(true)
	resp, err := cs.VirtualMachine.DestroyVirtualMachine(deleteParams)
	if err != nil {
		// log.Printf("Failed to destroy Vm with Id %s due to %v", vmId, err)
		utils.HandleError(err)
		return nil, err
	}
	return resp, nil
}

func StartVM_cs(cs *cloudstack.CloudStackClient, vmId string) error {
	p := cs.VirtualMachine.NewStartVirtualMachineParams(vmId)
	_, err := cs.VirtualMachine.StartVirtualMachine(p)
	if err != nil {
		// log.Printf("Failed to start vm with id %s due to %v", vmId, err)
		utils.HandleError(err)
		return err
	}
	return nil
}

func StartVM(cs *cloudstack.CloudStackClient, vmId string) (*cloudstack.StartVirtualMachineResponse, error) {
	p := cs.VirtualMachine.NewStartVirtualMachineParams(vmId)
	resp, err := cs.VirtualMachine.StartVirtualMachine(p)
	if err != nil {
		// log.Printf("Failed to start vm with id %s due to %v", vmId, err)
		utils.HandleError(err)
		return nil, err
	}
	return resp, nil
}

func StopVM_cs(cs *cloudstack.CloudStackClient, vmId string) error {
	p := cs.VirtualMachine.NewStopVirtualMachineParams(vmId)
	_, err := cs.VirtualMachine.StopVirtualMachine(p)
	if err != nil {
		// log.Printf("Failed to stop vm with id %s due to %v", vmId, err)
		utils.HandleError(err)
		return nil
	}
	return nil
}

func StopVM(cs *cloudstack.CloudStackClient, vmId string) (*cloudstack.StopVirtualMachineResponse, error) {
	p := cs.VirtualMachine.NewStopVirtualMachineParams(vmId)
	resp, err := cs.VirtualMachine.StopVirtualMachine(p)
	if err != nil {
		// log.Printf("Failed to stop vm with id %s due to %v", vmId, err)
		utils.HandleError(err)
		return nil, err
	}
	return resp, nil
}

func RebootVM(cs *cloudstack.CloudStackClient, vmId string) error {
	p := cs.VirtualMachine.NewRebootVirtualMachineParams(vmId)
	_, err := cs.VirtualMachine.RebootVirtualMachine(p)
	if err != nil {
		// log.Printf("Failed to reboot vm with id %s due to %v", vmId, err)
		utils.HandleError(err)
		return err
	}
	return nil
}

func CreateVMSnapshot(cs *cloudstack.CloudStackClient, vmId string) (*cloudstack.CreateVMSnapshotResponse, error) {
	p := cs.Snapshot.NewCreateVMSnapshotParams(vmId)
	resp, err := cs.Snapshot.CreateVMSnapshot(p)
	if err != nil {
		// log.Printf("Failed to create vmsnapshot with id %s due to %v", vmId, err)
		utils.HandleError(err)
		return nil, err
	}
	return resp, nil
}

func DeleteVMSnapshot(cs *cloudstack.CloudStackClient, snapshotId string) (*cloudstack.DeleteVMSnapshotResponse, error) {
	p := cs.Snapshot.NewDeleteVMSnapshotParams(snapshotId)
	resp, err := cs.Snapshot.DeleteVMSnapshot(p)
	if err != nil {
		// log.Printf("Failed to delete vmsnapshot with id %s due to %v", snapshotId, err)
		utils.HandleError(err)
		return nil, err
	}
	return resp, nil
}

// func AllocateVbmcToVM(cs *cloudstack.CloudStackClient, vmId string) (*cloudstack.AllocateVbmcToVMResponse, error) {
// 	p := cs.VirtualMachine.NewAllocateVbmcToVMParams(vmId)
// 	resp, err := cs.VirtualMachine.AllocateVbmcToVM(p)
// 	if err != nil {
// 		// log.Printf("Failed to allocate vbmc to vm with id %s due to %v", vmId, err)
// 		utils.HandleError(err)
// 		return nil, err
// 	}
// 	return resp, nil
// }

// func RemoveVbmcToVM(cs *cloudstack.CloudStackClient, vmId string) (*cloudstack.RemoveVbmcToVMResponse, error) {
// 	p := cs.VirtualMachine.NewRemoveVbmcToVMParams(vmId)
// 	resp, err := cs.VirtualMachine.RemoveVbmcToVM(p)
// 	if err != nil {
// 		// log.Printf("Failed to remove vbmc to vm with id %s due to %v", vmId, err)
// 		utils.HandleError(err)
// 		return nil, err
// 	}
// 	return resp, nil
// }

// func CloneVirtualMachine(cs *cloudstack.CloudStackClient, vmId string) (*cloudstack.CloneVirtualMachineResponse, error) {
// 	vmName := "vm-" + utils.RandomString(10)
// 	p := cs.VirtualMachine.NewCloneVirtualMachineParams(vmId)
// 	p.SetName(vmName)
// 	p.SetType("full")
// 	p.SetStartvm(config.StartVM)
// 	p.SetCount(1)
// 	resp, err := cs.VirtualMachine.CloneVirtualMachine(p)
// 	if err != nil {
// 		// log.Printf("Failed to clonevm due to: %v", err)
// 		utils.HandleError(err)
// 		return nil, err
// 	}
// 	return resp, nil
// }

func RestoreVirtualMachine(cs *cloudstack.CloudStackClient, vmId string) (*cloudstack.RestoreVirtualMachineResponse, error) {
	p := cs.VirtualMachine.NewRestoreVirtualMachineParams(vmId)
	resp, err := cs.VirtualMachine.RestoreVirtualMachine(p)
	if err != nil {
		// log.Printf("Failed to restore to vm with id %s due to %v", vmId, err)
		utils.HandleError(err)
		return nil, err
	}
	return resp, nil
}
