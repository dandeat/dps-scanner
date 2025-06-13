package mkpmobileutils

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/labstack/echo"
	"golang.org/x/crypto/bcrypt"
)

//Timestamp RFC3339
func Timestamp() string {
	return time.Now().Format(time.RFC3339)
}
func TimetoStr(t time.Time) string {
	return t.Local().String()
}
func TimetoStrFormat(t time.Time, layout string) string {
	return t.Local().Format(layout)
}
func StrToTimeLocal(tm string) (time.Time, error) {
	t, err := time.Parse(LAYOUT_TIMESTAMP, tm)
	if err != nil {
		return time.Now(), err
	}

	return t.Local(), nil
}

//DateTime
func DatetimeNowPgFormat(t time.Time) string {
	return time.Now().Format(LAYOUT_TIMESTAMP)
}
func DatetimeNow() string {
	return time.Now().Format(LAYOUT_DATETIME_STRING)
}
func DateNow() string {
	return time.Now().Format(LAYOUT_DATE)
}
func DatetimeLayoutNow(layout string) string {
	return time.Now().Format(layout)
}

//Masking 12345678 :> *****678
func MaskString(s string) string {
	rs := []rune(s)
	for i := 0; i < len(rs)-4; i++ {
		rs[i] = '*'
	}
	return string(rs)
}

//Bind Validate Struct
func BindValidateStruct(ctx echo.Context, i interface{}) error {
	if err := ctx.Bind(i); err != nil {
		return err
	}

	if err := ctx.Validate(i); err != nil {
		return err
	}
	return nil
}

//Response JSON
func ResponseJSON(success bool, code string, msg string, result interface{}, addMsg ...string) Response {
	tm := time.Now()
	response := Response{
		Meta: Meta{
			Code:          code,
			Success:       success,
			Message:       msg,
			AdditionalMsg: addMsg,
		},
		Result:           result,
		ResponseDatetime: tm,
	}

	return response
}

//Response JSON v1
func ResponseJSONV1(code string, msg string, result interface{}) ResponseV1 {
	tm := time.Now()
	response := ResponseV1{
		Result:           result,
		ResponseCode:     code,
		ResponseMessage:  msg,
		ResponseDatetime: tm,
	}

	return response
}

//Response JSON v1 Encrypted
func ResponseJSONV1Enc(code string, msg string, result interface{}, encIv, scretKey string) (ResponseV1Enc, error) {
	tm := time.Now()
	response := ResponseV1Enc{
		Result:           result,
		ResponseCode:     code,
		ResponseMessage:  msg,
		ResponseDatetime: tm,
	}

	resStr, err := json.Marshal(response)
	if err != nil {
		return response, err
	}

	encResult, err := AES256EncryptV2(string(resStr), encIv, scretKey)
	if err != nil {
		return response, err
	}
	response.MetaResult = encResult

	return response, nil
}

// Generate Random Number
func RandNumber(max int64) (n int64, err error) {
	nBig, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return n, err
	}
	n = nBig.Int64()
	return n, nil
}

func Int32ToString(n int32) string {
	buf := [11]byte{}
	pos := len(buf)
	i := int64(n)
	signed := i < 0
	if signed {
		i = -i
	}
	for {
		pos--
		buf[pos], i = '0'+byte(i%10), i/10
		if i == 0 {
			if signed {
				pos--
				buf[pos] = '-'
			}
			return string(buf[pos:])
		}
	}
}

func ReplaceSQL(old, searchPattern string) string {
	tmpCount := strings.Count(old, searchPattern)
	for m := 1; m <= tmpCount; m++ {
		old = strings.Replace(old, searchPattern, "$"+strconv.Itoa(m), 1)
	}
	return old
}

func ValBlankOrNull(request interface{}, keyName ...string) error {
	var params interface{}
	_ = json.Unmarshal([]byte(ToString(request)), &params)
	paramsValue := params.(map[string]interface{})

	for idx := range keyName {
		name := keyName[idx]
		if len(strings.TrimSpace(paramsValue[name].(string))) == 0 {
			return fmt.Errorf("%s must be filled", name)
		}
	}

	return nil
}

func InArray(v interface{}, in interface{}) (ok bool, i int) {
	val := reflect.Indirect(reflect.ValueOf(in))
	switch val.Kind() {
	case reflect.Slice, reflect.Array:
		for ; i < val.Len(); i++ {
			if ok = v == val.Index(i).Interface(); ok {
				return
			}
		}
	}
	return
}

// Make hash
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// Check hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func ToString(i interface{}) string {
	log, _ := json.Marshal(i)
	logString := string(log)

	return logString
}

func CreateCredential(secret string, value string) (result string, err error) {

	// Create a new HMAC by defining the hash type and the key (as byte array)
	h := hmac.New(sha256.New, []byte(secret))

	// Write Data to it
	h.Write([]byte(value))

	// Get result and encode as hexadecimal string
	sha := hex.EncodeToString(h.Sum(nil))

	db, err := decodeHex([]byte(sha))
	if err != nil {
		fmt.Printf("failed to decode hex: %s", err)
		return
	}

	f := base64Encode(db)

	return string(f), err
}

func base64Encode(input []byte) []byte {
	eb := make([]byte, base64.StdEncoding.EncodedLen(len(input)))
	base64.StdEncoding.Encode(eb, input)

	return eb
}

func decodeHex(input []byte) ([]byte, error) {
	db := make([]byte, hex.DecodedLen(len(input)))
	_, err := hex.Decode(db, input)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func Base64ToHex(s string) string {
	p, _ := base64.StdEncoding.DecodeString(s)
	h := hex.EncodeToString(p)
	return h
}

func AES256EncryptV2(message, encIv, secretKey string) (string, error) {

	block, err := aes.NewCipher([]byte(secretKey))
	if err != nil {
		return "", err
	}

	iv := []byte(encIv)

	enc := cipher.NewCBCEncrypter(block, iv)
	content := PKCS5Padding([]byte(message), block.BlockSize())
	crypted := make([]byte, len(content))
	enc.CryptBlocks(crypted, content)

	return base64.StdEncoding.EncodeToString(crypted), nil
}

func AES256DecryptV2(message, encIv, secretKey string) (string, error) {

	block, err := aes.NewCipher([]byte(secretKey))
	if err != nil {
		return "", err
	}

	iv := []byte(encIv)

	messageData, _ := base64.StdEncoding.DecodeString(message)
	dec := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(messageData))
	dec.CryptBlocks(decrypted, messageData)

	return string(PKCS5Unpadding(decrypted)), nil
}

// Decrypt from base64 to decrypted string
func Aes256Decrypt(cryptoText string, saltKey ...interface{}) (interface{}, error) {
	var result interface{}
	keyText := ""
	if len(saltKey) > 0 {
		keyText = saltKey[0].(string)
	}
	key := []byte(keyText)
	ciphertext, _ := base64.URLEncoding.DecodeString(cryptoText)

	block, err := aes.NewCipher(key)
	if err != nil {
		return result, err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		return result, err
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(ciphertext, ciphertext)
	unMarshall := json.Unmarshal(ciphertext, &result)
	fmt.Println(unMarshall)
	return result, nil
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5Unpadding(encrypt []byte) []byte {
	padding := encrypt[len(encrypt)-1]
	return encrypt[:len(encrypt)-int(padding)]
}

// -------------------------------- Digital Signature --------------------------------
func RmSpace(s string) string {
	rr := make([]rune, 0, len(s))
	for _, r := range s {
		if !unicode.IsSpace(r) {
			rr = append(rr, r)
		}
	}
	return string(rr)
}

func CreateDigitalSignature(signaturePayload DigitalSignature, scretKey string) (result string, err error) {

	minifyBody, err := json.Marshal(signaturePayload.RequestBody)
	if err != nil {
		return result, err
	}

	// trimmedBody := RmSpace(string(minifyBody))

	h := sha256.New()
	h.Write([]byte(minifyBody))
	b := h.Sum(nil)

	c := hex.EncodeToString(b)

	lower := strings.ToLower(c)

	strToSign := signaturePayload.HttpMethod +
		":" + signaturePayload.EndpointUrl +
		":" + signaturePayload.AccessToken +
		":" + lower +
		":" + signaturePayload.Timestamp

	sig := hmac.New(sha512.New, []byte(scretKey))
	sig.Write([]byte(strToSign))

	// result := hex.EncodeToString(sig.Sum(nil))
	result = base64.StdEncoding.EncodeToString(sig.Sum(nil))

	return result, err
}

// ------------------------------ End Digital Signature ------------------------------

// --------------------------------- Token Signature ---------------------------------
func CreateTokenSignature(req TokenSignature, keyPath string) (token string, err error) {

	signer, err := loadPrivateKey(keyPath)
	if err != nil {
		return "", err
	}

	stringToSign := req.Username + "|" + req.Timestamp

	token = signer.SHA256withRSA(stringToSign)
	return token, nil
}

func loadPrivateKey(path string) (Signer, error) {
	return parsePrivateKey([]byte(path))
}

func (r *rsaPrivateKey) SHA256withRSA(stringToSign string) string {
	h := sha256.New()
	h.Write([]byte(stringToSign))
	d := h.Sum(nil)
	signature, err := rsa.SignPKCS1v15(rand.Reader, r.PrivateKey, crypto.SHA256, d)
	if err != nil {
		panic(err)
	}
	encodedSig := base64.StdEncoding.EncodeToString(signature)
	return encodedSig
}

func parsePrivateKey(pemBytes []byte) (Signer, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("ssh: no key found")
	}

	var rawkey interface{}
	switch block.Type {
	case "RSA PRIVATE KEY":
		rsa, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		rawkey = rsa
	default:
		return nil, fmt.Errorf("ssh: unsupported key type %q", block.Type)
	}
	return newSignerFromKey(rawkey)
}

func newSignerFromKey(k interface{}) (Signer, error) {
	var sshKey Signer
	switch t := k.(type) {
	case *rsa.PrivateKey:
		sshKey = &rsaPrivateKey{t}
	default:
		return nil, fmt.Errorf("ssh: unsupported key type %T", k)
	}
	return sshKey, nil
}

// ------------------------------- End Token Signature -------------------------------

// DB Transactions
func DBTransaction(db *sql.DB, txFunc func(*sql.Tx) error) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // Rollback Panic
		} else if err != nil {
			tx.Rollback() // err is not nill
		} else {
			err = tx.Commit() // err is nil
		}
	}()
	err = txFunc(tx)
	return err
}
