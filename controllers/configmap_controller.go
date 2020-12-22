/*


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

package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"strconv"
	"strings"
)

const (
	ConfigName = "cicd-config"
)

// ConfigReconciler reconciles a Approval object
type ConfigReconciler struct {
	client typedcorev1.ConfigMapInterface
	Log    logr.Logger
}

func (r *ConfigReconciler) Start() {
	var err error
	r.client, err = newConfigMapClient()
	if err != nil {
		r.Log.Error(err, "")
		os.Exit(1)
	}

	// Get first to check the ConfigMap's existence
	_, err = r.client.Get(context.Background(), ConfigName, metav1.GetOptions{})
	if err != nil {
		r.Log.Error(err, "")
		os.Exit(1)
	}

	for {
		r.watch()
	}
}

func (r *ConfigReconciler) watch() {
	log := r.Log.WithName("config controller")

	watcher, err := r.client.Watch(context.Background(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", ConfigName),
	})
	if err != nil {
		log.Error(err, "")
		return
	}

	for ev := range watcher.ResultChan() {
		cm, ok := ev.Object.(*corev1.ConfigMap)
		if ok {
			if err := r.Reconcile(cm); err != nil {
				log.Error(err, "")
			}
		}
	}
}

type cfgType int

const (
	cfgTypeString cfgType = iota
	cfgTypeInt
	cfgTypeBool
)

type operatorConfig struct {
	Type cfgType

	StringVal     *string
	StringDefault string

	IntVal     *int
	IntDefault int

	BoolVal     *bool
	BoolDefault bool
}

func (r *ConfigReconciler) Reconcile(cm *corev1.ConfigMap) error {
	r.Log.Info("Config is changed")

	if cm == nil {
		return nil
	}

	vars := map[string]operatorConfig{
		"maxPipelineRun":   {Type: cfgTypeInt, IntVal: &configs.MaxPipelineRun, IntDefault: 5},    // Max PipelineRun count
		"enableMail":       {Type: cfgTypeBool, BoolVal: &configs.EnableMail, BoolDefault: false}, // Enable Mail
		"externalHostName": {Type: cfgTypeString, StringVal: &configs.ExternalHostName},           // External Hostname
		"smtpHost":         {Type: cfgTypeString, StringVal: &configs.SMTPHost},                   // SMTP Host
		"smtpUserSecret":   {Type: cfgTypeString, StringVal: &configs.SMTPUserSecret},             // SMTP Cred
	}

	getVars(cm.Data, vars)

	// Check SMTP config.s
	if configs.EnableMail && (configs.SMTPHost == "" || configs.SMTPUserSecret == "") {
		return fmt.Errorf("email is enaled but smtp access info. is not given")
	}

	return nil
}

func getVars(data map[string]string, vars map[string]operatorConfig) {
	for key, c := range vars {
		v, exist := data[key]
		// If not set, set as default
		if !exist {
			switch c.Type {
			case cfgTypeString:
				if c.StringVal == nil {
					continue
				}
				if c.StringDefault != "" {
					*c.StringVal = c.StringDefault
				}
			case cfgTypeInt:
				if c.IntVal == nil {
					continue
				}
				*c.IntVal = c.IntDefault
			case cfgTypeBool:
				if c.BoolVal == nil {
					continue
				}
				*c.BoolVal = c.BoolDefault
			}
		} else {
			// If set, set value
			switch c.Type {
			case cfgTypeString:
				if c.StringVal == nil {
					continue
				}
				if v != "" {
					*c.StringVal = v
				}
			case cfgTypeInt:
				if c.IntVal == nil {
					continue
				}
				i, err := strconv.Atoi(v)
				if err != nil {
					continue
				}
				*c.IntVal = i
			case cfgTypeBool:
				if c.BoolVal == nil {
					continue
				}
				switch strings.ToLower(v) {
				case "true":
					*c.BoolVal = true
				case "false":
					*c.BoolVal = false
				}
			}
		}
	}
}

func newConfigMapClient() (typedcorev1.ConfigMapInterface, error) {
	conf, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	clientSet, err := kubernetes.NewForConfig(conf)
	if err != nil {
		return nil, err
	}

	namespace, err := utils.Namespace()
	if err != nil {
		return nil, err
	}

	return clientSet.CoreV1().ConfigMaps(namespace), nil
}