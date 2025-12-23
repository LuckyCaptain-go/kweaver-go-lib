package oauth

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/url"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/bytedance/sonic"
	oauthcli "golang.org/x/oauth2"
	credentmgnt "golang.org/x/oauth2/clientcredentials"

	"github.com/AISHU-Technology/kweaver-go-lib/db"
	"github.com/AISHU-Technology/kweaver-go-lib/logger"
	"github.com/AISHU-Technology/kweaver-go-lib/rest"
)

//go:generate mockgen -package mock -source ./oauth2.go -destination ./mock/mock_oauth2.go

type OAuth2 interface {
	NewOAuth2HTTPClient(info AppAccountInfo) rest.HTTPClient
	NewOAuth2HTTPClientWithOptions(info AppAccountInfo, opts rest.HttpClientOptions) rest.HTTPClient
}

type UserMgntPrivateSetting struct {
	UserMgntPrivateProcotol string
	UserMgntPrivateHost     string
	UserMgntPrivatePort     int
}

type HydraPublicSetting struct {
	HydraPublicProcotol string
	HydraPublicHost     string
	HydraPublicPort     int
}

type OAuth2Setting struct {
	UserMgntPrivateSetting
	HydraPublicSetting
	rest.HydraAdminSetting
	DBSetting db.DBSetting
}

type AppAccountInfo struct {
	AppID              string
	AppName            string
	AppSecret          string
	AppApplyDocLibPerm []string
	AppApplyDocPerm    []string
}

type AppDBInfo struct {
	AppID      string
	AppName    string
	AppSecret  string
	CreateTime int64
}

const (
	DATA_BASE_NAME     string = "dip_mdl"
	TABLE_INTERNAL_APP string = "t_internal_app"
	ACCOUNT_TYPE       string = "internal"
)

type oauth2 struct {
	setting OAuth2Setting
	client  rest.HTTPClient
}

func NewOAuth2(setting OAuth2Setting) OAuth2 {
	o := &oauth2{
		setting: setting,
		client:  rest.NewHTTPClient(),
	}

	return o
}

// NewOauthHTTPClient 初始化管理实例
func (o *oauth2) NewOAuth2HTTPClient(info AppAccountInfo) rest.HTTPClient {
	return o.NewOAuth2HTTPClientWithOptions(info, rest.HttpClientOptions{TimeOut: 300})
}

// NewOAuth2HTTPClientWithOptions 初始化管理实例
func (o *oauth2) NewOAuth2HTTPClientWithOptions(info AppAccountInfo, opts rest.HttpClientOptions) rest.HTTPClient {

	ctx := context.Background()
	info, err := o.initInternalAccount(ctx, info)
	if err != nil {
		logger.Panic(err)
	}

	conf := &credentmgnt.Config{
		ClientID:     info.AppID,
		ClientSecret: info.AppSecret,
		Scopes:       []string{"all"},
		TokenURL:     o.GetTokenUrl(),
	}

	client := rest.NewRawHTTPClientWithOptions(opts)

	ctx = context.WithValue(ctx, oauthcli.HTTPClient, client)
	rawClient := conf.Client(ctx)
	return rest.NewHTTPClientWithRawClient(rawClient)
}

func (o *oauth2) initInternalAccount(ctx context.Context, info AppAccountInfo) (AppAccountInfo, error) {
	info.AppSecret = ComputeMD5(info.AppName)

	db := db.NewDB(&o.setting.DBSetting)
	info, exist, err := o.retrieveInternalAccountByDB(db, info)
	if err != nil {
		return info, err
	}
	if !exist {
		appID, err := o.createInternalAccount(ctx, info)
		if err != nil {
			return info, err
		}

		info.AppID = appID
		err = o.saveInternalAccountToDB(db, info)
		if err != nil {
			return info, err
		}
	}

	return info, nil
}

func ComputeMD5(input string) string {
	hash := md5.Sum([]byte(input))     // 计算 MD5
	return hex.EncodeToString(hash[:]) // 转为 32 位十六进制字符串
}

func (o *oauth2) retrieveInternalAccountByDB(db *sql.DB, info AppAccountInfo) (AppAccountInfo, bool, error) {

	sqlStr, args, err := sq.Select(
		"f_app_id",
		"f_app_name",
		"f_app_secret",
		"f_create_time",
	).From(TABLE_INTERNAL_APP).
		Where(sq.Eq{"f_app_name": info.AppName}).
		ToSql()

	if err != nil {
		logger.Errorf("failed to generate select sql for table internal_app, %v", err)
		return info, false, err
	}

	sqlInfo := fmt.Sprintf("The detail sql for getting an internal account: %s", sqlStr)
	logger.Debugf(sqlInfo)

	dbInfo := AppDBInfo{}
	row := db.QueryRow(sqlStr, args...)
	err = row.Scan(
		&dbInfo.AppID,
		&dbInfo.AppName,
		&dbInfo.AppSecret,
		&dbInfo.CreateTime,
	)
	if err == sql.ErrNoRows {
		return info, false, nil
	} else if err != nil {
		logger.Errorf("failed to run QueryRow : %v", err)
		return info, false, err
	}

	if dbInfo.AppSecret != info.AppSecret {
		err = fmt.Errorf("appSecret is not match: %s : %s", dbInfo.AppSecret, info.AppSecret)
		logger.Errorf(err.Error())
		return info, false, err
	}

	info.AppID = dbInfo.AppID
	return info, true, nil
}

func (o *oauth2) saveInternalAccountToDB(db *sql.DB, info AppAccountInfo) error {

	data := map[string]interface{}{
		"f_app_id":      info.AppID,
		"f_app_name":    info.AppName,
		"f_app_secret":  info.AppSecret,
		"f_create_time": time.Now().UnixMilli(),
	}
	sqlStr, args, err := sq.Insert(TABLE_INTERNAL_APP).
		SetMap(data).
		ToSql()
	if err != nil {
		logger.Errorf("failed to generate insert sql for table internal_app, %v", err)
		return err
	}

	sqlInfo := fmt.Sprintf("The detail sql for creating an intnernal account: %s", sqlStr)
	logger.Debugf(sqlInfo)

	_, err = db.Exec(sqlStr, args...)
	if err != nil {
		logger.Errorf("failed to run Exec : %v", err)
		return err
	}

	return nil
}

func (o *oauth2) createInternalAccount(ctx context.Context, info AppAccountInfo) (string, error) {
	// 新增内部账户
	url := url.URL{
		Scheme: o.setting.UserMgntPrivateProcotol,
		Host:   fmt.Sprintf("%v:%v", o.setting.UserMgntPrivateHost, o.setting.UserMgntPrivatePort),
		Path:   "/api/user-management/v1/apps",
	}
	urlStr := url.String()

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	reqBody := map[string]string{
		"name":     info.AppName,
		"type":     ACCOUNT_TYPE,
		"password": info.AppSecret,
	}
	reqBodyByte, _ := sonic.Marshal(reqBody)

	respCode, respData, err := o.client.Post(ctx, urlStr, headers, reqBodyByte)
	if err != nil {
		logger.Errorf("Failed to get user-management response; Detail: {%v}", err)
		return "", err
	}

	if respCode != 201 {
		err = fmt.Errorf("get internal account failed, status code is %d", respCode)
		return "", err
	}

	clientID := respData.(map[string]interface{})["id"].(string)
	return clientID, nil
}

func (o *oauth2) GetTokenUrl() string {
	url := url.URL{
		Scheme: o.setting.HydraPublicProcotol,
		Host:   fmt.Sprintf("%v:%v", o.setting.HydraPublicHost, o.setting.HydraPublicPort),
		Path:   "/oauth2/token",
	}
	urlStr := url.String()
	return urlStr
}
