package pack

import (
	"time"

	"github.com/gucooing/BaPs/common/enter"
	"github.com/gucooing/BaPs/common/rank"
	"github.com/gucooing/BaPs/game"
	"github.com/gucooing/BaPs/gdconf"
	"github.com/gucooing/BaPs/pkg/logger"
	"github.com/gucooing/BaPs/pkg/mx"
	"github.com/gucooing/BaPs/protocol/proto"
)

func RaidLogin(s *enter.Session, request, response mx.Message) {
	rsp := response.(*proto.RaidLoginResponse)

	rsp.SeasonType = game.GetRaidSeasonType()
}

func RaidLobby(s *enter.Session, request, response mx.Message) {
	rsp := response.(*proto.RaidLobbyResponse)

	curBattle := game.GetCurRaidBattleInfo(s)
	// 超时了
	if curBattle != nil &&
		!curBattle.IsClose &&
		time.Now().After(time.Unix(curBattle.Begin, 0).Add(1*time.Hour)) {
		parcelResult := game.RaidClose(s)
		rsp.RaidGiveUpDB = game.GetRaidGiveUpDB(s)
		rsp.ParcelResultDB = game.ParcelResultDB(s, parcelResult)
	}
	rsp.SeasonType = game.GetRaidSeasonType()
	rsp.RaidLobbyInfoDB = &proto.SingleRaidLobbyInfoDB{
		ClearDifficulty: game.GetClearDifficulty(s),
		RaidLobbyInfoDB: game.GetRaidLobbyInfoDB(s),
	}
}

func RaidOpponentList(s *enter.Session, request, response mx.Message) {
	req := request.(*proto.RaidOpponentListRequest)
	rsp := response.(*proto.RaidOpponentListResponse)

	rsp.OpponentUserDBs = make([]*proto.SingleRaidUserDB, 0)
	cur := gdconf.GetCurRaidSchedule()
	if cur == nil {
		return
	}
	for i := int64(0); i < 15; i++ {
		ranking := req.Rank + i
		uid, _ := rank.GetUidByRank(cur.SeasonId, ranking)
		as := enter.GetSessionByUid(uid)
		if as != nil {
			rsp.OpponentUserDBs = append(rsp.OpponentUserDBs, game.GetSingleRaidUserDB(as))
		}
	}
}

func RaidGetBestTeam(s *enter.Session, request, response mx.Message) {
	req := request.(*proto.RaidGetBestTeamRequest)
	rsp := response.(*proto.RaidGetBestTeamResponse)

	rsp.RaidTeamSettingDBs = make([]*proto.RaidTeamSettingDB, 0)
	as := enter.GetSessionByUid(req.AccountId)
	if as == nil {
		return
	}
	for _, bin := range game.GetCurRaidTeamList(s) {
		rsp.RaidTeamSettingDBs = append(rsp.RaidTeamSettingDBs, game.GetRaidTeamSettingDB(as, bin))
	}
}

func RaidCreateBattle(s *enter.Session, request, response mx.Message) {
	req := request.(*proto.RaidCreateBattleRequest)
	rsp := response.(*proto.RaidCreateBattleResponse)

	defer func() {
		rsp.AccountCurrencyDB = game.GetAccountCurrencyDB(s)
	}()

	if game.GetRaidSeasonType() != proto.RaidSeasonType_Open {
		// 没开就请求,nt了
		return
	}
	game.NewCurRaidBattleInfo(s, req.RaidUniqueId, req.IsPractice)

	curBattle := game.GetCurRaidBattleInfo(s)
	if curBattle == nil {
		logger.Debug("总力战实例创建失败")
		return
	}
	if assist := req.AssistUseInfo; assist != nil && !curBattle.IsAssist {
		ac := enter.GetSessionByUid(assist.CharacterAccountId)
		assistInfo := game.GetAssistInfo(ac, assist.EchelonType, assist.SlotNumber)
		rsp.AssistCharacterDB = game.GetAssistCharacterDB(ac, assistInfo, assist.AssistRelation)
	}

	if !req.IsPractice {
		// 扣票
		game.UpCurrency(s, proto.CurrencyTypes_RaidTicket, -1)
	}
	rsp.RaidBattleDB = game.GetRaidBattleDB(s)
	rsp.RaidDB = game.GetRaidDB(s)
}

func RaidEndBattle(s *enter.Session, request, response mx.Message) {
	req := request.(*proto.RaidEndBattleRequest)
	rsp := response.(*proto.RaidEndBattleResponse)

	curBattle := game.GetCurRaidBattleInfo(s)
	summary := req.Summary
	if summary == nil || curBattle == nil ||
		summary.RaidSummary == nil {
		return
	}
	raidSummary := summary.RaidSummary
	echelonInfo := game.GetEchelonInfo(s, proto.EchelonType_Raid, int64(req.EchelonId))
	// 参战角色保存
	if !game.CheckRaidCharacter(s, echelonInfo, summary) {
		return
	}
	// 记录boss情况
	for _, raidBossResult := range raidSummary.RaidBossResults {
		curBattle.AiPhase = raidBossResult.AIPhase
		curBattle.BossGroggyPoint += raidBossResult.RaidDamage.GivenGroggyPoint
		curBattle.GivenDamage += raidBossResult.RaidDamage.GivenDamage
		curBattle.IndexDamage = raidBossResult.RaidDamage.Index
	}
	curBattle.Frame += summary.EndFrame
	curBattle.ServerId++
	curBattle.IsClose = curBattle.MaxHp-curBattle.GivenDamage == 0
	// 判断是否结算
	if curBattle.IsClose {
		// 结算
		parcelResult := game.RaidClose(s)
		rsp.ClearTimePoint = curBattle.ClearTimePoint
		rsp.HPPercentScorePoint = curBattle.HpScorePoint
		rsp.DefaultClearPoint = curBattle.DefaultPoint
		rsp.RankingPoint = curBattle.ClearTimePoint + curBattle.HpScorePoint + curBattle.DefaultPoint
		rsp.BestRankingPoint = game.GetCurRaidInfo(s).GetBestScore()
		rsp.ParcelResultDB = game.ParcelResultDB(s, parcelResult)
	}
}

func RaidEnterBattle(s *enter.Session, request, response mx.Message) {
	req := request.(*proto.RaidEnterBattleRequest)
	rsp := response.(*proto.RaidEnterBattleResponse)

	curBattle := game.GetCurRaidBattleInfo(s)
	if curBattle == nil || // 没有战斗
		curBattle.RaidUniqueId != req.RaidUniqueId || // 实例不对
		game.GetRaidSeasonType() != proto.RaidSeasonType_Open || // 没开启
		time.Now().After(time.Unix(curBattle.Begin, 0).Add(1*time.Hour)) { // 超时了
		return
	}

	defer func() {
		rsp.AccountCurrencyDB = game.GetAccountCurrencyDB(s)
	}()

	if assist := req.AssistUseInfo; assist != nil && !curBattle.IsAssist {
		ac := enter.GetSessionByUid(assist.CharacterAccountId)
		assistInfo := game.GetAssistInfo(ac, assist.EchelonType, assist.SlotNumber)
		rsp.AssistCharacterDB = game.GetAssistCharacterDB(ac, assistInfo, assist.AssistRelation)
	}

	rsp.RaidBattleDB = game.GetRaidBattleDB(s)
	rsp.RaidDB = game.GetRaidDB(s)
}

func RaidGiveUp(s *enter.Session, request, response mx.Message) {
	req := request.(*proto.RaidGiveUpRequest)
	rsp := response.(*proto.RaidGiveUpResponse)

	curBattle := game.GetCurRaidBattleInfo(s)
	if curBattle == nil ||
		req.IsPractice != curBattle.IsPractice {
		return
	}
	parcelResult := game.RaidClose(s)
	if !curBattle.IsPractice {
		rsp.RaidGiveUpDB = game.GetRaidGiveUpDB(s)
		rsp.ParcelResultDB = game.ParcelResultDB(s, parcelResult)
	}
}
