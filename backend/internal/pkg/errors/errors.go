package errors

// HTTP状态码
const (
	Success       = 200
	BadRequest    = 400
	Unauthorized  = 401
	Forbidden     = 403
	NotFound      = 404
	InternalError = 500
)

// 业务错误码 10001-10999
const (
	ErrUserNotFound      = 10001
	ErrPasswordWrong     = 10002
	ErrUserDisabled      = 10003
	ErrDuplicateKey      = 10004
	ErrInsufficientStock = 10005
	ErrOrderNotFound     = 10006
	ErrInvalidOrderStatus = 10007
	ErrInvalidParam       = 10008
)

// 错误码消息映射
var CodeMessage = map[int]string{
	Success:              "success",
	BadRequest:           "请求参数错误",
	Unauthorized:         "未授权",
	Forbidden:            "禁止访问",
	NotFound:             "资源不存在",
	InternalError:        "服务器内部错误",
	ErrUserNotFound:      "用户不存在",
	ErrPasswordWrong:     "密码错误",
	ErrUserDisabled:      "用户已被禁用",
	ErrDuplicateKey:      "数据重复",
	ErrInsufficientStock: "库存不足",
	ErrOrderNotFound:     "订单不存在",
	ErrInvalidOrderStatus: "订单状态无效",
}

// GetMessage 获取错误码对应的消息
func GetMessage(code int) string {
	if msg, ok := CodeMessage[code]; ok {
		return msg
	}
	return "未知错误"
}
