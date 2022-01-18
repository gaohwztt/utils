package cos

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
	sts "github.com/tencentyun/qcloud-cos-sts-sdk/go"
)

// 获取临时密钥的参数
type TmpRequest struct {
	Appid     string // 用户 Appid
	SecretId  string // 用户的 SecretId
	SecretKey string // 用户的 SecretKey

	Bucket string // cos 的 bucket
	Region string // cos 的 region

	Action          []string // (选填) 访问权限
	Effect          bool     // (选填) true:allow  false: deny
	DurationSeconds int64    // (选填) 指定临时证书的有效期 单位:秒 默认3600秒, 最长可设置有效期为7200秒 区域之外的设置成默认时间
	Resource        []string // (选填) 不填写默认("qcs::cos:region:uid/appid:bucket/*") 授权操作的具体数据，可以是任意资源、指定路径前缀的资源、指定绝对路径的资源或它们的组合
}

// 返回临时密钥数据
type TmpResponse struct {
	TmpSecretID  string // 返回临时密钥 SecretId
	TmpSecretKey string // 返回临时密钥 SecretKey
	SessionToken string // 返回密钥 token
	StartTime    uint64 // 密钥生成时间
	ExpiredTime  uint64 // 密钥过期时间
}

// 获取临时密钥
func GetTmpData(tmpReq TmpRequest) (TmpResponse, error) {
	var tmpRsp TmpResponse

	// DurationSeconds
	if tmpReq.DurationSeconds <= 0 || tmpReq.DurationSeconds > durationSecondsMax {
		tmpReq.DurationSeconds = durationSecondsDefault
	}
	// Effect
	effect := effectFalse
	if tmpReq.Effect {
		effect = effectTrue
	}
	// Resource
	if len(tmpReq.Resource) == 0 {
		tmpReq.Resource = append(tmpReq.Resource, fmt.Sprintf(resourceDefault, tmpReq.Region, tmpReq.Appid, tmpReq.Bucket))
	}
	// Action
	if len(tmpReq.Action) == 0 {
		tmpReq.Action = append(tmpReq.Action, actionDefault()...)
	}

	c := sts.NewClient(tmpReq.SecretId, tmpReq.SecretKey, nil)
	opt := &sts.CredentialOptions{
		DurationSeconds: tmpReq.DurationSeconds,
		Region:          tmpReq.Region,
		Policy: &sts.CredentialPolicy{
			Statement: []sts.CredentialPolicyStatement{
				{
					Action:   tmpReq.Action,
					Effect:   effect,
					Resource: tmpReq.Resource,
				},
			},
		},
	}
	res, err := c.GetCredential(opt)
	if err != nil {
		return tmpRsp, err
	}

	tmpRsp.TmpSecretID = res.Credentials.TmpSecretID
	tmpRsp.TmpSecretKey = res.Credentials.TmpSecretKey
	tmpRsp.SessionToken = res.Credentials.SessionToken
	tmpRsp.StartTime = uint64(time.Now().Local().Unix())
	tmpRsp.ExpiredTime = uint64(tmpReq.DurationSeconds)

	return tmpRsp, nil
}

// 生成cos 的client
type ClientRequest struct {
	SecretID     string
	SecretKey    string
	SessionToken string // sts的 token, 上面两个变成sts的id和key
	SecretUrl    string // 访问域名
	EnableCRC    bool   // 是否关闭 CRC64 校验
	IsSts        bool   // 是否是 sts
}

func GetCosClient(req ClientRequest) *cos.Client {
	u, _ := url.Parse(req.SecretUrl)
	b := &cos.BaseURL{BucketURL: u}
	// 使用临时密钥生成 client
	transport := &cos.AuthorizationTransport{
		SecretID:  req.SecretID,  // 替换为用户的 SecretId，请登录访问管理控制台进行查看和管理，https://console.cloud.tencent.com/cam/capi
		SecretKey: req.SecretKey, // 替换为用户的 SecretKey，请登录访问管理控制台进行查看和管理，https://console.cloud.tencent.com/cam/capi
	}
	if req.IsSts {
		// 如果使用临时密钥需要填入，临时密钥生成和使用指引参见https://cloud.tencent.com/document/product/436/14048
		transport.SessionToken = req.SessionToken
	}

	client := cos.NewClient(b, &http.Client{
		Transport: transport,
	})

	client.Conf.EnableCRC = req.EnableCRC
	return client
}

// 直接上传图片
func UploadSimpleImg(client *cos.Client, file *multipart.FileHeader, secretUrl, fileName string) (string, error) {
	if msg := isImg(file); msg != "" {
		return "", errors.New(msg)
	}

	f, err := file.Open()
	if err != nil {
		return "", errors.New("file open fail")
	}
	defer f.Close()
	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: "text/html",
		},
		ACLHeaderOptions: &cos.ACLHeaderOptions{
			// 如果不是必要操作，建议上传文件时不要给单个文件设置权限，避免达到限制。若不设置默认继承桶的权限。
			XCosACL: "private",
		},
	}

	data, err := client.Object.Put(context.Background(), fileName, f, opt)

	if err != nil {
		return "", err
	}
	if data.StatusCode != http.StatusOK {
		return "", errors.New("cos.statusCode != http.statusOk")
	}

	return secretUrl + fileName, nil
}
