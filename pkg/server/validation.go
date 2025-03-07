package server

import (
	"net/http"

	"github.com/rancher/rancher/pkg/apis/management.cattle.io"
	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/rancher/rancher/pkg/apis/provisioning.cattle.io"
	v1 "github.com/rancher/rancher/pkg/apis/provisioning.cattle.io/v1"
	"github.com/rancher/webhook/pkg/clients"
	"github.com/rancher/webhook/pkg/resources/validation/cluster"
	"github.com/rancher/webhook/pkg/resources/validation/clusterroletemplatebinding"
	"github.com/rancher/webhook/pkg/resources/validation/feature"
	"github.com/rancher/webhook/pkg/resources/validation/globalrole"
	"github.com/rancher/webhook/pkg/resources/validation/globalrolebinding"
	"github.com/rancher/webhook/pkg/resources/validation/machineconfig"
	"github.com/rancher/webhook/pkg/resources/validation/projectroletemplatebinding"
	"github.com/rancher/webhook/pkg/resources/validation/roletemplate"
	"github.com/rancher/wrangler/pkg/webhook"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func Validation(clients *clients.Clients) (http.Handler, error) {
	router := webhook.NewRouter()

	router.Kind("Feature").Group(management.GroupName).Type(&v3.Feature{}).Handle(feature.NewValidator())
	router.Kind("Cluster").Group(management.GroupName).Type(&v3.Cluster{}).Handle(cluster.NewValidator(clients.K8s.AuthorizationV1().SubjectAccessReviews()))
	router.Kind("Cluster").Group(provisioning.GroupName).Type(&v1.Cluster{}).Handle(cluster.NewProvisioningClusterValidator(clients))
	router.Group("rke-machine-config.cattle.io").Type(&unstructured.Unstructured{}).Handle(machineconfig.NewMachineConfigValidator())

	if clients.MultiClusterManagement {
		globalRoleBindings := globalrolebinding.NewValidator(clients.Management.GlobalRole().Cache(), clients.EscalationChecker)
		globalRoles := globalrole.NewValidator(clients.EscalationChecker)
		prtbs := projectroletemplatebinding.NewValidator(clients.Management.RoleTemplate().Cache(), clients.EscalationChecker)
		crtbs := clusterroletemplatebinding.NewValidator(clients.Management.RoleTemplate().Cache(), clients.EscalationChecker)
		roleTemplates := roletemplate.NewValidator(clients.EscalationChecker)

		router.Kind("RoleTemplate").Group(management.GroupName).Type(&v3.RoleTemplate{}).Handle(roleTemplates)
		router.Kind("GlobalRoleBinding").Group(management.GroupName).Type(&v3.GlobalRoleBinding{}).Handle(globalRoleBindings)
		router.Kind("GlobalRole").Group(management.GroupName).Type(&v3.GlobalRole{}).Handle(globalRoles)
		router.Kind("ClusterRoleTemplateBinding").Group(management.GroupName).Type(&v3.ClusterRoleTemplateBinding{}).Handle(crtbs)
		router.Kind("ProjectRoleTemplateBinding").Group(management.GroupName).Type(&v3.ProjectRoleTemplateBinding{}).Handle(prtbs)
	}

	return router, nil
}
