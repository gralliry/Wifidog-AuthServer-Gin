package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/antchfx/htmlquery"
	"io"
	"math/big"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

const (
	hexE = "010001"
	hexM = "008e9fdac2a933c27a8262eb0ab8004aa74571e1e7c27beb436ce17c37df778d8861a9a2afddc04a6e80da995e34754e1e002864f2480f0471257880b55359e8232601244593333eb9f0f99b894fe13538a80bfd14aeb94bb8108959140231195a9e9f488f7d5cc72a112d6a19576cb05eaf629435538907ccc9b008d64595646d"
)

func strToHex(s string) string {
	encoded := ""
	for _, c := range s {
		encoded = fmt.Sprintf("%X", c) + encoded
	}
	return encoded
}

func encryptPassword(b string, e string, m string) (string, error) {
	// 将 16 进制字符串转换为大数，base 参数为 16
	bp := new(big.Int)
	be := new(big.Int)
	bm := new(big.Int)
	var success bool
	_, success = bp.SetString(b, 16)
	if !success {
		return "", errors.New("bp转换出错")
	}
	_, success = be.SetString(e, 16)
	if !success {
		return "", errors.New("be转换出错")
	}
	_, success = bm.SetString(m, 16)
	if !success {
		return "", errors.New("bm转换出错")
	}
	// 进行模幂运算: bp ^ be % bm
	modPow := new(big.Int).Exp(bp, be, bm)
	// 将大数转换为 16 进制字符串
	return modPow.Text(16), nil
}

func Verify(account string, password string) (bool, error) {
	var req *http.Request
	var resp *http.Response
	var err error
	// 创建一个 Cookie Jar，用于存储和管理 cookies
	jar, err := cookiejar.New(nil)
	if err != nil {
		return false, err
	}
	// 将 Cookie Jar 关联到 HTTP 客户端
	client := &http.Client{
		Jar: jar,
	}
	// -----------------------------------------------------------------------------------------------------------------
	req, err = http.NewRequest("GET", "https://cas.scau.edu.cn/lyuapServer/login", nil)
	if err != nil {
		return false, err
	}
	// 发送请求
	resp, err = client.Do(req)
	if err != nil {
		return false, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(resp.Body)
	// 解析 HTML
	doc, err := htmlquery.Parse(resp.Body)
	if err != nil {
		return false, err
	}
	// 使用 XPath 选择器查找元素
	node, err := htmlquery.Query(doc, "/html/body/div/div[3]/div[2]/div/div[2]/div[2]/form/div[5]/input[1]")
	if err != nil {
		return false, err
	}
	formLt := htmlquery.SelectAttr(node, "value")
	// -----------------------------------------------------------------------------------------------------------------
	// 创建一个不带数据的 POST 请求
	req, err = http.NewRequest("POST", "https://cas.scau.edu.cn/lyuapServer/LoginLogCount", nil)
	if err != nil {
		return false, err
	}
	// 发送请求
	resp, err = client.Do(req)
	if err != nil {
		return false, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(resp.Body)
	// -----------------------------------------------------------------------------------------------------------------
	// 对密码进行编码
	encodePwd, err := encryptPassword(url.QueryEscape(strToHex(password)), hexE, hexM)
	if err != nil {
		return false, err
	}
	// form表单
	form := url.Values{
		"username":  {account},
		"password":  {encodePwd},
		"captcha":   {""},
		"warn":      {"true"},
		"lt":        {formLt},
		"execution": {"e1s1"},
		"_eventId":  {"submit"},
		"submit":    {"登  录"},
	}
	// 创建 POST 请求
	req, err = http.NewRequest("POST", "https://cas.scau.edu.cn/lyuapServer/login", bytes.NewBufferString(form.Encode()))
	if err != nil {
		return false, err
	}
	// 设置请求头
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err = client.Do(req)
	if err != nil {
		return false, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(resp.Body)
	// 解析 HTML
	doc, err = htmlquery.Parse(resp.Body)
	if err != nil {
		return false, err
	}
	// 使用 XPath 选择器查找元素
	node, err = htmlquery.Query(doc, "/html/body/div/div[1]/div[1]/div/div/div[1]/input[1]")
	if err != nil {
		return false, err
	}
	return node != nil, nil
}
