package huaweicloud

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
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

// ExchangeRateBiz 汇率查询业务
type ExchangeRateBiz struct {
	ctx    context.Context
	client *http.Client
	logger *zap.Logger
	conf   *Config
}

// NewExchangeRateBiz 创建汇率查询业务
func NewExchangeRateBiz(ctx context.Context, conf *Config) *ExchangeRateBiz {
	return &ExchangeRateBiz{
		ctx:    ctx,
		client: &http.Client{Timeout: 10 * time.Second},
		logger: zap.L().With(zap.String("module", "huaweicloud.ExchangeRateBiz")),
		conf:   conf,
	}
}

// QueryExchangeRate 查询汇率
func (b *ExchangeRateBiz) QueryExchangeRate(fromCode, toCode string) (ExchangeRateRtnItem, error) {
	logger := b.logger.With(
		zap.String("fromCode", fromCode),
		zap.String("toCode", toCode),
	)
	defer func() {
		logger.Info("汇率查询")
	}()

	var rtnItem ExchangeRateRtnItem
	postData := url.Values{
		"fromCode": {fromCode},
		"toCode":   {toCode},
		"money":    {"1"},
	}

	logger = logger.With(zap.String("postData", postData.Encode()))

	//https://jmtyhlcxv2.apistore.huaweicloud.com/exchange-rate-v2/convert
	var covertHost = b.conf.ExchangeRateUrl
	req, err := http.NewRequestWithContext(b.ctx, "POST", covertHost,
		strings.NewReader(postData.Encode()))

	if err != nil {
		logger.Error("failed to create request", zap.Error(err))
		return rtnItem, err
	}

	s := Signer{
		Key:    b.conf.AccessKey,
		Secret: b.conf.SecretKey,
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	err = s.Sign(req)
	if err != nil {
		logger.Error("failed to sign request", zap.Error(err))
		return rtnItem, err
	}

	resp, err := b.client.Do(req)
	if err != nil {
		logger.Error("failed to send request", zap.Error(err))
		return rtnItem, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("failed to read response", zap.Error(err))
		return rtnItem, err
	}
	logger = logger.With(zap.String("response", string(body)))
	var queryRtn ExchangeRateQueryRtn
	err = json.Unmarshal(body, &queryRtn)
	if err != nil {
		logger.Error("failed to unmarshal response", zap.Error(err))
		return rtnItem, err
	}
	if !queryRtn.Success {
		return rtnItem, errors.New(queryRtn.Msg)
	}
	return queryRtn.Data, nil
}
