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

package main

import (
	"flag"
	"fmt"
	"io"
	"math"
	"math/rand"
	"moldbench/apirunner"
	"moldbench/config"
	"moldbench/domain"
	"moldbench/network"
	"moldbench/vm"
	"moldbench/volume"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/ablecloud-team/ablestack-mold-go/v2/cloudstack"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/montanaflynn/stats"
	"github.com/sourcegraph/conc/pool"
)

var (
	profiles = make(map[int]*config.Profile)
)

type Result struct {
	Success  bool
	Duration float64
}

func init() {
	logFile, err := os.OpenFile("csmetrics.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to create log file: %v", err)
	}

	mw := io.MultiWriter(os.Stdout, logFile)

	log.SetOutput(mw)
}

func readConfigurations(configFile string) map[int]*config.Profile {
	profiles, err := config.ReadProfiles(configFile)
	if err != nil {
		log.Fatal("Error reading profiles:", err)
	}

	return profiles
}

func logConfigurationDetails(profiles map[int]*config.Profile) {
	apiURL := config.URL
	iterations := config.Iterations
	page := config.Page
	pagesize := config.PageSize
	host := config.Host

	userProfileNames := make([]string, 0, len(profiles))
	for _, profile := range profiles {
		userProfileNames = append(userProfileNames, profile.Name)
	}

	fmt.Printf("\n\n\033[1;34mBenchmarking the CloudStack environment [%s] with the following configuration\033[0m\n\n", apiURL)
	fmt.Printf("Management server : %s\n", host)
	fmt.Printf("Roles : %s\n", strings.Join(userProfileNames, ","))
	fmt.Printf("Iterations : %d\n", iterations)
	fmt.Printf("Page : %d\n", page)
	fmt.Printf("PageSize : %d\n\n", pagesize)

	log.Infof("Found %d profiles in the configuration: ", len(profiles))
	log.Infof("Management server : %s", host)
}

func logReport() {
	fmt.Printf("\n\n\nLog file : csmetrics.log\n")
	fmt.Printf("Reports directory per API : report/%s/\n", config.Host)
	fmt.Printf("Number of APIs : %d\n", apirunner.APIscount)
	fmt.Printf("Successful APIs : %d\n", apirunner.SuccessAPIs)
	fmt.Printf("Failed APIs : %d\n", apirunner.FailedAPIs)
	fmt.Printf("Time in seconds per API: %.2f (avg)\n", apirunner.TotalTime/float64(apirunner.APIscount))
	fmt.Printf("\n\n\033[1;34m--------------------------------------------------------------------------------\033[0m\n" +
		"                            Done with benchmarking\n" +
		"\033[1;34m--------------------------------------------------------------------------------\033[0m\n\n")
}

func getSamples(results []*Result) (stats.Float64Data, stats.Float64Data, stats.Float64Data) {
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

/*
This function will generate a report with the following details:
 1. Total Number of executions
 2. Number of successful executions
 3. Number of failed exections
 4. Different statistics like min, max, avg, median, 90th percentile, 95th percentile, 99th percentile for above 3

Output format:
 1. CSV
 2. TSV
 3. Table
*/
func generateReport(results map[string][]*Result, format string, outputFile string) {
	fmt.Println("Generating report")

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Type", "Count", "Min", "Max", "Avg", "Median", "90th percentile", "95th percentile", "99th percentile"})

	for key, result := range results {
		allExecutionsSample, successfulExecutionSample, failedExecutionSample := getSamples(result)
		t.AppendRow(getRowFromSample(fmt.Sprintf("%s - All", key), allExecutionsSample))
		if failedExecutionSample.Len() != 0 {
			t.AppendRow(getRowFromSample(fmt.Sprintf("%s - Successful", key), successfulExecutionSample))
			t.AppendRow(getRowFromSample(fmt.Sprintf("%s - Failed", key), failedExecutionSample))
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

func main() {
	license := flag.Bool("license", false, "Test Mold APIs")
	simpletest := flag.Bool("simpletest", false, "Mold simple test APIs")
	dbprofile := flag.Int("dbprofile", 0, "DB profile number")
	create := flag.Bool("create", false, "Create resources. Specify at least one of the following options:\n\t"+
		"-domain - Create subdomains and accounts\n\t"+
		"-limits - Update limits to -1 for subdomains and accounts\n\t"+
		"-network - Create shared network in all subdomains\n\t"+
		"-vm - Deploy VMs in all networks in the subdomains\n\t"+
		"-volume - Create and attach Volumes to VMs")
	benchmark := flag.Bool("benchmark", false, "Benchmark list APIs")
	domainFlag := flag.Bool("domain", false, "Works with -create & -teardown\n\t"+
		"-create - Create subdomains and accounts\n\t"+
		"-teardown - Delete all subdomains and accounts")
	limitsFlag := flag.Bool("limits", false, "Update limits to -1 for subdomains and accounts")
	networkFlag := flag.Bool("network", false, "Works with -create & -teardown\n\t"+
		"-create - Create shared network in all subdomains\n\t"+
		"-teardown - Delete all networks in the subdomains")
	vmFlag := flag.Bool("vm", false, "Works with -create & -teardown\n\t"+
		"-create - Deploy VMs in all networks in the subdomains\n\t"+
		"-teardown - Delete all VMs in the subdomains")
	volumeFlag := flag.Bool("volume", false, "Works with -create & -teardown\n\t"+
		"-create - Create and attach Volumes to VMs\n\t"+
		"-teardown - Delete all volumes in the subdomains")
	vmAction := flag.String("vmaction", "", "Action to perform on VMs. Options:\n\t"+
		"start - start all VMs\n\t"+
		"stop - stop all VMs\n\t"+
		"reboot - reboot all running VMs\n\t"+
		"toggle - stop running VMs and start stopped VMs\n\t"+
		"random - Randomly toggle VMs")
	tearDown := flag.Bool("teardown", false, "Tear down resources. Specify at least one of the following options:\n\t"+
		"-domain - Delete all subdomains and accounts\n\t"+
		"-network - Delete all networks in the subdomains\n\t"+
		"-vm - Delete all VMs in the subdomains\n\t"+
		"-volume - Delete all volumes in the subdomains")
	workers := flag.Int("workers", 10, "Number of workers to use while creating resources")
	format := flag.String("format", "table", "Format of the report (csv, tsv, table). Valid only for create")
	outputFile := flag.String("output", "", "Path to output file. Valid only for create")
	configFile := flag.String("config", "config/config", "Path to config file")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if !(*license || *simpletest || *create || *benchmark || *tearDown || *vmAction != "") {
		log.Fatal("Please provide one of the following options: -simpletest, -create, -benchmark, -vmaction, -teardown")
	}

	if *create && *tearDown && *vmAction == "" {
		log.Fatal("Please provide only one of the following options: -create, -teardown, -vmaction")
	}

	if *vmAction != "" && !(*vmAction == "start" || *vmAction == "stop" || *vmAction == "reboot" || *vmAction == "toggle" || *vmAction == "random") {
		log.Fatal("Invalid VM action. Please provide one of the following: start, stop, reboot, toggle, random")
	}

	if *create && !(*domainFlag || *limitsFlag || *networkFlag || *vmFlag || *volumeFlag) {
		log.Fatal("Please provide one of the following options with create: -domain, -limits, -network, -vm, -volume")
	}

	if *tearDown && !(*domainFlag || *networkFlag || *vmFlag || *volumeFlag) {
		log.Fatal("Please provide one of the following options with teardown: -domain, -network, -vm, -volume")
	}

	switch *format {
	case "csv", "tsv", "table":
		// valid format, continue
	default:
		log.Fatal("Invalid format. Please provide one of the following: csv, tsv, table")
	}

	if *dbprofile < 0 {
		log.Fatal("Invalid DB profile number. Please provide a positive integer.")
	}

	profiles = readConfigurations(*configFile)
	apiURL := config.URL
	iterations := config.Iterations
	page := config.Page
	pagesize := config.PageSize

	if *simpletest {
		fmt.Printf("\n\n\033[1;34mBenchmarking the CloudStack environment [%s] with the following configuration\033[0m\n\n", apiURL)
		fmt.Printf("Management server : %s\n", config.Host)

		results, order := createResources_simpletest()
		apirunner.GenerateReport(results, order, *format, *outputFile)
	}

	if *create {
		results := createResources(domainFlag, limitsFlag, networkFlag, vmFlag, volumeFlag, workers)
		generateReport(results, *format, *outputFile)
	}

	if *vmAction != "" {
		results := executeVMAction(vmAction, workers)
		generateReport(results, *format, *outputFile)
	}

	if *tearDown {
		results := tearDownEnv(domainFlag, networkFlag, vmFlag, volumeFlag, workers)
		generateReport(results, *format, *outputFile)
	}

	if *benchmark {
		log.Infof("\nStarted benchmarking the CloudStack environment [%s]", apiURL)

		logConfigurationDetails(profiles)

		for i, profile := range profiles {
			userProfileName := profile.Name
			log.Infof("Using profile %d.%s for benchmarking", i, userProfileName)
			fmt.Printf("\n\033[1;34m============================================================\033[0m\n")
			fmt.Printf("                    Profile: [%s]\n", userProfileName)
			fmt.Printf("\033[1;34m============================================================\033[0m\n")
			apirunner.RunAPIs(userProfileName, apiURL, profile.ApiKey, profile.SecretKey, profile.Expires, profile.SignatureVersion, iterations, page, pagesize, *dbprofile)
		}
		logReport()

		log.Infof("Done with benchmarking the CloudStack environment [%s]", apiURL)
	}

}

func executeVMAction(vmAction *string, workers *int) map[string][]*Result {

	parentDomainId := config.ParentDomainId
	var cs *cloudstack.CloudStackClient
	workerPool := pool.NewWithResults[map[string]*Result]().WithMaxGoroutines(*workers)
	for _, profile := range profiles {
		if profile.Name == "admin" {
			cs = cloudstack.NewAsyncClient(config.URL, profile.ApiKey, profile.SecretKey, false)
			cs.Timeout(time.Duration(300 * time.Second))
		}
	}

	if cs == nil {
		log.Fatal("Failed to find admin profile")
	}

	log.Infof("Fetching all VMs in subdomains for domain %s", parentDomainId)
	domains := domain.ListSubDomains(cs, config.ParentDomainId)
	var allVMs []*cloudstack.VirtualMachine
	for _, dmn := range domains {
		vms, err := vm.ListVMs(cs, dmn.Id)
		if err != nil {
			log.Warn("Error listing VMs: ", err)
			continue
		}
		allVMs = append(allVMs, vms...)
	}

	progressMarker := int(math.Max(float64(len(allVMs))/10.0, 5))
	start := time.Now()

	for i, virtualMachine := range allVMs {
		virtualMachine := virtualMachine

		if (i+1)%progressMarker == 0 {
			log.Infof("Executed %d VMs", i+1)
		}

		if *vmAction == "random" && rand.Intn(100) < 50 {
			continue
		}

		workerPool.Go(func() map[string]*Result {
			taskStart := time.Now()
			result := false
			action := "skipped"
			switch virtualMachine.State {
			case "Running":
				if *vmAction == "stop" || *vmAction == "toggle" || *vmAction == "random" {
					err := vm.StopVM_cs(cs, virtualMachine.Id)
					result = err == nil
					action = "stop"
				} else if *vmAction == "reboot" {
					err := vm.RebootVM(cs, virtualMachine.Id)
					result = err == nil
					action = "reboot"
				}
			case "Stopped":
				if *vmAction == "start" || *vmAction == "toggle" || *vmAction == "random" {
					err := vm.StartVM_cs(cs, virtualMachine.Id)
					result = err == nil
					action = "start"
				} else if *vmAction == "reboot" {
					result = false
					action = "stop"
				}
			}
			return map[string]*Result{
				action: {
					Success:  result,
					Duration: time.Since(taskStart).Seconds(),
				},
			}
		})
	}
	res := workerPool.Wait()
	log.Infof("Executed %s on %d VMs in %.2f seconds", *vmAction, len(allVMs), time.Since(start).Seconds())
	var results = make(map[string][]*Result)
	for _, result := range res {
		for key, value := range result {
			key = "vmaction-" + key
			if results[key] == nil {
				results[key] = make([]*Result, 0)
			}
			results[key] = append(results[key], value)
		}
	}
	return results
}

func createResources(domainFlag, limitsFlag, networkFlag, vmFlag, volumeFlag *bool, workers *int) map[string][]*Result {
	apiURL := config.URL

	for _, profile := range profiles {
		if profile.Name == "admin" {

			numNetworksPerDomain := config.NumNetworks
			numVmsPerNetwork := config.NumVms
			numVolumesPerVM := config.NumVolumes

			cs := cloudstack.NewAsyncClient(apiURL, profile.ApiKey, profile.SecretKey, false)

			var results = make(map[string][]*Result)

			if *domainFlag {
				workerPool := pool.NewWithResults[*Result]().WithMaxGoroutines(*workers)
				if config.NumDomains > 0 {
					results["domain"] = createDomains(workerPool, cs, config.ParentDomainId, config.NumDomains)
				} else {
					log.Warn("Number of domains (numdomains) is less than 1. Skipping domain creation")
				}
			}

			if *limitsFlag {
				workerPool := pool.NewWithResults[*Result]().WithMaxGoroutines(*workers)
				results["limits"] = updateLimits(workerPool, cs, config.ParentDomainId)
			}

			if *networkFlag {
				workerPool := pool.NewWithResults[*Result]().WithMaxGoroutines(*workers)
				if numNetworksPerDomain > 0 {
					results["network"] = createNetwork(workerPool, cs, config.ParentDomainId, numNetworksPerDomain)
				} else {
					log.Warn("Number of networks per domain (numnetworks) is less than 1. Skipping network creation")
				}
			}

			if *vmFlag {
				workerPool := pool.NewWithResults[*Result]().WithMaxGoroutines(*workers)
				if numVmsPerNetwork > 0 {
					results["vm"] = createVms(workerPool, cs, config.ParentDomainId, numVmsPerNetwork)
				} else {
					log.Warn("Number of VMs per network (numvms) is less than 1. Skipping VM creation")
				}
			}

			if *volumeFlag {
				workerPool := pool.NewWithResults[*Result]().WithMaxGoroutines(*workers)
				if numVolumesPerVM > 0 {
					results["volume"] = createVolumes(workerPool, cs, config.ParentDomainId, numVolumesPerVM)
				} else {
					log.Warn("Number of volumes per VM (numvolumes) is less than 1. Skipping volume creation")
				}
			}

			return results
		}
	}
	return nil
}

func createDomains(workerPool *pool.ResultPool[*Result], cs *cloudstack.CloudStackClient, parentDomainId string, count int) []*Result {
	progressMarker := int(math.Max(float64(count)/10.0, 5))
	start := time.Now()
	log.Infof("Creating %d domains", count)
	for i := 0; i < count; i++ {
		if (i+1)%progressMarker == 0 {
			log.Infof("Created %d domains", i+1)
		}
		workerPool.Go(func() *Result {
			taskStart := time.Now()
			dmn, err := domain.CreateDomain(cs, parentDomainId)
			if err != nil {
				return &Result{
					Success:  false,
					Duration: time.Since(taskStart).Seconds(),
				}
			}
			_, err = domain.CreateAccount(cs, dmn.Id)
			if err != nil {
				return &Result{
					Success:  false,
					Duration: time.Since(taskStart).Seconds(),
				}
			}

			return &Result{
				Success:  true,
				Duration: time.Since(taskStart).Seconds(),
			}
		})
	}
	res := workerPool.Wait()
	log.Infof("Created %d domains in %.2f seconds", count, time.Since(start).Seconds())
	return res
}

func updateLimits(workerPool *pool.ResultPool[*Result], cs *cloudstack.CloudStackClient, parentDomainId string) []*Result {
	log.Infof("Fetching subdomains for domain %s", parentDomainId)
	domains := domain.ListSubDomains(cs, parentDomainId)
	accounts := make([]*cloudstack.Account, 0)
	for _, dmn := range domains {
		accounts = append(accounts, domain.ListAccounts(cs, dmn.Id)...)
	}

	progressMarker := int(math.Max(float64(len(accounts))/10.0, 5))
	start := time.Now()
	log.Infof("Updating limits for %d accounts", len(accounts))
	for i, account := range accounts {
		if (i+1)%progressMarker == 0 {
			log.Infof("Updated limits for %d accounts", i+1)
		}
		account := account
		workerPool.Go(func() *Result {
			taskStart := time.Now()
			resp := domain.UpdateLimits(cs, account)
			return &Result{
				Success:  resp,
				Duration: time.Since(taskStart).Seconds(),
			}
		})
	}
	res := workerPool.Wait()
	log.Infof("Updated limits for %d accounts in %.2f seconds", len(accounts), time.Since(start).Seconds())
	return res
}

func createNetwork(workerPool *pool.ResultPool[*Result], cs *cloudstack.CloudStackClient, parentDomainId string, numNetworkPerDomain int) []*Result {
	log.Infof("Fetching subdomains for domain %s", parentDomainId)
	domains := domain.ListSubDomains(cs, parentDomainId)

	progressMarker := int(math.Max(float64(len(domains)*numNetworkPerDomain)/10.0, 5))
	counter := 0
	start := time.Now()
	log.Infof("Creating %d networks", len(domains)*numNetworkPerDomain)
	for _, dmn := range domains {
		for j := 1; j <= numNetworkPerDomain; j++ {
			networkIdx := counter
			dmn := dmn
			counter++
			if counter%progressMarker == 0 {
				log.Infof("Created %d networks", counter)
			}

			workerPool.Go(func() *Result {
				taskStart := time.Now()
				_, err := network.CreateNetwork_cs(cs, dmn.Id, networkIdx)
				if err != nil {
					return &Result{
						Success:  false,
						Duration: time.Since(taskStart).Seconds(),
					}
				}
				return &Result{
					Success:  true,
					Duration: time.Since(taskStart).Seconds(),
				}
			})
		}
	}
	res := workerPool.Wait()
	log.Infof("Created %d networks in %.2f seconds", len(domains)*numNetworkPerDomain, time.Since(start).Seconds())
	return res
}

func createVms(workerPool *pool.ResultPool[*Result], cs *cloudstack.CloudStackClient, parentDomainId string, numVmPerNetwork int) []*Result {
	log.Infof("Fetching subdomains & accounts for domain %s", parentDomainId)
	domains := domain.ListSubDomains(cs, parentDomainId)
	var accounts []*cloudstack.Account
	for i := 0; i < len(domains); i++ {
		account := domain.ListAccounts(cs, domains[i].Id)
		accounts = append(accounts, account...)
	}

	domainIdAccountMapping := make(map[string]*cloudstack.Account)
	for _, account := range accounts {
		domainIdAccountMapping[account.Domainid] = account
	}

	log.Infof("Fetching networks for subdomains in domain %s", parentDomainId)
	var allNetworks []*cloudstack.Network
	for _, domain := range domains {
		network, _ := network.ListNetworks(cs, domain.Id)
		allNetworks = append(allNetworks, network...)
	}

	progressMarker := int(math.Max(float64(len(allNetworks)*numVmPerNetwork)/10.0, 5))
	counter := 0
	start := time.Now()
	log.Infof("Creating %d VMs", len(allNetworks)*numVmPerNetwork)
	for _, network := range allNetworks {
		network := network
		for j := 1; j <= numVmPerNetwork; j++ {
			counter++
			if counter%progressMarker == 0 {
				log.Infof("Created %d VMs", counter)
			}
			workerPool.Go(func() *Result {
				taskStart := time.Now()
				_, err := vm.DeployVm(cs, network.Domainid, network.Id, domainIdAccountMapping[network.Domainid].Name)
				if err != nil {
					return &Result{
						Success:  false,
						Duration: time.Since(taskStart).Seconds(),
					}
				}
				return &Result{
					Success:  true,
					Duration: time.Since(taskStart).Seconds(),
				}
			})
		}
	}
	res := workerPool.Wait()
	log.Infof("Created %d VMs in %.2f seconds", len(allNetworks)*numVmPerNetwork, time.Since(start).Seconds())
	return res
}

func createVolumes(workerPool *pool.ResultPool[*Result], cs *cloudstack.CloudStackClient, parentDomainId string, numVolumesPerVM int) []*Result {
	log.Infof("Fetching all VMs in subdomains for domain %s", parentDomainId)
	domains := domain.ListSubDomains(cs, parentDomainId)
	var allVMs []*cloudstack.VirtualMachine
	for _, dmn := range domains {
		vms, err := vm.ListVMs(cs, dmn.Id)
		if err != nil {
			log.Warn("Error listing VMs: ", err)
			continue
		}
		allVMs = append(allVMs, vms...)
	}

	progressMarker := int(math.Max(float64(len(allVMs)*numVolumesPerVM)/10.0, 5))
	start := time.Now()

	log.Infof("Creating %d volumes", len(allVMs)*numVolumesPerVM)
	unsuitableVmCount := 0
	counter := 0

	for _, vm := range allVMs {
		vm := vm
		if vm.State != "Running" && vm.State != "Stopped" {
			unsuitableVmCount++
			continue
		}
		for j := 1; j <= numVolumesPerVM; j++ {
			counter++
			if counter%progressMarker == 0 {
				log.Infof("Created %d volumes", counter)
			}

			workerPool.Go(func() *Result {
				taskStart := time.Now()
				vol, err := volume.CreateVolume(cs, vm.Domainid, vm.Account)
				if err != nil {
					return &Result{
						Success:  false,
						Duration: time.Since(taskStart).Seconds(),
					}
				}
				_, err = volume.AttachVolume(cs, vol.Id, vm.Id)
				if err != nil {
					return &Result{
						Success:  false,
						Duration: time.Since(taskStart).Seconds(),
					}
				}
				return &Result{
					Success:  true,
					Duration: time.Since(taskStart).Seconds(),
				}
			})
		}
	}
	if unsuitableVmCount > 0 {
		log.Warnf("Found %d VMs in unsuitable state", unsuitableVmCount)
	}
	res := workerPool.Wait()
	log.Infof("Created %d volumes in %.2f seconds", counter, time.Since(start).Seconds())
	return res
}

func tearDownEnv(domainFlag, networkFlag, vmFlag, volumeFlag *bool, workers *int) map[string][]*Result {
	apiURL := config.URL

	for _, profile := range profiles {
		userProfileName := profile.Name
		if userProfileName == "admin" {
			cs := cloudstack.NewAsyncClient(apiURL, profile.ApiKey, profile.SecretKey, false)

			var results = make(map[string][]*Result)

			if *vmFlag {
				workerPool := pool.NewWithResults[*Result]().WithMaxGoroutines(*workers)
				results["vm-destroy"] = destroyVms(workerPool, cs, config.ParentDomainId)
			}

			if *volumeFlag {
				workerPool := pool.NewWithResults[*Result]().WithMaxGoroutines(*workers)
				results["volume-delete"] = deleteVolumes(workerPool, cs, config.ParentDomainId)
			}

			if *networkFlag {
				workerPool := pool.NewWithResults[*Result]().WithMaxGoroutines(*workers)
				results["network-delete"] = deleteNetworks(workerPool, cs, config.ParentDomainId)
			}

			if *domainFlag {
				workerPool := pool.NewWithResults[*Result]().WithMaxGoroutines(*workers)
				results["domain-delete"] = deleteDomains(workerPool, cs, config.ParentDomainId)
			}

			return results
		}
	}
	return nil
}

func destroyVms(workerPool *pool.ResultPool[*Result], cs *cloudstack.CloudStackClient, parentDomainId string) []*Result {
	log.Infof("Fetching subdomains & accounts for domain %s", parentDomainId)
	domains := domain.ListSubDomains(cs, parentDomainId)
	var allVMs []*cloudstack.VirtualMachine
	for _, dmn := range domains {
		vms, err := vm.ListVMs(cs, dmn.Id)
		if err != nil {
			log.Warn("Error listing VMs: ", err)
			continue
		}
		allVMs = append(allVMs, vms...)
	}

	progressMarker := int(math.Max(float64(len(allVMs))/10.0, 5))
	start := time.Now()

	log.Infof("Destroying %d VMs", len(allVMs))

	for i, virtualMachine := range allVMs {
		virtualMachine := virtualMachine
		if i%progressMarker == 0 {
			log.Infof("Destroyed %d VMs", i+1)
		}

		workerPool.Go(func() *Result {
			taskStart := time.Now()
			err := vm.DestroyVm_cs(cs, virtualMachine.Id)
			if err != nil {
				return &Result{
					Success:  false,
					Duration: time.Since(taskStart).Seconds(),
				}
			}
			return &Result{
				Success:  true,
				Duration: time.Since(taskStart).Seconds(),
			}
		})
	}

	res := workerPool.Wait()
	log.Infof("Destroyed %d VMs in %.2f seconds", len(allVMs), time.Since(start).Seconds())
	return res
}

func deleteNetworks(workerPool *pool.ResultPool[*Result], cs *cloudstack.CloudStackClient, parentDomainId string) []*Result {
	log.Infof("Fetching subdomains & accounts for domain %s", parentDomainId)
	domains := domain.ListSubDomains(cs, parentDomainId)

	log.Infof("Fetching networks for subdomains in domain %s", parentDomainId)
	var allNetworks []*cloudstack.Network
	for _, domain := range domains {
		network, _ := network.ListNetworks(cs, domain.Id)
		allNetworks = append(allNetworks, network...)
	}

	progressMarker := int(math.Max(float64(len(allNetworks))/10.0, 5))
	start := time.Now()
	log.Infof("Deleting %d networks", len(allNetworks))
	for i, net := range allNetworks {
		net := net
		if (i+1)%progressMarker == 0 {
			log.Infof("Deleted %d networks", i+1)
		}
		workerPool.Go(func() *Result {
			taskStart := time.Now()
			resp, err := network.DeleteNetwork_cs(cs, net.Id)
			if err != nil || !resp {
				return &Result{
					Success:  false,
					Duration: time.Since(taskStart).Seconds(),
				}
			}
			return &Result{
				Success:  true,
				Duration: time.Since(taskStart).Seconds(),
			}
		})
	}
	res := workerPool.Wait()
	log.Infof("Deleted %d networks in %.2f seconds", len(allNetworks), time.Since(start).Seconds())
	return res
}

func deleteVolumes(workerPool *pool.ResultPool[*Result], cs *cloudstack.CloudStackClient, parentDomainId string) []*Result {
	log.Infof("Fetching subdomains & accounts for domain %s", parentDomainId)
	domains := domain.ListSubDomains(cs, parentDomainId)

	log.Infof("Fetching volumes for subdomains in domain %s", parentDomainId)
	var allVolumes []*cloudstack.Volume
	for _, domain := range domains {
		volumes, _ := volume.ListVolumes(cs, domain.Id)
		allVolumes = append(allVolumes, volumes...)
	}

	progressMarker := int(math.Max(float64(len(allVolumes))/10.0, 5))
	start := time.Now()
	log.Infof("Deleting %d volumes", len(allVolumes))
	for i, vol := range allVolumes {
		vol := vol
		if (i+1)%progressMarker == 0 {
			log.Infof("Deleted %d volumes", i+1)
		}
		workerPool.Go(func() *Result {
			taskStart := time.Now()
			_, err := volume.DestroyVolume(cs, vol.Id)
			if err != nil {
				return &Result{
					Success:  false,
					Duration: time.Since(taskStart).Seconds(),
				}
			}
			return &Result{
				Success:  true,
				Duration: time.Since(taskStart).Seconds(),
			}
		})
	}
	res := workerPool.Wait()
	log.Infof("Deleted %d volumes in %.2f seconds", len(allVolumes), time.Since(start).Seconds())
	return res
}

func deleteDomains(workerPool *pool.ResultPool[*Result], cs *cloudstack.CloudStackClient, parentDomainId string) []*Result {
	log.Infof("Fetching subdomains for domain %s", parentDomainId)
	domains := domain.ListSubDomains(cs, parentDomainId)

	progressMarker := int(math.Max(float64(len(domains))/10.0, 5))
	start := time.Now()
	log.Infof("Deleting %d domains", len(domains))
	for i, dmn := range domains {
		dmn := dmn
		if (i+1)%progressMarker == 0 {
			log.Infof("Deleted %d domains", i+1)
		}
		workerPool.Go(func() *Result {
			taskStart := time.Now()
			resp, err := domain.DeleteDomain_cs(cs, dmn.Id)
			if !resp || err != nil {
				return &Result{
					Success:  false,
					Duration: time.Since(taskStart).Seconds(),
				}
			}
			return &Result{
				Success:  true,
				Duration: time.Since(taskStart).Seconds(),
			}
		})
	}
	res := workerPool.Wait()
	log.Infof("Deleted %d domains in %.2f seconds", len(domains), time.Since(start).Seconds())
	return res
}

func createResources_simpletest() (result map[string][]*apirunner.Results, order []string) {
	apiURL := config.URL

	for _, profile := range profiles {
		if profile.Name == "admin" {
			cs := cloudstack.NewAsyncClient(apiURL, profile.ApiKey, profile.SecretKey, false)

			var result = make(map[string][]*apirunner.Results)
			var order []string

			fmt.Printf("\n\033[1;34m============================================================\033[0m\n")
			fmt.Printf("                    moldbench\n")
			fmt.Printf("\033[1;34m============================================================\033[0m\n")

			// Create domain
			domain := apirunner.CreateDomains(cs, config.ParentDomainId, config.NumDomains)
			result["createDomain"] = append(result["createDomain"], domain)
			order = append(order, "createDomain")
			domain_res := strings.Split(domain.Id, ",")
			if !domain.Success {
				apirunner.DeleteDomain(cs, domain_res[0], false)
				return result, order
			}

			// Create account
			account := apirunner.CreateAccount(cs, domain_res[0], config.NumDomains)
			result["createAccount"] = append(result["createAccount"], account)
			order = append(order, "createAccount")
			account_res := strings.Split(account.Id, ",")
			if !account.Success {
				apirunner.DeleteDomain(cs, domain_res[0], false)
				return result, order
			}

			fmt.Printf("\n\033[1;34m============================================================\033[0m\n")
			fmt.Printf("                    Domain: [%s]\n", domain_res[1])
			fmt.Printf("                    Account: [%s]\n", account_res[1])
			fmt.Printf("\033[1;34m============================================================\033[0m\n")

			// Register template
			template := apirunner.RegisterTemplate(cs, config.Format, config.Hypervisor, config.TemplateUrl, config.OsTypeId, config.ZoneId, domain_res[0], account_res[1])
			result["registerTemplate"] = append(result["registerTemplate"], template)
			order = append(order, "registerTemplate")
			if !template.Success {
				apirunner.DeleteDomain(cs, domain_res[0], true)
				return result, order
			}
			// log.Printf("Tepmlate: %v", template.Id)

			// Register ISO
			// iso := apirunner.RegisterIso(cs, config.Format, config.Hypervisor, config.TemplateUrl, config.OsTypeId, config.ZoneId)
			// result["registerIso"] = append(result["registerIso"], iso)
			// order = append(order, "registerIso")
			// if !iso.Success {
			// 	return result, order
			// }

			// Create network(Isolated & L2)
			network := apirunner.CreateNetwork(cs, config.NetworkOfferingId, "", config.ParentDomainId, domain_res[0], account_res[1], config.NumDomains)
			result["createNetwork(Isolated)"] = append(result["createNetwork(Isolated)"], network)
			order = append(order, "createNetwork(Isolated)")
			if !network.Success {
				apirunner.DeleteDomain(cs, domain_res[0], false)
				return result, order
			}

			networkl2 := apirunner.CreateNetwork(cs, config.L2NetworkOfferingId, "untagged", config.ParentDomainId, domain_res[0], account_res[1], config.NumDomains)
			result["createNetwork(L2)"] = append(result["createNetwork(L2)"], networkl2)
			order = append(order, "createNetwork(L2)")
			if !networkl2.Success {
				apirunner.DeleteDomain(cs, domain_res[0], false)
				return result, order
			}

			// Create vm
			vm := apirunner.CreateVms(cs, config.ParentDomainId, domain_res[0], account_res[1], network.Id, config.NumVms)
			result["createVms"] = append(result["createVms"], vm)
			order = append(order, "createVms")
			if !vm.Success {
				apirunner.DeleteDomain(cs, domain_res[0], false)
				return result, order
			}

			// Create volume & Attach volume
			volume := apirunner.CreateVolumes(cs, domain_res[0], account_res[1], vm.Id, config.NumVolumes)
			result["createVolumes"] = append(result["createVolumes"], volume)
			order = append(order, "createVolumes")
			if !volume.Success {
				apirunner.DeleteDomain(cs, domain_res[0], false)
				return result, order
			}

			// Create volume snapshot
			volumesnapshot := apirunner.CreateSnapshot(cs, volume.Id)
			result["createSnapshot"] = append(result["createSnapshot"], volumesnapshot)
			order = append(order, "createSnapshot")
			if !volumesnapshot.Success {
				apirunner.DeleteDomain(cs, domain_res[0], false)
				return result, order
			}

			// Delete volume snapshot
			deleteSnapshot := apirunner.DeleteSnapshot(cs, volumesnapshot.Id)
			result["deleteSnapshot"] = append(result["deleteSnapshot"], deleteSnapshot)
			order = append(order, "deleteSnapshot")
			if !deleteSnapshot.Success {
				apirunner.DeleteDomain(cs, domain_res[0], false)
				return result, order
			}

			// Detach volume
			detachVolume := apirunner.DetachVolume(cs, volume.Id)
			result["detachVolume"] = append(result["detachVolume"], detachVolume)
			order = append(order, "detachVolume")
			if !detachVolume.Success {
				apirunner.DeleteDomain(cs, domain_res[0], false)
				return result, order
			}

			// Delete volume
			destroyVolume := apirunner.DestroyVolume(cs, volume.Id)
			result["destroyVolume"] = append(result["destroyVolume"], destroyVolume)
			order = append(order, "destroyVolume")
			if !destroyVolume.Success {
				apirunner.DeleteDomain(cs, domain_res[0], false)
				return result, order
			}

			// Create vm snapshot(vm snapshot 존재할 경우, attach volume 안됨)
			vmsnapshot := apirunner.CreateVmSnapshot(cs, vm.Id)
			result["createVmSnapshot"] = append(result["createVmSnapshot"], vmsnapshot)
			order = append(order, "createVmSnapshot")
			if !vmsnapshot.Success {
				apirunner.DeleteDomain(cs, domain_res[0], false)
				return result, order
			}

			// Delete vm snapshot
			deleteVmSnapshot := apirunner.DeleteVmSnapshot(cs, vmsnapshot.Id)
			result["deleteVmSnapshot"] = append(result["deleteVmSnapshot"], deleteVmSnapshot)
			order = append(order, "deleteVmSnapshot")
			if !deleteVmSnapshot.Success {
				apirunner.DeleteDomain(cs, domain_res[0], false)
				return result, order
			}

			// Stop vm
			stopVm := apirunner.StopVm(cs, vm.Id)
			result["stopVm"] = append(result["stopVm"], stopVm)
			order = append(order, "stopVm")
			if !stopVm.Success {
				apirunner.DeleteDomain(cs, domain_res[0], false)
				return result, order
			}

			// Start vm
			startVm := apirunner.StartVm(cs, vm.Id)
			result["startVm"] = append(result["startVm"], startVm)
			order = append(order, "startVm")
			if !startVm.Success {
				apirunner.DeleteDomain(cs, domain_res[0], false)
				return result, order
			}

			// Destroy Vm
			destroyVm := apirunner.DestroyVm(cs, vm.Id)
			result["destroyVm"] = append(result["destroyVm"], destroyVm)
			order = append(order, "destroyVm")
			if !destroyVm.Success {
				apirunner.DeleteDomain(cs, domain_res[0], false)
				return result, order
			}

			// Delete Network(Isolated & L2)
			deleteNetwork := apirunner.DeleteNetwork(cs, network.Id)
			result["deleteNetwork(Isolated)"] = append(result["deleteNetwork(Isolated)"], deleteNetwork)
			order = append(order, "deleteNetwork(Isolated)")
			if !deleteNetwork.Success {
				apirunner.DeleteDomain(cs, domain_res[0], false)
				return result, order
			}

			deleteNetwork = apirunner.DeleteNetwork(cs, networkl2.Id)
			result["deleteNetwork(L2)"] = append(result["deleteNetwork(L2)"], deleteNetwork)
			order = append(order, "deleteNetwork(L2)")
			if !deleteNetwork.Success {
				apirunner.DeleteDomain(cs, domain_res[0], false)
				return result, order
			}

			// Delete template
			deleteTemplate := apirunner.DeleteTemplate(cs, domain_res[0])
			result["deleteTemplate"] = append(result["deleteTemplate"], deleteTemplate)
			order = append(order, "deleteTemplate")
			if !deleteTemplate.Success {
				apirunner.DeleteDomain(cs, domain_res[0], false)
				return result, order
			}

			// Delete Account
			deleteAccount := apirunner.DeleteAccount(cs, account_res[0])
			result["deleteAccount"] = append(result["deleteAccount"], deleteAccount)
			order = append(order, "deleteAccount")
			if !deleteAccount.Success {
				apirunner.DeleteDomain(cs, domain_res[0], false)
				return result, order
			}

			// Delete Domain
			deleteDomain := apirunner.DeleteDomain(cs, domain_res[0], true)
			result["deleteDomain"] = append(result["deleteDomain"], deleteDomain)
			order = append(order, "deleteDomain")
			if !deleteDomain.Success {
				apirunner.DeleteDomain(cs, domain_res[0], false)
				return result, order
			}

			// Create serviceOffering
			serviceOffering := apirunner.CreateServiceOffering(cs, config.NumDomains)
			result["createServiceOffering"] = append(result["createServiceOffering"], serviceOffering)
			order = append(order, "createServiceOffering")
			if !serviceOffering.Success {
				apirunner.DeleteDomain(cs, domain_res[0], false)
				return result, order
			}

			// Delete serviceOffering
			deleteserviceOffering := apirunner.DeleteServiceOffering(cs, serviceOffering.Id, config.NumDomains)
			result["deleteServiceOffering"] = append(result["deleteServiceOffering"], deleteserviceOffering)
			order = append(order, "deleteServiceOffering")
			if !deleteserviceOffering.Success {
				apirunner.DeleteDomain(cs, domain_res[0], false)
				return result, order
			}

			// Create diskoffering
			diskOffering := apirunner.CreateDiskOffering(cs, config.NumDomains)
			result["createDiskOffering"] = append(result["createDiskOffering"], diskOffering)
			order = append(order, "createDiskOffering")
			if !diskOffering.Success {
				apirunner.DeleteDomain(cs, domain_res[0], false)
				return result, order
			}

			// Delete diskoffering
			deletediskOffering := apirunner.DeleteDiskOffering(cs, diskOffering.Id, config.NumDomains)
			result["deleteDiskOffering"] = append(result["deleteDiskOffering"], deletediskOffering)
			order = append(order, "deleteDiskOffering")
			if !deletediskOffering.Success {
				apirunner.DeleteDomain(cs, domain_res[0], false)
				return result, order
			}

			// Create networkoffering
			networkOffering := apirunner.CreateNetworkOffering(cs, config.NumDomains)
			result["createNetworkOffering"] = append(result["createNetworkOffering"], networkOffering)
			order = append(order, "createNetworkOffering")
			if !networkOffering.Success {
				apirunner.DeleteDomain(cs, domain_res[0], false)
				return result, order
			}

			// Delete networkoffering
			deletenetworkOffering := apirunner.DeleteNetworkOffering(cs, networkOffering.Id, config.NumDomains)
			result["deleteNetworkOffering"] = append(result["deleteNetworkOffering"], deletenetworkOffering)
			order = append(order, "deleteNetworkOffering")
			if !deletenetworkOffering.Success {
				apirunner.DeleteDomain(cs, domain_res[0], false)
				return result, order
			}

			return result, order
		}
	}
	return nil, nil
}
