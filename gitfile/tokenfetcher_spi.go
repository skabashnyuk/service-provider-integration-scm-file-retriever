// Copyright (c) 2022 Red Hat, Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gitfile

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"

	"github.com/redhat-appstudio/service-provider-integration-operator/api/v1beta1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SpiTokenFetcher token fetcher implementation that looks for token in the specific ENV variable.
type SpiTokenFetcher struct {
	k8sClient client.Client
}

const (
	letterBytes = "abcdefghijklmnopqrstuvwxyz1234567890"
	duration    = 5 * time.Second
)

func NewSpiTokenFetcher() *SpiTokenFetcher {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	scheme := runtime.NewScheme()
	if err = corev1.AddToScheme(scheme); err != nil {
		panic(err.Error())
	}

	if err = v1beta1.AddToScheme(scheme); err != nil {
		panic(err.Error())
	}

	// creates the client
	k8sClient, err := client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		panic(err.Error())
	}
	return &SpiTokenFetcher{k8sClient: k8sClient}
}

func (s *SpiTokenFetcher) BuildHeader(ctx context.Context, namespace, repoUrl string, loginCallback func(ctx context.Context, url string)) (*HeaderStruct, error) {
	zap.L().Info(fmt.Sprintf("BuildHeader stag1: %d", 1))
	var tBindingName = "file-retriever-binding-" + randStringBytes(6)

	// create binding
	newBinding := newSPIATB(tBindingName, namespace, repoUrl)
	err := s.k8sClient.Create(ctx, newBinding)
	if err != nil {
		zap.L().Error("Error creating Token Binding item:", zap.Error(err))
		return nil, err
	}
	zap.L().Info(fmt.Sprintf("BuildHeader stag1: %d", 2))
	// scheduling the binding cleanup
	defer func() {
		// clean up token binding
		zap.L().Info(fmt.Sprintf("BuildHeader stag1: %d", 3))
		//err = s.k8sClient.Delete(ctx, newBinding)
		if err != nil {
			zap.L().Error("Error cleaning up TB item:", zap.Error(err))
		}
	}()

	// now re-reading SPITokenBinding to get updated fields
	var tokenName string
	for timeout := time.After(duration); ; {
		zap.L().Info(fmt.Sprintf("BuildHeader stag1: %d", 4))
		readBinding, err := readTB(ctx, namespace, tBindingName, s.k8sClient)
		if err != nil {
			zap.L().Error("Error reading TB item:", zap.Error(err))
		}
		zap.L().Info(fmt.Sprintf("BuildHeader stag1: %d", 5))
		tokenName = readBinding.Status.LinkedAccessTokenName
		if tokenName != "" {
			break
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("task is cancelled")
		case <-timeout:
			zap.L().Error("Timeout reached reading TB item:", zap.Error(err))
			return nil, fmt.Errorf("TB reading task is timed out")
		default:
			time.Sleep(200 * time.Millisecond)
		}
	}
	zap.L().Info(fmt.Sprintf("BuildHeader stag1: %d", 6))
	zap.L().Info(fmt.Sprintf("Access Token to watch: %s", tokenName))

	// now try read SPIAccessToken to get link
	var url string
	var loginCalled = false
	for timeout := time.After(10 * duration); ; {
		readToken := &v1beta1.SPIAccessToken{}
		_ = s.k8sClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: tokenName}, readToken)
		if readToken.Status.Phase == v1beta1.SPIAccessTokenPhaseAwaitingTokenData && !loginCalled {
			url = readToken.Status.OAuthUrl
			zap.L().Info(fmt.Sprintf("URL to OAUth: %s", url))
			go loginCallback(ctx, url)
			loginCalled = true
		} else if readToken.Status.Phase == v1beta1.SPIAccessTokenPhaseReady {
			// now we can exit the loop and read the secret
			break
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("task is cancelled")
		case <-timeout:
			zap.L().Error("Timeout reached reading Token item:", zap.Error(err))
			return nil, fmt.Errorf("token reading task is timed out")
		default:
			time.Sleep(200 * time.Millisecond)
		}
	}
	zap.L().Info(fmt.Sprintf("BuildHeader stag1: %d", 7))
	// now re-reading SPITokenBinding to get updated fields
	var secretName string
	for timeout := time.After(duration); ; {
		readBinding, err := readTB(ctx, namespace, tBindingName, s.k8sClient)
		if err != nil {
			zap.L().Error("Error reading TB item:", zap.Error(err))
		}
		err = s.k8sClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: tBindingName}, readBinding)
		if err != nil {
			zap.L().Error("Error reading TB item:", zap.Error(err))
		}
		secretName = readBinding.Status.SyncedObjectRef.Name
		if secretName != "" {
			break
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("task is cancelled")
		case <-timeout:
			zap.L().Error("Timeout reached reading TB item:", zap.Error(err))
			return nil, fmt.Errorf("TB reading task is timed out")
		default:
			time.Sleep(200 * time.Millisecond)
		}
	}
	zap.L().Info(fmt.Sprintf("BuildHeader stag1: %d", 8))
	zap.L().Info(fmt.Sprintf("Secret to watch: %s", secretName))

	// reading token secret
	for timeout := time.After(duration); ; {
		tokenSecret := &corev1.Secret{}
		err = s.k8sClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: secretName}, tokenSecret)
		if err != nil {
			zap.L().Error("Error reading Token Secret item:", zap.Error(err))
			return nil, err
		}
		if len(tokenSecret.Data) > 0 {
			return &HeaderStruct{Authorization: "Bearer " + string(tokenSecret.Data["password"])}, nil
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("task is cancelled")
		case <-timeout:
			zap.L().Error("Timeout reached reading Secret item:", zap.Error(err))
			return nil, fmt.Errorf("secred reading task is timed out")
		default:
			time.Sleep(200 * time.Millisecond)
		}
		zap.L().Info(fmt.Sprintf("BuildHeader stag1: %d", 9))
	}
}

func newSPIATB(tBindingName, namespace, repoUrl string) *v1beta1.SPIAccessTokenBinding {
	newBinding := &v1beta1.SPIAccessTokenBinding{
		ObjectMeta: metav1.ObjectMeta{Name: tBindingName, Namespace: namespace},
		Spec: v1beta1.SPIAccessTokenBindingSpec{
			RepoUrl: repoUrl,
			Permissions: v1beta1.Permissions{
				Required: []v1beta1.Permission{
					{
						Type: v1beta1.PermissionTypeReadWrite,
						Area: v1beta1.PermissionAreaRepository,
					},
				},
				AdditionalScopes: []string{"api"},
			},
			Secret: v1beta1.SecretSpec{
				Type: corev1.SecretTypeBasicAuth,
			},
		},
	}
	return newBinding
}

func readTB(ctx context.Context, namespace, tBindingName string, k8sClient client.Client) (*v1beta1.SPIAccessTokenBinding, error) {
	readBinding := &v1beta1.SPIAccessTokenBinding{}
	err := k8sClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: tBindingName}, readBinding)
	if err != nil {
		return nil, err
	}
	return readBinding, nil
}

func randStringBytes(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
