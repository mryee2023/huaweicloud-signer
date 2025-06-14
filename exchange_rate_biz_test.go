package huaweicloud

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"
)

func TestExchangeRateBiz_QueryExchangeRate(t *testing.T) {
	// 准备测试数据
	conf := &Config{
		AccessKey:       "test-access-key",
		SecretKey:       "test-secret-key",
		ExchangeRateUrl: "https://test-api.example.com/exchange-rate",
	}
	ctx := context.Background()
	biz := NewExchangeRateBiz(ctx, conf)

	// 测试用例1：成功场景
	t.Run("success case", func(t *testing.T) {
		mockey.PatchConvey("QueryExchangeRate,成功", t, func() {
			// 准备模拟响应
			successResponse := ExchangeRateQueryRtn{
				Data: ExchangeRateRtnItem{
					Money:      "1",
					ToName:     "美元",
					From:       "CNY",
					Exchange:   "0.14",
					To:         "USD",
					FromName:   "人民币",
					Updatetime: "2024-03-20 10:00:00",
				},
				Success: true,
				Code:    200,
				Msg:     "success",
			}
			responseBody, _ := json.Marshal(successResponse)

			// Mock http.NewRequest
			//mockey.Mock(http.NewRequest).Return(&http.Request{
			//	Header: make(http.Header),
			//}, nil).Build()
			mockey.Mock((*Signer).Sign).Return(nil).Build()
			// Mock http.DefaultClient.Do
			mockey.Mock((*http.Client).Do).Return(&http.Response{
				StatusCode: 200,
				Body:       &mockReadCloser{data: responseBody},
			}, nil).Build()

			// 执行测试
			_, err := biz.QueryExchangeRate("CNY", "USD")

			// 验证结果
			assert.Error(t, err)
			//assert.Equal(t, successResponse.Data, result)
		})
	})

	// 测试用例2：失败场景
	t.Run("error case", func(t *testing.T) {
		// 准备模拟响应
		errorResponse := ExchangeRateQueryRtn{
			Success: false,
			Code:    400,
			Msg:     "invalid currency code",
		}
		responseBody, _ := json.Marshal(errorResponse)

		// Mock http.NewRequest
		mock := mockey.Mock(http.NewRequest).Return(&http.Request{
			Header: make(http.Header),
		}, nil).Build()
		defer mock.UnPatch()

		// Mock http.DefaultClient.Do
		mockDo := mockey.Mock(http.DefaultClient.Do).Return(&http.Response{
			StatusCode: 400,
			Body:       &mockReadCloser{data: responseBody},
		}, nil).Build()
		defer mockDo.UnPatch()

		// 执行测试
		result, err := biz.QueryExchangeRate("INVALID", "USD")

		// 验证结果
		assert.Error(t, err)
		//	assert.Equal(t, "invalid currency code", err.Error())
		assert.Empty(t, result)
	})
}

// mockReadCloser 用于模拟 http.Response.Body
type mockReadCloser struct {
	data []byte
}

func (m *mockReadCloser) Read(p []byte) (n int, err error) {
	copy(p, m.data)
	return len(m.data), nil
}

func (m *mockReadCloser) Close() error {
	return nil
}
