package risks

import (
	accidentalsecretleak "github.com/threagile/threagile/pkg/model/risks/built-in/accidental-secret-leak"
)

func GetBuiltInRiskRules() []model.RiskRule {
	return []model.RiskRule{
		accidentalsecretleak.Rule(),

		codebackdooring.Rule(),
		containerbaseimagebackdooring.Rule(),
		containerplatformescape.Rule(),
		crosssiterequestforgery.Rule(),
		crosssitescripting.Rule(),
		dosriskyaccessacrosstrustboundary.Rule(),
		incompletemodel.Rule(),
		ldapinjection.Rule(),
		missingauthentication.Rule(),
		missingauthenticationsecondfactor.Rule(),
		missingbuildinfrastructure.Rule(),
		missingcloudhardening.Rule(),
		missingfilevalidation.Rule(),
		missinghardening.Rule(),
		missingidentitypropagation.Rule(),
		missingidentityproviderisolation.Rule(),
		missingidentitystore.Rule(),
		missingnetworksegmentation.Rule(),
		missingvault.Rule(),
		missingvaultisolation.Rule(),
		missingwaf.Rule(),
		mixedtargetsonsharedruntime.Rule(),
		pathtraversal.Rule(),
		pushinsteadofpulldeployment.Rule(),
		searchqueryinjection.Rule(),
		serversiderequestforgery.Rule(),
		serviceregistrypoisoning.Rule(),
		sqlnosqlinjection.Rule(),
		uncheckeddeployment.Rule(),
		unencryptedasset.Rule(),
		unencryptedcommunication.Rule(),
		unguardedaccessfrominternet.Rule(),
		unguardeddirectdatastoreaccess.Rule(),
		unnecessarycommunicationlink.Rule(),
		unnecessarydataasset.Rule(),
		unnecessarydatatransfer.Rule(),
		unnecessarytechnicalasset.Rule(),
		untrusteddeserialization.Rule(),
		wrongcommunicationlinkcontent.Rule(),
		wrongtrustboundarycontent.Rule(),
		xmlexternalentity.Rule(),
	}
}
