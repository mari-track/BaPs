package pack

import (
	"github.com/gucooing/BaPs/common/enter"
	"github.com/gucooing/BaPs/game"
	"github.com/gucooing/BaPs/gdconf"
	"github.com/gucooing/BaPs/pkg/logger"
	"github.com/gucooing/BaPs/protocol/proto"
)

func TimeAttackDungeonLogin(s *enter.Session, request, response proto.Message) {
	// rsp := response.(*proto.TimeAttackDungeonLoginResponse)
	//
	// rsp.PreviousRoomDB = &proto.TimeAttackDungeonRoomDB{
	// 	AccountId:         s.AccountServerId,
	// 	SeasonId:          0,
	// 	RoomId:            0,
	// 	CreateDate:        time.Time{},
	// 	RewardDate:        time.Time{},
	// 	IsPractice:        false,
	// 	SweepHistoryDates: nil,
	// 	BattleHistoryDBs:  nil,
	// }
}

func BillingPurchaseListByYostar(s *enter.Session, request, response proto.Message) {
	rsp := response.(*proto.BillingPurchaseListByYostarResponse)

	rsp.CountList = make([]*proto.PurchaseCountDB, 0)
	rsp.OrderList = make([]*proto.PurchaseOrderDB, 0)
	rsp.MonthlyProductList = make([]*proto.MonthlyProductPurchaseDB, 0)
	rsp.BlockedProductDBs = make([]*proto.BlockedProductDB, 0)
}

func EventContentPermanentList(s *enter.Session, request, response proto.Message) {
	rsp := response.(*proto.EventContentPermanentListResponse)

	rsp.PermanentDBs = make([]*proto.EventContentPermanentDB, 0)
	for _, id := range []int64{900801, 900802, 900803, 900804, 900805, 900806, 900808, 900809,
		900810, 900812, 900813, 900814, 900815, 900816, 900817, 900818, 900825, 900701} {
		rsp.PermanentDBs = append(rsp.PermanentDBs, &proto.EventContentPermanentDB{
			EventContentId:            id,
			IsStageAllClear:           false,
			IsReceivedCharacterReward: false,
		})
	}
}

func StickerLogin(s *enter.Session, request, response proto.Message) {
	rsp := response.(*proto.StickerLoginResponse)

	rsp.StickerBookDB = &proto.StickerBookDB{
		AccountId:        s.AccountServerId,
		UnusedStickerDBs: make([]*proto.StickerDB, 0),
		UsedStickerDBs:   make([]*proto.StickerDB, 0),
	}
}

func EventRewardIncrease(s *enter.Session, request, response proto.Message) {
	rsp := response.(*proto.EventRewardIncreaseResponse)

	rsp.EventRewardIncreaseDBs = make([]*proto.EventRewardIncreaseDB, 0)
}

func OpenConditionEventList(s *enter.Session, request, response proto.Message) {
	req := request.(*proto.OpenConditionEventListRequest)
	rsp := response.(*proto.OpenConditionEventListResponse)

	rsp.ConquestTiles = make(map[int64][]*proto.ConquestTileDB)
	rsp.WorldRaidLocalBossDBs = make(map[int64][]*proto.WorldRaidLocalBossDB)

	for _, conqusetEventId := range req.ConquestEventIds {
		rsp.ConquestTiles[conqusetEventId] = make([]*proto.ConquestTileDB, 0)
	}
	for seasonId, worldRaidBossGroupId := range req.WorldRaidSeasonAndGroupIds {
		bossList := make([]*proto.WorldRaidLocalBossDB, 0)
		boss := &proto.WorldRaidLocalBossDB{
			SeasonId:     seasonId,
			GroupId:      worldRaidBossGroupId,
			UniqueId:     0,
			IsScenario:   false,
			IsCleardEver: false,
			TacticMscSum: 0,
			RaidBattleDB: nil,
			IsContinue:   false,
		}
		bossList = append(bossList, boss)
		rsp.WorldRaidLocalBossDBs[seasonId] = bossList
	}
}

func NotificationEventContentReddotCheck(s *enter.Session, request, response proto.Message) {
	rsp := response.(*proto.NotificationEventContentReddotResponse)

	rsp.Reddots = make(map[int64][]proto.NotificationEventReddot)
	rsp.EventContentUnlockCGDBs = make(map[int64][]*proto.EventContentCollectionDB)
}

func ContentLogUIOpenStatistics(s *enter.Session, request, response proto.Message) {

}

func ContentSweepRequest(s *enter.Session, request, response proto.Message) {
	req := request.(*proto.ContentSweepRequest)
	rsp := response.(*proto.ContentSweepResponse)

	rsp.ClearParcels = make([][]*proto.ParcelInfo, 0)
	rsp.BonusParcels = make([]*proto.ParcelInfo, 0)
	rsp.EventContentBonusParcels = make([][]*proto.ParcelInfo, 0)

	switch req.Content {
	case proto.ContentType_WeekDungeon:
		parcelResultList, clearParcels := game.ContentSweepWeekDungeon(req.StageId, req.Count)
		// 扣钱
		conf := gdconf.GetWeekDungeonExcelTable(req.StageId)
		if conf != nil && (len(conf.StageEnterCostType) == len(conf.StageEnterCostId) &&
			len(conf.StageEnterCostId) == len(conf.StageEnterCostAmount)) {
			for index, rewardType := range conf.StageEnterCostType {
				parcelType := proto.ParcelType(proto.ParcelType_value[rewardType])
				parcelResultList = append(parcelResultList, &game.ParcelResult{
					ParcelType: parcelType,
					ParcelId:   conf.StageEnterCostId[index],
					Amount:     -conf.StageEnterCostAmount[index] * req.Count,
				})
			}
		}
		rsp.ParcelResult = game.ParcelResultDB(s, parcelResultList)
		rsp.ClearParcels = clearParcels
	case proto.ContentType_SchoolDungeon:
		parcelResultList, clearParcels := game.ContentSweepSchoolDungeon(req.StageId, req.Count)
		// 扣钱
		parcelResultList = append(parcelResultList,
			game.GetSchoolDungeonCost(true, req.Count)...)
		rsp.ParcelResult = game.ParcelResultDB(s, parcelResultList)
		rsp.ClearParcels = clearParcels
	default:
		logger.Warn("未处理的扫荡类型:%s", req.Content.String())
	}
}
