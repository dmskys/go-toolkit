package web

import "net/http"

// WebContext 作为一个web请求的上下文信息
type WebContext struct {
	Response *Response
	Request  *Request
}

// WebHandler 控制器方法
type WebHandler func(context *WebContext) HTTPResponse

// NewWebHandler 创建一个WebHandler，用于传递给Router
func NewWebHandler(handler WebHandler, decors ...HandlerDecorator) WebHandler {
	for i := range decors {
		d := decors[len(decors)-i-1]
		handler = d(handler)
	}

	return handler
}

// ServeHTTP 实现http.HandlerFunc接口
func (h WebHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	context := &WebContext{
		Response: &Response{
			w:       w,
			headers: make(map[string]string),
		},
		Request: &Request{r: r},
	}

	resp := h(context)
	if resp != nil {
		resp.CreateResponse()
	}
}

// NewJSONResponse 创建一个JSONResponse对象
func (ctx *WebContext) NewJSONResponse(res interface{}) JSONResponse {
	return NewJSONResponse(ctx.Response, res)
}

// NewAPIResponse 创建一个API响应
func (ctx *WebContext) NewAPIResponse(code string, message string, data interface{}) JSONResponse {
	return ctx.NewJSONResponse(struct {
		Code    string      `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// NewHTMLResponse 创建一个HTML响应
func (ctx *WebContext) NewHTMLResponse(res string) HTMLResponse {
	return NewHTMLResponse(ctx.Response, res)
}

// Redirect 页面跳转
func (ctx *WebContext) Redirect(location string, code int) RedirectResponse {
	return NewRedirectResponse(ctx.Response, ctx.Request, location, code)
}
