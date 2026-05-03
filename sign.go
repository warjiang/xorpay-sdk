package xorpay

import "github.com/warjiang/xorpay-sdk/internal/sign"

func Sign(parts ...string) string {
	return sign.Sign(parts...)
}

func SignPay(name, payType, price, orderID, notifyURL, appSecret string) string {
	return sign.Sign(name, payType, price, orderID, notifyURL, appSecret)
}

func SignBarcodePay(name, payType, price, orderID, notifyURL, barcode, appSecret string) string {
	return sign.Sign(name, payType, price, orderID, notifyURL, barcode, appSecret)
}

func SignQueryByOrderID(orderID, appSecret string) string {
	return sign.Sign(orderID, appSecret)
}

func SignRefund(price, appSecret string) string {
	return sign.Sign(price, appSecret)
}

func SignNotify(aoid, orderID, payPrice, payTime, appSecret string) string {
	return sign.Sign(aoid, orderID, payPrice, payTime, appSecret)
}
