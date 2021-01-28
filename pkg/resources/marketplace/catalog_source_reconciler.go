package marketplace

import (
	"context"
	integreatlyv1alpha1 "github.com/integr8ly/integreatly-operator/pkg/apis/integreatly/v1alpha1"
)

type CatalogSourceReconciler interface {
	Reconcile(ctx context.Context) (integreatlyv1alpha1.StatusPhase, error)
	CatalogSourceName() string
}
