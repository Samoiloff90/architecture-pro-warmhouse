const express = require('express');
const { Pool } = require('pg');
const amqp = require('amqplib');

const app = express();
app.use(express.json());

const PORT = process.env.PORT || 8083;

// PostgreSQL connection
const pool = new Pool({
    host: process.env.DB_HOST || 'localhost',
    port: process.env.DB_PORT || 5432,
    user: process.env.DB_USER || 'warmhouse',
    password: process.env.DB_PASSWORD || 'warmhouse123',
    database: process.env.DB_NAME || 'telemetry_db',
});

// RabbitMQ connection
let rabbitmqChannel = null;

async function connectRabbitMQ() {
    try {
        const rabbitmqUrl = process.env.RABBITMQ_URL || 'amqp://guest:guest@localhost';
        const connection = await amqp.connect(rabbitmqUrl);
        rabbitmqChannel = await connection.createChannel();

        await rabbitmqChannel.assertExchange('warmhouse', 'topic', { durable: true });
        const queue = await rabbitmqChannel.assertQueue('telemetry_queue', { durable: true });

        // Subscribe to device events
        await rabbitmqChannel.bindQueue(queue.queue, 'warmhouse', 'device.*');

        rabbitmqChannel.consume(queue.queue, (msg) => {
            if (msg) {
                const event = JSON.parse(msg.content.toString());
                console.log('Received event:', event.event_type, event.data);
                rabbitmqChannel.ack(msg);
            }
        });

        console.log('Connected to RabbitMQ');
    } catch (error) {
        console.error('Failed to connect to RabbitMQ:', error);
    }
}

// Publish event to RabbitMQ
async function publishEvent(eventType, data) {
    if (rabbitmqChannel) {
        try {
            const message = {
                event_type: eventType,
                timestamp: new Date().toISOString(),
                data
            };
            rabbitmqChannel.publish(
                'warmhouse',
                eventType,
                Buffer.from(JSON.stringify(message))
            );
            console.log('Published event:', eventType);
        } catch (error) {
            console.error('Failed to publish event:', error);
        }
    }
}

// Initialize database and RabbitMQ
(async () => {
    try {
        await pool.query('SELECT NOW()');
        console.log('Connected to PostgreSQL');
        await connectRabbitMQ();
    } catch (error) {
        console.error('Initialization error:', error);
    }
})();

// Routes
app.get('/', (req, res) => {
    res.json({ service: 'Telemetry Service', version: '1.0.0' });
});

app.get('/health', (req, res) => {
    res.json({ status: 'healthy' });
});

// Get telemetry data
app.get('/telemetry', async (req, res) => {
    try {
        const { device_id, metric_name, limit = 100 } = req.query;

        let query = 'SELECT * FROM telemetry WHERE 1=1';
        const params = [];

        if (device_id) {
            params.push(device_id);
            query += ` AND device_id = $${params.length}`;
        }

        if (metric_name) {
            params.push(metric_name);
            query += ` AND metric_name = $${params.length}`;
        }

        query += ' ORDER BY timestamp DESC';

        params.push(parseInt(limit));
        query += ` LIMIT $${params.length}`;

        const result = await pool.query(query, params);
        res.json(result.rows);
    } catch (error) {
        console.error('Error fetching telemetry:', error);
        res.status(500).json({ error: 'Internal server error' });
    }
});

// Post telemetry data
app.post('/telemetry', async (req, res) => {
    try {
        const { device_id, metric_name, value, unit } = req.body;

        if (!device_id || !metric_name || value === undefined) {
            return res.status(400).json({ error: 'Missing required fields' });
        }

        const result = await pool.query(
            `INSERT INTO telemetry (device_id, metric_name, value, unit, timestamp)
             VALUES ($1, $2, $3, $4, NOW())
             RETURNING *`,
            [device_id, metric_name, value, unit || 'unknown']
        );

        const telemetryData = result.rows[0];

        // Publish event
        await publishEvent('telemetry.received', {
            device_id: telemetryData.device_id,
            metric_name: telemetryData.metric_name,
            value: telemetryData.value,
            timestamp: telemetryData.timestamp
        });

        res.status(201).json(telemetryData);
    } catch (error) {
        console.error('Error creating telemetry:', error);
        res.status(500).json({ error: 'Internal server error' });
    }
});

// Get aggregated telemetry
app.get('/telemetry/aggregated', async (req, res) => {
    try {
        const { device_id, metric_name, interval = '1 hour' } = req.query;

        if (!device_id || !metric_name) {
            return res.status(400).json({ error: 'device_id and metric_name are required' });
        }

        const result = await pool.query(
            `SELECT
                date_trunc($1, timestamp) AS period,
                AVG(value) AS avg_value,
                MIN(value) AS min_value,
                MAX(value) AS max_value,
                COUNT(*) AS count
             FROM telemetry
             WHERE device_id = $2 AND metric_name = $3
             GROUP BY period
             ORDER BY period DESC
             LIMIT 24`,
            [interval, device_id, metric_name]
        );

        res.json(result.rows);
    } catch (error) {
        console.error('Error fetching aggregated telemetry:', error);
        res.status(500).json({ error: 'Internal server error' });
    }
});

app.listen(PORT, () => {
    console.log(`Telemetry Service listening on port ${PORT}`);
});
