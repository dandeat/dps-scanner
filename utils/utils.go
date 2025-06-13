package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	crRand "crypto/rand"
	cryptoutils "crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"database/sql"
	"dps-scanner-gateout/config"
	"dps-scanner-gateout/constants"
	"dps-scanner-gateout/models"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"mime/multipart"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	mkpmobileutils "github.com/dandeat/mkpmobile-utils/src/utils"
	"github.com/disintegration/imaging"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func GenerateBasicAuth(username, password string) string {
	basicAuthEncode := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
	return basicAuthEncode
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func StrToTimeLocal(tm, layout string) (time.Time, error) {
	t, err := time.Parse(layout, tm)
	if err != nil {
		return time.Now(), err
	}

	return t.Local(), nil
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

func ResponseJSON(success bool, code string, msg string, result interface{}) models.ResponseJSON {
	tm := time.Now()
	response := models.ResponseJSON{
		Success:          success,
		StatusCode:       code,
		Result:           result,
		Message:          msg,
		ResponseDatetime: tm,
	}

	return response
}

// Response JSON v1
func ResponseJSONV1(code string, msg string, result interface{}) models.ResponseV1 {
	tm := time.Now()
	response := models.ResponseV1{
		Result:           result,
		ResponseCode:     code,
		ResponseMessage:  msg,
		ResponseDatetime: tm,
	}

	return response
}

func GetExternalId() string {
	// t := time.Now()
	// dbTime := t.Format("060102150405")
	rand1 := rand.Intn(999999-100000) + 1
	rand1str := strconv.Itoa(rand1)
	// result := dbTime + rand1str + rand1str + rand1str + "00"
	result := rand1str + rand1str + rand1str + "00"
	// return result
	return result[0:9]
}

func CreateDigitalSignature(signaturePayload mkpmobileutils.DigitalSignature, scretKey string) (result string, err error) {

	minifyBody, err := json.Marshal(signaturePayload.RequestBody)
	if err != nil {
		return result, err
	}

	trimmedBody, err := Minify(minifyBody)
	if err != nil {

		fmt.Printf("err.Error(): %v\n", err.Error())
	}

	h := sha256.New()
	h.Write(trimmedBody)
	b := h.Sum(nil)

	c := hex.EncodeToString(b)

	lower := strings.ToLower(c)

	// strToSign := "path=" + signaturePayload.EndpointUrl +
	// 	"&verb=" + signaturePayload.HttpMethod +
	// 	"&token=Bearer " + signaturePayload.AccessToken +
	// 	"&timestamp=" + signaturePayload.Timestamp +
	// 	"&body=" + string(minifyBody)
	// fmt.Printf("strToSign: %v\n", strToSign)

	strToSign := signaturePayload.HttpMethod +
		":" + signaturePayload.EndpointUrl +
		":" + signaturePayload.AccessToken +
		":" + lower +
		":" + signaturePayload.Timestamp

	resultValue, _ := EscapeHTML(strToSign, true)

	sig := hmac.New(sha512.New, []byte(scretKey))
	log.Println(string(resultValue))
	sig.Write(resultValue)

	// result := hex.EncodeToString(sig.Sum(nil))
	result = base64.StdEncoding.EncodeToString(sig.Sum(nil))

	return result, err
}

func Minify(body []byte) ([]byte, error) {
	buff := new(bytes.Buffer)
	errCompact := json.Compact(buff, body)
	if errCompact != nil {
		newErr := fmt.Errorf("failure encountered compacting json := %v", errCompact)
		return nil, newErr
	}
	b, err := ioutil.ReadAll(buff)
	if err != nil {
		readErr := fmt.Errorf("read buffer error encountered := %v", err)
		return nil, readErr
	}
	return b, nil
}

func EscapeHTML(v interface{}, safeEncoding bool) ([]byte, error) {
	b, err := json.Marshal(v)

	if safeEncoding {
		b = bytes.Replace(b, []byte("\\u003c"), []byte("<"), -1)
		b = bytes.Replace(b, []byte("\\u003e"), []byte(">"), -1)
		b = bytes.Replace(b, []byte("\\u0026"), []byte("&"), -1)
	}
	return b, err
}

func GetStringInBetween(str string, start string, end string) (result string) {
	s := strings.Index(str, start)
	if s == -1 {
		return ""
	}
	s += len(start)
	e := strings.Index(str[s:], end)
	if e == -1 {
		return ""
	}
	e += s
	return str[s:e]
}

func TimeBetween(start, end, check time.Time) bool {
	if start.Before(end) {
		return !check.Before(start) && !check.After(end)
	}
	if start.Equal(end) {
		return check.Equal(start)
	}
	return !start.After(check) || !end.Before(check)
}

func FindTrxRes(input string) string {
	re := regexp.MustCompile(`\b(akan diproses)\b`)
	match := re.FindString(input)
	if match != "" {
		return match
	}
	return ""
}

func FindPendingByCheckStatus(input string) string {
	re := regexp.MustCompile(`\b(status Menunggu Jawaban)\b`)
	match := re.FindString(input)
	if match != "" {
		return match
	}
	return ""
}

func TimeStampNow() string {
	return time.Now().Format(constants.LAYOUT_TIMESTAMP)
}

func HourNow() string {
	return time.Now().Format(constants.LAYOUT_HOUR)
}

func ReplaceSQL(old, searchPattern string) string {
	tmpCount := strings.Count(old, searchPattern)
	for m := 1; m <= tmpCount; m++ {
		old = strings.Replace(old, searchPattern, "$"+strconv.Itoa(m), 1)
	}
	return old
}

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

func Stringify(input interface{}) string {
	bytes, err := json.Marshal(input)
	if err != nil {
		panic(err)
	}
	strings := string(bytes)
	bytes, err = json.Marshal(strings)
	if err != nil {
		panic(err)
	}

	return string(bytes)
}

func JSONPrettyfy(data interface{}) {
	bytesData, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(bytesData))
}

func JSONPrettyfyV2(data interface{}) string {
	bytesData, _ := json.MarshalIndent(data, "", "  ")
	return string(bytesData)
}

func ToString(i interface{}) string {
	log, _ := json.Marshal(i)
	logString := string(log)

	return logString
}

func QueryFill(query string) (new string) {
	query = strings.ReplaceAll(query, " ", "")
	split := strings.Split(query, ",")
	for range split {
		new += "?,"
	}

	return strings.TrimSuffix(new, ",")
}

func GenerateToken(merchantKey string) (resToken string, err error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["merchantKey"] = merchantKey
	claims["exp"] = time.Now().Add(time.Hour * 4).Unix()

	resToken, err = token.SignedString([]byte(config.GetEnv("JWT_KEY")))
	if err != nil {
		return
	}
	return
}

func BindValidateStruct(ctx echo.Context, i interface{}, function string) error {
	if err := ctx.Bind(i); err != nil {
		return err
	}
	bytes, _ := json.Marshal(i)
	log.Println("Incoming Request on", function, "=>", string(bytes))
	if err := ctx.Validate(i); err != nil {
		return err
	}

	return nil
}

func NewErrors(message string) error {
	return errors.New(message)
}

func DateString(year, month, day string) time.Time {
	y, _ := strconv.Atoi(year)
	m, _ := strconv.Atoi(month)
	d, _ := strconv.Atoi(day)

	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
}

// COUNT BETWEEN
func GetDaysBetween(dateFrom, dateTo string) int {
	if len(dateFrom) != 10 {
		from := strings.Split(dateFrom, " ")
		dateFrom = from[0]
	}

	if len(dateTo) != 10 {
		to := strings.Split(dateTo, " ")
		dateTo = to[0]
	}

	from := strings.Split(dateFrom, "-")
	to := strings.Split(dateTo, "-")

	if len(from) != 3 || len(to) != 3 {
		log.Println("format data tidak sesuai")
	}

	t1 := DateString(from[0], from[1], from[2])
	t2 := DateString(to[0], to[1], to[2])

	days := t2.Sub(t1).Hours() / 24

	return int(days)
}

func ValidateDays(listDays string) (err error) {

	days := strings.Split(listDays, "|")
	names := strings.Split(constants.DAYS_NAMES, "|")

	for _, x := range days {
		for j, y := range names {
			if x == y {
				break
			} else if j == len(names)-1 {
				return NewErrors("Invalid days name " + x)
			}
		}
	}
	return

}

func ListDates(dateFrom, dateTo string) (listDates []string) {
	listDates = append(listDates, dateFrom)
	totalDay := GetDaysBetween(dateFrom, dateTo)
	x, _ := time.Parse(constants.LAYOUT_DATE, dateFrom)
	for i := 0; i < totalDay; i++ {
		date := x.AddDate(0, 0, i+1).Format(constants.LAYOUT_DATE)
		listDates = append(listDates, date)
	}
	return
}

func FilterDatesByDayName(listDate []string, listDayName string) (dates []string) {
	dayNames := strings.Split(listDayName, "|")
	for i := 0; i < len(listDate); i++ {
		x, _ := time.Parse(constants.LAYOUT_DATE, listDate[i])
		days := x.Weekday().String()
		days = strings.ToUpper(days)
		for _, dayList := range dayNames {
			if days == dayList {
				dates = append(dates, listDate[i])
			}
		}
	}
	return
}

func GetDateInt(dates string) (int, int, int) {
	t, _ := time.Parse(constants.LAYOUT_DATE, dates)
	return t.Year(), int(t.Month()), t.Day()
}

func GetHourInt(hour string) (int, int) {
	t, _ := time.Parse(constants.LAYOUT_HOUR, hour)
	return t.Hour(), t.Minute()
}

func GetDepartureTime(previousDuration float64, date string, hour string) (departureDate string, departureHour string) {
	a, b, c := GetDateInt(date)
	d, e := GetHourInt(hour)
	f := time.Date(a, time.Month(b), c, d, e, 0, 0, time.Local)
	departure := f.Add(time.Minute * time.Duration(previousDuration))
	departureHour = departure.Format(constants.LAYOUT_HOUR)
	departureDate = departure.Format(constants.LAYOUT_DATE)
	return
}

func GetArrivalTime(departuredate, departurehour string, travelDuration float64) (arrivalDate string, arrivalHour string) {
	a, b, c := GetDateInt(departuredate)
	d, e := GetHourInt(departurehour)
	f := time.Date(a, time.Month(b), c, d, e, 0, 0, time.Local)
	arrival := f.Add(time.Minute * time.Duration(travelDuration))
	arrivalHour = arrival.Format(constants.LAYOUT_HOUR)
	arrivalDate = arrival.Format(constants.LAYOUT_DATE)
	return
}

func RoundUpToHundred(num float64) (result float64) {
	result = float64(((int(num) + 99) / 100) * 100)
	if result == constants.EMPTY_VALUE_INT {
		result = 100
	}

	return
}

func RoundUpToThousand(num float64) (result float64) {
	result = float64(((int(num) + 999) / 1000) * 1000)
	if result == constants.EMPTY_VALUE_INT {
		result = 1000
	}

	return
}

func RoundUpToFifthThousand(num float64) (result float64) {
	numInt := int(num)
	result = float64(((numInt + 4999) / 5000) * 5000)
	if result == constants.EMPTY_VALUE_INT {
		result = 5000
	}
	return
}

func GetTokenData(ctx echo.Context) models.TokenDetail {
	token := ctx.Get(constants.USER)
	if token == nil {
		result := models.TokenDetail{}
		return result
	}
	user := ctx.Get(constants.USER).(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	result := models.TokenDetail{
		PersonName: claims["personName"].(string),
		PersonID:   claims["personID"].(string),
		CID:        claims["cid"].(string),
	}

	return result
}

func GetUsernameToken(ctx echo.Context) string {
	user := ctx.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)

	return claims["username"].(string)
}

func GenerateRandomNumber(n int) (string, error) {
	seed := time.Now().Unix()
	r := rand.New(rand.NewSource(seed))
	charsets := []rune("1234567890")
	letters := make([]rune, n)
	for i := range letters {
		letters[i] = charsets[r.Intn(len(charsets))]
	}
	return string(letters), nil
}

func GenerateRandomString(n int) (string, error) {
	seed := time.Now().UnixNano()
	r := rand.New(rand.NewSource(seed))
	charsets := []rune("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNPQRSTUVWXYZ") //
	letters := make([]rune, n)
	for i := range letters {
		letters[i] = charsets[r.Intn(len(charsets))]
	}
	return string(letters), nil
}

func GenerateCombinedID(n, j int64) (int64, error) {
	combinedID := n*1000000 + j

	if combinedID > 999999999999 {
		return 0, fmt.Errorf("combined_id exceeds maximum 6 digits")
	}

	return combinedID, nil
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Make hash
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 4)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// Hash PIN Request
func HashPin(pin string) string {
	hasher := sha256.New()

	hasher.Write([]byte(pin))

	hashedPinBytes := hasher.Sum(nil)

	hashedPINHex := hex.EncodeToString(hashedPinBytes)

	return hashedPINHex
}

func FloatToString(amount float64) string {
	amountString := strconv.FormatFloat(amount, 'f', -1, 64)

	return amountString
}

func StringToInt(input string) int {
	intVar, err := strconv.Atoi(input)
	if err != nil {
		return 0
	}

	return intVar
}

func FormatCurrency(amount float64) string {
	intAmount := int64(amount)

	formatted := strconv.FormatInt(intAmount, 10)

	var result string
	for i, v := range formatted {
		if (len(formatted)-i)%3 == 0 && i != 0 {
			result += "."
		}
		result += string(v)
	}

	return result
}

func IntToString(n int64) string {
	result := strconv.FormatInt(n, 10)

	return result
}

func ValidatePIN(pin string) error {
	_, err := strconv.Atoi(pin)
	if err != nil {
		return errors.New("PIN harus berupa angka")
	}
	return nil
}

func CapitalInFront(n string) string {
	caser := cases.Title(language.English)

	lowerCaseStr := strings.ToLower(n)

	formattedStr := caser.String(lowerCaseStr)

	return formattedStr
}

func RemoveDots(input string) string {
	return strings.Replace(input, ".", "", -1)
}

// func CutoffTime(layout string) (time.Time, time.Time) {
// 	currentTime := time.Now()

// 	today := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, currentTime.Location())
// 	cutoffStart := today.Add(time.Hour*23 + time.Minute*55)
// 	cutoffEnd := today.Add(time.Hour*23 + time.Minute*59 + time.Second*59)

// 	return cutoffStart, cutoffEnd
// }

func CutoffTime(layout string) (time.Time, time.Time) {
	currentTime := time.Now()

	today := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, currentTime.Location())
	cutoffStart := today.Add(time.Hour*8 + time.Minute*00) // Jam 5.40 pagi
	cutoffEnd := today.Add(time.Hour*20 + time.Minute*30)  // Jam 6.00 pagi

	return cutoffStart, cutoffEnd
}

func Encrypt(data string) string {
	dataBytes := []byte(data)

	encryptData := base64.URLEncoding.EncodeToString(dataBytes)
	return encryptData
}

func Decrypt(encryptedData string) (string, error) {
	decryptedDataBytes, err := base64.URLEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", err
	}

	decryptedData := string(decryptedDataBytes)
	return decryptedData, nil
}

func EncryptBase64URL(data string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	plaintext := []byte(data)
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(cryptoutils.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	encodedData := base64.URLEncoding.EncodeToString(ciphertext)
	return encodedData, nil
}

func DecryptBase64URL(encryptedData string, key []byte) (string, error) {
	ciphertext, err := base64.URLEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}

func DebugFunc(doFunc func() error) {
	if config.DebugMode == "true" {
		err := doFunc()
		if err != nil {
			log.Println(":::::::::::::::::::: WARNING :::::::::::::::::::::: debugFunc :: ", err.Error())
		}
	}
}

// func UploadResizer(uploadImgProduct string, filename string, file *multipart.FileHeader, types string, width uint, height uint) (bool, string) {

// 	source, err := file.Open()
// 	if err != nil {
// 		return false, "Cannot Open File: " + err.Error()
// 	}
// 	defer source.Close()

// 	destination, err := os.Create(filepath.Join(uploadImgProduct, filename))
// 	if err != nil {
// 		return false, "Cannot Create File: " + err.Error()
// 	}
// 	defer destination.Close()

// 	if _, err = io.Copy(destination, source); err != nil {
// 		return false, "Cannot Create Copy File: " + err.Error()
// 	}

// 	return true, "Success"
// }

func UploadResizerImage(uploadImgProduct string, filename string, file *multipart.FileHeader, width int, maxSizeMB float64, compressLevel int) (bool, string) {
	source, err := file.Open()
	if err != nil {
		return false, "Cannot Open File: " + err.Error()
	}
	defer source.Close()

	img, _, err := image.Decode(source)
	if err != nil {
		return false, "Cannot Decode Image: " + err.Error()
	}

	resizedImg, err := ResizeAndCompressImage(img, width, maxSizeMB, compressLevel)
	if err != nil {
		return false, "Error Resizing/Compressing Image: " + err.Error()
	}

	savePath := filepath.Join(uploadImgProduct, filename)
	err = imaging.Save(resizedImg, savePath)
	if err != nil {
		return false, "Cannot Save Image: " + err.Error()
	}

	return true, "Success"
}

// Compress level & resize image service
func ResizeAndCompressImage(img image.Image, width int, maxSizeMB float64, compressLevel int) (image.Image, error) {
	var (
		quality int
	)

	if compressLevel == 1 {
		quality = 90
	} else if compressLevel == 2 {
		quality = 70
	} else if compressLevel == 3 {
		quality = 50
	} else if compressLevel == 4 {
		quality = 30
	} else {
		quality = 70
	}

	resizedImg := imaging.Resize(img, width, 0, imaging.Lanczos)

	buf := new(bytes.Buffer)
	err := imaging.Encode(buf, resizedImg, imaging.JPEG, imaging.JPEGQuality(quality))
	if err != nil {
		return nil, err
	}

	fileSizeMB := float64(buf.Len()) / (1024 * 1024)
	if fileSizeMB > maxSizeMB {
		scaleFactor := math.Sqrt(maxSizeMB / fileSizeMB)
		newWidth := int(float64(width) * scaleFactor)
		resizedImg = imaging.Resize(img, newWidth, 0, imaging.Lanczos)
	}

	return resizedImg, nil
}

// compress level pdf 1 - 4
func CompressPDF(inputPath string, outputPath string, compressLevel int) error {
	// optimizeConf := pdfcpu.New

	// switch compressLevel {
	// case 1:
	// 	optimizeConf.Optimize = &pdfcpu.Stats{ImageQuality: pdfcpu.Lossless}
	// case 2:
	// 	optimizeConf.Optimize = pdfcpu.Optimize{ImageQuality: pdfcpu.High}
	// case 3:
	// 	optimizeConf.Optimize = pdfcpu.Optimize{ImageQuality: pdfcpu.Medium}
	// case 4:
	// 	optimizeConf.Optimize = pdfcpu.Optimize{ImageQuality: pdfcpu.Low}
	// default:
	// 	optimizeConf.Optimize = pdfcpu.Optimize{ImageQuality: pdfcpu.Medium}
	// }

	// err := api.OptimizeFile(inputPath, outputPath, optimizeConf)
	// if err != nil {
	// 	return err
	// }
	return nil
}

// func ResizePDFImages(inputPath string, outputPath string, maxWidth, maxHeight int) error {
// 	// Load PDF
// 	pdf := gopdf.GoPdf{}
// 	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
// 	err := pdf.AddPage()
// 	if err != nil {
// 		return err
// 	}

// 	err = pdf.ImportPages(inputPath, "1-")
// 	if err != nil {
// 		return err
// 	}

// 	pages := pdf.PageCount()
// 	for i := 1; i <= pages; i++ {
// 		pdf.AddPage()
// 		err := pdf.ImportPage(inputPath, i)
// 		if err != nil {
// 			return err
// 		}

// 		// Resize images
// 		imgs, err := pdf.ExtractImages(i)
// 		if err != nil {
// 			return err
// 		}
// 		for _, img := range imgs {
// 			pdf.ImageOptions(img.Path, img.X, img.Y, &gopdf.Rect{W: float64(maxWidth), H: float64(maxHeight)}, gopdf.ImageOptions{ImageType: gopdf.ImageTypeJPEG})
// 		}
// 	}

// 	// Save output PDF
// 	err = pdf.WritePdf(outputPath)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

func UploadResizerPDF(uploadPdfProduct string, filename string, file *multipart.FileHeader, compressLevel int) (bool, string) {
	source, err := file.Open()
	if err != nil {
		return false, "Cannot Open File: " + err.Error()
	}
	defer source.Close()

	tmpFile, err := os.CreateTemp("", "upload-*.pdf")
	if err != nil {
		return false, "Cannot Create Temp File: " + err.Error()
	}
	defer os.Remove(tmpFile.Name())

	if _, err = io.Copy(tmpFile, source); err != nil {
		return false, "Cannot Copy to Temp File: " + err.Error()
	}

	outputFilePath := filepath.Join(uploadPdfProduct, filename)
	err = CompressPDF(tmpFile.Name(), outputFilePath, compressLevel)
	if err != nil {
		return false, "Error Compressing PDF: " + err.Error()
	}

	return true, "Success"
}

// generateClientID generates a UUID for the client ID
func GenerateClientID() (string, error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	return uuid.String(), nil
}

// generateClientSecret generates a random string for the client secret
func GenerateClientSecret() (string, error) {
	const secretLength = 32
	secret := make([]byte, secretLength)
	if _, err := crRand.Read(secret); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(secret), nil
}

func IsValidJSON(s string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(s), &js) == nil
}
