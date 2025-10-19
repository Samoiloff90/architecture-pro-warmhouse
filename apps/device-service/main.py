from fastapi import FastAPI, HTTPException, Depends
from pydantic import BaseModel
from typing import List, Optional
import os
import asyncpg
from datetime import datetime
import aio_pika
import json
import logging

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(title="Device Management Service", version="1.0.0")

# Database connection pool
db_pool = None
# RabbitMQ connection
rabbitmq_connection = None
rabbitmq_channel = None


# Pydantic models
class DeviceCreate(BaseModel):
    name: str
    device_type: str
    location: str
    room: str
    status: str = "active"


class Device(BaseModel):
    id: int
    name: str
    device_type: str
    location: str
    room: str
    status: str
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


# Database initialization
async def get_db_pool():
    global db_pool
    if db_pool is None:
        db_pool = await asyncpg.create_pool(
            host=os.getenv("DB_HOST", "localhost"),
            port=int(os.getenv("DB_PORT", "5432")),
            user=os.getenv("DB_USER", "warmhouse"),
            password=os.getenv("DB_PASSWORD", "warmhouse123"),
            database=os.getenv("DB_NAME", "device_db"),
            min_size=1,
            max_size=10
        )
        logger.info("Database pool created")
    return db_pool


# RabbitMQ initialization
async def get_rabbitmq_channel():
    global rabbitmq_connection, rabbitmq_channel
    if rabbitmq_channel is None:
        rabbitmq_url = os.getenv("RABBITMQ_URL", "amqp://guest:guest@localhost/")
        rabbitmq_connection = await aio_pika.connect_robust(rabbitmq_url)
        rabbitmq_channel = await rabbitmq_connection.channel()
        logger.info("RabbitMQ channel created")
    return rabbitmq_channel


# Publish event to RabbitMQ
async def publish_event(event_type: str, data: dict):
    try:
        channel = await get_rabbitmq_channel()
        exchange = await channel.declare_exchange("warmhouse", aio_pika.ExchangeType.TOPIC, durable=True)

        message = {
            "event_type": event_type,
            "timestamp": datetime.utcnow().isoformat(),
            "data": data
        }

        await exchange.publish(
            aio_pika.Message(body=json.dumps(message).encode()),
            routing_key=event_type
        )
        logger.info(f"Published event: {event_type}")
    except Exception as e:
        logger.error(f"Failed to publish event: {e}")


@app.on_event("startup")
async def startup():
    await get_db_pool()
    await get_rabbitmq_channel()
    logger.info("Device Service started")


@app.on_event("shutdown")
async def shutdown():
    global db_pool, rabbitmq_connection
    if db_pool:
        await db_pool.close()
    if rabbitmq_connection:
        await rabbitmq_connection.close()
    logger.info("Device Service stopped")


@app.get("/")
def root():
    return {"service": "Device Management Service", "version": "1.0.0"}


@app.get("/health")
def health():
    return {"status": "healthy"}


@app.get("/devices", response_model=List[Device])
async def get_devices(
        location: Optional[str] = None,
        status: Optional[str] = None
):
    pool = await get_db_pool()

    query = "SELECT * FROM devices WHERE 1=1"
    params = []

    if location:
        params.append(location)
        query += f" AND location = ${len(params)}"

    if status:
        params.append(status)
        query += f" AND status = ${len(params)}"

    query += " ORDER BY created_at DESC"

    async with pool.acquire() as conn:
        rows = await conn.fetch(query, *params)
        return [dict(row) for row in rows]


@app.get("/devices/{device_id}", response_model=Device)
async def get_device(device_id: int):
    pool = await get_db_pool()

    async with pool.acquire() as conn:
        row = await conn.fetchrow("SELECT * FROM devices WHERE id = $1", device_id)
        if not row:
            raise HTTPException(status_code=404, detail="Device not found")
        return dict(row)


@app.post("/devices", response_model=Device, status_code=201)
async def create_device(device: DeviceCreate):
    pool = await get_db_pool()

    async with pool.acquire() as conn:
        row = await conn.fetchrow(
            """
            INSERT INTO devices (name, device_type, location, room, status)
            VALUES ($1, $2, $3, $4, $5)
            RETURNING *
            """,
            device.name, device.device_type, device.location, device.room, device.status
        )

        device_data = dict(row)

        # Publish event
        await publish_event("device.created", {
            "device_id": device_data["id"],
            "name": device_data["name"],
            "device_type": device_data["device_type"],
            "location": device_data["location"]
        })

        return device_data


@app.put("/devices/{device_id}", response_model=Device)
async def update_device(device_id: int, device: DeviceCreate):
    pool = await get_db_pool()

    async with pool.acquire() as conn:
        row = await conn.fetchrow(
            """
            UPDATE devices
            SET name = $1, device_type = $2, location = $3, room = $4, status = $5, updated_at = NOW()
            WHERE id = $6
            RETURNING *
            """,
            device.name, device.device_type, device.location, device.room, device.status, device_id
        )

        if not row:
            raise HTTPException(status_code=404, detail="Device not found")

        device_data = dict(row)

        # Publish event
        await publish_event("device.updated", {
            "device_id": device_data["id"],
            "name": device_data["name"]
        })

        return device_data


@app.delete("/devices/{device_id}", status_code=204)
async def delete_device(device_id: int):
    pool = await get_db_pool()

    async with pool.acquire() as conn:
        result = await conn.execute("DELETE FROM devices WHERE id = $1", device_id)
        if result == "DELETE 0":
            raise HTTPException(status_code=404, detail="Device not found")

        # Publish event
        await publish_event("device.deleted", {"device_id": device_id})


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="0.0.0.0", port=8082)
