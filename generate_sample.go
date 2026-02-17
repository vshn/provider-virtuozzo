//go:build generate

// Clean samples dir
//go:generate rm -rf ./samples/*

// Generate sample files
//go:generate go run generate_sample.go ./samples

package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/vshn/provider-virtuozzo/apis"
	virtuozzov1 "github.com/vshn/provider-virtuozzo/apis/virtuozzo/v1"
	admissionv1 "k8s.io/api/admission/v1"
	authv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	serializerjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
)

var scheme = runtime.NewScheme()

func main() {
	failIfError(apis.AddToScheme(scheme))
	generateBucketSample()
	generateBucketAdmissionRequest()
}

func generateBucketSample() {
	spec := newBucketSample()
	serialize(spec, true)
}

func newBucketSample() *virtuozzov1.Bucket {
	return &virtuozzov1.Bucket{
		TypeMeta: metav1.TypeMeta{
			APIVersion: virtuozzov1.BucketGroupVersionKind.GroupVersion().String(),
			Kind:       virtuozzov1.BucketKind,
		},
		ObjectMeta: metav1.ObjectMeta{Name: "bucket"},
		Spec: virtuozzov1.BucketSpec{
			ForProvider: virtuozzov1.BucketParameters{
				BucketName:           "my-virtuozzo-test-bucket",
				BucketDeletionPolicy: virtuozzov1.DeleteAll,
			},
		},
	}
}

// generateBucketAdmissionRequest generates an update request that will fail.
func generateBucketAdmissionRequest() {
	oldSpec := newBucketSample()
	newSpec := newBucketSample()
	newSpec.Spec.ForProvider.BucketName = "another"
	oldSpec.Status.AtProvider.BucketName = oldSpec.Spec.ForProvider.BucketName

	gvk := metav1.GroupVersionKind{Group: virtuozzov1.Group, Version: virtuozzov1.Version, Kind: virtuozzov1.BucketKind}
	gvr := metav1.GroupVersionResource{Group: virtuozzov1.Group, Version: virtuozzov1.Version, Resource: virtuozzov1.BucketKind}
	admission := &admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{APIVersion: "admission.k8s.io/v1", Kind: "AdmissionReview"},
		Request: &admissionv1.AdmissionRequest{
			Object:          runtime.RawExtension{Object: newSpec},
			OldObject:       runtime.RawExtension{Object: oldSpec},
			Kind:            gvk,
			Resource:        gvr,
			RequestKind:     &gvk,
			RequestResource: &gvr,
			Name:            oldSpec.Name,
			Operation:       admissionv1.Update,
			UserInfo: authv1.UserInfo{
				Username: "admin",
				Groups:   []string{"system:authenticated"},
			},
		},
	}
	serialize(admission, false)
}

func failIfError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func serialize(object runtime.Object, useYaml bool) {
	gvk := object.GetObjectKind().GroupVersionKind()
	fileExt := "json"
	if useYaml {
		fileExt = "yaml"
	}
	fileName := fmt.Sprintf("%s_%s.%s", strings.ToLower(gvk.Group), strings.ToLower(gvk.Kind), fileExt)
	f := prepareFile(fileName)
	err := serializerjson.NewSerializerWithOptions(serializerjson.DefaultMetaFactory, scheme, scheme, serializerjson.SerializerOptions{Yaml: useYaml, Pretty: true}).Encode(object, f)
	failIfError(err)
}

func prepareFile(file string) io.Writer {
	dir := os.Args[1]
	err := os.MkdirAll(os.Args[1], 0775)
	failIfError(err)
	f, err := os.Create(filepath.Join(dir, file))
	failIfError(err)
	return f
}
