package http

import (
	"encoding/json"
	"net/http"

	"backend/internal/repo"
)

type mapLocation struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Icon     string `json:"icon"`
	District string `json:"district"`
	Size     string `json:"size"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
}

type mapDistrict struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Color  string `json:"color"`
}

type mapRoad struct {
	From uint   `json:"from"`
	To   uint   `json:"to"`
	Name string `json:"name"`
}

type mapData struct {
	Name      string        `json:"name"`
	Width     int           `json:"width"`
	Height    int           `json:"height"`
	Districts []mapDistrict `json:"districts"`
	Locations []mapLocation `json:"locations"`
	Roads     []mapRoad     `json:"roads"`
}

// staticCoords maps location names to API coordinates reverse-scaled from Godot map image positions.
// Godot map image is 2048x1152, API coordinate system is 1000x800.
// Formula: api_x = image_x * 1000 / 2048, api_y = image_y * 800 / 1152
var staticCoords = map[string]struct{ X, Y int }{
	"广场":   {475, 336},
	"咖啡馆":  {340, 152},
	"钟楼":   {690, 200},
	"市政厅":  {475, 168},
	"图书馆":  {820, 212},
	"花店":   {235, 144},
	"铁匠铺":  {620, 460},
	"诊所":   {340, 364},
	"农舍":   {115, 119},
	"钓鱼小屋": {370, 692},
	"学校":   {245, 480},
	"面包店":  {235, 256},
	"酒馆":   {715, 340},
	"公园凉亭": {115, 648},
	"手工工坊": {240, 360},
	"住宅区":  {485, 478},
	"森林营地": {925, 636},
}

var staticDistricts = []mapDistrict{
	{ID: "central", Name: "中央广场区", X: 350, Y: 250, Width: 300, Height: 200, Color: "#fffde7"},
	{ID: "left_commercial", Name: "左区商业街", X: 100, Y: 150, Width: 220, Height: 350, Color: "#f3e5f5"},
	{ID: "right", Name: "右区", X: 680, Y: 150, Width: 250, Height: 350, Color: "#e3f2fd"},
	{ID: "farmland", Name: "远左农田区", X: 0, Y: 50, Width: 180, Height: 200, Color: "#e8f5e9"},
	{ID: "lake", Name: "湖区", X: 350, Y: 0, Width: 200, Height: 100, Color: "#e0f7fa"},
	{ID: "park", Name: "公园区", X: 300, Y: 550, Width: 280, Height: 150, Color: "#c8e6c9"},
	{ID: "residential", Name: "住宅区", X: 150, Y: 600, Width: 150, Height: 130, Color: "#fff3e0"},
	{ID: "forest", Name: "森林边缘", X: 800, Y: 520, Width: 200, Height: 200, Color: "#dcedc8"},
}

type roadDef struct {
	From string
	To   string
	Name string
}

var staticRoads = []roadDef{
	{"广场", "咖啡馆", "商业街路"},
	{"广场", "钟楼", "钟楼路"},
	{"广场", "市政厅", "市政路"},
	{"广场", "图书馆", "学府路"},
	{"咖啡馆", "花店", "花巷"},
	{"咖啡馆", "面包店", "面包巷"},
	{"广场", "铁匠铺", "铁匠路"},
	{"铁匠铺", "酒馆", "酒馆巷"},
	{"咖啡馆", "诊所", "诊所路"},
	{"咖啡馆", "学校", "学校路"},
	{"农舍", "面包店", "田间路"},
	{"钓鱼小屋", "钟楼", "湖滨路"},
	{"广场", "公园凉亭", "公园路"},
	{"诊所", "住宅区", "住宅路"},
	{"酒馆", "森林营地", "森林小径"},
	{"铁匠铺", "森林营地", "林边路"},
}

func newMapHandler(locRepo *repo.LocationRepo) http.HandlerFunc {
	// Fetch real DB location IDs
	dbLocs, err := locRepo.FindByTownID(1)
	if err != nil {
		dbLocs = nil
	}
	nameToID := make(map[string]uint)
	for _, l := range dbLocs {
		nameToID[l.Name] = l.ID
	}

	// Build locations with real DB IDs (fallback to 0 if not found)
	var locations []mapLocation
	for name, coord := range staticCoords {
		id := nameToID[name]
		locations = append(locations, mapLocation{
			ID:       id,
			Name:     name,
			Icon:     locIcon(name),
			District: locDistrict(name),
			Size:     locSize(name),
			X:        coord.X,
			Y:        coord.Y,
		})
	}

	// Build roads with real DB IDs
	var roads []mapRoad
	for _, r := range staticRoads {
		fromID := nameToID[r.From]
		toID := nameToID[r.To]
		if fromID != 0 && toID != 0 {
			roads = append(roads, mapRoad{From: fromID, To: toID, Name: r.Name})
		}
	}

	data := mapData{
		Name:      "晨曦镇",
		Width:     1000,
		Height:    800,
		Districts: staticDistricts,
		Locations: locations,
		Roads:     roads,
	}

	body, _ := json.Marshal(data)

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(body)
	}
}

func locIcon(name string) string {
	icons := map[string]string{
		"广场": "🏛", "咖啡馆": "☕", "钟楼": "🕐", "市政厅": "🏛",
		"图书馆": "📚", "花店": "🌸", "铁匠铺": "🔨", "诊所": "🏥",
		"农舍": "🌾", "钓鱼小屋": "🎣", "学校": "🏫", "面包店": "🍞",
		"酒馆": "🍺", "公园凉亭": "🎵", "手工工坊": "🪚", "住宅区": "🏠", "森林营地": "🏕",
	}
	if i, ok := icons[name]; ok {
		return i
	}
	return "📍"
}

func locDistrict(name string) string {
	d := map[string]string{
		"广场": "central", "钟楼": "central", "市政厅": "central", "图书馆": "central",
		"咖啡馆": "left_commercial", "花店": "left_commercial", "面包店": "left_commercial",
		"诊所": "left_commercial", "学校": "left_commercial", "手工工坊": "left_commercial",
		"铁匠铺": "right", "酒馆": "right",
		"农舍": "farmland", "钓鱼小屋": "lake",
		"公园凉亭": "park", "住宅区": "residential", "森林营地": "forest",
	}
	if dd, ok := d[name]; ok {
		return dd
	}
	return ""
}

func locSize(name string) string {
	s := map[string]string{
		"广场": "large", "钟楼": "large",
		"花店": "small", "面包店": "small", "钓鱼小屋": "small", "手工工坊": "small",
	}
	if ss, ok := s[name]; ok {
		return ss
	}
	return "medium"
}
