-- Create the sensors table
CREATE TABLE IF NOT EXISTS sensors (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL,
    location VARCHAR(100) NOT NULL,
    value FLOAT DEFAULT 0,
    unit VARCHAR(20),
    status VARCHAR(20) NOT NULL DEFAULT 'inactive',
    last_updated TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_sensors_type ON sensors(type);
CREATE INDEX IF NOT EXISTS idx_sensors_location ON sensors(location);
CREATE INDEX IF NOT EXISTS idx_sensors_status ON sensors(status);

-- Insert sample data
INSERT INTO sensors (name, type, location, value, unit, status) VALUES
('Living Room Temperature Sensor', 'temperature', 'Living Room', 22.0, '°C', 'active'),
('Bedroom Temperature Sensor', 'temperature', 'Bedroom', 20.5, '°C', 'active'),
('Kitchen Temperature Sensor', 'temperature', 'Kitchen', 23.5, '°C', 'active')
ON CONFLICT DO NOTHING;
