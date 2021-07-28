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

package controllers

import (
	"github.com/bmizerany/assert"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"testing"
)

const (
	testConfigMapNamespace = "test-ns"
	testConfigMapName      = "test-cm"
)

func TestConfigReconciler_Reconcile(t *testing.T) {
	if _, exist := os.LookupEnv("CI"); !exist {
		ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	}

	reconciler := ConfigReconciler{
		client:               fake.NewSimpleClientset().CoreV1().ConfigMaps(testConfigMapNamespace),
		Log:                  logf.Log.WithName("test-logger"),
		lastResourceVersions: map[string]string{},
		Handlers:             map[string]configs.Handler{},
	}

	t.Run("new-resource", func(t *testing.T) {
		handled := false
		reconciler.Handlers[testConfigMapName] = func(cm *corev1.ConfigMap) error {
			handled = true
			return nil
		}
		cm := &corev1.ConfigMap{}
		cm.Namespace = testConfigMapNamespace
		cm.Name = testConfigMapName
		cm.ResourceVersion = "11111"
		_ = reconciler.Reconcile(cm)
		assert.Equal(t, true, handled)
	})

	t.Run("updated-resource", func(t *testing.T) {
		handled := false
		reconciler.Handlers[testConfigMapName] = func(cm *corev1.ConfigMap) error {
			handled = true
			return nil
		}
		reconciler.lastResourceVersions[testConfigMapName] = "11111"
		cm := &corev1.ConfigMap{}
		cm.Namespace = testConfigMapNamespace
		cm.Name = testConfigMapName
		cm.ResourceVersion = "11112"
		_ = reconciler.Reconcile(cm)
		assert.Equal(t, true, handled)
	})

	t.Run("same-resource-version", func(t *testing.T) {
		handled := false
		reconciler.Handlers[testConfigMapName] = func(cm *corev1.ConfigMap) error {
			handled = true
			return nil
		}
		reconciler.lastResourceVersions[testConfigMapName] = "11111"
		cm := &corev1.ConfigMap{}
		cm.Namespace = testConfigMapNamespace
		cm.Name = testConfigMapName
		cm.ResourceVersion = "11111"
		_ = reconciler.Reconcile(cm)
		assert.Equal(t, false, handled)
	})
}
