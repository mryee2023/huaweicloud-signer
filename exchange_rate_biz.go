package huaweicloud

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
)

type ExchangeRateRtnItem struct {
	Money      string `json:"money"`
	ToName     string `json:"to_name"`
	From       string `json:"from"`
	Exchange   string `json:"exchange"`
	To         string `json:"to"`
	FromName   string `json:"from_name"`
	Updatetime string `json:"updatetime"`
}
type ExchangeRateQueryRtn struct {
	Data    ExchangeRateRtnItem `json:"data"`
	Msg     string              `json:"msg"`
	Success bool                `json:"success"`
	Code    int                 `json:"code"`
	TaskNo  string              `json:"taskNo"`
}

type Config struct {
	AccessKey       string `json:"accessKey"`
	SecretKey       string `json:"secretKey"`
	ExchangeRateUrl string `json:"exchangeRateUrl"`
}
type ExchangeRateBiz struct {
	ctx    context.Context
	logger logx.Logger
	conf   *Config
}

func NewExchangeRateBiz(ctx context.Context, conf *Config) *ExchangeRateBiz {
	return &ExchangeRateBiz{
		ctx:    ctx,
		logger: logx.WithContext(ctx).WithFields(logx.Field("module", "huaweicloud.ExchangeRateBiz")),
		conf:   conf,
	}
}
func (b *ExchangeRateBiz) QueryExchangeRate(fromCode, toCode string) (ExchangeRateRtnItem, error) {
	logger := logx.WithContext(b.ctx)
	var rtnItem ExchangeRateRtnItem
	postData := url.Values{
		"fromCode": {fromCode},
		"toCode":   {toCode},
		"money":    {"1"},
	}

	defer func() {
		logger.Infow("queryExchangeRate")
	}()
	logger.WithFields(logx.Field("postData", postData.Encode()))

	//https://jmtyhlcxv2.apistore.huaweicloud.com/exchange-rate-v2/convert
	var covertHost = b.conf.ExchangeRateUrl
	r, err := http.NewRequest("POST", covertHost,
		strings.NewReader(postData.Encode()))

	if err != nil {
		logger.WithFields(logx.Field("error.NewRequest", err.Error()))
		return rtnItem, err
	}

	s := Signer{
		Key:    b.conf.AccessKey,
		Secret: b.conf.SecretKey,
	}

	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	err = s.Sign(r)
	if err != nil {
		logger.WithFields(logx.Field("error.sign", err.Error()))

		return rtnItem, err
	}

	client := http.DefaultClient
	resp, err := client.Do(r)
	if err != nil {

		logger.WithFields(logx.Field("error.client.do", err.Error()))
		return rtnItem, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.WithFields(logx.Field("error.readAll", err.Error()))
		return rtnItem, err
	}
	logger.WithFields(logx.Field("queryRtn", string(body)))
	var queryRtn ExchangeRateQueryRtn
	err = json.Unmarshal(body, &queryRtn)
	if err != nil {
		logger.WithFields(logx.Field("error.Unmarshal", err.Error()))
		return rtnItem, err
	}
	if !queryRtn.Success {
		return rtnItem, errors.New(queryRtn.Msg)
	}
	return queryRtn.Data, nil
}
