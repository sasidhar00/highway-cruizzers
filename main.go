// main.go - COMPLETE HIGHWAY CRUIZZERS WITH BOOKINGS TAB & TRIP PLANNER
package main

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
	"golang.org/x/crypto/bcrypt"
)

// ==================== STRUCT DEFINITIONS ====================

type Ride struct {
	ID           int
	User         string
	Handle       string
	Content      string
	Price        string
	Category     string
	Tags         string
	ImageURL     string
	Likes        int
	IsBoosted    bool
	IsFeatured   bool
	CreatedAt    time.Time
	UserID       int
	Status       string
	FromLocation string
	ToLocation   string
	DepartureDate string
	BikeModel    string
	Seats        int
}

type User struct {
	ID            int
	Username      string
	Handle        string
	Email         string
	Phone         string
	IsAdmin       bool
	Credits       int
	IsPremium     bool
	IsActive      bool
	PremiumUntil  *time.Time
	MembershipTier string
	BikeModel     string
	RidingExp     string
	AvatarURL     string
	IsVerified    bool
}

type BikingNews struct {
	ID        int
	Title     string
	Content   string
	Timestamp string
	Category  string
	Likes     int
}

type BikingTrend struct {
	ID          int
	Title       string
	Description string
	Trend       string
	Percentage  string
	Category    string
}

type RideRequest struct {
	ID            int
	RideID        int
	RiderID       int
	RiderName     string
	RiderEmail    string
	RiderPhone    string
	Message       string
	Status        string
	CreatedAt     time.Time
}

type Product struct {
	ID          int
	UserID      int
	UserName    string
	Title       string
	Description string
	Price       float64
	Category    string
	Condition   string
	ImageURL    string
	Location    string
	Status      string
	CreatedAt   time.Time
}

type Advertisement struct {
	ID          int
	Title       string
	ImageURL    string
	TargetURL   string
	Position    string
	StartDate   time.Time
	EndDate     time.Time
	Advertiser  string
	IsActive    bool
	Impressions int
	Clicks      int
}

type MarketplaceCategory struct {
	ID   int
	Name string
	Icon string
	Slug string
}

type SiteSettings struct {
	ID          int
	SiteName    string
	LogoURL     string
	FaviconURL  string
	PrimaryColor string
	SecondaryColor string
}

type RiderVerification struct {
	ID         int
	UserID     int
	DLNumber   string
	BikeRCNumber string
	Status     string
	VerifiedAt *time.Time
	VerifiedBy int
}

type Experience struct {
	ID             int
	Title          string
	Category       string
	Description    string
	DurationDays   int
	DurationNights int
	Price          int
	DiscountedPrice int
	MaxPeople      int
	MinPeople      int
	Location       string
	StartLocation  string
	EndLocation    string
	VehicleType    string
	IsFeatured     bool
	IsActive       bool
	CoverImage     string
	Rating         float64
	TotalReviews   int
	CreatedAt      time.Time
}

type ExperienceBooking struct {
	ID             int
	ExperienceID   int
	UserID         int
	TravelDate     time.Time
	NumberOfPeople int
	TotalPrice     int
	SpecialRequests string
	Status         string
	ContactName    string
	ContactPhone   string
	ContactEmail   string
	CreatedAt      time.Time
}

type RSSFeed struct {
	Channel struct {
		Title       string `xml:"title"`
		Link        string `xml:"link"`
		Description string `xml:"description"`
		Items       []struct {
			Title       string `xml:"title"`
			Link        string `xml:"link"`
			Description string `xml:"description"`
			PubDate     string `xml:"pubDate"`
			Guid        string `xml:"guid"`
		} `xml:"item"`
	} `xml:"channel"`
}

type PointOfInterest struct {
	ID              int
	Name            string
	Type            string
	Category        string
	Latitude        float64
	Longitude       float64
	Address         string
	City            string
	State           string
	Phone           string
	PriceRange      string
	Rating          float64
	TotalReviews    int
	Amenities       string
	Images          string
	IsPartner       bool
	DiscountPercent int
	OfferDetails    string
	Distance        float64
	OpeningTime     string
	ClosingTime     string
	Is24x7          bool
	Email           string
	Website         string
}

type PartnerOffer struct {
	ID             int
	PartnerID      int
	PartnerName    string
	Title          string
	Description    string
	DiscountType   string
	DiscountValue  int
	Code           string
	ValidUntil     string
	IsActive       bool
}

type UserTrip struct {
	ID              int
	UserID          int
	Title           string
	StartLocation   string
	EndLocation     string
	Waypoints       string
	DistanceKm      float64
	EstimatedTime   string
	CreatedAt       time.Time
}

// ==================== GLOBAL VARIABLES ====================

var db *sql.DB

// Rate limiting structure
var loginAttempts = make(map[string]struct {
	Count       int
	LastTry     time.Time
	LockedUntil time.Time
})

const (
	MinPasswordLength = 8
	MaxLoginAttempts  = 5
	LockoutDuration   = 15 * time.Minute
)

// ==================== HELPER FUNCTIONS ====================

func generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func validatePasswordStrength(password string) error {
	if len(password) < MinPasswordLength {
		return fmt.Errorf("password must be at least %d characters", MinPasswordLength)
	}
	
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)
	
	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return fmt.Errorf("password must contain uppercase, lowercase, number, and special character")
	}
	
	return nil
}

func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	return re.MatchString(strings.ToLower(email))
}

func calculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
    const R = 6371 // Earth's radius in kilometers
    
    lat1Rad := lat1 * math.Pi / 180
    lat2Rad := lat2 * math.Pi / 180
    deltaLat := (lat2 - lat1) * math.Pi / 180
    deltaLng := (lng2 - lng1) * math.Pi / 180
    
    a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
        math.Cos(lat1Rad)*math.Cos(lat2Rad)*
        math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
    c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
    
    return R * c
}

func generateTagHTML(tags string) template.HTML {
	if tags == "" {
		return ""
	}
	parts := strings.Split(tags, ",")
	var sb strings.Builder
	for _, t := range parts {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		color := "slate"
		switch t {
		case "Royal Enfield", "Harley", "KTM", "Ducati", "Triumph", "BMW", "Honda", "Yamaha", "Suzuki", "Bajaj", "TVS":
			color = "purple"
		case "Weekend Ride", "Long Trip", "Highway", "Mountain", "Coastal":
			color = "emerald"
		case "Beginner Friendly", "Experienced Only", "Pillion Available":
			color = "blue"
		case "Urgent", "Leaving Soon":
			color = "red"
		case "Gear Provided", "Petrol Share":
			color = "amber"
		}
		sb.WriteString(fmt.Sprintf(`<span class="inline-flex items-center px-3 py-1 text-xs font-bold rounded-2xl bg-%s-100 text-%s-700">%s</span>`, color, color, t))
	}
	return template.HTML(sb.String())
}

func createDefaultLogo() {
	logoPath := "./static/logos/logo.png"
	if _, err := os.Stat(logoPath); os.IsNotExist(err) {
		svgLogo := `<svg width="200" height="60" xmlns="http://www.w3.org/2000/svg">
			<rect width="200" height="60" fill="#e85d04" rx="10"/>
			<text x="100" y="38" font-size="24" text-anchor="middle" fill="white" font-weight="bold" font-family="Arial">🏍️ HC</text>
		</svg>`
		os.WriteFile(logoPath, []byte(svgLogo), 0644)
		log.Println("✅ Default logo created at", logoPath)
	}
}

func createSampleAdImages() {
	ads := map[string]string{
		"ad1.jpg": `<svg width="600" height="400" xmlns="http://www.w3.org/2000/svg">
			<rect width="600" height="400" fill="#e85d04"/>
			<text x="300" y="200" font-size="30" text-anchor="middle" fill="white" font-weight="bold">Royal Enfield</text>
			<text x="300" y="240" font-size="20" text-anchor="middle" fill="white">Accessories Sale</text>
		</svg>`,
		"ad2.jpg": `<svg width="800" height="200" xmlns="http://www.w3.org/2000/svg">
			<rect width="800" height="200" fill="#1a8cd8"/>
			<text x="400" y="100" font-size="35" text-anchor="middle" fill="white" font-weight="bold">RIDING GEAR SALE</text>
			<text x="400" y="140" font-size="20" text-anchor="middle" fill="white">Up to 50% OFF</text>
		</svg>`,
		"ad3.jpg": `<svg width="400" height="300" xmlns="http://www.w3.org/2000/svg">
			<rect width="400" height="300" fill="#10b981"/>
			<text x="200" y="150" font-size="25" text-anchor="middle" fill="white" font-weight="bold">Leh-Ladakh</text>
			<text x="200" y="190" font-size="18" text-anchor="middle" fill="white">Weekend Ride</text>
		</svg>`,
		"ad4.jpg": `<svg width="400" height="300" xmlns="http://www.w3.org/2000/svg">
			<rect width="400" height="300" fill="#ff6b35"/>
			<text x="200" y="150" font-size="30" text-anchor="middle" fill="white" font-weight="bold">HELMET STORE</text>
			<text x="200" y="190" font-size="18" text-anchor="middle" fill="white">30% OFF</text>
		</svg>`,
	}
	
	for filename, svg := range ads {
		path := "./static/ads/" + filename
		if _, err := os.Stat(path); os.IsNotExist(err) {
			os.WriteFile(path, []byte(svg), 0644)
		}
	}
}

func getCurrentUser(c *fiber.Ctx) *User {
	token := c.Cookies("auth_token")
	if token == "" {
		return nil
	}
	
	var userID int
	err := db.QueryRow(`
		SELECT user_id FROM sessions 
		WHERE token = ? AND expires_at > datetime('now')
	`, token).Scan(&userID)
	if err != nil {
		return nil
	}

	var u User
	var premiumUntil sql.NullTime
	err = db.QueryRow("SELECT id, username, handle, email, COALESCE(phone, ''), is_admin, credits, is_premium, COALESCE(is_active, 1), premium_until, membership_tier, COALESCE(bike_model, ''), COALESCE(riding_exp, ''), COALESCE(avatar_url, ''), COALESCE(is_verified, 0) FROM users WHERE id = ?", userID).
		Scan(&u.ID, &u.Username, &u.Handle, &u.Email, &u.Phone, &u.IsAdmin, &u.Credits, &u.IsPremium, &u.IsActive, &premiumUntil, &u.MembershipTier, &u.BikeModel, &u.RidingExp, &u.AvatarURL, &u.IsVerified)
	
	if err != nil {
		return nil
	}
	
	if !u.IsActive {
		return nil
	}
	
	if premiumUntil.Valid && premiumUntil.Time.After(time.Now()) {
		u.IsPremium = true
		u.PremiumUntil = &premiumUntil.Time
	} else if u.IsPremium {
		db.Exec("UPDATE users SET is_premium = false, membership_tier = 'free' WHERE id = ?", u.ID)
		u.IsPremium = false
	}
	
	return &u
}

func getAllNews() []BikingNews {
	rows, err := db.Query("SELECT id, title, content, category, likes, created_at FROM biking_news ORDER BY created_at DESC")
	if err != nil {
		log.Println("Error fetching all news:", err)
		return []BikingNews{}
	}
	defer rows.Close()

	var newsList []BikingNews
	for rows.Next() {
		var n BikingNews
		var createdAt time.Time
		rows.Scan(&n.ID, &n.Title, &n.Content, &n.Category, &n.Likes, &createdAt)
		diff := time.Since(createdAt)
		if diff < time.Hour {
			n.Timestamp = fmt.Sprintf("%d minutes ago", int(diff.Minutes()))
		} else if diff < 24*time.Hour {
			n.Timestamp = fmt.Sprintf("%d hours ago", int(diff.Hours()))
		} else {
			n.Timestamp = fmt.Sprintf("%d days ago", int(diff.Hours()/24))
		}
		newsList = append(newsList, n)
	}
	return newsList
}

func getBikingTrends() []BikingTrend {
	rows, err := db.Query("SELECT id, title, description, trend, percentage, category FROM biking_trends ORDER BY created_at DESC LIMIT 5")
	if err != nil {
		log.Println("Error fetching biking trends:", err)
		return []BikingTrend{}
	}
	defer rows.Close()

	var trends []BikingTrend
	for rows.Next() {
		var t BikingTrend
		rows.Scan(&t.ID, &t.Title, &t.Description, &t.Trend, &t.Percentage, &t.Category)
		trends = append(trends, t)
	}
	return trends
}

func getActiveAds(position string) []Advertisement {
	var rows *sql.Rows
	var err error
	
	if position == "" {
		rows, err = db.Query(`SELECT id, title, image_url, target_url, position, start_date, end_date, advertiser, is_active, impressions, clicks 
		          FROM advertisements WHERE is_active = 1 AND date('now') BETWEEN start_date AND end_date`)
	} else {
		rows, err = db.Query(`SELECT id, title, image_url, target_url, position, start_date, end_date, advertiser, is_active, impressions, clicks 
		          FROM advertisements WHERE is_active = 1 AND position = ? AND date('now') BETWEEN start_date AND end_date`, position)
	}
	
	if err != nil {
		log.Println("Error fetching ads:", err)
		return []Advertisement{}
	}
	defer rows.Close()

	var ads []Advertisement
	for rows.Next() {
		var a Advertisement
		rows.Scan(&a.ID, &a.Title, &a.ImageURL, &a.TargetURL, &a.Position, &a.StartDate, &a.EndDate, &a.Advertiser, &a.IsActive, &a.Impressions, &a.Clicks)
		ads = append(ads, a)
	}
	return ads
}

func getMarketplaceCategories() []MarketplaceCategory {
	rows, err := db.Query("SELECT id, name, icon, slug FROM marketplace_categories ORDER BY name")
	if err != nil {
		return []MarketplaceCategory{}
	}
	defer rows.Close()

	var categories []MarketplaceCategory
	for rows.Next() {
		var c MarketplaceCategory
		rows.Scan(&c.ID, &c.Name, &c.Icon, &c.Slug)
		categories = append(categories, c)
	}
	return categories
}

func getSiteSettings() SiteSettings {
	var settings SiteSettings
	err := db.QueryRow("SELECT id, site_name, logo_url, favicon_url, primary_color, secondary_color FROM site_settings LIMIT 1").
		Scan(&settings.ID, &settings.SiteName, &settings.LogoURL, &settings.FaviconURL, &settings.PrimaryColor, &settings.SecondaryColor)
	if err != nil {
		return SiteSettings{SiteName: "Highway Cruizzers", LogoURL: "/static/logos/logo.png", PrimaryColor: "#e85d04", SecondaryColor: "#d00000"}
	}
	return settings
}

func detectTags(content, category string) string {
	var detectedTags []string
	
	bikeKeywords := []string{"Royal Enfield", "Harley", "KTM", "Ducati", "Triumph", "BMW", "Honda", "Yamaha", "Suzuki", "Bajaj", "TVS"}
	rideKeywords := []string{"Weekend Ride", "Long Trip", "Highway", "Mountain", "Coastal", "Beginner Friendly", "Experienced Only", "Pillion Available", "Urgent", "Leaving Soon", "Gear Provided", "Petrol Share"}
	
	keywords := bikeKeywords
	if category == "Ride Buddy" || category == "Trip" {
		keywords = rideKeywords
	}
	
	for _, kw := range keywords {
		if strings.Contains(strings.ToUpper(content), strings.ToUpper(kw)) {
			detectedTags = append(detectedTags, kw)
		}
	}
	
	return strings.Join(detectedTags, ",")
}

func fetchRides(query string, args ...interface{}) ([]Ride, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rides []Ride
	for rows.Next() {
		var r Ride
		var boosted, featured int
		var createdAt time.Time
		err := rows.Scan(&r.ID, &r.User, &r.Handle, &r.Content, &r.Price, &r.Category, &r.Tags, &r.ImageURL, &r.Likes, &createdAt, &boosted, &featured, &r.UserID, &r.Status, &r.FromLocation, &r.ToLocation, &r.DepartureDate, &r.BikeModel, &r.Seats)
		if err != nil {
			continue
		}
		r.IsBoosted = boosted == 1
		r.IsFeatured = featured == 1
		r.CreatedAt = createdAt
		rides = append(rides, r)
	}
	return rides, nil
}

func formatRSSDate(dateStr string) string {
	layouts := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822,
		"Mon, 02 Jan 2006 15:04:05 -0700",
		"2006-01-02T15:04:05Z",
	}
	
	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			diff := time.Since(t)
			if diff < time.Hour {
				return fmt.Sprintf("%d minutes ago", int(diff.Minutes()))
			} else if diff < 24*time.Hour {
				return fmt.Sprintf("%d hours ago", int(diff.Hours()))
			}
			return t.Format("Jan 02, 2006")
		}
	}
	return "Recently"
}

func fetchRSSNews() ([]BikingNews, error) {
	feeds := []string{
		"https://www.rushlane.com/feed",
		"https://www.bikewale.com/news/feeds",
		"https://www.zigwheels.com/newsfeeds",
		"https://www.motorbeam.com/feed",
		"https://indianautosblog.com/feed",
		"https://www.news18.com/tag/motorcycles/feed",
	}
	
	var allNews []BikingNews
	
	for _, feedURL := range feeds {
		resp, err := http.Get(feedURL)
		if err != nil {
			log.Println("Error fetching feed:", feedURL, err)
			continue
		}
		defer resp.Body.Close()
		
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			continue
		}
		
		var feed RSSFeed
		err = xml.Unmarshal(body, &feed)
		if err != nil {
			continue
		}
		
		for _, item := range feed.Channel.Items {
			category := "news"
			titleLower := strings.ToLower(item.Title)
			descLower := strings.ToLower(item.Description)
			
			if strings.Contains(titleLower, "review") || strings.Contains(descLower, "review") {
				category = "reviews"
			} else if strings.Contains(titleLower, "electric") || strings.Contains(descLower, "electric") {
				category = "electric"
			} else if strings.Contains(titleLower, "safety") || strings.Contains(descLower, "safety") {
				category = "safety"
			} else if strings.Contains(titleLower, "event") || strings.Contains(descLower, "rally") {
				category = "events"
			} else if strings.Contains(titleLower, "launch") || strings.Contains(descLower, "launch") {
				category = "bikes"
			}
			
			desc := regexp.MustCompile(`<[^>]*>`).ReplaceAllString(item.Description, "")
			if len(desc) > 300 {
				desc = desc[:300] + "..."
			}
			if desc == "" {
				desc = item.Title
			}
			
			news := BikingNews{
				Title:     item.Title,
				Content:   desc,
				Timestamp: formatRSSDate(item.PubDate),
				Category:  category,
				Likes:     0,
			}
			allNews = append(allNews, news)
		}
	}
	
	return allNews, nil
}

// ==================== DATABASE INITIALIZATION ====================

func initDB() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ No .env file found, using environment variables")
	}

	dbURL := os.Getenv("TURSO_DB_URL")
	authToken := os.Getenv("TURSO_AUTH_TOKEN")

	var err error
	
	if dbURL != "" && authToken != "" {
		log.Println("🔗 Connecting to Turso database...")
		db, err = sql.Open("libsql", dbURL+"?authToken="+authToken)
		if err != nil {
			log.Fatal("Error connecting to Turso:", err)
		}
		log.Println("✅ Connected to Turso successfully!")
	} else {
		log.Println("📁 Using SQLite for local development")
		db, err = sql.Open("sqlite3", "./highwaycruizzers.db")
		if err != nil {
			log.Fatal("Error opening SQLite:", err)
		}
	}

	if err := db.Ping(); err != nil {
		log.Fatal("Database ping failed:", err)
	}

	createTablesSQL := `
-- ==================== ALL TABLES ====================

CREATE TABLE IF NOT EXISTS rides (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER DEFAULT 0,
    username TEXT,
    handle TEXT,
    content TEXT,
    price TEXT,
    category TEXT,
    tags TEXT,
    image_url TEXT,
    from_location TEXT,
    to_location TEXT,
    departure_date TEXT,
    bike_model TEXT,
    seats INTEGER DEFAULT 1,
    likes INTEGER DEFAULT 0,
    status TEXT DEFAULT 'pending',
    boost_until DATETIME,
    featured_until DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE,
    handle TEXT UNIQUE,
    email TEXT UNIQUE,
    phone TEXT DEFAULT '',
    password_hash TEXT,
    is_admin BOOLEAN DEFAULT FALSE,
    credits INTEGER DEFAULT 500,
    is_premium BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    premium_until DATETIME,
    membership_tier TEXT DEFAULT 'free',
    bike_model TEXT DEFAULT '',
    riding_exp TEXT DEFAULT '',
    avatar_url TEXT DEFAULT '',
    is_verified BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS user_profiles (
    user_id INTEGER PRIMARY KEY,
    bio TEXT,
    location TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS biking_news (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT,
    content TEXT,
    category TEXT,
    likes INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS biking_trends (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT,
    description TEXT,
    trend TEXT,
    percentage TEXT,
    category TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS transactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    amount INTEGER,
    type TEXT,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS password_resets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT,
    token TEXT,
    expires_at DATETIME,
    used BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS ride_requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ride_id INTEGER,
    rider_id INTEGER,
    rider_name TEXT,
    rider_email TEXT,
    rider_phone TEXT,
    message TEXT,
    status TEXT DEFAULT 'pending',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(ride_id) REFERENCES rides(id)
);

CREATE TABLE IF NOT EXISTS products (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    user_name TEXT,
    title TEXT,
    description TEXT,
    price REAL,
    category TEXT,
    condition TEXT,
    image_url TEXT,
    location TEXT,
    status TEXT DEFAULT 'available',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS advertisements (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT,
    image_url TEXT,
    target_url TEXT,
    position TEXT,
    start_date DATETIME,
    end_date DATETIME,
    advertiser TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    impressions INTEGER DEFAULT 0,
    clicks INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS marketplace_categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT,
    icon TEXT,
    slug TEXT
);

CREATE TABLE IF NOT EXISTS site_settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    site_name TEXT,
    logo_url TEXT,
    favicon_url TEXT,
    primary_color TEXT,
    secondary_color TEXT
);

CREATE TABLE IF NOT EXISTS rider_verifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER UNIQUE,
    dl_number TEXT,
    bike_rc_number TEXT,
    status TEXT DEFAULT 'pending',
    verified_at DATETIME,
    verified_by INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS security_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_type TEXT,
    user_id INTEGER,
    email TEXT,
    ip_address TEXT,
    user_agent TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    token TEXT UNIQUE,
    user_id INTEGER,
    ip_address TEXT,
    user_agent TEXT,
    expires_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS ride_safety_reports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ride_id INTEGER,
    reporter_id INTEGER,
    description TEXT,
    severity TEXT,
    status TEXT DEFAULT 'pending',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS experiences (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    category TEXT NOT NULL,
    description TEXT NOT NULL,
    duration_days INTEGER,
    duration_nights INTEGER,
    price INTEGER,
    discounted_price INTEGER,
    max_people INTEGER,
    min_people INTEGER DEFAULT 1,
    location TEXT,
    start_location TEXT,
    end_location TEXT,
    vehicle_type TEXT,
    is_featured BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    cover_image TEXT,
    rating REAL DEFAULT 0,
    total_reviews INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS experience_bookings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    experience_id INTEGER,
    user_id INTEGER,
    travel_date DATE,
    number_of_people INTEGER,
    total_price INTEGER,
    special_requests TEXT,
    status TEXT DEFAULT 'pending',
    contact_name TEXT,
    contact_phone TEXT,
    contact_email TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(experience_id) REFERENCES experiences(id),
    FOREIGN KEY(user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS points_of_interest (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    category TEXT,
    latitude REAL NOT NULL,
    longitude REAL NOT NULL,
    address TEXT,
    city TEXT,
    state TEXT,
    pincode TEXT,
    phone TEXT,
    email TEXT,
    website TEXT,
    price_range TEXT,
    rating REAL DEFAULT 0,
    total_reviews INTEGER DEFAULT 0,
    amenities TEXT,
    images TEXT,
    opening_time TEXT,
    closing_time TEXT,
    is_24x7 BOOLEAN DEFAULT FALSE,
    is_partner BOOLEAN DEFAULT FALSE,
    discount_percentage INTEGER DEFAULT 0,
    offer_details TEXT,
    distance_km REAL DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS partner_offers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    partner_id INTEGER,
    title TEXT NOT NULL,
    description TEXT,
    discount_type TEXT,
    discount_value INTEGER,
    min_purchase INTEGER DEFAULT 0,
    code TEXT,
    valid_from DATETIME,
    valid_to DATETIME,
    is_active BOOLEAN DEFAULT TRUE,
    terms TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(partner_id) REFERENCES points_of_interest(id)
);

CREATE TABLE IF NOT EXISTS user_saved_places (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    poi_id INTEGER,
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES users(id),
    FOREIGN KEY(poi_id) REFERENCES points_of_interest(id),
    UNIQUE(user_id, poi_id)
);

CREATE TABLE IF NOT EXISTS user_trips (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    title TEXT,
    start_location TEXT,
    end_location TEXT,
    waypoints TEXT,
    distance_km REAL,
    estimated_time TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS partner_offers_claimed (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    offer_id INTEGER,
    user_id INTEGER,
    claimed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    used_at DATETIME,
    status TEXT DEFAULT 'claimed',
    FOREIGN KEY(offer_id) REFERENCES partner_offers(id),
    FOREIGN KEY(user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS user_location_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    latitude REAL,
    longitude REAL,
    accuracy REAL,
    recorded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS user_follows (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    follower_id INTEGER,
    following_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(follower_id, following_id),
    FOREIGN KEY(follower_id) REFERENCES users(id),
    FOREIGN KEY(following_id) REFERENCES users(id)
);


`

	_, err = db.Exec(createTablesSQL)
	if err != nil {
		log.Fatal("Error creating tables:", err)
	}
	log.Println("✅ Tables created/verified")

	// Add missing columns
	db.Exec(`ALTER TABLE users ADD COLUMN avatar_url TEXT DEFAULT ''`)
	db.Exec(`ALTER TABLE users ADD COLUMN is_verified BOOLEAN DEFAULT FALSE`)
	db.Exec(`ALTER TABLE products ADD COLUMN user_name TEXT DEFAULT ''`)
	db.Exec(`ALTER TABLE points_of_interest ADD COLUMN is_active BOOLEAN DEFAULT TRUE`)

	// Create directories
	os.MkdirAll("./static/logos", 0755)
	os.MkdirAll("./static/ads", 0755)

	// Create default logo
	createDefaultLogo()
	createSampleAdImages()

	// Seed admin user
	var adminCount int
	db.QueryRow("SELECT COUNT(*) FROM users WHERE is_admin = 1").Scan(&adminCount)
	if adminCount == 0 {
		hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		_, err = db.Exec(`INSERT INTO users (username, handle, email, phone, password_hash, is_admin, credits, is_premium, membership_tier, bike_model, riding_exp, avatar_url, is_verified) 
		         VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`, 
			"Highway Cruizzers Admin", "highwayadmin", "admin@highwaycruizzers.com", "9999999999", hash, true, 9999, true, "premium", "Royal Enfield", "10+ years", "", true)
		if err != nil {
			log.Println("Warning: Could not create admin user:", err)
		} else {
			log.Println("✅ Admin created: admin@highwaycruizzers.com / admin123")
		}
	}

	// Seed site settings
	var settingsCount int
	db.QueryRow("SELECT COUNT(*) FROM site_settings").Scan(&settingsCount)
	if settingsCount == 0 {
		db.Exec(`INSERT INTO site_settings (site_name, logo_url, primary_color, secondary_color) VALUES (?, ?, ?, ?)`,
			"Highway Cruizzers", "/static/logos/logo.png", "#e85d04", "#d00000")
		log.Println("✅ Default site settings created")
	}

	// Seed news
	var newsCount int
	db.QueryRow("SELECT COUNT(*) FROM biking_news").Scan(&newsCount)
	if newsCount == 0 {
		newsData := []struct {
			title, content, category string
		}{
			{"🏍️ Royal Enfield Guerrilla 450 Launched in India at ₹2.39 Lakh", "Royal Enfield has launched the all-new Guerrilla 450 in India. Powered by the new 452cc liquid-cooled engine producing 40 bhp and 40 Nm torque. Bookings open across all dealerships.", "bikes"},
			{"⚡ Ola Roadster Electric Motorcycle Revealed - Launching August 2025", "Ola Electric has unveiled the Roadster electric motorcycle with a claimed range of 200 km. Features include a 6kW motor, digital display, and fast charging capability.", "electric"},
			{"🏁 MotoGP India 2025 Dates Announced - Buddh Circuit to Host Race", "The 2025 MotoGP Indian Grand Prix will take place from September 26-28 at the Buddh International Circuit. Early bird tickets available from next week.", "events"},
			{"⭐ TVS Apache RTR 310 Review: The Most Feature-Packed Apache Yet", "We spent a week with the new Apache RTR 310. With 35 bhp, cruise control, and 6 riding modes, this is TVS's most advanced motorcycle to date.", "reviews"},
			{"🛡️ New Helmet Safety Standards: What Riders Need to Know", "The Ministry of Road Transport has mandated new ISI standards for helmets from October 2025. Here's what changed and how to choose a safe helmet.", "safety"},
			{"🏍️ Triumph Speed 400 vs Harley Davidson X440 Comparison", "Two of the most popular mid-capacity roadsters go head-to-head. We compare performance, features, and value for money to help you choose.", "reviews"},
			{"⚡ Ultraviolette F77 Mach 2 Launched with 307 km Range", "Ultraviolette has launched the F77 Mach 2 with improved range and performance. The new variant offers 307 km in Eco mode and 0-100 kmph in 7 seconds.", "electric"},
			{"🏁 India Bike Week 2025: Complete Guide to Asia's Largest Biking Festival", "Everything you need to know about IBW 2025 happening in Goa from December 5-7. Workshops, stunt shows, new launches, and celebrity appearances.", "events"},
			{"🏍️ Bajaj CNG Bike Launch Date Confirmed for August 2025", "Bajaj Auto has confirmed the launch date for world's first CNG motorcycle. Expected price around ₹70,000 with running cost of just 1 rupee per km.", "bikes"},
			{"🛡️ 10 Essential Safety Checks Before Your Monsoon Road Trip", "Planning a monsoon ride? Don't miss these 10 critical safety checks including tire tread, brake pads, chain lubrication, and rain gear preparation.", "safety"},
			{"⭐ Honda CB350 RS Long Term Review: 6 Months of Ownership", "Our long-term review of the Honda CB350 RS after 6 months and 5,000 km. Real-world mileage, service costs, and everything you need to know.", "reviews"},
			{"⚡ Ather 450X Gets Big Price Cut - New Starting Price ₹1.05 Lakh", "Ather Energy has reduced prices of the 450X by up to ₹20,000. The move comes ahead of festive season to compete with Ola and Bajaj Chetak.", "electric"},
			{"🏁 Himalayan Rally 2025: Registrations Open for World's Toughest Biker Rally", "The annual Himalayan Motorcycle Rally returns with a new route through Leh-Ladakh. Registrations limited to 100 riders. Early bird discount available.", "events"},
			{"🏍️ Yamaha R3 vs KTM RC 390: Which Sportbike Should You Buy?", "Detailed comparison between the new Yamaha R3 and KTM RC 390. Track performance, street comfort, and maintenance costs compared.", "reviews"},
			{"🛡️ Road Safety Tips: How to Ride Safely on Indian Highways", "Expert tips for safe highway riding including proper lane discipline, overtaking techniques, and handling emergency situations on Indian roads.", "safety"},
		}
		
		for _, news := range newsData {
			_, err := db.Exec(`INSERT INTO biking_news (title, content, category, likes, created_at) 
				VALUES (?, ?, ?, ?, datetime('now'))`,
				news.title, news.content, news.category, 0)
			if err != nil {
				log.Println("Error seeding news:", err)
			}
		}
		log.Println("✅ Seeded 15+ real biking news articles")
	}

	// Seed trends
	var trendsCount int
	db.QueryRow("SELECT COUNT(*) FROM biking_trends").Scan(&trendsCount)
	if trendsCount == 0 {
		trendsData := []struct {
			title, description, trend, percentage, category string
		}{
			{"Leh-Ladakh", "Most popular riding destination this season", "up", "+45%", "destination"},
			{"Adventure Touring", "Adventure bike sales increasing", "up", "+32%", "trending"},
			{"Weekend Rides", "Group rides gaining popularity", "up", "+28%", "activity"},
			{"Vintage Restoration", "Classic bike restoration trending", "stable", "+15%", "hobby"},
			{"Safety Gear", "Premium riding gear demand rising", "up", "+40%", "gear"},
		}
		for _, trend := range trendsData {
			db.Exec("INSERT INTO biking_trends (title, description, trend, percentage, category) VALUES (?, ?, ?, ?, ?)",
				trend.title, trend.description, trend.trend, trend.percentage, trend.category)
		}
		log.Println("✅ Seeded biking trends")
	}

	// Seed marketplace categories
	var catCount int
	db.QueryRow("SELECT COUNT(*) FROM marketplace_categories").Scan(&catCount)
	if catCount == 0 {
		categories := []struct{ name, icon, slug string }{
			{"Helmets", "🪖", "helmets"},
			{"Riding Jackets", "🧥", "jackets"},
			{"Gloves", "🧤", "gloves"},
			{"Boots", "👢", "boots"},
			{"Bike Parts", "🔧", "parts"},
			{"Accessories", "🎒", "accessories"},
		}
		for _, cat := range categories {
			db.Exec("INSERT INTO marketplace_categories (name, icon, slug) VALUES (?, ?, ?)",
				cat.name, cat.icon, cat.slug)
		}
		log.Println("✅ Seeded marketplace categories")
	}

	// Seed sample advertisements
	var adCount int
	db.QueryRow("SELECT COUNT(*) FROM advertisements").Scan(&adCount)
	if adCount == 0 {
		ads := []struct {
			title, imageURL, targetURL, position, advertiser string
		}{
			{"Royal Enfield Accessories", "/static/ads/ad1.jpg", "https://royalenfield.com", "sidebar", "Royal Enfield"},
			{"Premium Riding Gear Sale", "/static/ads/ad2.jpg", "https://example.com/gear", "banner", "RidingGear Pro"},
			{"Weekend Ride to Leh", "/static/ads/ad3.jpg", "https://example.com/leh-tour", "featured", "Biking Tours India"},
			{"Helmet Store - 30% OFF", "/static/ads/ad4.jpg", "https://example.com/helmets", "sidebar", "SafeRide Helmets"},
		}
		for _, ad := range ads {
			db.Exec(`INSERT INTO advertisements (title, image_url, target_url, position, advertiser, start_date, end_date, is_active) 
			         VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now', '+30 days'), 1)`,
				ad.title, ad.imageURL, ad.targetURL, ad.position, ad.advertiser)
		}
		log.Println("✅ Seeded sample advertisements")
	}

	// Seed sample products
	var productCount int
	db.QueryRow("SELECT COUNT(*) FROM products").Scan(&productCount)
	if productCount == 0 {
		products := []struct {
			title, description, category, condition, imageURL, location string
			price float64
		}{
			{"Royal Enfield Helmet", "Premium full-face helmet with DOT certification", "Helmets", "Like New", "/static/logo.png", "Mumbai", 3499},
			{"Riding Jacket", "Waterproof riding jacket with armor", "Riding Jackets", "New", "/static/logo.png", "Delhi", 5999},
			{"Bike Cover", "Waterproof bike cover for Royal Enfield", "Accessories", "New", "/static/logo.png", "Bangalore", 899},
			{"Handlebar Grips", "Premium rubber grips for better control", "Bike Parts", "New", "/static/logo.png", "Pune", 499},
		}
		for _, p := range products {
			db.Exec(`INSERT INTO products (user_id, user_name, title, description, price, category, condition, image_url, location, status) 
			         VALUES (1, 'Highway Cruizzers Admin', ?, ?, ?, ?, ?, ?, ?, 'available')`,
				p.title, p.description, p.price, p.category, p.condition, p.imageURL, p.location)
		}
		log.Println("✅ Seeded sample products")
	}

	// Seed Points of Interest
	var poiCount int
	db.QueryRow("SELECT COUNT(*) FROM points_of_interest").Scan(&poiCount)
	if poiCount == 0 {
		log.Println("Seeding points of interest...")
		
		pois := []struct {
			name, poiType, category, address, city, phone, priceRange string
			lat, lng float64
			rating float64
			discount int
			offerDetails string
		}{
			{"Hotel Himalayan Gateway", "hotel", "luxury", "NH 44, Near Bus Stand", "Manali", "9812345678", "₹₹₹", 32.2432, 77.1896, 4.5, 10, "10% discount for members"},
			{"Zostel Manali", "hotel", "budget", "Old Manali", "Manali", "9812345679", "₹", 32.2420, 77.1880, 4.3, 15, "15% off on dorm beds"},
			{"JW Marriott Chandigarh", "hotel", "luxury", "Sector 35", "Chandigarh", "9812345680", "₹₹₹₹", 30.7333, 76.7794, 4.7, 0, ""},
			{"Johnson's Cafe", "restaurant", "cafe", "Club House Road", "Manali", "9812345681", "₹₹", 32.2410, 77.1870, 4.6, 10, "10% off on total bill"},
			{"The Lazy Dog", "restaurant", "cafe", "Old Manali", "Manali", "9812345682", "₹₹", 32.2405, 77.1865, 4.4, 5, "5% discount on food"},
			{"Dhaba 29", "restaurant", "dhaba", "NH 44", "Kullu", "9812345683", "₹", 31.9580, 77.1100, 4.2, 0, ""},
			{"Himalayan Spa & Wellness", "relax_zone", "spa", "Mall Road", "Manali", "9812345684", "₹₹₹", 32.2430, 77.1900, 4.5, 20, "20% off on spa treatments"},
			{"Yoga House", "relax_zone", "yoga", "Old Manali", "Manali", "9812345685", "₹₹", 32.2415, 77.1860, 4.4, 15, "15% off on yoga classes"},
			{"Royal Enfield Service Center", "service_center", "bike", "NH 44, Near Petrol Pump", "Manali", "9812345686", "₹₹", 32.2440, 77.1910, 4.3, 5, "5% discount on service"},
			{"Bike Point Service", "service_center", "multi-brand", "Mall Road", "Kullu", "9812345687", "₹₹", 31.9600, 77.1110, 4.1, 10, "10% off on all services"},
			{"Tyre Pro", "service_center", "tyres", "NH 44", "Mandi", "9812345688", "₹₹", 31.7100, 76.9300, 4.2, 0, ""},
			{"EV Charging Station", "charging_point", "electric", "HP Petrol Pump, NH 44", "Manali", "9812345689", "₹", 32.2450, 77.1920, 4.0, 5, "5% discount on charging"},
			{"Green EV Charging", "charging_point", "electric", "Near Bus Stand", "Kullu", "9812345690", "₹", 31.9590, 77.1120, 4.1, 0, ""},
		}
		
		for _, poi := range pois {
			_, err := db.Exec(`INSERT INTO points_of_interest (name, type, category, address, city, phone, 
			          price_range, latitude, longitude, rating, is_partner, discount_percentage, offer_details, is_active) 
			          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1)`,
				poi.name, poi.poiType, poi.category, poi.address, poi.city, poi.phone,
				poi.priceRange, poi.lat, poi.lng, poi.rating, 
				poi.discount > 0, poi.discount, poi.offerDetails)
			if err != nil {
				log.Println("Error seeding POI:", err)
			}
		}
		log.Println("✅ Seeded 13+ points of interest")
	}

	// Seed sample experiences
	var expCount int
	db.QueryRow("SELECT COUNT(*) FROM experiences").Scan(&expCount)
	if expCount == 0 {
		experiences := []struct {
			title, category, description string
			durationDays, durationNights int
			price, discountedPrice int
			maxPeople, minPeople int
			location, startLocation, endLocation, vehicleType string
			isFeatured bool
		}{
			{"Royal Rajasthan Caravan Tour", "family", "Explore the royal heritage of Rajasthan in a luxury caravan. Visit Jaipur, Jodhpur, Udaipur, and Jaisalmer with complete comfort.", 7, 6, 85000, 75000, 6, 2, "Rajasthan", "Jaipur", "Jaisalmer", "Luxury Caravan", true},
			{"Kerala Backwaters Caravan", "family", "Experience the serene backwaters of Kerala in a fully-equipped caravan. Perfect for family vacations.", 5, 4, 65000, 55000, 6, 2, "Kerala", "Kochi", "Trivandrum", "Premium Caravan", true},
			{"Himalayan Circuit", "biker", "Epic bike trip through the Himalayas. Cover Manali, Leh, and Pangong Lake.", 12, 11, 45000, 39999, 15, 4, "Himachal", "Manali", "Leh", "Royal Enfield", true},
			{"Goa Beach Caravan", "group", "Beach hopping caravan tour across Goa's best beaches. Perfect for groups of friends.", 4, 3, 35000, 29999, 8, 4, "Goa", "North Goa", "South Goa", "Standard Caravan", false},
			{"Golden Triangle Tour", "family", "Delhi-Agra-Jaipur tour in luxury caravan. Visit Taj Mahal, Amer Fort, and more.", 6, 5, 75000, 65000, 6, 2, "North India", "Delhi", "Jaipur", "Luxury Caravan", true},
			{"Western Ghats Expedition", "biker", "Thrilling bike ride through the scenic Western Ghats. Includes camping and local experiences.", 5, 4, 28000, 24999, 12, 3, "Karnataka", "Bangalore", "Coorg", "Adventure Bike", false},
		}
		for _, exp := range experiences {
			db.Exec(`INSERT INTO experiences (title, category, description, duration_days, duration_nights, price, discounted_price, max_people, min_people, location, start_location, end_location, vehicle_type, is_featured) 
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				exp.title, exp.category, exp.description, exp.durationDays, exp.durationNights,
				exp.price, exp.discountedPrice, exp.maxPeople, exp.minPeople, exp.location,
				exp.startLocation, exp.endLocation, exp.vehicleType, exp.isFeatured)
		}
		log.Println("✅ Seeded sample experiences")
	}
}

// ==================== MAIN FUNCTION ====================

func main() {
	initDB()
	defer db.Close()

	// Setup template engine
	engine := html.New("./views", ".html")
	engine.AddFunc("tagHTML", generateTagHTML)
	engine.Reload(true)

	app := fiber.New(fiber.Config{
		Views: engine,
		ViewsLayout: "",
	})

	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New())

	// Security headers middleware
	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		return c.Next()
	})

	// Static files
	app.Static("/static", "./static")

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Seed endpoint
	app.Get("/seed", func(c *fiber.Ctx) error {
		db.Exec("DELETE FROM rides")
		
		dummyData := []struct {
			user, handle, content, price, category, tags, image string
			boosted, featured bool
			userID int
			fromLocation, toLocation, departureDate, bikeModel string
			seats int
		}{
			{"Biker Raj", "bikerraj", "Leh-Ladakh trip next week. Royal Enfield Classic 350. Looking for riding partner.", "Petrol Share", "Ride Buddy", "Royal Enfield,Long Trip", "https://picsum.photos/id/104/800/450", false, true, 1, "Delhi", "Leh", "2024-06-15", "Royal Enfield Classic 350", 2},
			{"Mountain Rider", "mountainrider", "Weekend ride to Kasauli. Experienced riders only.", "₹500", "Trip", "Weekend Ride,Experienced Only", "https://picsum.photos/id/107/800/450", true, false, 1, "Chandigarh", "Kasauli", "2024-05-20", "Ducati Scrambler", 3},
			{"Bike Rental Co", "bikerental", "Royal Enfield Himalayan available for rent. Well maintained.", "₹1500/day", "Rental", "Royal Enfield,Gear Provided", "https://picsum.photos/id/111/800/450", false, false, 1, "Manali", "Local", "", "Royal Enfield Himalayan", 1},
			{"Service Pro", "servicepro", "Professional bike servicing and custom modifications.", "Contact Owner", "Service", "Maintenance,Custom", "https://picsum.photos/id/119/800/450", false, false, 1, "", "", "", "", 0},
		}

		for _, d := range dummyData {
			query := `INSERT INTO rides (user_id, username, handle, content, price, category, tags, image_url, status, from_location, to_location, departure_date, bike_model, seats`
			args := []interface{}{d.userID, d.user, d.handle, d.content, d.price, d.category, d.tags, d.image, "approved", d.fromLocation, d.toLocation, d.departureDate, d.bikeModel, d.seats}
			
			if d.boosted {
				query += ", boost_until"
				args = append(args, time.Now().Add(7*24*time.Hour))
			}
			if d.featured {
				query += ", featured_until"
				args = append(args, time.Now().Add(24*time.Hour))
			}
			query += ") VALUES (" + strings.Repeat("?,", len(args)) + "?)"
			query = strings.Replace(query, ",?)", ")", 1)
			
			db.Exec(query, args...)
		}
		return c.SendString("✅ Dummy ride data loaded! <a href='/'>Go Home</a>")
	})

	// ==================== TERMS & PRIVACY PAGES ====================
	
	app.Get("/terms", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		settings := getSiteSettings()
		return c.Render("terms", fiber.Map{
			"CurrentUser": currentUser,
			"Settings":    settings,
		})
	})

	app.Get("/privacy", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		settings := getSiteSettings()
		return c.Render("privacy", fiber.Map{
			"CurrentUser": currentUser,
			"Settings":    settings,
		})
	})

	// ==================== RIDER VERIFICATION ====================
	
	app.Get("/verify", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil {
			return c.Redirect("/")
		}
		
		var verification RiderVerification
		db.QueryRow("SELECT id, dl_number, bike_rc_number, status FROM rider_verifications WHERE user_id = ?", currentUser.ID).
			Scan(&verification.ID, &verification.DLNumber, &verification.BikeRCNumber, &verification.Status)
		
		settings := getSiteSettings()
		return c.Render("verify", fiber.Map{
			"CurrentUser":  currentUser,
			"Verification": verification,
			"Settings":     settings,
		})
	})

	app.Post("/verify/submit", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil {
			return c.Status(401).SendString("Please login")
		}
		
		dlNumber := c.FormValue("dl_number")
		bikeRCNumber := c.FormValue("bike_rc_number")
		
		if len(dlNumber) < 5 {
			return c.Status(400).SendString("Invalid driver's license number")
		}
		
		_, err := db.Exec(`INSERT INTO rider_verifications (user_id, dl_number, bike_rc_number, status) 
		                   VALUES (?, ?, ?, 'pending') 
		                   ON CONFLICT(user_id) DO UPDATE SET 
		                   dl_number = excluded.dl_number, 
		                   bike_rc_number = excluded.bike_rc_number,
		                   status = 'pending'`,
			currentUser.ID, dlNumber, bikeRCNumber)
		
		if err != nil {
			return c.Status(500).SendString("Failed to submit verification")
		}
		
		return c.SendString(`<div class="p-4 text-emerald-600">✅ Verification submitted! Admin will review within 24 hours.<script>setTimeout(function(){ window.location.href = "/profile"; }, 2000);</script></div>`)
	})

	// ==================== SAFETY CHECK & REPORTING ====================
	
	app.Get("/safety-check/:rideId", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil {
			return c.Redirect("/")
		}
		
		rideID := c.Params("rideId")
		settings := getSiteSettings()
		
		return c.Render("safety-check", fiber.Map{
			"CurrentUser": currentUser,
			"RideID":      rideID,
			"Settings":    settings,
		})
	})

	app.Post("/safety-check/:rideId", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil {
			return c.Status(401).JSON(fiber.Map{"error": "Please login"})
		}
		
		rideID := c.Params("rideId")
		helmet := c.FormValue("helmet") == "on"
		
		if !helmet {
			return c.Status(400).JSON(fiber.Map{"error": "Helmet is mandatory for safety"})
		}
		
		_, err := db.Exec(`INSERT INTO ride_safety_reports (ride_id, reporter_id, description, severity) 
		                   VALUES (?, ?, ?, ?)`,
			rideID, currentUser.ID, "Safety check passed: Helmet confirmed, Documents ready", "low")
		
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to save safety check"})
		}
		
		return c.JSON(fiber.Map{
			"success": true,
			"message": "Safety check completed! You can now join the ride.",
		})
	})

	app.Post("/report/:rideId", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil {
			return c.Status(401).JSON(fiber.Map{"error": "Please login"})
		}
		
		rideID := c.Params("rideId")
		description := c.FormValue("description")
		severity := c.FormValue("severity")
		
		_, err := db.Exec(`INSERT INTO ride_safety_reports (ride_id, reporter_id, description, severity) 
		                   VALUES (?, ?, ?, ?)`,
			rideID, currentUser.ID, description, severity)
		
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to submit report"})
		}
		
		return c.JSON(fiber.Map{
			"success": true,
			"message": "Report submitted. Our team will investigate.",
		})
	})

	// ==================== HOME PAGE ====================
	
	app.Get("/", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		bikingTrends := getBikingTrends()
		sidebarAds := getActiveAds("sidebar")
		bannerAds := getActiveAds("banner")
		featuredAds := getActiveAds("featured")
		settings := getSiteSettings()

		rides, _ := fetchRides(`
			SELECT id, username, handle, content, price, category, tags, image_url, likes, created_at,
			       CASE WHEN boost_until > datetime('now') THEN 1 ELSE 0 END,
			       CASE WHEN featured_until > datetime('now') THEN 1 ELSE 0 END,
			       user_id, status, COALESCE(from_location, ''), COALESCE(to_location, ''), COALESCE(departure_date, ''), COALESCE(bike_model, ''), COALESCE(seats, 0)
			FROM rides 
			WHERE status = 'approved' 
			ORDER BY featured_until DESC, boost_until DESC, id DESC
			LIMIT 20
		`)

		return c.Render("index", fiber.Map{
			"Rides":         rides,
			"CurrentUser":   currentUser,
			"IsProfile":     false,
			"SearchQuery":   "",
			"IsAdmin":       currentUser != nil && currentUser.IsAdmin,
			"BikingTrends":  bikingTrends,
			"SidebarAds":    sidebarAds,
			"BannerAds":     bannerAds,
			"FeaturedAds":   featuredAds,
			"Settings":      settings,
		})
	})

	// ==================== EXPERIENCES PAGE ====================

	app.Get("/experiences", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		settings := getSiteSettings()
		bikingTrends := getBikingTrends()
		sidebarAds := getActiveAds("sidebar")
		
		rows, err := db.Query(`
			SELECT id, title, category, description, duration_days, duration_nights, 
				   price, discounted_price, max_people, min_people, location, 
				   vehicle_type, is_featured, rating, total_reviews, cover_image
			FROM experiences WHERE is_active = 1 ORDER BY is_featured DESC, created_at DESC
		`)
		if err != nil {
			return c.Status(500).SendString("Error loading experiences")
		}
		defer rows.Close()
		
		type ExperienceDisplay struct {
			ID             int
			Title          string
			Category       string
			Description    string
			DurationDays   int
			DurationNights int
			Price          int
			DiscountedPrice int
			MaxPeople      int
			MinPeople      int
			Location       string
			VehicleType    string
			IsFeatured     bool
			Rating         float64
			TotalReviews   int
			CoverImage     string
		}
		
		var experiences []ExperienceDisplay
		for rows.Next() {
			var e ExperienceDisplay
			var coverImage sql.NullString
			rows.Scan(&e.ID, &e.Title, &e.Category, &e.Description, &e.DurationDays, &e.DurationNights,
				&e.Price, &e.DiscountedPrice, &e.MaxPeople, &e.MinPeople, &e.Location,
				&e.VehicleType, &e.IsFeatured, &e.Rating, &e.TotalReviews, &coverImage)
			if coverImage.Valid {
				e.CoverImage = coverImage.String
			}
			experiences = append(experiences, e)
		}
		
		categories := []string{"all", "family", "group", "biker"}
		
		return c.Render("experiences", fiber.Map{
			"Title":         "Travel Experiences - " + settings.SiteName,
			"CurrentUser":   currentUser,
			"Experiences":   experiences,
			"Categories":    categories,
			"Settings":      settings,
			"BikingTrends":  bikingTrends,
			"SidebarAds":    sidebarAds,
			"IsAdmin":       currentUser != nil && currentUser.IsAdmin,
		})
	})

	app.Get("/experience/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		currentUser := getCurrentUser(c)
		settings := getSiteSettings()
		bikingTrends := getBikingTrends()
		sidebarAds := getActiveAds("sidebar")
		
		var exp struct {
			ID             int
			Title          string
			Category       string
			Description    string
			DurationDays   int
			DurationNights int
			Price          int
			DiscountedPrice int
			MaxPeople      int
			MinPeople      int
			Location       string
			StartLocation  string
			EndLocation    string
			VehicleType    string
			IsFeatured     bool
			CoverImage     string
		}
		var coverImage sql.NullString
		err := db.QueryRow(`
			SELECT id, title, category, description, duration_days, duration_nights, 
				   price, discounted_price, max_people, min_people, location, 
				   start_location, end_location, vehicle_type, is_featured, cover_image
			FROM experiences WHERE id = ? AND is_active = 1
		`, id).Scan(&exp.ID, &exp.Title, &exp.Category, &exp.Description, &exp.DurationDays,
			&exp.DurationNights, &exp.Price, &exp.DiscountedPrice, &exp.MaxPeople, &exp.MinPeople,
			&exp.Location, &exp.StartLocation, &exp.EndLocation, &exp.VehicleType, &exp.IsFeatured, &coverImage)
		if err != nil {
			return c.Redirect("/experiences")
		}
		if coverImage.Valid {
			exp.CoverImage = coverImage.String
		}
		
		itinerary := []map[string]string{
			{"day": "1", "title": "Arrival & Welcome", "description": "Pickup from airport, welcome dinner, and orientation"},
			{"day": "2", "title": "City Tour", "description": "Explore local attractions and cultural sites"},
			{"day": "3", "title": "Adventure Activities", "description": "Optional adventure sports and local experiences"},
		}
		
		included := []string{"Accommodation in caravan", "All meals (breakfast, lunch, dinner)", "Fuel and toll charges", "Professional driver/guide", "24/7 support"}
		excluded := []string{"Airfare/train tickets", "Personal expenses", "Entry fees to monuments", "Travel insurance"}
		
		return c.Render("experience-detail", fiber.Map{
			"Title":         exp.Title + " - " + settings.SiteName,
			"CurrentUser":   currentUser,
			"Experience":    exp,
			"Itinerary":     itinerary,
			"Included":      included,
			"Excluded":      excluded,
			"Settings":      settings,
			"BikingTrends":  bikingTrends,
			"SidebarAds":    sidebarAds,
			"IsAdmin":       currentUser != nil && currentUser.IsAdmin,
		})
	})

	app.Post("/experience/:id/book", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil {
			return c.Status(401).SendString("Please login to book")
		}
		
		experienceID := c.Params("id")
		travelDate := c.FormValue("travel_date")
		peopleStr := c.FormValue("people")
		requests := c.FormValue("requests")
		
		people, _ := strconv.Atoi(peopleStr)
		if people < 1 {
			people = 1
		}
		
		var discountedPrice int
		db.QueryRow("SELECT discounted_price FROM experiences WHERE id = ?", experienceID).Scan(&discountedPrice)
		totalPrice := discountedPrice * people
		
		_, err := db.Exec(`INSERT INTO experience_bookings (experience_id, user_id, travel_date, number_of_people, total_price, special_requests, contact_name, contact_phone, contact_email, status) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'pending')`,
			experienceID, currentUser.ID, travelDate, people, totalPrice, requests, currentUser.Username, currentUser.Phone, currentUser.Email)
		
		if err != nil {
			return c.Redirect("/experience/" + experienceID + "?error=Booking+failed.+Please+try+again.")
		}
		
		return c.Redirect("/profile?success=Booking+request+submitted%21+We%27ll+contact+you+shortly.")
	})

	// ==================== MARKETPLACE ====================
	
	app.Get("/marketplace", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		category := c.Query("category")
		categories := getMarketplaceCategories()
		sidebarAds := getActiveAds("sidebar")
		settings := getSiteSettings()
		
		var products []Product
		var rows *sql.Rows
		var err error
		
		if category != "" && category != "All Items" {
			rows, err = db.Query(`
				SELECT id, user_id, user_name, title, description, price, category, condition, image_url, location, status, created_at
				FROM products WHERE category = ? AND status = 'available' ORDER BY created_at DESC
			`, category)
		} else {
			rows, err = db.Query(`
				SELECT id, user_id, user_name, title, description, price, category, condition, image_url, location, status, created_at
				FROM products WHERE status = 'available' ORDER BY created_at DESC
			`)
		}
		
		if err == nil && rows != nil {
			defer rows.Close()
			for rows.Next() {
				var p Product
				if err := rows.Scan(&p.ID, &p.UserID, &p.UserName, &p.Title, &p.Description, &p.Price, &p.Category, &p.Condition, &p.ImageURL, &p.Location, &p.Status, &p.CreatedAt); err == nil {
					products = append(products, p)
				}
			}
		}
		
		return c.Render("marketplace", fiber.Map{
			"CurrentUser":     currentUser,
			"Categories":      categories,
			"Products":        products,
			"SelectedCategory": category,
			"SidebarAds":      sidebarAds,
			"Settings":        settings,
			"IsAdmin":         currentUser != nil && currentUser.IsAdmin,
		})
	})

	app.Post("/marketplace/add", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil {
			return c.Status(401).SendString(`<div class="p-4 text-red-500">Please login to sell items</div>`)
		}
		
		title := c.FormValue("title")
		description := c.FormValue("description")
		priceStr := c.FormValue("price")
		category := c.FormValue("category")
		condition := c.FormValue("condition")
		imageURL := c.FormValue("image_url")
		location := c.FormValue("location")
		
		var priceFloat float64
		fmt.Sscanf(priceStr, "%f", &priceFloat)
		
		_, err := db.Exec(`INSERT INTO products (user_id, user_name, title, description, price, category, condition, image_url, location) 
						   VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			currentUser.ID, currentUser.Username, title, description, priceFloat, category, condition, imageURL, location)
		
		if err != nil {
			log.Println("Error adding product:", err)
			return c.Status(500).SendString(`<div class="p-4 text-red-500">Failed to add product</div>`)
		}
		
		return c.SendString(`<div class="p-4 text-emerald-600">✅ Product listed successfully!<script>setTimeout(function(){ window.location.href = "/marketplace"; }, 1000);</script></div>`)
	})

	// ==================== NEWS PAGE ====================

	app.Get("/news", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		settings := getSiteSettings()
		sidebarAds := getActiveAds("sidebar")
		
		// Fetch all news from database
		rows, err := db.Query(`
			SELECT id, title, content, category, likes, 
				   strftime('%Y-%m-%d %H:%M:%S', created_at) as created_at
			FROM biking_news 
			ORDER BY created_at DESC
		`)
		
		var allNews []BikingNews
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var n BikingNews
				var createdAt string
				rows.Scan(&n.ID, &n.Title, &n.Content, &n.Category, &n.Likes, &createdAt)
				
				// Format timestamp
				t, _ := time.Parse("2006-01-02 15:04:05", createdAt)
				diff := time.Since(t)
				if diff < time.Hour {
					n.Timestamp = fmt.Sprintf("%d minutes ago", int(diff.Minutes()))
				} else if diff < 24*time.Hour {
					n.Timestamp = fmt.Sprintf("%d hours ago", int(diff.Hours()))
				} else {
					n.Timestamp = fmt.Sprintf("%d days ago", int(diff.Hours()/24))
				}
				
				allNews = append(allNews, n)
			}
		}

		return c.Render("news", fiber.Map{
			"CurrentUser": currentUser,
			"News":        allNews,
			"IsAdmin":     currentUser != nil && currentUser.IsAdmin,
			"SidebarAds":  sidebarAds,
			"Settings":    settings,
		})
	})

	// Like news endpoint
	app.Post("/news/like/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		_, err := db.Exec("UPDATE biking_news SET likes = likes + 1 WHERE id = ?", id)
		if err != nil {
			return c.Status(500).SendString("0")
		}
		var likes int
		db.QueryRow("SELECT likes FROM biking_news WHERE id = ?", id).Scan(&likes)
		return c.SendString(fmt.Sprintf("%d", likes))
	})

	// ==================== ADMIN ROUTES ====================
	
	app.Post("/admin/news/add", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil || !currentUser.IsAdmin {
			return c.Status(403).SendString("Access denied")
		}

		title := c.FormValue("title")
		content := c.FormValue("content")
		category := c.FormValue("category")

		_, err := db.Exec("INSERT INTO biking_news (title, content, category) VALUES (?, ?, ?)", title, content, category)
		if err != nil {
			return c.Status(500).SendString("Error adding news")
		}

		return c.SendString(`✅ News added!<script>window.location.reload()</script>`)
	})

	app.Post("/admin/news/delete/:id", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil || !currentUser.IsAdmin {
			return c.Status(403).SendString("Access denied")
		}
		id := c.Params("id")
		db.Exec("DELETE FROM biking_news WHERE id = ?", id)
		return c.SendString("✅ News deleted")
	})

	app.Post("/admin/news/edit/:id", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil || !currentUser.IsAdmin {
			return c.Status(403).SendString("Access denied")
		}
		id := c.Params("id")
		title := c.FormValue("title")
		content := c.FormValue("content")
		category := c.FormValue("category")
		db.Exec("UPDATE biking_news SET title = ?, content = ?, category = ? WHERE id = ?", title, content, category, id)
		return c.SendString(`✅ News updated!<script>window.location.reload()</script>`)
	})

	// Admin verification approval
	app.Post("/admin/verify-user/:id", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil || !currentUser.IsAdmin {
			return c.Status(403).SendString("Access denied")
		}
		
		userID := c.Params("id")
		action := c.FormValue("action")
		
		var status string
		if action == "approve" {
			status = "verified"
			db.Exec("UPDATE users SET is_verified = 1 WHERE id = ?", userID)
		} else {
			status = "rejected"
		}
		
		now := time.Now()
		db.Exec(`UPDATE rider_verifications SET status = ?, verified_at = ?, verified_by = ? WHERE user_id = ?`,
			status, now, currentUser.ID, userID)
		
		return c.Redirect("/admin")
	})

	// Admin trends management
	app.Post("/admin/trends/add", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil || !currentUser.IsAdmin {
			return c.Status(403).SendString("Access denied")
		}
		
		title := c.FormValue("title")
		description := c.FormValue("description")
		trend := c.FormValue("trend")
		percentage := c.FormValue("percentage")
		category := c.FormValue("category")
		
		db.Exec("INSERT INTO biking_trends (title, description, trend, percentage, category) VALUES (?, ?, ?, ?, ?)",
			title, description, trend, percentage, category)
		return c.Redirect("/admin")
	})

	app.Post("/admin/trends/delete/:id", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil || !currentUser.IsAdmin {
			return c.Status(403).SendString("Access denied")
		}
		id := c.Params("id")
		db.Exec("DELETE FROM biking_trends WHERE id = ?", id)
		return c.Redirect("/admin")
	})
	
	app.Post("/admin/ads/add", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil || !currentUser.IsAdmin {
			return c.Status(403).SendString("Access denied")
		}

		title := c.FormValue("title")
		imageURL := c.FormValue("image_url")
		targetURL := c.FormValue("target_url")
		position := c.FormValue("position")
		advertiser := c.FormValue("advertiser")

		db.Exec(`INSERT INTO advertisements (title, image_url, target_url, position, advertiser, start_date, end_date) 
				 VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now', '+30 days'))`,
			title, imageURL, targetURL, position, advertiser)
		return c.SendString(`✅ Ad added!<script>window.location.reload()</script>`)
	})

	app.Post("/admin/ads/delete/:id", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil || !currentUser.IsAdmin {
			return c.Status(403).SendString("Access denied")
		}
		id := c.Params("id")
		db.Exec("DELETE FROM advertisements WHERE id = ?", id)
		return c.SendString("✅ Ad deleted")
	})

	app.Post("/admin/ads/toggle/:id", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil || !currentUser.IsAdmin {
			return c.Status(403).SendString("Access denied")
		}
		id := c.Params("id")
		db.Exec("UPDATE advertisements SET is_active = NOT is_active WHERE id = ?", id)
		return c.SendString("✅ Ad status updated")
	})

	app.Get("/ad/click/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		db.Exec("UPDATE advertisements SET clicks = clicks + 1 WHERE id = ?", id)
		var targetURL string
		db.QueryRow("SELECT target_url FROM advertisements WHERE id = ?", id).Scan(&targetURL)
		return c.Redirect(targetURL)
	})

	app.Post("/admin/settings", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil || !currentUser.IsAdmin {
			return c.Status(403).SendString("Access denied")
		}
		siteName := c.FormValue("site_name")
		logoURL := c.FormValue("logo_url")
		primaryColor := c.FormValue("primary_color")
		secondaryColor := c.FormValue("secondary_color")
		db.Exec(`UPDATE site_settings SET site_name = ?, logo_url = ?, primary_color = ?, secondary_color = ? WHERE id = 1`,
			siteName, logoURL, primaryColor, secondaryColor)
		return c.SendString(`✅ Settings updated!<script>window.location.reload()</script>`)
	})

	app.Post("/admin/product/delete/:id", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil || !currentUser.IsAdmin {
			return c.Status(403).SendString("Access denied")
		}
		id := c.Params("id")
		db.Exec("DELETE FROM products WHERE id = ?", id)
		return c.SendString("✅ Product deleted")
	})

	// ==================== SEARCH ====================
	
	app.Get("/search", func(c *fiber.Ctx) error {
		query := c.Query("q")
		if query == "" {
			return c.Redirect("/")
		}

		rides, _ := fetchRides(`
			SELECT id, username, handle, content, price, category, tags, image_url, likes, created_at,
				   CASE WHEN boost_until > datetime('now') THEN 1 ELSE 0 END,
				   CASE WHEN featured_until > datetime('now') THEN 1 ELSE 0 END,
				   user_id, status, COALESCE(from_location, ''), COALESCE(to_location, ''), COALESCE(departure_date, ''), COALESCE(bike_model, ''), COALESCE(seats, 0)
			FROM rides 
			WHERE (content LIKE ? OR tags LIKE ? OR category LIKE ? OR from_location LIKE ? OR to_location LIKE ?) AND status = 'approved'
			ORDER BY id DESC
		`, "%"+query+"%", "%"+query+"%", "%"+query+"%", "%"+query+"%", "%"+query+"%")

		sidebarAds := getActiveAds("sidebar")
		settings := getSiteSettings()
		bikingTrends := getBikingTrends()

		return c.Render("index", fiber.Map{
			"Rides":        rides,
			"CurrentUser":  getCurrentUser(c),
			"SearchQuery":  query,
			"IsProfile":    false,
			"IsAdmin":      false,
			"BikingTrends": bikingTrends,
			"SidebarAds":   sidebarAds,
			"Settings":     settings,
		})
	})

	// ==================== PROFILE ====================
	
	app.Get("/profile", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil {
			return c.Redirect("/")
		}

		successMessage := c.Query("success")
		errorMessage := c.Query("error")
		bookingSuccess := c.Query("booking") == "success"

		rides, _ := fetchRides(`
			SELECT id, username, handle, content, price, category, tags, image_url, likes, created_at,
				   CASE WHEN boost_until > datetime('now') THEN 1 ELSE 0 END,
				   CASE WHEN featured_until > datetime('now') THEN 1 ELSE 0 END,
				   user_id, status, COALESCE(from_location, ''), COALESCE(to_location, ''), COALESCE(departure_date, ''), COALESCE(bike_model, ''), COALESCE(seats, 0)
			FROM rides 
			WHERE user_id = ? AND status = 'approved'
			ORDER BY id DESC
		`, currentUser.ID)

		// Fetch user's experience bookings
		type UserBooking struct {
			ID              int
			ExperienceID    int
			Title           string
			CoverImage      string
			TravelDate      string
			NumberOfPeople  int
			TotalPrice      int
			SpecialRequests string
			Status          string
			CreatedAt       string
		}
		
		var bookings []UserBooking
		bookingRows, err := db.Query(`
			SELECT eb.id, eb.experience_id, e.title, COALESCE(e.cover_image, ''),
				   eb.travel_date, eb.number_of_people, eb.total_price, 
				   COALESCE(eb.special_requests, ''), eb.status, eb.created_at
			FROM experience_bookings eb
			JOIN experiences e ON eb.experience_id = e.id
			WHERE eb.user_id = ?
			ORDER BY eb.created_at DESC
		`, currentUser.ID)
		
		if err == nil {
			defer bookingRows.Close()
			for bookingRows.Next() {
				var b UserBooking
				var travelDate time.Time
				var createdAt time.Time
				bookingRows.Scan(&b.ID, &b.ExperienceID, &b.Title, &b.CoverImage,
					&travelDate, &b.NumberOfPeople, &b.TotalPrice, &b.SpecialRequests,
					&b.Status, &createdAt)
				b.TravelDate = travelDate.Format("2006-01-02")
				b.CreatedAt = createdAt.Format("Jan 02, 2006")
				bookings = append(bookings, b)
			}
		}

		var totalRides, totalLikes, totalCreditsEarned int
		db.QueryRow("SELECT COUNT(*) FROM rides WHERE user_id = ? AND status = 'approved'", currentUser.ID).Scan(&totalRides)
		db.QueryRow("SELECT COALESCE(SUM(likes), 0) FROM rides WHERE user_id = ?", currentUser.ID).Scan(&totalLikes)
		
		var joinedAt time.Time
		db.QueryRow("SELECT created_at FROM users WHERE id = ?", currentUser.ID).Scan(&joinedAt)
		
		var bio, location string
		db.QueryRow("SELECT COALESCE(bio, ''), COALESCE(location, '') FROM user_profiles WHERE user_id = ?", currentUser.ID).Scan(&bio, &location)
		
		sidebarAds := getActiveAds("sidebar")
		settings := getSiteSettings()
		bikingTrends := getBikingTrends()

		return c.Render("profile", fiber.Map{
			"Rides":              rides,
			"CurrentUser":        currentUser,
			"IsProfile":          true,
			"SearchQuery":        "",
			"IsAdmin":            currentUser.IsAdmin,
			"BikingTrends":       bikingTrends,
			"SidebarAds":         sidebarAds,
			"Settings":           settings,
			"TotalRides":         totalRides,
			"TotalLikes":         totalLikes,
			"TotalCreditsEarned": totalCreditsEarned,
			"JoinedAt":           joinedAt,
			"Bio":                bio,
			"Location":           location,
			"SuccessMessage":     successMessage,
			"ErrorMessage":       errorMessage,
			"BookingSuccess":     bookingSuccess,
			"Bookings":           bookings,
		})
	})

	app.Get("/profile/edit", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil {
			return c.Redirect("/login")
		}
		
		var bio, location string
		db.QueryRow("SELECT COALESCE(bio, ''), COALESCE(location, '') FROM user_profiles WHERE user_id = ?", currentUser.ID).Scan(&bio, &location)
		
		settings := getSiteSettings()
		bikingTrends := getBikingTrends()
		sidebarAds := getActiveAds("sidebar")
		
		return c.Render("profile-edit", fiber.Map{
			"Title":         "Edit Profile - " + settings.SiteName,
			"CurrentUser":   currentUser,
			"Bio":           bio,
			"Location":      location,
			"Settings":      settings,
			"BikingTrends":  bikingTrends,
			"SidebarAds":    sidebarAds,
			"IsAdmin":       currentUser.IsAdmin,
		})
	})

	app.Post("/profile/update", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil {
			return c.Status(401).SendString(`<div class="p-4 text-red-500">Please login</div>`)
		}
		
		username := c.FormValue("username")
		bikeModel := c.FormValue("bike_model")
		ridingExp := c.FormValue("riding_exp")
		avatarURL := c.FormValue("avatar_url")
		bio := c.FormValue("bio")
		location := c.FormValue("location")
		
		if username != "" {
			db.Exec("UPDATE users SET username = ? WHERE id = ?", username, currentUser.ID)
		}
		db.Exec("UPDATE users SET bike_model = ?, riding_exp = ?, avatar_url = ? WHERE id = ?",
			bikeModel, ridingExp, avatarURL, currentUser.ID)
		
		db.Exec(`INSERT INTO user_profiles (user_id, bio, location) VALUES (?, ?, ?) 
				 ON CONFLICT(user_id) DO UPDATE SET bio = excluded.bio, location = excluded.location`,
			currentUser.ID, bio, location)
		
		return c.SendString(`<div class="p-4 text-emerald-600">✅ Profile updated successfully!<script>setTimeout(function(){ window.location.href = "/profile"; }, 1500);</script></div>`)
	})

	// Change Password
	app.Post("/change-password", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil {
			return c.Status(401).SendString("Please login")
		}
		
		currentPassword := c.FormValue("current_password")
		newPassword := c.FormValue("new_password")
		
		var hash string
		err := db.QueryRow("SELECT password_hash FROM users WHERE id = ?", currentUser.ID).Scan(&hash)
		if err != nil {
			return c.Status(500).SendString("Error verifying password")
		}
		
		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(currentPassword)) != nil {
			return c.SendString(`<div class="p-2 bg-red-100 text-red-700 rounded text-sm">❌ Current password is incorrect</div>`)
		}
		
		if err := validatePasswordStrength(newPassword); err != nil {
			return c.SendString(`<div class="p-2 bg-red-100 text-red-700 rounded text-sm">❌ ` + err.Error() + `</div>`)
		}
		
		newHash, _ := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		
		_, err = db.Exec("UPDATE users SET password_hash = ? WHERE id = ?", newHash, currentUser.ID)
		if err != nil {
			return c.Status(500).SendString(`<div class="p-2 bg-red-100 text-red-700 rounded text-sm">❌ Failed to update password</div>`)
		}
		
		return c.SendString(`<div class="p-2 bg-green-100 text-green-700 rounded text-sm">✅ Password changed successfully! Please login again.</div><script>setTimeout(function(){ window.location.href = "/logout"; }, 2000);</script>`)
	})

	// ==================== RIDE POSTING ====================
	
	app.Post("/post", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil {
			return c.Status(401).SendString(`<div class="p-4 text-red-500">Please login to post</div>`)
		}

		content := c.FormValue("content")
		category := c.FormValue("category")
		imageURL := c.FormValue("image_url")
		price := c.FormValue("price")
		fromLocation := c.FormValue("from_location")
		toLocation := c.FormValue("to_location")
		departureDate := c.FormValue("departure_date")
		bikeModel := c.FormValue("bike_model")
		seats := c.FormValue("seats")
		
		if price == "" {
			price = "Contact Owner"
		}
		
		seatsInt := 1
		if seats != "" {
			fmt.Sscanf(seats, "%d", &seatsInt)
		}

		if content == "" {
			return c.Status(400).SendString(`<div class="p-4 text-red-500">Content cannot be empty</div>`)
		}

		tagsString := detectTags(content, category)
		
		status := "pending"
		if currentUser.IsAdmin || currentUser.IsPremium {
			status = "approved"
		}

		_, err := db.Exec(`
			INSERT INTO rides (user_id, username, handle, content, price, category, tags, image_url, status, from_location, to_location, departure_date, bike_model, seats)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, currentUser.ID, currentUser.Username, currentUser.Handle, content, price, category, tagsString, imageURL, status, fromLocation, toLocation, departureDate, bikeModel, seatsInt)

		if err != nil {
			log.Println("Error creating post:", err)
			return c.Status(500).SendString(`<div class="p-4 text-red-500">Failed to create post</div>`)
		}

		msg := "✅ Ride posted successfully!"
		if status == "pending" {
			msg = "📝 Ride submitted for admin approval."
		}
		
		return c.SendString(`<div class="p-4 text-emerald-600">` + msg + `</div><script>window.location.reload()</script>`)
	})

	// ==================== PASSWORD RESET ====================
	
	app.Get("/forgot-password", func(c *fiber.Ctx) error {
		settings := getSiteSettings()
		return c.Render("forgot-password", fiber.Map{
			"CurrentUser": getCurrentUser(c),
			"Settings":    settings,
		})
	})

	app.Post("/forgot-password", func(c *fiber.Ctx) error {
		email := c.FormValue("email")
		
		var userID int
		err := db.QueryRow("SELECT id FROM users WHERE email = ? AND is_active = 1", email).Scan(&userID)
		if err != nil {
			return c.SendString(`<div class="p-4 text-emerald-600 text-center">If your email exists, you will receive a reset link.</div>`)
		}
		
		token := fmt.Sprintf("%d_%s", time.Now().UnixNano(), email)
		hash := sha256.Sum256([]byte(token))
		hashToken := hex.EncodeToString(hash[:])
		
		expiresAt := time.Now().Add(1 * time.Hour)
		db.Exec(`INSERT INTO password_resets (email, token, expires_at) VALUES (?, ?, ?)`, email, hashToken, expiresAt)
		
		resetLink := fmt.Sprintf("/reset-password?token=%s", token)
		
		return c.SendString(fmt.Sprintf(`
			<div class="p-4 text-emerald-600 text-center">
				✅ Reset link generated!<br>
				<a href="%s" style="color: %s">Click here to reset your password</a>
			</div>`, resetLink, getSiteSettings().PrimaryColor))
	})

	app.Get("/reset-password", func(c *fiber.Ctx) error {
		token := c.Query("token")
		if token == "" {
			return c.Redirect("/")
		}
		
		hash := sha256.Sum256([]byte(token))
		hashToken := hex.EncodeToString(hash[:])
		
		var email string
		var expiresAt time.Time
		err := db.QueryRow(`SELECT email, expires_at FROM password_resets WHERE token = ? AND used = FALSE AND expires_at > datetime('now')`,
			hashToken).Scan(&email, &expiresAt)
		
		if err != nil {
			return c.SendString(`<div class="p-4 text-red-500 text-center">Invalid or expired reset link.</div>`)
		}
		
		settings := getSiteSettings()
		return c.Render("reset-password", fiber.Map{
			"Token":    token,
			"Email":    email,
			"Settings": settings,
		})
	})

	app.Post("/reset-password", func(c *fiber.Ctx) error {
		token := c.FormValue("token")
		password := c.FormValue("password")
		
		if len(password) < 8 {
			return c.SendString(`<div class="p-4 text-red-500 text-center">Password must be at least 8 characters</div>`)
		}
		
		hash := sha256.Sum256([]byte(token))
		hashToken := hex.EncodeToString(hash[:])
		
		var email string
		err := db.QueryRow(`SELECT email FROM password_resets WHERE token = ? AND used = FALSE AND expires_at > datetime('now')`,
			hashToken).Scan(&email)
		
		if err != nil {
			return c.SendString(`<div class="p-4 text-red-500 text-center">Invalid or expired reset link</div>`)
		}
		
		hashPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		db.Exec("UPDATE users SET password_hash = ? WHERE email = ?", hashPassword, email)
		db.Exec("UPDATE password_resets SET used = TRUE WHERE token = ?", hashToken)
		
		return c.SendString(`<div class="p-4 text-emerald-600 text-center">Password reset successfully! <a href="/">Login now</a></div>`)
	})

	// ==================== PREMIUM & BOOST ====================
	
	app.Post("/upgrade", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil {
			return c.Status(401).JSON(fiber.Map{"success": false, "message": "Please login first"})
		}

		if currentUser.IsPremium {
			return c.JSON(fiber.Map{"success": false, "message": "You are already a premium member!"})
		}

		cost := 499
		if currentUser.Credits < cost {
			return c.JSON(fiber.Map{"success": false, "message": fmt.Sprintf("Not enough credits! Need %d credits.", cost)})
		}

		tx, _ := db.Begin()
		tx.Exec("UPDATE users SET credits = credits - ? WHERE id = ?", cost, currentUser.ID)
		premiumUntil := time.Now().Add(30 * 24 * time.Hour)
		tx.Exec("UPDATE users SET is_premium = true, premium_until = ?, membership_tier = 'premium' WHERE id = ?", premiumUntil, currentUser.ID)
		tx.Exec("UPDATE users SET credits = credits + 100 WHERE id = ?", currentUser.ID)
		tx.Exec("INSERT INTO transactions (user_id, amount, type, description) VALUES (?, ?, ?, ?)", currentUser.ID, cost, "debit", "Premium membership upgrade (30 days)")
		tx.Commit()

		return c.JSON(fiber.Map{"success": true, "message": "✅ Successfully upgraded to Premium! You received 100 bonus credits."})
	})

	app.Post("/boost/:id", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil {
			return c.Status(401).SendString("Please login")
		}

		cost := 99
		if currentUser.Credits < cost {
			return c.SendString(fmt.Sprintf("Not enough credits! Need %d credits.", cost))
		}

		id := c.Params("id")
		tx, _ := db.Begin()
		tx.Exec("UPDATE users SET credits = credits - ? WHERE id = ?", cost, currentUser.ID)
		tx.Exec(`UPDATE rides SET boost_until = datetime('now', '+7 days') WHERE id = ?`, id)
		tx.Commit()

		return c.SendString("✅ Ride boosted successfully for 7 days!")
	})

	app.Post("/feature/:id", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil {
			return c.Status(401).SendString("Please login")
		}

		cost := 199
		if currentUser.Credits < cost {
			return c.SendString(fmt.Sprintf("Not enough credits! Need %d credits.", cost))
		}

		id := c.Params("id")
		tx, _ := db.Begin()
		tx.Exec("UPDATE users SET credits = credits - ? WHERE id = ?", cost, currentUser.ID)
		tx.Exec(`UPDATE rides SET featured_until = datetime('now', '+1 day') WHERE id = ?`, id)
		tx.Commit()

		return c.SendString("✅ Ride featured for 24 hours!")
	})

	app.Post("/buy-credits", func(c *fiber.Ctx) error {
		return c.SendString(`<div class="p-4 text-amber-600 text-center">🚧 Payment integration coming soon!</div>`)
	})

	// ==================== LIKES & JOIN ====================
	
	app.Post("/like/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		db.Exec("UPDATE rides SET likes = likes + 1 WHERE id = ?", id)
		var likes int
		db.QueryRow("SELECT likes FROM rides WHERE id = ?", id).Scan(&likes)
		return c.SendString(fmt.Sprintf(`<button hx-post="/like/%s" hx-swap="outerHTML" class="flex items-center gap-1.5 text-slate-400 hover:text-red-500 transition-all active:scale-110">❤️ <span class="text-sm font-medium">%d</span></button>`, id, likes))
	})

	app.Post("/join/:id", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil {
			return c.Status(401).JSON(fiber.Map{"success": false, "message": "Please login to join rides"})
		}

		rideID := c.Params("id")
		message := c.FormValue("message")

		if message == "" {
			return c.Status(400).JSON(fiber.Map{"success": false, "message": "Please tell the rider about yourself"})
		}

		db.Exec(`INSERT INTO ride_requests (ride_id, rider_id, rider_name, rider_email, rider_phone, message) VALUES (?, ?, ?, ?, ?, ?)`,
			rideID, currentUser.ID, currentUser.Username, currentUser.Email, currentUser.Phone, message)

		return c.JSON(fiber.Map{"success": true, "message": "✅ Request submitted! The ride organizer will contact you."})
	})

	// ==================== ADMIN DASHBOARD ====================
	
	app.Get("/admin", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil || !currentUser.IsAdmin {
			return c.Status(403).SendString("Access denied. Admin only.")
		}

		pendingRides, _ := fetchRides(`
			SELECT id, username, handle, content, price, category, tags, image_url, likes, created_at,
				   CASE WHEN boost_until > datetime('now') THEN 1 ELSE 0 END,
				   CASE WHEN featured_until > datetime('now') THEN 1 ELSE 0 END,
				   user_id, status, COALESCE(from_location, ''), COALESCE(to_location, ''), COALESCE(departure_date, ''), COALESCE(bike_model, ''), COALESCE(seats, 0)
			FROM rides WHERE status = 'pending' ORDER BY created_at DESC
		`)

		allNews := getAllNews()
		bikingTrends := getBikingTrends()
		allAds := getActiveAds("")
		
		var allProducts []Product
		rows, _ := db.Query(`SELECT id, user_id, user_name, title, description, price, category, condition, image_url, location, status, created_at FROM products ORDER BY created_at DESC`)
		if rows != nil {
			defer rows.Close()
			for rows.Next() {
				var p Product
				rows.Scan(&p.ID, &p.UserID, &p.UserName, &p.Title, &p.Description, &p.Price, &p.Category, &p.Condition, &p.ImageURL, &p.Location, &p.Status, &p.CreatedAt)
				allProducts = append(allProducts, p)
			}
		}

		// Fetch all experiences for admin
		expRows, err := db.Query(`
			SELECT id, title, category, description, duration_days, duration_nights,
				   price, discounted_price, max_people, location, is_featured, is_active
			FROM experiences ORDER BY created_at DESC
		`)
		var allExperiences []map[string]interface{}
		if err == nil {
			defer expRows.Close()
			for expRows.Next() {
				var id, price, discountedPrice, maxPeople, durationDays, durationNights int
				var title, category, description, location string
				var isFeatured, isActive bool
				
				expRows.Scan(&id, &title, &category, &description, &durationDays, &durationNights,
					&price, &discountedPrice, &maxPeople, &location, &isFeatured, &isActive)
				
				allExperiences = append(allExperiences, map[string]interface{}{
					"ID":              id,
					"Title":           title,
					"Category":        category,
					"Description":     description,
					"DurationDays":    durationDays,
					"DurationNights":  durationNights,
					"Price":           price,
					"DiscountedPrice": discountedPrice,
					"MaxPeople":       maxPeople,
					"Location":        location,
					"IsFeatured":      isFeatured,
					"IsActive":        isActive,
				})
			}
		}

		// Fetch all POIs for admin
		poiRows, err := db.Query(`
			SELECT id, name, type, city, state, is_partner, discount_percentage, is_active
			FROM points_of_interest ORDER BY created_at DESC
		`)
		var allPOIs []map[string]interface{}
		if err == nil {
			defer poiRows.Close()
			for poiRows.Next() {
				var id, discount int
				var name, poiType, city, state string
				var isPartner, isActive bool
				
				poiRows.Scan(&id, &name, &poiType, &city, &state, &isPartner, &discount, &isActive)
				
				allPOIs = append(allPOIs, map[string]interface{}{
					"ID": id, "Name": name, "Type": poiType,
					"City": city, "State": state, "IsPartner": isPartner,
					"DiscountPercent": discount, "IsActive": isActive,
				})
			}
		}

		userRows, _ := db.Query(`SELECT id, username, handle, email, COALESCE(phone, ''), is_admin, credits, is_premium, is_active, bike_model, riding_exp, avatar_url, is_verified FROM users ORDER BY id DESC`)
		var users []User
		if userRows != nil {
			defer userRows.Close()
			for userRows.Next() {
				var u User
				userRows.Scan(&u.ID, &u.Username, &u.Handle, &u.Email, &u.Phone, &u.IsAdmin, &u.Credits, &u.IsPremium, &u.IsActive, &u.BikeModel, &u.RidingExp, &u.AvatarURL, &u.IsVerified)
				users = append(users, u)
			}
		}

		// Get pending verifications
		verifRows, _ := db.Query(`SELECT rv.id, rv.user_id, u.username, u.email, rv.dl_number, rv.bike_rc_number, rv.status 
								  FROM rider_verifications rv JOIN users u ON rv.user_id = u.id WHERE rv.status = 'pending'`)
		var pendingVerifications []struct {
			ID           int
			UserID       int
			Username     string
			Email        string
			DLNumber     string
			BikeRCNumber string
			Status       string
		}
		if verifRows != nil {
			defer verifRows.Close()
			for verifRows.Next() {
				var v struct {
					ID           int
					UserID       int
					Username     string
					Email        string
					DLNumber     string
					BikeRCNumber string
					Status       string
				}
				verifRows.Scan(&v.ID, &v.UserID, &v.Username, &v.Email, &v.DLNumber, &v.BikeRCNumber, &v.Status)
				pendingVerifications = append(pendingVerifications, v)
			}
		}

		var totalRides, totalUsers, totalCreditsSpent int
		db.QueryRow("SELECT COUNT(*) FROM rides WHERE status = 'approved'").Scan(&totalRides)
		db.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalUsers)
		db.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM transactions WHERE type = 'debit'").Scan(&totalCreditsSpent)

		categories := getMarketplaceCategories()
		settings := getSiteSettings()

		return c.Render("admin", fiber.Map{
			"CurrentUser":          currentUser,
			"PendingRides":         pendingRides,
			"Users":                users,
			"TotalRides":           totalRides,
			"TotalUsers":           totalUsers,
			"TotalCreditsSpent":    totalCreditsSpent,
			"AllNews":              allNews,
			"BikingTrends":         bikingTrends,
			"AllAds":               allAds,
			"AllProducts":          allProducts,
			"AllExperiences":       allExperiences,
			"AllPOIs":              allPOIs,
			"Categories":           categories,
			"Settings":             settings,
			"PendingVerifications": pendingVerifications,
		})
	})

	app.Post("/admin/approve/:id", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil || !currentUser.IsAdmin {
			return c.Status(403).SendString("Access denied")
		}
		id := c.Params("id")
		db.Exec("UPDATE rides SET status = 'approved' WHERE id = ?", id)
		return c.SendString("✅ Ride approved")
	})

	app.Post("/admin/delete/:id", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil || !currentUser.IsAdmin {
			return c.Status(403).SendString("Access denied")
		}
		id := c.Params("id")
		db.Exec("DELETE FROM rides WHERE id = ?", id)
		return c.SendString("✅ Ride deleted")
	})

	app.Post("/admin/give-credits", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil || !currentUser.IsAdmin {
			return c.Status(403).SendString("Access denied")
		}
		userID := c.FormValue("user_id")
		credits := c.FormValue("credits")
		var creditsInt int
		fmt.Sscanf(credits, "%d", &creditsInt)
		db.Exec("UPDATE users SET credits = credits + ? WHERE id = ?", creditsInt, userID)
		db.Exec("INSERT INTO transactions (user_id, amount, type, description) VALUES (?, ?, ?, ?)", userID, creditsInt, "credit", "Admin grant")
		return c.SendString(fmt.Sprintf("✅ Added %d credits", creditsInt))
	})

	// ==================== ADMIN EXPERIENCE MANAGEMENT ====================

	// Get all experiences for admin
	app.Get("/admin/experiences", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil || !currentUser.IsAdmin {
			return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
		}
		
		rows, err := db.Query(`
			SELECT id, title, category, description, duration_days, duration_nights,
				   price, discounted_price, max_people, min_people, location,
				   start_location, end_location, vehicle_type, is_featured, is_active,
				   cover_image, rating, total_reviews, created_at
			FROM experiences ORDER BY created_at DESC
		`)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to load experiences"})
		}
		defer rows.Close()
		
		var experiences []map[string]interface{}
		for rows.Next() {
			var id, price, discountedPrice, maxPeople, minPeople, durationDays, durationNights int
			var title, category, description, location, startLocation, endLocation, vehicleType, coverImage string
			var isFeatured, isActive bool
			var rating float64
			var totalReviews int
			var createdAt time.Time
			
			rows.Scan(&id, &title, &category, &description, &durationDays, &durationNights,
				&price, &discountedPrice, &maxPeople, &minPeople, &location,
				&startLocation, &endLocation, &vehicleType, &isFeatured, &isActive,
				&coverImage, &rating, &totalReviews, &createdAt)
			
			exp := map[string]interface{}{
				"ID": id, "Title": title, "Category": category, "Description": description,
				"DurationDays": durationDays, "DurationNights": durationNights,
				"Price": price, "DiscountedPrice": discountedPrice,
				"MaxPeople": maxPeople, "MinPeople": minPeople, "Location": location,
				"StartLocation": startLocation, "EndLocation": endLocation,
				"VehicleType": vehicleType, "IsFeatured": isFeatured, "IsActive": isActive,
				"CoverImage": coverImage, "Rating": rating, "TotalReviews": totalReviews,
				"CreatedAt": createdAt,
			}
			experiences = append(experiences, exp)
		}
		
		return c.JSON(fiber.Map{"success": true, "experiences": experiences})
	})

	// Get single experience for editing
	app.Get("/admin/experience/:id", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil || !currentUser.IsAdmin {
			return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
		}
		
		id := c.Params("id")
		var exp struct {
			ID              int
			Title           string
			Category        string
			Description     string
			DurationDays    int
			DurationNights  int
			Price           int
			DiscountedPrice int
			MaxPeople       int
			MinPeople       int
			Location        string
			StartLocation   string
			EndLocation     string
			VehicleType     string
			IsFeatured      bool
			IsActive        bool
			CoverImage      string
		}
		
		var coverImage sql.NullString
		err := db.QueryRow(`
			SELECT id, title, category, description, duration_days, duration_nights,
				   price, discounted_price, max_people, min_people, location,
				   COALESCE(start_location, ''), COALESCE(end_location, ''), vehicle_type,
				   is_featured, is_active, COALESCE(cover_image, '')
			FROM experiences WHERE id = ?
		`, id).Scan(&exp.ID, &exp.Title, &exp.Category, &exp.Description, &exp.DurationDays,
			&exp.DurationNights, &exp.Price, &exp.DiscountedPrice, &exp.MaxPeople, &exp.MinPeople,
			&exp.Location, &exp.StartLocation, &exp.EndLocation, &exp.VehicleType,
			&exp.IsFeatured, &exp.IsActive, &coverImage)
		
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"success": false, "error": "Experience not found"})
		}
		
		if coverImage.Valid {
			exp.CoverImage = coverImage.String
		}
		
		return c.JSON(fiber.Map{"success": true, "experience": exp})
	})

	// Add new experience
	app.Post("/admin/experience/add", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil || !currentUser.IsAdmin {
			return c.Status(403).SendString("Access denied")
		}
		
		title := c.FormValue("title")
		category := c.FormValue("category")
		description := c.FormValue("description")
		durationDays, _ := strconv.Atoi(c.FormValue("duration_days"))
		durationNights, _ := strconv.Atoi(c.FormValue("duration_nights"))
		price, _ := strconv.Atoi(c.FormValue("price"))
		discountedPrice, _ := strconv.Atoi(c.FormValue("discounted_price"))
		maxPeople, _ := strconv.Atoi(c.FormValue("max_people"))
		minPeople, _ := strconv.Atoi(c.FormValue("min_people"))
		location := c.FormValue("location")
		startLocation := c.FormValue("start_location")
		endLocation := c.FormValue("end_location")
		vehicleType := c.FormValue("vehicle_type")
		coverImage := c.FormValue("cover_image")
		isFeatured := c.FormValue("is_featured") == "true" || c.FormValue("is_featured") == "on"
		isActive := c.FormValue("is_active") == "true" || c.FormValue("is_active") == "on"
		
		if title == "" || description == "" || price == 0 {
			return c.Status(400).SendString("❌ Title, description and price are required")
		}
		
		if discountedPrice == 0 {
			discountedPrice = price
		}
		
		if durationDays == 0 {
			durationDays = 5
		}
		if durationNights == 0 {
			durationNights = 4
		}
		if maxPeople == 0 {
			maxPeople = 6
		}
		if minPeople == 0 {
			minPeople = 1
		}
		
		_, err := db.Exec(`
			INSERT INTO experiences (title, category, description, duration_days, duration_nights,
				price, discounted_price, max_people, min_people, location, start_location,
				end_location, vehicle_type, cover_image, is_featured, is_active)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, title, category, description, durationDays, durationNights, price, discountedPrice,
			maxPeople, minPeople, location, startLocation, endLocation, vehicleType,
			coverImage, isFeatured, isActive)
		
		if err != nil {
			log.Println("Error adding experience:", err)
			return c.Status(500).SendString("❌ Failed to add experience: " + err.Error())
		}
		
		return c.SendString("✅ Experience added successfully!")
	})

	// Update experience
	app.Post("/admin/experience/update/:id", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil || !currentUser.IsAdmin {
			return c.Status(403).SendString("Access denied")
		}
		
		id := c.Params("id")
		title := c.FormValue("title")
		category := c.FormValue("category")
		description := c.FormValue("description")
		durationDays, _ := strconv.Atoi(c.FormValue("duration_days"))
		durationNights, _ := strconv.Atoi(c.FormValue("duration_nights"))
		price, _ := strconv.Atoi(c.FormValue("price"))
		discountedPrice, _ := strconv.Atoi(c.FormValue("discounted_price"))
		maxPeople, _ := strconv.Atoi(c.FormValue("max_people"))
		minPeople, _ := strconv.Atoi(c.FormValue("min_people"))
		location := c.FormValue("location")
		startLocation := c.FormValue("start_location")
		endLocation := c.FormValue("end_location")
		vehicleType := c.FormValue("vehicle_type")
		coverImage := c.FormValue("cover_image")
		isFeatured := c.FormValue("is_featured") == "true" || c.FormValue("is_featured") == "on"
		isActive := c.FormValue("is_active") == "true" || c.FormValue("is_active") == "on"
		
		_, err := db.Exec(`
			UPDATE experiences SET
				title = ?, category = ?, description = ?, duration_days = ?, duration_nights = ?,
				price = ?, discounted_price = ?, max_people = ?, min_people = ?, location = ?,
				start_location = ?, end_location = ?, vehicle_type = ?, cover_image = ?,
				is_featured = ?, is_active = ?
			WHERE id = ?
		`, title, category, description, durationDays, durationNights, price, discountedPrice,
			maxPeople, minPeople, location, startLocation, endLocation, vehicleType,
			coverImage, isFeatured, isActive, id)
		
		if err != nil {
			return c.Status(500).SendString("❌ Failed to update experience")
		}
		
		return c.SendString("✅ Experience updated successfully!")
	})

	// Toggle experience active status
	app.Post("/admin/experience/toggle/:id", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil || !currentUser.IsAdmin {
			return c.Status(403).SendString("Access denied")
		}
		
		id := c.Params("id")
		db.Exec("UPDATE experiences SET is_active = NOT is_active WHERE id = ?", id)
		return c.SendString("✅ Experience status toggled")
	})

	// Toggle experience featured status
	app.Post("/admin/experience/feature/:id", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil || !currentUser.IsAdmin {
			return c.Status(403).SendString("Access denied")
		}
		
		id := c.Params("id")
		db.Exec("UPDATE experiences SET is_featured = NOT is_featured WHERE id = ?", id)
		return c.SendString("✅ Experience featured status toggled")
	})

	// Delete experience
	app.Post("/admin/experience/delete/:id", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil || !currentUser.IsAdmin {
			return c.Status(403).SendString("Access denied")
		}
		
		id := c.Params("id")
		
		// Check if there are bookings
		var bookingCount int
		db.QueryRow("SELECT COUNT(*) FROM experience_bookings WHERE experience_id = ?", id).Scan(&bookingCount)
		
		if bookingCount > 0 {
			// Soft delete - just mark inactive
			db.Exec("UPDATE experiences SET is_active = 0 WHERE id = ?", id)
			return c.SendString("✅ Experience deactivated (has existing bookings)")
		}
		
		db.Exec("DELETE FROM experiences WHERE id = ?", id)
		return c.SendString("✅ Experience deleted")
	})

	// ==================== TRIP PLANNER & POI ROUTES ====================

	// Trip Planner Page
	app.Get("/trip-planner", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		settings := getSiteSettings()
		sidebarAds := getActiveAds("sidebar")
		
		return c.Render("trip-planner", fiber.Map{
			"CurrentUser": currentUser,
			"Settings":    settings,
			"SidebarAds":  sidebarAds,
			"IsAdmin":     currentUser != nil && currentUser.IsAdmin,
		})
	})

	// API: Get nearby places based on location
	app.Get("/api/nearby", func(c *fiber.Ctx) error {
		latStr := c.Query("lat")
		lngStr := c.Query("lng")
		poiType := c.Query("type")
		
		lat, err := strconv.ParseFloat(latStr, 64)
		if err != nil || lat == 0 {
			lat = 28.6139 // Default Delhi
		}
		
		lng, err := strconv.ParseFloat(lngStr, 64)
		if err != nil || lng == 0 {
			lng = 77.2090 // Default Delhi
		}
		
		// Build query
		query := `
			SELECT id, name, type, category, latitude, longitude, 
				   COALESCE(address, '') as address, 
				   COALESCE(city, '') as city, 
				   COALESCE(phone, '') as phone, 
				   COALESCE(price_range, '₹₹') as price_range, 
				   COALESCE(rating, 0) as rating, 
				   COALESCE(total_reviews, 0) as total_reviews,
				   COALESCE(is_partner, 0) as is_partner, 
				   COALESCE(discount_percentage, 0) as discount_percentage,
				   COALESCE(offer_details, '') as offer_details
			FROM points_of_interest 
			WHERE is_active = 1
		`
		args := []interface{}{}
		
		if poiType != "" && poiType != "all" {
			query += " AND type = ?"
			args = append(args, poiType)
		}
		
		query += " ORDER BY id DESC LIMIT 50"
		
		rows, err := db.Query(query, args...)
		if err != nil {
			log.Println("Error fetching POIs:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch places"})
		}
		defer rows.Close()
		
		var places []map[string]interface{}
		for rows.Next() {
			var id int
			var name, poiTypeVal, category, address, city, phone, priceRange, offerDetails string
			var latitude, longitude, rating float64
			var totalReviews, discountPercentage int
			var isPartner bool
			
			err := rows.Scan(&id, &name, &poiTypeVal, &category, &latitude, &longitude, &address, &city,
				&phone, &priceRange, &rating, &totalReviews, &isPartner, &discountPercentage, &offerDetails)
			if err != nil {
				continue
			}
			
			distance := calculateDistance(lat, lng, latitude, longitude)
			
			places = append(places, map[string]interface{}{
				"id": id, "name": name, "type": poiTypeVal, "category": category,
				"latitude": latitude, "longitude": longitude, "address": address,
				"city": city, "phone": phone, "price_range": priceRange,
				"rating": rating, "total_reviews": totalReviews,
				"distance": math.Round(distance*10)/10,
				"discount_percent": discountPercentage,
				"is_partner": isPartner,
			})
		}
		
		// Get nearby offers
		var offers []map[string]interface{}
		offerRows, err := db.Query(`
			SELECT DISTINCT po.id, po.name, po.discount_percentage, po.offer_details
			FROM points_of_interest po
			WHERE po.is_partner = 1 AND po.discount_percentage > 0 AND po.is_active = 1
			LIMIT 10
		`)
		
		if err == nil {
			defer offerRows.Close()
			for offerRows.Next() {
				var id, discount int
				var name, details string
				offerRows.Scan(&id, &name, &discount, &details)
				offers = append(offers, map[string]interface{}{
					"id": id, "title": name, "description": details,
					"discount_value": discount,
				})
			}
		}
		
		return c.JSON(fiber.Map{
			"success": true,
			"places": places,
			"offers": offers,
			"user_location": map[string]float64{"lat": lat, "lng": lng},
			"count": len(places),
		})
	})

	// API: Save user location for personalized offers
	app.Post("/api/save-location", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil {
			return c.Status(401).JSON(fiber.Map{"error": "Not logged in"})
		}
		
		var data struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		}
		
		if err := c.BodyParser(&data); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid data"})
		}
		
		db.Exec(`INSERT INTO user_location_history (user_id, latitude, longitude) 
				 VALUES (?, ?, ?)`, currentUser.ID, data.Latitude, data.Longitude)
		
		return c.JSON(fiber.Map{"success": true})
	})

	// API: Claim partner offer
	app.Post("/api/claim-offer", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil {
			return c.Status(401).JSON(fiber.Map{"error": "Please login"})
		}
		
		var data struct {
			OfferID int `json:"offer_id"`
		}
		
		if err := c.BodyParser(&data); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid data"})
		}
		
		// Generate unique claim code
		claimCode := fmt.Sprintf("HC%d%d", currentUser.ID, time.Now().UnixNano())
		
		_, err := db.Exec(`INSERT INTO partner_offers_claimed (offer_id, user_id, status) 
						   VALUES (?, ?, 'claimed')`, data.OfferID, currentUser.ID)
		
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to claim offer"})
		}
		
		return c.JSON(fiber.Map{
			"success": true,
			"code": claimCode[:10],
			"message": "Offer claimed successfully! Show this code at the partner location.",
		})
	})

	// Place detail page
// Place detail page - SIMPLIFIED VERSION
app.Get("/place/:id", func(c *fiber.Ctx) error {
    id := c.Params("id")
    currentUser := getCurrentUser(c)
    settings := getSiteSettings()
    
    var place PointOfInterest
    
    // Simplified query - only select columns that definitely exist
    err := db.QueryRow(`
        SELECT id, name, type, 
               COALESCE(category, '') as category,
               latitude, longitude, 
               COALESCE(address, '') as address,
               COALESCE(city, '') as city,
               COALESCE(state, '') as state,
               COALESCE(phone, '') as phone,
               COALESCE(price_range, '₹₹') as price_range,
               COALESCE(rating, 0) as rating,
               COALESCE(total_reviews, 0) as total_reviews,
               COALESCE(is_partner, 0) as is_partner,
               COALESCE(discount_percentage, 0) as discount_percentage,
               COALESCE(offer_details, '') as offer_details
        FROM points_of_interest 
        WHERE id = ? AND is_active = 1
    `, id).Scan(
        &place.ID, &place.Name, &place.Type, &place.Category,
        &place.Latitude, &place.Longitude, &place.Address, &place.City,
        &place.State, &place.Phone, &place.PriceRange, &place.Rating,
        &place.TotalReviews, &place.IsPartner, &place.DiscountPercent, &place.OfferDetails)
    
    if err != nil {
        log.Println("Place not found - ID:", id, "Error:", err)
        return c.Redirect("/trip-planner")
    }
    
    return c.Render("place-detail", fiber.Map{
        "CurrentUser": currentUser,
        "Place":       place,
        "Settings":    settings,
        "IsAdmin":     currentUser != nil && currentUser.IsAdmin,
    })
})

	// Admin endpoint to seed POIs
	app.Get("/admin/seed-pois", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil || !currentUser.IsAdmin {
			return c.Status(403).SendString("Access denied")
		}
		
		pois := []struct {
			name, poiType, category, address, city, phone, priceRange string
			lat, lng, rating float64
			discount int
			offerDetails string
		}{
			{"Hotel Himalayan View", "hotel", "luxury", "Mall Road, Near Bus Stand", "Manali", "+91 9812345601", "₹₹₹", 32.2432, 77.1896, 4.5, 15, "15% discount for Highway Cruizzers members"},
			{"Zostel Manali", "hotel", "budget", "Old Manali", "Manali", "+91 9812345602", "₹", 32.2420, 77.1880, 4.3, 10, "10% off on dorm beds and private rooms"},
			{"Johnson's Cafe", "restaurant", "cafe", "Club House Road", "Manali", "+91 9812345605", "₹₹", 32.2410, 77.1870, 4.6, 10, "10% off on total bill"},
			{"The Lazy Dog", "restaurant", "cafe", "Old Manali", "Manali", "+91 9812345606", "₹₹", 32.2405, 77.1865, 4.4, 5, "5% discount on food and beverages"},
			{"Royal Enfield Service Center", "service_center", "bike", "NH 44, Near Petrol Pump", "Manali", "+91 9812345609", "₹₹", 32.2440, 77.1910, 4.3, 5, "5% discount on service and spare parts"},
			{"EV Charging Station", "charging_point", "electric", "HP Petrol Pump, NH 44", "Manali", "+91 9812345612", "₹", 32.2450, 77.1920, 4.0, 5, "5% discount on charging"},
			{"Himalayan Spa & Wellness", "relax_zone", "spa", "Mall Road", "Manali", "+91 9812345614", "₹₹₹", 32.2430, 77.1900, 4.5, 20, "20% off on massage and spa treatments"},
		}
		
		for _, p := range pois {
			_, err := db.Exec(`INSERT OR IGNORE INTO points_of_interest (name, type, category, address, city, phone, price_range, latitude, longitude, rating, is_partner, discount_percentage, offer_details, is_active) 
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?, 1)`,
				p.name, p.poiType, p.category, p.address, p.city, p.phone, p.priceRange, p.lat, p.lng, p.rating, p.discount, p.offerDetails)
			if err != nil {
				log.Println("Error seeding POI:", err)
			}
		}
		
		return c.SendString("✅ POIs seeded successfully! <a href='/admin'>Go back</a>")
	})

	// Get single POI for editing
	app.Get("/admin/poi/:id", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil || !currentUser.IsAdmin {
			return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
		}
		
		id := c.Params("id")
		var poi struct {
			ID              int
			Name            string
			Type            string
			Category        string
			Latitude        float64
			Longitude       float64
			Address         string
			City            string
			State           string
			Phone           string
			PriceRange      string
			Rating          float64
			IsPartner       bool
			DiscountPercent int
			OfferDetails    string
			IsActive        bool
		}
		
		err := db.QueryRow(`
			SELECT id, name, type, COALESCE(category, ''), latitude, longitude, 
				   COALESCE(address, ''), COALESCE(city, ''), COALESCE(state, ''),
				   COALESCE(phone, ''), COALESCE(price_range, '₹₹'), COALESCE(rating, 0),
				   is_partner, COALESCE(discount_percentage, 0), COALESCE(offer_details, ''), is_active
			FROM points_of_interest WHERE id = ?
		`, id).Scan(&poi.ID, &poi.Name, &poi.Type, &poi.Category, &poi.Latitude, &poi.Longitude,
			&poi.Address, &poi.City, &poi.State, &poi.Phone, &poi.PriceRange, &poi.Rating,
			&poi.IsPartner, &poi.DiscountPercent, &poi.OfferDetails, &poi.IsActive)
		
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"success": false, "error": "POI not found"})
		}
		
		return c.JSON(fiber.Map{"success": true, "poi": poi})
	})

	// Add new POI
	app.Post("/admin/poi/add", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil || !currentUser.IsAdmin {
			return c.Status(403).SendString("Access denied")
		}
		
		name := c.FormValue("name")
		poiType := c.FormValue("type")
		category := c.FormValue("category")
		latitude, _ := strconv.ParseFloat(c.FormValue("latitude"), 64)
		longitude, _ := strconv.ParseFloat(c.FormValue("longitude"), 64)
		address := c.FormValue("address")
		city := c.FormValue("city")
		state := c.FormValue("state")
		phone := c.FormValue("phone")
		priceRange := c.FormValue("price_range")
		rating, _ := strconv.ParseFloat(c.FormValue("rating"), 64)
		discount, _ := strconv.Atoi(c.FormValue("discount_percentage"))
		offerDetails := c.FormValue("offer_details")
		isPartner := c.FormValue("is_partner") == "true" || c.FormValue("is_partner") == "on"
		isActive := c.FormValue("is_active") == "true" || c.FormValue("is_active") == "on"
		
		if name == "" || latitude == 0 || longitude == 0 {
			return c.Status(400).SendString("❌ Name, latitude and longitude are required")
		}
		
		_, err := db.Exec(`
			INSERT INTO points_of_interest (name, type, category, latitude, longitude, address, 
				city, state, phone, price_range, rating, is_partner, discount_percentage, 
				offer_details, is_active)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, name, poiType, category, latitude, longitude, address, city, state, phone,
			priceRange, rating, isPartner, discount, offerDetails, isActive)
		
		if err != nil {
			log.Println("Error adding POI:", err)
			return c.Status(500).SendString("❌ Failed to add location: " + err.Error())
		}
		
		return c.SendString("✅ Location added successfully!")
	})

	// Update POI
	app.Post("/admin/poi/update/:id", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil || !currentUser.IsAdmin {
			return c.Status(403).SendString("Access denied")
		}
		
		id := c.Params("id")
		name := c.FormValue("name")
		poiType := c.FormValue("type")
		category := c.FormValue("category")
		latitude, _ := strconv.ParseFloat(c.FormValue("latitude"), 64)
		longitude, _ := strconv.ParseFloat(c.FormValue("longitude"), 64)
		address := c.FormValue("address")
		city := c.FormValue("city")
		state := c.FormValue("state")
		phone := c.FormValue("phone")
		priceRange := c.FormValue("price_range")
		rating, _ := strconv.ParseFloat(c.FormValue("rating"), 64)
		discount, _ := strconv.Atoi(c.FormValue("discount_percentage"))
		offerDetails := c.FormValue("offer_details")
		isPartner := c.FormValue("is_partner") == "true" || c.FormValue("is_partner") == "on"
		isActive := c.FormValue("is_active") == "true" || c.FormValue("is_active") == "on"
		
		_, err := db.Exec(`
			UPDATE points_of_interest SET
				name = ?, type = ?, category = ?, latitude = ?, longitude = ?, 
				address = ?, city = ?, state = ?, phone = ?, price_range = ?, 
				rating = ?, is_partner = ?, discount_percentage = ?, 
				offer_details = ?, is_active = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`, name, poiType, category, latitude, longitude, address, city, state, phone,
			priceRange, rating, isPartner, discount, offerDetails, isActive, id)
		
		if err != nil {
			return c.Status(500).SendString("❌ Failed to update location")
		}
		
		return c.SendString("✅ Location updated successfully!")
	})

	// Toggle POI active status
	app.Post("/admin/poi/toggle/:id", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil || !currentUser.IsAdmin {
			return c.Status(403).SendString("Access denied")
		}
		
		id := c.Params("id")
		db.Exec("UPDATE points_of_interest SET is_active = NOT is_active WHERE id = ?", id)
		return c.SendString("✅ Location status toggled")
	})

	// Delete POI
	app.Post("/admin/poi/delete/:id", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil || !currentUser.IsAdmin {
			return c.Status(403).SendString("Access denied")
		}
		
		id := c.Params("id")
		db.Exec("DELETE FROM points_of_interest WHERE id = ?", id)
		return c.SendString("✅ Location deleted")
	})

	// API: Save trip to database
	app.Post("/api/save-trip", func(c *fiber.Ctx) error {
		currentUser := getCurrentUser(c)
		if currentUser == nil {
			return c.Status(401).JSON(fiber.Map{"error": "Please login"})
		}
		
		var data struct {
			Title         string  `json:"title"`
			StartLocation string  `json:"start_location"`
			EndLocation   string  `json:"end_location"`
			DistanceKm    float64 `json:"distance_km"`
			EstimatedTime string  `json:"estimated_time"`
		}
		
		if err := c.BodyParser(&data); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid data"})
		}
		
		_, err := db.Exec(`INSERT INTO user_trips (user_id, title, start_location, end_location, distance_km, estimated_time) 
						   VALUES (?, ?, ?, ?, ?, ?)`,
			currentUser.ID, data.Title, data.StartLocation, data.EndLocation, data.DistanceKm, data.EstimatedTime)
		
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to save trip"})
		}
		
		return c.JSON(fiber.Map{"success": true})
	})

	// ==================== AUTHENTICATION ====================
	
	app.Post("/login", func(c *fiber.Ctx) error {
		email := c.FormValue("email")
		password := c.FormValue("password")
		ip := c.IP()
		userAgent := c.Get("User-Agent")
		
		key := ip + ":" + email
		
		if attempt, exists := loginAttempts[key]; exists {
			if time.Now().Before(attempt.LockedUntil) {
				return c.Status(429).SendString(`<div class="p-4 text-red-500 text-center">Too many attempts. Try again later.</div>`)
			}
		}

		var user User
		var hash string
		err := db.QueryRow("SELECT id, username, handle, is_admin, is_active, password_hash, is_verified FROM users WHERE email = ?", email).
			Scan(&user.ID, &user.Username, &user.Handle, &user.IsAdmin, &user.IsActive, &hash, &user.IsVerified)
		
		if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
			if attempt, exists := loginAttempts[key]; exists {
				attempt.Count++
				attempt.LastTry = time.Now()
				if attempt.Count >= MaxLoginAttempts {
					attempt.LockedUntil = time.Now().Add(LockoutDuration)
				}
				loginAttempts[key] = attempt
			} else {
				loginAttempts[key] = struct {
					Count       int
					LastTry     time.Time
					LockedUntil time.Time
				}{Count: 1, LastTry: time.Now(), LockedUntil: time.Time{}}
			}
			
			db.Exec(`INSERT INTO security_logs (event_type, email, ip_address, user_agent) VALUES (?, ?, ?, ?)`,
				"failed_login", email, ip, userAgent)
			return c.Status(401).SendString(`<div class="p-4 text-red-500 text-center">Invalid credentials</div>`)
		}
		
		if !user.IsActive {
			return c.Status(403).SendString(`<div class="p-4 text-red-500 text-center">Account deactivated. Contact support.</div>`)
		}
		
		sessionToken, _ := generateSecureToken()
		expiresAt := time.Now().Add(7 * 24 * time.Hour)
		_, err = db.Exec(`INSERT INTO sessions (token, user_id, ip_address, user_agent, expires_at) VALUES (?, ?, ?, ?, ?)`,
			sessionToken, user.ID, ip, userAgent, expiresAt)
		
		if err != nil {
			log.Println("Session creation error:", err)
			return c.Status(500).SendString(`<div class="p-4 text-red-500 text-center">Login failed</div>`)
		}
		
		c.Cookie(&fiber.Cookie{
			Name:     "auth_token",
			Value:    sessionToken,
			Path:     "/",
			HTTPOnly: true,
			Secure:   false,
			SameSite: "Lax",
			MaxAge:   7 * 24 * 60 * 60,
		})
		
		delete(loginAttempts, key)
		
		db.Exec(`INSERT INTO security_logs (event_type, user_id, ip_address, user_agent) VALUES (?, ?, ?, ?)`,
			"successful_login", user.ID, ip, userAgent)
		
		return c.SendString(`<div class="p-4 text-emerald-600 text-center">Logged in!<script>window.location.reload()</script></div>`)
	})

	app.Post("/signup", func(c *fiber.Ctx) error {
		username := c.FormValue("username")
		handle := c.FormValue("handle")
		email := c.FormValue("email")
		phone := c.FormValue("phone")
		password := c.FormValue("password")
		bikeModel := c.FormValue("bike_model")
		ridingExp := c.FormValue("riding_exp")

		if !isValidEmail(email) {
			return c.SendString(`<div class="p-4 text-red-500 text-center">Invalid email format</div>`)
		}
		
		if err := validatePasswordStrength(password); err != nil {
			return c.SendString(`<div class="p-4 text-red-500 text-center">` + err.Error() + `</div>`)
		}

		hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		_, err := db.Exec(`INSERT INTO users (username, handle, email, phone, password_hash, credits, bike_model, riding_exp) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			username, handle, email, phone, hash, 500, bikeModel, ridingExp)
		if err != nil {
			return c.SendString(`<div class="p-4 text-red-500 text-center">Handle or email already taken</div>`)
		}

		return c.SendString(`<div class="p-4 text-emerald-600 text-center">Account created! Please login.<script>window.location.reload()</script></div>`)
	})

	app.Post("/logout", func(c *fiber.Ctx) error {
		token := c.Cookies("auth_token")
		if token != "" {
			db.Exec("DELETE FROM sessions WHERE token = ?", token)
		}
		c.ClearCookie("auth_token")
		return c.SendString(`<div class="p-4 text-center">Logged out<script>window.location.reload()</script></div>`)
	})


// ==================== FOLLOW/UNFOLLOW ROUTES ====================

// Follow a user
app.Post("/follow/:userId", func(c *fiber.Ctx) error {
    currentUser := getCurrentUser(c)
    if currentUser == nil {
        return c.Status(401).JSON(fiber.Map{
            "success": false, 
            "message": "Please login to follow users",
        })
    }
    
    targetUserId := c.Params("userId")
    userIdInt, err := strconv.Atoi(targetUserId)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid user ID"})
    }
    
    if currentUser.ID == userIdInt {
        return c.JSON(fiber.Map{"success": false, "message": "You cannot follow yourself"})
    }
    
    // Check if already following
    var count int
    db.QueryRow("SELECT COUNT(*) FROM user_follows WHERE follower_id = ? AND following_id = ?", 
        currentUser.ID, userIdInt).Scan(&count)
    
    if count == 0 {
        _, err := db.Exec("INSERT INTO user_follows (follower_id, following_id) VALUES (?, ?)", 
            currentUser.ID, userIdInt)
        if err != nil {
            return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to follow user"})
        }
    }
    
    return c.JSON(fiber.Map{"success": true, "message": "Now following user"})
})

// Unfollow a user
app.Post("/unfollow/:userId", func(c *fiber.Ctx) error {
    currentUser := getCurrentUser(c)
    if currentUser == nil {
        return c.Status(401).JSON(fiber.Map{
            "success": false, 
            "message": "Please login to unfollow users",
        })
    }
    
    targetUserId := c.Params("userId")
    userIdInt, err := strconv.Atoi(targetUserId)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid user ID"})
    }
    
    _, err = db.Exec("DELETE FROM user_follows WHERE follower_id = ? AND following_id = ?", 
        currentUser.ID, userIdInt)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to unfollow user"})
    }
    
    return c.JSON(fiber.Map{"success": true, "message": "Unfollowed user"})
})

// Check if current user is following another user
app.Get("/api/check-follow/:userId", func(c *fiber.Ctx) error {
    currentUser := getCurrentUser(c)
    if currentUser == nil {
        return c.JSON(fiber.Map{"isFollowing": false})
    }
    
    targetUserId := c.Params("userId")
    userIdInt, _ := strconv.Atoi(targetUserId)
    
    var count int
    db.QueryRow("SELECT COUNT(*) FROM user_follows WHERE follower_id = ? AND following_id = ?", 
        currentUser.ID, userIdInt).Scan(&count)
    
    return c.JSON(fiber.Map{"isFollowing": count > 0})
})

// Get follower count for a user
app.Get("/api/followers/:userId", func(c *fiber.Ctx) error {
    targetUserId := c.Params("userId")
    userIdInt, _ := strconv.Atoi(targetUserId)
    
    var followers, following int
    db.QueryRow("SELECT COUNT(*) FROM user_follows WHERE following_id = ?", userIdInt).Scan(&followers)
    db.QueryRow("SELECT COUNT(*) FROM user_follows WHERE follower_id = ?", userIdInt).Scan(&following)
    
    return c.JSON(fiber.Map{
        "followers": followers,
        "following": following,
    })
})
	// ==================== START SERVER ====================
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("🏍️ Highway Cruizzers running on http://localhost:%s", port)
	log.Fatal(app.Listen(":" + port))
}