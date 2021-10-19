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

package v1

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/node/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NodepoolSpec defines the desired state of Nodepool
//定义nodepollspec结构体
type NodepoolSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	//定义污点 标签 selector
	Taints       []corev1.Taint    `json:"taints,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
	Nodeselector map[string]string `json:"nodeselector,omitemptm"`
	// Foo is an example field of Nodepool. Edit nodepool_types.go to remove/update
	//NCL对应 runtime class
	Ncl string `json:"ncl,omitempty"`
}

//定义clannode 清理节点标签，仅保留系统标签 定义指针s
func (s *NodepoolSpec) CleanNode(node corev1.Node) *corev1.Node {
	//只保留包含kubernetes相关的标签，除节点池标签
	//当前逻辑一个节点只属于一个节点池
	nodeLabels := map[string]string{}
	for k, v := range node.Labels {
		if strings.Contains(k, "kubernetes") {
			nodeLabels[k] = v
		}
	}
	node.Labels = nodeLabels
	//污点同理 污点是key包含kubernetes
	var taints []corev1.Taint
	for _, taint := range node.Spec.Taints {
		if strings.Contains(taint.Key, "kubernetes") {
			taints = append(taints, taint)
		}
		node.Spec.Taints = taints
	}

	return &node
}

//定义apply 方法生成node结构， 方便Patch n指定为指针类型
func (s *NodepoolSpec) ApplyNode(node corev1.Node) *corev1.Node {
	n := s.CleanNode(node)
	for k, v := range s.Labels {
		n.Labels[k] = v
	}
	n.Spec.Taints = append(n.Spec.Taints, s.Taints...)
	return n
}

// NodepoolStatus defines the observed state of Nodepool
type NodepoolStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Nodepool is the Schema for the nodepools API
type Nodepool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodepoolSpec   `json:"spec,omitempty"`
	Status NodepoolStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NodepoolList contains a list of Nodepool
type NodepoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Nodepool `json:"items"`
}

func (n *Nodepool) RuntimeClass() *v1.RuntimeClass {
	s := n.Spec
	tolerations := make([]corev1.Toleration, len(s.Taints))
	for i, t := range s.Taints {
		tolerations[i] = corev1.Toleration{
			Key:      t.Key,
			Value:    t.Value,
			Effect:   t.Effect,
			Operator: corev1.TolerationOpEqual,
		}
	}
	return &v1.RuntimeClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-pool-" + n.Name,
		},
		Handler: "runc",
		Scheduling: &v1.Scheduling{
			NodeSelector: s.Labels,
			Tolerations:  tolerations,
		},
	}
}
func (n *Nodepool) NodeRole() string {
	return "node-role.kubernetes.io/" + n.Name
}
func (n *Nodepool) NodeLabelSelector() labels.Selector {
	return labels.SelectorFromSet(map[string]string{
		n.NodeRole(): "",
	})
}
func init() {
	SchemeBuilder.Register(&Nodepool{}, &NodepoolList{})
}
