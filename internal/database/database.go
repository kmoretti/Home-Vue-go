package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"home-vue-go/internal/ent"
	"home-vue-go/internal/ent/migrate"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type Database struct {
	Client *ent.Client
}

// contactSeed 对应 links.json 中一条联系人的结构
type contactSeed struct {
	Type       string `json:"type"`
	Icon       string `json:"icon"`
	URL        string `json:"url,omitempty"`
	QrCode     string `json:"qrCode,omitempty"`
	HoverColor string `json:"hoverColor,omitempty"`
}

// siteSeed 对应 site.json 中一个站点的结构
type siteSeed struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Icon string `json:"icon"`
}

func Init(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_fk=1")
	if err != nil {
		return nil, err
	}

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := ent.NewClient(ent.Driver(drv))

	// 运行数据库迁移
	ctx := context.Background()
	if err := client.Schema.Create(ctx, migrate.WithForeignKeys(false)); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 初始化默认数据
	if err := initDefaultData(ctx, client); err != nil {
		log.Printf("初始化默认数据失败: %v", err)
	}

	return &Database{Client: client}, nil
}

func (d *Database) Close() error {
	return d.Client.Close()
}

func initDefaultData(ctx context.Context, client *ent.Client) error {
	// 初始化站点配置（仅首次）
	if err := initSiteConfig(ctx, client); err != nil {
		return err
	}

	// 初始化联系人（仅首次）
	if err := initContacts(ctx, client); err != nil {
		log.Printf("初始化联系人失败: %v", err)
	}

	// 初始化站点（仅首次）
	if err := initSites(ctx, client); err != nil {
		log.Printf("初始化站点失败: %v", err)
	}

	// 初始化管理员用户（仅首次）
	return initAdminUser(ctx, client)
}

func initSiteConfig(ctx context.Context, client *ent.Client) error {
	_, err := client.SiteConfig.Get(ctx, 1)
	if err == nil {
		return nil // 已存在
	}

	_, err = client.SiteConfig.Create().
		SetSiteName("个人主页").
		SetSiteURL("https://example.com").
		SetSiteIcon("/favicon.ico").
		SetSiteDescription("一个基于Vue3的个人主页").
		SetSiteKeywords("个人主页,Vue3").
		SetUserName("用户").
		SetProfileImageURL("").
		SetIcpNumber("").
		SetPoliceNumber("").
		SetPageTitle("个人主页").
		SetFavicon("/favicon.ico").
		SetUmamiScript("").
		SetUmamiWebsiteID("").
		SetIconLibrary("//lib.baomitu.com/font-awesome/6.5.0/css/all.min.css").
		SetFontLibrary("").
		Save(ctx)
	return err
}

func initAdminUser(ctx context.Context, client *ent.Client) error {
	userCount, err := client.User.Query().Count(ctx)
	if err != nil {
		return err
	}
	if userCount > 0 {
		return nil // 已有用户
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = client.User.Create().
		SetUsername("admin").
		SetPassword(string(hashedPassword)).
		Save(ctx)
	return err
}

func initContacts(ctx context.Context, client *ent.Client) error {
	// 获取已存在的联系人类型列表
	existingTypes := make(map[string]bool)
	existing, err := client.Contact.Query().All(ctx)
	if err != nil {
		return err
	}
	for _, c := range existing {
		existingTypes[c.Type] = true
	}

	// 有数据时只补全缺失的联系人（适配后续新增的场景）
	// 空表时全部写入
	isFresh := len(existing) == 0

	// 先尝试从文件读取 links.json，回退到内嵌默认数据
	defaults := defaultContacts()
	data := readSeedFile("src/config/links.json")
	if data != nil {
		var contacts []contactSeed
		if err := json.Unmarshal(data, &contacts); err == nil && len(contacts) > 0 {
			defaults = contacts
		}
	}

	added := 0
	for i, c := range defaults {
		if !isFresh && existingTypes[c.Type] {
			continue // 已有该类型，跳过
		}
		builder := client.Contact.Create().
			SetType(c.Type).
			SetIcon(c.Icon).
			SetSortOrder(i)
		if c.URL != "" {
			builder.SetURL(c.URL)
		}
		if c.QrCode != "" {
			builder.SetQrCode(c.QrCode)
		}
		if c.HoverColor != "" {
			builder.SetHoverColor(c.HoverColor)
		}
		if _, err := builder.Save(ctx); err != nil {
			return err
		}
		added++
	}
	if added > 0 {
		log.Printf("已初始化 %d 个默认联系人", added)
	}
	return nil
}

func initSites(ctx context.Context, client *ent.Client) error {
	// 获取已存在的站点名称列表
	existingNames := make(map[string]bool)
	existing, err := client.Site.Query().All(ctx)
	if err != nil {
		return err
	}
	for _, s := range existing {
		existingNames[s.Name] = true
	}

	isFresh := len(existing) == 0

	// 先尝试从文件读取 site.json，回退到内嵌默认数据
	defaults := defaultSites()
	data := readSeedFile("src/config/site.json")
	if data != nil {
		var sites []siteSeed
		if err := json.Unmarshal(data, &sites); err == nil && len(sites) > 0 {
			defaults = sites
		}
	}

	added := 0
	for i, s := range defaults {
		if !isFresh && existingNames[s.Name] {
			continue
		}
		if _, err := client.Site.Create().
			SetName(s.Name).
			SetURL(s.URL).
			SetIcon(s.Icon).
			SetSortOrder(i).
			Save(ctx); err != nil {
			return err
		}
		added++
	}
	if added > 0 {
		log.Printf("已初始化 %d 个默认站点", added)
	}
	return nil
}

// readSeedFile 尝试从路径读取 JSON 配置文件
func readSeedFile(path string) []byte {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil
	}
	data, err := os.ReadFile(abs)
	if err == nil {
		return data
	}
	return nil
}

// defaultContacts 返回内嵌的默认联系人数据（对应 links.json）
func defaultContacts() []contactSeed {
	return []contactSeed{
		{Type: "Email", Icon: "fas fa-envelope", URL: "mailto:i@bsgun.cn", HoverColor: "#e78b0a"},
		{Type: "Github", Icon: "fab fa-github", URL: "https://github.com/JLinmr", HoverColor: "#6500fc"},
		{Type: "支付宝", Icon: "fab fa-alipay", QrCode: "https://lib.bsgun.cn/Hexo-static/img/zfbzf.avif", HoverColor: "#007aff"},
		{Type: "微信", Icon: "fab fa-weixin", QrCode: "https://lib.bsgun.cn/Hexo-static/img/wxzf.avif", HoverColor: "#247700"},
		{Type: "QQ", Icon: "fab fa-qq", QrCode: "https://imgbed.081531.xyz/file/telegram/qq.gif", HoverColor: "#F3A694"},
		{Type: "Telegram", Icon: "fab fa-telegram", URL: "https://t.me/Kemeow0815", HoverColor: "#3271AE"},
		{Type: "BiliBili", Icon: "fab fa-bilibili", URL: "https://space.bilibili.com/3546643173477234", HoverColor: "#3271AE"},
	}
}

// defaultSites 返回内嵌的默认站点数据（对应 site.json）
func defaultSites() []siteSeed {
	return []siteSeed{
		{Name: "博客", URL: "https://blog.bsgun.cn", Icon: "fa fa-blog"},
		{Name: "雨云", URL: "https://www.rainyun.com/Lin_", Icon: "fa fa-cloud"},
		{Name: "图床", URL: "https://dev.bsgun.cn", Icon: "fa fa-image"},
		{Name: "封面", URL: "https://cover.bsgun.cn", Icon: "fa fa-panorama"},
		{Name: "监测", URL: "https://status.bsgun.cn", Icon: "fa fa-chart-line"},
		{Name: "图标", URL: "https://icon.bsgun.cn/", Icon: "fa fa-icons"},
	}
}
