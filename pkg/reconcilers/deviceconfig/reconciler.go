/*
Copyright 2024 Nokia.

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

package deviceconfig

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/henderiw/logger/log"
	conditionv1alpha1 "github.com/kuidio/kuid/apis/condition/v1alpha1"
	"github.com/kuidio/kuid/pkg/reconcilers/resource"
	"github.com/kuidio/kuid/pkg/resources"
	netwv1alpha1 "github.com/kuidio/kuidapps/apis/network/v1alpha1"
	"github.com/kuidio/nokia-srl/pkg/parser"
	"github.com/kuidio/nokia-srl/pkg/reconcilers"
	"github.com/kuidio/nokia-srl/pkg/reconcilers/ctrlconfig"
	perrors "github.com/pkg/errors"
	configv1alpha1 "github.com/sdcio/config-server/apis/config/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/yaml"
)

func init() {
	reconcilers.Register("deviceconfig", &reconciler{})
}

const (
	crName         = "networkdevice"
	controllerName = "SRLDeviceConfigController"
	finalizer      = "config.srl.nokia.app.kuid.dev/finalizer"
	// errors
	errGetCr        = "cannot get cr"
	errUpdateStatus = "cannot update status"
)

// SetupWithManager sets up the controller with the Manager.
func (r *reconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, c interface{}) (map[schema.GroupVersionKind]chan event.GenericEvent, error) {

	_, ok := c.(*ctrlconfig.ControllerConfig)
	if !ok {
		return nil, fmt.Errorf("cannot initialize, expecting controllerConfig, got: %s", reflect.TypeOf(c).Name())
	}

	r.Client = mgr.GetClient()
	//r.APIPatchingApplicator = resource.NewAPIPatchingApplicator(mgr.GetClient())
	r.finalizer = resource.NewAPIFinalizer(mgr.GetClient(), finalizer)
	r.recorder = mgr.GetEventRecorderFor(controllerName)

	var err error
	r.parser, err = parser.New("templates")
	if err != nil {
		return nil, err
	}

	return nil, ctrl.NewControllerManagedBy(mgr).
		Named(controllerName).
		For(&netwv1alpha1.NetworkDevice{}).
		Owns(&configv1alpha1.Config{}).
		Complete(r)
}

type reconciler struct {
	//resource.APIPatchingApplicator
	client.Client
	finalizer *resource.APIFinalizer
	recorder  record.EventRecorder
	parser    *parser.Parser
}

func (r *reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx = ctrlconfig.InitContext(ctx, controllerName, req.NamespacedName)
	log := log.FromContext(ctx)
	log.Info("reconcile")

	cr := &netwv1alpha1.NetworkDevice{}
	if err := r.Client.Get(ctx, req.NamespacedName, cr); err != nil {
		// if the resource no longer exists the reconcile loop is done
		if resource.IgnoreNotFound(err) != nil {
			log.Error(errGetCr, "error", err)
			return ctrl.Result{}, perrors.Wrap(resource.IgnoreNotFound(err), errGetCr)
		}
		return ctrl.Result{}, nil
	}
	cr = cr.DeepCopy()

	// TODO add provider logic

	if !cr.GetDeletionTimestamp().IsZero() {
		if err := r.delete(ctx, cr); err != nil {
			r.handleError(ctx, cr, "canot delete resources", err)
			return reconcile.Result{Requeue: true}, perrors.Wrap(r.Client.Status().Update(ctx, cr), errUpdateStatus)
		}

		if err := r.finalizer.RemoveFinalizer(ctx, cr); err != nil {
			r.handleError(ctx, cr, "cannot remove finalizer", err)
			return ctrl.Result{Requeue: true}, perrors.Wrap(r.Client.Status().Update(ctx, cr), errUpdateStatus)
		}
		log.Debug("Successfully deleted resource")
		return ctrl.Result{}, nil
	}

	if err := r.finalizer.AddFinalizer(ctx, cr); err != nil {
		r.handleError(ctx, cr, "cannot add finalizer", err)
		return ctrl.Result{Requeue: true}, perrors.Wrap(r.Client.Status().Update(ctx, cr), errUpdateStatus)
	}

	if err := r.apply(ctx, cr); err != nil {
		if errd := r.delete(ctx, cr); errd != nil {
			err = errors.Join(err, errd)
			r.handleError(ctx, cr, "cannot delete resources after populate failed", err)
			return reconcile.Result{Requeue: true}, perrors.Wrap(r.Client.Status().Update(ctx, cr), errUpdateStatus)
		}
		r.handleError(ctx, cr, "cannot apply topology resources", err)
		return ctrl.Result{}, perrors.Wrap(r.Client.Status().Update(ctx, cr), errUpdateStatus)
	}

	cr.SetConditions(conditionv1alpha1.Ready())
	r.recorder.Eventf(cr, corev1.EventTypeNormal, crName, "ready")
	return ctrl.Result{}, perrors.Wrap(r.Client.Status().Update(ctx, cr), errUpdateStatus)
}

func (r *reconciler) handleError(ctx context.Context, cr *netwv1alpha1.NetworkDevice, msg string, err error) {
	log := log.FromContext(ctx)
	if err == nil {
		cr.SetConditions(conditionv1alpha1.Failed(msg))
		log.Error(msg)
		r.recorder.Eventf(cr, corev1.EventTypeWarning, crName, msg)
	} else {
		cr.SetConditions(conditionv1alpha1.Failed(err.Error()))
		log.Error(msg, "error", err)
		r.recorder.Eventf(cr, corev1.EventTypeWarning, crName, fmt.Sprintf("%s, err: %s", msg, err.Error()))
	}
}

func (r *reconciler) apply(ctx context.Context, cr *netwv1alpha1.NetworkDevice) error {
	resources := resources.New(r.Client, resources.Config{
		Owns: []schema.GroupVersionKind{
			configv1alpha1.SchemeGroupVersion.WithKind(configv1alpha1.ConfigKind),
		},
	})

	/*
		nds, err := r.getNetworkDeviceConfigs(ctx, cr)
		if err != nil {
			return err
		}
		for _, nd := range nds {
	*/
	var buf bytes.Buffer
	if err := r.parser.Render(ctx, cr, &buf); err != nil {
		return err
	}

	cfg := configv1alpha1.BuildConfig(
		metav1.ObjectMeta{
			Namespace: cr.GetNamespace(),
			Name:      cr.GetName(),
			Labels: map[string]string{
				"config.sdcio.dev/targetName":      getNodeName(cr.GetName()),
				"config.sdcio.dev/targetNamespace": cr.GetNamespace(),
			},
		},
		configv1alpha1.ConfigSpec{
			Priority: 10,
			Config: []configv1alpha1.ConfigBlob{
				{
					Path: "/",
					Value: runtime.RawExtension{
						Raw: buf.Bytes(),
					},
				},
			},
		},
		configv1alpha1.ConfigStatus{},
	)
	b, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	//}

	return resources.APIApply(ctx, cr)
}

func (r *reconciler) delete(ctx context.Context, cr *netwv1alpha1.NetworkDevice) error {
	resources := resources.New(r.Client, resources.Config{
		Owns: []schema.GroupVersionKind{
			configv1alpha1.SchemeGroupVersion.WithKind(configv1alpha1.ConfigKind),
		},
	})

	return resources.APIDelete(ctx, cr)
}


/*
func (r *reconciler) getNetworkDeviceConfigs(ctx context.Context, cr *netwv1alpha1.Network) ([]*netwv1alpha1.NetworkDevice, error) {
	nds := []*netwv1alpha1.NetworkDevice{}

	opts := []client.ListOption{
		client.InNamespace(cr.Namespace),
	}
	ndList := &netwv1alpha1.NetworkDeviceList{}
	if err := r.Client.List(ctx, ndList, opts...); err != nil {
		return nil, fmt.Errorf("cannot get nodeModel from api, err: %s", err.Error())
	}

	for _, nd := range ndList.Items {
		if strings.HasPrefix(nd.Name, cr.Name) {
			nds = append(nds, &nd)
		}
	}
	return nds, nil
}
*/
func getNodeName(name string) string {
	lastDotIndex := strings.LastIndex(name, ".")
	if lastDotIndex == -1 {
		// If there's no dot, return the original string
		return name
	}
	return name[lastDotIndex+1:]
}
