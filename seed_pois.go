// seed_pois.go - Run once to seed POI and Offer data
package main

import (
    "database/sql"
    "log"
    _ "github.com/mattn/go-sqlite3"
)

func main() {
    db, err := sql.Open("sqlite3", "./highwaycruizzers.db")
    if err != nil {
        log.Fatal("Error opening database:", err)
    }
    defer db.Close()

    // Clear existing data
    log.Println("Clearing existing POI data...")
    db.Exec("DELETE FROM partner_offers")
    db.Exec("DELETE FROM points_of_interest")

    // Insert Hotels
    log.Println("Seeding Hotels...")
    hotels := []struct {
        name, category, address, city, phone, priceRange string
        lat, lng, rating float64
        discount int
        offer string
    }{
        {"Hotel Himalayan View", "luxury", "Mall Road, Near Bus Stand", "Manali", "9812345601", "₹₹₹", 32.2432, 77.1896, 4.5, 15, "15% discount on room booking for Highway Cruizzers members"},
        {"Zostel Manali", "budget", "Old Manali", "Manali", "9812345602", "₹", 32.2420, 77.1880, 4.3, 10, "10% off on dorm beds and private rooms"},
        {"JW Marriott Chandigarh", "luxury", "Sector 35", "Chandigarh", "9812345603", "₹₹₹₹", 30.7333, 76.7794, 4.7, 0, ""},
        {"Hotel City Park", "mid-range", "Connaught Place", "Delhi", "9812345604", "₹₹", 28.6139, 77.2090, 4.0, 20, "20% off on weekend stays"},
    }

    for _, h := range hotels {
        _, err := db.Exec(`INSERT INTO points_of_interest 
            (name, type, category, latitude, longitude, address, city, phone, price_range, rating, is_partner, discount_percentage, offer_details, is_active) 
            VALUES (?, 'hotel', ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?, 1)`,
            h.name, h.category, h.lat, h.lng, h.address, h.city, h.phone, h.priceRange, h.rating, h.discount, h.offer)
        if err != nil {
            log.Println("Error seeding hotel:", err)
        }
    }

    // Insert Restaurants
    log.Println("Seeding Restaurants...")
    restaurants := []struct {
        name, category, address, city, phone, priceRange string
        lat, lng, rating float64
        discount int
        offer string
    }{
        {"Johnson's Cafe", "cafe", "Club House Road", "Manali", "9812345605", "₹₹", 32.2410, 77.1870, 4.6, 10, "10% off on total bill"},
        {"The Lazy Dog", "cafe", "Old Manali", "Manali", "9812345606", "₹₹", 32.2405, 77.1865, 4.4, 5, "5% discount on food and beverages"},
        {"Dhaba 29", "dhaba", "NH 44", "Kullu", "9812345607", "₹", 31.9580, 77.1100, 4.2, 0, ""},
        {"Saravana Bhavan", "veg", "Connaught Place", "Delhi", "9812345608", "₹", 28.6140, 77.2100, 4.3, 10, "10% off on orders above ₹500"},
    }

    for _, r := range restaurants {
        _, err := db.Exec(`INSERT INTO points_of_interest 
            (name, type, category, latitude, longitude, address, city, phone, price_range, rating, is_partner, discount_percentage, offer_details, is_active) 
            VALUES (?, 'restaurant', ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?, 1)`,
            r.name, r.category, r.lat, r.lng, r.address, r.city, r.phone, r.priceRange, r.rating, r.discount, r.offer)
        if err != nil {
            log.Println("Error seeding restaurant:", err)
        }
    }

    // Insert Service Centers
    log.Println("Seeding Service Centers...")
    services := []struct {
        name, category, address, city, phone, priceRange string
        lat, lng, rating float64
        discount int
        offer string
    }{
        {"Royal Enfield Service Center", "bike", "NH 44, Near Petrol Pump", "Manali", "9812345609", "₹₹", 32.2440, 77.1910, 4.3, 5, "5% discount on service and spare parts"},
        {"Bike Point Service", "multi-brand", "Mall Road", "Kullu", "9812345610", "₹₹", 31.9600, 77.1110, 4.1, 10, "10% off on all services"},
        {"Tyre Pro", "tyres", "NH 44", "Mandi", "9812345611", "₹₹", 31.7100, 76.9300, 4.2, 0, ""},
    }

    for _, s := range services {
        _, err := db.Exec(`INSERT INTO points_of_interest 
            (name, type, category, latitude, longitude, address, city, phone, price_range, rating, is_partner, discount_percentage, offer_details, is_active) 
            VALUES (?, 'service_center', ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?, 1)`,
            s.name, s.category, s.lat, s.lng, s.address, s.city, s.phone, s.priceRange, s.rating, s.discount, s.offer)
        if err != nil {
            log.Println("Error seeding service center:", err)
        }
    }

    // Insert Charging Points
    log.Println("Seeding Charging Points...")
    charging := []struct {
        name, category, address, city, phone, priceRange string
        lat, lng, rating float64
        discount int
        offer string
    }{
        {"EV Charging Station", "electric", "HP Petrol Pump, NH 44", "Manali", "9812345612", "₹", 32.2450, 77.1920, 4.0, 5, "5% discount on charging"},
        {"Green EV Charging", "electric", "Near Bus Stand", "Kullu", "9812345613", "₹", 31.9590, 77.1120, 4.1, 0, ""},
    }

    for _, c := range charging {
        _, err := db.Exec(`INSERT INTO points_of_interest 
            (name, type, category, latitude, longitude, address, city, phone, price_range, rating, is_partner, discount_percentage, offer_details, is_active) 
            VALUES (?, 'charging_point', ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?, 1)`,
            c.name, c.category, c.lat, c.lng, c.address, c.city, c.phone, c.priceRange, c.rating, c.discount, c.offer)
        if err != nil {
            log.Println("Error seeding charging point:", err)
        }
    }

    // Insert Relax Zones
    log.Println("Seeding Relax Zones...")
    relax := []struct {
        name, category, address, city, phone, priceRange string
        lat, lng, rating float64
        discount int
        offer string
    }{
        {"Himalayan Spa & Wellness", "spa", "Mall Road", "Manali", "9812345614", "₹₹₹", 32.2430, 77.1900, 4.5, 20, "20% off on massage and spa treatments"},
        {"Yoga House", "yoga", "Old Manali", "Manali", "9812345615", "₹₹", 32.2415, 77.1860, 4.4, 15, "15% off on yoga classes"},
    }

    for _, r := range relax {
        _, err := db.Exec(`INSERT INTO points_of_interest 
            (name, type, category, latitude, longitude, address, city, phone, price_range, rating, is_partner, discount_percentage, offer_details, is_active) 
            VALUES (?, 'relax_zone', ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?, 1)`,
            r.name, r.category, r.lat, r.lng, r.address, r.city, r.phone, r.priceRange, r.rating, r.discount, r.offer)
        if err != nil {
            log.Println("Error seeding relax zone:", err)
        }
    }

    // Insert Partner Offers
    log.Println("Seeding Partner Offers...")
    
    // Get IDs for offers
    var hotelID, zostelID, cafeID, serviceID, spaID int
    
    db.QueryRow("SELECT id FROM points_of_interest WHERE name = 'Hotel Himalayan View' LIMIT 1").Scan(&hotelID)
    db.QueryRow("SELECT id FROM points_of_interest WHERE name = 'Zostel Manali' LIMIT 1").Scan(&zostelID)
    db.QueryRow("SELECT id FROM points_of_interest WHERE name = \"Johnson's Cafe\" LIMIT 1").Scan(&cafeID)
    db.QueryRow("SELECT id FROM points_of_interest WHERE name = 'Royal Enfield Service Center' LIMIT 1").Scan(&serviceID)
    db.QueryRow("SELECT id FROM points_of_interest WHERE name = 'Himalayan Spa & Wellness' LIMIT 1").Scan(&spaID)

    offers := []struct {
        partnerID int
        title, description, discountType, code string
        discountValue int
    }{
        {hotelID, "Himalayan View Hotel Special", "Flat 15% off on all room bookings", "percentage", "HC15OFF", 15},
        {zostelID, "Zostel Member Discount", "10% off for Highway Cruizzers members", "percentage", "ZOSTEL10", 10},
        {cafeID, "Johnson's Cafe Deal", "10% off on total bill", "percentage", "JOHNSON10", 10},
        {serviceID, "Royal Enfield Service Discount", "5% off on service and parts", "percentage", "RE5OFF", 5},
        {spaID, "Spa & Wellness Offer", "20% off on all treatments", "percentage", "SPA20", 20},
    }

    for _, offer := range offers {
        if offer.partnerID > 0 {
            _, err := db.Exec(`INSERT INTO partner_offers 
                (partner_id, title, description, discount_type, discount_value, code, valid_to, is_active) 
                VALUES (?, ?, ?, ?, ?, ?, datetime('now', '+90 days'), 1)`,
                offer.partnerID, offer.title, offer.description, offer.discountType, offer.discountValue, offer.code)
            if err != nil {
                log.Println("Error seeding offer:", err)
            }
        }
    }

    // Verify
    var poiCount, offerCount int
    db.QueryRow("SELECT COUNT(*) FROM points_of_interest").Scan(&poiCount)
    db.QueryRow("SELECT COUNT(*) FROM partner_offers").Scan(&offerCount)
    
    log.Printf("✅ Success! Seeded %d POIs and %d offers", poiCount, offerCount)
}