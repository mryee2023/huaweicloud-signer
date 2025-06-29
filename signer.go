// HWS API Gateway Signature
// based on https://github.com/datastream/aws/blob/master/signv4.go
// Copyright (c) 2014, Xianjie

package huaweicloud

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"

	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"
)

const (
	DateFormat           = "20060102T150405Z"
	SignAlgorithm        = "SDK-HMAC-SHA256"
	HeaderXDateTime      = "X-Sdk-Date"
	HeaderXHost          = "host"
	HeaderXAuthorization = "Authorization"
	HeaderXContentSha256 = "X-Sdk-Content-Sha256"
)

func hmacsha256(keyByte []byte, dataStr string) ([]byte, error) {
	hm := hmac.New(sha256.New, []byte(keyByte))
	if _, err := hm.Write([]byte(dataStr)); err != nil {
		return nil, err
	}
	return hm.Sum(nil), nil
}

// CanonicalRequest Build a CanonicalRequest from a regular request string
func CanonicalRequest(request *http.Request, signedHeaders []string) (string, error) {
	var hexencode string
	var err error
	if hex := request.Header.Get(HeaderXContentSha256); hex != "" {
		hexencode = hex
	} else {
		bodyData, err := RequestPayload(request)
		if err != nil {
			return "", err
		}
		hexencode, err = HexEncodeSHA256Hash(bodyData)
		if err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s", request.Method, CanonicalURI(request), CanonicalQueryString(request), CanonicalHeaders(request, signedHeaders), strings.Join(signedHeaders, ";"), hexencode), err
}

// CanonicalURI returns request uri
func CanonicalURI(request *http.Request) string {
	pattens := strings.Split(request.URL.Path, "/")
	var uriSlice []string
	for _, v := range pattens {
		uriSlice = append(uriSlice, escape(v))
	}
	urlpath := strings.Join(uriSlice, "/")
	if len(urlpath) == 0 || urlpath[len(urlpath)-1] != '/' {
		urlpath = urlpath + "/"
	}
	return urlpath
}

func CanonicalQueryString(request *http.Request) string {
	var keys []string
	queryMap := request.URL.Query()
	for key := range queryMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var query []string
	for _, key := range keys {
		k := escape(key)
		sort.Strings(queryMap[key])
		for _, v := range queryMap[key] {
			kv := fmt.Sprintf("%s=%s", k, escape(v))
			query = append(query, kv)
		}
	}
	queryStr := strings.Join(query, "&")
	request.URL.RawQuery = queryStr
	return queryStr
}

func CanonicalHeaders(request *http.Request, signerHeaders []string) string {
	var canonicalHeaders []string
	header := make(map[string][]string)
	for k, v := range request.Header {
		header[strings.ToLower(k)] = v
	}
	for _, key := range signerHeaders {
		value := header[key]
		if strings.EqualFold(key, HeaderXHost) {
			value = []string{request.Host}
		}
		sort.Strings(value)
		for _, v := range value {
			canonicalHeaders = append(canonicalHeaders, key+":"+strings.TrimSpace(v))
		}
	}
	return fmt.Sprintf("%s\n", strings.Join(canonicalHeaders, "\n"))
}

func SignedHeaders(r *http.Request) []string {
	var signedHeaders []string
	for key := range r.Header {
		signedHeaders = append(signedHeaders, strings.ToLower(key))
	}
	sort.Strings(signedHeaders)
	return signedHeaders
}

func RequestPayload(request *http.Request) ([]byte, error) {
	if request.Body == nil {
		return []byte(""), nil
	}
	bodyByte, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return []byte(""), err
	}
	request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyByte))
	return bodyByte, err
}

// StringToSign Create a "String to Sign".
func StringToSign(canonicalRequest string, t time.Time) (string, error) {
	hashStruct := sha256.New()
	_, err := hashStruct.Write([]byte(canonicalRequest))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s\n%s\n%x",
		SignAlgorithm, t.UTC().Format(DateFormat), hashStruct.Sum(nil)), nil
}

// SignStringToSign Create the HWS Signature.
func SignStringToSign(stringToSign string, signingKey []byte) (string, error) {
	hmsha, err := hmacsha256(signingKey, stringToSign)
	return fmt.Sprintf("%x", hmsha), err
}

// HexEncodeSHA256Hash returns hexcode of sha256
func HexEncodeSHA256Hash(body []byte) (string, error) {
	hashStruct := sha256.New()
	if len(body) == 0 {
		body = []byte("")
	}
	_, err := hashStruct.Write(body)
	return fmt.Sprintf("%x", hashStruct.Sum(nil)), err
}

// AuthHeaderValue Get the finalized value for the "Authorization" header. The signature parameter is the output from SignStringToSign
func AuthHeaderValue(signatureStr, accessKeyStr string, signedHeaders []string) string {
	return fmt.Sprintf("%s Access=%s, SignedHeaders=%s, Signature=%s", SignAlgorithm, accessKeyStr, strings.Join(signedHeaders, ";"), signatureStr)
}

// Signature HWS meta
type Signer struct {
	Key    string
	Secret string
}

// SignRequest set Authorization header
func (s *Signer) Sign(request *http.Request) error {
	var t time.Time
	var err error
	var date string
	if date = request.Header.Get(HeaderXDateTime); date != "" {
		t, err = time.Parse(DateFormat, date)
	}
	if err != nil || date == "" {
		t = time.Now()
		request.Header.Set(HeaderXDateTime, t.UTC().Format(DateFormat))
	}
	signedHeaders := SignedHeaders(request)
	canonicalRequest, err := CanonicalRequest(request, signedHeaders)
	if err != nil {
		return err
	}
	stringToSignStr, err := StringToSign(canonicalRequest, t)
	if err != nil {
		return err
	}
	signatureStr, err := SignStringToSign(stringToSignStr, []byte(s.Secret))
	if err != nil {
		return err
	}
	authValueStr := AuthHeaderValue(signatureStr, s.Key, signedHeaders)
	request.Header.Set(HeaderXAuthorization, authValueStr)
	return nil
}
