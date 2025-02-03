package gdconf

import (
	"encoding/json"
	"os"

	sro "github.com/gucooing/BaPs/common/server_only"
	"github.com/gucooing/BaPs/pkg/logger"
)

func (g *GameConfig) loadCharacterSkillListExcelTable() {
	g.GetExcel().CharacterSkillListExcelTable = make([]*sro.CharacterSkillListExcelTable, 0)
	name := "CharacterSkillListExcelTable.json"
	file, err := os.ReadFile(g.excelPath + name)
	if err != nil {
		logger.Error("文件:%s 读取失败,err:%s", name, err)
		return
	}
	if err := json.Unmarshal(file, &g.GetExcel().CharacterSkillListExcelTable); err != nil {
		logger.Error("文件:%s 解析失败,err:%s", name, err)
		return
	}

	logger.Info("文件:%s 读取成功,解析数量:%v", name, len(g.GetExcel().GetCharacterSkillListExcelTable()))
}

type CharacterSkillListExcel struct {
	CharacterSkillListExcelMap map[int64]map[int32]*sro.CharacterSkillListExcelTable
}

func (g *GameConfig) gppCharacterSkillListExcelTable() {
	g.GetGPP().CharacterSkillListExcel = &CharacterSkillListExcel{
		CharacterSkillListExcelMap: make(map[int64]map[int32]*sro.CharacterSkillListExcelTable),
	}

	for _, v := range g.GetExcel().GetCharacterSkillListExcelTable() {
		if g.GetGPP().CharacterSkillListExcel.CharacterSkillListExcelMap[v.CharacterSkillListGroupId] == nil {
			g.GetGPP().CharacterSkillListExcel.CharacterSkillListExcelMap[v.CharacterSkillListGroupId] =
				make(map[int32]*sro.CharacterSkillListExcelTable)
		}
		if v.FormIndex != 0 || v.MinimumGradeCharacterWeapon != 0 {
			continue
		}
		g.GetGPP().CharacterSkillListExcel.CharacterSkillListExcelMap[v.CharacterSkillListGroupId][v.MinimumGradeCharacterWeapon] = v
	}

	logger.Info("处理角色技能配置完成,角色技能配置:%v个",
		len(g.GetGPP().CharacterSkillListExcel.CharacterSkillListExcelMap))
}

func GetCharacterSkillListExcelTable(characterId int64, weaponLevel int32) *sro.CharacterSkillListExcelTable {
	list := GC.GetGPP().CharacterSkillListExcel.CharacterSkillListExcelMap[characterId]
	if list == nil {
		return nil
	}

	maxLevel := int32(0)
	for _, conf := range list {
		if weaponLevel > conf.MinimumGradeCharacterWeapon &&
			maxLevel < conf.MinimumGradeCharacterWeapon {
			maxLevel = conf.MinimumGradeCharacterWeapon
		}
	}
	return list[maxLevel]
}
