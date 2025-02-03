package gdconf

import (
	"encoding/json"
	"os"

	sro "github.com/gucooing/BaPs/common/server_only"
	"github.com/gucooing/BaPs/pkg/logger"
)

func (g *GameConfig) loadCharacterGearExcel() {
	g.GetExcel().CharacterGearExcel = make([]*sro.CharacterGearExcel, 0)
	name := "CharacterGearExcel.json"
	file, err := os.ReadFile(g.excelDbPath + name)
	if err != nil {
		logger.Error("文件:%s 读取失败,err:%s", name, err)
		return
	}
	if err := json.Unmarshal(file, &g.GetExcel().CharacterGearExcel); err != nil {
		logger.Error("文件:%s 解析失败,err:%s", name, err)
		return
	}
	logger.Info("文件:%s 读取成功,解析数量:%v", name, len(g.GetExcel().CharacterGearExcel))
}

type CharacterGear struct {
	CharacterGearList          map[int64]*sro.CharacterGearExcel
	CharacterGearByCharacterId map[int64]map[int32]*sro.CharacterGearExcel
}

func (g *GameConfig) gppCharacterGearExcel() {
	g.GetGPP().CharacterGear = &CharacterGear{
		CharacterGearList:          make(map[int64]*sro.CharacterGearExcel),
		CharacterGearByCharacterId: make(map[int64]map[int32]*sro.CharacterGearExcel),
	}
	for _, v := range g.GetExcel().GetCharacterGearExcel() {
		g.GetGPP().CharacterGear.CharacterGearList[v.Id] = v
		if g.GetGPP().CharacterGear.CharacterGearByCharacterId[v.CharacterId] == nil {
			g.GetGPP().CharacterGear.CharacterGearByCharacterId[v.CharacterId] = make(map[int32]*sro.CharacterGearExcel)
		}
		g.GetGPP().CharacterGear.CharacterGearByCharacterId[v.CharacterId][v.Tier] = v
	}

	logger.Info("处理角色爱用品配置完成,数量:%v个",
		len(g.GetGPP().CharacterGear.CharacterGearList))
}

func GetUnlockCharacterGear(characterId int64) *sro.CharacterGearExcel {
	if GC.GetGPP().CharacterGear.CharacterGearByCharacterId[characterId] == nil {
		return nil
	}
	return GC.GetGPP().CharacterGear.CharacterGearByCharacterId[characterId][1]
}

func GetCharacterGearExcel(id int64) *sro.CharacterGearExcel {
	return GC.GetGPP().CharacterGear.CharacterGearList[id]
}
