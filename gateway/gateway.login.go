package gateway

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gucooing/BaPs/common/enter"
	"github.com/gucooing/BaPs/db"
	"github.com/gucooing/BaPs/game"
	"github.com/gucooing/BaPs/pkg/alg"
	"github.com/gucooing/BaPs/pkg/logger"
	"github.com/gucooing/BaPs/pkg/mx"
	"github.com/gucooing/BaPs/protocol/proto"
	pb "google.golang.org/protobuf/proto"
)

func (g *Gateway) getEnterTicket(c *gin.Context) {
	if !alg.CheckGateWay(c) {
		errTokenBestHTTP(c)
		return
	}
	bin, err := mx.GetFormMx(c)
	if err != nil {
		return
	}
	rsp := &proto.QueuingGetTicketResponse{}
	defer g.send(c, rsp)
	req := new(proto.QueuingGetTicketRequest)
	err = json.Unmarshal(bin, req)
	if err != nil {
		logger.Debug("request err:%s c--->s:%s", err.Error(), string(bin))
		return
	}
	yoStarUserLogin := db.GetYoStarUserLoginByYostarUid(req.YostarUID)
	if yoStarUserLogin == nil {
		return
	}
	if yoStarUserLogin.YostarLoginToken != req.YostarToken ||
		yoStarUserLogin.YostarLoginToken == "" {
		return
	}
	yoStarUserLogin.YostarLoginToken = ""
	if err = db.UpdateYoStarUserLogin(yoStarUserLogin); err != nil {
		return
	}
	enterTicket := fmt.Sprintf("%v%s", alg.GetSnow().GenId(), alg.RandStr(10))
	if !enter.AddEnterTicket(yoStarUserLogin.AccountServerId, req.YostarUID, enterTicket) {
		return
	}
	rsp.EnterTicket = enterTicket
	rsp.SetSessionKey(&proto.BasePacket{
		Protocol: req.Protocol,
	})
	logger.Debug("EnterTicket交换成功:%s", rsp.EnterTicket)
}

func (g *Gateway) AccountCheckYostar(s *enter.Session, request, response proto.Message) {
	req := request.(*proto.AccountCheckYostarRequest)
	rsp := response.(*proto.AccountCheckYostarResponse)
	var err error

	tickInfo := enter.GetEnterTicketInfo(req.EnterTicket)
	if tickInfo == nil {
		rsp.ResultMessag = "EnterTicket验证失败"
		logger.Debug("EnterTicket验证失败")
		return
	}
	if enter.GetSessionNum() >= enter.MaxPlayerNum &&
		enter.MaxPlayerNum > 0 {
		rsp.ResultMessag = "在线玩家满"
		logger.Debug("在线玩家满")
		return
	}
	enter.DelEnterTicket(req.EnterTicket)
	s = enter.GetSessionByAccountServerId(tickInfo.AccountServerId)
	mxToken := mx.GetMxToken(tickInfo.AccountServerId, 64)
	if s == nil {
		yostarGame := db.GetYostarGameByAccountServerId(tickInfo.AccountServerId)
		if yostarGame == nil {
			// new Game Player
			yostarGame, err = db.AddYostarGameByYostarUid(tickInfo.AccountServerId)
			if err != nil {
				logger.Debug("账号创建失败:%s", err.Error())
				return
			}
		}
		s = enter.NewSession(tickInfo.AccountServerId)
		s.YostarUID = tickInfo.YostarUID
		if yostarGame.BinData != nil {
			pb.Unmarshal(yostarGame.BinData, s.PlayerBin)
		} else {
			s.PlayerBin = game.NewYostarGame(tickInfo.AccountServerId)
			logger.Debug("AccountServerId:%v,新玩家登录Game,创建新账号中", tickInfo.AccountServerId)
		}
	}
	// 更新一次账号缓存
	s.MxToken = mxToken
	s.EndTime = time.Now().Add(time.Duration(enter.MaxCachePlayerTime) * time.Minute)
	if !enter.AddSession(s) {
		logger.Info("AccountServerId:%v,重复上线账号,如果老客户端在线则会被离线", tickInfo.AccountServerId)
	} else {
		logger.Info("AccountServerId:%v,上线账号", tickInfo.AccountServerId)
	}
	rsp.ResultState = 1
	base := &proto.BasePacket{
		SessionKey: &proto.SessionKey{
			AccountServerId: tickInfo.AccountServerId,
			MxToken:         s.MxToken,
		},
		Protocol:           response.GetProtocol(),
		AccountId:          tickInfo.AccountServerId,
		ServerNotification: game.GetServerNotification(s),
		ServerTimeTicks:    game.GetServerTime(),
	}
	response.SetSessionKey(base)
}
