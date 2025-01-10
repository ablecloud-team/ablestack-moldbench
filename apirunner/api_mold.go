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

package apirunner

import (
	"fmt"
	"math"
	"moldbench/config"
	"moldbench/domain"
	"moldbench/network"
	"moldbench/offering"
	"moldbench/storage"
	"moldbench/template"
	"moldbench/vm"
	"moldbench/volume"
	"os"
	"strconv"
	"time"

	"github.com/ablecloud-team/ablestack-mold-go/v2/cloudstack"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/montanaflynn/stats"
	log "github.com/sirupsen/logrus"
)

type Results struct {
	Success  bool
	Duration float64
	Id       string
}

func GenerateReport(results map[string][]*Results, order []string, format string, outputFile string) {
	fmt.Println("Generating report")

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Type", "Count", "Min", "Max", "Avg", "Median", "90th percentile", "95th percentile", "99th percentile"})

	// for key, _ := range results {
	// 	data := results[key]
	// 	// log.Infof("results ===== %s, %s", key, result)
	// 	allExecutionsSample, successfulExecutionSample, failedExecutionSample := GetSamples(data)
	// 	t.AppendRow(getRowFromSample(fmt.Sprintf("%s - All", key), allExecutionsSample))
	// 	if failedExecutionSample.Len() != 0 {
	// 		t.AppendRow(getRowFromSample(fmt.Sprintf("%s - Successful", key), successfulExecutionSample))
	// 		t.AppendRow(getRowFromSample(fmt.Sprintf("%s - Failed", key), failedExecutionSample))
	// 	}
	// }

	for _, api := range order {
		data := results[api]
		// log.Infof("results ===== %s, %s, %s", key, api, data)
		allExecutionsSample, successfulExecutionSample, failedExecutionSample := GetSamples(data)
		t.AppendRow(getRowFromSample(fmt.Sprintf("%s - All", api), allExecutionsSample))
		if failedExecutionSample.Len() != 0 {
			t.AppendRow(getRowFromSample(fmt.Sprintf("%s - Successful", api), successfulExecutionSample))
			t.AppendRow(getRowFromSample(fmt.Sprintf("%s - Failed", api), failedExecutionSample))
		}
	}

	if outputFile != "" {
		f, err := os.Create(outputFile)
		if err != nil {
			log.Error("Error creating file: ", err)
		}
		defer f.Close()
		t.SetOutputMirror(f)
	}
	switch format {
	case "csv":
		t.RenderCSV()
	case "tsv":
		t.RenderTSV()
	case "table":
		t.Render()
	}
}

func GetSamples(results []*Results) (stats.Float64Data, stats.Float64Data, stats.Float64Data) {
	var allExecutionsSample stats.Float64Data
	var successfulExecutionSample stats.Float64Data
	var failedExecutionSample stats.Float64Data

	for _, result := range results {
		duration := math.Round(result.Duration*1000) / 1000
		allExecutionsSample = append(allExecutionsSample, duration)
		if result.Success {
			successfulExecutionSample = append(successfulExecutionSample, duration)
		} else {
			failedExecutionSample = append(failedExecutionSample, duration)
		}
	}

	return allExecutionsSample, successfulExecutionSample, failedExecutionSample
}

func getRowFromSample(key string, sample stats.Float64Data) table.Row {
	min, _ := sample.Min()
	min = math.Round(min*1000) / 1000
	max, _ := sample.Max()
	max = math.Round(max*1000) / 1000
	mean, _ := sample.Mean()
	mean = math.Round(mean*1000) / 1000
	median, _ := sample.Median()
	median = math.Round(median*1000) / 1000
	percentile90, _ := sample.Percentile(90)
	percentile90 = math.Round(percentile90*1000) / 1000
	percentile95, _ := sample.Percentile(95)
	percentile95 = math.Round(percentile95*1000) / 1000
	percentile99, _ := sample.Percentile(99)
	percentile99 = math.Round(percentile99*1000) / 1000

	return table.Row{key, len(sample), min, max, mean, median, percentile90, percentile95, percentile99}
}

func CreateDomains(cs *cloudstack.CloudStackClient, parentDomainId string, count int) *Results {
	start := time.Now()
	log.Infof("Creating domain")
	dmn, err := domain.CreateDomain(cs, parentDomainId)
	if err != nil {
		return &Results{
			Success:  false,
			Duration: time.Since(start).Seconds(),
			Id:       "",
		}
	}

	return &Results{
		Success:  true,
		Duration: time.Since(start).Seconds(),
		Id:       dmn.Id + "," + dmn.Name,
	}
}

func CreateAccount(cs *cloudstack.CloudStackClient, subdomain string, count int) *Results {
	start := time.Now()
	log.Infof("Creating account")
	acn, err := domain.CreateAccount(cs, subdomain)
	if err != nil {
		return &Results{
			Success:  false,
			Duration: time.Since(start).Seconds(),
			Id:       "",
		}
	}

	return &Results{
		Success:  true,
		Duration: time.Since(start).Seconds(),
		Id:       acn.Id + "," + acn.Name,
	}
}

func CreateNetwork(cs *cloudstack.CloudStackClient, netnetworkofferingid string, vlan string, parentDomainId string, subdomain string, account string, count int) *Results {
	start := time.Now()
	log.Infof("Creating %d network", count)
	result, err := network.CreateNetwork(cs, netnetworkofferingid, vlan, parentDomainId, subdomain, account, count)
	if err != nil {
		return &Results{
			Success:  false,
			Duration: time.Since(start).Seconds(),
			Id:       "",
		}
	}

	return &Results{
		Success:  true,
		Duration: time.Since(start).Seconds(),
		Id:       result.Id,
	}
}

func CreateVms(cs *cloudstack.CloudStackClient, parentDomainId string, subdomain string, account string, networkId string, count int) *Results {
	start := time.Now()
	log.Infof("Creating %d vm", count)
	result, err := vm.DeployVm(cs, subdomain, networkId, account)
	if err != nil {
		return &Results{
			Success:  false,
			Duration: time.Since(start).Seconds(),
			Id:       "",
		}
	}
	// vm state = Running일때까지 체크
	// log.Infof("wait for running vm...")
	// time.Sleep(20 * time.Second)

	return &Results{
		Success:  true,
		Duration: time.Since(start).Seconds(),
		Id:       result.Id,
	}
}

func StartVm(cs *cloudstack.CloudStackClient, vmId string) *Results {
	start := time.Now()
	log.Infof("Starting vm")
	result, err := vm.StartVM(cs, vmId)
	if err != nil {
		return &Results{
			Success:  false,
			Duration: time.Since(start).Seconds(),
			Id:       "",
		}
	}

	return &Results{
		Success:  true,
		Duration: time.Since(start).Seconds(),
		Id:       result.Id,
	}
}

func StopVm(cs *cloudstack.CloudStackClient, vmId string) *Results {
	start := time.Now()
	log.Infof("Stopping vm")
	result, err := vm.StopVM(cs, vmId)
	// time.Sleep(10 * time.Second)
	if err != nil {
		return &Results{
			Success:  false,
			Duration: time.Since(start).Seconds(),
			Id:       "",
		}
	}

	return &Results{
		Success:  true,
		Duration: time.Since(start).Seconds(),
		Id:       result.Id,
	}
}

func CreateVmSnapshot(cs *cloudstack.CloudStackClient, vmId string) *Results {
	start := time.Now()
	log.Infof("Creating vm snapshot")
	result, err := vm.CreateVMSnapshot(cs, vmId)
	if err != nil {
		return &Results{
			Success:  false,
			Duration: time.Since(start).Seconds(),
			Id:       "",
		}
	}

	return &Results{
		Success:  true,
		Duration: time.Since(start).Seconds(),
		Id:       result.Id,
	}
}

func DeleteVmSnapshot(cs *cloudstack.CloudStackClient, snapshotId string) *Results {
	start := time.Now()
	log.Infof("Deleting vmsnapshot")
	_, err := vm.DeleteVMSnapshot(cs, snapshotId)
	if err != nil {
		return &Results{
			Success:  false,
			Duration: time.Since(start).Seconds(),
			Id:       "",
		}
	}

	return &Results{
		Success:  true,
		Duration: time.Since(start).Seconds(),
		Id:       "",
	}
}

func CreateVolumes(cs *cloudstack.CloudStackClient, subdomain string, account string, vmId string, count int) *Results {
	start := time.Now()
	log.Infof("Creating %d volume", count)
	result, err := volume.CreateVolume(cs, subdomain, account)
	if err != nil {
		return &Results{
			Success:  false,
			Duration: time.Since(start).Seconds(),
			Id:       "",
		}
	}
	// time.Sleep(3 * time.Second)

	_, err = volume.AttachVolume(cs, result.Id, vmId)
	if err != nil {
		return &Results{
			Success:  false,
			Duration: time.Since(start).Seconds(),
			Id:       "",
		}
	}

	return &Results{
		Success:  true,
		Duration: time.Since(start).Seconds(),
		Id:       result.Id,
	}
}

func CreateSnapshot(cs *cloudstack.CloudStackClient, volumeId string) *Results {
	start := time.Now()
	log.Infof("Creating snapshot")
	result, err := volume.CreateSnapshot(cs, volumeId)
	if err != nil {
		return &Results{
			Success:  false,
			Duration: time.Since(start).Seconds(),
			Id:       "",
		}
	}

	return &Results{
		Success:  true,
		Duration: time.Since(start).Seconds(),
		Id:       result.Id,
	}
}

func RegisterTemplate(cs *cloudstack.CloudStackClient, format, hypervisor, url, ostypeid, zoneid, subdomain, account string) *Results {
	start := time.Now()
	log.Infof("Creating template")
	result, err := template.RegisterTemplate(cs, format, hypervisor, url, ostypeid, zoneid, subdomain, account)
	if err != nil {
		return &Results{
			Success:  false,
			Duration: time.Since(start).Seconds(),
			Id:       "",
		}
	}
	//log.Infof("result : %s", result)
	/*	RegisterTemplate 값이 배열
		type RegisterTemplateResponse struct {
			Count            int                 `json:"count"`
			RegisterTemplate []*RegisterTemplate `json:"template"`
		}
	*/
	var template []*cloudstack.RegisterTemplate
	template = result.RegisterTemplate

	return &Results{
		Success:  true,
		Duration: time.Since(start).Seconds(),
		Id:       template[0].Id,
	}
}

func ListTemplates(cs *cloudstack.CloudStackClient, templatefilter, templateId string) *Results {
	start := time.Now()
	log.Infof("Check ready template")
	result, err := template.ListTemplates(cs, templatefilter, templateId)
	if err != nil {
		return &Results{
			Success:  false,
			Duration: time.Since(start).Seconds(),
			Id:       "",
		}
	}
	/*	Templates 값이 배열
		type ListTemplatesResponse struct {
			Count     int         `json:"count"`
			Templates []*Template `json:"template"`
		}
	*/
	var template []*cloudstack.Template
	template = result.Templates

	return &Results{
		Success:  true,
		Duration: time.Since(start).Seconds(),
		Id:       strconv.FormatBool(template[0].Isready),
	}
}

func ListTemplatesForDeleteDomain(cs *cloudstack.CloudStackClient, templatefilter, subdomain string) *Results {
	start := time.Now()
	log.Infof("Check ready template for delete domain")
	result, err := template.ListTemplatesForDeleteDomain(cs, templatefilter, subdomain)
	if err != nil {
		return &Results{
			Success:  false,
			Duration: time.Since(start).Seconds(),
			Id:       "",
		}
	}
	/*	Templates 값이 배열
		type ListTemplatesResponse struct {
			Count     int         `json:"count"`
			Templates []*Template `json:"template"`
		}
	*/
	var template []*cloudstack.Template
	template = result.Templates
	downloadState := template[0].Downloaddetails[0]["downloadState"]
	return &Results{
		Success:  true,
		Duration: time.Since(start).Seconds(),
		Id:       downloadState,
	}
}

func DeleteSnapshot(cs *cloudstack.CloudStackClient, snapshotId string) *Results {
	start := time.Now()
	log.Infof("Deleting snapshot")
	_, err := volume.DeleteSnapshot(cs, snapshotId)
	if err != nil {
		return &Results{
			Success:  false,
			Duration: time.Since(start).Seconds(),
			Id:       "",
		}
	}

	return &Results{
		Success:  true,
		Duration: time.Since(start).Seconds(),
		Id:       "",
	}
}

func DetachVolume(cs *cloudstack.CloudStackClient, volumeId string) *Results {
	start := time.Now()
	log.Infof("Detaching volume")
	result, err := volume.DetachVolume(cs, volumeId)
	if err != nil {
		return &Results{
			Success:  false,
			Duration: time.Since(start).Seconds(),
			Id:       "",
		}
	}

	return &Results{
		Success:  true,
		Duration: time.Since(start).Seconds(),
		Id:       result.Id,
	}
}

func DestroyVolume(cs *cloudstack.CloudStackClient, volumeId string) *Results {
	start := time.Now()
	log.Infof("Destroy volume")
	result, err := volume.DestroyVolume(cs, volumeId)
	if err != nil {
		return &Results{
			Success:  false,
			Duration: time.Since(start).Seconds(),
			Id:       "",
		}
	}

	return &Results{
		Success:  true,
		Duration: time.Since(start).Seconds(),
		Id:       result.Id,
	}
}

func DeleteTemplate(cs *cloudstack.CloudStackClient, subdomain string) *Results {
	start := time.Now()
	log.Infof("Deleting template")
	i := 1
	for {
		var res *Results
		res = ListTemplatesForDeleteDomain(cs, config.TemplateFilter, subdomain)
		if res.Id != "DOWNLOAD_IN_PROGRESS" && res.Id != "UPLOAD_IN_PROGRESS" && res.Id != "NOT_DOWNLOADED" {
			log.Infof("Check Template can be deleted")
			break
		} else {
			log.Infof("Template cannot be deleted because it is being uploaded.(try: %d)", i)
			time.Sleep(10 * time.Second)
			i++

			if i == 10 {
				break
			}
		}
	}

	return &Results{
		Success:  true,
		Duration: time.Since(start).Seconds(),
		Id:       "",
	}
}

func DestroyVm(cs *cloudstack.CloudStackClient, vmId string) *Results {
	start := time.Now()
	log.Infof("Destroy vm")
	result, err := vm.DestroyVm(cs, vmId)
	if err != nil {
		return &Results{
			Success:  false,
			Duration: time.Since(start).Seconds(),
			Id:       "",
		}
	}

	return &Results{
		Success:  true,
		Duration: time.Since(start).Seconds(),
		Id:       result.Id,
	}
}

func DeleteNetwork(cs *cloudstack.CloudStackClient, networkId string) *Results {
	start := time.Now()
	log.Infof("Deleting network")
	_, err := network.DeleteNetwork(cs, networkId)
	if err != nil {
		return &Results{
			Success:  false,
			Duration: time.Since(start).Seconds(),
			Id:       "",
		}
	}

	return &Results{
		Success:  true,
		Duration: time.Since(start).Seconds(),
		Id:       "",
	}
}

func DeleteAccount(cs *cloudstack.CloudStackClient, accountId string) *Results {
	start := time.Now()
	log.Infof("Deleting account")
	_, err := domain.DeleteAccount(cs, accountId)
	if err != nil {
		return &Results{
			Success:  false,
			Duration: time.Since(start).Seconds(),
			Id:       "",
		}
	}

	return &Results{
		Success:  true,
		Duration: time.Since(start).Seconds(),
		Id:       "",
	}
}

func DeleteDomain(cs *cloudstack.CloudStackClient, domainId string, flag bool) *Results {
	start := time.Now()
	log.Infof("Deleting domain")
	// flag true=delete domain, false=error delete domain
	if !flag {
		i := 1
		for {
			var res *Results
			res = ListTemplatesForDeleteDomain(cs, config.TemplateFilter, domainId)
			if res.Id != "DOWNLOAD_IN_PROGRESS" && res.Id != "UPLOAD_IN_PROGRESS" && res.Id != "NOT_DOWNLOADED" {
				log.Infof("Check Template can be deleted")
				break
			} else {
				log.Infof("Template cannot be deleted because it is being uploaded.(try: %d)", i)
				time.Sleep(10 * time.Second)
				i++

				if i == 10 {
					break
				}
			}
		}
	}

	_, err := domain.DeleteDomain(cs, domainId)
	if err != nil {
		return &Results{
			Success:  false,
			Duration: time.Since(start).Seconds(),
			Id:       "",
		}
	}

	return &Results{
		Success:  true,
		Duration: time.Since(start).Seconds(),
		Id:       "",
	}
}

func CreateServiceOffering(cs *cloudstack.CloudStackClient, count int) *Results {
	start := time.Now()
	log.Infof("Creating %d computeofferings", count)
	result, err := offering.CreateServiceOffering(cs)
	if err != nil {
		return &Results{
			Success:  false,
			Duration: time.Since(start).Seconds(),
			Id:       "",
		}
	}

	return &Results{
		Success:  true,
		Duration: time.Since(start).Seconds(),
		Id:       result.Id,
	}
}

func DeleteServiceOffering(cs *cloudstack.CloudStackClient, offeringId string, count int) *Results {
	start := time.Now()
	log.Infof("Deleting %d computeofferings", count)
	_, err := offering.DeleteServiceOffering(cs, offeringId)
	if err != nil {
		return &Results{
			Success:  false,
			Duration: time.Since(start).Seconds(),
			Id:       "",
		}
	}

	return &Results{
		Success:  true,
		Duration: time.Since(start).Seconds(),
		Id:       "",
	}
}

func CreateDiskOffering(cs *cloudstack.CloudStackClient, count int) *Results {
	start := time.Now()
	log.Infof("Creating %d diskofferings", count)
	result, err := offering.CreateDiskOffering(cs)
	if err != nil {
		return &Results{
			Success:  false,
			Duration: time.Since(start).Seconds(),
			Id:       "",
		}
	}

	return &Results{
		Success:  true,
		Duration: time.Since(start).Seconds(),
		Id:       result.Id,
	}
}

func DeleteDiskOffering(cs *cloudstack.CloudStackClient, offeringId string, count int) *Results {
	start := time.Now()
	log.Infof("Deleting %d diskofferings", count)
	_, err := offering.DeleteDiskOffering(cs, offeringId)
	if err != nil {
		return &Results{
			Success:  false,
			Duration: time.Since(start).Seconds(),
			Id:       "",
		}
	}

	return &Results{
		Success:  true,
		Duration: time.Since(start).Seconds(),
		Id:       "",
	}
}

func CreateNetworkOffering(cs *cloudstack.CloudStackClient, count int) *Results {
	start := time.Now()
	log.Infof("Creating %d networkofferings", count)
	result, err := offering.CreateNetworkOffering(cs)
	if err != nil {
		return &Results{
			Success:  false,
			Duration: time.Since(start).Seconds(),
			Id:       "",
		}
	}

	return &Results{
		Success:  true,
		Duration: time.Since(start).Seconds(),
		Id:       result.Id,
	}
}

func DeleteNetworkOffering(cs *cloudstack.CloudStackClient, offeringId string, count int) *Results {
	start := time.Now()
	log.Infof("Deleting %d networkofferings", count)
	_, err := offering.DeleteNetworkOffering(cs, offeringId)
	if err != nil {
		return &Results{
			Success:  false,
			Duration: time.Since(start).Seconds(),
			Id:       "",
		}
	}

	return &Results{
		Success:  true,
		Duration: time.Since(start).Seconds(),
		Id:       "",
	}
}

func CreateStoragePool(cs *cloudstack.CloudStackClient, name, url, zoneid, pod_id, cluster_id string) *Results {
	start := time.Now()
	log.Infof("Creating Storage Pool %s %s %s %s %s", name, url, zoneid, pod_id, cluster_id)
	result, err := storage.CreateStoragePool(cs, name, url, zoneid, pod_id, cluster_id)
	if err != nil {
		return &Results{
			Success:  false,
			Duration: time.Since(start).Seconds(),
			Id:       "",
		}
	}

	return &Results{
		Success:  true,
		Duration: time.Since(start).Seconds(),
		Id:       result.Id,
	}
}
