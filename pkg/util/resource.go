package util

import (
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
	tapi "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/k8sdb/cli/pkg/decoder"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

func GetSupportedResource(resource string) (string, error) {
	switch strings.ToLower(resource) {
	case strings.ToLower(tapi.ResourceKindElasticsearch),
		strings.ToLower(tapi.ResourceTypeElasticsearch),
		strings.ToLower(tapi.ResourceCodeElasticsearch),
		strings.ToLower(tapi.ResourceNameElasticsearch):
		return tapi.ResourceTypeElasticsearch + "." + tapi.SchemeGroupVersion.Group, nil
	case strings.ToLower(tapi.ResourceKindPostgres),
		strings.ToLower(tapi.ResourceTypePostgres),
		strings.ToLower(tapi.ResourceCodePostgres),
		strings.ToLower(tapi.ResourceNamePostgres):
		return tapi.ResourceTypePostgres + "." + tapi.SchemeGroupVersion.Group, nil
	case strings.ToLower(tapi.ResourceKindSnapshot),
		strings.ToLower(tapi.ResourceTypeSnapshot),
		strings.ToLower(tapi.ResourceCodeSnapshot),
		strings.ToLower(tapi.ResourceNameSnapshot):
		return tapi.ResourceTypeSnapshot + "." + tapi.SchemeGroupVersion.Group, nil
	case strings.ToLower(tapi.ResourceKindDormantDatabase),
		strings.ToLower(tapi.ResourceTypeDormantDatabase),
		strings.ToLower(tapi.ResourceCodeDormantDatabase),
		strings.ToLower(tapi.ResourceNameDormantDatabase):
		return tapi.ResourceTypeDormantDatabase + "." + tapi.SchemeGroupVersion.Group, nil
	default:
		return "", fmt.Errorf(`kubedb doesn't support a resource type "%v"`, resource)
	}
}

func GetResourceType(resource string) (string, error) {
	switch strings.ToLower(resource) {
	case strings.ToLower(tapi.ResourceKindElasticsearch),
		strings.ToLower(tapi.ResourceTypeElasticsearch),
		strings.ToLower(tapi.ResourceCodeElasticsearch),
		strings.ToLower(tapi.ResourceNameElasticsearch):
		return tapi.ResourceTypeElasticsearch, nil
	case strings.ToLower(tapi.ResourceKindPostgres),
		strings.ToLower(tapi.ResourceTypePostgres),
		strings.ToLower(tapi.ResourceCodePostgres),
		strings.ToLower(tapi.ResourceNamePostgres):
		return tapi.ResourceTypePostgres, nil
	case strings.ToLower(tapi.ResourceKindSnapshot),
		strings.ToLower(tapi.ResourceTypeSnapshot),
		strings.ToLower(tapi.ResourceCodeSnapshot),
		strings.ToLower(tapi.ResourceNameSnapshot):
		return tapi.ResourceTypeSnapshot, nil
	case strings.ToLower(tapi.ResourceKindDormantDatabase),
		strings.ToLower(tapi.ResourceTypeDormantDatabase),
		strings.ToLower(tapi.ResourceCodeDormantDatabase),
		strings.ToLower(tapi.ResourceNameDormantDatabase):
		return tapi.ResourceTypeDormantDatabase, nil
	default:
		return "", fmt.Errorf(`kubedb doesn't support a resource type "%v"`, resource)
	}
}

func CheckSupportedResource(kind string) error {
	switch kind {
	case tapi.ResourceKindElasticsearch,
		tapi.ResourceKindPostgres,
		tapi.ResourceKindSnapshot,
		tapi.ResourceKindDormantDatabase:
		return nil
	default:
		return fmt.Errorf(`kubedb doesn't support a resource type "%v"`, kind)
	}
}

func GetAllSupportedResources(f cmdutil.Factory) ([]string, error) {

	resources := []string{
		tapi.ResourceTypeElasticsearch + "." + tapi.SchemeGroupVersion.Group,
		tapi.ResourceTypePostgres + "." + tapi.SchemeGroupVersion.Group,
		tapi.ResourceTypeSnapshot + "." + tapi.SchemeGroupVersion.Group,
		tapi.ResourceTypeDormantDatabase + "." + tapi.SchemeGroupVersion.Group,
	}

	restConfig, err := f.ClientConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := crd_cs.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	availableResources := make([]string, 0)
	for _, val := range resources {
		_, err := clientset.CustomResourceDefinitions().Get(val, metav1.GetOptions{})
		if err != nil {
			if kerr.IsNotFound(err) {
				continue
			}
			return nil, err
		}
		availableResources = append(availableResources, val)
	}

	return availableResources, nil
}

var ShortForms = map[string]string{
	tapi.ResourceCodeElasticsearch:   tapi.ResourceTypeElasticsearch,
	tapi.ResourceCodePostgres:        tapi.ResourceTypePostgres,
	tapi.ResourceCodeSnapshot:        tapi.ResourceTypeSnapshot,
	tapi.ResourceCodeDormantDatabase: tapi.ResourceTypeDormantDatabase,
}

func ResourceShortFormFor(resource string) (string, bool) {
	resourceType, err := GetResourceType(resource)
	if err != nil {
		return "", false
	}

	var alias string
	exists := false
	for k, val := range ShortForms {
		if val == resourceType {
			alias = k
			exists = true
			break
		}
	}
	return alias, exists
}

func GetObjectData(obj runtime.Object) ([]byte, error) {
	return yaml.Marshal(obj)
}

func GetStructuredObject(obj runtime.Object) (runtime.Object, error) {
	kind := obj.GetObjectKind().GroupVersionKind().Kind
	data, err := GetObjectData(obj)
	if err != nil {
		return obj, err
	}
	return decoder.Decode(kind, data)
}

func checkChainKeyUnchanged(key string, mapData map[string]interface{}) bool {
	keys := strings.Split(key, ".")
	val, ok := mapData[keys[0]]
	if !ok || len(keys) == 1 {
		return !ok
	}

	newKey := strings.Join(keys[1:], ".")
	return checkChainKeyUnchanged(newKey, val.(map[string]interface{}))
}

func RequireChainKeyUnchanged(key string) mergepatch.PreconditionFunc {
	return func(patch interface{}) bool {
		patchMap, ok := patch.(map[string]interface{})
		if !ok {
			fmt.Println("Invalid data")
			return true
		}
		return checkChainKeyUnchanged(key, patchMap)
	}
}

func GetPreconditionFunc(kind string) []mergepatch.PreconditionFunc {
	preconditions := []mergepatch.PreconditionFunc{
		mergepatch.RequireKeyUnchanged("apiVersion"),
		mergepatch.RequireKeyUnchanged("kind"),
		mergepatch.RequireMetadataKeyUnchanged("name"),
		mergepatch.RequireMetadataKeyUnchanged("namespace"),
		mergepatch.RequireKeyUnchanged("status"),
	}
	return preconditions
}

var PreconditionSpecField = map[string][]string{
	tapi.ResourceKindElasticsearch: {
		"spec.version",
		"spec.storage",
		"spec.nodeSelector",
		"spec.init",
	},
	tapi.ResourceKindPostgres: {
		"spec.version",
		"spec.storage",
		"spec.databaseSecret",
		"spec.nodeSelector",
		"spec.init",
	},
	tapi.ResourceKindDormantDatabase: {
		"spec.origin",
	},
}

func GetConditionalPreconditionFunc(kind string) []mergepatch.PreconditionFunc {
	preconditions := []mergepatch.PreconditionFunc{}

	if fields, found := PreconditionSpecField[kind]; found {
		for _, field := range fields {
			preconditions = append(preconditions,
				RequireChainKeyUnchanged(field),
			)
		}
	}

	return preconditions
}

func CheckResourceExists(client internalclientset.Interface, kind, name, namespace string) (bool, error) {
	var err error
	switch kind {
	case tapi.ResourceKindElasticsearch:
		statefulSetName := fmt.Sprintf("%v-%v", name, tapi.ResourceCodeElasticsearch)
		_, err = client.Apps().StatefulSets(namespace).Get(statefulSetName, metav1.GetOptions{})
	case tapi.ResourceKindPostgres:
		statefulSetName := fmt.Sprintf("%v-%v", name, tapi.ResourceCodePostgres)
		_, err = client.Apps().StatefulSets(namespace).Get(statefulSetName, metav1.GetOptions{})
	}

	if err != nil {
		if kerr.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func IsPreconditionFailed(err error) bool {
	_, ok := err.(errPreconditionFailed)
	return ok
}

type errPreconditionFailed struct {
	message string
}

func newErrPreconditionFailed(target map[string]interface{}) errPreconditionFailed {
	s := fmt.Sprintf("precondition failed for: %v", target)
	return errPreconditionFailed{s}
}

func (err errPreconditionFailed) Error() string {
	return err.message
}

func CheckConditionalPrecondition(patchData []byte, fns ...mergepatch.PreconditionFunc) error {
	patch := make(map[string]interface{})
	if err := json.Unmarshal(patchData, &patch); err != nil {
		return err
	}
	for _, fn := range fns {
		if !fn(patch) {
			return newErrPreconditionFailed(patch)
		}
	}
	return nil
}
