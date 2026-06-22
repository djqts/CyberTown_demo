package service

import (
	"fmt"
	"strconv"
	"strings"

	"backend/internal/model"
	"backend/internal/repo"
)

// NPCMove 描述一次 NPC 待执行的移动。
type NPCMove struct {
	NPC          model.NPC
	Schedule     model.NPCSchedule
	FromLocation uint
	ToLocation   uint
}

// NPCService NPC 日程和移动业务逻辑。
type NPCService struct {
	npcRepo      *repo.NPCRepo
	scheduleRepo *repo.ScheduleRepo
}

// NewNPCService 创建 NPC 服务。
func NewNPCService(npcRepo *repo.NPCRepo, scheduleRepo *repo.ScheduleRepo) *NPCService {
	return &NPCService{npcRepo: npcRepo, scheduleRepo: scheduleRepo}
}

// FindActiveMoves 查找当前分钟应执行移动的 NPC。
// 日程的 StartTime/EndTime 格式为 "HH:MM"，与当前 minuteOfDay (0~1439) 比较。
func (s *NPCService) FindActiveMoves(townID uint, minuteOfDay int) ([]NPCMove, error) {
	schedules, err := s.scheduleRepo.FindByTownID(townID)
	if err != nil {
		return nil, fmt.Errorf("find schedules: %w", err)
	}

	var moves []NPCMove
	for _, sched := range schedules {
		start := timeStrToMinute(sched.StartTime)
		end := timeStrToMinute(sched.EndTime)

		if minuteOfDay >= start && minuteOfDay < end {
			npc, err := s.npcRepo.FindByID(sched.NPCID)
			if err != nil {
				return nil, fmt.Errorf("find npc %d: %w", sched.NPCID, err)
			}
			if npc.LocationID != sched.LocationID {
				moves = append(moves, NPCMove{
					NPC:          *npc,
					Schedule:     sched,
					FromLocation: npc.LocationID,
					ToLocation:   sched.LocationID,
				})
			}
		}
	}
	return moves, nil
}

// MoveNPC 更新 NPC 位置和状态。
func (s *NPCService) MoveNPC(npcID, newLocationID uint) error {
	npc, err := s.npcRepo.FindByID(npcID)
	if err != nil {
		return fmt.Errorf("find npc: %w", err)
	}

	if npc.LocationID == newLocationID {
		return nil
	}

	npc.LocationID = newLocationID
	npc.Status = "moving"

	return s.npcRepo.Update(npc)
}

// timeStrToMinute 将 "HH:MM" 转换为一天中的分钟数。
func timeStrToMinute(s string) int {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return 0
	}
	h, _ := strconv.Atoi(parts[0])
	m, _ := strconv.Atoi(parts[1])
	return h*60 + m
}
