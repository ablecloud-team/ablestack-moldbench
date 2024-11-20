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

package domain

import (
	"moldbench/config"
	"moldbench/utils"

	"github.com/ablecloud-team/ablestack-mold-go/v2/cloudstack"
)

func CreateDomain(cs *cloudstack.CloudStackClient, parentDomainId string) (*cloudstack.CreateDomainResponse, error) {
	domainName := "Domain-" + utils.RandomString(10)
	p := cs.Domain.NewCreateDomainParams(domainName)
	p.SetParentdomainid(parentDomainId)
	resp, err := cs.Domain.CreateDomain(p)
	if err != nil {
		// log.Printf("Failed to create domain due to: %v", err)
		utils.HandleError(err)
		return nil, err
	}
	return resp, nil
}

func DeleteDomain_cs(cs *cloudstack.CloudStackClient, domainId string) (bool, error) {
	deleteParams := cs.Domain.NewDeleteDomainParams(domainId)
	deleteParams.SetCleanup(true)
	delResp, err := cs.Domain.DeleteDomain(deleteParams)
	if err != nil {
		// log.Printf("Failed to delete domain with id  %s due to %v", domainId, err)
		utils.HandleError(err)
		return delResp.Success, err
	}
	return delResp.Success, nil
}

func DeleteDomain(cs *cloudstack.CloudStackClient, domainId string) (*cloudstack.DeleteDomainResponse, error) {
	deleteParams := cs.Domain.NewDeleteDomainParams(domainId)
	deleteParams.SetCleanup(true)
	resp, err := cs.Domain.DeleteDomain(deleteParams)
	if err != nil {
		// log.Printf("Failed to delete domain with id  %s due to %v", domainId, err)
		utils.HandleError(err)
		return nil, err
	}
	return resp, nil
}

func CreateAccount(cs *cloudstack.CloudStackClient, domainId string) (*cloudstack.CreateAccountResponse, error) {
	accountName := "Account-" + utils.RandomString(10)
	p := cs.Account.NewCreateAccountParams("test@test", accountName, "Account", "Ablecloud1!", accountName)
	p.SetDomainid(domainId)
	p.SetAccounttype(2)
	resp, err := cs.Account.CreateAccount(p)
	if err != nil {
		// log.Printf("Failed to create account due to: %v", err)
		utils.HandleError(err)
		return nil, err
	}
	return resp, nil
}

func ListSubDomains(cs *cloudstack.CloudStackClient, domainId string) []*cloudstack.DomainChildren {
	result := make([]*cloudstack.DomainChildren, 0)
	page := 1
	p := cs.Domain.NewListDomainChildrenParams()
	p.SetId(domainId)
	p.SetPagesize(config.PageSize)
	for {
		p.SetPage(page)
		resp, err := cs.Domain.ListDomainChildren(p)
		if err != nil {
			// log.Printf("Failed to list domains due to: %v", err)
			utils.HandleError(err)
			return result
		}
		result = append(result, resp.DomainChildren...)
		if len(result) < resp.Count {
			page++
		} else {
			break
		}
	}
	return result
}

func ListAccounts(cs *cloudstack.CloudStackClient, domainId string) []*cloudstack.Account {
	result := make([]*cloudstack.Account, 0)
	page := 1
	p := cs.Account.NewListAccountsParams()
	p.SetDomainid(domainId)
	p.SetPagesize(config.PageSize)
	for {
		p.SetPage(page)
		resp, err := cs.Account.ListAccounts(p)
		if err != nil {
			// log.Printf("Failed to list accounts due to: %v", err)
			utils.HandleError(err)
			return result
		}
		result = append(result, resp.Accounts...)
		if len(result) < resp.Count {
			page++
		} else {
			break
		}
	}
	return result
}

func UpdateLimits(cs *cloudstack.CloudStackClient, account *cloudstack.Account) bool {
	for i := 0; i <= 11; i++ {
		p := cs.Limit.NewUpdateResourceLimitParams(i)
		p.SetDomainid(account.Domainid)
		p.SetMax(-1)
		_, err := cs.Limit.UpdateResourceLimit(p)
		if err != nil {
			// log.Printf("Failed to update domain's resource limit due to: %v", err)
			utils.HandleError(err)
			return false
		}
		p.SetAccount(account.Name)
		_, err = cs.Limit.UpdateResourceLimit(p)
		if err != nil {
			// log.Printf("Failed to update account's resource limit due to: %v", err)
			utils.HandleError(err)
			return false
		}
	}
	return true
}

func DeleteAccount(cs *cloudstack.CloudStackClient, accountId string) (*cloudstack.DeleteAccountResponse, error) {
	deleteParams := cs.Account.NewDeleteAccountParams(accountId)
	resp, err := cs.Account.DeleteAccount(deleteParams)
	if err != nil {
		// log.Printf("Failed to delete account with id  %s due to %v", accountId, err)
		utils.HandleError(err)
		return nil, err
	}
	return resp, nil
}
