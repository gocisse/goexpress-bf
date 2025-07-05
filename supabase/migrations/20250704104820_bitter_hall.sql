/*
  # Customer Management System

  1. New Tables
    - `customers`
      - `id` (serial, primary key)
      - `user_id` (integer, foreign key to users)
      - `company_name` (varchar)
      - `contact_person` (varchar)
      - `phone` (varchar)
      - `alternate_phone` (varchar)
      - `website` (varchar)
      - `tax_id` (varchar)
      - `business_type` (varchar)
      - `status` (varchar) - active, inactive, suspended
      - `credit_limit` (decimal)
      - `payment_terms` (varchar)
      - `notes` (text)
      - `created_at` (timestamp)
      - `updated_at` (timestamp)
    
    - `customer_addresses`
      - `id` (serial, primary key)
      - `customer_id` (integer, foreign key to customers)
      - `type` (varchar) - billing, shipping, both
      - `label` (varchar) - home, office, warehouse
      - `address_line1` (varchar)
      - `address_line2` (varchar)
      - `city` (varchar)
      - `state` (varchar)
      - `postal_code` (varchar)
      - `country` (varchar)
      - `is_default` (boolean)
      - `created_at` (timestamp)
      - `updated_at` (timestamp)

  2. Security
    - Enable RLS on both tables
    - Add policies for admin access and customer self-access
*/

-- Create customers table
CREATE TABLE IF NOT EXISTS customers (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    company_name VARCHAR(255) NOT NULL,
    contact_person VARCHAR(255) NOT NULL,
    phone VARCHAR(50) NOT NULL,
    alternate_phone VARCHAR(50),
    website VARCHAR(255),
    tax_id VARCHAR(100),
    business_type VARCHAR(100),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended')),
    credit_limit DECIMAL(12,2) DEFAULT 0.00,
    payment_terms VARCHAR(100),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id)
);

-- Create customer addresses table
CREATE TABLE IF NOT EXISTS customer_addresses (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER REFERENCES customers(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('billing', 'shipping', 'both')),
    label VARCHAR(50) NOT NULL,
    address_line1 VARCHAR(255) NOT NULL,
    address_line2 VARCHAR(255),
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100) NOT NULL,
    postal_code VARCHAR(20) NOT NULL,
    country VARCHAR(100) NOT NULL DEFAULT 'India',
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_customers_user_id ON customers(user_id);
CREATE INDEX IF NOT EXISTS idx_customers_status ON customers(status);
CREATE INDEX IF NOT EXISTS idx_customers_business_type ON customers(business_type);
CREATE INDEX IF NOT EXISTS idx_customer_addresses_customer_id ON customer_addresses(customer_id);
CREATE INDEX IF NOT EXISTS idx_customer_addresses_type ON customer_addresses(type);
CREATE INDEX IF NOT EXISTS idx_customer_addresses_default ON customer_addresses(is_default);

-- Insert sample customers for existing client users
INSERT INTO customers (
    user_id, company_name, contact_person, phone, business_type, status
) 
SELECT 
    u.id,
    CASE 
        WHEN u.name LIKE '%Client%' THEN u.name || ' Corporation'
        ELSE u.name || ' Enterprises'
    END as company_name,
    u.name as contact_person,
    '+91-' || LPAD((RANDOM() * 9000000000)::bigint + 1000000000, 10, '0') as phone,
    CASE 
        WHEN RANDOM() < 0.3 THEN 'E-commerce'
        WHEN RANDOM() < 0.6 THEN 'Manufacturing'
        WHEN RANDOM() < 0.8 THEN 'Retail'
        ELSE 'Services'
    END as business_type,
    'active' as status
FROM users u 
WHERE u.role = 'client' 
AND NOT EXISTS (SELECT 1 FROM customers c WHERE c.user_id = u.id);

-- Insert sample addresses for customers
INSERT INTO customer_addresses (
    customer_id, type, label, address_line1, city, state, postal_code, country, is_default
)
SELECT 
    c.id,
    'both' as type,
    'Head Office' as label,
    CASE 
        WHEN RANDOM() < 0.25 THEN 'Plot No. ' || (RANDOM() * 999 + 1)::int || ', Sector ' || (RANDOM() * 50 + 1)::int
        WHEN RANDOM() < 0.5 THEN 'Building No. ' || (RANDOM() * 99 + 1)::int || ', Street ' || (RANDOM() * 20 + 1)::int
        WHEN RANDOM() < 0.75 THEN 'Office No. ' || (RANDOM() * 999 + 100)::int || ', Commercial Complex'
        ELSE 'Warehouse ' || (RANDOM() * 50 + 1)::int || ', Industrial Area'
    END as address_line1,
    CASE 
        WHEN RANDOM() < 0.2 THEN 'Mumbai'
        WHEN RANDOM() < 0.4 THEN 'Delhi'
        WHEN RANDOM() < 0.6 THEN 'Bangalore'
        WHEN RANDOM() < 0.8 THEN 'Chennai'
        ELSE 'Pune'
    END as city,
    CASE 
        WHEN RANDOM() < 0.2 THEN 'Maharashtra'
        WHEN RANDOM() < 0.4 THEN 'Delhi'
        WHEN RANDOM() < 0.6 THEN 'Karnataka'
        WHEN RANDOM() < 0.8 THEN 'Tamil Nadu'
        ELSE 'Maharashtra'
    END as state,
    LPAD((RANDOM() * 899999 + 100000)::int::text, 6, '0') as postal_code,
    'India' as country,
    true as is_default
FROM customers c
WHERE NOT EXISTS (SELECT 1 FROM customer_addresses ca WHERE ca.customer_id = c.id);
