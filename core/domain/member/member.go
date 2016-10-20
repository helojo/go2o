/**
 * Copyright 2014 @ z3q.net.
 * name :
 * author : jarryliu
 * date : 2013-12-09 10:12
 * description :
 * history :
 */

package member

//todo: 要注意UpdateTime的更新

import (
	"bytes"
	"errors"
	"fmt"
	"go2o/core/domain/interface/member"
	"go2o/core/domain/interface/mss"
	"go2o/core/domain/interface/mss/notify"
	"go2o/core/domain/interface/valueobject"
	"go2o/core/infrastructure/domain"
	"go2o/core/infrastructure/tool/sms"
	"regexp"
	"strconv"
	"strings"
	"time"
)

//todo: 依赖商户的 MSS 发送通知消息,应去掉
//todo: 会员升级 应单独来处理
var _ member.IMember = new(memberImpl)

type memberImpl struct {
	manager         member.IMemberManager
	value           *member.Member
	account         member.IAccount
	level           *member.Level
	rep             member.IMemberRep
	relation        *member.Relation
	invitation      member.IInvitationManager
	mssRep          mss.IMssRep
	valRep          valueobject.IValueRep
	profileManager  member.IProfileManager
	favoriteManager member.IFavoriteManager
	giftCardManager member.IGiftCardManager
}

func NewMember(manager member.IMemberManager, val *member.Member, rep member.IMemberRep,
	mp mss.IMssRep, valRep valueobject.IValueRep) member.IMember {
	return &memberImpl{
		manager: manager,
		value:   val,
		rep:     rep,
		mssRep:  mp,
		valRep:  valRep,
	}
}

// 获取聚合根编号
func (m *memberImpl) GetAggregateRootId() int {
	return m.value.Id
}

// 会员资料服务
func (m *memberImpl) Profile() member.IProfileManager {
	if m.profileManager == nil {
		m.profileManager = newProfileManagerImpl(m,
			m.GetAggregateRootId(), m.rep, m.valRep)
	}
	return m.profileManager
}

// 会员收藏服务
func (m *memberImpl) Favorite() member.IFavoriteManager {
	if m.favoriteManager == nil {
		m.favoriteManager = newFavoriteManagerImpl(
			m.GetAggregateRootId(), m.rep)
	}
	return m.favoriteManager
}

// 礼品卡服务
func (m *memberImpl) GiftCard() member.IGiftCardManager {
	if m.giftCardManager == nil {
		m.giftCardManager = newGiftCardManagerImpl(
			m.GetAggregateRootId(), m.rep)
	}
	return m.giftCardManager
}

// 邀请管理
func (m *memberImpl) Invitation() member.IInvitationManager {
	if m.invitation == nil {
		m.invitation = &invitationManager{
			member: m,
		}
	}
	return m.invitation
}

// 获取值
func (m *memberImpl) GetValue() member.Member {
	return *m.value
}

var (
	userRegex  = regexp.MustCompile("^[a-zA-Z0-9_]{6,}$")
	emailRegex = regexp.MustCompile("\\w+([-+.']\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*")
	phoneRegex = regexp.MustCompile("^(13[0-9]|14[5|7]|15[0-9]|16[8]|" +
		"18[0-9]|17[0|1|2|3|4|6|7|8])(\\d{8})$")
)

// 验证用户名
func validUsr(usr string) error {
	usr = strings.ToLower(strings.TrimSpace(usr)) // 小写并删除空格
	if len([]rune(usr)) < 6 {
		return member.ErrUsrLength
	}
	if !userRegex.MatchString(usr) {
		return member.ErrUsrValidErr
	}
	return nil
}

// 设置值
func (m *memberImpl) SetValue(v *member.Member) error {
	v.Usr = m.value.Usr
	if len(m.value.InvitationCode) == 0 {
		m.value.InvitationCode = v.InvitationCode
	}
	if v.Exp != 0 {
		m.value.Exp = v.Exp
	}
	if v.Level > 0 {
		m.value.Level = v.Level
	}
	if len(v.TradePwd) == 0 {
		m.value.TradePwd = v.TradePwd
	}
	return nil
}

// 发送验证码,并返回验证码
func (m *memberImpl) SendCheckCode(operation string, mssType int) (string, error) {
	const expiresMinutes = 10 //10分钟生效
	code := domain.NewCheckCode()
	m.value.CheckCode = code
	m.value.CheckExpires = time.Now().Add(time.Minute * expiresMinutes).Unix()
	_, err := m.Save()
	if err == nil {
		mgr := m.mssRep.NotifyManager()
		pro := m.Profile().GetProfile()

		// 创建参数
		data := map[string]interface{}{
			"code":      code,
			"operation": operation,
			"minutes":   expiresMinutes,
		}

		// 根据消息类型发送信息
		switch mssType {
		case notify.TypePhoneMessage:
			// 某些短信平台要求传入模板ID,在这里附加参数
			provider, _ := m.valRep.GetDefaultSmsApiPerm()
			data = sms.AppendCheckPhoneParams(provider, data)

			// 构造并发送短信
			n := mgr.GetNotifyItem("验证手机")
			c := notify.PhoneMessage(n.Content)
			err = mgr.SendPhoneMessage(pro.Phone, c, data)

		default:
		case notify.TypeEmailMessage:
			n := mgr.GetNotifyItem("验证邮箱")
			c := &notify.MailMessage{
				Subject: operation + "验证码",
				Body:    n.Content,
			}
			err = mgr.SendEmail(pro.Phone, c, data)
		}
	}
	return code, err
}

// 对比验证码
func (m *memberImpl) CompareCode(code string) error {
	if m.value.CheckCode != strings.TrimSpace(code) {
		return member.ErrCheckCodeError
	}
	if m.value.CheckExpires < time.Now().Unix() {
		return member.ErrCheckCodeExpires
	}
	return nil
}

// 获取账户
func (m *memberImpl) GetAccount() member.IAccount {
	if m.account == nil {
		v := m.rep.GetAccount(m.value.Id)
		return NewAccount(m, v, m.rep, m.manager, m.valRep)
	}
	return m.account
}

// 增加经验值
func (m *memberImpl) AddExp(exp int) error {
	m.value.Exp += exp
	_, err := m.Save()
	//判断是否升级
	m.checkLevelUp()

	return err
}

// 更改会员等级
func (m *memberImpl) ChangeLevel(level int) error {
	lg := m.manager.LevelManager()
	lv := lg.GetLevelById(level)
	// 判断等级是否启用
	if lv == nil || lv.Enabled == 0 {
		return member.ErrLevelDisabled
	}
	m.value.Exp = lv.RequireExp
	m.value.Level = level
	_, err := m.Save()
	m.level = nil
	return err
}

// 获取等级
func (m *memberImpl) GetLevel() *member.Level {
	if m.level == nil {
		m.level = m.manager.LevelManager().
			GetLevelById(m.value.Level)
	}
	return m.level
}

// 检查升级
func (m *memberImpl) checkLevelUp() bool {
	lg := m.manager.LevelManager()
	levelId := lg.GetLevelIdByExp(m.value.Exp)
	if levelId == 0 {
		return false
	}
	// 判断是否大于当前等级
	if m.value.Level > levelId {
		return false
	}
	// 判断等级是否启用
	lv := lg.GetLevelById(levelId)
	if lv.Enabled == 0 {
		return false
	}
	m.value.Level = levelId
	m.Save()
	m.level = nil
	return true
}

// 获取会员关联
func (m *memberImpl) GetRelation() *member.Relation {
	if m.relation == nil {
		m.relation = m.rep.GetRelation(m.value.Id)
	}
	return m.relation
}

// 保存关系
func (m *memberImpl) SaveRelation(r *member.Relation) error {
	m.relation = r
	m.relation.MemberId = m.value.Id
	m.updateReferStr(m.relation)
	return m.rep.SaveRelation(m.relation)
}

// 更换用户名
func (m *memberImpl) ChangeUsr(usr string) error {
	if usr == m.value.Usr {
		return member.ErrSameUsr
	}
	if len([]rune(usr)) < 6 {
		return member.ErrUsrLength
	}
	if !userRegex.MatchString(usr) {
		return member.ErrUsrValidErr
	}
	if m.usrIsExist(usr) {
		return member.ErrUsrExist
	}
	m.value.Usr = usr
	_, err := m.Save()
	return err
}

// 更新登录时间
func (m *memberImpl) UpdateLoginTime() error {
	unix := time.Now().Unix()
	m.value.LastLoginTime = m.value.LoginTime
	m.value.LoginTime = unix
	m.value.UpdateTime = unix
	_, err := m.Save()
	return err
}

// 保存
func (m *memberImpl) Save() (int, error) {
	m.value.UpdateTime = time.Now().Unix() // 更新时间，数据以更新时间触发
	if m.value.Id > 0 {
		return m.rep.SaveMember(m.value)
	}
	return m.create(m.value, nil)
}

// 锁定会员
func (m *memberImpl) Lock() error {
	m.value.State = 0
	_, err := m.Save()
	return err
}

// 解锁会员
func (m *memberImpl) Unlock() error {
	m.value.State = 1
	_, err := m.Save()
	return err
}

// 创建会员
func (m *memberImpl) create(v *member.Member, pro *member.Profile) (int, error) {
	if err := validUsr(m.value.Usr); err != nil {
		return 0, err
	}
	if m.usrIsExist(v.Usr) {
		return 0, member.ErrUsrExist
	}
	t := time.Now().Unix()
	v.State = 1
	v.RegTime = t
	v.LastLoginTime = t
	v.Level = 1
	v.Exp = 0
	v.DynamicToken = v.Pwd
	if len(v.RegFrom) == 0 {
		v.RegFrom = "API-INTERNAL"
	}
	v.InvitationCode = m.generateInvitationCode() // 创建一个邀请码
	id, err := m.rep.SaveMember(v)
	if err == nil {
		m.value.Id = id
		go m.memberInit()
	}
	return id, err
}

// 会员初始化
func (m *memberImpl) memberInit() {
	conf := m.valRep.GetRegistry()
	// 注册后赠送积分
	if conf.PresentIntegralNumOfRegister > 0 {
		m.GetAccount().AddIntegral(member.TypeIntegralPresent, "",
			conf.PresentIntegralNumOfRegister, "新会员注册赠送积分")
	}
}

// 创建邀请码
func (m *memberImpl) generateInvitationCode() string {
	var code string
	for {
		code = domain.GenerateInvitationCode()
		if memberId := m.rep.GetMemberIdByInvitationCode(code); memberId == 0 {
			break
		}
	}
	return code
}

// 用户是否已经存在
func (m *memberImpl) usrIsExist(usr string) bool {
	return m.rep.CheckUsrExist(usr, m.GetAggregateRootId())
}

// 获取推荐数组
func (m *memberImpl) getReferArr(memberId int, level int) []int {
	arr := make([]int, level)
	i := 0
	referId := memberId
	for i <= level-1 {
		rl := m.rep.GetRelation(referId)
		if rl == nil || rl.RefereesId <= 0 {
			break
		}
		arr[i] = rl.RefereesId
		referId = arr[i]
		i++
	}
	return arr
}

// 强制更新邀请关系
func (m *memberImpl) forceUpdateReferStr(r *member.Relation) {
	// 无邀请关系
	if r.RefereesId == 0 {
		r.ReferStr = ""
		return
	}
	level := m.valRep.GetRegistry().MemberReferLayer - 1
	arr := m.getReferArr(r.RefereesId, level)
	arr = append([]int{r.RefereesId}, arr...)

	if len(arr) > 0 {
		// 有邀请关系
		buf := bytes.NewBuffer([]byte("{"))
		for i, v := range arr {
			if v == 0 {
				continue
			}
			if buf.Len() > 1 {
				buf.WriteString(",")
			}
			buf.WriteString("'r")
			buf.WriteString(strconv.Itoa(i))
			buf.WriteString("':")
			buf.WriteString(strconv.Itoa(v))
		}
		buf.WriteString("}")
		r.ReferStr = buf.String()
	}
}

// 更新邀请关系
func (m *memberImpl) updateReferStr(r *member.Relation) {
	prefix := fmt.Sprintf("{'r0':%d,", r.RefereesId)
	if !strings.HasPrefix(r.ReferStr, prefix) {
		m.forceUpdateReferStr(r)
	}
}

// 更改邀请人
func (m *memberImpl) ChangeReferees(memberId int) error {
	if memberId > 0 {
		rm := m.rep.GetMember(memberId)
		if rm == nil {
			return member.ErrNoSuchMember
		}
	}
	rl := m.GetRelation()
	if rl.RefereesId != memberId {
		rl.RefereesId = memberId
		if memberId == 0 {
			rl.ReferStr = ""
		}
		return m.SaveRelation(rl)
	}
	return nil
}

var _ member.IFavoriteManager = new(favoriteManagerImpl)

// 收藏服务
type favoriteManagerImpl struct {
	_memberId int
	_rep      member.IMemberRep
}

func newFavoriteManagerImpl(memberId int,
	rep member.IMemberRep) member.IFavoriteManager {
	if memberId == 0 {
		//如果会员不存在,则不应创建服务
		panic(errors.New("member not exists"))
	}
	return &favoriteManagerImpl{
		_memberId: memberId,
		_rep:      rep,
	}
}

// 收藏
func (m *favoriteManagerImpl) Favorite(favType, referId int) error {
	if m.Favored(favType, referId) {
		return member.ErrFavored
	}
	return m._rep.Favorite(m._memberId, favType, referId)
}

// 是否已收藏
func (m *favoriteManagerImpl) Favored(favType, referId int) bool {
	return m._rep.Favored(m._memberId, favType, referId)
}

// 取消收藏
func (m *favoriteManagerImpl) Cancel(favType, referId int) error {
	return m._rep.CancelFavorite(m._memberId, favType, referId)
}

// 收藏商品
func (m *favoriteManagerImpl) FavoriteGoods(goodsId int) error {
	return m.Favorite(member.FavTypeGoods, goodsId)
}

// 取消收藏商品
func (m *favoriteManagerImpl) CancelGoodsFavorite(goodsId int) error {
	return m.Cancel(member.FavTypeGoods, goodsId)
}

// 商品是否已收藏
func (m *favoriteManagerImpl) GoodsFavored(goodsId int) bool {
	return m.Favored(member.FavTypeGoods, goodsId)
}

// 收藏店铺
func (m *favoriteManagerImpl) FavoriteShop(shopId int) error {
	return m.Favorite(member.FavTypeShop, shopId)
}

// 取消收藏店铺
func (m *favoriteManagerImpl) CancelShopFavorite(shopId int) error {
	return m.Cancel(member.FavTypeShop, shopId)
}

// 商店是否已收藏
func (m *favoriteManagerImpl) ShopFavored(shopId int) bool {
	return m.Favored(member.FavTypeShop, shopId)
}
