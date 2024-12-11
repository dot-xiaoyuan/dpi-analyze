package license

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"strings"
	"time"
)

const SecretKey = "5e1575e9c3bb70837f2c1ed5764ce71d"

// CheckLicense 校验许可是否合法
func CheckLicense() error {
	// 判断是否为空
	if len(config.Cfg.License.Sn) == 0 {
		return errors.New("系统未授权，请获取授权后使用")
	}

	if len(config.Cfg.License.Sn) < 32 {
		return errors.New("授权许可证无效，请联系研发获取授权")
	}

	// 如果校验过期时间大于当前时间 就直接返回 ok
	if config.Cfg.License.CheckTime.Unix() > 0 && config.Cfg.License.CheckTime.Before(time.Now()) {
		return nil
	}

	// 去掉分隔符的授权码
	LicenseSn := strings.ReplaceAll(config.Cfg.License.Sn, "-", "")

	// 计算md5值 规则："机器码:项目名称:到期时间:盐值"
	var Md5Builder bytes.Buffer
	Md5Builder.WriteString(config.Cfg.MachineID)
	Md5Builder.WriteString(":srun-dpi:")
	Md5Builder.WriteString(LicenseSn[26:])
	Md5Builder.WriteString(":")
	Md5Builder.WriteString(SecretKey)
	SnMd5 := fmt.Sprintf("%x", md5.Sum(Md5Builder.Bytes()))
	//fmt.Println(Md5Builder.String())
	// 如果长度不是 32 或者 取前三十二位做验证码
	if SnMd5[:26] != LicenseSn[:26] {
		return errors.New("授权许可证无效，请先获取授权")
	}

	// 校验时间是否超时
	licenseTimeStr := fmt.Sprintf("20%s-%s-%s 23:59:59", LicenseSn[26:28], LicenseSn[28:30], LicenseSn[30:32])

	licenseTime, err := time.Parse("2006-01-02 15:04:05", licenseTimeStr)

	if err != nil {
		return errors.New("授权许可证时间校验无效，请确认许可证是否合法")
	}

	// 如果授权过期了 就提示
	if licenseTime.Before(time.Now()) {
		return errors.New("授权许可证已过期，请获取新许可证")
	}

	// 每 2小时 验证一次
	config.Cfg.License.ExpireTime = licenseTime
	config.Cfg.License.CheckTime = time.Now().Add(time.Hour * 2)

	// 校验通过
	return nil
}
