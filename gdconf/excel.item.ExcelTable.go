package gdconf

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	sro "github.com/gucooing/BaPs/common/server_only"
	"github.com/gucooing/BaPs/pkg/logger"
)

func (g *GameConfig) loadItemExcelTable() {
	g.GetExcel().ItemExcelTable = make([]*sro.ItemExcelTable, 0)
	name := "ItemExcelTable.json"
	file, err := os.ReadFile(g.excelPath + name)
	if err != nil {
		logger.Error("文件:%s 读取失败,err:%s", name, err)
		return
	}
	if err := json.Unmarshal(file, &g.GetExcel().ItemExcelTable); err != nil {
		logger.Error("文件:%s 解析失败,err:%s", name, err)
		return
	}
	logger.Info("文件:%s 读取成功,解析数量:%v", name, len(g.GetExcel().ItemExcelTable))
}

type ItemExcel struct {
	RecruitCoin          *RecruitCoin
	recruitCoinSync      sync.Mutex
	ItemExcelMap         map[int64]*sro.ItemExcelTable
	ItemExcelCategoryMap map[string][]*sro.ItemExcelTable
}

type RecruitCoin struct {
	Item   *sro.ItemExcelTable
	EnTime time.Time
}

func (g *GameConfig) gppItemExcelTable() {
	g.GetGPP().ItemExcel = &ItemExcel{
		ItemExcelMap:         make(map[int64]*sro.ItemExcelTable),
		ItemExcelCategoryMap: make(map[string][]*sro.ItemExcelTable),
		recruitCoinSync:      sync.Mutex{},
	}

	for _, v := range g.GetExcel().GetItemExcelTable() {
		if v.ExpirationDateTime != "" {
			enTime, err := time.Parse("2006-01-02 15:04:05", v.ExpirationDateTime)
			if err != nil {
				continue
			}
			if time.Now().After(enTime) {
				continue
			}
		}
		g.GetGPP().ItemExcel.ItemExcelMap[v.Id] = v
		if g.GetGPP().ItemExcel.ItemExcelCategoryMap[v.ItemCategory_] == nil {
			g.GetGPP().ItemExcel.ItemExcelCategoryMap[v.ItemCategory_] = make([]*sro.ItemExcelTable, 0)
		}
		g.GetGPP().ItemExcel.ItemExcelCategoryMap[v.ItemCategory_] = append(
			g.GetGPP().ItemExcel.ItemExcelCategoryMap[v.ItemCategory_], v)
	}

	logger.Info("处理道具配置完成,道具:%v个,类型:%v个", len(g.GetGPP().ItemExcel.ItemExcelMap),
		len(g.GetGPP().ItemExcel.ItemExcelCategoryMap))
}

func GetRecruitCoin() *sro.ItemExcelTable {
	bin := GC.GetGPP().ItemExcel
	bin.recruitCoinSync.Lock()
	defer bin.recruitCoinSync.Unlock()
	if bin.RecruitCoin != nil &&
		!time.Now().After(bin.RecruitCoin.EnTime) {
		return bin.RecruitCoin.Item
	}
	confList := GC.GetGPP().ItemExcel.ItemExcelCategoryMap["RecruitCoin"]
	for _, conf := range confList {
		if conf.ExpirationDateTime == "2099-12-31 23:59:59" {
			continue
		}
		enTime, err := time.Parse("2006-01-02 15:04:05", conf.ExpirationDateTime)
		if err != nil {
			continue
		}
		if !time.Now().After(enTime) {
			bin.RecruitCoin = &RecruitCoin{
				Item:   conf,
				EnTime: enTime,
			}
			return conf
		}

	}
	return nil
}

func GetItemExcelCategoryMap(itemCategory string) []*sro.ItemExcelTable {
	return GC.GetGPP().ItemExcel.ItemExcelCategoryMap[itemCategory]
}

func GetItemExcelTable(id int64) *sro.ItemExcelTable {
	return GC.GetGPP().ItemExcel.ItemExcelMap[id]
}
