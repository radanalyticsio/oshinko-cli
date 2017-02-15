package api

import (
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/runtime"
)

const GroupName = ""
const FutureGroupName = "image.openshift.io"

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = unversioned.GroupVersion{Group: GroupName, Version: runtime.APIVersionInternal}

// Kind takes an unqualified kind and returns back a Group qualified GroupKind
func Kind(kind string) unversioned.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// Resource takes an unqualified resource and returns back a Group qualified GroupResource
func Resource(resource string) unversioned.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

// Adds the list of known types to api.Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&Image{},
		&ImageList{},
		&DockerImage{},
		&ImageSignature{},
		&ImageStream{},
		&ImageStreamList{},
		&ImageStreamMapping{},
		&ImageStreamTag{},
		&ImageStreamTagList{},
		&ImageStreamImage{},
		&ImageStreamImport{},
	)
	return nil
}

func (obj *Image) GetObjectKind() unversioned.ObjectKind              { return &obj.TypeMeta }
func (obj *ImageList) GetObjectKind() unversioned.ObjectKind          { return &obj.TypeMeta }
func (obj *DockerImage) GetObjectKind() unversioned.ObjectKind        { return &obj.TypeMeta }
func (obj *ImageSignature) GetObjectKind() unversioned.ObjectKind     { return &obj.TypeMeta }
func (obj *ImageStream) GetObjectKind() unversioned.ObjectKind        { return &obj.TypeMeta }
func (obj *ImageStreamList) GetObjectKind() unversioned.ObjectKind    { return &obj.TypeMeta }
func (obj *ImageStreamMapping) GetObjectKind() unversioned.ObjectKind { return &obj.TypeMeta }
func (obj *ImageStreamTag) GetObjectKind() unversioned.ObjectKind     { return &obj.TypeMeta }
func (obj *ImageStreamTagList) GetObjectKind() unversioned.ObjectKind { return &obj.TypeMeta }
func (obj *ImageStreamImage) GetObjectKind() unversioned.ObjectKind   { return &obj.TypeMeta }
func (obj *ImageStreamImport) GetObjectKind() unversioned.ObjectKind  { return &obj.TypeMeta }
