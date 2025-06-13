package models

type (
	Client struct {
		ID                            int64  `json:"id"`
		PartnerID                     int64  `json:"partner_id"`
		RoleID                        int64  `json:"role_id"`
		ClientID                      string `json:"client_id"`
		ClientSecret                  string `json:"client_secret"`
		PartnerPrivateKey             string `json:"private_key"`
		PartnerPublicKey              string `json:"public_key"`
		CallbackUpgradeSavingUrl      string `json:"callback_upgrade_saving_url"`
		CallbackActivationPaylaterUrl string `json:"callback_activation_paylater_url"`
		CallbackPaymentUrl            string `json:"callback_payment_url"`
		CreatedAt                     string `json:"created_at"`
		UpdatedAt                     string `json:"updated_at"`
		CreatedBy                     string `json:"created_by"`
		UpdatedBy                     string `json:"updated_by"`
		Active                        string `json:"active"`
	}

	ClientReturn struct {
		ID                            int64   `json:"id"`
		PartnerID                     int64   `json:"partner_id"`
		PartnerName                   string  `json:"partner_name"`
		RoleID                        int64   `json:"role_id"`
		RoleName                      string  `json:"role_name"`
		ClientID                      string  `json:"client_id"`
		ClientSecret                  string  `json:"client_secret"`
		PartnerPrivateKey             string  `json:"private_key"`
		PartnerPublicKey              string  `json:"public_key"`
		CallbackUpgradeSavingUrl      string  `json:"callback_upgrade_saving_url"`
		CallbackActivationPaylaterUrl string  `json:"callback_activation_paylater_url"`
		CallbackPaymentUrl            string  `json:"callback_payment_url"`
		CreatedAt                     string  `json:"created_at"`
		UpdatedAt                     *string `json:"updated_at"`
		CreatedBy                     string  `json:"created_by"`
		UpdatedBy                     string  `json:"updated_by"`
		Active                        string  `json:"active"`
	}

	ClientIndex struct {
		PartnerID    int64  `json:"partnerID"`
		RoleID       int64  `json:"roleID"`
		ClientID     string `json:"clientID"`
		ClientSecret string `json:"clientSecret"`
	}

	RequestGetClientByCifID struct {
		CifID int64 `json:"cifID"`
	}

	RequestGetClientList struct {
		PartnerID int64 `json:"partnerID"`
		RoleID    int64 `json:"roleID"`

		Limit  int64 `json:"limit" validate:"required"`
		Offset int64 `json:"offset"`
	}

	RequestGetClientListDashboard struct {
		PartnerID int64 `json:"partnerID"`
		RoleID    int64 `json:"roleID"`

		Search     string `json:"search"`
		OrderBy    string `json:"orderBy"`
		Order      string `json:"order"`
		PageNumber int64  `json:"pageNumber" validate:"required"`
		PageSize   int64  `json:"pageSize" validate:"required"`
	}

	ResponseGetClientListDashboard struct {
		RecordsFiltered int            `json:"recordsFiltered"`
		RecordsTotal    int            `json:"recordsTotal"`
		Value           []ClientReturn `json:"value"`
	}

	RequestGetClient struct {
		ClientID int64 `json:"clientID" validate:"required"`
	}

	RequestAddClient struct {
		PartnerID                     int64  `json:"partnerID" validate:"required"`
		RoleID                        int64  `json:"roleID" validate:"required"`
		PartnerPrivateKey             string `json:"partnerPrivateKey"`
		PartnerPublicKey              string `json:"partnerPublicKey" `
		CallbackUpgradeSavingUrl      string `json:"callback_upgrade_saving_url"`
		CallbackActivationPaylaterUrl string `json:"callback_activation_paylater_url"`
		CallbackPaymentUrl            string `json:"callback_payment_url"`
	}

	RequestUpdateClient struct {
		ID                            int64  `json:"id" validate:"required"`
		RoleID                        int64  `json:"roleID" validate:"required"`
		PartnerPrivateKey             string `json:"partnerPrivateKey" validate:"required"`
		PartnerPublicKey              string `json:"partnerPublicKey" validate:"required"`
		CallbackUpgradeSavingUrl      string `json:"callback_upgrade_saving_url"`
		CallbackActivationPaylaterUrl string `json:"callback_activation_paylater_url"`
		CallbackPaymentUrl            string `json:"callback_payment_url"`
	}

	RequestRemoveClient struct {
		ClientID int64 `json:"clientID" validate:"required"`
	}

	ResposenGetClientByCifInternal struct {
		LinkedAccountID               int64  `json:"linkedAccountID"`
		LinkedAccountParterID         int64  `json:"linkedAccountParterID"`
		ClientID                      int64  `json:"clientID"`
		ClientPartnerID               int64  `json:"clientPartnerID"`
		CallbackUpgradeSavingUrl      string `json:"callbackUpgradeSavingUrl"`
		CallbackActivationPaylaterUrl string `json:"callbackActivationPaylaterEmail"`
		CallbackPaymentUrl            string `json:"callback_payment_url"`
	}
)
