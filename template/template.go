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

package template

import (
	"moldbench/utils"

	"github.com/ablecloud-team/ablestack-mold-go/v2/cloudstack"
)

func ListTemplates(cs *cloudstack.CloudStackClient, templatefilter, templateId string) (*cloudstack.ListTemplatesResponse, error) {
	p := cs.Template.NewListTemplatesParams(templatefilter)
	p.SetId(templateId)
	resp, err := cs.Template.ListTemplates(p)
	if err != nil {
		// log.Printf("Failed to list template due to: %v", err)
		utils.HandleError(err)
		return nil, err
	}
	return resp, nil
}

func ListTemplatesForDeleteDomain(cs *cloudstack.CloudStackClient, templatefilter, subdomain string) (*cloudstack.ListTemplatesResponse, error) {
	p := cs.Template.NewListTemplatesParams(templatefilter)
	p.SetName("Template-" + subdomain)
	p.SetDomainid(subdomain)
	p.SetDomainid(subdomain)
	resp, err := cs.Template.ListTemplates(p)
	if err != nil {
		// log.Printf("Failed to list template due to: %v", err)
		utils.HandleError(err)
		return nil, err
	}
	return resp, nil
}

func CreateTemplate(cs *cloudstack.CloudStackClient, ostypeid string, snapshotId string) (*cloudstack.CreateTemplateResponse, error) {
	temName := "Template-" + utils.RandomString(10)
	p := cs.Template.NewCreateTemplateParams(temName, temName, ostypeid)
	p.SetSnapshotid(snapshotId)
	resp, err := cs.Template.CreateTemplate(p)
	if err != nil {
		// log.Printf("Failed to create template due to: %v", err)
		utils.HandleError(err)
		return nil, err
	}
	return resp, nil
}

func RegisterTemplate(cs *cloudstack.CloudStackClient, format, hypervisor, url, ostypeid, zoneid, domainId, account string) (*cloudstack.RegisterTemplateResponse, error) {
	// temName := "Template-" + utils.RandomString(10)
	temName := "Template-" + domainId
	p := cs.Template.NewRegisterTemplateParams(temName, format, hypervisor, temName, url)
	p.SetOstypeid(ostypeid)
	p.SetZoneid(zoneid)
	p.SetRequireshvm(true)
	p.SetDirectdownload(false)
	p.SetDomainid(domainId)
	p.SetAccount(account)
	resp, err := cs.Template.RegisterTemplate(p)
	if err != nil {
		// log.Printf("Failed to register template due to: %v", err)
		utils.HandleError(err)
		return nil, err
	}
	return resp, nil
}

func DeleteTemplate(cs *cloudstack.CloudStackClient, templateId string) (*cloudstack.DeleteTemplateResponse, error) {
	p := cs.Template.NewDeleteTemplateParams(templateId)
	resp, err := cs.Template.DeleteTemplate(p)
	if err != nil {
		// log.Printf("Failed to delete template due to: %v", err)
		utils.HandleError(err)
		return nil, err
	}
	return resp, nil
}

/*
func GetUploadParamsForTemplate(cs *cloudstack.CloudStackClient, format, hypervisor, zoneid, ostypeid string) (*cloudstack.CreateTemplateResponse, error) {
	temName := "Template-" + utils.RandomString(10)
	p := cs.Template.NewCreateTemplateParams(temName, format, hypervisor, temName, zoneid)
	p.SetOstypeid(ostypeid)
	resp, err := cs.Template.GetUploadParamsForTemplate(p)
	if err != nil {
		log.Printf("Failed to create template due to: %v", err)
		return nil, err
	}
	return resp, nil
}
*/
