package podsecuritypolicy

import (
	"context"

	v3 "github.com/rancher/types/apis/management.cattle.io/v3"
	v12 "github.com/rancher/types/apis/rbac.authorization.k8s.io/v1"
	"github.com/rancher/types/config"
	v1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func RegisterClusterRole(ctx context.Context, context *config.UserContext) {
	c := clusterRoleHandler{
		psptLister:   context.Management.Management.PodSecurityPolicyTemplates("").Controller().Lister(),
		clusterRoles: context.RBAC.ClusterRoles(""),
	}

	context.RBAC.ClusterRoles("").AddHandler(ctx, "cluster-role-sync", c.sync)
}

type clusterRoleHandler struct {
	psptLister   v3.PodSecurityPolicyTemplateLister
	clusterRoles v12.ClusterRoleInterface
}

// sync checks if a clusterRole has a parent pspt based on the annotation and if that parent no longer
// exists will delete the clusterRole
func (c *clusterRoleHandler) sync(key string, obj *v1.ClusterRole) (runtime.Object, error) {
	if obj == nil || obj.DeletionTimestamp != nil {
		return obj, nil
	}
	if templateID, ok := obj.Annotations[podSecurityPolicyTemplateParentAnnotation]; ok {
		_, err := c.psptLister.Get("", templateID)
		if err != nil {
			// parent template is gone, delete the clusterRole
			if k8serrors.IsNotFound(err) {
				return obj, c.clusterRoles.Delete(obj.Name, &metav1.DeleteOptions{})

			}
			return obj, err
		}

	}
	return obj, nil
}
