package constants

const (
	REAL_SUCCESS_CODE             = "201"
	REAL_PENDING_CODE             = "200"
	INVALID_MANDATORY_FIELD_CODE  = "214"
	VALIDATE_FIELD_ERROR_CODE     = "215"
	INVALID_AUTHENTICATION_CODE   = "216"
	INVALID_PIN_CODE              = "217"
	INVALID_INQUIRY_CODE          = "218"
	INVALID_SESSION_ID_CODE       = "219"
	INVALID_TRANSACTION_TIME_CODE = "220"

	BENEFICIARY_ACCOUNT_NOT_FOUND_CODE  = "421"
	BENEFICIARY_ACCOUNT_EXIST_CODE      = "422"
	BENEFICIARY_ACCOUNT_INACTIVE_CODE   = "423"
	BENEFICIARY_ACCOUNT_ACTIVE_CODE     = "424"
	BENEFICIARY_ACCOUNT_UNVERIFIED_CODE = "425"

	ACCOUNT_NOT_FOUND_CODE  = "426"
	ACCOUNT_EXIST_CODE      = "427"
	ACCOUNT_INACTIVE_CODE   = "428"
	ACCOUNT_CONNECTED_CODE  = "429"
	ACCOUNT_UNVERIFIED_CODE = "430"

	ACCOUNT_BLOCKED_SYSTEM_CODE        = "431"
	ACCOUNT_BLOCKED_PIN_CODE           = "432"
	INSUFFICIENT_BALANCE_CODE          = "433"
	ACCOUNT_BALANCE_EXCEED_CODE        = "434"
	ACCOUNT_BALANCE_BELOW_MINIMUM_CODE = "435"
	TRX_EXCEED_LIMITATION_CODE         = "436"
	TRX_BELOW_LIMITATION_CODE          = "437"
	ACCOUNT_NOT_UPGRADE_CODE           = "438"
	ACCOUNT_NOT_CONNECTED_CODE         = "439"
	TRX_NOT_FOUND_CODE                 = "440"
	TRX_FOUND_CODE                     = "441"

	SYSTEM_ERROR_CODE = "501"
	// End Status Code

	// Status Message
	REAL_SUCCESS_MESSAGE             = "Success"
	REAL_PENDING_MESSAGE             = "Pending"
	INVALID_MANDATORY_FIELD_MESSAGE  = "Invalid Mandatory Field or Value"
	VALIDATE_FIELD_ERROR_MESSAGE     = "Invalid Field"
	INVALID_AUTHENTICATION_MESSAGE   = "Invalid Authentication"
	INVALID_PIN_MESSAGE              = "PIN salah, Akun Akan diblokir setelah 3 percobaan gagal"
	INVALID_INQUIRY_MESSAGE          = "Invalid Inquiry"
	INVALID_SESSION_ID_MESSAGE       = "Invalid Session ID"
	INVALID_TRANSACTION_TIME_MESSAGE = "Transaksi tidak dapat diproses pada waktu cutoff"

	BENEFICIARY_ACCOUNT_NOT_FOUND_MESSAGE  = "Akun Penerima Tidak Ditemukan"
	BENEFICIARY_ACCOUNT_EXIST_MESSAGE      = "Akun Penerima Tersedia"
	BENEFICIARY_ACCOUNT_INACTIVE_MESSAGE   = "Akun Penerima Tidak aktif"
	BENEFICIARY_ACCOUNT_ACTIVE_MESSAGE     = "Akun Penerima Aktif"
	BENEFICIARY_ACCOUNT_UNVERIFIED_MESSAGE = "Akun Penerima Belum Verifikasi"

	ACCOUNT_NOT_FOUND_MESSAGE     = "Akun Tidak Ditemukan"
	ACCOUNT_EXIST_MESSAGE         = "Akun Telah Tersedia"
	ACCOUNT_INACTIVE_MESSAGE      = "Akun Tidak Aktif"
	ACCOUNT_CONNECTED_MESSAGE     = "Akun Telah Terhubung"
	ACCOUNT_NOT_CONNECTED_MESSAGE = "Akun Belum Terhubung"
	ACCOUNT_UNVERIFIED_MESSAGE    = "Akun Belum Verifikasi"
	ACCOUNT_NOT_UPGRADE_MESSAGE   = "Akun Belum Upgrade Ke Premium"

	ACCOUNT_BLOCKED_SYSTEM_MESSAGE      = "Akun Terblokir Oleh System"
	ACCOUNT_BLOCKED_PIN_MESSAGE         = "Akun Terblokir, Kegagalan Autentikasi Berulang"
	INSUFFICIENT_BALANCE_MESSAGE        = "Saldo Tidak Cukup"
	ACCOUNT_BALANCE_EXCEED_MESSAGE      = "Saldo Melebihi Batas Maksimum"
	ACCOUNT_BALANCE_BELOW_MINUM_MESSAGE = "Saldo Dibawah Batas Minimum"
	TRX_EXCEED_LIMITATION_MESSAGE       = "Transaksi Melebihi Batas Yang Ditetapkan"
	TRX_BELOW_LIMITATION_MESSAGE        = "Transaksi Dibawah Batas Yang Ditetapkan"
	TRX_NOT_FOUND_MESSAGE               = "Transaksi Tidak Ditemukan"
	TRX_FOUND_MESSAGE                   = "Transaksi Telah Tersedia"

	SYSTEM_ERROR_MESSAGE = "Terjadi Kesalahan Pada Sistem, Silahkan Coba Kembali"
)
