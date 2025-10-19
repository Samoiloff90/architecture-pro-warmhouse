from fastapi import FastAPI, Query
from fastapi.responses import JSONResponse
import random
from datetime import datetime

app = FastAPI(title="Temperature API", version="1.0.0")

# Mapping между location и sensorId
LOCATION_TO_SENSOR_ID = {
    "Living Room": "1",
    "Bedroom": "2",
    "Kitchen": "3",
}

SENSOR_ID_TO_LOCATION = {
    "1": "Living Room",
    "2": "Bedroom",
    "3": "Kitchen",
}


@app.get("/")
def root():
    return {"message": "Temperature API is running", "version": "1.0.0"}


@app.get("/health")
def health_check():
    return {"status": "healthy"}


@app.get("/temperature")
def get_temperature(
        location: str = Query(None, description="Location name (e.g., Living Room)"),
        sensorId: str = Query(None, alias="sensorId", description="Sensor ID (e.g., 1, 2, 3)")
):
    """
    Возвращает случайную температуру для заданной локации или sensor ID.

    Логика из монолита:
    - Если location не указан, определяем его по sensorId
    - Если sensorId не указан, определяем его по location
    - Возвращаем случайное значение температуры
    """

    # If no location is provided, use a default based on sensor ID
    if not location and sensorId:
        location = SENSOR_ID_TO_LOCATION.get(sensorId, "Unknown")

    # If no sensor ID is provided, generate one based on location
    if not sensorId and location:
        sensorId = LOCATION_TO_SENSOR_ID.get(location, "0")

    # Если ничего не указано, возвращаем дефолт
    if not location and not sensorId:
        location = "Unknown"
        sensorId = "0"

    # Генерируем случайную температуру (от 15 до 30 градусов)
    temperature = round(random.uniform(15.0, 30.0), 2)

    response = {
        "location": location,
        "sensorId": sensorId,
        "temperature": temperature,
        "unit": "celsius",
        "timestamp": datetime.utcnow().isoformat() + "Z"
    }

    return JSONResponse(content=response)


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="0.0.0.0", port=8081)
