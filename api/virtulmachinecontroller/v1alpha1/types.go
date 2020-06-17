package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type VirtulMachinePhase string

type VirtulMachineConditionType string

type ConditionStatus string

// ResourceList is a set of (resource name, quantity) pairs.
type ResourceList map[ResourceName]resource.Quantity

// ResourceName is the name identifying various resources in a ResourceList.
type ResourceName string

type VirtulMachineAddressType string

type UniqueVolumeName string

// AttachedVolume describes a volume attached to a node
type AttachedVolume struct {
	// Name of the attached volume
	Name UniqueVolumeName `json:"name" protobuf:"bytes,1,rep,name=name"`

	// DevicePath represents the device path where the volume should be available
	DevicePath string `json:"devicePath" protobuf:"bytes,2,rep,name=devicePath"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VirtulMachine is a worker node in Kubernetes.
// Each node will have a unique identifier in the cache (i.e. in etcd).
type VirtulMachine struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Spec defines the behavior of a node.
	// https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status
	// +optional
	Spec VirtulMachineSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// Most recently observed status of the node.
	// Populated by the system.
	// Read-only.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status
	// +optional
	Status VirtulMachineStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// The node this Taint is attached to has the "effect" on
// any pod that does not tolerate the Taint.
type Taint struct {
	// Required. The taint key to be applied to a node.
	Key string `json:"key" protobuf:"bytes,1,opt,name=key"`
	// Required. The taint value corresponding to the taint key.
	// +optional
	Value string `json:"value,omitempty" protobuf:"bytes,2,opt,name=value"`
	// Required. The effect of the taint on pods
	// that do not tolerate the taint.
	// Valid effects are NoSchedule, PreferNoSchedule and NoExecute.
	Effect TaintEffect `json:"effect" protobuf:"bytes,3,opt,name=effect,casttype=TaintEffect"`
	// TimeAdded represents the time at which the taint was added.
	// It is only written for NoExecute taints.
	// +optional
	TimeAdded *metav1.Time `json:"timeAdded,omitempty" protobuf:"bytes,4,opt,name=timeAdded"`
}

type TaintEffect string

// VirtulMachineSpec describes the attributes that a node is created with.
type VirtulMachineSpec struct {
	// PodCIDR represents the pod IP range assigned to the node.
	// +optional
	PodCIDR string `json:"podCIDR,omitempty" protobuf:"bytes,1,opt,name=podCIDR"`
	// ID of the node assigned by the cloud provider in the format: <ProviderName>://<ProviderSpecificVirtulMachineID>
	// +optional
	ProviderID string `json:"providerID,omitempty" protobuf:"bytes,3,opt,name=providerID"`
	// Unschedulable controls node schedulability of new pods. By default, node is schedulable.
	// More info: https://kubernetes.io/docs/concepts/nodes/node/#manual-node-administration
	// +optional
	Unschedulable bool `json:"unschedulable,omitempty" protobuf:"varint,4,opt,name=unschedulable"`
	// If specified, the node's taints.
	// +optional
	Taints []Taint `json:"taints,omitempty" protobuf:"bytes,5,opt,name=taints"`
	// If specified, the source to get node configuration from
	// The DynamicKubeletConfig feature gate must be enabled for the Kubelet to use this field
	// +optional
	ConfigSource *VirtulMachineConfigSource `json:"configSource,omitempty" protobuf:"bytes,6,opt,name=configSource"`

	// Deprecated. Not all kubelets will set this field. Remove field after 1.13.
	// see: https://issues.k8s.io/61966
	// +optional
	DoNotUse_ExternalID string `json:"externalID,omitempty" protobuf:"bytes,2,opt,name=externalID"`
}

// VirtulMachineConfigSource specifies a source of node configuration. Exactly one subfield (excluding metadata) must be non-nil.
type VirtulMachineConfigSource struct {
	// For historical context, regarding the below kind, apiVersion, and configMapRef deprecation tags:
	// 1. kind/apiVersion were used by the kubelet to persist this struct to disk (they had no protobuf tags)
	// 2. configMapRef and proto tag 1 were used by the API to refer to a configmap,
	//    but used a generic ObjectReference type that didn't really have the fields we needed
	// All uses/persistence of the VirtulMachineConfigSource struct prior to 1.11 were gated by alpha feature flags,
	// so there was no persisted data for these fields that needed to be migrated/handled.

	// +k8s:deprecated=kind
	// +k8s:deprecated=apiVersion
	// +k8s:deprecated=configMapRef,protobuf=1

	// ConfigMap is a reference to a VirtulMachine's ConfigMap
	ConfigMap *ConfigMapVirtulMachineConfigSource `json:"configMap,omitempty" protobuf:"bytes,2,opt,name=configMap"`
}

// ConfigMapVirtulMachineConfigSource contains the information to reference a ConfigMap as a config source for the VirtulMachine.
type ConfigMapVirtulMachineConfigSource struct {
	// Namespace is the metadata.namespace of the referenced ConfigMap.
	// This field is required in all cases.
	Namespace string `json:"namespace" protobuf:"bytes,1,opt,name=namespace"`

	// Name is the metadata.name of the referenced ConfigMap.
	// This field is required in all cases.
	Name string `json:"name" protobuf:"bytes,2,opt,name=name"`

	// UID is the metadata.UID of the referenced ConfigMap.
	// This field is forbidden in VirtulMachine.Spec, and required in VirtulMachine.Status.
	// +optional
	UID types.UID `json:"uid,omitempty" protobuf:"bytes,3,opt,name=uid"`

	// ResourceVersion is the metadata.ResourceVersion of the referenced ConfigMap.
	// This field is forbidden in VirtulMachine.Spec, and required in VirtulMachine.Status.
	// +optional
	ResourceVersion string `json:"resourceVersion,omitempty" protobuf:"bytes,4,opt,name=resourceVersion"`

	// KubeletConfigKey declares which key of the referenced ConfigMap corresponds to the KubeletConfiguration structure
	// This field is required in all cases.
	KubeletConfigKey string `json:"kubeletConfigKey" protobuf:"bytes,5,opt,name=kubeletConfigKey"`
}

// VirtulMachineStatus is information about the current status of a node.
type VirtulMachineStatus struct {
	// Capacity represents the total resources of a node.
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#capacity
	// +optional
	Capacity ResourceList `json:"capacity,omitempty" protobuf:"bytes,1,rep,name=capacity,casttype=ResourceList,castkey=ResourceName"`
	// Allocatable represents the resources of a node that are available for scheduling.
	// Defaults to Capacity.
	// +optional
	Allocatable ResourceList `json:"allocatable,omitempty" protobuf:"bytes,2,rep,name=allocatable,casttype=ResourceList,castkey=ResourceName"`
	// VirtulMachinePhase is the recently observed lifecycle phase of the node.
	// More info: https://kubernetes.io/docs/concepts/nodes/node/#phase
	// The field is never populated, and now is deprecated.
	// +optional
	Phase VirtulMachinePhase `json:"phase,omitempty" protobuf:"bytes,3,opt,name=phase,casttype=VirtulMachinePhase"`
	// Conditions is an array of current observed node conditions.
	// More info: https://kubernetes.io/docs/concepts/nodes/node/#condition
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []VirtulMachineCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,4,rep,name=conditions"`
	// List of addresses reachable to the node.
	// Queried from cloud provider, if available.
	// More info: https://kubernetes.io/docs/concepts/nodes/node/#addresses
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Addresses []VirtulMachineAddress `json:"addresses,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,5,rep,name=addresses"`
	// Endpoints of daemons running on the VirtulMachine.
	// +optional
	DaemonEndpoints VirtulMachineDaemonEndpoints `json:"daemonEndpoints,omitempty" protobuf:"bytes,6,opt,name=daemonEndpoints"`
	// Set of ids/uuids to uniquely identify the node.
	// More info: https://kubernetes.io/docs/concepts/nodes/node/#info
	// +optional
	VirtulMachineInfo VirtulMachineSystemInfo `json:"nodeInfo,omitempty" protobuf:"bytes,7,opt,name=nodeInfo"`
	// List of container images on this node
	// +optional
	Images []ContainerImage `json:"images,omitempty" protobuf:"bytes,8,rep,name=images"`
	// List of attachable volumes in use (mounted) by the node.
	// +optional
	VolumesInUse []UniqueVolumeName `json:"volumesInUse,omitempty" protobuf:"bytes,9,rep,name=volumesInUse"`
	// List of volumes that are attached to the node.
	// +optional
	VolumesAttached []AttachedVolume `json:"volumesAttached,omitempty" protobuf:"bytes,10,rep,name=volumesAttached"`
	// Status of the config assigned to the node via the dynamic Kubelet config feature.
	// +optional
	Config *VirtulMachineConfigStatus `json:"config,omitempty" protobuf:"bytes,11,opt,name=config"`
}

// VirtulMachineConfigStatus describes the status of the config assigned by VirtulMachine.Spec.ConfigSource.
type VirtulMachineConfigStatus struct {
	// Assigned reports the checkpointed config the node will try to use.
	// When VirtulMachine.Spec.ConfigSource is updated, the node checkpoints the associated
	// config payload to local disk, along with a record indicating intended
	// config. The node refers to this record to choose its config checkpoint, and
	// reports this record in Assigned. Assigned only updates in the status after
	// the record has been checkpointed to disk. When the Kubelet is restarted,
	// it tries to make the Assigned config the Active config by loading and
	// validating the checkpointed payload identified by Assigned.
	// +optional
	Assigned *VirtulMachineConfigSource `json:"assigned,omitempty" protobuf:"bytes,1,opt,name=assigned"`
	// Active reports the checkpointed config the node is actively using.
	// Active will represent either the current version of the Assigned config,
	// or the current LastKnownGood config, depending on whether attempting to use the
	// Assigned config results in an error.
	// +optional
	Active *VirtulMachineConfigSource `json:"active,omitempty" protobuf:"bytes,2,opt,name=active"`
	// LastKnownGood reports the checkpointed config the node will fall back to
	// when it encounters an error attempting to use the Assigned config.
	// The Assigned config becomes the LastKnownGood config when the node determines
	// that the Assigned config is stable and correct.
	// This is currently implemented as a 10-minute soak period starting when the local
	// record of Assigned config is updated. If the Assigned config is Active at the end
	// of this period, it becomes the LastKnownGood. Note that if Spec.ConfigSource is
	// reset to nil (use local defaults), the LastKnownGood is also immediately reset to nil,
	// because the local default config is always assumed good.
	// You should not make assumptions about the node's method of determining config stability
	// and correctness, as this may change or become configurable in the future.
	// +optional
	LastKnownGood *VirtulMachineConfigSource `json:"lastKnownGood,omitempty" protobuf:"bytes,3,opt,name=lastKnownGood"`
	// Error describes any problems reconciling the Spec.ConfigSource to the Active config.
	// Errors may occur, for example, attempting to checkpoint Spec.ConfigSource to the local Assigned
	// record, attempting to checkpoint the payload associated with Spec.ConfigSource, attempting
	// to load or validate the Assigned config, etc.
	// Errors may occur at different points while syncing config. Earlier errors (e.g. download or
	// checkpointing errors) will not result in a rollback to LastKnownGood, and may resolve across
	// Kubelet retries. Later errors (e.g. loading or validating a checkpointed config) will result in
	// a rollback to LastKnownGood. In the latter case, it is usually possible to resolve the error
	// by fixing the config assigned in Spec.ConfigSource.
	// You can find additional information for debugging by searching the error message in the Kubelet log.
	// Error is a human-readable description of the error state; machines can check whether or not Error
	// is empty, but should not rely on the stability of the Error text across Kubelet versions.
	// +optional
	Error string `json:"error,omitempty" protobuf:"bytes,4,opt,name=error"`
}

// Describe a container image
type ContainerImage struct {
	// Names by which this image is known.
	// e.g. ["k8s.gcr.io/hyperkube:v1.0.7", "dockerhub.io/google_containers/hyperkube:v1.0.7"]
	Names []string `json:"names" protobuf:"bytes,1,rep,name=names"`
	// The size of the image in bytes.
	// +optional
	SizeBytes int64 `json:"sizeBytes,omitempty" protobuf:"varint,2,opt,name=sizeBytes"`
}

// VirtulMachineSystemInfo is a set of ids/uuids to uniquely identify the node.
type VirtulMachineSystemInfo struct {
	// MachineID reported by the node. For unique machine identification
	// in the cluster this field is preferred. Learn more from man(5)
	// machine-id: http://man7.org/linux/man-pages/man5/machine-id.5.html
	MachineID string `json:"machineID" protobuf:"bytes,1,opt,name=machineID"`
	// SystemUUID reported by the node. For unique machine identification
	// MachineID is preferred. This field is specific to Red Hat hosts
	// https://access.redhat.com/documentation/en-US/Red_Hat_Subscription_Management/1/html/RHSM/getting-system-uuid.html
	SystemUUID string `json:"systemUUID" protobuf:"bytes,2,opt,name=systemUUID"`
	// Boot ID reported by the node.
	BootID string `json:"bootID" protobuf:"bytes,3,opt,name=bootID"`
	// Kernel Version reported by the node from 'uname -r' (e.g. 3.16.0-0.bpo.4-amd64).
	KernelVersion string `json:"kernelVersion" protobuf:"bytes,4,opt,name=kernelVersion"`
	// OS Image reported by the node from /etc/os-release (e.g. Debian GNU/Linux 7 (wheezy)).
	OSImage string `json:"osImage" protobuf:"bytes,5,opt,name=osImage"`
	// ContainerRuntime Version reported by the node through runtime remote API (e.g. docker://1.5.0).
	ContainerRuntimeVersion string `json:"containerRuntimeVersion" protobuf:"bytes,6,opt,name=containerRuntimeVersion"`
	// Kubelet Version reported by the node.
	KubeletVersion string `json:"kubeletVersion" protobuf:"bytes,7,opt,name=kubeletVersion"`
	// KubeProxy Version reported by the node.
	KubeProxyVersion string `json:"kubeProxyVersion" protobuf:"bytes,8,opt,name=kubeProxyVersion"`
	// The Operating System reported by the node
	OperatingSystem string `json:"operatingSystem" protobuf:"bytes,9,opt,name=operatingSystem"`
	// The Architecture reported by the node
	Architecture string `json:"architecture" protobuf:"bytes,10,opt,name=architecture"`
}

// VirtulMachineDaemonEndpoints lists ports opened by daemons running on the VirtulMachine.
type VirtulMachineDaemonEndpoints struct {
	// Endpoint on which Kubelet is listening.
	// +optional
	KubeletEndpoint DaemonEndpoint `json:"kubeletEndpoint,omitempty" protobuf:"bytes,1,opt,name=kubeletEndpoint"`
}

// DaemonEndpoint contains information about a single Daemon endpoint.
type DaemonEndpoint struct {
	/*
		The port tag was not properly in quotes in earlier releases, so it must be
		uppercased for backwards compat (since it was falling back to var name of
		'Port').
	*/

	// Port number of the given endpoint.
	Port int32 `json:"Port" protobuf:"varint,1,opt,name=Port"`
}

// VirtulMachineCondition contains condition information for a node.
type VirtulMachineCondition struct {
	// Type of node condition.
	Type VirtulMachineConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=VirtulMachineConditionType"`
	// Status of the condition, one of True, False, Unknown.
	Status ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=ConditionStatus"`
	// Last time we got an update on a given condition.
	// +optional
	LastHeartbeatTime metav1.Time `json:"lastHeartbeatTime,omitempty" protobuf:"bytes,3,opt,name=lastHeartbeatTime"`
	// Last time the condition transit from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,4,opt,name=lastTransitionTime"`
	// (brief) reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,5,opt,name=reason"`
	// Human readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,6,opt,name=message"`
}

// VirtulMachineAddress contains information for the node's address.
type VirtulMachineAddress struct {
	// VirtulMachine address type, one of Hostname, ExternalIP or InternalIP.
	Type VirtulMachineAddressType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=VirtulMachineAddressType"`
	// The node address.
	Address string `json:"address" protobuf:"bytes,2,opt,name=address"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VirtulMachineList is the whole list of all VirtulMachines which have been registered with master.
type VirtulMachineList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// List of nodes
	Items []VirtulMachine `json:"items" protobuf:"bytes,2,rep,name=items"`
}
