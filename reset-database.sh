#!/bin/bash

# GoExpress Database Reset Script
# This script completely wipes the database and recreates it with fresh schema

echo "🗄️ GoExpress Database Reset & Migration"
echo "======================================="

# Database connection details
DB_HOST="localhost"
DB_USER="goexpress"
DB_NAME="goexpress_db"
DB_PASSWORD="goexpress"

echo "⚠️  WARNING: This will completely wipe the existing database!"
echo "Database: $DB_NAME"
echo "User: $DB_USER"
echo ""

# Confirmation prompt
read -p "Are you sure you want to proceed? (yes/no): " confirm
if [ "$confirm" != "yes" ]; then
    echo "❌ Operation cancelled"
    exit 1
fi

echo ""
echo "🧹 Step 1: Dropping existing database..."

# Drop and recreate database
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U postgres << EOF
-- Terminate all connections to the database
SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '$DB_NAME';

-- Drop and recreate database
DROP DATABASE IF EXISTS $DB_NAME;
CREATE DATABASE $DB_NAME;

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;

\q
EOF

if [ $? -eq 0 ]; then
    echo "✅ Database dropped and recreated successfully"
else
    echo "❌ Failed to drop/recreate database"
    exit 1
fi

echo ""
echo "🏗️ Step 2: Creating fresh schema..."

# Create the new enhanced schema
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME << 'EOF'
-- GoExpress Enhanced Database Schema for Burkina Faso Operations

-- Users table (enhanced)
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) CHECK (role IN ('admin', 'driver', 'client')) NOT NULL,
    phone VARCHAR(50),
    avatar_url VARCHAR(500),
    is_active BOOLEAN DEFAULT true,
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Zones table (enhanced for West Africa)
CREATE TABLE IF NOT EXISTS zones (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price_per_kg DECIMAL(10,2) NOT NULL,
    countries TEXT[], -- Array of countries covered
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Customers table (enhanced business profiles)
CREATE TABLE IF NOT EXISTS customers (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    company_name VARCHAR(255) NOT NULL,
    contact_person VARCHAR(255) NOT NULL,
    phone VARCHAR(50) NOT NULL,
    alternate_phone VARCHAR(50),
    email VARCHAR(255),
    website VARCHAR(255),
    tax_id VARCHAR(100),
    business_type VARCHAR(100),
    business_registration VARCHAR(100),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended')),
    credit_limit DECIMAL(12,2) DEFAULT 0.00,
    payment_terms VARCHAR(100),
    preferred_currency VARCHAR(3) DEFAULT 'XOF',
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id)
);

-- Customer addresses table
CREATE TABLE IF NOT EXISTS customer_addresses (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER REFERENCES customers(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('billing', 'shipping', 'both')),
    label VARCHAR(50) NOT NULL,
    address_line1 VARCHAR(255) NOT NULL,
    address_line2 VARCHAR(255),
    city VARCHAR(100) NOT NULL,
    state_province VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(100) NOT NULL DEFAULT 'Burkina Faso',
    is_default BOOLEAN DEFAULT FALSE,
    coordinates JSONB, -- {latitude: float, longitude: float}
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Drivers table (enhanced with vehicle and performance data)
CREATE TABLE IF NOT EXISTS drivers (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    license_number VARCHAR(100),
    license_expiry DATE,
    vehicle_type VARCHAR(50),
    vehicle_number VARCHAR(50),
    vehicle_model VARCHAR(100),
    vehicle_year INTEGER,
    insurance_number VARCHAR(100),
    insurance_expiry DATE,
    status VARCHAR(20) DEFAULT 'available' CHECK (status IN ('available', 'busy', 'offline', 'suspended')),
    current_location VARCHAR(255),
    coordinates JSONB, -- {latitude: float, longitude: float}
    rating DECIMAL(3,2) DEFAULT 0.00,
    total_deliveries INTEGER DEFAULT 0,
    successful_deliveries INTEGER DEFAULT 0,
    emergency_contact_name VARCHAR(255),
    emergency_contact_phone VARCHAR(50),
    hire_date DATE DEFAULT CURRENT_DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id)
);

-- Enhanced shipments table
CREATE TABLE IF NOT EXISTS shipments (
    id SERIAL PRIMARY KEY,
    tracking_number VARCHAR(255) UNIQUE NOT NULL,
    origin VARCHAR(255) NOT NULL,
    destination VARCHAR(255) NOT NULL,
    origin_coordinates JSONB, -- {latitude: float, longitude: float}
    destination_coordinates JSONB, -- {latitude: float, longitude: float}
    weight DECIMAL(10,2) NOT NULL,
    dimensions JSONB, -- {length: float, width: float, height: float}
    package_type VARCHAR(50) DEFAULT 'standard',
    zone_id INTEGER REFERENCES zones(id),
    status VARCHAR(50) DEFAULT 'pending',
    priority VARCHAR(20) DEFAULT 'normal' CHECK (priority IN ('low', 'normal', 'high', 'urgent')),
    customer_id INTEGER REFERENCES users(id),
    driver_id INTEGER REFERENCES users(id),
    estimated_cost DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    actual_cost DECIMAL(10,2),
    estimated_delivery_date TIMESTAMP,
    actual_delivery_date TIMESTAMP,
    pickup_date TIMESTAMP,
    special_instructions TEXT,
    fragile BOOLEAN DEFAULT false,
    requires_signature BOOLEAN DEFAULT false,
    insurance_value DECIMAL(10,2) DEFAULT 0.00,
    payment_method VARCHAR(50) DEFAULT 'cash_on_delivery',
    payment_status VARCHAR(20) DEFAULT 'pending' CHECK (payment_status IN ('pending', 'paid', 'refunded')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Enhanced tracking updates table
CREATE TABLE IF NOT EXISTS tracking_updates (
    id SERIAL PRIMARY KEY,
    shipment_id INTEGER REFERENCES shipments(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL,
    location VARCHAR(255),
    coordinates JSONB, -- {latitude: float, longitude: float}
    description TEXT,
    driver_notes TEXT,
    photo_url VARCHAR(500),
    temperature DECIMAL(5,2), -- For temperature-sensitive packages
    humidity DECIMAL(5,2), -- For humidity-sensitive packages
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Driver ratings table
CREATE TABLE IF NOT EXISTS driver_ratings (
    id SERIAL PRIMARY KEY,
    shipment_id INTEGER REFERENCES shipments(id) ON DELETE CASCADE,
    driver_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    customer_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    rating INTEGER CHECK (rating >= 1 AND rating <= 5),
    feedback TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    type VARCHAR(50) DEFAULT 'info' CHECK (type IN ('info', 'success', 'warning', 'error')),
    is_read BOOLEAN DEFAULT false,
    related_shipment_id INTEGER REFERENCES shipments(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- System settings table
CREATE TABLE IF NOT EXISTS system_settings (
    id SERIAL PRIMARY KEY,
    key VARCHAR(100) UNIQUE NOT NULL,
    value TEXT,
    description TEXT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_shipments_tracking ON shipments(tracking_number);
CREATE INDEX IF NOT EXISTS idx_shipments_customer ON shipments(customer_id);
CREATE INDEX IF NOT EXISTS idx_shipments_driver ON shipments(driver_id);
CREATE INDEX IF NOT EXISTS idx_shipments_status ON shipments(status);
CREATE INDEX IF NOT EXISTS idx_shipments_priority ON shipments(priority);
CREATE INDEX IF NOT EXISTS idx_shipments_created_at ON shipments(created_at);
CREATE INDEX IF NOT EXISTS idx_tracking_shipment ON tracking_updates(shipment_id);
CREATE INDEX IF NOT EXISTS idx_tracking_timestamp ON tracking_updates(timestamp);
CREATE INDEX IF NOT EXISTS idx_customers_user_id ON customers(user_id);
CREATE INDEX IF NOT EXISTS idx_customers_status ON customers(status);
CREATE INDEX IF NOT EXISTS idx_drivers_user_id ON drivers(user_id);
CREATE INDEX IF NOT EXISTS idx_drivers_status ON drivers(status);
CREATE INDEX IF NOT EXISTS idx_customer_addresses_customer_id ON customer_addresses(customer_id);
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_read ON notifications(is_read);

\q
EOF

if [ $? -eq 0 ]; then
    echo "✅ Schema created successfully"
else
    echo "❌ Failed to create schema"
    exit 1
fi

echo ""
echo "📊 Step 3: Inserting sample data..."

# Insert sample data
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME << 'EOF'
-- Insert West African delivery zones with XOF pricing
INSERT INTO zones (name, description, price_per_kg, countries) VALUES 
('Ouagadougou Express', 'Livraison rapide dans Ouagadougou', 1500.00, ARRAY['Burkina Faso']),
('Burkina National', 'Livraison nationale au Burkina Faso', 2500.00, ARRAY['Burkina Faso']),
('Afrique de l''Ouest', 'Livraison régionale en Afrique de l''Ouest', 4500.00, ARRAY['Mali', 'Niger', 'Côte d''Ivoire', 'Ghana', 'Togo', 'Bénin']),
('Express International', 'Livraison express internationale', 8500.00, ARRAY['France', 'Europe', 'International']),
('Même Jour Ouaga', 'Livraison le jour même à Ouagadougou', 3000.00, ARRAY['Burkina Faso'])
ON CONFLICT DO NOTHING;

-- Insert default admin user (password: goexpress123)
INSERT INTO users (name, email, password_hash, role, phone) VALUES 
('GoExpress Admin', 'admin@goexpress.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin', '+226 70 12 34 56')
ON CONFLICT (email) DO NOTHING;

-- Insert sample driver users
INSERT INTO users (name, email, password_hash, role, phone) VALUES 
('Amadou Traoré', 'amadou.traore@goexpress.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'driver', '+226 70 11 22 33'),
('Fatimata Ouédraogo', 'fatimata.ouedraogo@goexpress.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'driver', '+226 70 44 55 66'),
('Ibrahim Sawadogo', 'ibrahim.sawadogo@goexpress.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'driver', '+226 70 77 88 99')
ON CONFLICT (email) DO NOTHING;

-- Insert sample client users
INSERT INTO users (name, email, password_hash, role, phone) VALUES 
('Entreprise Kaboré SARL', 'contact@kabore-sarl.bf', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'client', '+226 25 30 40 50'),
('Boutique Wend-Kuni', 'info@wendkuni.bf', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'client', '+226 25 60 70 80'),
('Pharmacie du Centre', 'pharmacie.centre@gmail.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'client', '+226 25 90 10 20')
ON CONFLICT (email) DO NOTHING;

-- Insert driver profiles
INSERT INTO drivers (user_id, license_number, vehicle_type, vehicle_number, status, current_location, rating, total_deliveries, successful_deliveries)
SELECT 
    u.id,
    CASE 
        WHEN u.name = 'Amadou Traoré' THEN 'BF001234567'
        WHEN u.name = 'Fatimata Ouédraogo' THEN 'BF002345678'
        WHEN u.name = 'Ibrahim Sawadogo' THEN 'BF003456789'
    END,
    CASE 
        WHEN u.name = 'Amadou Traoré' THEN 'Moto'
        WHEN u.name = 'Fatimata Ouédraogo' THEN 'Voiture'
        WHEN u.name = 'Ibrahim Sawadogo' THEN 'Camionnette'
    END,
    CASE 
        WHEN u.name = 'Amadou Traoré' THEN '11 BF 2024'
        WHEN u.name = 'Fatimata Ouédraogo' THEN '22 BF 2024'
        WHEN u.name = 'Ibrahim Sawadogo' THEN '33 BF 2024'
    END,
    'available',
    'Ouagadougou, Burkina Faso',
    4.5 + (RANDOM() * 0.5),
    FLOOR(RANDOM() * 100) + 50,
    FLOOR(RANDOM() * 95) + 45
FROM users u 
WHERE u.role = 'driver'
ON CONFLICT (user_id) DO NOTHING;

-- Insert customer profiles
INSERT INTO customers (user_id, company_name, contact_person, phone, business_type, status)
SELECT 
    u.id,
    CASE 
        WHEN u.name = 'Entreprise Kaboré SARL' THEN 'Kaboré SARL'
        WHEN u.name = 'Boutique Wend-Kuni' THEN 'Wend-Kuni'
        WHEN u.name = 'Pharmacie du Centre' THEN 'Pharmacie du Centre'
    END,
    CASE 
        WHEN u.name = 'Entreprise Kaboré SARL' THEN 'Moussa Kaboré'
        WHEN u.name = 'Boutique Wend-Kuni' THEN 'Awa Compaoré'
        WHEN u.name = 'Pharmacie du Centre' THEN 'Dr. Salif Ouattara'
    END,
    u.phone,
    CASE 
        WHEN u.name = 'Entreprise Kaboré SARL' THEN 'Commerce'
        WHEN u.name = 'Boutique Wend-Kuni' THEN 'Vente au détail'
        WHEN u.name = 'Pharmacie du Centre' THEN 'Santé'
    END,
    'active'
FROM users u 
WHERE u.role = 'client'
ON CONFLICT (user_id) DO NOTHING;

-- Insert sample customer addresses
INSERT INTO customer_addresses (customer_id, type, label, address_line1, city, country, is_default)
SELECT 
    c.id,
    'both',
    'Siège social',
    CASE 
        WHEN c.company_name = 'Kaboré SARL' THEN 'Avenue Kwame Nkrumah, Secteur 4'
        WHEN c.company_name = 'Wend-Kuni' THEN 'Rue de la Révolution, Secteur 12'
        WHEN c.company_name = 'Pharmacie du Centre' THEN 'Avenue de la Nation, Centre-ville'
    END,
    'Ouagadougou',
    'Burkina Faso',
    true
FROM customers c
ON CONFLICT DO NOTHING;

-- Insert sample shipments
INSERT INTO shipments (
    tracking_number, origin, destination, weight, package_type, zone_id, 
    status, priority, customer_id, estimated_cost, special_instructions
)
SELECT 
    'GEX' || LPAD((ROW_NUMBER() OVER())::text, 8, '0'),
    origins.origin,
    destinations.destination,
    (RANDOM() * 10 + 0.5)::DECIMAL(10,2),
    package_types.package_type,
    (RANDOM() * 5 + 1)::INTEGER,
    statuses.status,
    priorities.priority,
    customers.customer_id,
    (RANDOM() * 50000 + 5000)::DECIMAL(10,2),
    instructions.instruction
FROM 
    (VALUES 
        ('Ouagadougou, Secteur 15'),
        ('Bobo-Dioulasso, Centre'),
        ('Koudougou, Secteur 3'),
        ('Ouagadougou, Zone du Bois')
    ) AS origins(origin),
    (VALUES 
        ('Ouagadougou, Secteur 30'),
        ('Banfora, Centre'),
        ('Ouahigouya, Secteur 2'),
        ('Ouagadougou, Pissy')
    ) AS destinations(destination),
    (VALUES 
        ('standard'),
        ('fragile'),
        ('documents'),
        ('electronics')
    ) AS package_types(package_type),
    (VALUES 
        ('pending'),
        ('confirmed'),
        ('in_transit'),
        ('delivered')
    ) AS statuses(status),
    (VALUES 
        ('normal'),
        ('high'),
        ('urgent'),
        ('low')
    ) AS priorities(priority),
    (SELECT u.id as customer_id FROM users u WHERE u.role = 'client' LIMIT 1) AS customers,
    (VALUES 
        ('Livraison standard'),
        ('Fragile - Manipuler avec précaution'),
        ('Documents importants'),
        ('Appeler avant livraison')
    ) AS instructions(instruction)
LIMIT 12;

-- Insert system settings
INSERT INTO system_settings (key, value, description) VALUES 
('company_name', 'GoExpress', 'Nom de l''entreprise'),
('company_country', 'Burkina Faso', 'Pays principal d''opération'),
('default_currency', 'XOF', 'Devise par défaut (Franc CFA)'),
('timezone', 'Africa/Ouagadougou', 'Fuseau horaire'),
('language_default', 'fr', 'Langue par défaut'),
('max_package_weight', '100', 'Poids maximum des colis (kg)'),
('delivery_radius_km', '500', 'Rayon de livraison (km)'),
('emergency_phone', '+226 70 00 00 00', 'Numéro d''urgence')
ON CONFLICT (key) DO NOTHING;

\q
EOF

if [ $? -eq 0 ]; then
    echo "✅ Sample data inserted successfully"
else
    echo "❌ Failed to insert sample data"
    exit 1
fi

echo ""
echo "🔧 Step 4: Updating sequences..."

# Reset sequences to ensure proper auto-increment
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME << 'EOF'
-- Reset all sequences to ensure proper auto-increment
SELECT setval('users_id_seq', COALESCE((SELECT MAX(id) FROM users), 1));
SELECT setval('zones_id_seq', COALESCE((SELECT MAX(id) FROM zones), 1));
SELECT setval('customers_id_seq', COALESCE((SELECT MAX(id) FROM customers), 1));
SELECT setval('customer_addresses_id_seq', COALESCE((SELECT MAX(id) FROM customer_addresses), 1));
SELECT setval('drivers_id_seq', COALESCE((SELECT MAX(id) FROM drivers), 1));
SELECT setval('shipments_id_seq', COALESCE((SELECT MAX(id) FROM shipments), 1));
SELECT setval('tracking_updates_id_seq', COALESCE((SELECT MAX(id) FROM tracking_updates), 1));
SELECT setval('driver_ratings_id_seq', COALESCE((SELECT MAX(id) FROM driver_ratings), 1));
SELECT setval('notifications_id_seq', COALESCE((SELECT MAX(id) FROM notifications), 1));
SELECT setval('system_settings_id_seq', COALESCE((SELECT MAX(id) FROM system_settings), 1));

\q
EOF

echo "✅ Sequences updated successfully"

echo ""
echo "📋 Step 5: Database summary..."

# Show database summary
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME << 'EOF'
-- Database summary
SELECT 'Users' as table_name, COUNT(*) as record_count FROM users
UNION ALL
SELECT 'Zones', COUNT(*) FROM zones
UNION ALL
SELECT 'Customers', COUNT(*) FROM customers
UNION ALL
SELECT 'Customer Addresses', COUNT(*) FROM customer_addresses
UNION ALL
SELECT 'Drivers', COUNT(*) FROM drivers
UNION ALL
SELECT 'Shipments', COUNT(*) FROM shipments
UNION ALL
SELECT 'Tracking Updates', COUNT(*) FROM tracking_updates
UNION ALL
SELECT 'System Settings', COUNT(*) FROM system_settings
ORDER BY table_name;

\q
EOF

echo ""
echo "🎉 Database reset completed successfully!"
echo "========================================"
echo ""
echo "📊 Database Summary:"
echo "• Fresh GoExpress schema created"
echo "• Enhanced tables for Burkina Faso operations"
echo "• Sample data with West African context"
echo "• XOF (F CFA) currency support"
echo "• Bilingual French/English support"
echo ""
echo "🔐 Default Login Credentials:"
echo "Email: admin@goexpress.com"
echo "Password: goexpress123"
echo ""
echo "🚀 Next Steps:"
echo "1. Restart your Go backend: go run main.go"
echo "2. Restart your admin panel: npm run preview"
echo "3. Login with the admin credentials above"
echo ""
echo "🌍 GoExpress - Rapide, Fiable, Suivi! 🇧🇫"



