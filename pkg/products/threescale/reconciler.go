package threescale

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/integr8ly/integreatly-operator/pkg/resources/quota"

	"net/http"
	"strconv"
	"strings"

	"github.com/integr8ly/integreatly-operator/pkg/metrics"

	l "github.com/integr8ly/integreatly-operator/pkg/resources/logger"

	envoyapi "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoycore "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	consolev1 "github.com/openshift/api/console/v1"
	oauthv1 "github.com/openshift/api/oauth/v1"

	"github.com/integr8ly/integreatly-operator/pkg/resources/events"
	"github.com/integr8ly/integreatly-operator/pkg/resources/ratelimit"

	"github.com/integr8ly/integreatly-operator/pkg/products/monitoring"
	"github.com/integr8ly/integreatly-operator/pkg/resources/backup"
	"github.com/integr8ly/integreatly-operator/pkg/resources/owner"
	"github.com/integr8ly/integreatly-operator/version"
	rbacv1 "k8s.io/api/rbac/v1"
	k8sresource "k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"

	crov1 "github.com/integr8ly/cloud-resource-operator/apis/integreatly/v1alpha1"
	"github.com/integr8ly/cloud-resource-operator/apis/integreatly/v1alpha1/types"
	croUtil "github.com/integr8ly/cloud-resource-operator/pkg/client"
	userHelper "github.com/integr8ly/integreatly-operator/pkg/resources/user"

	threescalev1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	monitoringv1alpha1 "github.com/integr8ly/application-monitoring-operator/pkg/apis/applicationmonitoring/v1alpha1"
	integreatlyv1alpha1 "github.com/integr8ly/integreatly-operator/apis/v1alpha1"
	keycloak "github.com/keycloak/keycloak-operator/pkg/apis/keycloak/v1alpha1"

	"github.com/integr8ly/integreatly-operator/pkg/config"
	"github.com/integr8ly/integreatly-operator/pkg/products/rhsso"
	"github.com/integr8ly/integreatly-operator/pkg/resources"
	"github.com/integr8ly/integreatly-operator/pkg/resources/marketplace"

	"github.com/integr8ly/integreatly-operator/pkg/resources/constants"
	appsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	usersv1 "github.com/openshift/api/user/v1"
	appsv1Client "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	oauthClient "github.com/openshift/client-go/oauth/clientset/versioned/typed/oauth/v1"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	defaultInstallationNamespace = "3scale"
	manifestPackage              = "integreatly-3scale"
	apiManagerName               = "3scale"
	clientID                     = "3scale"
	rhssoIntegrationName         = "rhsso"

	s3CredentialsSecretName        = "s3-credentials"
	externalRedisSecretName        = "system-redis"
	externalBackendRedisSecretName = "backend-redis"
	externalPostgresSecretName     = "system-database"
	apicastStagingDCName           = "apicast-staging"
	apicastProductionDCName        = "apicast-production"
	backendListenerDCName          = "backend-listener"
	systemSeedSecretName           = "system-seed"
	systemMasterApiCastSecretName  = "system-master-apicast"
	systemAppDCName                = "system-app"

	apicastRatelimiting              = "apicast-ratelimit"
	backendListenerEnvoyConfigNodeID = "backend-listener-envoyconfig"
	registrySecretName               = "threescale-registry-auth"

	threeScaleIcon = "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCAxMDAgMTAwIj48ZGVmcz48c3R5bGU+LmNscy0xe2ZpbGw6I2Q3MWUwMDt9LmNscy0ye2ZpbGw6I2MyMWEwMDt9LmNscy0ze2ZpbGw6I2ZmZjt9PC9zdHlsZT48L2RlZnM+PHRpdGxlPnByb2R1Y3RpY29uc18xMDE3X1JHQl9BUEkgZmluYWwgY29sb3I8L3RpdGxlPjxnIGlkPSJMYXllcl8xIiBkYXRhLW5hbWU9IkxheWVyIDEiPjxjaXJjbGUgY2xhc3M9ImNscy0xIiBjeD0iNTAiIGN5PSI1MCIgcj0iNTAiIHRyYW5zZm9ybT0idHJhbnNsYXRlKC0yMC43MSA1MCkgcm90YXRlKC00NSkiLz48cGF0aCBjbGFzcz0iY2xzLTIiIGQ9Ik04NS4zNiwxNC42NEE1MCw1MCwwLDAsMSwxNC42NCw4NS4zNloiLz48cGF0aCBjbGFzcz0iY2xzLTMiIGQ9Ik01MC4yNSwzMC44M2EyLjY5LDIuNjksMCwxLDAtMi42OC0yLjY5QTIuNjUsMi42NSwwLDAsMCw1MC4yNSwzMC44M1pNNDMuMzYsMzkuNGEzLjM1LDMuMzUsMCwwLDAsMy4zMiwzLjM0LDMuMzQsMy4zNCwwLDAsMCwwLTYuNjdBMy4zNSwzLjM1LDAsMCwwLDQzLjM2LDM5LjRabTMuOTIsOS44OUEyLjY4LDIuNjgsMCwxLDAsNDQuNiw1MiwyLjcsMi43LDAsMCwwLDQ3LjI4LDQ5LjI5Wk0zMi42MywyOS42NWEzLjI2LDMuMjYsMCwxLDAtMy4yNC0zLjI2QTMuMjYsMy4yNiwwLDAsMCwzMi42MywyOS42NVpNNDAuNTMsMzRhMi43NywyLjc3LDAsMCwwLDAtNS41MywyLjc5LDIuNzksMCwwLDAtMi43NiwyLjc3QTIuODUsMi44NSwwLDAsMCw0MC41MywzNFptMS43Ni05LjMxYTQuNCw0LjQsMCwxLDAtNC4zOC00LjRBNC4zNyw0LjM3LDAsMCwwLDQyLjI5LDI0LjcxWk0zMi43OCw0OWE3LDcsMCwxLDAtNy03QTcsNywwLDAsMCwzMi43OCw0OVptMzIuMTMtNy43YTQuMjMsNC4yMywwLDAsMCw0LjMsNC4zMSw0LjMxLDQuMzEsMCwxLDAtNC4zLTQuMzFabTYuOSwxMC4wNmEzLjA4LDMuMDgsMCwxLDAsMy4wOC0zLjA5QTMuMDksMy4wOSwwLDAsMCw3MS44MSw1MS4zOFpNNzMuOSwzNC43N2E0LjMxLDQuMzEsMCwxLDAtNC4zLTQuMzFBNC4yOCw0LjI4LDAsMCwwLDczLjksMzQuNzdaTTUyLjE2LDQ1LjA2YTMuNjUsMy42NSwwLDEsMCwzLjY1LTMuNjZBMy42NCwzLjY0LDAsMCwwLDUyLjE2LDQ1LjA2Wk01NSwyMmEzLjE3LDMuMTcsMCwwLDAsMy4xNi0zLjE3QTMuMjMsMy4yMywwLDAsMCw1NSwxNS42MywzLjE3LDMuMTcsMCwwLDAsNTUsMjJabS0uNDcsMTAuMDlBNS4zNyw1LjM3LDAsMCwwLDYwLDM3LjU0YTUuNDgsNS40OCwwLDEsMC01LjQ1LTUuNDhaTTY2LjI1LDI1LjVhMi42OSwyLjY5LDAsMSwwLTIuNjgtMi42OUEyLjY1LDIuNjUsMCwwLDAsNjYuMjUsMjUuNVpNNDUuNyw2My4xYTMuNDIsMy40MiwwLDEsMC0zLjQxLTMuNDJBMy40MywzLjQzLDAsMCwwLDQ1LjcsNjMuMVptMTQsMTEuMTlhNC40LDQuNCwwLDEsMCw0LjM4LDQuNEE0LjM3LDQuMzcsMCwwLDAsNTkuNzMsNzQuMjlaTTYyLjMsNTAuNTFhOS4yLDkuMiwwLDEsMCw5LjE2LDkuMkE5LjIyLDkuMjIsMCwwLDAsNjIuMyw1MC41MVpNNTAuMSw2Ni43N2EyLjY5LDIuNjksMCwxLDAsMi42OCwyLjY5QTIuNywyLjcsMCwwLDAsNTAuMSw2Ni43N1pNODEuMjUsNDEuMTJhMi43LDIuNywwLDAsMC0yLjY4LDIuNjksMi42NSwyLjY1LDAsMCwwLDIuNjgsMi42OSwyLjY5LDIuNjksMCwwLDAsMC01LjM3Wk00NC40OSw3Ni40N2EzLjczLDMuNzMsMCwwLDAtMy43MywzLjc0LDMuNzcsMy43NywwLDEsMCwzLjczLTMuNzRaTTc5LjA2LDU2LjcyYTQsNCwwLDEsMCw0LDRBNCw0LDAsMCwwLDc5LjA2LDU2LjcyWm0tNiwxMS43OEEzLjA5LDMuMDksMCwwLDAsNzAsNzEuNmEzLDMsMCwwLDAsMy4wOCwzLjA5LDMuMDksMy4wOSwwLDAsMCwwLTYuMTlaTTI4LjMsNjhhNC4xNiw0LjE2LDAsMCwwLTQuMTQsNC4xNUE0LjIxLDQuMjEsMCwwLDAsMjguMyw3Ni4zYTQuMTUsNC4xNSwwLDAsMCwwLTguM1ptLTguMjItOWEzLDMsMCwxLDAsMywzQTMuMDUsMy4wNSwwLDAsMCwyMC4wOCw1OVptMS44NC05Ljc0YTMsMywwLDEsMCwzLDNBMy4wNSwzLjA1LDAsMCwwLDIxLjkxLDQ5LjIyWk0yMi4zNyw0MmEzLjI0LDMuMjQsMCwxLDAtMy4yNCwzLjI2QTMuMjYsMy4yNiwwLDAsMCwyMi4zNyw0MlpNNDMuMTEsNzAuMmEzLjgsMy44LDAsMCwwLTMuODEtMy43NCwzLjczLDMuNzMsMCwwLDAtMy43MywzLjc0QTMuOCwzLjgsMCwwLDAsMzkuMyw3NCwzLjg3LDMuODcsMCwwLDAsNDMuMTEsNzAuMlpNMzcuNTYsNTguNDNhNC42OCw0LjY4LDAsMCwwLTQuNjItNC42NCw0LjYzLDQuNjMsMCwwLDAtNC42Miw0LjY0LDQuNTgsNC41OCwwLDAsMCw0LjYyLDQuNjRBNC42Myw0LjYzLDAsMCwwLDM3LjU2LDU4LjQzWk0yMy4xMSwzMy44MmEyLjUyLDIuNTIsMCwxLDAtMi41MS0yLjUyQTIuNTMsMi41MywwLDAsMCwyMy4xMSwzMy44MloiLz48L2c+PC9zdmc+"

	user3ScaleID = "3scale_user_id"
)

var (
	threeScaleDeploymentConfigs = []string{
		"apicast-production",
		"apicast-staging",
		"backend-cron",
		"backend-listener",
		"backend-worker",
		"system-app",
		"system-memcache",
		"system-sidekiq",
		"system-sphinx",
		"zync",
		"zync-database",
		"zync-que",
	}
)

func NewReconciler(configManager config.ConfigReadWriter, installation *integreatlyv1alpha1.RHMI, appsv1Client appsv1Client.AppsV1Interface, oauthv1Client oauthClient.OauthV1Interface, tsClient ThreeScaleInterface, mpm marketplace.MarketplaceInterface, recorder record.EventRecorder, logger l.Logger, productDeclaration *marketplace.ProductDeclaration) (*Reconciler, error) {
	if productDeclaration == nil {
		return nil, fmt.Errorf("no product declaration found for 3scale")
	}

	ns := installation.Spec.NamespacePrefix + defaultInstallationNamespace
	threescaleConfig, err := configManager.ReadThreeScale()
	if err != nil {
		return nil, fmt.Errorf("could not retrieve threescale config: %w", err)
	}
	if threescaleConfig.GetNamespace() == "" {
		threescaleConfig.SetNamespace(ns)
		configManager.WriteConfig(threescaleConfig)
	}
	if threescaleConfig.GetOperatorNamespace() == "" {
		if installation.Spec.OperatorsInProductNamespace {
			threescaleConfig.SetOperatorNamespace(threescaleConfig.GetNamespace())
		} else {
			threescaleConfig.SetOperatorNamespace(threescaleConfig.GetNamespace() + "-operator")
		}
	}
	threescaleConfig.SetBlackboxTargetPathForAdminUI("/p/login/")

	return &Reconciler{
		ConfigManager: configManager,
		Config:        threescaleConfig,
		mpm:           mpm,
		installation:  installation,
		tsClient:      tsClient,
		appsv1Client:  appsv1Client,
		oauthv1Client: oauthv1Client,
		Reconciler:    resources.NewReconciler(mpm).WithProductDeclaration(*productDeclaration),
		recorder:      recorder,
		log:           logger,
	}, nil
}

type Reconciler struct {
	ConfigManager config.ConfigReadWriter
	Config        *config.ThreeScale
	mpm           marketplace.MarketplaceInterface
	installation  *integreatlyv1alpha1.RHMI
	tsClient      ThreeScaleInterface
	appsv1Client  appsv1Client.AppsV1Interface
	oauthv1Client oauthClient.OauthV1Interface
	*resources.Reconciler
	extraParams map[string]string
	recorder    record.EventRecorder
	log         l.Logger
}

func (r *Reconciler) GetPreflightObject(ns string) runtime.Object {
	return &appsv1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "system-app",
			Namespace: ns,
		},
	}
}

func (r *Reconciler) VerifyVersion(installation *integreatlyv1alpha1.RHMI) bool {
	return version.VerifyProductAndOperatorVersion(
		installation.Status.Stages[integreatlyv1alpha1.ProductsStage].Products[integreatlyv1alpha1.Product3Scale],
		string(integreatlyv1alpha1.Version3Scale),
		string(integreatlyv1alpha1.OperatorVersion3Scale),
	)
}

func (r *Reconciler) Reconcile(ctx context.Context, installation *integreatlyv1alpha1.RHMI, product *integreatlyv1alpha1.RHMIProductStatus, serverClient k8sclient.Client, productConfig quota.ProductConfig) (integreatlyv1alpha1.StatusPhase, error) {
	r.log.Info("Start Reconciling")

	operatorNamespace := r.Config.GetOperatorNamespace()
	productNamespace := r.Config.GetNamespace()

	phase, err := r.ReconcileFinalizer(ctx, serverClient, installation, string(r.Config.GetProductName()), func() (integreatlyv1alpha1.StatusPhase, error) {
		if installation.Spec.Type == string(integreatlyv1alpha1.InstallationTypeManagedApi) {
			phase, err := ratelimit.DeleteEnvoyConfigsInNamespaces(ctx, serverClient, productNamespace)
			if err != nil || phase != integreatlyv1alpha1.PhaseCompleted {
				return phase, err
			}
		}

		phase, err := resources.RemoveNamespace(ctx, installation, serverClient, productNamespace, r.log)
		if err != nil || phase != integreatlyv1alpha1.PhaseCompleted {
			return phase, err
		}

		phase, err = resources.RemoveNamespace(ctx, installation, serverClient, operatorNamespace, r.log)
		if err != nil || phase != integreatlyv1alpha1.PhaseCompleted {
			return phase, err
		}

		err = resources.RemoveOauthClient(r.oauthv1Client, r.getOAuthClientName(), r.log)
		if err != nil {
			return integreatlyv1alpha1.PhaseFailed, err
		}

		phase, err = r.deleteConsoleLink(ctx, serverClient)
		if err != nil || phase != integreatlyv1alpha1.PhaseCompleted {
			return phase, err
		}

		return integreatlyv1alpha1.PhaseCompleted, nil
	}, r.log)
	if err != nil || phase != integreatlyv1alpha1.PhaseCompleted {
		events.HandleError(r.recorder, installation, phase, "Failed to reconcile finalizer", err)
		return phase, err
	}

	phase, err = r.ReconcileNamespace(ctx, operatorNamespace, installation, serverClient, r.log)
	if err != nil || phase != integreatlyv1alpha1.PhaseCompleted {
		events.HandleError(r.recorder, installation, phase, fmt.Sprintf("Failed to reconcile %s ns", operatorNamespace), err)
		return phase, err
	}

	phase, err = r.ReconcileNamespace(ctx, productNamespace, installation, serverClient, r.log)
	if err != nil || phase != integreatlyv1alpha1.PhaseCompleted {
		events.HandleError(r.recorder, installation, phase, fmt.Sprintf("Failed to reconcile %s ns", productNamespace), err)
		return phase, err
	}

	phase, err = r.restoreSystemSecrets(ctx, serverClient, installation)
	if err != nil || phase != integreatlyv1alpha1.PhaseCompleted {
		events.HandleError(r.recorder, installation, phase, fmt.Sprintf("Failed to reconcile %s ns", productNamespace), err)
		return phase, err
	}

	err = resources.CopyPullSecretToNameSpace(ctx, installation.GetPullSecretSpec(), productNamespace, registrySecretName, serverClient)
	if err != nil {
		events.HandleError(r.recorder, installation, phase, "Failed to reconcile pull secret", err)
		return integreatlyv1alpha1.PhaseFailed, err
	}

	phase, err = r.reconcileSubscription(ctx, serverClient, installation, productNamespace, operatorNamespace)
	if err != nil || phase != integreatlyv1alpha1.PhaseCompleted {
		events.HandleError(r.recorder, installation, phase, fmt.Sprintf("Failed to reconcile %s subscription", constants.ThreeScaleSubscriptionName), err)
		return phase, err
	}

	if r.installation.GetDeletionTimestamp() == nil {
		phase, err = r.reconcileSMTPCredentials(ctx, serverClient)
		if err != nil || phase != integreatlyv1alpha1.PhaseCompleted {
			events.HandleError(r.recorder, installation, phase, "Failed to reconcile smtp credentials", err)
			return phase, err
		}

		phase, err = r.reconcileExternalDatasources(ctx, serverClient)
		if err != nil || phase != integreatlyv1alpha1.PhaseCompleted {
			events.HandleError(r.recorder, installation, phase, "Failed to reconcile external data sources", err)
			return phase, err
		}

		phase, err = r.reconcileBlobStorage(ctx, serverClient)
		if err != nil || phase != integreatlyv1alpha1.PhaseCompleted {
			events.HandleError(r.recorder, installation, phase, "Failed to reconcile blob storage", err)
			return phase, err
		}
	}

	phase, err = r.reconcileComponents(ctx, serverClient, productConfig)
	r.log.Infof("reconcileComponents", l.Fields{"phase": phase})
	if err != nil || phase != integreatlyv1alpha1.PhaseCompleted {
		events.HandleError(r.recorder, installation, phase, "Failed to reconcile components", err)
		return phase, err
	}

	r.log.Info("Successfully deployed")

	phase, err = r.reconcileOutgoingEmailAddress(ctx, serverClient)
	r.log.Infof("reconcileOutgoingEmailAddress", l.Fields{"phase": phase})
	if err != nil || phase != integreatlyv1alpha1.PhaseCompleted {
		if err != nil {
			r.log.Warning("Failed to reconcileOutgoingEmailAddress: " + err.Error())
			events.HandleError(r.recorder, installation, phase, "Failed to reconcileOutgoingEmailAddress", err)
		}
		return phase, err
	}

	phase, err = r.reconcileRHSSOIntegration(ctx, serverClient)
	r.log.Infof("reconcileRHSSOIntegration", l.Fields{"phase": phase})
	if err != nil || phase != integreatlyv1alpha1.PhaseCompleted {
		events.HandleError(r.recorder, installation, phase, "Failed to reconcile rhsso integration", err)
		return phase, err
	}

	phase, err = r.reconcileBlackboxTargets(ctx, installation, serverClient)
	r.log.Infof("reconcileBlackboxTargets", l.Fields{"phase": phase})
	if err != nil || phase != integreatlyv1alpha1.PhaseCompleted {
		events.HandleError(r.recorder, installation, phase, "Failed to reconcile blackbox targets", err)
		return phase, err
	}

	phase, err = r.reconcileOpenshiftUsers(ctx, installation, serverClient)
	r.log.Infof("reconcileOpenshiftUsers", l.Fields{"phase": phase})
	if err != nil || phase != integreatlyv1alpha1.PhaseCompleted {
		events.HandleError(r.recorder, installation, phase, "Failed to reconcile openshift users", err)
		return phase, err
	}

	clientSecret, err := r.getOauthClientSecret(ctx, serverClient)
	if err != nil {
		events.HandleError(r.recorder, installation, integreatlyv1alpha1.PhaseFailed, "Failed to get oauth client secret", err)
		return integreatlyv1alpha1.PhaseFailed, err
	}

	threescaleMasterRoute, err := r.getThreescaleRoute(ctx, serverClient, "system-master", nil)
	if err != nil || threescaleMasterRoute == nil {
		return integreatlyv1alpha1.PhaseFailed, err
	}
	phase, err = r.ReconcileOauthClient(ctx, installation, &oauthv1.OAuthClient{
		ObjectMeta: metav1.ObjectMeta{
			Name: r.getOAuthClientName(),
		},
		Secret: clientSecret,
		RedirectURIs: []string{
			"https://" + threescaleMasterRoute.Spec.Host,
		},
		GrantMethod: oauthv1.GrantHandlerAuto,
	}, serverClient)
	r.log.Infof("ReconcileOauthClient", l.Fields{"phase": phase})
	if err != nil || phase != integreatlyv1alpha1.PhaseCompleted {
		events.HandleError(r.recorder, installation, phase, "Failed to reconcile oauth client", err)
		return phase, err
	}

	phase, err = r.reconcileServiceDiscovery(ctx, serverClient)
	r.log.Infof("reconcileServiceDiscovery", l.Fields{"phase": phase})
	if err != nil || phase != integreatlyv1alpha1.PhaseCompleted {
		events.HandleError(r.recorder, installation, phase, "Failed to reconcile service discovery", err)
		return phase, err
	}

	phase, err = r.backupSystemSecrets(ctx, serverClient, installation)
	r.log.Infof("backupSystemSecrets", l.Fields{"phase": phase})
	if err != nil || phase != integreatlyv1alpha1.PhaseCompleted {
		events.HandleError(r.recorder, installation, phase, "Failed to reconcile templates", err)
		return phase, err
	}

	phase, err = r.reconcileRouteEditRole(ctx, serverClient)
	r.log.Infof("reconcileRouteEditRole", l.Fields{"phase": phase})
	if err != nil || phase != integreatlyv1alpha1.PhaseCompleted {
		events.HandleError(r.recorder, installation, phase, "Failed to reconcile roles", err)
		return phase, err
	}

	if installation.Spec.Type == string(integreatlyv1alpha1.InstallationTypeManagedApi) {

		phase, err = r.reconcileRatelimitingTo3scaleComponents(ctx, serverClient, r.installation)
		if err != nil || phase != integreatlyv1alpha1.PhaseCompleted {
			events.HandleError(r.recorder, installation, phase, "Failed to reconcile rate limiting to 3scale components", err)
			return phase, err
		}

		alertsReconciler := r.newEnvoyAlertReconciler(r.log, r.installation.Spec.Type)
		if phase, err := alertsReconciler.ReconcileAlerts(ctx, serverClient); err != nil || phase != integreatlyv1alpha1.PhaseCompleted {
			events.HandleError(r.recorder, installation, phase, "Failed to reconcile threescale alerts", err)
			return phase, err
		}
	}

	alertsReconciler := r.newAlertReconciler(r.log, r.installation.Spec.Type)
	if phase, err := alertsReconciler.ReconcileAlerts(ctx, serverClient); err != nil || phase != integreatlyv1alpha1.PhaseCompleted {
		events.HandleError(r.recorder, installation, phase, "Failed to reconcile threescale alerts", err)
		return phase, err
	}

	phase, err = r.reconcileConsoleLink(ctx, serverClient)
	r.log.Infof("reconcileConsoleLink", l.Fields{"phase": phase})
	if err != nil || phase != integreatlyv1alpha1.PhaseCompleted {
		events.HandleError(r.recorder, installation, phase, "Failed to reconcile console link", err)
		return phase, err
	}

	phase, err = r.reconcileDeploymentConfigs(ctx, serverClient, productNamespace)
	r.log.Infof("reconcileDeploymentConfigs", l.Fields{"phase": phase})
	if err != nil || phase != integreatlyv1alpha1.PhaseCompleted {
		events.HandleError(r.recorder, installation, phase, "Failed to reconcile deployment configs", err)
		return phase, err
	}

	phase, err = r.changesDeploymentConfigsEnvVar(ctx, serverClient)
	r.log.Infof("changesDeploymentConfigsEnvVar", l.Fields{"phase": phase})
	if err != nil || phase != integreatlyv1alpha1.PhaseCompleted {
		events.HandleError(r.recorder, installation, phase, "Failed to change deployment config envvars", err)
		return phase, err
	}

	// Ensure deployment configs are ready before returning phase complete
	phase, err = r.ensureDeploymentConfigsReady(ctx, serverClient, productNamespace)
	r.log.Infof("ensureDeploymentConfigsReady", l.Fields{"phase": phase})
	if err != nil || phase != integreatlyv1alpha1.PhaseCompleted {
		events.HandleError(r.recorder, installation, phase, "Failed to ensure deployment configs are ready", err)
		return phase, err
	}

	product.Host = r.Config.GetHost()
	product.Version = r.Config.GetProductVersion()
	product.OperatorVersion = r.Config.GetOperatorVersion()

	events.HandleProductComplete(r.recorder, installation, integreatlyv1alpha1.ProductsStage, r.Config.GetProductName())
	r.log.Infof("Installation reconciled successfully", l.Fields{"product": r.Config.GetProductName()})
	return integreatlyv1alpha1.PhaseCompleted, nil
}

// restores seed and master api cast secrets if available
func (r *Reconciler) restoreSystemSecrets(ctx context.Context, serverClient k8sclient.Client, installation *integreatlyv1alpha1.RHMI) (integreatlyv1alpha1.StatusPhase, error) {
	for _, secretName := range []string{systemSeedSecretName, systemMasterApiCastSecretName} {
		err := resources.CopySecret(ctx, serverClient, secretName, installation.Namespace, secretName, r.Config.GetNamespace())
		if err != nil {
			if !k8serr.IsNotFound(err) && !k8serr.IsConflict(err) {
				return integreatlyv1alpha1.PhaseFailed, err
			}
			r.log.Info(fmt.Sprintf("no backed up secret %v found in %v", secretName, installation.Namespace))
		}
	}

	return integreatlyv1alpha1.PhaseCompleted, nil
}

// Copies the seed and master api cast secrets for later restoration
func (r *Reconciler) backupSystemSecrets(ctx context.Context, serverClient k8sclient.Client, installation *integreatlyv1alpha1.RHMI) (integreatlyv1alpha1.StatusPhase, error) {
	for _, secretName := range []string{systemSeedSecretName, systemMasterApiCastSecretName} {
		err := resources.CopySecret(ctx, serverClient, secretName, r.Config.GetNamespace(), secretName, installation.Namespace)
		if err != nil {
			return integreatlyv1alpha1.PhaseFailed, err
		}
	}
	return integreatlyv1alpha1.PhaseCompleted, nil
}

func (r *Reconciler) getOauthClientSecret(ctx context.Context, serverClient k8sclient.Client) (string, error) {
	oauthClientSecrets := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: r.ConfigManager.GetOauthClientsSecretName(),
		},
	}

	err := serverClient.Get(ctx, k8sclient.ObjectKey{Name: oauthClientSecrets.Name, Namespace: r.ConfigManager.GetOperatorNamespace()}, oauthClientSecrets)
	if err != nil {
		return "", fmt.Errorf("Could not find %s Secret: %w", oauthClientSecrets.Name, err)
	}

	clientSecretBytes, ok := oauthClientSecrets.Data[string(r.Config.GetProductName())]
	if !ok {
		return "", fmt.Errorf("Could not find %s key in %s Secret", string(r.Config.GetProductName()), oauthClientSecrets.Name)
	}
	return string(clientSecretBytes), nil
}

func (r *Reconciler) reconcileSMTPCredentials(ctx context.Context, serverClient k8sclient.Client) (integreatlyv1alpha1.StatusPhase, error) {
	r.log.Info("Reconciling smtp credentials")

	// get the secret containing smtp credentials
	credSec := &corev1.Secret{}
	err := serverClient.Get(ctx, k8sclient.ObjectKey{Name: r.installation.Spec.SMTPSecret, Namespace: r.installation.Namespace}, credSec)
	if err != nil {
		r.log.Warningf("could not obtain smtp credentials secret", l.Fields{"error": err})
	}
	smtpConfigSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "system-smtp",
			Namespace: r.Config.GetNamespace(),
		},
		Data: map[string][]byte{},
	}

	// reconcile the smtp configmap for 3scale
	_, err = controllerutil.CreateOrUpdate(ctx, serverClient, smtpConfigSecret, func() error {
		owner.AddIntegreatlyOwnerAnnotations(smtpConfigSecret, r.installation)
		if smtpConfigSecret.Data == nil {
			smtpConfigSecret.Data = map[string][]byte{}
		}

		smtpUpdated := false

		if string(credSec.Data["host"]) != string(smtpConfigSecret.Data["address"]) {
			smtpConfigSecret.Data["address"] = credSec.Data["host"]
			smtpUpdated = true
		}
		if string(credSec.Data["authentication"]) != string(smtpConfigSecret.Data["authentication"]) {
			smtpConfigSecret.Data["authentication"] = credSec.Data["authentication"]
			smtpUpdated = true
		}
		if string(credSec.Data["domain"]) != string(smtpConfigSecret.Data["domain"]) {
			smtpConfigSecret.Data["domain"] = credSec.Data["domain"]
			smtpUpdated = true
		}
		if string(credSec.Data["openssl.verify.mode"]) != string(smtpConfigSecret.Data["openssl.verify.mode"]) {
			smtpConfigSecret.Data["openssl.verify.mode"] = credSec.Data["openssl.verify.mode"]
			smtpUpdated = true
		}
		if string(credSec.Data["password"]) != string(smtpConfigSecret.Data["password"]) {
			smtpConfigSecret.Data["password"] = credSec.Data["password"]
			smtpUpdated = true
		}
		if string(credSec.Data["port"]) != string(smtpConfigSecret.Data["port"]) {
			smtpConfigSecret.Data["port"] = credSec.Data["port"]
			smtpUpdated = true
		}
		if string(credSec.Data["username"]) != string(smtpConfigSecret.Data["username"]) {
			smtpConfigSecret.Data["username"] = credSec.Data["username"]
			smtpUpdated = true
		}

		if smtpUpdated {
			err = r.RolloutDeployment(ctx, "system-app")
			if err != nil {
				r.log.Error("Rollout system-app deployment", err)
			}

			err = r.RolloutDeployment(ctx, "system-sidekiq")
			if err != nil {
				r.log.Error("Rollout system-sidekiq deployment", err)
			}
		}

		return nil
	})
	if err != nil {
		return integreatlyv1alpha1.PhaseFailed, fmt.Errorf("failed to create or update 3scale smtp configmap: %w", err)
	}

	return integreatlyv1alpha1.PhaseCompleted, nil
}

func (r *Reconciler) reconcileComponents(ctx context.Context, serverClient k8sclient.Client, productConfig quota.ProductConfig) (integreatlyv1alpha1.StatusPhase, error) {

	fss, err := r.getBlobStorageFileStorageSpec(ctx, serverClient)
	if err != nil {
		return integreatlyv1alpha1.PhaseFailed, err
	}

	// create the 3scale api manager
	resourceRequirements := r.installation.Spec.Type != string(integreatlyv1alpha1.InstallationTypeWorkshop)
	apim := &threescalev1.APIManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      apiManagerName,
			Namespace: r.Config.GetNamespace(),
		},
		Spec: threescalev1.APIManagerSpec{
			HighAvailability:    &threescalev1.HighAvailabilitySpec{},
			PodDisruptionBudget: &threescalev1.PodDisruptionBudgetSpec{},
			Monitoring:          &threescalev1.MonitoringSpec{},
			APIManagerCommonSpec: threescalev1.APIManagerCommonSpec{
				ResourceRequirementsEnabled: &resourceRequirements,
			},
			System: &threescalev1.SystemSpec{
				DatabaseSpec: &threescalev1.SystemDatabaseSpec{
					PostgreSQL: &threescalev1.SystemPostgreSQLSpec{},
				},
				FileStorageSpec: &threescalev1.SystemFileStorageSpec{
					S3: &threescalev1.SystemS3Spec{},
				},
				AppSpec:     &threescalev1.SystemAppSpec{Replicas: &[]int64{0}[0]},
				SidekiqSpec: &threescalev1.SystemSidekiqSpec{Replicas: &[]int64{0}[0]},
			},
			Apicast: &threescalev1.ApicastSpec{
				ProductionSpec: &threescalev1.ApicastProductionSpec{Replicas: &[]int64{0}[0]},
				StagingSpec:    &threescalev1.ApicastStagingSpec{Replicas: &[]int64{0}[0]},
			},
			Backend: &threescalev1.BackendSpec{
				ListenerSpec: &threescalev1.BackendListenerSpec{Replicas: &[]int64{0}[0]},
				WorkerSpec:   &threescalev1.BackendWorkerSpec{Replicas: &[]int64{0}[0]},
				CronSpec:     &threescalev1.BackendCronSpec{Replicas: &[]int64{0}[0]},
			},
			Zync: &threescalev1.ZyncSpec{
				AppSpec: &threescalev1.ZyncAppSpec{Replicas: &[]int64{0}[0]},
				QueSpec: &threescalev1.ZyncQueSpec{Replicas: &[]int64{0}[0]},
			},
		},
	}

	antiAffinityRequired, err := resources.IsAntiAffinityRequired(ctx, serverClient)
	if err != nil {
		r.log.Warning("Failure when deciding if pod anti affinity is required. Defaulted to false: " + err.Error())
		antiAffinityRequired = false
	}

	key, err := k8sclient.ObjectKeyFromObject(apim)
	if err != nil {
		return integreatlyv1alpha1.PhaseFailed, err
	}

	err = serverClient.Get(ctx, key, apim)
	if err != nil && !k8serr.IsNotFound(err) {
		return integreatlyv1alpha1.PhaseFailed, err
	}

	status, err := controllerutil.CreateOrUpdate(ctx, serverClient, apim, func() error {

		apim.Spec.HighAvailability = &threescalev1.HighAvailabilitySpec{Enabled: true}
		apim.Spec.APIManagerCommonSpec.ResourceRequirementsEnabled = &resourceRequirements
		apim.Spec.APIManagerCommonSpec.WildcardDomain = r.installation.Spec.RoutingSubdomain
		apim.Spec.System.FileStorageSpec = fss
		apim.Spec.PodDisruptionBudget = &threescalev1.PodDisruptionBudgetSpec{Enabled: true}
		apim.Spec.Monitoring = &threescalev1.MonitoringSpec{Enabled: false}

		replicas := r.Config.GetReplicasConfig(r.installation)

		if *apim.Spec.System.AppSpec.Replicas < replicas["systemApp"] {
			*apim.Spec.System.AppSpec.Replicas = replicas["systemApp"]
		}
		if *apim.Spec.System.SidekiqSpec.Replicas < replicas["systemSidekiq"] {
			*apim.Spec.System.SidekiqSpec.Replicas = replicas["systemSidekiq"]
		}
		if *apim.Spec.Apicast.StagingSpec.Replicas < replicas["apicastStage"] {
			*apim.Spec.Apicast.StagingSpec.Replicas = replicas["apicastStage"]
		}
		if *apim.Spec.Backend.CronSpec.Replicas < replicas["backendCron"] {
			*apim.Spec.Backend.CronSpec.Replicas = replicas["backendCron"]
		}
		if *apim.Spec.Zync.AppSpec.Replicas < replicas["zyncApp"] {
			*apim.Spec.Zync.AppSpec.Replicas = replicas["zyncApp"]
		}
		if *apim.Spec.Zync.QueSpec.Replicas < replicas["zyncQue"] {
			*apim.Spec.Zync.QueSpec.Replicas = replicas["zyncQue"]
		}

		apim.Spec.System.AppSpec.Affinity = resources.SelectAntiAffinityForCluster(antiAffinityRequired, map[string]string{
			"threescale_component":         "system",
			"threescale_component_element": "app",
		})
		apim.Spec.System.SidekiqSpec.Affinity = resources.SelectAntiAffinityForCluster(antiAffinityRequired, map[string]string{
			"threescale_component":         "system",
			"threescale_component_element": "sidekiq",
		})
		apim.Spec.Apicast.ProductionSpec.Affinity = resources.SelectAntiAffinityForCluster(antiAffinityRequired, map[string]string{
			"threescale_component":         "apicast",
			"threescale_component_element": "production",
		})
		apim.Spec.Apicast.StagingSpec.Affinity = resources.SelectAntiAffinityForCluster(antiAffinityRequired, map[string]string{
			"threescale_component":         "apicast",
			"threescale_component_element": "staging",
		})

		apim.Spec.Backend.ListenerSpec.Affinity = resources.SelectAntiAffinityForCluster(antiAffinityRequired, map[string]string{
			"threescale_component":         "backend",
			"threescale_component_element": "listener",
		})
		apim.Spec.Backend.WorkerSpec.Affinity = resources.SelectAntiAffinityForCluster(antiAffinityRequired, map[string]string{
			"threescale_component":         "backend",
			"threescale_component_element": "worker",
		})
		apim.Spec.Backend.CronSpec.Affinity = resources.SelectAntiAffinityForCluster(antiAffinityRequired, map[string]string{
			"threescale_component":         "backend",
			"threescale_component_element": "cron",
		})
		apim.Spec.Zync.AppSpec.Affinity = resources.SelectAntiAffinityForCluster(antiAffinityRequired, map[string]string{
			"threescale_component":         "zync",
			"threescale_component_element": "zync",
		})
		apim.Spec.Zync.QueSpec.Affinity = resources.SelectAntiAffinityForCluster(antiAffinityRequired, map[string]string{
			"threescale_component":         "zync",
			"threescale_component_element": "zync-que",
		})

		if r.installation.Spec.Type == string(integreatlyv1alpha1.InstallationTypeManaged) {
			if *apim.Spec.Apicast.ProductionSpec.Replicas < replicas["apicastProd"] {
				*apim.Spec.Apicast.ProductionSpec.Replicas = replicas["apicastProd"]
			}
			if *apim.Spec.Backend.ListenerSpec.Replicas < replicas["backendListener"] {
				*apim.Spec.Backend.ListenerSpec.Replicas = replicas["backendListener"]
			}
			if *apim.Spec.Backend.WorkerSpec.Replicas < replicas["backendWorker"] {
				*apim.Spec.Backend.WorkerSpec.Replicas = replicas["backendWorker"]
			}
			apicastProdResources := corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceCPU: k8sresource.MustParse("300m"), corev1.ResourceMemory: k8sresource.MustParse("250Mi")},
				Limits:   corev1.ResourceList{corev1.ResourceCPU: k8sresource.MustParse("600m"), corev1.ResourceMemory: k8sresource.MustParse("300Mi")},
			}
			apim.Spec.Apicast.ProductionSpec.Resources = &apicastProdResources

			backendWorkerResources := corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceCPU: k8sresource.MustParse("150m"), corev1.ResourceMemory: k8sresource.MustParse("100Mi")},
				Limits:   corev1.ResourceList{corev1.ResourceCPU: k8sresource.MustParse("300m"), corev1.ResourceMemory: k8sresource.MustParse("100Mi")},
			}
			apim.Spec.Backend.WorkerSpec.Resources = &backendWorkerResources

			backendListenerResources := corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceCPU: k8sresource.MustParse("250m"), corev1.ResourceMemory: k8sresource.MustParse("450Mi")},
				Limits:   corev1.ResourceList{corev1.ResourceCPU: k8sresource.MustParse("600m"), corev1.ResourceMemory: k8sresource.MustParse("500Mi")},
			}
			apim.Spec.Backend.ListenerSpec.Resources = &backendListenerResources
		}

		if r.installation.Spec.Type == string(integreatlyv1alpha1.InstallationTypeManagedApi) {
			err = productConfig.Configure(apim)

			if err != nil {
				return err
			}
		}

		owner.AddIntegreatlyOwnerAnnotations(apim, r.installation)

		return nil
	})

	if err != nil {
		return integreatlyv1alpha1.PhaseFailed, err
	}

	r.log.Infof("API Manager: ", l.Fields{"status": status})

	if len(apim.Status.Deployments.Starting) == 0 && len(apim.Status.Deployments.Stopped) == 0 && len(apim.Status.Deployments.Ready) > 0 {

		threescaleRoute, err := r.getThreescaleRoute(ctx, serverClient, "system-provider", func(r routev1.Route) bool {
			return strings.HasPrefix(r.Spec.Host, "3scale-admin.")
		})
		if threescaleRoute != nil {
			r.Config.SetHost("https://" + threescaleRoute.Spec.Host)
			err = r.ConfigManager.WriteConfig(r.Config)
			if err != nil {
				return integreatlyv1alpha1.PhaseFailed, err
			}
		} else if err != nil {
			r.log.Error("Error getting system-provider route", err)
			return integreatlyv1alpha1.PhaseFailed, err
		}
		// Its not enough to just check if the system-provider route exists. This can exist but system-master, for example, may not
		exist, err := r.routesExist(ctx, serverClient)
		if err != nil {
			return integreatlyv1alpha1.PhaseFailed, err
		}
		if exist {
			return integreatlyv1alpha1.PhaseCompleted, nil
		} else {
			// If the system-provider route does not exist at this point (i.e. when Deployments are ready)
			// we can force a resync of routes. see below for more details on why this is required:
			// https://access.redhat.com/documentation/en-us/red_hat_3scale_api_management/2.7/html/operating_3scale/backup-restore#creating_equivalent_zync_routes
			// This scenario will manifest during a backup and restore and also if the product ns was accidentally deleted.
			return r.resyncRoutes(ctx, serverClient)
		}
	}
	r.log.Infof("3Scale Deployments in progress",
		l.Fields{"starting": len(apim.Status.Deployments.Starting), "stopped": len(apim.Status.Deployments.Stopped), "ready": len(apim.Status.Deployments.Ready)})

	return integreatlyv1alpha1.PhaseInProgress, nil
}

func (r *Reconciler) routesExist(ctx context.Context, serverClient k8sclient.Client) (bool, error) {
	expectedRoutes := 4
	opts := k8sclient.ListOptions{
		Namespace: r.Config.GetNamespace(),
	}

	routes := routev1.RouteList{}
	err := serverClient.List(ctx, &routes, &opts)
	if err != nil {
		return false, err
	}

	if len(routes.Items) >= expectedRoutes {
		return true, nil
	}
	r.log.Warningf("Required number of routes do not exist", l.Fields{"found": len(routes.Items), "required": expectedRoutes})
	return false, nil
}

func (r *Reconciler) resyncRoutes(ctx context.Context, client k8sclient.Client) (integreatlyv1alpha1.StatusPhase, error) {
	ns := r.Config.GetNamespace()
	podname := ""

	pods := &corev1.PodList{}
	listOpts := []k8sclient.ListOption{
		k8sclient.InNamespace(ns),
		k8sclient.MatchingLabels(map[string]string{"deploymentConfig": "system-sidekiq"}),
	}
	err := client.List(ctx, pods, listOpts...)

	for _, pod := range pods.Items {
		if pod.Status.Phase == "Running" {
			podname = pod.ObjectMeta.Name
			break
		}
	}

	if podname == "" {
		r.log.Info("Waiting on system-sidekiq pod to start, 3Scale install in progress")
		return integreatlyv1alpha1.PhaseInProgress, nil
	}

	stdout, stderr, err := resources.ExecuteRemoteCommand(ns, podname, "bundle exec rake zync:resync:domains", r.log)
	if err != nil {
		r.log.Error("Failed to resync 3Scale routes", err)
		return integreatlyv1alpha1.PhaseFailed, nil
	} else if stderr != "" {
		err := errors.New(stderr)
		r.log.Error("Error attempting to resync 3Scale routes", err)
		return integreatlyv1alpha1.PhaseFailed, err
	} else {
		r.log.Infof("Resync 3Scale routes result", l.Fields{"stdout": stdout})
		return integreatlyv1alpha1.PhaseInProgress, nil
	}
}

func (r *Reconciler) reconcileBlobStorage(ctx context.Context, serverClient k8sclient.Client) (integreatlyv1alpha1.StatusPhase, error) {
	r.log.Info("Reconciling blob storage")
	ns := r.installation.Namespace

	// setup blob storage cr for the cloud resource operator
	blobStorageName := fmt.Sprintf("%s%s", constants.ThreeScaleBlobStoragePrefix, r.installation.Name)
	blobStorage, err := croUtil.ReconcileBlobStorage(ctx, serverClient, defaultInstallationNamespace, r.installation.Spec.Type, croUtil.TierProduction, blobStorageName, ns, blobStorageName, ns, func(cr metav1.Object) error {
		owner.AddIntegreatlyOwnerAnnotations(cr, r.installation)
		return nil
	})
	if err != nil {
		return integreatlyv1alpha1.PhaseFailed, fmt.Errorf("failed to reconcile blob storage request: %w", err)
	}

	// wait for the blob storage cr to reconcile
	if blobStorage.Status.Phase != types.PhaseComplete {
		return integreatlyv1alpha1.PhaseAwaitingComponents, nil
	}

	return integreatlyv1alpha1.PhaseCompleted, nil
}

func (r *Reconciler) getBlobStorageFileStorageSpec(ctx context.Context, serverClient k8sclient.Client) (*threescalev1.SystemFileStorageSpec, error) {
	// create blob storage cr
	blobStorage := &crov1.BlobStorage{}
	err := serverClient.Get(ctx, k8sclient.ObjectKey{Name: fmt.Sprintf("%s%s", constants.ThreeScaleBlobStoragePrefix, r.installation.Name), Namespace: r.installation.Namespace}, blobStorage)
	if err != nil {
		return nil, fmt.Errorf("failed to get blob storage custom resource: %w", err)
	}

	// get blob storage connection secret
	blobStorageSec := &corev1.Secret{}
	err = serverClient.Get(ctx, k8sclient.ObjectKey{Name: blobStorage.Status.SecretRef.Name, Namespace: blobStorage.Status.SecretRef.Namespace}, blobStorageSec)
	if err != nil {
		return nil, fmt.Errorf("failed to get blob storage connection secret: %w", err)
	}

	// create s3 credentials secret
	credSec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s3CredentialsSecretName,
			Namespace: r.Config.GetNamespace(),
		},
		Data: map[string][]byte{},
	}

	_, err = controllerutil.CreateOrUpdate(ctx, serverClient, credSec, func() error {
		// Map known key names from CRO, and append any additional values that may be used for Minio
		for key, value := range blobStorageSec.Data {
			switch key {
			case "credentialKeyID":
				credSec.Data["AWS_ACCESS_KEY_ID"] = blobStorageSec.Data["credentialKeyID"]
			case "credentialSecretKey":
				credSec.Data["AWS_SECRET_ACCESS_KEY"] = blobStorageSec.Data["credentialSecretKey"]
			case "bucketName":
				credSec.Data["AWS_BUCKET"] = blobStorageSec.Data["bucketName"]
			case "bucketRegion":
				credSec.Data["AWS_REGION"] = blobStorageSec.Data["bucketRegion"]
			default:
				credSec.Data[key] = value
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create or update blob storage aws credentials secret: %w", err)
	}
	// return the file storage spec
	return &threescalev1.SystemFileStorageSpec{
		S3: &threescalev1.SystemS3Spec{
			ConfigurationSecretRef: corev1.LocalObjectReference{
				Name: s3CredentialsSecretName,
			},
		},
	}, nil
}

// reconcileExternalDatasources provisions 2 redis caches and a postgres instance
// which are used when 3scale HighAvailability mode is enabled
func (r *Reconciler) reconcileExternalDatasources(ctx context.Context, serverClient k8sclient.Client) (integreatlyv1alpha1.StatusPhase, error) {
	r.log.Info("Reconciling external datastores")
	ns := r.installation.Namespace

	// setup backend redis custom resource
	// this will be used by the cloud resources operator to provision a redis instance
	r.log.Info("Creating backend redis instance")
	backendRedisName := fmt.Sprintf("%s%s", constants.ThreeScaleBackendRedisPrefix, r.installation.Name)
	backendRedis, err := croUtil.ReconcileRedis(ctx, serverClient, defaultInstallationNamespace, r.installation.Spec.Type, croUtil.TierProduction, backendRedisName, ns, backendRedisName, ns, true, func(cr metav1.Object) error {
		owner.AddIntegreatlyOwnerAnnotations(cr, r.installation)
		return nil
	})
	if err != nil {
		return integreatlyv1alpha1.PhaseFailed, fmt.Errorf("failed to reconcile backend redis request: %w", err)
	}

	// setup system redis custom resource
	// this will be used by the cloud resources operator to provision a redis instance
	r.log.Info("Creating system redis instance")
	systemRedisName := fmt.Sprintf("%s%s", constants.ThreeScaleSystemRedisPrefix, r.installation.Name)
	systemRedis, err := croUtil.ReconcileRedis(ctx, serverClient, defaultInstallationNamespace, r.installation.Spec.Type, croUtil.TierProduction, systemRedisName, ns, systemRedisName, ns, true, func(cr metav1.Object) error {
		owner.AddIntegreatlyOwnerAnnotations(cr, r.installation)
		return nil
	})
	if err != nil {
		return integreatlyv1alpha1.PhaseFailed, fmt.Errorf("failed to reconcile system redis request: %w", err)
	}

	// setup postgres cr for the cloud resource operator
	// this will be used by the cloud resources operator to provision a postgres instance
	r.log.Info("Creating postgres instance")
	postgresName := fmt.Sprintf("%s%s", constants.ThreeScalePostgresPrefix, r.installation.Name)
	postgres, err := croUtil.ReconcilePostgres(ctx, serverClient, defaultInstallationNamespace, r.installation.Spec.Type, croUtil.TierProduction, postgresName, ns, postgresName, ns, constants.PostgresApplyImmediately, func(cr metav1.Object) error {
		owner.AddIntegreatlyOwnerAnnotations(cr, r.installation)
		return nil
	})
	if err != nil {
		return integreatlyv1alpha1.PhaseFailed, fmt.Errorf("failed to reconcile postgres request: %w", err)
	}

	phase, err := resources.ReconcileRedisAlerts(ctx, serverClient, r.installation, backendRedis, r.log)
	if err != nil {
		return integreatlyv1alpha1.PhaseFailed, fmt.Errorf("failed to reconcile redis alerts: %w", err)
	}
	if phase != integreatlyv1alpha1.PhaseCompleted {
		return phase, nil
	}

	// create Redis Cpu Usage High alert
	err = resources.CreateRedisCpuUsageAlerts(ctx, serverClient, r.installation, backendRedis, r.log)
	if err != nil {
		return integreatlyv1alpha1.PhaseFailed, fmt.Errorf("failed to create backend redis prometheus Cpu usage high alerts for threescale: %s", err)
	}
	// wait for the backend redis cr to reconcile
	if backendRedis.Status.Phase != types.PhaseComplete {
		return integreatlyv1alpha1.PhaseAwaitingComponents, nil
	}

	// get the secret created by the cloud resources operator
	// containing backend redis connection details
	credSec := &corev1.Secret{}
	err = serverClient.Get(ctx, k8sclient.ObjectKey{Name: backendRedis.Status.SecretRef.Name, Namespace: backendRedis.Status.SecretRef.Namespace}, credSec)
	if err != nil {
		return integreatlyv1alpha1.PhaseFailed, fmt.Errorf("failed to get backend redis credential secret: %w", err)
	}

	// create backend redis external connection secret needed for the 3scale apimanager
	backendRedisSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      externalBackendRedisSecretName,
			Namespace: r.Config.GetNamespace(),
		},
		Data: map[string][]byte{},
	}
	_, err = controllerutil.CreateOrUpdate(ctx, serverClient, backendRedisSecret, func() error {
		uri := credSec.Data["uri"]
		port := credSec.Data["port"]
		backendRedisSecret.Data["REDIS_STORAGE_URL"] = []byte(fmt.Sprintf("redis://%s:%s/0", uri, port))
		backendRedisSecret.Data["REDIS_QUEUES_URL"] = []byte(fmt.Sprintf("redis://%s:%s/1", uri, port))
		return nil
	})
	if err != nil {
		return integreatlyv1alpha1.PhaseFailed, fmt.Errorf("failed to create or update 3scale %s connection secret: %w", externalBackendRedisSecretName, err)
	}

	phase, err = resources.ReconcileRedisAlerts(ctx, serverClient, r.installation, systemRedis, r.log)
	if err != nil {
		return integreatlyv1alpha1.PhaseFailed, fmt.Errorf("failed to reconcile redis alerts: %w", err)
	}
	if phase != integreatlyv1alpha1.PhaseCompleted {
		return phase, nil
	}
	// wait for the system redis cr to reconcile
	if systemRedis.Status.Phase != types.PhaseComplete {
		return integreatlyv1alpha1.PhaseAwaitingComponents, nil
	}

	// get the secret created by the cloud resources operator
	// containing system redis connection details
	systemCredSec := &corev1.Secret{}
	err = serverClient.Get(ctx, k8sclient.ObjectKey{Name: systemRedis.Status.SecretRef.Name, Namespace: systemRedis.Status.SecretRef.Namespace}, systemCredSec)
	if err != nil {
		return integreatlyv1alpha1.PhaseFailed, fmt.Errorf("failed to get system redis credential secret: %w", err)
	}

	// create system redis external connection secret needed for the 3scale apimanager
	redisSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      externalRedisSecretName,
			Namespace: r.Config.GetNamespace(),
		},
		Data: map[string][]byte{},
	}
	_, err = controllerutil.CreateOrUpdate(ctx, serverClient, redisSecret, func() error {
		uri := systemCredSec.Data["uri"]
		port := systemCredSec.Data["port"]
		conn := fmt.Sprintf("redis://%s:%s/1", uri, port)
		redisSecret.Data["URL"] = []byte(conn)
		redisSecret.Data["MESSAGE_BUS_URL"] = []byte(conn)
		return nil
	})
	if err != nil {
		return integreatlyv1alpha1.PhaseFailed, fmt.Errorf("failed to create or update 3scale %s connection secret: %w", externalRedisSecretName, err)
	}

	// reconcile postgres alerts
	phase, err = resources.ReconcilePostgresAlerts(ctx, serverClient, r.installation, postgres, r.log)
	productName := postgres.Labels["productName"]
	if err != nil {
		return integreatlyv1alpha1.PhaseFailed, fmt.Errorf("failed to reconcile postgres alerts for %s: %w", productName, err)
	}
	if phase != integreatlyv1alpha1.PhaseCompleted {
		return phase, nil
	}

	// get the secret containing redis credentials
	postgresCredSec := &corev1.Secret{}
	err = serverClient.Get(ctx, k8sclient.ObjectKey{Name: postgres.Status.SecretRef.Name, Namespace: postgres.Status.SecretRef.Namespace}, postgresCredSec)
	if err != nil {
		return integreatlyv1alpha1.PhaseFailed, fmt.Errorf("failed to get postgres credential secret: %w", err)
	}

	// create postgres external connection secret
	postgresSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      externalPostgresSecretName,
			Namespace: r.Config.GetNamespace(),
		},
		Data: map[string][]byte{},
	}
	_, err = controllerutil.CreateOrUpdate(ctx, serverClient, postgresSecret, func() error {
		username := postgresCredSec.Data["username"]
		password := postgresCredSec.Data["password"]
		url := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", username, password, postgresCredSec.Data["host"], postgresCredSec.Data["port"], postgresCredSec.Data["database"])

		postgresSecret.Data["URL"] = []byte(url)
		postgresSecret.Data["DB_USER"] = username
		postgresSecret.Data["DB_PASSWORD"] = password
		return nil
	})
	if err != nil {
		return integreatlyv1alpha1.PhaseFailed, fmt.Errorf("failed to create or update 3scale %s connection secret: %w", externalPostgresSecretName, err)
	}

	return integreatlyv1alpha1.PhaseCompleted, nil
}

func (r *Reconciler) reconcileOutgoingEmailAddress(ctx context.Context, serverClient k8sclient.Client) (integreatlyv1alpha1.StatusPhase, error) {
	monitoringConfig, err := r.ConfigManager.ReadMonitoring()
	if err != nil {
		return integreatlyv1alpha1.PhaseFailed, err
	}

	existingSMTPFromAddress, err := resources.GetExistingSMTPFromAddress(ctx, serverClient, monitoringConfig.GetOperatorNamespace())

	if err != nil {
		if !k8serr.IsNotFound(err) {
			r.log.Error("Error getting smtp_from address from secret alertmanager-application-monitoring", err)
			return integreatlyv1alpha1.PhaseFailed, nil
		}
	}
	if existingSMTPFromAddress == "" {
		r.log.Warning("Couldn't find SMTP in a secret, retrieving it from the envar")
		existingSMTPFromAddress = os.Getenv(integreatlyv1alpha1.EnvKeyAlertSMTPFrom)
	}
	accessToken, err := r.GetAdminToken(ctx, serverClient)
	if err != nil {
		r.log.Info("Failed to get admin token in reconcileOutgoingEmailAddresss: " + err.Error())
		return integreatlyv1alpha1.PhaseInProgress, err
	}
	_, err = r.tsClient.SetFromEmailAddress(existingSMTPFromAddress, *accessToken)
	if err != nil {
		r.log.Error("Failed to set email from address:", err)
		return integreatlyv1alpha1.PhaseFailed, err
	}
	return integreatlyv1alpha1.PhaseCompleted, nil

}

func (r *Reconciler) reconcileRHSSOIntegration(ctx context.Context, serverClient k8sclient.Client) (integreatlyv1alpha1.StatusPhase, error) {
	rhssoConfig, err := r.ConfigManager.ReadRHSSO()
	if err != nil {
		return integreatlyv1alpha1.PhaseFailed, err
	}
	rhssoNamespace := rhssoConfig.GetNamespace()
	rhssoRealm := rhssoConfig.GetRealm()
	if rhssoNamespace == "" || rhssoRealm == "" {
		r.log.Warningf("Cannot configure SSO integration without SSO", l.Fields{"ns": rhssoNamespace, "realm": rhssoRealm})
		return integreatlyv1alpha1.PhaseInProgress, nil
	}

	kcClient := &keycloak.KeycloakClient{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clientID,
			Namespace: rhssoNamespace,
		},
	}

	// keycloak-operator sets the spec.client.id, we need to preserve that value
	apiClientID := ""
	err = serverClient.Get(ctx, k8sclient.ObjectKey{
		Namespace: rhssoNamespace,
		Name:      clientID,
	}, kcClient)
	if err == nil {
		apiClientID = kcClient.Spec.Client.ID
	}

	clientSecret, err := r.getOauthClientSecret(ctx, serverClient)
	if err != nil {
		r.log.Error("Error retrieving client secret", err)
		return integreatlyv1alpha1.PhaseFailed, err
	}

	opRes, err := controllerutil.CreateOrUpdate(ctx, serverClient, kcClient, func() error {
		kcClient.Spec = r.getKeycloakClientSpec(apiClientID, clientSecret)
		return nil
	})
	if err != nil {
		return integreatlyv1alpha1.PhaseFailed, fmt.Errorf("could not create/update 3scale keycloak client: %w operation: %v", err, opRes)
	}

	accessToken, err := r.GetAdminToken(ctx, serverClient)
	if err != nil {
		r.log.Info("Failed to get admin token: " + err.Error())
		return integreatlyv1alpha1.PhaseInProgress, err
	}
	_, err = r.tsClient.GetAuthenticationProviderByName(rhssoIntegrationName, *accessToken)
	if err != nil && !tsIsNotFoundError(err) {
		r.log.Info("Failed to get authentication provider:" + err.Error())
		return integreatlyv1alpha1.PhaseInProgress, err
	}
	if tsIsNotFoundError(err) {
		site := rhssoConfig.GetHost() + "/auth/realms/" + rhssoRealm
		res, err := r.tsClient.AddAuthenticationProvider(map[string]string{
			"kind":                              "keycloak",
			"name":                              rhssoIntegrationName,
			"client_id":                         clientID,
			"client_secret":                     clientSecret,
			"site":                              site,
			"skip_ssl_certificate_verification": "true",
			"published":                         "true",
		}, *accessToken)
		if err != nil || res.StatusCode != http.StatusCreated {
			if err != nil {
				r.log.Info("Failed to add authentication provider:" + err.Error())
			}
			return integreatlyv1alpha1.PhaseInProgress, err
		}
	}

	return integreatlyv1alpha1.PhaseCompleted, nil
}

func (r *Reconciler) getOAuthClientName() string {
	return r.installation.Spec.NamespacePrefix + string(r.Config.GetProductName())
}

func (r *Reconciler) reconcileOpenshiftUsers(ctx context.Context, installation *integreatlyv1alpha1.RHMI, serverClient k8sclient.Client) (integreatlyv1alpha1.StatusPhase, error) {
	r.log.Info("Reconciling openshift users to 3scale")

	rhssoConfig, err := r.ConfigManager.ReadRHSSO()
	if err != nil {
		return integreatlyv1alpha1.PhaseFailed, err
	}

	accessToken, err := r.GetAdminToken(ctx, serverClient)
	if err != nil {
		return integreatlyv1alpha1.PhaseFailed, err
	}

	systemAdminUsername, _, err := r.GetAdminNameAndPassFromSecret(ctx, serverClient)
	if err != nil {
		r.log.Info("Failed to retrieve admin name and password from secret: " + err.Error())
		return integreatlyv1alpha1.PhaseInProgress, err
	}

	kcu, err := rhsso.GetKeycloakUsers(ctx, serverClient, rhssoConfig.GetNamespace())
	if err != nil {
		return integreatlyv1alpha1.PhaseFailed, err
	}

	tsUsers, err := r.tsClient.GetUsers(*accessToken)
	if err != nil {
		r.log.Info("Failed to get users:" + err.Error())
		return integreatlyv1alpha1.PhaseInProgress, err
	}

	added, deleted, updated := r.getUserDiff(ctx, serverClient, kcu, tsUsers.Users)
	// reset the user action metric before we re-reconcile
	// in order to get up to date metrics on user creation
	metrics.ResetThreeScaleUserAction()
	// the deleted entries are addressed first
	// a common use case is where one idp is added to give early access to the cluster
	// later that idp is removed and a more permanent one is added
	// if there are any duplicate emails across the set of users the user from the first idp
	// should be removed first and that allows for the new one which had a potential conflict
	// can now be added.
	for _, tsUser := range deleted {
		if tsUser.UserDetails.Username != *systemAdminUsername {
			statusCode := http.StatusServiceUnavailable

			res, err := r.tsClient.DeleteUser(tsUser.UserDetails.Id, *accessToken)
			if err != nil {
				r.log.Error(fmt.Sprintf("Failed to delete keycloak user %d from 3scale", tsUser.UserDetails.Id), err)
			} else {
				statusCode = res.StatusCode
			}

			metrics.SetThreeScaleUserAction(statusCode, strconv.Itoa(tsUser.UserDetails.Id), http.MethodDelete)

			if statusCode != http.StatusOK {
				r.log.Error(fmt.Sprintf("Failed to delete keycloak user %d from 3scale with status code %d", tsUser.UserDetails.Id, statusCode), errors.New("error on http request"))
			}
		}
	}

	for _, tsUser := range updated {
		if tsUser.UserDetails.Username != *systemAdminUsername {
			genKcUser, err := getGeneratedKeycloakUser(ctx, serverClient, rhssoConfig.GetNamespace(), tsUser)

			if err != nil {
				r.log.Warning("Failed to get generate keycloak user: " + err.Error())
				continue
			}

			_, err = r.tsClient.UpdateUser(tsUser.UserDetails.Id, strings.ToLower(genKcUser.Spec.User.UserName), tsUser.UserDetails.Email, *accessToken)
			if err != nil {
				r.log.Warning("Failed to updating 3scale user details: " + err.Error())
			}
		}
	}

	for _, kcUser := range added {
		user, _ := r.tsClient.GetUser(strings.ToLower(kcUser.UserName), *accessToken)
		// recheck the user is new.
		// 3scale user may being update during the update phase
		if user == nil {
			statusCode := http.StatusServiceUnavailable
			res, err := r.tsClient.AddUser(strings.ToLower(kcUser.UserName), strings.ToLower(kcUser.Email), "", *accessToken)

			if err != nil {
				r.log.Error(fmt.Sprintf("Failed to add keycloak user %s to 3scale", kcUser.UserName), err)
			} else {
				statusCode = res.StatusCode
			}

			// when the failure of user happens we don't want to block the reconciler.
			// failure to create a user can happen in the case of the username being too long
			// the max allowed user length is 40 characters in 3scale.
			// The reconciler will continue to allow the installation to happen and a metric
			// will be exposed and alert fire to alert to the creation failure
			metrics.SetThreeScaleUserAction(statusCode, kcUser.UserName, http.MethodPost)

			if statusCode != http.StatusCreated {
				r.log.Error(fmt.Sprintf("Failed to add keycloak user %s to 3scale with status code %d", kcUser.UserName, statusCode), errors.New("error on http request"))
			}
		}
	}

	// update KeycloakUser attribute after user is created in 3scale
	phase, err := r.updateKeycloakUsersAttributeWith3ScaleUserId(ctx, serverClient, kcu, accessToken)
	if err != nil || phase != integreatlyv1alpha1.PhaseCompleted {
		return phase, err
	}

	openshiftAdminGroup := &usersv1.Group{}
	err = serverClient.Get(ctx, k8sclient.ObjectKey{Name: "dedicated-admins"}, openshiftAdminGroup)
	if err != nil && !k8serr.IsNotFound(err) {
		r.log.Info("Failed to retrieve dedicated admins: " + err.Error())
		return integreatlyv1alpha1.PhaseInProgress, err
	}
	newTsUsers, err := r.tsClient.GetUsers(*accessToken)
	if err != nil {
		r.log.Info("Failed to get users: " + err.Error())
		return integreatlyv1alpha1.PhaseInProgress, err
	}

	isWorkshop := installation.Spec.Type == string(integreatlyv1alpha1.InstallationTypeWorkshop)

	err = syncOpenshiftAdminMembership(openshiftAdminGroup, newTsUsers, *systemAdminUsername, isWorkshop, r.tsClient, *accessToken)
	if err != nil {
		r.log.Info("Failed to sync openshift admin membership: " + err.Error())
		return integreatlyv1alpha1.PhaseInProgress, err
	}

	return integreatlyv1alpha1.PhaseCompleted, nil
}

func (r *Reconciler) updateKeycloakUsersAttributeWith3ScaleUserId(ctx context.Context, serverClient k8sclient.Client, kcu []keycloak.KeycloakAPIUser, accessToken *string) (integreatlyv1alpha1.StatusPhase, error) {
	rhssoConfig, err := r.ConfigManager.ReadRHSSO()
	if err != nil {
		return integreatlyv1alpha1.PhaseFailed, err
	}

	userCreated3ScaleName := "3scale_user_created"
	for _, user := range kcu {
		tsUser, err := r.tsClient.GetUser(strings.ToLower(user.UserName), *accessToken)
		if err != nil {
			return integreatlyv1alpha1.PhaseInProgress,
				fmt.Errorf("failed to get 3scale user with keycloak username %s, err: %s", user.UserName, err)
		}

		if user.Attributes == nil {
			user.Attributes = map[string][]string{
				userCreated3ScaleName: {"true"},
			}
		}

		kcUser := &keycloak.KeycloakUser{
			ObjectMeta: metav1.ObjectMeta{
				Name:      userHelper.GetValidGeneratedUserName(user),
				Namespace: rhssoConfig.GetNamespace(),
			},
		}

		_, err = controllerutil.CreateOrUpdate(ctx, serverClient, kcUser, func() error {
			user.Attributes[userCreated3ScaleName] = []string{"true"}
			user.Attributes[user3ScaleID] = []string{fmt.Sprint(tsUser.UserDetails.Id)}
			kcUser.Spec.User = user
			return nil
		})
		if err != nil {
			return integreatlyv1alpha1.PhaseInProgress,
				fmt.Errorf("failed to update KeycloakUser CR with %s attribute: %w", userCreated3ScaleName, err)
		}
	}

	return integreatlyv1alpha1.PhaseCompleted, nil
}

func (r *Reconciler) preUpgradeBackupExecutor() backup.BackupExecutor {
	if r.installation.Spec.UseClusterStorage != "false" {
		return backup.NewNoopBackupExecutor()
	}

	return backup.NewConcurrentBackupExecutor(
		backup.NewAWSBackupExecutor(
			r.installation.Namespace,
			"threescale-postgres-rhmi",
			backup.PostgresSnapshotType,
		),
		backup.NewAWSBackupExecutor(
			r.installation.Namespace,
			"threescale-backend-redis-rhmi",
			backup.RedisSnapshotType,
		),
		backup.NewAWSBackupExecutor(
			r.installation.Namespace,
			"threescale-redis-rhmi",
			backup.RedisSnapshotType,
		),
	)
}

func syncOpenshiftAdminMembership(openshiftAdminGroup *usersv1.Group, newTsUsers *Users, systemAdminUsername string, isWorkshop bool, tsClient ThreeScaleInterface, accessToken string) error {
	for _, tsUser := range newTsUsers.Users {
		// skip if ts user is the system user admin
		if tsUser.UserDetails.Username == systemAdminUsername {
			continue
		}

		// In workshop mode, developer users also get admin permissions in 3scale
		if (userIsOpenshiftAdmin(tsUser, openshiftAdminGroup) || isWorkshop) && tsUser.UserDetails.Role != adminRole {
			res, err := tsClient.SetUserAsAdmin(tsUser.UserDetails.Id, accessToken)
			if err != nil || res.StatusCode != http.StatusOK {
				return err
			}
		}
	}

	return nil
}

func (r *Reconciler) reconcileServiceDiscovery(ctx context.Context, serverClient k8sclient.Client) (integreatlyv1alpha1.StatusPhase, error) {

	if string(r.Config.GetProductVersion()) != string(integreatlyv1alpha1.Version3Scale) {
		r.Config.SetProductVersion(string(integreatlyv1alpha1.Version3Scale))
		r.ConfigManager.WriteConfig(r.Config)
	}

	if string(r.Config.GetOperatorVersion()) != string(integreatlyv1alpha1.OperatorVersion3Scale) {
		r.Config.SetOperatorVersion(string(integreatlyv1alpha1.OperatorVersion3Scale))
		r.ConfigManager.WriteConfig(r.Config)
	}

	system := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "system",
			Namespace: r.Config.GetNamespace(),
		},
	}

	status, err := controllerutil.CreateOrUpdate(ctx, serverClient, system, func() error {
		clientSecret, err := r.getOauthClientSecret(ctx, serverClient)
		if err != nil {
			return err
		}
		sdConfig := fmt.Sprintf("production:\n  enabled: true\n  authentication_method: oauth\n  oauth_server_type: builtin\n  client_id: '%s'\n  client_secret: '%s'\n", r.getOAuthClientName(), clientSecret)

		system.Data["service_discovery.yml"] = sdConfig
		return nil
	})

	if err != nil {
		r.log.Info("Failed to get oauth client secret:" + err.Error())
		return integreatlyv1alpha1.PhaseInProgress, err
	}

	if status != controllerutil.OperationResultNone {
		err = r.RolloutDeployment(ctx, "system-app")
		if err != nil {
			r.log.Info("Failed to rollout deployment (system-app):" + err.Error())
			return integreatlyv1alpha1.PhaseInProgress, err
		}

		err = r.RolloutDeployment(ctx, "system-sidekiq")
		if err != nil {
			r.log.Info("Failed to rollout deployment (system-sidekiq)" + err.Error())
			return integreatlyv1alpha1.PhaseInProgress, err
		}
	}

	return integreatlyv1alpha1.PhaseCompleted, nil
}

func (r *Reconciler) reconcileBlackboxTargets(ctx context.Context, installation *integreatlyv1alpha1.RHMI, client k8sclient.Client) (integreatlyv1alpha1.StatusPhase, error) {
	cfg, err := r.ConfigManager.ReadMonitoring()
	if err != nil {
		return integreatlyv1alpha1.PhaseInProgress, nil
	}

	err = monitoring.CreateBlackboxTarget(ctx, "integreatly-3scale-admin-ui", monitoringv1alpha1.BlackboxtargetData{
		Url:     r.Config.GetHost() + "/" + r.Config.GetBlackboxTargetPathForAdminUI(),
		Service: "3scale-admin-ui",
	}, cfg, installation, client)
	if err != nil {
		r.log.Error("Error creating threescale blackbox target", err)
		return integreatlyv1alpha1.PhaseFailed, fmt.Errorf("error creating threescale blackbox target: %w", err)
	}

	// Create a blackbox target for the developer console ui
	threescaleRoute, err := r.getThreescaleRoute(ctx, client, "system-developer", func(r routev1.Route) bool {
		return strings.HasPrefix(r.Spec.Host, "3scale.")
	})
	if err != nil {
		r.log.Info("Failed to retrieve threescale threescaleRoute: " + err.Error())
		return integreatlyv1alpha1.PhaseInProgress, nil
	}
	err = monitoring.CreateBlackboxTarget(ctx, "integreatly-3scale-system-developer", monitoringv1alpha1.BlackboxtargetData{
		Url:     "https://" + threescaleRoute.Spec.Host,
		Service: "3scale-developer-console-ui",
	}, cfg, installation, client)
	if err != nil {
		r.log.Error("Error creating blackbox target (system-developer)", err)
		return integreatlyv1alpha1.PhaseFailed, fmt.Errorf("error creating threescale blackbox target (system-developer): %w", err)
	}

	// Create a blackbox target for the master console ui
	threescaleRoute, err = r.getThreescaleRoute(ctx, client, "system-master", nil)
	if err != nil {
		return integreatlyv1alpha1.PhaseInProgress, nil
	}
	err = monitoring.CreateBlackboxTarget(ctx, "integreatly-3scale-system-master", monitoringv1alpha1.BlackboxtargetData{
		Url:     "https://" + threescaleRoute.Spec.Host,
		Service: "3scale-system-admin-ui",
	}, cfg, installation, client)
	if err != nil {
		r.log.Error("Error creating blackbox target (system-master)", err)
		return integreatlyv1alpha1.PhaseFailed, fmt.Errorf("error creating threescale blackbox target (system-master): %w", err)
	}

	return integreatlyv1alpha1.PhaseCompleted, nil
}

func (r *Reconciler) getThreescaleRoute(ctx context.Context, serverClient k8sclient.Client, label string, filterFn func(r routev1.Route) bool) (*routev1.Route, error) {
	// Add backwards compatible filter function, first element will do
	if filterFn == nil {
		filterFn = func(r routev1.Route) bool { return true }
	}

	selector, err := labels.Parse(fmt.Sprintf("zync.3scale.net/route-to=%v", label))
	if err != nil {
		return nil, err
	}

	opts := k8sclient.ListOptions{
		LabelSelector: selector,
		Namespace:     r.Config.GetNamespace(),
	}

	routes := routev1.RouteList{}
	err = serverClient.List(ctx, &routes, &opts)
	if err != nil {
		return nil, err
	}

	if len(routes.Items) == 0 {
		return nil, nil
	}

	var foundRoute *routev1.Route
	for _, rt := range routes.Items {
		if filterFn(rt) {
			foundRoute = &rt
			break
		}
	}
	return foundRoute, nil
}

func (r *Reconciler) GetAdminNameAndPassFromSecret(ctx context.Context, serverClient k8sclient.Client) (*string, *string, error) {
	s := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: r.Config.GetNamespace(),
			Name:      "system-seed",
		},
	}
	err := serverClient.Get(ctx, k8sclient.ObjectKey{Name: s.Name, Namespace: r.Config.GetNamespace()}, s)
	if err != nil {
		return nil, nil, err
	}

	username := string(s.Data["ADMIN_USER"])
	email := string(s.Data["ADMIN_EMAIL"])
	return &username, &email, nil
}

func (r *Reconciler) SetAdminDetailsOnSecret(ctx context.Context, serverClient k8sclient.Client, username string, email string) error {
	s := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: r.Config.GetNamespace(),
			Name:      "system-seed",
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, serverClient, s, func() error {
		s.Data["ADMIN_USER"] = []byte(username)
		s.Data["ADMIN_EMAIL"] = []byte(email)
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

func (r *Reconciler) GetAdminToken(ctx context.Context, serverClient k8sclient.Client) (*string, error) {
	s := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-seed",
		},
	}
	err := serverClient.Get(ctx, k8sclient.ObjectKey{Name: s.Name, Namespace: r.Config.GetNamespace()}, s)
	if err != nil {
		return nil, err
	}

	accessToken := string(s.Data["ADMIN_ACCESS_TOKEN"])
	return &accessToken, nil
}

func (r *Reconciler) RolloutDeployment(ctx context.Context, name string) error {
	_, err := r.appsv1Client.DeploymentConfigs(r.Config.GetNamespace()).Instantiate(ctx, name, &appsv1.DeploymentRequest{
		Name:   name,
		Force:  true,
		Latest: true,
	}, metav1.CreateOptions{})

	return err
}

func (r *Reconciler) getUserDiff(ctx context.Context, serverClient k8sclient.Client, kcUsers []keycloak.KeycloakAPIUser, tsUsers []*User) ([]keycloak.KeycloakAPIUser, []*User, []*User) {
	var added []keycloak.KeycloakAPIUser
	var deleted []*User
	var updated []*User

	rhssoConfig, err := r.ConfigManager.ReadRHSSO()
	if err != nil {
		r.log.Warning("Failed to get rhsso config: " + err.Error())
		return added, deleted, updated
	}

	for _, kcUser := range kcUsers {
		if !tsContainsKc(tsUsers, kcUser) {
			added = append(added, kcUser)
		}
	}

	var expectedDeleted []*User
	for _, tsUser := range tsUsers {
		if !kcContainsTs(kcUsers, tsUser) {
			expectedDeleted = append(expectedDeleted, tsUser)
		}
	}

	// compare the id fields in the generated user to that of the expected deleted user
	for _, user := range expectedDeleted {
		toDelete := true
		for _, kuUser := range kcUsers {
			genKcUser := &keycloak.KeycloakUser{
				ObjectMeta: metav1.ObjectMeta{
					Name:      userHelper.GetValidGeneratedUserName(kuUser),
					Namespace: rhssoConfig.GetNamespace(),
				},
			}
			objectKey, err := k8sclient.ObjectKeyFromObject(genKcUser)
			if err != nil {
				r.log.Warning("Failed to get object key from object: " + err.Error())
				continue
			}

			err = serverClient.Get(ctx, objectKey, genKcUser)
			if err != nil {
				r.log.Warning("Failed get generated Keycloak User: " + err.Error())
				continue
			}

			if tsUserIDInKc(user, genKcUser) {
				updated = append(updated, user)
				toDelete = false
				break
			}
		}
		if toDelete {
			deleted = append(deleted, user)
		}
	}

	return added, deleted, updated
}

// getGeneratedKeycloakUser returns a keycloakUser CR for a matching 3scale user ID
func getGeneratedKeycloakUser(ctx context.Context, serverClient k8sclient.Client, ns string, tsUser *User) (*keycloak.KeycloakUser, error) {

	var users keycloak.KeycloakUserList

	listOptions := []k8sclient.ListOption{
		k8sclient.MatchingLabels(rhsso.GetInstanceLabels()),
		k8sclient.InNamespace(ns),
	}
	err := serverClient.List(ctx, &users, listOptions...)
	if err != nil {
		return nil, err
	}

	for _, user := range users.Items {
		if tsUserIDInKc(tsUser, &user) {
			return &user, nil
		}
	}

	return nil, fmt.Errorf("Genrated Keycloak user was not found")
}

// tsUserIDInKc checks if a 3scale user ID is listed in the keycloak user attributes
func tsUserIDInKc(tsUser *User, kcUser *keycloak.KeycloakUser) bool {
	if len(kcUser.Spec.User.Attributes[user3ScaleID]) == 0 {
		return false
	}

	if strings.EqualFold(fmt.Sprint(tsUser.UserDetails.Id), kcUser.Spec.User.Attributes[user3ScaleID][0]) {
		return true
	}
	return false
}

func kcContainsTs(kcUsers []keycloak.KeycloakAPIUser, tsUser *User) bool {
	for _, kcu := range kcUsers {
		if strings.ToLower(kcu.UserName) == tsUser.UserDetails.Username {
			return true
		}
	}

	return false
}

func tsContainsKc(tsusers []*User, kcUser keycloak.KeycloakAPIUser) bool {
	for _, tsu := range tsusers {
		if tsu.UserDetails.Username == strings.ToLower(kcUser.UserName) {
			return true
		}
	}

	return false
}

func userIsOpenshiftAdmin(tsUser *User, adminGroup *usersv1.Group) bool {
	for _, userName := range adminGroup.Users {
		if strings.EqualFold(tsUser.UserDetails.Username, userName) {
			return true
		}
	}

	return false
}

func (r *Reconciler) getKeycloakClientSpec(id, clientSecret string) keycloak.KeycloakClientSpec {
	fullScopeAllowed := true

	return keycloak.KeycloakClientSpec{
		RealmSelector: &metav1.LabelSelector{
			MatchLabels: rhsso.GetInstanceLabels(),
		},
		Client: &keycloak.KeycloakAPIClient{
			ID:                      id,
			ClientID:                clientID,
			Enabled:                 true,
			Secret:                  clientSecret,
			ClientAuthenticatorType: "client-secret",
			RedirectUris: []string{
				fmt.Sprintf("https://3scale-admin.%s/*", r.installation.Spec.RoutingSubdomain),
			},
			StandardFlowEnabled: true,
			RootURL:             fmt.Sprintf("https://3scale-admin.%s", r.installation.Spec.RoutingSubdomain),
			FullScopeAllowed:    &fullScopeAllowed,
			Access: map[string]bool{
				"view":      true,
				"configure": true,
				"manage":    true,
			},
			ProtocolMappers: []keycloak.KeycloakProtocolMapper{
				{
					Name:            "given name",
					Protocol:        "openid-connect",
					ProtocolMapper:  "oidc-usermodel-property-mapper",
					ConsentRequired: true,
					ConsentText:     "${givenName}",
					Config: map[string]string{
						"userinfo.token.claim": "true",
						"user.attribute":       "firstName",
						"id.token.claim":       "true",
						"access.token.claim":   "true",
						"claim.name":           "given_name",
						"jsonType.label":       "String",
					},
				},
				{
					Name:            "email verified",
					Protocol:        "openid-connect",
					ProtocolMapper:  "oidc-usermodel-property-mapper",
					ConsentRequired: true,
					ConsentText:     "${emailVerified}",
					Config: map[string]string{
						"userinfo.token.claim": "true",
						"user.attribute":       "emailVerified",
						"id.token.claim":       "true",
						"access.token.claim":   "true",
						"claim.name":           "email_verified",
						"jsonType.label":       "String",
					},
				},
				{
					Name:            "full name",
					Protocol:        "openid-connect",
					ProtocolMapper:  "oidc-full-name-mapper",
					ConsentRequired: true,
					ConsentText:     "${fullName}",
					Config: map[string]string{
						"id.token.claim":     "true",
						"access.token.claim": "true",
					},
				},
				{
					Name:            "family name",
					Protocol:        "openid-connect",
					ProtocolMapper:  "oidc-usermodel-property-mapper",
					ConsentRequired: true,
					ConsentText:     "${familyName}",
					Config: map[string]string{
						"userinfo.token.claim": "true",
						"user.attribute":       "lastName",
						"id.token.claim":       "true",
						"access.token.claim":   "true",
						"claim.name":           "family_name",
						"jsonType.label":       "String",
					},
				},
				{
					Name:            "role list",
					Protocol:        "saml",
					ProtocolMapper:  "saml-role-list-mapper",
					ConsentRequired: false,
					ConsentText:     "${familyName}",
					Config: map[string]string{
						"single":               "false",
						"attribute.nameformat": "Basic",
						"attribute.name":       "Role",
					},
				},
				{
					Name:            "email",
					Protocol:        "openid-connect",
					ProtocolMapper:  "oidc-usermodel-property-mapper",
					ConsentRequired: true,
					ConsentText:     "${email}",
					Config: map[string]string{
						"userinfo.token.claim": "true",
						"user.attribute":       "email",
						"id.token.claim":       "true",
						"access.token.claim":   "true",
						"claim.name":           "email",
						"jsonType.label":       "String",
					},
				},
				{
					Name:            "org_name",
					Protocol:        "openid-connect",
					ProtocolMapper:  "oidc-usermodel-property-mapper",
					ConsentRequired: false,
					ConsentText:     "n.a.",
					Config: map[string]string{
						"userinfo.token.claim": "true",
						"user.attribute":       "org_name",
						"id.token.claim":       "true",
						"access.token.claim":   "true",
						"claim.name":           "org_name",
						"jsonType.label":       "String",
					},
				},
			},
		},
	}
}

func (r *Reconciler) reconcileRouteEditRole(ctx context.Context, client k8sclient.Client) (integreatlyv1alpha1.StatusPhase, error) {

	// Allow dedicated-admin group to edit routes. This is enabled to allow the public API in 3Scale, on private clusters, to be exposed.
	// This is achieved by labelling the route to match the additional router created by SRE for private clusters. INTLY-7398.

	r.log.Info("reconciling edit routes role to the dedicated admins group")

	editRoutesRole := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "edit-3scale-routes",
			Namespace: r.Config.GetNamespace(),
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, client, editRoutesRole, func() error {
		owner.AddIntegreatlyOwnerAnnotations(editRoutesRole, r.installation)

		editRoutesRole.Rules = []rbacv1.PolicyRule{
			{
				APIGroups: []string{"route.openshift.io"},
				Resources: []string{"routes"},
				Verbs:     []string{"get", "update", "list", "watch", "patch"},
			},
		}

		return nil
	})
	if err != nil {
		return integreatlyv1alpha1.PhaseFailed, fmt.Errorf("Failed reconciling edit routes role %v: %w", editRoutesRole, err)
	}

	// Bind the amq online service admin role to the dedicated-admins group
	editRoutesRoleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dedicated-admins-edit-routes",
			Namespace: r.Config.GetNamespace(),
		},
	}

	_, err = controllerutil.CreateOrUpdate(ctx, client, editRoutesRoleBinding, func() error {
		owner.AddIntegreatlyOwnerAnnotations(editRoutesRoleBinding, r.installation)

		editRoutesRoleBinding.RoleRef = rbacv1.RoleRef{
			Name: editRoutesRole.GetName(),
			Kind: "Role",
		}
		editRoutesRoleBinding.Subjects = []rbacv1.Subject{
			{
				Name: "dedicated-admins",
				Kind: "Group",
			},
		}

		return nil
	})
	if err != nil {
		return integreatlyv1alpha1.PhaseFailed, fmt.Errorf("Failed reconciling service admin role binding %v: %w", editRoutesRoleBinding, err)
	}

	return integreatlyv1alpha1.PhaseCompleted, nil
}

func (r *Reconciler) reconcileSubscription(ctx context.Context, serverClient k8sclient.Client, inst *integreatlyv1alpha1.RHMI, productNamespace string, operatorNamespace string) (integreatlyv1alpha1.StatusPhase, error) {
	target := marketplace.Target{
		SubscriptionName: constants.ThreeScaleSubscriptionName,
		Namespace:        operatorNamespace,
	}

	catalogSourceReconciler, err := r.GetProductDeclaration().PrepareTarget(
		r.log,
		serverClient,
		marketplace.CatalogSourceName,
		&target,
	)
	if err != nil {
		return integreatlyv1alpha1.PhaseFailed, err
	}

	return r.Reconciler.ReconcileSubscription(
		ctx,
		target,
		[]string{productNamespace},
		r.preUpgradeBackupExecutor(),
		serverClient,
		catalogSourceReconciler,
		r.log,
	)
}

func (r *Reconciler) reconcileConsoleLink(ctx context.Context, serverClient k8sclient.Client) (integreatlyv1alpha1.StatusPhase, error) {
	cl := &consolev1.ConsoleLink{
		ObjectMeta: metav1.ObjectMeta{
			Name: "rhmi-3scale-console-link",
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, serverClient, cl, func() error {
		cl.Spec = consolev1.ConsoleLinkSpec{
			ApplicationMenu: &consolev1.ApplicationMenuSpec{
				ImageURL: threeScaleIcon,
				Section:  "OpenShift Managed Services",
			},
			Link: consolev1.Link{
				Href: fmt.Sprintf("%v/auth/rhsso/bounce", r.Config.GetHost()),
				Text: "API Management",
			},
			Location: consolev1.ApplicationMenu,
		}
		return nil
	})
	if err != nil {
		return integreatlyv1alpha1.PhaseFailed, fmt.Errorf("error creating or updating 3Scale console link, %s", err)
	}

	return integreatlyv1alpha1.PhaseCompleted, nil
}

func (r *Reconciler) deleteConsoleLink(ctx context.Context, serverClient k8sclient.Client) (integreatlyv1alpha1.StatusPhase, error) {
	cl := &consolev1.ConsoleLink{
		ObjectMeta: metav1.ObjectMeta{
			Name: "rhmi-3scale-console-link",
		},
	}

	err := serverClient.Delete(ctx, cl)
	if err != nil && !k8serr.IsNotFound(err) {
		return integreatlyv1alpha1.PhaseFailed, fmt.Errorf("error removing 3Scale console link, %s", err)
	}
	return integreatlyv1alpha1.PhaseCompleted, nil
}

func (r *Reconciler) reconcileDeploymentConfigs(ctx context.Context, serverClient k8sclient.Client, productNamespace string) (integreatlyv1alpha1.StatusPhase, error) {

	for _, name := range threeScaleDeploymentConfigs {
		deploymentConfig := &appsv1.DeploymentConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: productNamespace,
			},
		}

		podPriorityMutation := resources.NoopMutate
		if r.installation.Spec.Type == string(integreatlyv1alpha1.InstallationTypeManagedApi) {
			podPriorityMutation = resources.MutatePodPriority(r.installation.Spec.PriorityClassName)

		}

		phase, err := resources.UpdatePodTemplateIfExists(
			ctx,
			serverClient,
			resources.SelectFromDeploymentConfig,
			resources.AllMutationsOf(
				resources.MutateZoneTopologySpreadConstraints("app"),
				podPriorityMutation,
			),
			deploymentConfig,
		)
		if err != nil || phase != integreatlyv1alpha1.PhaseCompleted {
			return phase, err
		}
	}

	return integreatlyv1alpha1.PhaseCompleted, nil
}

func (r *Reconciler) changesDeploymentConfigsEnvVar(ctx context.Context, serverClient k8sclient.Client) (integreatlyv1alpha1.StatusPhase, error) {

	for _, name := range threeScaleDeploymentConfigs {
		deploymentConfig := &appsv1.DeploymentConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: r.Config.GetNamespace(),
			},
		}

		objKey, err := k8sclient.ObjectKeyFromObject(deploymentConfig)
		if err != nil {
			return integreatlyv1alpha1.PhaseFailed, err
		}

		if err := serverClient.Get(ctx, objKey, deploymentConfig); err != nil {
			if k8serr.IsNotFound(err) {
				return integreatlyv1alpha1.PhaseInProgress, nil
			}
			return integreatlyv1alpha1.PhaseFailed, err
		}

		if name == systemAppDCName {
			envVars := make(map[string]corev1.EnvVarSource)
			backendListenerServiceEndpoint := &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "backend-listener",
					},
					Key: "service_endpoint",
				},
			}

			backendListenerRoute := &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "backend-listener",
					},
					Key: "route_endpoint",
				},
			}
			envVars["APICAST_BACKEND_ROOT_ENDPOINT"] = *backendListenerRoute
			envVars["BACKEND_ROUTE"] = *backendListenerServiceEndpoint
			envVars["BACKEND_PUBLIC_URL"] = *backendListenerRoute

			// Have to use the index when iterating here because when using range go creates a copy of the variable
			// so any update will be applyed to the copy
			for envVarName, _ := range envVars {
				foundEnv := false
				envVarValue := envVars[envVarName]

				if deploymentConfig.Spec.Strategy.RollingParams != nil {
					for i, env := range deploymentConfig.Spec.Strategy.RollingParams.Pre.ExecNewPod.Env {
						if env.Name == envVarName {
							deploymentConfig.Spec.Strategy.RollingParams.Pre.ExecNewPod.Env[i].ValueFrom = &envVarValue
						}
					}
				}

				for i, container := range deploymentConfig.Spec.Template.Spec.Containers {
					for j, env := range container.Env {
						if env.Name == envVarName {
							foundEnv = true
							r.log.Infof("updating env variable to system app", l.Fields{"envVarName": envVarName, "envVarValue": envVarValue, "foundVariable": foundEnv})
							deploymentConfig.Spec.Template.Spec.Containers[i].Env[j].ValueFrom = &envVarValue
						}
					}

					if foundEnv == false {
						r.log.Infof("adding env variable to system app", l.Fields{"envVarName": envVarName, "envVarValue": envVarValue, "foundVariable": foundEnv})

						deploymentConfig.Spec.Template.Spec.Containers[i].Env = append(
							deploymentConfig.Spec.Template.Spec.Containers[i].Env,
							corev1.EnvVar{Name: envVarName, ValueFrom: &envVarValue},
						)
					}
				}
			}

			if err := serverClient.Update(ctx, deploymentConfig); err != nil {
				return integreatlyv1alpha1.PhaseFailed, err
			}
		}
	}
	return integreatlyv1alpha1.PhaseCompleted, nil
}

// Deployment configs are rescaled when adding topologySpreadConstraints, PodTopology etc
// Should check that these deployment configs are ready before returning phase complete in CR
func (r *Reconciler) ensureDeploymentConfigsReady(ctx context.Context, serverClient k8sclient.Client, productNamespace string) (integreatlyv1alpha1.StatusPhase, error) {
	for _, name := range threeScaleDeploymentConfigs {
		deploymentConfig := &appsv1.DeploymentConfig{}

		err := serverClient.Get(ctx, k8sclient.ObjectKey{Name: name, Namespace: productNamespace}, deploymentConfig)

		if err != nil {
			return integreatlyv1alpha1.PhaseFailed, err
		}

		// Rollout new dc if there is a failed condition
		for _, condition := range deploymentConfig.Status.Conditions {
			if condition.Status == corev1.ConditionFalse {
				r.log.Warningf("3scale dc in a failed condition, rolling out new deployment", l.Fields{"dc": name})
				err = r.RolloutDeployment(ctx, name)
				if err != nil {
					return integreatlyv1alpha1.PhaseFailed, err
				}

				return integreatlyv1alpha1.PhaseCreatingComponents, nil
			}
		}

		//  Check that replicas are fully rolled out
		for _, condition := range deploymentConfig.Status.Conditions {
			if condition.Status != corev1.ConditionTrue || (deploymentConfig.Status.Replicas != deploymentConfig.Status.AvailableReplicas ||
				deploymentConfig.Status.ReadyReplicas != deploymentConfig.Status.UpdatedReplicas) {
				r.log.Infof("waiting for 3scale dc to become ready", l.Fields{"dc": name})
				return integreatlyv1alpha1.PhaseInProgress, fmt.Errorf("waiting for 3scale deployment config %s to become available", name)
			}
		}
	}

	return integreatlyv1alpha1.PhaseCompleted, nil
}

func (r *Reconciler) reconcileRatelimitingTo3scaleComponents(ctx context.Context, serverClient k8sclient.Client, installation *integreatlyv1alpha1.RHMI) (integreatlyv1alpha1.StatusPhase, error) {

	r.log.Info("Reconciling rate limiting settings to 3scale components")

	proxyServer := ratelimit.NewEnvoyProxyServer(ctx, serverClient, r.log)

	err := r.createBackendListenerProxyService(ctx, serverClient)
	if err != nil {
		return integreatlyv1alpha1.PhaseInProgress, err
	}

	// creates envoy proxy sidecar container for apicast staging
	phase, err := proxyServer.CreateEnvoyProxyContainer(
		apicastStagingDCName,
		r.Config.GetNamespace(),
		ApicastNodeID,
		apicastStagingDCName,
		"gateway",
		ApicastEnvoyProxyPort,
	)
	if phase != integreatlyv1alpha1.PhaseCompleted {
		return phase, err
	}

	// creates envoy proxy sidecar container for apicast production
	phase, err = proxyServer.CreateEnvoyProxyContainer(
		apicastProductionDCName,
		r.Config.GetNamespace(),
		ApicastNodeID,
		apicastProductionDCName,
		"gateway",
		ApicastEnvoyProxyPort,
	)
	if phase != integreatlyv1alpha1.PhaseCompleted {
		return phase, err
	}

	// creates envoy proxy sidecar container for backend listener
	phase, err = proxyServer.CreateEnvoyProxyContainer(
		backendListenerDCName,
		r.Config.GetNamespace(),
		BackendNodeID,
		BackendServiceName,
		"http",
		BackendEnvoyProxyPort,
	)
	if phase != integreatlyv1alpha1.PhaseCompleted {
		return phase, err
	}

	r.log.Info("Finished creating envoy sidecar containers for 3scale components")

	// setting up envoy config
	ratelimitServiceCR, err := r.getRateLimitServiceCR(ctx, serverClient)
	if err != nil {
		return integreatlyv1alpha1.PhaseFailed, err
	}

	// rate limit cluster
	ratelimitClusterResource := ratelimit.CreateClusterResource(
		ratelimitServiceCR.Spec.ClusterIP,
		ratelimit.RateLimitClusterName,
		getRatelimitServicePort(ratelimitServiceCR),
	)
	ratelimitClusterResource.Http2ProtocolOptions = &envoycore.Http2ProtocolOptions{}

	// apicast cluster
	apiCastClusterResource := ratelimit.CreateClusterResource(
		ApicastContainerAddress,
		ApicastClusterName,
		ApicastContainerPort,
	)

	// apicast listener
	apiCastFilters, _ := getListenerResourceFilters(
		getAPICastVirtualHosts(ApicastClusterName),
		getAPICastHTTPFilters(),
	)
	apiCastListenerResource := ratelimit.CreateListenerResource(
		ApicastListenerName,
		ApicastEnvoyProxyAddress,
		ApicastEnvoyProxyPort,
		apiCastFilters,
	)

	// create envoy config for apicast
	apiCastProxyConfig := ratelimit.NewEnvoyConfig(ApicastClusterName, r.Config.GetNamespace(), ApicastNodeID)
	err = apiCastProxyConfig.CreateEnvoyConfig(ctx, serverClient, []*envoyapi.Cluster{apiCastClusterResource, ratelimitClusterResource}, []*envoyapi.Listener{apiCastListenerResource}, installation)
	if err != nil {
		r.log.Errorf("Failed to create envoyconfig for apicast", l.Fields{"APICast": ApicastClusterName}, err)
		return integreatlyv1alpha1.PhaseFailed, err
	}

	// backend-listener cluster
	backendClusterResource := ratelimit.CreateClusterResource(
		BackendContainerAddress,
		BackendClusterName,
		BackendContainerPort,
	)

	backendHTTPFilters, _ := getBackendListenerHTTPFilters()
	// backend listener listener
	backendFilters, _ := getListenerResourceFilters(
		getBackendListenerVitualHosts(BackendClusterName),
		backendHTTPFilters,
	)
	backendListenerResource := ratelimit.CreateListenerResource(
		BackendListenerName,
		BackendEnvoyProxyAddress,
		BackendEnvoyProxyPort,
		backendFilters,
	)

	// create envoy config for backend listener
	backendProxyConfig := ratelimit.NewEnvoyConfig(BackendClusterName, r.Config.GetNamespace(), BackendNodeID)
	err = backendProxyConfig.CreateEnvoyConfig(ctx, serverClient, []*envoyapi.Cluster{backendClusterResource, ratelimitClusterResource}, []*envoyapi.Listener{backendListenerResource}, installation)
	if err != nil {
		r.log.Errorf("Failed to create envoyconfig for backend-listener", l.Fields{"BackendListener": BackendClusterName}, err)
		return integreatlyv1alpha1.PhaseFailed, err
	}

	return integreatlyv1alpha1.PhaseCompleted, nil
}

func (r *Reconciler) getRateLimitServiceCR(ctx context.Context, serverClient k8sclient.Client) (*corev1.Service, error) {
	rateLimitService := &corev1.Service{}
	marin3rConfig, err := r.ConfigManager.ReadMarin3r()
	if err != nil {
		return nil, fmt.Errorf("failed to load marin3r config in 3scale reconciler: %v", err)
	}
	err = serverClient.Get(ctx, k8sclient.ObjectKey{
		Namespace: marin3rConfig.GetNamespace(),
		Name:      "ratelimit",
	}, rateLimitService)

	if err != nil {
		return nil, fmt.Errorf("failed to rate limiting service: %v", err)
	}

	return rateLimitService, nil
}

func getRatelimitServicePort(rateLimitService *corev1.Service) int {
	for _, port := range rateLimitService.Spec.Ports {
		if port.Name == "grpc" {
			return port.TargetPort.IntValue()
		}
	}
	return 0
}

func (r *Reconciler) createBackendListenerProxyService(ctx context.Context, serverClient k8sclient.Client) error {

	backendListenerService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      BackendServiceName,
			Namespace: r.Config.GetNamespace(),
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, serverClient, backendListenerService, func() error {
		owner.AddIntegreatlyOwnerAnnotations(backendListenerService, r.installation)
		backendListenerService.Spec.Ports = []corev1.ServicePort{
			{
				Name:       "http",
				Protocol:   "TCP",
				Port:       BackendEnvoyProxyPort,
				TargetPort: intstr.FromInt(BackendEnvoyProxyPort),
			},
		}
		backendListenerService.Spec.Selector = map[string]string{
			"deploymentConfig": backendListenerDCName,
		}
		return nil
	}); err != nil {
		return err
	}

	// links the backend listener proxy service to the external backend listener route
	backendRoute, err := r.getBackendListenerRoute(ctx, serverClient)
	if err != nil {
		return err
	}

	backendRoute.Spec.To.Name = backendListenerService.Name
	err = serverClient.Update(ctx, backendRoute)
	if err != nil {
		return fmt.Errorf("Error updating the backend-listener external route to the backend-listener proxy server: %v", err)
	}

	r.log.Infof("Created service to rate limit external backend-listener route",
		l.Fields{"ServiceName": backendListenerService.Name},
	)
	return nil
}

func (r *Reconciler) getBackendListenerRoute(ctx context.Context, serverClient k8sclient.Client) (*routev1.Route, error) {
	backendRoute := &routev1.Route{}
	err := serverClient.Get(ctx, k8sclient.ObjectKey{
		Namespace: r.Config.GetNamespace(),
		Name:      "backend",
	}, backendRoute)
	if err != nil {
		return nil, fmt.Errorf("Error getting the backend-listener external route: %v", err)
	}
	return backendRoute, nil
}
