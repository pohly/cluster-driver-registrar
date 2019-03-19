/*
Copyright 2018 The Kubernetes Authors.

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

package main

import (
	"os"
	"os/signal"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog"
)

func kubernetesRegister(
	name string,
	add func() error,
	remove func() error,
) {
	// Set up goroutine to cleanup (aka deregister) on termination.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go cleanup(c, name, remove)

	// Run forever
	for {
		verifyAndAddCSIDriverInfo(name, add)
		time.Sleep(sleepDuration)
	}
}

func cleanup(c <-chan os.Signal, name string, remove func() error) {
	<-c
	verifyAndDeleteCSIDriverInfo(name, remove)
	os.Exit(1)
}

// Registers CSI driver by creating a CSIDriver object
func verifyAndAddCSIDriverInfo(
	name string,
	add func() error,
) error {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		err := add()
		if err == nil {
			klog.V(1).Infof("CSIDriver object created for driver %s", name)
			return nil
		} else if apierrors.IsAlreadyExists(err) {
			klog.V(1).Info("CSIDriver CRD already had been registered")
			return nil
		}
		klog.Errorf("Failed to create CSIDriver object: %v", err)
		return err
	})
	return retryErr
}

// Deregister CSI Driver by deleting CSIDriver object
func verifyAndDeleteCSIDriverInfo(
	name string,
	remove func() error,
) error {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		err := remove()
		if err == nil {
			klog.V(1).Infof("CSIDriver object deleted for driver %s", name)
			return nil
		} else if apierrors.IsNotFound(err) {
			klog.V(1).Info("No need to clean up CSIDriver since it does not exist")
			return nil
		}
		klog.Errorf("Failed to delete CSIDriver object: %v", err)
		return err
	})
	return retryErr
}
