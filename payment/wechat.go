package payment

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/miaomimg3/photo/config"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
)

type WechatPay struct {
	AppID     string
	MchID     string
	APIKey    string
	NotifyURL string
	client    *core.Client
}

type Order struct {
	OrderID   string
	QRCodeURL string
	Amount    float64
	Status    string
	CreatedAt time.Time
}

// NewWechatPay 创建微信支付实例
func NewWechatPay() *WechatPay {
	cfg := config.Load()
	// 加载商户私钥
	privateKey, err := utils.LoadPrivateKeyWithPath("pem/apiclient_key.pem")
	if err != nil {
		log.Fatalf("加载商户私钥失败: %v", err)
	}

	// 创建微信支付客户端
	ctx := context.Background()
	client, err := core.NewClient(ctx,
		option.WithWechatPayAutoAuthCipher(cfg.WechatPay.MchID, "46DB018B089C83CECD451A1A28487003881B5221", privateKey, cfg.WechatPay.APIKey),
	)
	if err != nil {
		log.Fatalf("创建支付客户端失败: %v", err)
	}

	return &WechatPay{
		AppID:     cfg.WechatPay.AppID,
		MchID:     cfg.WechatPay.MchID,
		APIKey:    cfg.WechatPay.APIKey,
		NotifyURL: cfg.WechatPay.NotifyURL,
		client:    client,
	}
}

// CreateOrder 创建支付订单
func (w *WechatPay) CreateOrder(userID uint, amount float64, description string) (*Order, error) {
	// 生成订单号
	orderID := fmt.Sprintf("PHOTO%d%d", userID, time.Now().Unix())

	// 创建Native支付服务
	svc := native.NativeApiService{Client: w.client}
	resp, _, err := svc.Prepay(context.Background(), native.PrepayRequest{
		Appid:       core.String(w.AppID),
		Mchid:       core.String(w.MchID),
		Description: core.String(description),
		OutTradeNo:  core.String(orderID),
		NotifyUrl:   core.String(w.NotifyURL),
		Amount: &native.Amount{
			Total: core.Int64(int64(amount * 100)), // 金额单位：分
		},
	})

	if err != nil {
		return nil, fmt.Errorf("创建预支付订单失败: %v", err)
	}

	// 记录订单信息（实际应该保存到数据库）
	order := &Order{
		OrderID:   orderID,
		QRCodeURL: *resp.CodeUrl,
		Amount:    amount,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	return order, nil
}

// HandleNotify 处理支付回调
// HandleNotify 处理支付回调
func (w *WechatPay) HandleNotify(data []byte) (bool, error) {
	// 解析回调数据
	var notifyReq map[string]interface{}
	if err := json.Unmarshal(data, &notifyReq); err != nil {
		return false, fmt.Errorf("解析回调数据失败: %v", err)
	}

	// 验证签名
	// 实际实现需使用微信支付SDK的签名验证功能

	// 更新订单状态
	orderID := notifyReq["out_trade_no"].(string)
	// 这里应该更新数据库中的订单状态

	// 如果支付成功，创建照片记录
	if notifyReq["trade_state"].(string) == "SUCCESS" {
		// 实现照片创建逻辑
	}

	return true, nil
}

// generateSign 生成签名（示例）
func (w *WechatPay) generateSign(params map[string]string) string {
	// 实现微信支付签名算法
	// ...
	hash := md5.Sum([]byte(signStr))
	return hex.EncodeToString(hash[:])
}
