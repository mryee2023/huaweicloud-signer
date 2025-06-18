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

	"github.com/sirupsen/logrus"
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
	conf   *Config
}

func init() {
	// 只显示 Info 级别以上的日志
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{})

}

// NewExchangeRateBiz 创建汇率查询业务
func NewExchangeRateBiz(ctx context.Context, conf *Config) *ExchangeRateBiz {
	return &ExchangeRateBiz{
		ctx:    ctx,
		client: &http.Client{Timeout: 10 * time.Second},
		conf:   conf,
	}
}

// QueryExchangeRate 查询汇率
func (b *ExchangeRateBiz) QueryExchangeRate(fromCode, toCode string) (ExchangeRateRtnItem, error) {
	// 创建一个新的日志条目，用于收集所有字段
	entry := logrus.WithFields(logrus.Fields{
		"fromCode": fromCode,
		"toCode":   toCode,
	})
	defer func() {
		if err := recover(); err != nil {

		}
		entry.Info("查询汇率")
	}()
	var rtnItem ExchangeRateRtnItem
	postData := url.Values{
		"fromCode": {fromCode},
		"toCode":   {toCode},
		"money":    {"1"},
	}

	entry = entry.WithField("postData", postData.Encode())

	var covertHost = b.conf.ExchangeRateUrl
	req, err := http.NewRequestWithContext(b.ctx, "POST", covertHost,
		strings.NewReader(postData.Encode()))

	if err != nil {
		entry.WithError(err).Error("创建请求失败")
		return rtnItem, err
	}

	s := Signer{
		Key:    b.conf.AccessKey,
		Secret: b.conf.SecretKey,
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	err = s.Sign(req)
	if err != nil {
		entry.WithError(err).Error("签名失败")
		return rtnItem, err
	}

	resp, err := b.client.Do(req)
	if err != nil {
		entry.WithError(err).Error("发送请求失败")
		return rtnItem, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		entry.WithError(err).Error("读取响应失败")
		return rtnItem, err
	}

	entry = entry.WithField("response", string(body))

	var queryRtn ExchangeRateQueryRtn
	err = json.Unmarshal(body, &queryRtn)
	if err != nil {
		entry.WithError(err).Error("解析响应失败")
		return rtnItem, err
	}
	if !queryRtn.Success {
		entry = entry.WithFields(logrus.Fields{
			"code": queryRtn.Code,
			"msg":  queryRtn.Msg,
		})
		return rtnItem, errors.New(queryRtn.Msg)
	}

	return queryRtn.Data, nil
}
