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

package offering

import (
	"moldbench/utils"

	"github.com/ablecloud-team/ablestack-mold-go/v2/cloudstack"
)

func ListServiceOfferings(cs *cloudstack.CloudStackClient) ([]*cloudstack.ServiceOffering, error) {
	result := make([]*cloudstack.ServiceOffering, 0)
	page := 1
	p := cs.ServiceOffering.NewListServiceOfferingsParams()
	p.SetName("System Offering For Software Router")
	for {
		p.SetPage(page)
		resp, err := cs.ServiceOffering.ListServiceOfferings(p)
		if err != nil {
			// log.Printf("Failed to list serviceoffering due to %v", err)
			utils.HandleError(err)
			return result, err
		}
		result = append(result, resp.ServiceOfferings...)
		if len(result) < resp.Count {
			page++
		} else {
			break
		}
	}
	return result, nil
}

func CreateServiceOffering(cs *cloudstack.CloudStackClient) (*cloudstack.CreateServiceOfferingResponse, error) {
	offeringName := "ComputeOffering-" + utils.RandomString(10)
	p := cs.ServiceOffering.NewCreateServiceOfferingParams(offeringName, offeringName)
	resp, err := cs.ServiceOffering.CreateServiceOffering(p)
	if err != nil {
		// log.Printf("Failed to create computeofferings due to: %v", err)
		utils.HandleError(err)
		return nil, err
	}
	return resp, nil
}

func DeleteServiceOffering(cs *cloudstack.CloudStackClient, offeringId string) (*cloudstack.DeleteServiceOfferingResponse, error) {
	p := cs.ServiceOffering.NewDeleteServiceOfferingParams(offeringId)
	resp, err := cs.ServiceOffering.DeleteServiceOffering(p)
	if err != nil {
		// log.Printf("Failed to delete computeofferings due to: %v", err)
		utils.HandleError(err)
		return nil, err
	}
	return resp, nil
}

func CreateDiskOffering(cs *cloudstack.CloudStackClient) (*cloudstack.CreateDiskOfferingResponse, error) {
	offeringName := "DiskOffering-" + utils.RandomString(10)
	p := cs.DiskOffering.NewCreateDiskOfferingParams(offeringName, offeringName)
	p.SetCustomized(true)
	resp, err := cs.DiskOffering.CreateDiskOffering(p)
	if err != nil {
		// log.Printf("Failed to create diskofferings due to: %v", err)
		utils.HandleError(err)
		return nil, err
	}
	return resp, nil
}

func DeleteDiskOffering(cs *cloudstack.CloudStackClient, offeringId string) (*cloudstack.DeleteDiskOfferingResponse, error) {
	p := cs.DiskOffering.NewDeleteDiskOfferingParams(offeringId)
	resp, err := cs.DiskOffering.DeleteDiskOffering(p)
	if err != nil {
		// log.Printf("Failed to delete diskofferings due to: %v", err)
		utils.HandleError(err)
		return nil, err
	}
	return resp, nil
}

func CreateNetworkOffering(cs *cloudstack.CloudStackClient) (*cloudstack.CreateNetworkOfferingResponse, error) {
	offeringName := "NetworkOffering-" + utils.RandomString(10)
	p := cs.NetworkOffering.NewCreateNetworkOfferingParams(offeringName, "isolated", offeringName, "GUEST")
	p.SetInternetprotocol("ipv4")
	p.SetConservemode(true)
	p.SetEnable(true)
	p.SetSpecifyvlan(false)
	p.SetSupportedservices([]string{"Dhcp"})
	serviceprovider := map[string]string{
		"Dhcp": "VirtualRouter",
	}
	p.SetServiceproviderlist(serviceprovider)
	resp, err := cs.NetworkOffering.CreateNetworkOffering(p)
	if err != nil {
		// log.Printf("Failed to create networkoffering due to: %v", err)
		utils.HandleError(err)
		return nil, err
	}
	return resp, nil
}

func DeleteNetworkOffering(cs *cloudstack.CloudStackClient, offeringId string) (*cloudstack.DeleteNetworkOfferingResponse, error) {
	p := cs.NetworkOffering.NewDeleteNetworkOfferingParams(offeringId)
	resp, err := cs.NetworkOffering.DeleteNetworkOffering(p)
	if err != nil {
		// log.Printf("Failed to delete networkoffering due to: %v", err)
		utils.HandleError(err)
		return nil, err
	}
	return resp, nil
}
