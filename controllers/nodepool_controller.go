/*
Copyright 2021.

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

	appsv1 "github.com/SunhaoKim/nodepool_operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	//"k8s.io/api/node/v1beta1"

	v1 "k8s.io/api/node/v1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NodepoolReconciler reconciles a Nodepool object
type NodepoolReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//定义数据类型
var (
	nodes corev1.NodeList
)

//+kubebuilder:rbac:groups=apps.operator.com,resources=nodepools,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.operator.com,resources=nodepools/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.operator.com,resources=nodepools/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Nodepool object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *NodepoolReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//_ = log.FromContext(ctx)
	_ = r.Log.WithValues("nodepool", req.NamespacedName)
	fmt.Println(req.NamespacedName)
	//获取对象
	pool := &appsv1.Nodepool{}
	//检测是否获取到pool
	err := r.Get(ctx, req.NamespacedName, pool)
	if err != nil {
		r.Log.Error(err, "can not get node pool")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	//定义err 如果存在即给节点加数据
	err = r.List(ctx, &nodes, &client.ListOptions{LabelSelector: pool.NodeLabelSelector()})
	//检测err是否为空
	if client.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, err
	}
	if len(nodes.Items) > 0 {
		r.Log.Info("find nodes, will merge data", "nodes", len(nodes.Items))
		for _, n := range nodes.Items {
			n := n
			err := r.Update(ctx, pool.Spec.ApplyNode(n))
			//err := r.Patch(ctx, pool.Spec.ApplyNode(n), client.Merge)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
	}
	//调用runtimeclass方法
	runtimeClass := &v1.RuntimeClass{}
	err = r.Get(ctx, client.ObjectKeyFromObject(pool.RuntimeClass()), runtimeClass)
	fmt.Println(err)
	fmt.Println(client.IgnoreNotFound(err))
	if client.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, err
	}
	//检测runtimeclass是否存在，不存在则创建
	fmt.Println("debug11111111111111111111111111111111", runtimeClass.Name)
	if runtimeClass.Name == "" {
		runtimeClass = pool.RuntimeClass()
		//err = r.Create(ctx, pool.RuntimeClass())
		err = controllerutil.SetOwnerReference(pool, runtimeClass, r.Scheme)
		fmt.Print(err)
		if err != nil {
			return ctrl.Result{}, err
		}
		err = r.Create(ctx, pool.RuntimeClass())
		fmt.Println("debug runtimeclass", err)
		return ctrl.Result{}, err
	}
	//存在则更新
	runtimeClass.Scheduling = pool.RuntimeClass().Scheduling
	runtimeClass.Handler = pool.RuntimeClass().Handler
	err = r.Client.Update(ctx, runtimeClass)
	if err != nil {
		return ctrl.Result{}, err
	}
	//_ = r.Log.WithValues("application", req.NamespacedName)
	// your logic here
	//r.Log.Info("app changed", "ns", req.Namespace)
	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodepoolReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Nodepool{}).
		Complete(r)
}
