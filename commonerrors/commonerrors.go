package commonerrors

import "errors"

var ErrRepoNotFound = errors.New("数据没找到")

var ErrServiceBusy = errors.New("服务繁忙，请稍后再试")

var ErrSystemError = errors.New("系统错误，请稍后重试")
