/*
 Copyright 2021 The CI/CD Operator Authors

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package configs

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

// Configs to be configured by command line arguments

// Names of config maps
const (
	ConfigMapNameCICDConfig    = "cicd-config"
	ConfigMapNameEmailTemplate = "email-template"
)

var controllerConfigUpdateChan []chan struct{}

// RegisterControllerConfigUpdateChan registers a channel which accepts controller config's update event
func RegisterControllerConfigUpdateChan(ch chan struct{}) {
	controllerConfigUpdateChan = append(controllerConfigUpdateChan, ch)
}

// ApplyControllerConfigChange is a configmap handler for cicd-config configmap
func ApplyControllerConfigChange(cm *corev1.ConfigMap) error {
	getVars(cm.Data, map[string]operatorConfig{
		"maxPipelineRun":            {Type: cfgTypeInt, IntVal: &MaxPipelineRun, IntDefault: 5},                                // Max PipelineRun count
		"enableMail":                {Type: cfgTypeBool, BoolVal: &EnableMail, BoolDefault: false},                             // Enable Mail
		"externalHostName":          {Type: cfgTypeString, StringVal: &ExternalHostName},                                       // External Hostname
		"exposeMode":                {Type: cfgTypeString, StringVal: &ExposeMode, StringDefault: "Ingress"},                   // Expose mode
		"reportRedirectUriTemplate": {Type: cfgTypeString, StringVal: &ReportRedirectURITemplate},                              // RedirectUriTemplate for report access
		"smtpHost":                  {Type: cfgTypeString, StringVal: &SMTPHost},                                               // SMTP Host
		"smtpUserSecret":            {Type: cfgTypeString, StringVal: &SMTPUserSecret},                                         // SMTP Cred
		"collectPeriod":             {Type: cfgTypeInt, IntVal: &CollectPeriod, IntDefault: 120},                               // GC period
		"integrationJobTTL":         {Type: cfgTypeInt, IntVal: &IntegrationJobTTL, IntDefault: 120},                           // GC threshold
		"ingressClass":              {Type: cfgTypeString, StringVal: &IngressClass, StringDefault: ""},                        // Ingress class
		"ingressHost":               {Type: cfgTypeString, StringVal: &IngressHost, StringDefault: ""},                         // Ingress host
		"gitImage":                  {Type: cfgTypeString, StringVal: &GitImage, StringDefault: "docker.io/alpine/git:1.0.30"}, // Git image
	})

	// Check SMTP config.s
	if EnableMail && (SMTPHost == "" || SMTPUserSecret == "") {
		return fmt.Errorf("email is enaled but smtp access info. is not given")
	}

	// Init
	if !ControllerInitiated {
		ControllerInitiated = true
		if len(ControllerInitCh) < cap(ControllerInitCh) {
			ControllerInitCh <- struct{}{}
		}
	}

	// Notify channels (non-blocking way)
	for _, ch := range controllerConfigUpdateChan {
		if len(ch) < cap(ch) {
			ch <- struct{}{}
		}
	}

	return nil
}

// ApplyEmailTemplateConfigChange is a configmap handler for email-template configmap
func ApplyEmailTemplateConfigChange(cm *corev1.ConfigMap) error {
	getVars(cm.Data, map[string]operatorConfig{
		"request-title":   {Type: cfgTypeString, StringVal: &ApprovalRequestMailTitle, StringDefault: "[CI/CD] Approval '{{.Name}}' is requested to you"},
		"request-content": {Type: cfgTypeString, StringVal: &ApprovalRequestMailContent, StringDefault: "{{.Name}}"},
		"result-title":    {Type: cfgTypeString, StringVal: &ApprovalResultMailTitle, StringDefault: "[CI/CD] Approval is {{.Status.Result}}"},
		"result-content":  {Type: cfgTypeString, StringVal: &ApprovalResultMailContent, StringDefault: "{{.Name}}"},
	})
	return nil
}

// Configs for manager
var (
	// MaxPipelineRun is the number of PipelineRuns that can run simultaneously
	MaxPipelineRun int

	// ExternalHostName to be used for webhook server (default is ingress host name)
	ExternalHostName string

	// CurrentExternalHostName is NOT a configurable variable! it just stores current hostname which will be used for
	// exposing webhook/result server
	CurrentExternalHostName string

	// ReportRedirectURITemplate is a uri template for report page redirection
	ReportRedirectURITemplate string

	// CollectPeriod is a garbage collection period (in hour)
	CollectPeriod int

	// IntegrationJobTTL is a garbage collection threshold (in hour).
	// If IntegrationJob's .status.completionTime + TTL < now, it's collected
	IntegrationJobTTL int

	// EnableMail is whether to enable mail feature or not
	EnableMail bool

	// SMTPHost is a host (IP:PORT) of the SMTP server
	SMTPHost string

	// SMTPUserSecret is a credential secret for the SMTP server (should be basic type)
	SMTPUserSecret string

	// ApprovalRequestMailTitle is a title for the approval request mail
	ApprovalRequestMailTitle string
	// ApprovalRequestMailContent is a content of the approval request mail
	ApprovalRequestMailContent string

	// ApprovalResultMailTitle is a title for the approval result mail
	ApprovalResultMailTitle string
	// ApprovalResultMailContent is a content of the approval result mail
	ApprovalResultMailContent string

	// ExposeMode is a mode to be used for exposing the webhook server (Ingress/LoadBalancer/ClusterIP)
	ExposeMode string

	// IngressClass is a class for ingress instance
	IngressClass string

	// IngressHost is a host for ingress instance
	IngressHost string

	// GitImage is an image url for the git-checkout step
	GitImage string
)
