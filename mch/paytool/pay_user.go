package paytool

import (
	"errors"
	"fmt"
	"time"

	"github.com/astaxie/beego"
	"github.com/qnfnypen/gowechat/mch/base"
	"github.com/qnfnypen/gowechat/util"
)

//官方文档： https://pay.weixin.qq.com/wiki/doc/api/tools/mch_pay.php?chapter=14_2

//PayInput 付款的配置
type PayInput struct {
	ToOpenID string //接红包的OpenID
	MoneyFen int    //分为单位
	Remark   string //备注 String(100)

	IP string
}

//Pay 付款
func (c *PayTool) Pay(input PayInput) (isSuccess bool, err error) {
	now := time.Now()
	dayStr := beego.Date(now, "Ymd")

	billno := c.MchID + dayStr + util.RandomStr(10)

	var signMap = make(map[string]string)
	signMap["nonce_str"] = util.RandomStr(5)
	signMap["partner_trade_no"] = billno //mch_id+yyyymmdd+10位一天内不能重复的数字
	signMap["mchid"] = c.MchID
	signMap["mch_appid"] = c.AppID
	signMap["check_name"] = "NO_CHECK"
	signMap["openid"] = input.ToOpenID
	signMap["amount"] = util.ToStr(input.MoneyFen)
	signMap["spbill_create_ip"] = input.IP
	signMap["desc"] = input.Remark
	signMap["sign"] = base.Sign(signMap, c.MchAPIKey, nil)

	respMap, err := c.PayRaw(signMap)
	if err != nil {
		return false, err
	}

	resultCode, ok := respMap["result_code"]
	if !ok {
		err = errors.New("no result_code")
		return false, err
	}

	if resultCode != "SUCCESS" {
		returnMsg, _ := respMap["return_msg"]
		errMsg, _ := respMap["err_code_des"]
		errCode, _ := respMap["err_code"]

		if errCode == "NOTENOUGH" {
			return false, ErrNoEnoughMoney
		}

		err = fmt.Errorf("Err:%s return_msg:%s err_code:%s err_code_des:%s", "result code is not success", returnMsg, errCode, errMsg)
		return false, err
	}

	mchBillNo, ok := respMap["mch_billno"]
	if !ok {
		err = errors.New("no mch_billno")
		return false, err
	}

	if billno != mchBillNo {
		err = errors.New("billno is not correct")
		return false, err
	}

	return true, nil
}
