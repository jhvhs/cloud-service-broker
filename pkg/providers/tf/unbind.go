package tf

import (
	"context"

	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/cloud-service-broker/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/utils/correlation"
)

// Unbind performs a terraform destroy on the binding.
func (provider *TerraformProvider) Unbind(ctx context.Context, instanceGUID, bindingID string, vc *varcontext.VarContext) error {
	tfID := generateTfID(instanceGUID, bindingID)
	provider.logger.Debug("terraform-unbind", correlation.ID(ctx), lager.Data{
		"instance": instanceGUID,
		"binding":  bindingID,
		"tfId":     tfID,
	})

	if err := provider.UpdateWorkspaceHCL(tfID, provider.serviceDefinition.BindSettings, vc.ToMap()); err != nil {
		return err
	}

	if err := provider.destroy(ctx, tfID, vc.ToMap(), models.UnbindOperationType); err != nil {
		return err
	}

	return provider.Wait(ctx, tfID)
}
