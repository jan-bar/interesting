package calendar

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type (
	PerpetualCalendar struct {
		//Status       string `json:"status"`
		//T            string `json:"t"`
		//SetCacheTime string `json:"set_cache_time"`
		Data []PerpetualCalendarData `json:"data"`
	}
	PerpetualCalendarData struct {
		//ExtendedLocation string `json:"ExtendedLocation"`
		//OriginQuery      string `json:"OriginQuery"`
		//SiteID           int    `json:"SiteId"`
		//StdStg           int    `json:"StdStg"`
		//StdStl           int    `json:"StdStl"`
		//SelectTime       int    `json:"_select_time"`
		//UpdateTime       string `json:"_update_time"`
		//Version          int    `json:"_version"`
		//Appinfo          string `json:"appinfo"`
		//CambrianAppid    string `json:"cambrian_appid"`
		//DispType         int    `json:"disp_type"`
		//Fetchkey         string `json:"fetchkey"`
		//Key              string `json:"key"`
		//Loc              string `json:"loc"`
		//Resourceid       string `json:"resourceid"`
		//RoleID           int    `json:"role_id"`
		//Showlamp         string `json:"showlamp"`
		//Tplt             string `json:"tplt"`
		//URL              string `json:"url"`
		Almanac []PerpetualCalendarAlmanac `json:"almanac"`
	}
	PerpetualCalendarAlmanac struct {
		SolarId int `gorm:"column:solarId;primarykey"`             // 阳历ymd,作为主键
		LunarId int `gorm:"column:lunarId;not null;index:lunarId"` // 阴历ymd,加入索引

		Animal         string    `json:"animal" gorm:"column:animal;not null;size:4"`                     // 生肖
		Suit           string    `json:"suit" gorm:"column:suit;not null"`                                // 宜
		Avoid          string    `json:"avoid" gorm:"column:avoid;not null"`                              // 忌
		CnDay          string    `json:"cnDay" gorm:"column:cnDay;not null;size:4"`                       // 星期
		Day            int       `json:"day,string" gorm:"column:day;not null;uniqueIndex:Solar"`         // 阳历日
		Month          int       `json:"month,string" gorm:"column:month;not null;uniqueIndex:Solar"`     // 阳历月
		Year           int       `json:"year,string" gorm:"column:year;not null;uniqueIndex:Solar"`       // 阳历年
		GzDate         string    `json:"gzDate" gorm:"column:gzDate;not null;size:8"`                     // 干支日
		GzMonth        string    `json:"gzMonth" gorm:"column:gzMonth;not null;size:8"`                   // 干支月
		GzYear         string    `json:"gzYear" gorm:"column:gzYear;not null;size:8"`                     // 干支年
		IsBigMonth     string    `json:"isBigMonth" gorm:"-"`                                             // json取数据,忽略gorm
		IsBigMonthBool bool      `gorm:"column:isBigMonth;not null;default:0"`                            // 是否为阴历大月
		LDate          string    `json:"lDate" gorm:"column:lDate;not null;size:4"`                       // 阴历日,汉字
		LMonth         string    `json:"lMonth" gorm:"column:lMonth;not null;size:4"`                     // 阴历月,汉字,带'闰'字表示闰月
		LunarDate      int       `json:"lunarDate,string" gorm:"column:lunarDate;not null;index:Lunar"`   // 阴历日,数字
		LunarMonth     int       `json:"lunarMonth,string" gorm:"column:lunarMonth;not null;index:Lunar"` // 阴历月,数字
		LunarYear      int       `json:"lunarYear,string" gorm:"column:lunarYear;not null;index:Lunar"`   // 阴历年,数字
		ODate          time.Time `json:"oDate" gorm:"column:oDate;not null"`                              // ODate.Local(),阳历当天0点
		Term           string    `json:"term,omitempty" gorm:"column:term;not null"`                      // 如'除夕','万圣节','三伏'等
		Desc           string    `json:"desc,omitempty" gorm:"column:desc;not null"`                      // 如'腊八节','下元节'等
		Type           string    `json:"type,omitempty" gorm:"column:type;not null;size:2"`               // a,c,h,i,t,用这种查询感觉看不出来啥:"SELECT *FROM perpetualCalendarAlmanac WHERE TYPE ='a' GROUP BY TERM HAVING COUNT(TERM)>1"
		Value          string    `json:"value,omitempty" gorm:"column:value;not null"`                    // 如'国际残疾人日'等
		Status         int       `json:"status,string,omitempty" gorm:"column:status;not null;default:0"` // 0 工作日,1 休假,2 上班,3 周末
	}
)

func (PerpetualCalendarAlmanac) TableName() string {
	return "perpetualCalendarAlmanac"
}

// GetPerpetualCalendar 返回[前一个月,本月,后一个月]的数据
func GetPerpetualCalendar(year, mouth int) ([]PerpetualCalendarAlmanac, error) {
	u := url.Values{}
	u.Add("tn", "wisetpl")
	u.Add("oe", "utf8")
	u.Add("format", "json")
	u.Add("resource_id", "39043") // 这个用浏览器请求后得到的
	u.Add("query", fmt.Sprintf("%d年%d月", year, mouth))
	u.Add("t", strconv.FormatInt(time.Now().UnixMilli(), 10))

	urls := "https://sp1.baidu.com/8aQDcjqpAAV3otqbppnN2DJv/api.php?" + u.Encode()
	resp, err := http.Get(urls) // 百度这个接口可能现在请求速度,所以可能报错
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var ret PerpetualCalendar
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return nil, err
	}
	if len(ret.Data) != 1 { // 该数组目前只会有一个
		return nil, errors.New("get Data error")
	}

	for i, v := range ret.Data[0].Almanac {
		// 阳历ymd,作为主键
		ret.Data[0].Almanac[i].SolarId = v.Year*10000 + v.Month*100 + v.Day
		// 阴历ymd,加入索引
		ret.Data[0].Almanac[i].LunarId = v.LunarYear*10000 + v.LunarMonth*100 + v.LunarDate

		// 赋值大月
		ret.Data[0].Almanac[i].IsBigMonthBool = v.IsBigMonth == "1"

		if v.Status == 0 && (v.CnDay == "六" || v.CnDay == "日") {
			ret.Data[0].Almanac[i].Status = 3 // 不是特殊类型,且为周末则赋值
		}
	}
	return ret.Data[0].Almanac, nil
}

func SaveCalendar(dsnSrc string) error {
	ts := time.Now()
	defer func() {
		fmt.Println(time.Since(ts))
	}()

	iDsn := strings.Index(dsnSrc, ":")
	if iDsn < 0 {
		return errors.New("dsn error")
	}

	var isMysql bool
	var gormOpen gorm.Dialector
	switch strings.ToLower(dsnSrc[:iDsn]) {
	case "mysql":
		gormOpen = mysql.Open(dsnSrc[iDsn+1:])
		isMysql = true
	case "sqlite":
		gormOpen = sqlite.Open(dsnSrc[iDsn+1:])
	default:
		return errors.New("just support mysql or sqlite")
	}

	db, err := gorm.Open(gormOpen, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return err
	}

	res := db.Exec("DROP table if exists " + PerpetualCalendarAlmanac{}.TableName())
	if res.Error != nil {
		return res.Error
	}

	if isMysql {
		db = db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4")
	}
	err = db.AutoMigrate(PerpetualCalendarAlmanac{})
	if err != nil {
		return err
	}

	var data []PerpetualCalendarAlmanac
	// 起止时间按照百度万年历得到
	start := time.Date(1900, time.February, 1, 0, 0, 0, 0, time.Local)
	end := time.Date(2050, time.December, 1, 0, 0, 0, 0, time.Local)
	// 由于每次查询包含前一个月,当月,下个月,因此每次都增加3个月进行查询
	for ; start.Before(end); start = start.AddDate(0, 3, 0) {
		y, m, _ := start.Date()
		for { // 使用协程并发请求,提高速度,出现错误时重试
			data, err = GetPerpetualCalendar(y, int(m))
			if err == nil {
				break
			}
			// 报错重试,直到成功
			fmt.Println("GetPerpetualCalendar", y, m, err)
		}

		for {
			res := db.Create(&data)
			if res.Error == nil {
				break
			}
			// 报错重试,直到成功
			fmt.Println("Create", y, m, res.Error)
		}
	}
	return nil
}
