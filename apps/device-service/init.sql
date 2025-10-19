-- Create devices table
CREATE TABLE IF NOT EXISTS devices (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    device_type VARCHAR(100) NOT NULL,
    location VARCHAR(255) NOT NULL,
    room VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_devices_location ON devices(location);
CREATE INDEX IF NOT EXISTS idx_devices_type ON devices(device_type);
CREATE INDEX IF NOT EXISTS idx_devices_status ON devices(status);

-- Insert sample devices
INSERT INTO devices (name, device_type, location, room, status) VALUES
('Temperature Sensor 1', 'temperature', 'Living Room', 'Main Floor', 'active'),
('Humidity Sensor 1', 'humidity', 'Bedroom', 'Second Floor', 'active'),
('Motion Detector 1', 'motion', 'Kitchen', 'Main Floor', 'active')
ON CONFLICT DO NOTHING;
