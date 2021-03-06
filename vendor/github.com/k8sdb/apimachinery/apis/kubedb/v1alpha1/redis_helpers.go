package v1alpha1

import (
	"fmt"
	"strings"

	"github.com/appscode/kutil/tools/monitoring/api"
	core "k8s.io/api/core/v1"
)

func (r Redis) OffshootName() string {
	return r.Name
}

func (r Redis) OffshootLabels() map[string]string {
	return map[string]string{
		LabelDatabaseName: r.Name,
		LabelDatabaseKind: ResourceKindRedis,
	}
}

func (r Redis) StatefulSetLabels() map[string]string {
	labels := r.OffshootLabels()
	for key, val := range r.Labels {
		if !strings.HasPrefix(key, GenericKey+"/") && !strings.HasPrefix(key, RedisKey+"/") {
			labels[key] = val
		}
	}
	return labels
}

func (r Redis) StatefulSetAnnotations() map[string]string {
	annotations := make(map[string]string)
	for key, val := range r.Annotations {
		if !strings.HasPrefix(key, GenericKey+"/") && !strings.HasPrefix(key, RedisKey+"/") {
			annotations[key] = val
		}
	}
	annotations[RedisDatabaseVersion] = string(r.Spec.Version)
	return annotations
}

func (r Redis) ResourceCode() string {
	return ResourceCodeRedis
}

func (r Redis) ResourceKind() string {
	return ResourceKindRedis
}

func (r Redis) ResourceName() string {
	return ResourceNameRedis
}

func (r Redis) ResourceType() string {
	return ResourceTypeRedis
}

func (r Redis) ObjectReference() *core.ObjectReference {
	return &core.ObjectReference{
		APIVersion:      SchemeGroupVersion.String(),
		Kind:            ResourceKindRedis,
		Namespace:       r.Namespace,
		Name:            r.Name,
		UID:             r.UID,
		ResourceVersion: r.ResourceVersion,
	}
}

func (r Redis) ServiceName() string {
	return r.OffshootName()
}

func (r Redis) ServiceMonitorName() string {
	return fmt.Sprintf("kubedb-%s-%s", r.Namespace, r.Name)
}

func (r Redis) Path() string {
	return fmt.Sprintf("/kubedb.com/v1alpha1/namespaces/%s/%s/%s/metrics", r.Namespace, r.ResourceType(), r.Name)
}

func (r Redis) Scheme() string {
	return ""
}

func (r *Redis) StatsAccessor() api.StatsAccessor {
	return r
}
