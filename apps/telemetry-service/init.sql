-- Create telemetry table
CREATE TABLE IF NOT EXISTS telemetry (
    id BIGSERIAL PRIMARY KEY,
    device_id INTEGER NOT NULL,
    metric_name VARCHAR(100) NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    unit VARCHAR(50),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for time-series queries
CREATE INDEX IF NOT EXISTS idx_telemetry_device_time ON telemetry(device_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_telemetry_metric_time ON telemetry(metric_name, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_telemetry_timestamp ON telemetry(timestamp DESC);

-- Insert sample telemetry data
INSERT INTO telemetry (device_id, metric_name, value, unit, timestamp) VALUES
(1, 'temperature', 22.5, 'celsius', NOW() - INTERVAL '5 minutes'),
(1, 'temperature', 22.7, 'celsius', NOW() - INTERVAL '4 minutes'),
(2, 'humidity', 45.0, 'percent', NOW() - INTERVAL '5 minutes'),
(3, 'motion', 1.0, 'boolean', NOW() - INTERVAL '2 minutes')
ON CONFLICT DO NOTHING;
