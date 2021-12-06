package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tungyao/tjson"
	cedar "github.com/tungyao/ultimate-cedar"
)

func init() {
	// 校验本地路径是否存在
	fs, err := os.Open("./temp")
	if os.IsNotExist(err) {
		os.Mkdir("./temp", 755)
	}
	fs.Close()
}
func main() {
	r := cedar.NewRouter()
	r.Get("/", func(writer cedar.ResponseWriter, request cedar.Request) {
		fs, err := os.OpenFile("../template/index.html", os.O_RDONLY, 0777)
		if err != nil {
			log.Println(err)
		}
		html, err := ioutil.ReadAll(fs)
		if err != nil {
			log.Println(err)
		}
		_, _ = writer.Write(html)
	})
	r.Post("newnamedesk", func(writer cedar.ResponseWriter, request cedar.Request) {
		data := make([]byte, 1024)
		nt, _ := request.Body.Read(data)
		request.Body.Close()
		t, _ := tjson.Decode(data[:nt])
		first := t["first"].(string)
		number := t["number"].(string)
		alln := "2000"
		if len(t) > 2 {
			alln = t["alln"].(string)
		}
		var m strings.Builder
		n, _ := strconv.Atoi(number)
		na, _ := strconv.Atoi(alln)
		for i := 1; i < na+1; i++ {
			ss := GetRandomName(n, first)
			if i%10 != 0 {
				m.WriteString(ss + "\t")
			} else {
				m.WriteString(ss + "\n")
			}
		}
		tm := strconv.Itoa(int(time.Now().Unix()) + rand.Int())
		WriteStringToFileS("./static/temp/"+tm+"-name.txt", m.String())
		err := Zip("./static/temp/"+tm+"-name.txt", "./static/temp/"+tm+"-name.zip")
		if err != nil {
			log.Println(err)
		}
		writer.Header().Set("content-type", "application/x-zip-compressed")
		_, _ = writer.Write([]byte("./static/temp/" + tm + "-name.zip"))
	})
	r.Post("newname", func(writer cedar.ResponseWriter, request cedar.Request) {
		data := make([]byte, 1024)
		nt, _ := request.Body.Read(data)
		request.Body.Close()
		t, _ := tjson.Decode(data[:nt])
		first := t["first"].(string)
		number := t["number"].(string)
		alln := "2000"
		if len(t) > 2 {
			alln = t["alln"].(string)
		}
		var m strings.Builder
		n, _ := strconv.Atoi(number)
		na, _ := strconv.Atoi(alln)
		for i := 1; i < na+1; i++ {
			ss := GetRandomName(n, first)
			if i%10 != 0 {
				m.WriteString(ss + "\t")
			} else {
				m.WriteString(ss + "\n")
			}
		}
		tm := strconv.Itoa(int(time.Now().Unix()) + rand.Int())
		WriteStringToFileS("./temp/"+tm+"-name.txt", m.String())
		err := Zip("./temp/"+tm+"-name.txt", "./temp/"+tm+"-name.zip")
		if err != nil {
			log.Println(err)
		}
		writer.Header().Set("content-type", "application/x-zip-compressed")
		_, _ = writer.Write([]byte("./temp/" + tm))
	})
	r.Get("temp/:name", func(writer cedar.ResponseWriter, request cedar.Request) {
		f, err := os.Open("./temp/" + request.Data.Get("name"))
		if err != nil {
			log.Println(err)
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		defer f.Close()
		io.Copy(writer, f)
	})
	// 在线查看
	// https://newname.tungyao.com/temp/8674665224710888359-name.zip
	r.Get("look/:name", func(writer cedar.ResponseWriter, request cedar.Request) {
		f, err := os.Open("./temp/" + request.Data.Get("name") + "-name.txt")
		if err != nil {
			log.Println(err)
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		defer f.Close()
		d, err := io.ReadAll(f)
		writer.Header().Set("content-type", "text/html")
		writer.Write([]byte("<!DOCTYPE html><html lang=\"en\"><head><meta charset=\"UTF-8\"><title>Look</title></head><body><div style='max-width:1108px;margin:auto;font-size:x-large;'>" + string(d) + "</div></body></html>"))
	})
	http.ListenAndServe(":8400", r)
}
func Zip(srcFile string, destZip string) error {
	zipfile, err := os.Create(destZip)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()
	filepath.Walk(srcFile, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = strings.TrimPrefix(path, filepath.Dir(srcFile)+"/")
		// header.Name = path
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}
		header.Name = filepath.Base(srcFile)
		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
		}
		return err
	})

	return err
}

var (
	familyNames    = []string{"周"}
	middleNamesMap = map[string][]string{}

	// 定义一堆名字
	names = "纤云弄巧飞星传恨银迢迢暗风玉露一相逢胜间无数饮醒复家息雷应倚杖听江声长此何时忘却营阑风静纹平小舟从此逝江海寄余生梦" +
		"后楼台高锁帘幕年春恨却时落花立微雨双生若只如初见何事秋风画胜寻芳泗水滨无边光景一时新等闲识得风千春江潮水连海平海明月共潮生随波千" +
		"万里何处春江无月明江流宛转绕芳甸月照花林皆似霰空里流霜觉飞汀白沙看见江天一色无纤尘皎皎空中孤月轮江畔何初见月江月何年初照生代" +
		"无已江月年只相似知江月待何但见长江送流水白云一去悠悠青枫浦谁家今扁舟子何相思明月楼可楼月徘徊应照离妆镜台玉户帘中卷去衣砧拂还" +
		"此时相望相闻愿逐月华流照君鸿长飞光鱼潜跃水成文昨闲潭梦落花可春半还家江水流春去欲尽江潭落月复西斜斜月沉沉藏海雾碣石潇湘无限路知乘" +
		"月几归落月摇情满江树院深几许杨柳烟帘幕无数玉勒鞍游冶楼高见章台路雨风三月暮门掩黄昏无计留春住泪眼问花花语乱飞过秋千去" +
		"秋风清秋月明落叶聚还散寒复惊相思相见知何此时此难为情入我相思门知我相思苦长相思兮长相忆相思兮无极早知如此绊心何如当初莫相识" +
		"天门中断楚江开碧流回两岸青山相对出帆明月出天山苍茫云海间长风几万里吹玉门关白登道胡青海由来征战地见有还客望边邑思归多苦颜望边楼未应闲" +
		"中岁颇好道晚家南山陲来每独往胜事空自知行到水穷坐看云起时偶然值林叟笑无还期新丰美酒斗十千阳游年相逢意气为君饮系马高楼垂柳边出身仕汉羽林郎" +
		"初随骠战渔阳知向边庭苦纵死犹闻侠香一身能两弧千只似无偏坐金鞍调白羽纷纷射五单于家君臣欢宴终高议云台论战功天子临轩赐侯印将佩明" +
		"积雨空林烟火迟炊黍饷东漠水田飞白鹭阴阴夏木啭黄鹂山中习静观朝槿松下清斋折露葵与席何事更相疑尊前拟把归期说欲语春容先惨咽生自是有情痴此恨关风与月离歌且莫翻新阕一曲结直须看尽洛城花始共春风容易别" +
		"驿外断桥边寂寞开无主已是黄昏独自愁更著风和雨无意苦争春一任群芳妒零落成泥碾作尘只有香如故早岁那知世事艰中原北望气如山楼船夜雪瓜洲渡铁马秋风大散关塞上长城空自许镜中衰鬓已先斑出师一表真名世千载谁堪伯仲" +
		"红藕香残玉簟秋轻解罗裳独上兰舟云中谁寄锦书来雁字回月满西楼花自飘零水自流一种相思两闲愁此情无计可消除才下眉头却上心头" +
		"薄雾浓云愁永昼瑞脑销金兽佳节又重阳玉枕纱橱半夜初透东篱把酒黄昏后有暗香盈袖莫道销魂帘卷西风比" +
		"暗淡轻黄体性柔情疏迹远只香留何须浅碧深红色自是花中第一流梅定妒菊应羞画阑开冠中秋骚人可煞无情思何事当年见收" +
		"小楼寒夜长帘幕低垂恨萧萧无情风雨夜来揉损琼肌也似贵妃醉脸也似孙寿愁眉韩令偷香徐娘傅粉莫将比拟未新奇细看取屈平陶令风韵正相宜微风起清芬酝减酴醿渐秋雪清玉瘦向无限依依似愁凝汉皋解佩似泪洒纨扇题诗朗月清风浓烟暗雨天教憔悴度芳姿纵爱惜知从此留得几多时人情好何须更忆泽畔东篱" +
		"落熔金暮云合璧在何染柳烟浓吹梅笛春意知几许元宵佳节融和天气次第岂无风雨来相召香车宝马谢他酒朋诗侣中州盛闺门多暇记得偏重三五铺翠冠儿捻金雪柳簇带争济楚如今憔悴风鬟霜鬓见夜间出去如向帘儿底下听笑语" +
		"寒萧萧上琐窗梧桐应恨夜来霜酒阑更喜团茶苦梦断偏宜瑞脑香秋已尽犹长仲宣怀远更凉如随分尊前醉莫负东篱菊蕊黄"
	lastName  []string
	lastNames []string
)

func GetRandomName(n int, first string) string {
	familyName := first
	var s string
	if n == 3 {
		middleName := lastNames[GetRandomInt(0, len(lastNames)-1)]
		lastName := lastNames[GetRandomInt(0, len(lastNames)-1)]
		s += familyName + middleName + lastName
	} else if n == 2 {
		lastName := lastNames[GetRandomInt(0, len(lastNames)-1)]
		s += familyName + lastName
	} else if n == 4 {
		lastName := lastNames[GetRandomInt(0, len(lastNames)-1)]
		lastNamee := lastNames[GetRandomInt(0, len(lastNames)-1)]
		lastNameee := lastNames[GetRandomInt(0, len(lastNames)-1)]
		s += familyName + lastName + lastNamee + lastNameee
	}
	return s
}
func init() {
	data, err := ioutil.ReadFile("../shici.txt")
	var get string
	if err != nil {
		fmt.Println("File reading error", err)
		get = string(data)
		reg := regexp.MustCompile(`\s*\pP*`)
		get = reg.ReplaceAllString(get, "")
		for _, s := range get {
			lastNames = append(lastNames, string(s))
		}

	} else {
		get = string(data)
		reg := regexp.MustCompile(`\s*\pP*`)
		get = reg.ReplaceAllString(get, "")
		for _, s := range get {
			lastNames = append(lastNames, string(s))
		}
	}
	fmt.Println(lastNames)
	for _, x := range familyNames {
		middleNamesMap[x] = []string{"德", "惟", "守", "世", "令", "子", "伯", "师", "希", "与", "孟", "由", "宜", "顺", "元", "允", "宗", "仲", "士", "不", "善", "汝", "崇", "必", "良", "友", "季", "同"}
	}
}

var (
	// 随机数互斥锁（确保GetRandomInt不能被并发访问）
	randomMutex sync.Mutex
)

func GetRandomInt(start, end int) int {
	randomMutex.Lock()
	<-time.After(1 * time.Nanosecond)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	n := start + r.Intn(end-start+1)
	randomMutex.Unlock()
	return n
}

// 定义cmd参数
var (
	numberFlag    = flag.Int("number", 100, "需要名字个数 : -number")
	lengthFlag    = flag.Int("length", 3, "需要名字长度 : -length")
	firstNameFlag = flag.String("first", "周", "姓氏 : -first")
)

func WriteStringToFileS(filepath, content string) {
	fs, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		log.Println(err)
	}
	defer fs.Close()
	_, err = fs.WriteString(content)
	if err != nil {
		log.Println(err)
	}
}
